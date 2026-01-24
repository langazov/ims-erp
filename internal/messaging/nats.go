package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/ims-erp/system/internal/commands"
	"github.com/ims-erp/system/internal/events"
	"github.com/ims-erp/system/pkg/logger"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type NATSConfig struct {
	URLs           []string
	Username       string
	Password       string
	Token          string
	MaxReconnect   int
	ReconnectWait  time.Duration
	ConnectTimeout time.Duration
	JetStream      bool
	Domain         string
	StreamPrefix   string
}

type Publisher struct {
	conn   *nats.Conn
	js     jetstream.JetStream
	config NATSConfig
	logger *logger.Logger
	mu     sync.RWMutex
}

func NewPublisher(config NATSConfig, log *logger.Logger) (*Publisher, error) {
	opts := []nats.Option{
		nats.MaxReconnects(config.MaxReconnect),
		nats.ReconnectWait(config.ReconnectWait),
		nats.ErrorHandler(func(nc *nats.Conn, s *nats.Subscription, err error) {
			log.Error("NATS error", "error", err.Error(), "subject", s.Subject)
		}),
	}

	if config.Username != "" && config.Password != "" {
		opts = append(opts, nats.UserInfo(config.Username, config.Password))
	}
	if config.Token != "" {
		opts = append(opts, nats.Token(config.Token))
	}

	conn, err := nats.Connect(config.URLs[0], opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	var js jetstream.JetStream
	if config.JetStream {
		js, err = jetstream.New(conn)
		if err != nil {
			conn.Close()
			return nil, fmt.Errorf("failed to create JetStream: %w", err)
		}
	}

	return &Publisher{
		conn:   conn,
		js:     js,
		config: config,
		logger: log,
	}, nil
}

func (p *Publisher) PublishEvent(ctx context.Context, event *events.EventEnvelope) error {
	tracer := otel.Tracer("messaging")
	ctx, span := tracer.Start(ctx, "nats.publish.event",
		trace.WithAttributes(
			attribute.String("event.type", event.Type),
			attribute.String("event.aggregate_id", event.AggregateID),
			attribute.String("event.tenant_id", event.TenantID),
		),
	)
	defer span.End()

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	subject := p.config.StreamPrefix + event.Subject()
	msg := nats.NewMsg(subject)
	msg.Data = data
	msg.Header.Set("event-type", event.Type)
	msg.Header.Set("aggregate-id", event.AggregateID)
	msg.Header.Set("aggregate-type", event.AggregateType)
	msg.Header.Set("tenant-id", event.TenantID)
	msg.Header.Set("user-id", event.UserID)
	msg.Header.Set("trace-id", span.SpanContext().TraceID().String())

	if p.js == nil {
		if err := p.conn.PublishMsg(msg); err != nil {
			span.RecordError(err)
			return fmt.Errorf("failed to publish event: %w", err)
		}
	} else {
		_, err := p.js.Publish(ctx, subject, data)
		if err != nil {
			span.RecordError(err)
			return fmt.Errorf("failed to publish event to JetStream: %w", err)
		}
	}

	p.logger.New(ctx).Debug("Published event",
		"event_type", event.Type,
		"aggregate_id", event.AggregateID,
		"subject", subject,
	)

	return nil
}

func (p *Publisher) PublishCommand(ctx context.Context, cmd *commands.CommandEnvelope) error {
	tracer := otel.Tracer("messaging")
	ctx, span := tracer.Start(ctx, "nats.publish.command",
		trace.WithAttributes(
			attribute.String("command.type", cmd.Type),
			attribute.String("command.target_id", cmd.TargetID),
			attribute.String("command.tenant_id", cmd.TenantID),
		),
	)
	defer span.End()

	data, err := json.Marshal(cmd)
	if err != nil {
		return fmt.Errorf("failed to marshal command: %w", err)
	}

	subject := p.config.StreamPrefix + cmd.Subject()
	msg := nats.NewMsg(subject)
	msg.Data = data
	msg.Header.Set("command-type", cmd.Type)
	msg.Header.Set("target-id", cmd.TargetID)
	msg.Header.Set("tenant-id", cmd.TenantID)
	msg.Header.Set("user-id", cmd.UserID)
	msg.Header.Set("trace-id", span.SpanContext().TraceID().String())

	if p.js == nil {
		if err := p.conn.PublishMsg(msg); err != nil {
			span.RecordError(err)
			return fmt.Errorf("failed to publish command: %w", err)
		}
	} else {
		_, err := p.js.Publish(ctx, subject, data)
		if err != nil {
			span.RecordError(err)
			return fmt.Errorf("failed to publish command to JetStream: %w", err)
		}
	}

	p.logger.New(ctx).Debug("Published command",
		"command_type", cmd.Type,
		"target_id", cmd.TargetID,
		"subject", subject,
	)

	return nil
}

func (p *Publisher) RequestReply(ctx context.Context, subject string, data []byte, timeout time.Duration) ([]byte, error) {
	msg := nats.NewMsg(subject)
	msg.Data = data
	msg.Header.Set("trace-id", trace.SpanFromContext(ctx).SpanContext().TraceID().String())

	resp, err := p.conn.RequestMsgWithContext(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp.Data, nil
}

func (p *Publisher) Close() error {
	p.conn.Close()
	return nil
}

func (p *Publisher) Connected() bool {
	return p.conn.IsConnected()
}

type Subscriber struct {
	conn     *nats.Conn
	js       jetstream.JetStream
	config   NATSConfig
	logger   *logger.Logger
	handlers map[string][]nats.MsgHandler
	mu       sync.RWMutex
	subs     []*nats.Subscription
}

func NewSubscriber(config NATSConfig, log *logger.Logger) (*Subscriber, error) {
	opts := []nats.Option{
		nats.MaxReconnects(config.MaxReconnect),
		nats.ReconnectWait(config.ReconnectWait),
	}

	if config.Username != "" && config.Password != "" {
		opts = append(opts, nats.UserInfo(config.Username, config.Password))
	}

	conn, err := nats.Connect(config.URLs[0], opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	var js jetstream.JetStream
	if config.JetStream {
		js, err = jetstream.New(conn)
		if err != nil {
			conn.Close()
			return nil, fmt.Errorf("failed to create JetStream: %w", err)
		}
	}

	return &Subscriber{
		conn:     conn,
		js:       js,
		config:   config,
		logger:   log,
		handlers: make(map[string][]nats.MsgHandler),
		subs:     make([]*nats.Subscription, 0),
	}, nil
}

func (s *Subscriber) Subscribe(subject string, handler nats.MsgHandler) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.handlers[subject] = append(s.handlers[subject], handler)

	sub, err := s.conn.Subscribe(subject, handler)
	if err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	s.subs = append(s.subs, sub)
	return nil
}

func (s *Subscriber) SubscribeQueue(subject, queue string, handler nats.MsgHandler) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.handlers[subject] = append(s.handlers[subject], handler)

	sub, err := s.conn.QueueSubscribe(subject, queue, handler)
	if err != nil {
		return fmt.Errorf("failed to subscribe to queue: %w", err)
	}

	s.subs = append(s.subs, sub)
	return nil
}

func (s *Subscriber) SubscribeJetStream(ctx context.Context, streamName, consumerName, subject string, handler jetstream.MessageHandler) error {
	stream, err := s.js.Stream(ctx, streamName)
	if err != nil {
		return fmt.Errorf("failed to get stream: %w", err)
	}

	consumerCfg := jetstream.ConsumerConfig{
		Name:          consumerName,
		Durable:       consumerName,
		FilterSubject: subject,
		AckPolicy:     jetstream.AckAllPolicy,
	}

	_, err = stream.CreateConsumer(ctx, consumerCfg)
	if err != nil {
		return fmt.Errorf("failed to create consumer: %w", err)
	}

	return nil
}

func (s *Subscriber) UnsubscribeAll() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, sub := range s.subs {
		sub.Unsubscribe()
	}
	s.subs = make([]*nats.Subscription, 0)
}

