package logger

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	*zap.Logger
	sugar  *zap.SugaredLogger
	config Config
}

type Config struct {
	Level       string `mapstructure:"level"`
	Format      string `mapstructure:"format"`
	OutputPath  string `mapstructure:"output_path"`
	ErrorPath   string `mapstructure:"error_path"`
	AddSource   bool   `mapstructure:"add_source"`
	Caller      bool   `mapstructure:"caller"`
	ServiceName string `mapstructure:"service_name"`
}

func New(cfg Config) (*Logger, error) {
	var level zapcore.Level
	switch cfg.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn", "warning":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	case "fatal":
		level = zapcore.FatalLevel
	default:
		level = zapcore.InfoLevel
	}

	var cores []zapcore.Core

	if cfg.Format == "json" {
		encoderConfig := zap.NewProductionEncoderConfig()
		encoderConfig.TimeKey = "timestamp"
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

		if cfg.AddSource {
			encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
		}

		jsonEncoder := zapcore.NewJSONEncoder(encoderConfig)

		if cfg.OutputPath != "" {
			if err := os.MkdirAll(filepath.Dir(cfg.OutputPath), 0755); err != nil {
				return nil, fmt.Errorf("failed to create log directory: %w", err)
			}
			outputFile, err := os.OpenFile(cfg.OutputPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return nil, fmt.Errorf("failed to open log file: %w", err)
			}
			cores = append(cores, zapcore.NewCore(jsonEncoder, outputFile, level))
		} else {
			cores = append(cores, zapcore.NewCore(jsonEncoder, os.Stdout, level))
		}

		if cfg.ErrorPath != "" {
			if err := os.MkdirAll(filepath.Dir(cfg.ErrorPath), 0755); err != nil {
				return nil, fmt.Errorf("failed to create error log directory: %w", err)
			}
			errorFile, err := os.OpenFile(cfg.ErrorPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return nil, fmt.Errorf("failed to open error log file: %w", err)
			}
			errorCore := zapcore.NewCore(jsonEncoder, errorFile, zapcore.ErrorLevel)
			cores = append(cores, errorCore)
		}
	} else {
		encoderConfig := zap.NewDevelopmentEncoderConfig()
		encoderConfig.TimeKey = "timestamp"
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

		if cfg.AddSource {
			encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
		}

		consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
		cores = append(cores, zapcore.NewCore(consoleEncoder, os.Stdout, level))
	}

	var options []zap.Option
	if cfg.Caller {
		options = append(options, zap.AddCaller())
	}
	options = append(options, zap.AddStacktrace(zapcore.ErrorLevel))

	combinedCore := zapcore.NewTee(cores...)
	logger := zap.New(combinedCore, options...)

	return &Logger{
		Logger: logger,
		sugar:  logger.Sugar().With("service", cfg.ServiceName),
		config: cfg,
	}, nil
}

func (l *Logger) With(ctx context.Context) *zap.SugaredLogger {
	traceID := GetTraceID(ctx)
	userID := GetUserID(ctx)
	tenantID := GetTenantID(ctx)

	sugar := l.sugar
	if traceID != "" {
		sugar = sugar.With("trace_id", traceID)
	}
	if userID != "" {
		sugar = sugar.With("user_id", userID)
	}
	if tenantID != "" {
		sugar = sugar.With("tenant_id", tenantID)
	}
	return sugar
}

func (l *Logger) Debug(msg string, fields ...interface{}) {
	l.sugar.Debugw(msg, fields...)
}

func (l *Logger) Info(msg string, fields ...interface{}) {
	l.sugar.Infow(msg, fields...)
}

func (l *Logger) Warn(msg string, fields ...interface{}) {
	l.sugar.Warnw(msg, fields...)
}

func (l *Logger) Error(msg string, fields ...interface{}) {
	l.sugar.Errorw(msg, fields...)
}

func (l *Logger) Fatal(msg string, fields ...interface{}) {
	l.sugar.Fatalw(msg, fields...)
}

