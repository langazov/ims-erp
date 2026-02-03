package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ims-erp/system/internal/config"
	"github.com/ims-erp/system/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Redis struct {
	client redis.UniversalClient
	config config.RedisConfig
	logger *logger.Logger
	tracer trace.Tracer
}

func NewRedis(cfg config.RedisConfig, log *logger.Logger) (*Redis, error) {
	var client redis.UniversalClient

	if cfg.Mode == "cluster" {
		client = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:        cfg.Addresses,
			Username:     cfg.Username,
			Password:     cfg.Password,
			PoolSize:     cfg.PoolSize,
			MinIdleConns: cfg.MinIdleConns,
			MaxRetries:   cfg.MaxRetries,
			DialTimeout:  cfg.DialTimeout,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
			PoolTimeout:  cfg.PoolTimeout,
		})
	} else if cfg.Mode == "sentinel" {
		client = redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:    cfg.MasterName,
			SentinelAddrs: cfg.Addresses,
			Username:      cfg.Username,
			Password:      cfg.Password,
			PoolSize:      cfg.PoolSize,
			MinIdleConns:  cfg.MinIdleConns,
			MaxRetries:    cfg.MaxRetries,
			DialTimeout:   cfg.DialTimeout,
			ReadTimeout:   cfg.ReadTimeout,
			WriteTimeout:  cfg.WriteTimeout,
		})
	} else {
		addr := cfg.Addresses[0]
		if addr == "" {
			addr = "localhost:6379"
		}
		client = redis.NewClient(&redis.Options{
			Addr:         addr,
			Username:     cfg.Username,
			Password:     cfg.Password,
			DB:           cfg.Database,
			PoolSize:     cfg.PoolSize,
			MinIdleConns: cfg.MinIdleConns,
			MaxRetries:   cfg.MaxRetries,
			DialTimeout:  cfg.DialTimeout,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
			PoolTimeout:  cfg.PoolTimeout,
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Redis{
		client: client,
		config: cfg,
		logger: log,
		tracer: otel.Tracer("redis"),
	}, nil
}

func (r *Redis) Client() redis.UniversalClient {
	return r.client
}

func (r *Redis) Close() error {
	return r.client.Close()
}

func (r *Redis) Health(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

type Cache struct {
	redis  *Redis
	prefix string
	logger *logger.Logger
	tracer trace.Tracer
}

func NewCache(redis *Redis, prefix string, log *logger.Logger) *Cache {
	return &Cache{
		redis:  redis,
		prefix: prefix,
		logger: log,
		tracer: otel.Tracer("cache"),
	}
}

func (c *Cache) key(key string) string {
	return fmt.Sprintf("%s:%s", c.prefix, key)
}

func (c *Cache) Get(ctx context.Context, key string) (string, error) {
	ctx, span := c.tracer.Start(ctx, "redis.get",
		trace.WithAttributes(attribute.String("cache.key", key)),
	)
	defer span.End()

	result, err := c.redis.client.Get(ctx, c.key(key)).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil
		}
		span.RecordError(err)
		return "", fmt.Errorf("failed to get from cache: %w", err)
	}

	return result, nil
}

func (c *Cache) GetBytes(ctx context.Context, key string) ([]byte, error) {
	ctx, span := c.tracer.Start(ctx, "redis.get_bytes",
		trace.WithAttributes(attribute.String("cache.key", key)),
	)
	defer span.End()

	result, err := c.redis.client.Get(ctx, c.key(key)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get bytes from cache: %w", err)
	}

	return result, nil
}

func (c *Cache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	ctx, span := c.tracer.Start(ctx, "redis.set",
		trace.WithAttributes(
			attribute.String("cache.key", key),
			attribute.String("cache.expiration", expiration.String()),
		),
	)
	defer span.End()

	var data []byte
	var err error

	switch v := value.(type) {
	case string:
		data = []byte(v)
	case []byte:
		data = v
	default:
		data, err = json.Marshal(value)
		if err != nil {
			span.RecordError(err)
			return fmt.Errorf("failed to marshal value: %w", err)
		}
	}

	if err := c.redis.client.Set(ctx, c.key(key), data, expiration).Err(); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to set cache: %w", err)
	}

	return nil
}

func (c *Cache) Delete(ctx context.Context, keys ...string) error {
	ctx, span := c.tracer.Start(ctx, "redis.delete",
		trace.WithAttributes(attribute.Int("cache.keys_count", len(keys))),
	)
	defer span.End()

	if len(keys) == 0 {
		return nil
	}

	redisKeys := make([]string, len(keys))
	for i, key := range keys {
		redisKeys[i] = c.key(key)
	}

	if err := c.redis.client.Del(ctx, redisKeys...).Err(); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to delete from cache: %w", err)
	}

	return nil
}

