package analytics

import (
	"context"
	"time"

	"github.com/ims-erp/system/internal/events"
)

// EventAggregator aggregates events for analytics
type EventAggregator struct {
	buckets   map[string]*EventBucket
	interval  time.Duration
	processor EventProcessor
}

// EventBucket holds aggregated events for a time period
type EventBucket struct {
	StartTime time.Time
	EndTime   time.Time
	Events    []events.EventEnvelope
	Counts    map[string]int
}

// EventProcessor processes aggregated events
type EventProcessor interface {
	Process(bucket *EventBucket) error
}

// NewEventAggregator creates a new event aggregator
func NewEventAggregator(interval time.Duration, processor EventProcessor) *EventAggregator {
	return &EventAggregator{
		buckets:   make(map[string]*EventBucket),
		interval:  interval,
		processor: processor,
	}
}

// Start begins the aggregation process
func (ea *EventAggregator) Start(ctx context.Context) {
	ticker := time.NewTicker(ea.interval)

	go func() {
		for {
			select {
			case <-ticker.C:
				ea.flushBuckets(ctx)
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}

// AddEvent adds an event to the aggregator
func (ea *EventAggregator) AddEvent(event events.EventEnvelope) {
	bucketID := ea.getBucketID(event.Timestamp)

	if _, exists := ea.buckets[bucketID]; !exists {
		ea.buckets[bucketID] = &EventBucket{
			StartTime: event.Timestamp,
			EndTime:   event.Timestamp.Add(ea.interval),
			Events:    []events.EventEnvelope{},
			Counts:    make(map[string]int),
		}
	}

	bucket := ea.buckets[bucketID]
	bucket.Events = append(bucket.Events, event)
	bucket.Counts[event.Type]++
}

// flushBuckets processes and clears full buckets
func (ea *EventAggregator) flushBuckets(ctx context.Context) {
	now := time.Now()

	for id, bucket := range ea.buckets {
		if now.After(bucket.EndTime) {
			// Process bucket
			if err := ea.processor.Process(bucket); err != nil {
				// Log error
			}

			// Remove processed bucket
			delete(ea.buckets, id)
		}
	}
}

// getBucketID generates a bucket ID from timestamp
func (ea *EventAggregator) getBucketID(t time.Time) string {
	return t.Truncate(ea.interval).Format("200601021504")
}

// TimeSeriesStore stores time-series metrics
type TimeSeriesStore struct {
	metrics map[string][]MetricPoint
}

// MetricPoint represents a single metric value
type MetricPoint struct {
	Timestamp time.Time
	Value     float64
	Labels    map[string]string
}

// NewTimeSeriesStore creates a new time series store
func NewTimeSeriesStore() *TimeSeriesStore {
	return &TimeSeriesStore{
		metrics: make(map[string][]MetricPoint),
	}
}

// Write adds a metric point
func (ts *TimeSeriesStore) Write(metricName string, value float64, labels map[string]string) {
	point := MetricPoint{
		Timestamp: time.Now(),
		Value:     value,
		Labels:    labels,
	}

	ts.metrics[metricName] = append(ts.metrics[metricName], point)

	// Keep only last 1000 points per metric (in production, use proper retention)
	if len(ts.metrics[metricName]) > 1000 {
		ts.metrics[metricName] = ts.metrics[metricName][len(ts.metrics[metricName])-1000:]
	}
}

// Query queries metrics by name and time range
func (ts *TimeSeriesStore) Query(metricName string, from, to time.Time) []MetricPoint {
	points, exists := ts.metrics[metricName]
	if !exists {
		return nil
	}

	var result []MetricPoint
	for _, p := range points {
		if p.Timestamp.After(from) && p.Timestamp.Before(to) {
			result = append(result, p)
		}
	}

	return result
}

// Dashboard represents a BI dashboard
type Dashboard struct {
	ID          string
	Name        string
	Description string
	Widgets     []Widget
	TenantID    string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Widget represents a dashboard widget
type Widget struct {
	ID       string
	Type     string // chart, metric, table, etc.
	Title    string
	Query    string // Query definition
	Config   WidgetConfig
	Position Position
}

// WidgetConfig holds widget configuration
type WidgetConfig struct {
	ChartType  string // line, bar, pie, etc.
	TimeRange  string // 1h, 24h, 7d, 30d, etc.
	Dimensions map[string]string
}

// Position defines widget position on dashboard
type Position struct {
	X int
	Y int
	W int
	H int
}

// Report represents a BI report
type Report struct {
	ID          string
	Name        string
	Description string
	Type        string // scheduled, on-demand
	Query       ReportQuery
	Schedule    string   // cron expression for scheduled reports
	Recipients  []string // email addresses
	TenantID    string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ReportQuery defines report data query
type ReportQuery struct {
	DataSource string
	Metrics    []string
	Filters    map[string]interface{}
	GroupBy    []string
	TimeRange  TimeRange
}

// TimeRange defines a time range
type TimeRange struct {
	From time.Time
	To   time.Time
}

// KPI represents a Key Performance Indicator
type KPI struct {
	ID          string
	Name        string
	Description string
	Formula     string
	Target      float64
	Unit        string
	Current     float64
	Trend       TrendDirection
	TenantID    string
}

// TrendDirection indicates KPI trend
type TrendDirection string

const (
	TrendUp     TrendDirection = "up"
	TrendDown   TrendDirection = "down"
	TrendStable TrendDirection = "stable"
)

// AnalyticsEngine processes analytics queries
type AnalyticsEngine struct {
	tsStore *TimeSeriesStore
}

// NewAnalyticsEngine creates a new analytics engine
func NewAnalyticsEngine(tsStore *TimeSeriesStore) *AnalyticsEngine {
	return &AnalyticsEngine{
		tsStore: tsStore,
	}
}

// ExecuteQuery executes an analytics query
func (ae *AnalyticsEngine) ExecuteQuery(ctx context.Context, query ReportQuery) ([]map[string]interface{}, error) {
	// In a real implementation, this would query the database
	// For now, return mock data
	return []map[string]interface{}{
		{
			"metric": "revenue",
			"value":  123456.78,
			"period": "2024-01",
		},
	}, nil
}

// CalculateKPI calculates a KPI value
func (ae *AnalyticsEngine) CalculateKPI(ctx context.Context, kpi KPI) (float64, error) {
	// In a real implementation, this would evaluate the KPI formula
	// For now, return mock value
	return 85.5, nil
}

// GenerateDashboard generates a dashboard from configuration
func (ae *AnalyticsEngine) GenerateDashboard(ctx context.Context, dashboard Dashboard) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for _, widget := range dashboard.Widgets {
		data, err := ae.ExecuteQuery(ctx, ReportQuery{
			Metrics: []string{widget.Query},
		})
		if err != nil {
			continue
		}
		result[widget.ID] = data
	}

	return result, nil
}

// DefaultDashboards returns default dashboard configurations
func DefaultDashboards() []Dashboard {
	return []Dashboard{
		{
			ID:          "executive-summary",
			Name:        "Executive Summary",
			Description: "High-level business metrics",
			Widgets: []Widget{
				{
					ID:       "total-revenue",
					Type:     "metric",
					Title:    "Total Revenue",
					Query:    "sum(invoice.total) WHERE status = 'paid'",
					Position: Position{X: 0, Y: 0, W: 3, H: 2},
				},
				{
					ID:       "active-clients",
					Type:     "metric",
					Title:    "Active Clients",
					Query:    "count(client) WHERE status = 'active'",
					Position: Position{X: 3, Y: 0, W: 3, H: 2},
				},
				{
					ID:       "revenue-chart",
					Type:     "chart",
					Title:    "Revenue Trend",
					Query:    "sum(invoice.total) GROUP BY month",
					Config:   WidgetConfig{ChartType: "line", TimeRange: "12m"},
					Position: Position{X: 0, Y: 2, W: 6, H: 4},
				},
			},
		},
		{
			ID:          "inventory-dashboard",
			Name:        "Inventory Dashboard",
			Description: "Inventory and warehouse metrics",
			Widgets: []Widget{
				{
					ID:       "low-stock-items",
					Type:     "table",
					Title:    "Low Stock Items",
					Query:    "inventory WHERE quantity <= reorderPoint",
					Position: Position{X: 0, Y: 0, W: 6, H: 4},
				},
				{
					ID:       "warehouse-utilization",
					Type:     "chart",
					Title:    "Warehouse Utilization",
					Query:    "avg(warehouse.utilization) GROUP BY warehouse",
					Config:   WidgetConfig{ChartType: "bar"},
					Position: Position{X: 6, Y: 0, W: 6, H: 4},
				},
			},
		},
	}
}