func (l *Logger) Debugf(template string, args ...interface{}) {
	l.sugar.Debugf(template, args...)
}

func (l *Logger) Infof(template string, args ...interface{}) {
	l.sugar.Infof(template, args...)
}

func (l *Logger) Warnf(template string, args ...interface{}) {
	l.sugar.Warnf(template, args...)
}

func (l *Logger) Errorf(template string, args ...interface{}) {
	l.sugar.Errorf(template, args...)
}

func (l *Logger) Fatalf(template string, args ...interface{}) {
	l.sugar.Fatalf(template, args...)
}

func (l *Logger) Sync() error {
	return l.Logger.Sync()
}

func (l *Logger) WithFields(fields map[string]interface{}) *zap.SugaredLogger {
	return l.sugar.With(fields)
}

func (l *Logger) Named(name string) *Logger {
	return &Logger{
		Logger: l.Logger.Named(name),
		sugar:  l.sugar.Named(name),
		config: l.config,
	}
}

type contextKey string

const (
	TraceIDKey   contextKey = "trace_id"
	UserIDKey    contextKey = "user_id"
	TenantIDKey  contextKey = "tenant_id"
	RequestIDKey contextKey = "request_id"
)

func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIDKey, traceID)
}

func GetTraceID(ctx context.Context) string {
	if v := ctx.Value(TraceIDKey); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

func GetUserID(ctx context.Context) string {
	if v := ctx.Value(UserIDKey); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func WithTenantID(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, TenantIDKey, tenantID)
}

func GetTenantID(ctx context.Context) string {
	if v := ctx.Value(TenantIDKey); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

func GetRequestID(ctx context.Context) string {
	if v := ctx.Value(RequestIDKey); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func (l *Logger) New(ctx context.Context) *zap.SugaredLogger {
	fields := make([]interface{}, 0)

	if traceID := GetTraceID(ctx); traceID != "" {
		fields = append(fields, "trace_id", traceID)
	}
	if userID := GetUserID(ctx); userID != "" {
		fields = append(fields, "user_id", userID)
	}
	if tenantID := GetTenantID(ctx); tenantID != "" {
		fields = append(fields, "tenant_id", tenantID)
	}
	if requestID := GetRequestID(ctx); requestID != "" {
		fields = append(fields, "request_id", requestID)
	}

	if len(fields) > 0 {
		return l.sugar.With(fields...)
	}
	return l.sugar
}

func (l *Logger) Log(level zapcore.Level, msg string, fields map[string]interface{}) {
	switch level {
	case zapcore.DebugLevel:
		l.sugar.Debugw(msg, toSlice(fields)...)
	case zapcore.InfoLevel:
		l.sugar.Infow(msg, toSlice(fields)...)
	case zapcore.WarnLevel:
		l.sugar.Warnw(msg, toSlice(fields)...)
	case zapcore.ErrorLevel:
		l.sugar.Errorw(msg, toSlice(fields)...)
	case zapcore.FatalLevel:
		l.sugar.Fatalw(msg, toSlice(fields)...)
	}
}

func toSlice(fields map[string]interface{}) []interface{} {
	s := make([]interface{}, 0, len(fields)*2)
	for k, v := range fields {
		s = append(s, k, v)
	}
	return s
}

type Timer struct {
	start   time.Time
	logger  *Logger
	message string
	fields  map[string]interface{}
}

func (l *Logger) StartTimer(ctx context.Context, message string, fields map[string]interface{}) *Timer {
	return &Timer{
		start:   time.Now(),
		logger:  l,
		message: message,
		fields:  fields,
	}
}

func (t *Timer) Stop(ctx context.Context) {
	elapsed := time.Since(t.start)
	fields := t.fields
	if fields == nil {
		fields = make(map[string]interface{})
	}
	fields["elapsed_ms"] = elapsed.Milliseconds()
	t.logger.Log(zapcore.InfoLevel, t.message, fields)
}