func (c *Cache) DeletePattern(ctx context.Context, pattern string) error {
	ctx, span := c.tracer.Start(ctx, "redis.delete_pattern",
		trace.WithAttributes(attribute.String("cache.pattern", pattern)),
	)
	defer span.End()

	iter := c.redis.client.Scan(ctx, 0, c.key(pattern), 100).Iterator()
	for iter.Next(ctx) {
		if err := c.redis.client.Del(ctx, iter.Val()).Err(); err != nil {
			span.RecordError(err)
			return fmt.Errorf("failed to delete pattern: %w", err)
		}
	}

	if err := iter.Err(); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to scan keys: %w", err)
	}

	return nil
}

func (c *Cache) Exists(ctx context.Context, key string) (bool, error) {
	ctx, span := c.tracer.Start(ctx, "redis.exists",
		trace.WithAttributes(attribute.String("cache.key", key)),
	)
	defer span.End()

	result, err := c.redis.client.Exists(ctx, c.key(key)).Result()
	if err != nil {
		span.RecordError(err)
		return false, fmt.Errorf("failed to check existence: %w", err)
	}

	return result > 0, nil
}

func (c *Cache) Expire(ctx context.Context, key string, expiration time.Duration) error {
	ctx, span := c.tracer.Start(ctx, "redis.expire",
		trace.WithAttributes(
			attribute.String("cache.key", key),
			attribute.String("cache.expiration", expiration.String()),
		),
	)
	defer span.End()

	if err := c.redis.client.Expire(ctx, c.key(key), expiration).Err(); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to set expiration: %w", err)
	}

	return nil
}

func (c *Cache) TTL(ctx context.Context, key string) (time.Duration, error) {
	ctx, span := c.tracer.Start(ctx, "redis.ttl",
		trace.WithAttributes(attribute.String("cache.key", key)),
	)
	defer span.End()

	ttl, err := c.redis.client.TTL(ctx, c.key(key)).Result()
	if err != nil {
		span.RecordError(err)
		return 0, fmt.Errorf("failed to get TTL: %w", err)
	}

	return ttl, nil
}

func (c *Cache) Incr(ctx context.Context, key string) (int64, error) {
	ctx, span := c.tracer.Start(ctx, "redis.incr",
		trace.WithAttributes(attribute.String("cache.key", key)),
	)
	defer span.End()

	result, err := c.redis.client.Incr(ctx, c.key(key)).Result()
	if err != nil {
		span.RecordError(err)
		return 0, fmt.Errorf("failed to increment: %w", err)
	}

	return result, nil
}

func (c *Cache) Decr(ctx context.Context, key string) (int64, error) {
	ctx, span := c.tracer.Start(ctx, "redis.decr",
		trace.WithAttributes(attribute.String("cache.key", key)),
	)
	defer span.End()

	result, err := c.redis.client.Decr(ctx, c.key(key)).Result()
	if err != nil {
		span.RecordError(err)
		return 0, fmt.Errorf("failed to decrement: %w", err)
	}

	return result, nil
}

func (c *Cache) HGet(ctx context.Context, key, field string) (string, error) {
	ctx, span := c.tracer.Start(ctx, "redis.hget",
		trace.WithAttributes(
			attribute.String("cache.key", key),
			attribute.String("cache.field", field),
		),
	)
	defer span.End()

	result, err := c.redis.client.HGet(ctx, c.key(key), field).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil
		}
		span.RecordError(err)
		return "", fmt.Errorf("failed to get hash field: %w", err)
	}

	return result, nil
}

func (c *Cache) HSet(ctx context.Context, key, field string, value interface{}) error {
	ctx, span := c.tracer.Start(ctx, "redis.hset",
		trace.WithAttributes(
			attribute.String("cache.key", key),
			attribute.String("cache.field", field),
		),
	)
	defer span.End()

	var data string
	switch v := value.(type) {
	case string:
		data = v
	default:
		jsonData, err := json.Marshal(value)
		if err != nil {
			span.RecordError(err)
			return fmt.Errorf("failed to marshal value: %w", err)
		}
		data = string(jsonData)
	}

	if err := c.redis.client.HSet(ctx, c.key(key), field, data).Err(); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to set hash field: %w", err)
	}

	return nil
}

