<script lang="ts">
  import { onMount } from 'svelte';
  import BarChart from '$lib/shared/components/display/charts/BarChart.svelte';
  import LineChart from '$lib/shared/components/display/charts/LineChart.svelte';
  import DonutChart from '$lib/shared/components/display/charts/DonutChart.svelte';
  import Card from '$lib/shared/components/layout/Card.svelte';

  // Sales data
  const salesData = [
    { label: 'Jan', value: 45000 },
    { label: 'Feb', value: 52000 },
    { label: 'Mar', value: 48000 },
    { label: 'Apr', value: 61000 },
    { label: 'May', value: 58000 },
    { label: 'Jun', value: 67000 },
  ];

  // Order trends data
  const orderTrendsData = [
    {
      name: 'Orders',
      data: [
        { x: 'Mon', y: 45 },
        { x: 'Tue', y: 52 },
        { x: 'Wed', y: 48 },
        { x: 'Thu', y: 61 },
        { x: 'Fri', y: 58 },
        { x: 'Sat', y: 42 },
        { x: 'Sun', y: 35 },
      ],
      color: '#3b82f6',
    },
    {
      name: 'Completed',
      data: [
        { x: 'Mon', y: 38 },
        { x: 'Tue', y: 48 },
        { x: 'Wed', y: 42 },
        { x: 'Thu', y: 55 },
        { x: 'Fri', y: 52 },
        { x: 'Sat', y: 38 },
        { x: 'Sun', y: 30 },
      ],
      color: '#10b981',
    },
  ];

  // Revenue by category
  const revenueByCategory = [
    { label: 'Electronics', value: 35000, color: '#3b82f6' },
    { label: 'Clothing', value: 28000, color: '#10b981' },
    { label: 'Home & Garden', value: 22000, color: '#f59e0b' },
    { label: 'Sports', value: 15000, color: '#ef4444' },
  ];

  // Inventory status
  const inventoryStatus = [
    { label: 'In Stock', value: 65, color: '#10b981' },
    { label: 'Low Stock', value: 20, color: '#f59e0b' },
    { label: 'Out of Stock', value: 10, color: '#ef4444' },
    { label: 'On Order', value: 5, color: '#3b82f6' },
  ];

  // Top products
  const topProducts = [
    { label: 'Laptop Pro X1', value: 125 },
    { label: 'Wireless Mouse', value: 98 },
    { label: 'USB-C Hub', value: 87 },
    { label: 'Monitor 27"', value: 76 },
    { label: 'Keyboard K2', value: 65 },
  ];

  let loading = true;

  onMount(() => {
    // Simulate loading
    setTimeout(() => {
      loading = false;
    }, 500);
  });
</script>

