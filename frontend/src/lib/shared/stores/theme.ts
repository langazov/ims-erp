import { writable } from 'svelte/store';
import { browser } from '$app/environment';

type Theme = 'light' | 'dark' | 'auto';

function createThemeStore() {
  const stored = browser ? (localStorage.getItem('theme') as Theme | null) : null;
  const initial: Theme = stored || 'auto';
  const { subscribe, set, update } = writable<Theme>(initial);

  function applyTheme(theme: Theme) {
    if (!browser) return;
    const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
    const isDark = theme === 'dark' || (theme === 'auto' && prefersDark);
    document.documentElement.classList.toggle('dark', isDark);
  }

  if (browser) {
    applyTheme(initial);
  }

  return {
    subscribe,
    set: (theme: Theme) => {
      if (browser) {
        localStorage.setItem('theme', theme);
        applyTheme(theme);
      }
      set(theme);
    },
    toggle: () => {
      update(t => {
        const next: Theme = t === 'dark' ? 'light' : 'dark';
        if (browser) {
          localStorage.setItem('theme', next);
          applyTheme(next);
        }
        return next;
      });
    },
    init: () => {
      const stored = browser ? (localStorage.getItem('theme') as Theme | null) : null;
      const theme: Theme = stored || 'auto';
      if (browser) applyTheme(theme);
      set(theme);
    }
  };
}

export const theme = createThemeStore();
export type { Theme };
