/**
 * Phase 6 Load Testing Script
 * 
 * This script tests the IMS ERP system under high load to validate:
 * - P95 latency < 200ms
 * - 99.9% success rate
 * - Support for 5,000 concurrent users
 * 
 * Run with: k6 run scripts/load-test-phase6.js
 */

import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Rate, Trend, Counter, Gauge } from 'k6/metrics';
import { htmlReport } from 'https://raw.githubusercontent.com/benc-uk/k6-reporter/main/dist/bundle.js';

// Custom metrics
const errorRate = new Rate('error_rate');
const p95Latency = new Trend('p95_latency');
const p99Latency = new Trend('p99_latency');
const invoiceCreationTrend = new Trend('invoice_creation_duration');
const paymentProcessingTrend = new Trend('payment_processing_duration');
const dashboardQueryTrend = new Trend('dashboard_query_duration');
const activeUsers = new Gauge('active_users');
const requestCounter = new Counter('total_requests');

// Test configuration
export const options = {
  stages: [
    // Ramp up to 5,000 users over 5 minutes
    { duration: '5m', target: 5000 },
    // Stay at 5,000 users for 10 minutes
    { duration: '10m', target: 5000 },
    // Ramp down
    { duration: '5m', target: 0 },
  ],
  thresholds: {
    // P95 latency must be below 200ms
    http_req_duration: ['p(95)<200'],
    // P99 latency must be below 500ms
    'p99_latency': ['p(99)<500'],
    // Error rate must be below 0.1% (99.9% success rate)
    error_rate: ['rate<0.001'],
    // 95% of invoice creations under 300ms
    'invoice_creation_duration': ['p(95)<300'],
    // 95% of payment processing under 250ms
    'payment_processing_duration': ['p(95)<250'],
    // 95% of dashboard queries under 150ms
    'dashboard_query_duration': ['p(95)<150'],
  },
  ext: {
    loadimpact: {
      distribution: {
        'amazon:us:ashburn': { loadZone: 'amazon:us:ashburn', percent: 40 },
        'amazon:ie:dublin': { loadZone: 'amazon:ie:dublin', percent: 30 },
        'amazon:sg:singapore': { loadZone: 'amazon:sg:singapore', percent: 30 },
      },
    },
  },
};

// Base URL - update for your environment
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const API_VERSION = '/api/v1';

// Authentication token (set via environment variable)
const AUTH_TOKEN = __ENV.AUTH_TOKEN || '';

// Default headers
const headers = {
  'Content-Type': 'application/json',
  'Accept': 'application/json',
  'X-Tenant-ID': 'test-tenant',
};

if (AUTH_TOKEN) {
  headers['Authorization'] = `Bearer ${AUTH_TOKEN}`;
}

// Helper function to make requests with metrics
function makeRequest(method, url, body = null, tags = {}) {
  const startTime = new Date();
  
  const response = body 
    ? http.request(method, url, JSON.stringify(body), { headers, tags })
    : http.request(method, url, null, { headers, tags });
  
  const duration = new Date() - startTime;
  
  // Track metrics
  requestCounter.add(1);
  p95Latency.add(duration);
  p99Latency.add(duration);
  
  // Check for errors
  const success = check(response, {
    'status is 200 or 201': (r) => r.status === 200 || r.status === 201,
    'response time < 200ms': (r) => r.timings.duration < 200,
  });
  
  errorRate.add(!success);
  
  return response;
}

// Generate random invoice data
function generateInvoiceData() {
  const now = new Date();
  const dueDate = new Date(now.getTime() + 30 * 24 * 60 * 60 * 1000);
  
  return {
    clientId: `client-${Math.floor(Math.random() * 1000)}`,
    dueDate: dueDate.toISOString(),
    currency: 'USD',
    lineItems: [
      {
        productId: `product-${Math.floor(Math.random() * 500)}`,
        description: `Test Product ${Math.floor(Math.random() * 100)}`,
        quantity: Math.floor(Math.random() * 10) + 1,
        unitPrice: (Math.random() * 1000).toFixed(2),
      },
      {
        productId: `product-${Math.floor(Math.random() * 500)}`,
        description: `Test Product ${Math.floor(Math.random() * 100)}`,
        quantity: Math.floor(Math.random() * 5) + 1,
        unitPrice: (Math.random() * 500).toFixed(2),
      },
    ],
    notes: 'Load test invoice',
  };
}

