<script lang="ts">
  import { goto } from '$app/navigation';
  import Input from '$lib/shared/components/forms/Input.svelte';
  import Button from '$lib/shared/components/forms/Button.svelte';
  import { auth } from '$lib/shared/stores/auth';

  let tenantId = $state('');
  let email = $state('');
  let password = $state('');
  let error = $state('');
  let loading = $state(false);

  const DEMO_TENANT_ID = '00000000-0000-0000-0000-000000000001';

  async function handleSubmit(event: Event) {
    event.preventDefault();
    error = '';
    loading = true;

    const tenantToUse = tenantId || DEMO_TENANT_ID;

    if (result.success) {
      goto('/');
    } else {
      error = result.error || 'Login failed. Please try again.';
    }

    loading = false;
  }

  function handleKeydown(event: KeyboardEvent) {
    if (event.key === 'Enter' && !event.shiftKey) {
      handleSubmit(event);
    }
  }
</script>

<svelte:head>
  <title>Login - ERP System</title>
</svelte:head>

<div class="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900 px-4 py-12">
  <div class="w-full max-w-md">
    <div class="text-center mb-8">
      <div
        class="mx-auto h-12 w-12 rounded-xl bg-primary-600 flex items-center justify-center mb-4"
      >
        <svg
          class="h-7 w-7 text-white"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4"
          />
        </svg>
      </div>
      <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Welcome back</h1>
      <p class="mt-2 text-gray-600 dark:text-gray-400">
        Sign in to your ERP account
      </p>
    </div>

    <div class="bg-white dark:bg-gray-800 rounded-2xl shadow-xl border border-gray-200 dark:border-gray-700 p-8">
      {#if error}
        <div
          class="mb-6 p-4 rounded-lg bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800"
        >
          <div class="flex items-center">
            <svg
              class="h-5 w-5 text-red-500 dark:text-red-400 mr-2"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            </svg>
            <p class="text-sm text-red-700 dark:text-red-300">{error}</p>
          </div>
        </div>
      {/if}

      <form onsubmit={handleSubmit} class="space-y-5">
        <Input
          id="tenantId"
          label="Tenant ID"
          type="text"
          bind:value={tenantId}
          placeholder="Enter your tenant ID"
          helpText="Leave empty to use demo tenant"
          autocomplete="organization"
        />

        <Input
          id="email"
          label="Email"
          type="email"
          bind:value={email}
          placeholder="you@example.com"
          required
          autocomplete="email"
          onkeydown={handleKeydown}
        />

        <Input
          id="password"
          label="Password"
          type="password"
          bind:value={password}
          placeholder="Enter your password"
          required
          autocomplete="current-password"
          onkeydown={handleKeydown}
        />

        <div class="flex items-center justify-between">
          <label class="flex items-center">
            <input
              type="checkbox"
              class="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500"
            />
            <span class="ml-2 text-sm text-gray-600 dark:text-gray-400">
              Remember me
            </span>
          </label>
          <a
            href="/forgot-password"
            class="text-sm font-medium text-primary-600 hover:text-primary-500"
          >
            Forgot password?
          </a>
        </div>

        <Button type="submit" fullWidth {loading} disabled={loading}>
          {#if loading}
            Signing in...
          {:else}
            Sign in
          {/if}
        </Button>
      </form>

      <div class="mt-6 text-center">
        <p class="text-sm text-gray-600 dark:text-gray-400">
          Don't have an account?
          <a
            href="/register"
            class="font-medium text-primary-600 hover:text-primary-500"
          >
            Contact your administrator
          </a>
        </p>
      </div>
    </div>

    <div class="mt-8 text-center">
      <p class="text-xs text-gray-500 dark:text-gray-400">
        ERP System v1.0.0
      </p>
    </div>
  </div>
</div>
