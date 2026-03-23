// Analytics API
import api from './index';

export interface DashboardKPIs {
  revenueToday: string;
  revenueThisMonth: string;
  revenueGrowthPct: number;
  activeOrders: number;
  pendingOrders: number;
  overdueInvoices: number;
  overdueAmount: string;
  lowStockItems: number;
  totalClients: number;
  newClientsThisMonth: number;
}

export interface TimeSeriesPoint {
  date: string;
  value: number;
}

export interface RevenueTimeSeries {
  current: TimeSeriesPoint[];
  previous: TimeSeriesPoint[];
  period: 'daily' | 'weekly' | 'monthly';
}

export interface TopClient {
  clientId: string;
  clientName: string;
  totalRevenue: string;
  invoiceCount: number;
  lastActivity: string;
}

export interface InvoiceAging {
  current: string;
  days1_30: string;
  days31_60: string;
  days61_90: string;
  over90: string;
  totalOutstanding: string;
}

export interface PaymentMethodBreakdown {
  method: string;
  count: number;
  amount: string;
  percentage: number;
}

export interface OrderFulfillmentMetrics {
  totalOrders: number;
  fulfilled: number;
  pending: number;
  cancelled: number;
  fulfillmentRate: number;
  avgFulfillmentDays: number;
}

export type DatePeriod = 'today' | 'week' | 'month' | 'quarter' | 'year' | 'custom';

export async function getDashboardKPIs(period?: DatePeriod): Promise<DashboardKPIs> {
  return api.get('/analytics/dashboard', { period });
}

export async function getRevenueTimeSeries(period: 'daily' | 'weekly' | 'monthly', dateFrom?: string, dateTo?: string): Promise<RevenueTimeSeries> {
  return api.get('/analytics/revenue', { period, dateFrom, dateTo });
}

export async function getTopClients(limit = 10, period?: DatePeriod): Promise<TopClient[]> {
  return api.get('/analytics/top-clients', { limit, period });
}

export async function getInvoiceAging(): Promise<InvoiceAging> {
  return api.get('/analytics/invoice-aging');
}

export async function getPaymentMethodBreakdown(period?: DatePeriod): Promise<PaymentMethodBreakdown[]> {
  return api.get('/analytics/payment-methods', { period });
}

export async function getOrderFulfillmentMetrics(period?: DatePeriod): Promise<OrderFulfillmentMetrics> {
  return api.get('/analytics/order-fulfillment', { period });
}

export async function getInventoryTurnover(warehouseId?: string): Promise<{
  turnoverRate: number;
  avgDaysInStock: number;
  slowMovingItems: number;
}> {
  return api.get('/analytics/inventory-turnover', { warehouseId });
}
