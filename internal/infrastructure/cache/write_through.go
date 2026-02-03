package cache

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/redis/go-redis/v9"
)

// WriteThroughCache implements write-through caching with compression
type WriteThroughCache struct {
	client           redis.Cmdable
	compress         bool
	compressionLevel int
}

// NewWriteThroughCache creates a new write-through cache
func NewWriteThroughCache(client redis.Cmdable, compress bool) *WriteThroughCache {
	return &WriteThroughCache{
		client:           client,
		compress:         compress,
		compressionLevel: gzip.BestSpeed,
	}
}

// Set writes data to cache and database (write-through)
func (c *WriteThroughCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration, dbWriteFunc func() error) error {
	// First write to database
	if err := dbWriteFunc(); err != nil {
		return fmt.Errorf("database write failed: %w", err)
	}

	// Then write to cache
	if err := c.writeToCache(ctx, key, value, ttl); err != nil {
		// Log but don't fail - we already wrote to DB
		// In production, this would trigger a cache sync
	}

	return nil
}

// writeToCache writes data to Redis with optional compression
func (c *WriteThroughCache) writeToCache(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	// Compress if enabled and data is large enough (>1KB)
	if c.compress && len(data) > 1024 {
		compressed, err := c.compressData(data)
		if err == nil {
			data = compressed
			key = key + ":gzip"
		}
	}

	return c.client.Set(ctx, key, data, ttl).Err()
}

// Get reads data from cache
func (c *WriteThroughCache) Get(ctx context.Context, key string, dest interface{}) error {
	// Try compressed key first
	data, err := c.client.Get(ctx, key+":gzip").Bytes()
	if err == nil {
		// Decompress
		decompressed, err := c.decompressData(data)
		if err != nil {
			return fmt.Errorf("failed to decompress: %w", err)
		}
		return json.Unmarshal(decompressed, dest)
	}

	// Try uncompressed key
	data, err = c.client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}

	return json.Unmarshal(data, dest)
}

// Invalidate removes data from cache
func (c *WriteThroughCache) Invalidate(ctx context.Context, pattern string) error {
	// Find keys matching pattern
	keys, err := c.client.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		return c.client.Del(ctx, keys...).Err()
	}

	return nil
}

// compressData compresses data using gzip
func (c *WriteThroughCache) compressData(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer, err := gzip.NewWriterLevel(&buf, c.compressionLevel)
	if err != nil {
		return nil, err
	}

	if _, err := writer.Write(data); err != nil {
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// decompressData decompresses gzip data
func (c *WriteThroughCache) decompressData(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}

// CacheWarmer warms up cache with hot data
type CacheWarmer struct {
	cache    *WriteThroughCache
	warmers  []CacheWarmerFunc
	interval time.Duration
	stopCh   chan struct{}
}

// CacheWarmerFunc is a function that warms cache for a specific entity
type CacheWarmerFunc func(ctx context.Context) error

// NewCacheWarmer creates a new cache warmer
func NewCacheWarmer(cache *WriteThroughCache, interval time.Duration) *CacheWarmer {
	return &CacheWarmer{
		cache:    cache,
		warmers:  []CacheWarmerFunc{},
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

// Register adds a warmer function
func (w *CacheWarmer) Register(warmer CacheWarmerFunc) {
	w.warmers = append(w.warmers, warmer)
}

// Start begins the warming process
func (w *CacheWarmer) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)

	go func() {
		// Warm immediately on start
		w.warm(ctx)

		for {
			select {
			case <-ticker.C:
				w.warm(ctx)
			case <-w.stopCh:
				ticker.Stop()
				return
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}

// Stop stops the warmer
func (w *CacheWarmer) Stop() {
	close(w.stopCh)
}

// warm executes all warmer functions
func (w *CacheWarmer) warm(ctx context.Context) {
	for _, warmer := range w.warmers {
		if err := warmer(ctx); err != nil {
			// Log error but continue with other warmers
			continue
		}
	}
}

// TenantShardedCache implements cache sharding by tenant
type TenantShardedCache struct {
	shards   map[string]redis.Cmdable
	hashFunc func(tenantID string) string
}

// NewTenantShardedCache creates a new sharded cache
func NewTenantShardedCache(shards []redis.Cmdable) *TenantShardedCache {
	shardMap := make(map[string]redis.Cmdable)
	for i, shard := range shards {
		shardMap[fmt.Sprintf("shard-%d", i)] = shard
	}

	return &TenantShardedCache{
		shards:   shardMap,
		hashFunc: defaultHashFunc(len(shards)),
	}
}

func defaultHashFunc(numShards int) func(string) string {
	return func(tenantID string) string {
		// Simple hash function - in production use consistent hashing
		hash := 0
		for _, c := range tenantID {
			hash += int(c)
		}
		return fmt.Sprintf("shard-%d", hash%numShards)
	}
}

// getShard returns the Redis client for a tenant
func (c *TenantShardedCache) getShard(tenantID string) redis.Cmdable {
	shardKey := c.hashFunc(tenantID)
	return c.shards[shardKey]
}

// Set writes data to the appropriate shard
func (c *TenantShardedCache) Set(ctx context.Context, tenantID, key string, value interface{}, ttl time.Duration) error {
	shard := c.getShard(tenantID)
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return shard.Set(ctx, fmt.Sprintf("%s:%s", tenantID, key), data, ttl).Err()
}

// Get reads data from the appropriate shard
func (c *TenantShardedCache) Get(ctx context.Context, tenantID, key string, dest interface{}) error {
	shard := c.getShard(tenantID)
	data, err := shard.Get(ctx, fmt.Sprintf("%s:%s", tenantID, key)).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}