func (s *Subscriber) Close() error {
	s.UnsubscribeAll()
	s.conn.Close()
	return nil
}

func (s *Subscriber) Connected() bool {
	return s.conn.IsConnected()
}

func ExtractTraceID(headers nats.Header) string {
	return headers.Get("trace-id")
}

func InjectTraceID(ctx context.Context, headers nats.Header) {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().HasTraceID() {
		headers.Set("trace-id", span.SpanContext().TraceID().String())
	}
}

func SetupTracePropagation() {
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
}

type StreamConfig struct {
	Name      string
	Subjects  []string
	Retention jetstream.RetentionPolicy
	MaxAge    time.Duration
	MaxBytes  int64
	MaxMsgs   int64
	Storage   jetstream.StorageType
	Replicas  int
	Discard   jetstream.DiscardPolicy
}

func (p *Publisher) CreateStream(ctx context.Context, cfg StreamConfig) error {
	if p.js == nil {
		return fmt.Errorf("JetStream not enabled")
	}

	_, err := p.js.CreateStream(ctx, jetstream.StreamConfig{
		Name:      cfg.Name,
		Subjects:  cfg.Subjects,
		Retention: cfg.Retention,
		MaxAge:    cfg.MaxAge,
		MaxBytes:  cfg.MaxBytes,
		MaxMsgs:   cfg.MaxMsgs,
		Storage:   cfg.Storage,
		Replicas:  cfg.Replicas,
		Discard:   cfg.Discard,
	})
	return err
}

func (p *Publisher) CreateConsumer(ctx context.Context, streamName, consumerName string, cfg ConsumerConfig) error {
	if p.js == nil {
		return fmt.Errorf("JetStream not enabled")
	}

	stream, err := p.js.Stream(ctx, streamName)
	if err != nil {
		return fmt.Errorf("failed to get stream: %w", err)
	}

	consumerCfg := jetstream.ConsumerConfig{
		Name:          consumerName,
		Durable:       consumerName,
		FilterSubject: cfg.FilterSubject,
		AckPolicy:     jetstream.AckAllPolicy,
		MaxDeliver:    cfg.MaxDeliver,
		MaxAckPending: cfg.MaxAckPending,
		AckWait:       cfg.AckWait,
	}

	_, err = stream.CreateConsumer(ctx, consumerCfg)
	return err
}

type ConsumerConfig struct {
	FilterSubject string
	MaxDeliver    int
	MaxAckPending int
	AckWait       time.Duration
}