// Generate random payment data
function generatePaymentData(invoiceId) {
  return {
    invoiceId: invoiceId,
    amount: (Math.random() * 5000 + 100).toFixed(2),
    method: ['credit_card', 'bank_transfer', 'paypal'][Math.floor(Math.random() * 3)],
    reference: `REF-${Date.now()}-${Math.floor(Math.random() * 10000)}`,
  };
}

// Test scenario: Invoice Creation
function testInvoiceCreation() {
  group('Invoice Creation', () => {
    const startTime = new Date();
    
    const invoiceData = generateInvoiceData();
    const response = makeRequest(
      'POST',
      `${BASE_URL}${API_VERSION}/invoices`,
      invoiceData,
      { name: 'Create Invoice' }
    );
    
    const duration = new Date() - startTime;
    invoiceCreationTrend.add(duration);
    
    if (response.status === 201) {
      const invoice = JSON.parse(response.body);
      // Store invoice ID for payment test
      return invoice.id;
    }
    
    return null;
  });
}

// Test scenario: Payment Processing
function testPaymentProcessing(invoiceId) {
  group('Payment Processing', () => {
    const startTime = new Date();
    
    const paymentData = generatePaymentData(invoiceId);
    const response = makeRequest(
      'POST',
      `${BASE_URL}${API_VERSION}/payments`,
      paymentData,
      { name: 'Process Payment' }
    );
    
    const duration = new Date() - startTime;
    paymentProcessingTrend.add(duration);
    
    return response.status === 201;
  });
}

// Test scenario: Dashboard Queries
function testDashboardQueries() {
  group('Dashboard Queries', () => {
    const startTime = new Date();
    
    // Get main dashboard
    const dashboardResponse = makeRequest(
      'GET',
      `${BASE_URL}${API_VERSION}/dashboard`,
      null,
      { name: 'Get Dashboard' }
    );
    
    let duration = new Date() - startTime;
    dashboardQueryTrend.add(duration);
    
    // Get revenue metrics
    const revenueStart = new Date();
    makeRequest(
      'GET',
      `${BASE_URL}${API_VERSION}/metrics/revenue?start=${new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString()}&end=${new Date().toISOString()}`,
      null,
      { name: 'Get Revenue Metrics' }
    );
    
    duration = new Date() - revenueStart;
    dashboardQueryTrend.add(duration);
    
    // Get aging report
    const agingStart = new Date();
    makeRequest(
      'GET',
      `${BASE_URL}${API_VERSION}/metrics/aging`,
      null,
      { name: 'Get Aging Report' }
    );
    
    duration = new Date() - agingStart;
    dashboardQueryTrend.add(duration);
    
    return dashboardResponse.status === 200;
  });
}

// Test scenario: Invoice List Query
function testInvoiceListQuery() {
  group('Invoice List Query', () => {
    const response = makeRequest(
      'GET',
      `${BASE_URL}${API_VERSION}/invoices?page=1&pageSize=20`,
      null,
      { name: 'List Invoices' }
    );
    
    return response.status === 200;
  });
}

// Test scenario: GraphQL Queries
function testGraphQLQueries() {
  group('GraphQL Queries', () => {
    const query = `
      query {
        invoices(first: 10) {
          edges {
            node {
              id
              invoiceNumber
              status
              total
            }
          }
        }
      }
    `;
    
    const response = makeRequest(
      'POST',
      `${BASE_URL}/graphql`,
      { query },
      { name: 'GraphQL Invoices Query' }
    );
    
    return response.status === 200;
  });
}