func (c *Cache) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	ctx, span := c.tracer.Start(ctx, "redis.hgetall",
		trace.WithAttributes(attribute.String("cache.key", key)),
	)
	defer span.End()

	result, err := c.redis.client.HGetAll(ctx, c.key(key)).Result()
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get all hash fields: %w", err)
	}

	return result, nil
}

func (c *Cache) ZAdd(ctx context.Context, key string, score float64, member string) error {
	ctx, span := c.tracer.Start(ctx, "redis.zadd",
		trace.WithAttributes(
			attribute.String("cache.key", key),
			attribute.Float64("cache.score", score),
		),
	)
	defer span.End()

	if err := c.redis.client.ZAdd(ctx, c.key(key), redis.Z{
		Score:  score,
		Member: member,
	}).Err(); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to add to sorted set: %w", err)
	}

	return nil
}

func (c *Cache) ZRangeByScore(ctx context.Context, key string, min, max string) ([]string, error) {
	ctx, span := c.tracer.Start(ctx, "redis.zrangebyscore",
		trace.WithAttributes(
			attribute.String("cache.key", key),
			attribute.String("cache.min", min),
			attribute.String("cache.max", max),
		),
	)
	defer span.End()

	result, err := c.redis.client.ZRangeByScore(ctx, c.key(key), &redis.ZRangeBy{
		Min: min,
		Max: max,
	}).Result()
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get range by score: %w", err)
	}

	return result, nil
}

func (c *Cache) Lock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	ctx, span := c.tracer.Start(ctx, "redis.lock",
		trace.WithAttributes(
			attribute.String("cache.key", key),
			attribute.String("cache.ttl", ttl.String()),
		),
	)
	defer span.End()

	lockKey := c.key("lock:" + key)
	result, err := c.redis.client.SetNX(ctx, lockKey, "locked", ttl).Result()
	if err != nil {
		span.RecordError(err)
		return false, fmt.Errorf("failed to acquire lock: %w", err)
	}

	return result, nil
}

func (c *Cache) Unlock(ctx context.Context, key string) error {
	ctx, span := c.tracer.Start(ctx, "redis.unlock",
		trace.WithAttributes(attribute.String("cache.key", key)),
	)
	defer span.End()

	lockKey := c.key("lock:" + key)
	if err := c.redis.client.Del(ctx, lockKey).Err(); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to release lock: %w", err)
	}

	return nil
}

type DistributedLock struct {
	cache *Cache
	key   string
	ttl   time.Duration
}

func (c *Cache) AcquireLock(key string, ttl time.Duration) *DistributedLock {
	return &DistributedLock{
		cache: c,
		key:   key,
		ttl:   ttl,
	}
}

func (l *DistributedLock) TryLock(ctx context.Context) (bool, error) {
	return l.cache.Lock(ctx, l.key, l.ttl)
}

func (l *DistributedLock) Unlock(ctx context.Context) error {
	return l.cache.Unlock(ctx, l.key)
}

type RateLimiter struct {
	redis  *Redis
	logger *logger.Logger
	tracer trace.Tracer
}

func NewRateLimiter(redis *Redis, log *logger.Logger) *RateLimiter {
	return &RateLimiter{
		redis:  redis,
		logger: log,
		tracer: otel.Tracer("rate-limiter"),
	}
}

func (r *RateLimiter) Allow(ctx context.Context, identifier string, limit int, window time.Duration) (bool, int, error) {
	ctx, span := r.tracer.Start(ctx, "redis.rate_limit",
		trace.WithAttributes(
			attribute.String("rate_limit.identifier", identifier),
			attribute.Int("rate_limit.limit", limit),
			attribute.String("rate_limit.window", window.String()),
		),
	)
	defer span.End()

	key := fmt.Sprintf("ratelimit:%s", identifier)
	now := time.Now().UnixNano()
	windowStart := now - int64(window)

	pipe := r.redis.client.Pipeline()

	pipe.ZAdd(ctx, key, redis.Z{
		Score:  float64(now),
		Member: fmt.Sprintf("%d", now),
	})

	pipe.ZRemRangeByScore(ctx, key, "-inf", fmt.Sprintf("%d", windowStart))

	count := pipe.ZCard(ctx, key)

	pipe.Expire(ctx, key, window)

	_, err := pipe.Exec(ctx)
	if err != nil {
		span.RecordError(err)
		return false, 0, fmt.Errorf("failed to execute rate limit pipeline: %w", err)
	}

	currentCount := int(count.Val())
	span.SetAttributes(attribute.Int("rate_limit.current", currentCount))

	if currentCount > limit {
		return false, currentCount, nil
	}

	return true, currentCount, nil
}
