package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App           AppConfig           `mapstructure:"app"`
	MongoDB       MongoDBConfig       `mapstructure:"mongodb"`
	Redis         RedisConfig         `mapstructure:"redis"`
	NATS          NATSConfig          `mapstructure:"nats"`
	MinIO         MinIOConfig         `mapstructure:"minio"`
	Elasticsearch ElasticsearchConfig `mapstructure:"elasticsearch"`
	Auth          AuthConfig          `mapstructure:"auth"`
	Security      SecurityConfig      `mapstructure:"security"`
	Tracing       TracingConfig       `mapstructure:"tracing"`
	Logging       LoggingConfig       `mapstructure:"logging"`
}

type AppConfig struct {
	Name            string        `mapstructure:"name"`
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	Environment     string        `mapstructure:"environment"`
	Version         string        `mapstructure:"version"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
}

type MongoDBConfig struct {
	URI               string        `mapstructure:"uri"`
	Database          string        `mapstructure:"database"`
	Username          string        `mapstructure:"username"`
	Password          string        `mapstructure:"password"`
	AuthDatabase      string        `mapstructure:"auth_database"`
	MaxPoolSize       uint64        `mapstructure:"max_pool_size"`
	MinPoolSize       uint64        `mapstructure:"min_pool_size"`
	MaxConnIdleTime   time.Duration `mapstructure:"max_conn_idle_time"`
	ConnectTimeout    time.Duration `mapstructure:"connect_timeout"`
	ServerSelection   time.Duration `mapstructure:"server_selection_timeout"`
	HeartbeatInterval time.Duration `mapstructure:"heartbeat_interval"`
}

type RedisConfig struct {
	Mode            string        `mapstructure:"mode"` // standalone, sentinel, cluster
	Addresses       []string      `mapstructure:"addresses"`
	MasterName      string        `mapstructure:"master_name"`
	Username        string        `mapstructure:"username"`
	Password        string        `mapstructure:"password"`
	Database        int           `mapstructure:"database"`
	PoolSize        int           `mapstructure:"pool_size"`
	MinIdleConns    int           `mapstructure:"min_idle_conns"`
	MaxRetries      int           `mapstructure:"max_retries"`
	MinRetryBackoff time.Duration `mapstructure:"min_retry_backoff"`
	MaxRetryBackoff time.Duration `mapstructure:"max_retry_backoff"`
	DialTimeout     time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	PoolTimeout     time.Duration `mapstructure:"pool_timeout"`
	IdleTimeout     time.Duration `mapstructure:"idle_timeout"`
	IdleCheckFreq   time.Duration `mapstructure:"idle_check_freq"`
	TLSEnabled      bool          `mapstructure:"tls_enabled"`
	TLSCertFile     string        `mapstructure:"tls_cert_file"`
	TLSKeyFile      string        `mapstructure:"tls_key_file"`
}

type NATSConfig struct {
	URLs           []string        `mapstructure:"urls"`
	Username       string          `mapstructure:"username"`
	Password       string          `mapstructure:"password"`
	Token          string          `mapstructure:"token"`
	MaxReconnect   int             `mapstructure:"max_reconnect"`
	ReconnectWait  time.Duration   `mapstructure:"reconnect_wait"`
	ConnectTimeout time.Duration   `mapstructure:"connect_timeout"`
	PingInterval   time.Duration   `mapstructure:"ping_interval"`
	MaxPingsOut    int             `mapstructure:"max_pings_out"`
	TLSEnabled     bool            `mapstructure:"tls_enabled"`
	TLSCertFile    string          `mapstructure:"tls_cert_file"`
	TLSKeyFile     string          `mapstructure:"tls_key_file"`
	TLSCACertFile  string          `mapstructure:"tls_ca_cert_file"`
	JetStream      JetStreamConfig `mapstructure:"jetstream"`
}

type JetStreamConfig struct {
	Enabled      bool   `mapstructure:"enabled"`
	Domain       string `mapstructure:"domain"`
	StreamPrefix string `mapstructure:"stream_prefix"`
}

type MinIOConfig struct {
	Endpoint     string `mapstructure:"endpoint"`
	AccessKey    string `mapstructure:"access_key"`
	SecretKey    string `mapstructure:"secret_key"`
	UseSSL       bool   `mapstructure:"use_ssl"`
	Secure       bool   `mapstructure:"secure"`
	Region       string `mapstructure:"region"`
	BucketPrefix string `mapstructure:"bucket_prefix"`
	MaxPartSize  int64  `mapstructure:"max_part_size"`
}

type ElasticsearchConfig struct {
	Addresses     []string      `mapstructure:"addresses"`
	Username      string        `mapstructure:"username"`
	Password      string        `mapstructure:"password"`
	CloudID       string        `mapstructure:"cloud_id"`
	APIKey        string        `mapstructure:"api_key"`
	MaxRetries    int           `mapstructure:"max_retries"`
	RetryOnStatus []int         `mapstructure:"retry_on_status"`
	RetryBackoff  time.Duration `mapstructure:"retry_backoff"`
	Transport     time.Duration `mapstructure:"transport"`
	IndexPrefix   string        `mapstructure:"index_prefix"`
	SessionTTL    time.Duration `mapstructure:"session_ttl"`
}

type AuthConfig struct {
	JWT_SECRET             string        `mapstructure:"jwt_secret"`
	JWT_ISSUER             string        `mapstructure:"jwt_issuer"`
	AccessTokenExpiry      time.Duration `mapstructure:"access_token_expiry"`
	RefreshTokenExpiry     time.Duration `mapstructure:"refresh_token_expiry"`
	PasswordMinLength      int           `mapstructure:"password_min_length"`
	PasswordRequireUpper   bool          `mapstructure:"password_require_upper"`
	PasswordRequireLower   bool          `mapstructure:"password_require_lower"`
	PasswordRequireNumber  bool          `mapstructure:"password_require_number"`
	PasswordRequireSpecial bool          `mapstructure:"password_require_special"`
	MaxLoginAttempts       int           `mapstructure:"max_login_attempts"`
	LockoutDuration        time.Duration `mapstructure:"lockout_duration"`
	SessionTTL             time.Duration `mapstructure:"session_ttl"`
	MFAEnabled             bool          `mapstructure:"mfa_enabled"`
	MFAType                string        `mapstructure:"mfa_type"` // totp, email, sms
}

type SecurityConfig struct {
	EncryptionKey       string        `mapstructure:"encryption_key"`
	EncryptionAlgorithm string        `mapstructure:"encryption_algorithm"`
	RateLimitRequests   int           `mapstructure:"rate_limit_requests"`
	RateLimitWindow     time.Duration `mapstructure:"rate_limit_window"`
	CORSDomain          []string      `mapstructure:"cors_domain"`
	AllowedHeaders      []string      `mapstructure:"allowed_headers"`
	AllowedMethods      []string      `mapstructure:"allowed_methods"`
	MaxRequestBodySize  int64         `mapstructure:"max_request_body_size"`
}

type TracingConfig struct {
	Enabled      bool    `mapstructure:"enabled"`
	ServiceName  string  `mapstructure:"service_name"`
	ExporterType string  `mapstructure:"exporter_type"` // stdout, otlp, jaeger
	Endpoint     string  `mapstructure:"endpoint"`
	SamplerType  string  `mapstructure:"sampler_type"` // always, never, ratio
	SamplerRatio float64 `mapstructure:"sampler_ratio"`
	SampleParam  string  `mapstructure:"sample_param"`
}

type LoggingConfig struct {
	Level      string `mapstructure:"level"`  // debug, info, warn, error
	Format     string `mapstructure:"format"` // json, console
	OutputPath string `mapstructure:"output_path"`
	ErrorPath  string `mapstructure:"error_path"`
	AddSource  bool   `mapstructure:"add_source"`
	Caller     bool   `mapstructure:"caller"`
}

func Load(configPath string, configName string) (*Config, error) {
	v := viper.New()

	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName(configName)
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("./config")
		v.AddConfigPath("/etc/erp-system")
		v.AddConfigPath("$HOME/.erp-system")
	}

	v.SetEnvPrefix("ERP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	cfg.applyDefaults()
	cfg.validate()

	return &cfg, nil
}

func (c *Config) applyDefaults() {
	if c.App.Port == 0 {
		c.App.Port = 8080
	}
	if c.App.ShutdownTimeout == 0 {
		c.App.ShutdownTimeout = 30 * time.Second
	}
	if c.App.ReadTimeout == 0 {
		c.App.ReadTimeout = 30 * time.Second
	}
	if c.App.WriteTimeout == 0 {
		c.App.WriteTimeout = 30 * time.Second
	}
	if c.MongoDB.MaxPoolSize == 0 {
		c.MongoDB.MaxPoolSize = 100
	}
	if c.MongoDB.MinPoolSize == 0 {
		c.MongoDB.MinPoolSize = 10
	}
	if c.MongoDB.MaxConnIdleTime == 0 {
		c.MongoDB.MaxConnIdleTime = 5 * time.Minute
	}
	if c.MongoDB.ConnectTimeout == 0 {
		c.MongoDB.ConnectTimeout = 10 * time.Second
	}
	if c.MongoDB.ServerSelection == 0 {
		c.MongoDB.ServerSelection = 5 * time.Second
	}
	if c.Redis.PoolSize == 0 {
		c.Redis.PoolSize = 100
	}
	if c.Redis.MaxRetries == 0 {
		c.Redis.MaxRetries = 3
	}
	if c.NATS.MaxReconnect == 0 {
		c.NATS.MaxReconnect = 60
	}
	if c.NATS.ReconnectWait == 0 {
		c.NATS.ReconnectWait = 2 * time.Second
	}
	if c.Auth.AccessTokenExpiry == 0 {
		c.Auth.AccessTokenExpiry = 15 * time.Minute
	}
	if c.Auth.RefreshTokenExpiry == 0 {
		c.Auth.RefreshTokenExpiry = 7 * 24 * time.Hour
	}
	if c.Auth.PasswordMinLength == 0 {
		c.Auth.PasswordMinLength = 8
	}
	if c.Auth.MaxLoginAttempts == 0 {
		c.Auth.MaxLoginAttempts = 5
	}
	if c.Auth.LockoutDuration == 0 {
		c.Auth.LockoutDuration = 15 * time.Minute
	}
	if c.Security.RateLimitRequests == 0 {
		c.Security.RateLimitRequests = 1000
	}
	if c.Security.RateLimitWindow == 0 {
		c.Security.RateLimitWindow = time.Minute
	}
	if c.Tracing.SamplerRatio == 0 {
		c.Tracing.SamplerRatio = 1.0
	}
	if c.Logging.Level == "" {
		c.Logging.Level = "info"
	}
	if c.Logging.Format == "" {
		c.Logging.Format = "json"
	}
}

func (c *Config) validate() error {
	if c.App.Name == "" {
		return fmt.Errorf("app.name is required")
	}
	if c.MongoDB.URI == "" {
		return fmt.Errorf("mongodb.uri is required")
	}
	if c.MongoDB.Database == "" {
		return fmt.Errorf("mongodb.database is required")
	}
	if len(c.NATS.URLs) == 0 {
		return fmt.Errorf("nats.urls is required")
	}
	return nil
}

func (c *Config) GetMongoURI() string {
	if c.MongoDB.Username != "" && c.MongoDB.Password != "" {
		return fmt.Sprintf("mongodb://%s:%s@%s/%s?authSource=%s",
			c.MongoDB.Username,
			c.MongoDB.Password,
			c.MongoDB.URI,
			c.MongoDB.Database,
			c.MongoDB.AuthDatabase,
		)
	}
	return fmt.Sprintf("mongodb://%s/%s", c.MongoDB.URI, c.MongoDB.Database)
}

func (c *Config) GetRedisAddr() string {
	if len(c.Redis.Addresses) > 0 {
		return c.Redis.Addresses[0]
	}
	return "localhost:6379"
}
