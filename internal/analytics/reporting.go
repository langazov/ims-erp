package analytics

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ims-erp/system/internal/events"
	"github.com/ims-erp/system/internal/repository"
	"github.com/ims-erp/system/pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// ReportingService provides BI analytics and reporting
type ReportingService struct {
	readModelStore *repository.ReadModelStore
	cache          *repository.Cache
	logger         *logger.Logger
	tracer         trace.Tracer
}

// NewReportingService creates a new reporting service
func NewReportingService(
	readModelStore *repository.ReadModelStore,
	cache *repository.Cache,
	logger *logger.Logger,
) *ReportingService {
	return &ReportingService{
		readModelStore: readModelStore,
		cache:          cache,
		logger:         logger,
		tracer:         otel.Tracer("reporting-service"),
	}
}

// RevenueSummary contains revenue analytics
type RevenueSummary struct {
	Period         string  `json:"period"`
	StartDate      string  `json:"startDate"`
	EndDate        string  `json:"endDate"`
	TotalRevenue   float64 `json:"totalRevenue"`
	InvoiceCount   int     `json:"invoiceCount"`
	AverageInvoice float64 `json:"averageInvoice"`
	PaidAmount     float64 `json:"paidAmount"`
	Outstanding    float64 `json:"outstanding"`
	OverdueAmount  float64 `json:"overdueAmount"`
}

// AgingBucket represents an aging category
type AgingBucket struct {
	Range        string  `json:"range"`
	InvoiceCount int     `json:"invoiceCount"`
	Amount       float64 `json:"amount"`
}

// AgingReport contains invoice aging analysis
type AgingReport struct {
	AsOfDate         time.Time     `json:"asOfDate"`
	TotalOutstanding float64       `json:"totalOutstanding"`
	Buckets          []AgingBucket `json:"buckets"`
}

// PaymentSummary contains payment analytics
type PaymentSummary struct {
	Period           string         `json:"period"`
	StartDate        string         `json:"startDate"`
	EndDate          string         `json:"endDate"`
	TotalPayments    int            `json:"totalPayments"`
	TotalVolume      float64        `json:"totalVolume"`
	SuccessRate      float64        `json:"successRate"`
	FailedCount      int            `json:"failedCount"`
	RefundedAmount   float64        `json:"refundedAmount"`
	MethodsBreakdown map[string]int `json:"methodsBreakdown"`
}

// DashboardData contains combined metrics for dashboard
type DashboardData struct {
	TenantID       string                  `json:"tenantId"`
	GeneratedAt    time.Time               `json:"generatedAt"`
	Revenue        RevenueSummary          `json:"revenue"`
	Aging          AgingReport             `json:"aging"`
	Payments       PaymentSummary          `json:"payments"`
	RecentInvoices []events.InvoiceSummary `json:"recentInvoices"`
	KeyMetrics     map[string]interface{}  `json:"keyMetrics"`
}

