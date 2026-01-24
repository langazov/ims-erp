package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	RequestsTotal      *prometheus.CounterVec
	RequestDuration    *prometheus.HistogramVec
	RequestsInFlight   prometheus.Gauge
	CacheHits          *prometheus.CounterVec
	CacheMisses        *prometheus.CounterVec
	DatabaseOperations *prometheus.CounterVec
	DatabaseDuration   *prometheus.HistogramVec
	NATSMessages       *prometheus.CounterVec
	NATSMsgDuration    *prometheus.HistogramVec
	ServiceHealth      *prometheus.GaugeVec
	ErrorsTotal        *prometheus.CounterVec
)

func Initialize(namespace string) {
	RequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request duration in seconds",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	RequestsInFlight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "http_requests_in_flight",
			Help:      "Current number of HTTP requests being processed",
		},
	)

	CacheHits = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "cache_hits_total",
			Help:      "Total number of cache hits",
		},
		[]string{"cache_type"},
	)

	CacheMisses = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "cache_misses_total",
			Help:      "Total number of cache misses",
		},
		[]string{"cache_type"},
	)

	DatabaseOperations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "db_operations_total",
			Help:      "Total number of database operations",
		},
		[]string{"operation", "collection"},
	)

	DatabaseDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "db_operation_duration_seconds",
			Help:      "Database operation duration in seconds",
			Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"operation", "collection"},
	)

	NATSMessages = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "nats_messages_total",
			Help:      "Total number of NATS messages",
		},
		[]string{"subject", "direction", "status"},
	)

	NATSMsgDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "nats_message_duration_seconds",
			Help:      "NATS message processing duration in seconds",
			Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"subject"},
	)

	ServiceHealth = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "service_health",
			Help:      "Service health status (1=healthy, 0=unhealthy)",
		},
		[]string{"component"},
	)

	ErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "errors_total",
			Help:      "Total number of errors",
		},
		[]string{"type", "component"},
	)
}

func RecordCacheHit(cacheType string) {
	CacheHits.WithLabelValues(cacheType).Inc()
}

func RecordCacheMiss(cacheType string) {
	CacheMisses.WithLabelValues(cacheType).Inc()
}

func RecordDBOperation(operation, collection string, duration float64) {
	DatabaseOperations.WithLabelValues(operation, collection).Inc()
	DatabaseDuration.WithLabelValues(operation, collection).Observe(duration)
}

func RecordNATSMessage(subject, direction string, status string, duration float64) {
	NATSMessages.WithLabelValues(subject, direction, status).Inc()
	NATSMsgDuration.WithLabelValues(subject).Observe(duration)
}

func RecordHTTPRequest(method, endpoint, status string, duration float64) {
	RequestsTotal.WithLabelValues(method, endpoint, status).Inc()
	RequestDuration.WithLabelValues(method, endpoint).Observe(duration)
}

func RecordError(errorType, component string) {
	ErrorsTotal.WithLabelValues(errorType, component).Inc()
}

func SetServiceHealth(component string, healthy bool) {
	var value float64
	if healthy {
		value = 1
	}
	ServiceHealth.WithLabelValues(component).Set(value)
}
