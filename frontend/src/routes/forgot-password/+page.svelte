<script lang="ts">
  import Input from '$lib/shared/components/forms/Input.svelte';
  import Button from '$lib/shared/components/forms/Button.svelte';

  let email = $state('');
  let loading = $state(false);
  let submitted = $state(false);
  let error = $state('');

  async function handleSubmit(event: Event) {
    event.preventDefault();
    error = '';
    loading = true;

    try {
      const response = await fetch('/api/auth/forgot-password', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email }),
      });

      if (response.ok) {
        submitted = true;
      } else {
        const data = await response.json().catch(() => ({}));
        error = data.message || 'Failed to send reset link. Please try again.';
      }
    } catch {
      error = 'An unexpected error occurred. Please try again.';
    } finally {
      loading = false;
    }
  }
</script>

<svelte:head>
  <title>Forgot Password - ERP System</title>
</svelte:head>

<div class="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900 px-4 py-12">
  <div class="w-full max-w-md">
    <div class="text-center mb-8">
      <div class="mx-auto h-12 w-12 rounded-xl bg-primary-600 flex items-center justify-center mb-4">
        <svg class="h-7 w-7 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
            d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" />
        </svg>
      </div>
      <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Forgot your password?</h1>
      <p class="mt-2 text-gray-600 dark:text-gray-400">
        Enter your email and we'll send you a reset link.
      </p>
    </div>

    <div class="bg-white dark:bg-gray-800 rounded-2xl shadow-xl border border-gray-200 dark:border-gray-700 p-8">
      {#if submitted}
        <div class="text-center space-y-4">
          <div class="mx-auto h-12 w-12 rounded-full bg-emerald-100 flex items-center justify-center">
            <svg class="h-6 w-6 text-emerald-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
            </svg>
          </div>
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">Check your email</h2>
          <p class="text-sm text-gray-600 dark:text-gray-400">
            If an account exists for <strong>{email}</strong>, you'll receive a password reset link shortly.
          </p>
          <a
            href="/login"
            class="inline-block mt-2 text-sm font-medium text-primary-600 hover:text-primary-500"
          >
            ← Back to Login
          </a>
        </div>
      {:else}
        {#if error}
          <div class="mb-6 p-4 rounded-lg bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800">
            <div class="flex items-center gap-2">
              <svg class="h-5 w-5 text-red-500 dark:text-red-400 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                  d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <p class="text-sm text-red-700 dark:text-red-300">{error}</p>
            </div>
          </div>
        {/if}

        <form onsubmit={handleSubmit} class="space-y-5">
          <Input
            id="email"
            label="Email address"
            type="email"
            bind:value={email}
            placeholder="you@example.com"
            required
            autocomplete="email"
          />

          <Button type="submit" fullWidth {loading} disabled={loading}>
            {#if loading}
              Sending...
            {:else}
              Send Reset Link
            {/if}
          </Button>
        </form>

        <div class="mt-6 text-center">
          <a href="/login" class="text-sm font-medium text-primary-600 hover:text-primary-500">
            ← Back to Login
          </a>
        </div>
      {/if}
    </div>

    <div class="mt-8 text-center">
      <p class="text-xs text-gray-500 dark:text-gray-400">ERP System v1.0.0</p>
    </div>
  </div>
</div>
