package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

// LoadTestConfig holds configuration for load testing
type LoadTestConfig struct {
	BaseURL         string
	ConcurrentUsers int
	TestDuration    time.Duration
	RampUpTime      time.Duration
	TenantID        string
}

// LoadTestMetrics tracks test metrics
type LoadTestMetrics struct {
	TotalRequests      int64
	SuccessfulRequests int64
	FailedRequests     int64
	TotalLatency       int64
	MinLatency         int64
	MaxLatency         int64
	LatencyHistogram   map[string]int64 // P50, P95, P99
	mu                 sync.RWMutex
}

// LoadTestRunner executes load tests
type LoadTestRunner struct {
	config  LoadTestConfig
	metrics *LoadTestMetrics
	client  *http.Client
	stopCh  chan struct{}
}

// NewLoadTestRunner creates a new load test runner
func NewLoadTestRunner(config LoadTestConfig) *LoadTestRunner {
	return &LoadTestRunner{
		config: config,
		metrics: &LoadTestMetrics{
			MinLatency:       999999999,
			LatencyHistogram: make(map[string]int64),
		},
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        1000,
				MaxIdleConnsPerHost: 100,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		stopCh: make(chan struct{}),
	}
}

// Run executes the load test
func (r *LoadTestRunner) Run(ctx context.Context) error {
	fmt.Printf("Starting Phase 2 Load Test\n")
	fmt.Printf("Configuration:\n")
	fmt.Printf("  - Base URL: %s\n", r.config.BaseURL)
	fmt.Printf("  - Concurrent Users: %d\n", r.config.ConcurrentUsers)
	fmt.Printf("  - Test Duration: %s\n", r.config.TestDuration)
	fmt.Printf("  - Ramp Up Time: %s\n", r.config.RampUpTime)
	fmt.Printf("\n")

	// Start metrics reporter
	go r.reportMetrics(ctx)

	// Create worker pool
	var wg sync.WaitGroup
	usersPerSecond := float64(r.config.ConcurrentUsers) / r.config.RampUpTime.Seconds()

	for i := 0; i < r.config.ConcurrentUsers; i++ {
		wg.Add(1)
		go r.worker(ctx, &wg, i)

		// Ramp up gradually
		if i > 0 && usersPerSecond > 0 {
			time.Sleep(time.Duration(float64(time.Second) / usersPerSecond))
		}
	}

	// Wait for test duration
	time.Sleep(r.config.TestDuration)

	// Signal workers to stop
	close(r.stopCh)

	// Wait for all workers to finish
	wg.Wait()

	// Print final results
	r.printResults()

	return nil
}

// worker simulates a single user
func (r *LoadTestRunner) worker(ctx context.Context, wg *sync.WaitGroup, id int) {
	defer wg.Done()

	userID := uuid.New().String()
	clientID := uuid.New().String()

	for {
		select {
		case <-r.stopCh:
			return
		case <-ctx.Done():
			return
		default:
			// Execute test scenario
			r.executeScenario(userID, clientID)
		}
	}
}

// executeScenario runs a random test scenario
func (r *LoadTestRunner) executeScenario(userID, clientID string) {
	scenarios := []func(string, string) error{
		r.scenarioCreateInvoice,
		r.scenarioAddLineItem,
		r.scenarioFinalizeInvoice,
		r.scenarioRecordPayment,
		r.scenarioListInvoices,
		r.scenarioGetInvoice,
		r.scenarioCreatePayment,
		r.scenarioProcessPayment,
	}

	// Weight scenarios (more reads than writes)
	weights := []int{15, 10, 8, 12, 25, 15, 8, 7}

	// Select scenario based on weights
	totalWeight := 0
	for _, w := range weights {
		totalWeight += w
	}

	random := rand.Intn(totalWeight)
	cumulative := 0
	selectedScenario := 0

	for i, w := range weights {
		cumulative += w
		if random < cumulative {
			selectedScenario = i
			break
		}
	}

	start := time.Now()
	err := scenarios[selectedScenario](userID, clientID)
	latency := time.Since(start).Milliseconds()

	// Record metrics
	atomic.AddInt64(&r.metrics.TotalRequests, 1)
	atomic.AddInt64(&r.metrics.TotalLatency, latency)

	if err != nil {
		atomic.AddInt64(&r.metrics.FailedRequests, 1)
	} else {
		atomic.AddInt64(&r.metrics.SuccessfulRequests, 1)
	}

	// Update min/max latency
	r.metrics.mu.Lock()
	if latency < r.metrics.MinLatency {
		r.metrics.MinLatency = latency
	}
	if latency > r.metrics.MaxLatency {
		r.metrics.MaxLatency = latency
	}
	r.metrics.mu.Unlock()

	// Small delay between requests
	time.Sleep(time.Duration(rand.Intn(100)+50) * time.Millisecond)
}

