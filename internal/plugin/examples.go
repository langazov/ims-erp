package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ims-erp/system/internal/events"
)

// SlackNotificationPlugin sends notifications to Slack
type SlackNotificationPlugin struct {
	*BasePlugin
	webhookURL string
	httpClient *http.Client
}

// NewSlackNotificationPlugin creates a new Slack notification plugin
func NewSlackNotificationPlugin(webhookURL string) *SlackNotificationPlugin {
	plugin := &SlackNotificationPlugin{
		BasePlugin: NewBasePlugin("slack-notifications", "1.0.0", "Send notifications to Slack", "ERP System"),
		webhookURL: webhookURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}

	plugin.SetEventTypes([]string{
		"client.created",
		"invoice.paid",
		"payment.received",
		"order.fulfilled",
	})

	return plugin
}

func (p *SlackNotificationPlugin) HandleEvent(ctx context.Context, event events.EventEnvelope) error {
	if !p.CanHandleEvent(event.Type) {
		return nil
	}

	message := p.formatMessage(event)

	return p.sendToSlack(message)
}

func (p *SlackNotificationPlugin) formatMessage(event events.EventEnvelope) map[string]interface{} {
	var text string

	switch event.Type {
	case "client.created":
		name, _ := event.Data["name"].(string)
		email, _ := event.Data["email"].(string)
		text = fmt.Sprintf("New client created: %s (%s)", name, email)
	case "invoice.paid":
		invNumber, _ := event.Data["invoiceNumber"].(string)
		amount, _ := event.Data["amount"].(string)
		text = fmt.Sprintf("Invoice %s paid: %s", invNumber, amount)
	case "payment.received":
		amount, _ := event.Data["amount"].(string)
		method, _ := event.Data["method"].(string)
		text = fmt.Sprintf("Payment received: %s via %s", amount, method)
	case "order.fulfilled":
		orderID, _ := event.Data["orderId"].(string)
		text = fmt.Sprintf("Order %s has been fulfilled", orderID)
	default:
		text = fmt.Sprintf("Event: %s", event.Type)
	}

	return map[string]interface{}{
		"text": text,
		"attachments": []map[string]interface{}{
			{
				"color": "good",
				"fields": []map[string]interface{}{
					{
						"title": "Event Type",
						"value": event.Type,
						"short": true,
					},
					{
						"title": "Tenant",
						"value": event.TenantID,
						"short": true,
					},
					{
						"title": "Timestamp",
						"value": event.Timestamp.Format(time.RFC3339),
						"short": true,
					},
				},
			},
		},
	}
}

func (p *SlackNotificationPlugin) sendToSlack(message map[string]interface{}) error {
	if p.webhookURL == "" {
		return fmt.Errorf("no webhook URL configured")
	}

	payload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	req, err := http.NewRequest("POST", p.webhookURL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send to Slack: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Slack returned status: %d", resp.StatusCode)
	}

	return nil
}

// DataExportPlugin exports data on a schedule
type DataExportPlugin struct {
	*BasePlugin
	schedule string
	timezone string
}

// NewDataExportPlugin creates a new data export plugin
func NewDataExportPlugin(schedule string) *DataExportPlugin {
	return &DataExportPlugin{
		BasePlugin: NewBasePlugin("data-export", "1.0.0", "Export data to external systems", "ERP System"),
		schedule:   schedule,
		timezone:   "UTC",
	}
}

func (p *DataExportPlugin) Schedule() string {
	return p.schedule
}

func (p *DataExportPlugin) TimeZone() string {
	return p.timezone
}

func (p *DataExportPlugin) Run(ctx context.Context) error {
	sdk := p.GetSDK()
	logger := sdk.Logger()

	logger.Info("Starting data export", "plugin", p.Name())

	// Simulate export
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(5 * time.Second):
		logger.Info("Data export completed", "plugin", p.Name())
		return nil
	}
}

// CustomReportPlugin provides custom reporting API
type CustomReportPlugin struct {
	*BasePlugin
	basePath string
}

// NewCustomReportPlugin creates a new custom report plugin
func NewCustomReportPlugin() *CustomReportPlugin {
	plugin := &CustomReportPlugin{
		BasePlugin: NewBasePlugin("custom-reports", "1.0.0", "Generate custom reports", "ERP System"),
		basePath:   "/api/v1/plugins/reports",
	}

	// Add routes
	plugin.AddRoute(Route{
		Path:    "/sales",
		Method:  "GET",
		Handler: plugin.handleSalesReport,
	})

	plugin.AddRoute(Route{
		Path:    "/inventory",
		Method:  "GET",
		Handler: plugin.handleInventoryReport,
	})

	plugin.AddRoute(Route{
		Path:    "/export",
		Method:  "POST",
		Handler: plugin.handleExportReport,
	})

	return plugin
}

func (p *CustomReportPlugin) BasePath() string {
	return p.basePath
}

func (p *CustomReportPlugin) handleSalesReport(w http.ResponseWriter, r *http.Request) {
	report := map[string]interface{}{
		"report": "sales",
		"total":  123456.78,
		"count":  150,
		"date":   time.Now().Format("2006-01-02"),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

func (p *CustomReportPlugin) handleInventoryReport(w http.ResponseWriter, r *http.Request) {
	report := map[string]interface{}{
		"report":          "inventory",
		"totalItems":      5000,
		"lowStockCount":   12,
		"outOfStockCount": 3,
		"date":            time.Now().Format("2006-01-02"),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

func (p *CustomReportPlugin) handleExportReport(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ReportType string `json:"reportType"`
		Format     string `json:"format"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"status":   "queued",
		"reportId": generatePluginID("report"),
		"format":   request.Format,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(response)
}