// Main test function
export default function () {
  // Track active users
  activeUsers.add(1);
  
  // Execute test scenarios
  
  // 1. Dashboard queries (40% of traffic)
  if (Math.random() < 0.4) {
    testDashboardQueries();
  }
  
  // 2. Invoice list queries (30% of traffic)
  if (Math.random() < 0.3) {
    testInvoiceListQuery();
  }
  
  // 3. Invoice creation (15% of traffic)
  let invoiceId = null;
  if (Math.random() < 0.15) {
    invoiceId = testInvoiceCreation();
  }
  
  // 4. Payment processing (10% of traffic)
  if (invoiceId && Math.random() < 0.67) {
    testPaymentProcessing(invoiceId);
  }
  
  // 5. GraphQL queries (5% of traffic)
  if (Math.random() < 0.05) {
    testGraphQLQueries();
  }
  
  // Random sleep between 1-3 seconds to simulate real user behavior
  sleep(Math.random() * 2 + 1);
  
  activeUsers.add(-1);
}

// Setup function
export function setup() {
  console.log('Starting Phase 6 Load Test');
  console.log(`Target: 5,000 concurrent users`);
  console.log(`Requirements:`);
  console.log(`  - P95 latency < 200ms`);
  console.log(`  - 99.9% success rate`);
  console.log(`  - Base URL: ${BASE_URL}`);
  
  // Health check
  const response = http.get(`${BASE_URL}${API_VERSION}/health`);
  if (response.status !== 200) {
    throw new Error('Health check failed - aborting test');
  }
  
  return {
    startTime: new Date().toISOString(),
    baseUrl: BASE_URL,
  };
}

// Teardown function
export function teardown(data) {
  console.log('Load test completed');
  console.log(`Started at: ${data.startTime}`);
  console.log(`Base URL: ${data.baseUrl}`);
}

// Handle summary
export function handleSummary(data) {
  const report = {
    'load-test-report.html': htmlReport(data),
    'load-test-results.json': JSON.stringify(data, null, 2),
  };
  
  // Print summary to console
  console.log('\n=== LOAD TEST SUMMARY ===');
  console.log(`Total Requests: ${data.metrics.http_reqs.values.count}`);
  console.log(`Failed Requests: ${data.metrics.http_req_failed.values.passes}`);
  console.log(`Success Rate: ${((1 - data.metrics.http_req_failed.values.rate) * 100).toFixed(3)}%`);
  console.log(`\nLatency Metrics:`);
  console.log(`  P50: ${data.metrics.http_req_duration.values['p(50)'].toFixed(2)}ms`);
  console.log(`  P95: ${data.metrics.http_req_duration.values['p(95)'].toFixed(2)}ms`);
  console.log(`  P99: ${data.metrics.http_req_duration.values['p(99)'].toFixed(2)}ms`);
  console.log(`  Avg: ${data.metrics.http_req_duration.values.avg.toFixed(2)}ms`);
  console.log(`\nScenario Metrics:`);
  console.log(`  Invoice Creation P95: ${data.metrics.invoice_creation_duration?.values['p(95)']?.toFixed(2) || 'N/A'}ms`);
  console.log(`  Payment Processing P95: ${data.metrics.payment_processing_duration?.values['p(95)']?.toFixed(2) || 'N/A'}ms`);
  console.log(`  Dashboard Query P95: ${data.metrics.dashboard_query_duration?.values['p(95)']?.toFixed(2) || 'N/A'}ms`);
  console.log('\n=== VALIDATION ===');
  
  const p95Valid = data.metrics.http_req_duration.values['p(95)'] < 200;
  const successRateValid = (1 - data.metrics.http_req_failed.values.rate) >= 0.999;
  
  console.log(`P95 Latency < 200ms: ${p95Valid ? 'PASS' : 'FAIL'}`);
  console.log(`Success Rate >= 99.9%: ${successRateValid ? 'PASS' : 'FAIL'}`);
  console.log(`Overall: ${p95Valid && successRateValid ? 'ALL TESTS PASSED' : 'SOME TESTS FAILED'}`);
  
  return report;
}
