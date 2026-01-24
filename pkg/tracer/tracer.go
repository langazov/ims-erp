package tracer

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

type Config struct {
	Enabled      bool
	ServiceName  string
	ExporterType string
	Endpoint     string
	SamplerType  string
	SamplerRatio float64
	SampleParam  string
}

type Tracer struct {
	provider *sdktrace.TracerProvider
	tracer   trace.Tracer
	config   Config
}

func New(cfg Config) (*Tracer, error) {
	if !cfg.Enabled {
		return &Tracer{
			tracer: otel.Tracer(cfg.ServiceName),
			config: cfg,
		}, nil
	}

	ctx := context.Background()

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion("1.0.0"),
			attribute.String("environment", "production"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	var exporter sdktrace.SpanExporter
	switch cfg.ExporterType {
	case "otlp":
		exp, err := otlptracegrpc.New(ctx,
			otlptracegrpc.WithEndpoint(cfg.Endpoint),
			otlptracegrpc.WithInsecure(),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
		}
		exporter = exp
	case "stdout":
		exp, err := stdouttrace.New(
			stdouttrace.WithPrettyPrint(),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create stdout exporter: %w", err)
		}
		exporter = exp
	default:
		exp, err := stdouttrace.New(
			stdouttrace.WithPrettyPrint(),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create stdout exporter: %w", err)
		}
		exporter = exp
	}

	var sampler sdktrace.Sampler
	switch cfg.SamplerType {
	case "always":
		sampler = sdktrace.AlwaysSample()
	case "never":
		sampler = sdktrace.NeverSample()
	case "ratio":
		sampler = sdktrace.TraceIDRatioBased(cfg.SamplerRatio)
	case "parent":
		sampler = sdktrace.ParentBased(sdktrace.AlwaysSample())
	default:
		sampler = sdktrace.AlwaysSample()
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return &Tracer{
		provider: tp,
		tracer:   tp.Tracer(cfg.ServiceName),
		config:   cfg,
	}, nil
}

func (t *Tracer) Tracer() trace.Tracer {
	return t.tracer
}

func (t *Tracer) Shutdown(ctx context.Context) error {
	if t.provider != nil {
		return t.provider.Shutdown(ctx)
	}
	return nil
}

func (t *Tracer) Start(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return t.tracer.Start(ctx, name, opts...)
}

func (t *Tracer) AddEvent(ctx context.Context, name string, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	span.AddEvent(name, trace.WithAttributes(attrs...))
}

func (t *Tracer) SetError(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	span.RecordError(err)
}

func (t *Tracer) SetAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attrs...)
}

type Span struct {
	span    trace.Span
	tracer  *Tracer
	started bool
}

func (t *Tracer) Span(ctx context.Context, name string) *Span {
	ctx, span := t.tracer.Start(ctx, name)
	return &Span{
		span:    span,
		tracer:  t,
		started: true,
	}
}

func (s *Span) End() {
	if s.started {
		s.span.End()
		s.started = false
	}
}

func (s *Span) AddEvent(name string, attrs ...attribute.KeyValue) {
	s.span.AddEvent(name, trace.WithAttributes(attrs...))
}

func (s *Span) SetError(err error) {
	s.span.RecordError(err)
}

func (s *Span) SetAttributes(attrs ...attribute.KeyValue) {
	s.span.SetAttributes(attrs...)
}

func (s *Span) Context() context.Context {
	return context.Background()
}

func ExtractTraceID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().HasTraceID() {
		return span.SpanContext().TraceID().String()
	}
	return ""
}

func ExtractSpanID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().HasSpanID() {
		return span.SpanContext().SpanID().String()
	}
	return ""
}