// GetRevenueSummary returns revenue analytics for a period
func (s *ReportingService) GetRevenueSummary(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (*RevenueSummary, error) {
	ctx, span := s.tracer.Start(ctx, "reporting.revenue_summary",
		trace.WithAttributes(
			attribute.String("tenant_id", tenantID.String()),
			attribute.String("start_date", startDate.Format(time.RFC3339)),
			attribute.String("end_date", endDate.Format(time.RFC3339)),
		),
	)
	defer span.End()

	// Check cache first
	cacheKey := fmt.Sprintf("report:revenue:%s:%s:%s", tenantID.String(), startDate.Format("20060102"), endDate.Format("20060102"))
	if cached, err := s.cache.Get(ctx, cacheKey); err == nil {
		var summary RevenueSummary
		if err := json.Unmarshal([]byte(cached), &summary); err == nil {
			return &summary, nil
		}
	}

	// Query read model for invoice data
	filter := bson.M{
		"tenantId": tenantID.String(),
		"issueDate": bson.M{
			"$gte": startDate,
			"$lte": endDate,
		},
	}

	results, err := s.readModelStore.Find(ctx, filter)
	if err != nil {
		s.logger.New(ctx).Error("Failed to query invoices for revenue summary", "error", err)
		return nil, fmt.Errorf("failed to generate revenue summary: %w", err)
	}

	var invoices []events.InvoiceSummary
	for _, r := range results {
		if doc, ok := r.(bson.M); ok {
			var inv events.InvoiceSummary
			if data, err := bson.Marshal(doc); err == nil {
				if err := bson.Unmarshal(data, &inv); err == nil {
					invoices = append(invoices, inv)
				}
			}
		}
	}

	summary := &RevenueSummary{
		Period:    fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
		StartDate: startDate.Format(time.RFC3339),
		EndDate:   endDate.Format(time.RFC3339),
	}

	for _, inv := range invoices {
		summary.InvoiceCount++
		// Parse total from string
		var total float64
		fmt.Sscanf(inv.Total, "%f", &total)
		summary.TotalRevenue += total

		var paid float64
		fmt.Sscanf(inv.AmountPaid, "%f", &paid)
		summary.PaidAmount += paid

		var due float64
		fmt.Sscanf(inv.AmountDue, "%f", &due)
		summary.Outstanding += due

		if inv.Status == "overdue" {
			summary.OverdueAmount += due
		}
	}

	if summary.InvoiceCount > 0 {
		summary.AverageInvoice = summary.TotalRevenue / float64(summary.InvoiceCount)
	}

	// Cache result
	if data, err := json.Marshal(summary); err == nil {
		s.cache.Set(ctx, cacheKey, string(data), 5*time.Minute)
	}

	return summary, nil
}

// GetAgingReport returns invoice aging analysis
func (s *ReportingService) GetAgingReport(ctx context.Context, tenantID uuid.UUID, asOfDate time.Time) (*AgingReport, error) {
	ctx, span := s.tracer.Start(ctx, "reporting.aging_report",
		trace.WithAttributes(
			attribute.String("tenant_id", tenantID.String()),
			attribute.String("as_of_date", asOfDate.Format(time.RFC3339)),
		),
	)
	defer span.End()

	// Check cache
	cacheKey := fmt.Sprintf("report:aging:%s:%s", tenantID.String(), asOfDate.Format("20060102"))
	if cached, err := s.cache.Get(ctx, cacheKey); err == nil {
		var report AgingReport
		if err := json.Unmarshal([]byte(cached), &report); err == nil {
			return &report, nil
		}
	}

	// Query outstanding invoices
	filter := bson.M{
		"tenantId": tenantID.String(),
		"status":   bson.M{"$in": []string{"sent", "partial", "overdue"}},
	}

	results, err := s.readModelStore.Find(ctx, filter)
	if err != nil {
		s.logger.New(ctx).Error("Failed to query invoices for aging report", "error", err)
		return nil, fmt.Errorf("failed to generate aging report: %w", err)
	}

	var invoices []events.InvoiceSummary
	for _, r := range results {
		if doc, ok := r.(bson.M); ok {
			var inv events.InvoiceSummary
			if data, err := bson.Marshal(doc); err == nil {
				if err := bson.Unmarshal(data, &inv); err == nil {
					invoices = append(invoices, inv)
				}
			}
		}
	}

	report := &AgingReport{
		AsOfDate: asOfDate,
		Buckets: []AgingBucket{
			{Range: "Current", InvoiceCount: 0, Amount: 0},
			{Range: "1-30 days", InvoiceCount: 0, Amount: 0},
			{Range: "31-60 days", InvoiceCount: 0, Amount: 0},
			{Range: "61-90 days", InvoiceCount: 0, Amount: 0},
			{Range: "90+ days", InvoiceCount: 0, Amount: 0},
		},
	}

	for _, inv := range invoices {
		var amountDue float64
		fmt.Sscanf(inv.AmountDue, "%f", &amountDue)

		daysOverdue := int(asOfDate.Sub(inv.DueDate).Hours() / 24)

		switch {
		case daysOverdue <= 0:
			report.Buckets[0].InvoiceCount++
			report.Buckets[0].Amount += amountDue
		case daysOverdue <= 30:
			report.Buckets[1].InvoiceCount++
			report.Buckets[1].Amount += amountDue
		case daysOverdue <= 60:
			report.Buckets[2].InvoiceCount++
			report.Buckets[2].Amount += amountDue
		case daysOverdue <= 90:
			report.Buckets[3].InvoiceCount++
			report.Buckets[3].Amount += amountDue
		default:
			report.Buckets[4].InvoiceCount++
			report.Buckets[4].Amount += amountDue
		}

		report.TotalOutstanding += amountDue
	}

	// Cache result
	if data, err := json.Marshal(report); err == nil {
		s.cache.Set(ctx, cacheKey, string(data), 5*time.Minute)
	}

	return report, nil
}

// GetPaymentSummary returns payment analytics
func (s *ReportingService) GetPaymentSummary(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (*PaymentSummary, error) {
	ctx, span := s.tracer.Start(ctx, "reporting.payment_summary",
		trace.WithAttributes(
			attribute.String("tenant_id", tenantID.String()),
		),
	)
	defer span.End()

	// Check cache
	cacheKey := fmt.Sprintf("report:payments:%s:%s:%s", tenantID.String(), startDate.Format("20060102"), endDate.Format("20060102"))
	if cached, err := s.cache.Get(ctx, cacheKey); err == nil {
		var summary PaymentSummary
		if err := json.Unmarshal([]byte(cached), &summary); err == nil {
			return &summary, nil
		}
	}

	// Query payments
	filter := bson.M{
		"tenantId": tenantID.String(),
		"createdAt": bson.M{
			"$gte": startDate,
			"$lte": endDate,
		},
	}

	results, err := s.readModelStore.Find(ctx, filter)
	if err != nil {
		s.logger.New(ctx).Error("Failed to query payments for summary", "error", err)
		return nil, fmt.Errorf("failed to generate payment summary: %w", err)
	}

	var payments []events.PaymentSummary
	for _, r := range results {
		if doc, ok := r.(bson.M); ok {
			var p events.PaymentSummary
			if data, err := bson.Marshal(doc); err == nil {
				if err := bson.Unmarshal(data, &p); err == nil {
					payments = append(payments, p)
				}
			}
		}
	}

	summary := &PaymentSummary{
		Period:           fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
		StartDate:        startDate.Format(time.RFC3339),
		EndDate:          endDate.Format(time.RFC3339),
		MethodsBreakdown: make(map[string]int),
	}

	for _, p := range payments {
		summary.TotalPayments++
		var amount float64
		fmt.Sscanf(p.Amount, "%f", &amount)
		summary.TotalVolume += amount

		summary.MethodsBreakdown[p.Method]++

		switch p.Status {
		case "completed":
			// success
		case "failed":
			summary.FailedCount++
		case "refunded":
			summary.RefundedAmount += amount
		}
	}

	if summary.TotalPayments > 0 {
		summary.SuccessRate = float64(summary.TotalPayments-summary.FailedCount) / float64(summary.TotalPayments) * 100
	}

	// Cache result
	if data, err := json.Marshal(summary); err == nil {
		s.cache.Set(ctx, cacheKey, string(data), 5*time.Minute)
	}

	return summary, nil
}

// GetDashboardData returns combined metrics for dashboard
func (s *ReportingService) GetDashboardData(ctx context.Context, tenantID uuid.UUID) (*DashboardData, error) {
	ctx, span := s.tracer.Start(ctx, "reporting.dashboard",
		trace.WithAttributes(attribute.String("tenant_id", tenantID.String())),
	)
	defer span.End()

	// Check cache
	cacheKey := fmt.Sprintf("report:dashboard:%s", tenantID.String())
	if cached, err := s.cache.Get(ctx, cacheKey); err == nil {
		var data DashboardData
		if err := json.Unmarshal([]byte(cached), &data); err == nil {
			return &data, nil
		}
	}

	now := time.Now().UTC()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	// Get revenue summary for current month
	revenue, err := s.GetRevenueSummary(ctx, tenantID, startOfMonth, now)
	if err != nil {
		s.logger.New(ctx).Error("Failed to get revenue summary for dashboard", "error", err)
	}

	// Get aging report
	aging, err := s.GetAgingReport(ctx, tenantID, now)
	if err != nil {
		s.logger.New(ctx).Error("Failed to get aging report for dashboard", "error", err)
	}

	// Get payment summary for current month
	payments, err := s.GetPaymentSummary(ctx, tenantID, startOfMonth, now)
	if err != nil {
		s.logger.New(ctx).Error("Failed to get payment summary for dashboard", "error", err)
	}

	// Get recent invoices
	var recentInvoices []events.InvoiceSummary
	filter := bson.M{"tenantId": tenantID.String()}
	opts := options.Find().SetSort(bson.M{"createdAt": -1}).SetLimit(5)
	results, err := s.readModelStore.Find(ctx, filter, opts)
	if err != nil {
		s.logger.New(ctx).Error("Failed to get recent invoices for dashboard", "error", err)
	} else {
		for _, r := range results {
			if doc, ok := r.(bson.M); ok {
				var inv events.InvoiceSummary
				if data, err := bson.Marshal(doc); err == nil {
					if err := bson.Unmarshal(data, &inv); err == nil {
						recentInvoices = append(recentInvoices, inv)
					}
				}
			}
		}
	}

	dashboard := &DashboardData{
		TenantID:       tenantID.String(),
		GeneratedAt:    now,
		Revenue:        *revenue,
		Aging:          *aging,
		Payments:       *payments,
		RecentInvoices: recentInvoices,
		KeyMetrics: map[string]interface{}{
			"collectionRate":        0.0,
			"averageCollectionDays": 0,
			"outstandingInvoices":   0,
		},
	}

	// Calculate collection rate
	if revenue.TotalRevenue > 0 {
		dashboard.KeyMetrics["collectionRate"] = (revenue.PaidAmount / revenue.TotalRevenue) * 100
	}

	// Count outstanding invoices
	for _, inv := range recentInvoices {
		if inv.Status == "sent" || inv.Status == "partial" || inv.Status == "overdue" {
			count := dashboard.KeyMetrics["outstandingInvoices"].(int)
			dashboard.KeyMetrics["outstandingInvoices"] = count + 1
		}
	}

	// Cache result for 1 minute
	if data, err := json.Marshal(dashboard); err == nil {
		s.cache.Set(ctx, cacheKey, string(data), 1*time.Minute)
	}

	return dashboard, nil
}
