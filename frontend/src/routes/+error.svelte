<script lang="ts">
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';

  function goHome() {
    goto('/dashboard');
  }

  function goBack() {
    history.back();
  }
</script>

<div class="error-page">
  <div class="error-container">
    <div class="error-illustration">
      {#if $page.status === 404}
        <svg class="error-icon" viewBox="0 0 200 200" fill="none" xmlns="http://www.w3.org/2000/svg">
          <circle cx="100" cy="100" r="80" stroke="#E5E7EB" stroke-width="4"/>
          <path d="M70 80L90 80" stroke="#9CA3AF" stroke-width="6" stroke-linecap="round"/>
          <path d="M110 80L130 80" stroke="#9CA3AF" stroke-width="6" stroke-linecap="round"/>
          <path d="M75 130C85 120 115 120 125 130" stroke="#9CA3AF" stroke-width="6" stroke-linecap="round"/>
          <path d="M40 40L160 160" stroke="#E5E7EB" stroke-width="4" stroke-linecap="round"/>
        </svg>
      {:else}
        <svg class="error-icon" viewBox="0 0 200 200" fill="none" xmlns="http://www.w3.org/2000/svg">
          <circle cx="100" cy="100" r="80" stroke="#FEE2E2" stroke-width="4"/>
          <path d="M70 70L130 130" stroke="#EF4444" stroke-width="6" stroke-linecap="round"/>
          <path d="M130 70L70 130" stroke="#EF4444" stroke-width="6" stroke-linecap="round"/>
          <circle cx="100" cy="160" r="4" fill="#EF4444"/>
        </svg>
      {/if}
    </div>

    <div class="error-content">
      <h1 class="error-code">{$page.status}</h1>
      <h2 class="error-title">
        {#if $page.status === 404}
          Page Not Found
        {:else if $page.status === 500}
          Server Error
        {:else}
          {$page.error?.message || 'An error occurred'}
        {/if}
      </h2>
      <p class="error-message">
        {#if $page.status === 404}
          The page you're looking for doesn't exist or has been moved.
        {:else if $page.status === 500}
          Something went wrong on our end. Please try again later.
        {:else}
          {$page.error?.message || 'An unexpected error occurred.'}
        {/if}
      </p>

      <div class="error-actions">
        <button class="btn btn-primary" on:click={goHome}>
          <svg class="btn-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6"/>
          </svg>
          Go Home
        </button>
        <button class="btn btn-secondary" on:click={goBack}>
          <svg class="btn-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18"/>
          </svg>
          Go Back
        </button>
      </div>

      {#if $page.status === 500}
        <div class="support-section">
          <p>If this problem persists, please contact support.</p>
          <a href="/support" class="support-link">Contact Support →</a>
        </div>
      {/if}
    </div>
  </div>
</div>

<style>
  .error-page {
    min-height: 100vh;
    display: flex;
    align-items: center;
    justify-content: center;
    background: linear-gradient(135deg, #f8fafc 0%, #e2e8f0 100%);
    padding: 2rem;
  }

  .error-container {
    max-width: 600px;
    width: 100%;
    text-align: center;
  }

  .error-illustration {
    margin-bottom: 2rem;
  }

  .error-icon {
    width: 200px;
    height: 200px;
    margin: 0 auto;
  }

  .error-content {
    background: white;
    border-radius: 1rem;
    padding: 2.5rem;
    box-shadow:
      0 4px 6px -1px rgba(0, 0, 0, 0.1),
      0 2px 4px -1px rgba(0, 0, 0, 0.06);
  }

  .error-code {
    font-size: 6rem;
    font-weight: 800;
    color: #e5e7eb;
    line-height: 1;
    margin: 0 0 0.5rem 0;
    background: linear-gradient(135deg, #1e293b 0%, #0f172a 100%);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    background-clip: text;
  }

  .error-title {
    font-size: 1.875rem;
    font-weight: 700;
    color: #111827;
    margin: 0 0 1rem 0;
  }

  .error-message {
    font-size: 1.125rem;
    color: #6b7280;
    margin: 0 0 2rem 0;
    line-height: 1.6;
  }

  .error-actions {
    display: flex;
    gap: 1rem;
    justify-content: center;
    flex-wrap: wrap;
  }

  .btn {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.875rem 1.5rem;
    border-radius: 0.5rem;
    font-size: 1rem;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s ease;
    border: none;
  }

  .btn-primary {
    background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
    color: white;
    box-shadow: 0 4px 6px -1px rgba(59, 130, 246, 0.2);
  }

  .btn-primary:hover {
    transform: translateY(-1px);
    box-shadow: 0 6px 8px -1px rgba(59, 130, 246, 0.3);
  }

  .btn-secondary {
    background: white;
    color: #374151;
    border: 1px solid #d1d5db;
  }

  .btn-secondary:hover {
    background: #f9fafb;
    border-color: #9ca3af;
  }

  .btn-icon {
    width: 1.25rem;
    height: 1.25rem;
  }

  .support-section {
    margin-top: 2rem;
    padding-top: 2rem;
    border-top: 1px solid #e5e7eb;
  }

  .support-section p {
    font-size: 0.875rem;
    color: #6b7280;
    margin: 0 0 0.5rem 0;
  }

  .support-link {
    font-size: 0.875rem;
    font-weight: 500;
    color: #3b82f6;
    text-decoration: none;
  }

  .support-link:hover {
    text-decoration: underline;
  }

  @media (max-width: 640px) {
    .error-page {
      padding: 1rem;
    }

    .error-content {
      padding: 1.5rem;
    }

    .error-code {
      font-size: 4rem;
    }

    .error-title {
      font-size: 1.5rem;
    }

    .error-message {
      font-size: 1rem;
    }

    .error-actions {
      flex-direction: column;
    }

    .btn {
      width: 100%;
      justify-content: center;
    }
  }
</style>
