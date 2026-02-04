import { writable, derived } from 'svelte/store';

export type ToastVariant = 'success' | 'error' | 'warning' | 'info';

export interface Toast {
  id: string;
  variant: ToastVariant;
  title?: string;
  message: string;
  duration: number;
  dismissible: boolean;
}

export interface ToastOptions {
  variant?: ToastVariant;
  title?: string;
  duration?: number;
  dismissible?: boolean;
}

function createToastStore() {
  const { subscribe, update } = writable<Toast[]>([]);

  function generateId(): string {
    return `toast-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
  }

  function add(message: string, options: ToastOptions = {}): string {
    const id = generateId();
    const toast: Toast = {
      id,
      variant: options.variant || 'info',
      title: options.title,
      message,
      duration: options.duration ?? 5000,
      dismissible: options.dismissible ?? true,
    };

    update((toasts) => [...toasts, toast]);
    return id;
  }

  function remove(id: string) {
    update((toasts) => toasts.filter((t) => t.id !== id));
  }

  function clear() {
    update(() => []);
  }

  // Convenience methods
  function success(message: string, options: Omit<ToastOptions, 'variant'> = {}): string {
    return add(message, { ...options, variant: 'success' });
  }

  function error(message: string, options: Omit<ToastOptions, 'variant'> = {}): string {
    return add(message, { ...options, variant: 'error', duration: options.duration ?? 8000 });
  }

  function warning(message: string, options: Omit<ToastOptions, 'variant'> = {}): string {
    return add(message, { ...options, variant: 'warning' });
  }

  function info(message: string, options: Omit<ToastOptions, 'variant'> = {}): string {
    return add(message, { ...options, variant: 'info' });
  }

  return {
    subscribe,
    add,
    remove,
    clear,
    success,
    error,
    warning,
    info,
  };
}

export const toast = createToastStore();

// Derived store to get toast count
export const toastCount = derived(toast, ($toasts) => $toasts.length);