// scenarioCreateInvoice creates a new invoice
func (r *LoadTestRunner) scenarioCreateInvoice(userID, clientID string) error {
	payload := map[string]interface{}{
		"clientId":    clientID,
		"type":        "standard",
		"currency":    "USD",
		"paymentTerm": "net_30",
		"notes":       "Load test invoice",
	}

	return r.makeRequest("POST", "/api/v1/invoices", payload, userID)
}

// scenarioAddLineItem adds a line item to an invoice
func (r *LoadTestRunner) scenarioAddLineItem(userID, clientID string) error {
	invoiceID := uuid.New().String() // In real test, would use created invoice
	payload := map[string]interface{}{
		"description": "Test product",
		"quantity":    "5",
		"unitPrice":   "100.00",
		"taxRate":     "20",
	}

	return r.makeRequest("POST", fmt.Sprintf("/api/v1/invoices/%s/lines", invoiceID), payload, userID)
}

// scenarioFinalizeInvoice finalizes an invoice
func (r *LoadTestRunner) scenarioFinalizeInvoice(userID, clientID string) error {
	invoiceID := uuid.New().String()
	return r.makeRequest("PUT", fmt.Sprintf("/api/v1/invoices/%s/finalize", invoiceID), nil, userID)
}

// scenarioRecordPayment records a payment for an invoice
func (r *LoadTestRunner) scenarioRecordPayment(userID, clientID string) error {
	invoiceID := uuid.New().String()
	payload := map[string]interface{}{
		"amount":        "500.00",
		"paymentMethod": "credit_card",
		"reference":     "load-test-001",
	}

	return r.makeRequest("POST", fmt.Sprintf("/api/v1/invoices/%s/payments", invoiceID), payload, userID)
}

// scenarioListInvoices lists invoices
func (r *LoadTestRunner) scenarioListInvoices(userID, clientID string) error {
	return r.makeRequest("GET", fmt.Sprintf("/api/v1/invoices?tenantId=%s&page=1&pageSize=50", r.config.TenantID), nil, userID)
}

// scenarioGetInvoice gets a single invoice
func (r *LoadTestRunner) scenarioGetInvoice(userID, clientID string) error {
	invoiceID := uuid.New().String()
	return r.makeRequest("GET", fmt.Sprintf("/api/v1/invoices/%s", invoiceID), nil, userID)
}

// scenarioCreatePayment creates a payment
func (r *LoadTestRunner) scenarioCreatePayment(userID, clientID string) error {
	payload := map[string]interface{}{
		"invoiceId":   uuid.New().String(),
		"clientId":    clientID,
		"amount":      "500.00",
		"currency":    "USD",
		"method":      "credit_card",
		"provider":    "stripe",
		"description": "Load test payment",
	}

	return r.makeRequest("POST", "/api/v1/payments", payload, userID)
}

// scenarioProcessPayment processes a payment
func (r *LoadTestRunner) scenarioProcessPayment(userID, clientID string) error {
	paymentID := uuid.New().String()
	return r.makeRequest("POST", fmt.Sprintf("/api/v1/payments/%s/process", paymentID), nil, userID)
}

// makeRequest makes an HTTP request
func (r *LoadTestRunner) makeRequest(method, path string, payload interface{}, userID string) error {
	url := r.config.BaseURL + path

	var body []byte
	var err error
	if payload != nil {
		body, err = json.Marshal(payload)
		if err != nil {
			return err
		}
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", r.config.TenantID)
	req.Header.Set("X-User-ID", userID)

	resp, err := r.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	return nil
}

// reportMetrics periodically reports metrics
func (r *LoadTestRunner) reportMetrics(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			r.printCurrentMetrics()
		case <-ctx.Done():
			return
		case <-r.stopCh:
			return
		}
	}
}