<div class="chart-widgets">
  <div class="widgets-grid">
    <!-- Revenue Trend -->
    <div class="widget widget-large">
      <Card>
        <div class="widget-header">
          <div>
            <h3 class="widget-title">Revenue Trend</h3>
            <p class="widget-subtitle">Monthly revenue performance</p>
          </div>
          <div class="widget-actions">
            <select class="period-select">
              <option value="6m">Last 6 months</option>
              <option value="1y">Last year</option>
              <option value="all">All time</option>
            </select>
          </div>
        </div>
        {#if loading}
          <div class="loading-state">
            <div class="skeleton-chart"></div>
          </div>
        {:else}
          <BarChart data={salesData} height={280} showValues={true} />
        {/if}
      </Card>
    </div>

    <!-- Order Trends -->
    <div class="widget widget-large">
      <Card>
        <div class="widget-header">
          <div>
            <h3 class="widget-title">Order Trends</h3>
            <p class="widget-subtitle">Daily orders vs completed</p>
          </div>
          <div class="widget-badge">
            <span class="badge positive">+12.5%</span>
          </div>
        </div>
        {#if loading}
          <div class="loading-state">
            <div class="skeleton-chart"></div>
          </div>
        {:else}
          <LineChart series={orderTrendsData} height={280} showGrid={true} />
        {/if}
      </Card>
    </div>

    <!-- Revenue by Category -->
    <div class="widget">
      <Card>
        <div class="widget-header">
          <div>
            <h3 class="widget-title">Revenue by Category</h3>
            <p class="widget-subtitle">Sales distribution</p>
          </div>
        </div>
        {#if loading}
          <div class="loading-state">
            <div class="skeleton-chart circular"></div>
          </div>
        {:else}
          <DonutChart data={revenueByCategory} size={240} donutWidth={50} />
        {/if}
      </Card>
    </div>

    <!-- Inventory Status -->
    <div class="widget">
      <Card>
        <div class="widget-header">
          <div>
            <h3 class="widget-title">Inventory Status</h3>
            <p class="widget-subtitle">Current stock levels</p>
          </div>
        </div>
        {#if loading}
          <div class="loading-state">
            <div class="skeleton-chart circular"></div>
          </div>
        {:else}
          <DonutChart data={inventoryStatus} size={240} donutWidth={50} />
        {/if}
      </Card>
    </div>

    <!-- Top Products -->
    <div class="widget widget-full">
      <Card>
        <div class="widget-header">
          <div>
            <h3 class="widget-title">Top Products</h3>
            <p class="widget-subtitle">Best selling items this month</p>
          </div>
          <a href="/products" class="view-all-link">View all →</a>
        </div>
        {#if loading}
          <div class="loading-state">
            <div class="skeleton-list">
              {#each Array(5) as _}
                <div class="skeleton-item"></div>
              {/each}
            </div>
          </div>
        {:else}
          <div class="top-products-list">
            {#each topProducts as product, index}
              <div class="product-item">
                <div class="product-rank">#{index + 1}</div>
                <div class="product-info">
                  <span class="product-name">{product.label}</span>
                  <div class="product-bar">
                    <div
                      class="product-progress"
                      style="width: {(product.value / topProducts[0].value) * 100}%"
                    ></div>
                  </div>
                </div>
                <span class="product-value">{product.value} sold</span>
              </div>
            {/each}
          </div>
        {/if}
      </Card>
    </div>
  </div>
</div>

<style>
  .chart-widgets {
    width: 100%;
  }

  .widgets-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 1.5rem;
  }

  .widget {
    min-width: 0;
  }

  .widget-large {
    grid-column: span 2;
  }

  .widget-full {
    grid-column: span 2;
  }

  .widget-header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    margin-bottom: 1.25rem;
  }

  .widget-title {
    font-size: 1rem;
    font-weight: 600;
    color: #111827;
    margin: 0;
  }

  .widget-subtitle {
    font-size: 0.875rem;
    color: #6b7280;
    margin: 0.25rem 0 0 0;
  }

  .widget-actions {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .period-select {
    padding: 0.375rem 0.75rem;
    border: 1px solid #d1d5db;
    border-radius: 0.375rem;
    font-size: 0.875rem;
    color: #374151;
    background: white;
    cursor: pointer;
  }

  .period-select:focus {
    outline: none;
    border-color: #3b82f6;
    box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
  }

  .widget-badge {
    display: flex;
    align-items: center;
  }

  .badge {
    display: inline-flex;
    align-items: center;
    padding: 0.25rem 0.625rem;
    border-radius: 9999px;
    font-size: 0.75rem;
    font-weight: 600;
  }

  .badge.positive {
    background: #d1fae5;
    color: #065f46;
  }

  .view-all-link {
    font-size: 0.875rem;
    font-weight: 500;
    color: #3b82f6;
    text-decoration: none;
  }

  .view-all-link:hover {
    text-decoration: underline;
  }

  /* Loading states */
  .loading-state {
    padding: 1rem 0;
  }

  .skeleton-chart {
    height: 280px;
    background: linear-gradient(90deg, #f3f4f6 25%, #e5e7eb 50%, #f3f4f6 75%);
    background-size: 200% 100%;
    animation: shimmer 1.5s infinite;
    border-radius: 0.5rem;
  }

  .skeleton-chart.circular {
    width: 240px;
    height: 240px;
    margin: 0 auto;
    border-radius: 50%;
  }

  .skeleton-list {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .skeleton-item {
    height: 48px;
    background: linear-gradient(90deg, #f3f4f6 25%, #e5e7eb 50%, #f3f4f6 75%);
    background-size: 200% 100%;
    animation: shimmer 1.5s infinite;
    border-radius: 0.375rem;
  }

  @keyframes shimmer {
    0% {
      background-position: 200% 0;
    }
    100% {
      background-position: -200% 0;
    }
  }

  /* Top products list */
  .top-products-list {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .product-item {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 0.75rem;
    background: #f9fafb;
    border-radius: 0.5rem;
    transition: background 0.2s ease;
  }

  .product-item:hover {
    background: #f3f4f6;
  }

  .product-rank {
    width: 2rem;
    height: 2rem;
    display: flex;
    align-items: center;
    justify-content: center;
    background: #e5e7eb;
    border-radius: 0.375rem;
    font-size: 0.875rem;
    font-weight: 600;
    color: #374151;
    flex-shrink: 0;
  }

  .product-item:first-child .product-rank {
    background: linear-gradient(135deg, #fbbf24 0%, #f59e0b 100%);
    color: white;
  }

  .product-item:nth-child(2) .product-rank {
    background: linear-gradient(135deg, #9ca3af 0%, #6b7280 100%);
    color: white;
  }

  .product-item:nth-child(3) .product-rank {
    background: linear-gradient(135deg, #b45309 0%, #92400e 100%);
    color: white;
  }

  .product-info {
    flex: 1;
    min-width: 0;
  }

  .product-name {
    display: block;
    font-size: 0.875rem;
    font-weight: 500;
    color: #111827;
    margin-bottom: 0.375rem;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .product-bar {
    height: 4px;
    background: #e5e7eb;
    border-radius: 2px;
    overflow: hidden;
  }

  .product-progress {
    height: 100%;
    background: linear-gradient(90deg, #3b82f6 0%, #2563eb 100%);
    border-radius: 2px;
    transition: width 0.5s ease;
  }

  .product-value {
    font-size: 0.875rem;
    font-weight: 500;
    color: #6b7280;
    flex-shrink: 0;
  }

  @media (max-width: 1024px) {
    .widgets-grid {
      grid-template-columns: 1fr;
    }

    .widget-large {
      grid-column: span 1;
    }

    .widget-full {
      grid-column: span 1;
    }
  }

  @media (max-width: 640px) {
    .widget-header {
      flex-direction: column;
      gap: 0.75rem;
    }

    .widget-actions {
      width: 100%;
    }

    .period-select {
      width: 100%;
    }
  }
</style>