// printCurrentMetrics prints current metrics
func (r *LoadTestRunner) printCurrentMetrics() {
	total := atomic.LoadInt64(&r.metrics.TotalRequests)
	success := atomic.LoadInt64(&r.metrics.SuccessfulRequests)
	failed := atomic.LoadInt64(&r.metrics.FailedRequests)
	totalLatency := atomic.LoadInt64(&r.metrics.TotalLatency)

	var avgLatency int64
	if total > 0 {
		avgLatency = totalLatency / total
	}

	successRate := float64(0)
	if total > 0 {
		successRate = float64(success) / float64(total) * 100
	}

	fmt.Printf("[%s] Requests: %d | Success: %d | Failed: %d | Success Rate: %.2f%% | Avg Latency: %dms\n",
		time.Now().Format("15:04:05"),
		total,
		success,
		failed,
		successRate,
		avgLatency,
	)
}

// printResults prints final test results
func (r *LoadTestRunner) printResults() {
	fmt.Printf("\n")
	fmt.Printf("========================================\n")
	fmt.Printf("      PHASE 2 LOAD TEST RESULTS         \n")
	fmt.Printf("========================================\n")
	fmt.Printf("\n")

	total := atomic.LoadInt64(&r.metrics.TotalRequests)
	success := atomic.LoadInt64(&r.metrics.SuccessfulRequests)
	failed := atomic.LoadInt64(&r.metrics.FailedRequests)
	totalLatency := atomic.LoadInt64(&r.metrics.TotalLatency)

	var avgLatency int64
	if total > 0 {
		avgLatency = totalLatency / total
	}

	successRate := float64(0)
	if total > 0 {
		successRate = float64(success) / float64(total) * 100
	}

	rps := float64(total) / r.config.TestDuration.Seconds()

	fmt.Printf("Test Configuration:\n")
	fmt.Printf("  Concurrent Users: %d\n", r.config.ConcurrentUsers)
	fmt.Printf("  Test Duration: %s\n", r.config.TestDuration)
	fmt.Printf("\n")

	fmt.Printf("Results:\n")
	fmt.Printf("  Total Requests: %d\n", total)
	fmt.Printf("  Successful: %d\n", success)
	fmt.Printf("  Failed: %d\n", failed)
	fmt.Printf("  Success Rate: %.2f%%\n", successRate)
	fmt.Printf("  Requests/Second: %.2f\n", rps)
	fmt.Printf("\n")

	fmt.Printf("Latency:\n")
	fmt.Printf("  Average: %dms\n", avgLatency)

	r.metrics.mu.RLock()
	fmt.Printf("  Min: %dms\n", r.metrics.MinLatency)
	fmt.Printf("  Max: %dms\n", r.metrics.MaxLatency)
	r.metrics.mu.RUnlock()

	fmt.Printf("\n")

	// Validate against requirements
	fmt.Printf("Validation:\n")
	passed := true

	if successRate >= 99.9 {
		fmt.Printf("  ✅ Success Rate >= 99.9%% (%.2f%%)\n", successRate)
	} else {
		fmt.Printf("  ❌ Success Rate >= 99.9%% (%.2f%%)\n", successRate)
		passed = false
	}

	// Note: P95 calculation would require storing all latencies
	fmt.Printf("  ⏳ P95 Latency < 200ms (requires detailed histogram)\n")

	fmt.Printf("\n")
	if passed {
		fmt.Printf("✅ PHASE 2 LOAD TEST PASSED\n")
	} else {
		fmt.Printf("❌ PHASE 2 LOAD TEST FAILED\n")
	}
	fmt.Printf("========================================\n")
}

func main() {
	config := LoadTestConfig{
		BaseURL:         "http://localhost:8080",
		ConcurrentUsers: 5000,
		TestDuration:    5 * time.Minute,
		RampUpTime:      30 * time.Second,
		TenantID:        "test-tenant",
	}

	runner := NewLoadTestRunner(config)

	ctx, cancel := context.WithTimeout(context.Background(), config.TestDuration+1*time.Minute)
	defer cancel()

	if err := runner.Run(ctx); err != nil {
		fmt.Printf("Load test failed: %v\n", err)
	}
}
