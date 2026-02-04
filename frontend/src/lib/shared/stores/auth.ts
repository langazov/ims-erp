import { writable, derived, get } from 'svelte/store';
import type { User, TokenPair, AuthState } from '../api/auth';
import { login as loginApi, logout as logoutApi, refreshToken as refreshTokenApi } from '../api/auth';
import { api } from '../api';

const initialState: AuthState = {
  user: null,
  tokens: null,
  sessionId: null,
  isAuthenticated: false,
  isLoading: true,
};

function createAuthStore() {
  const { subscribe, set, update } = writable<AuthState>(initialState);

  const STORAGE_KEY = 'auth_state';

  function loadFromStorage(): AuthState {
    if (typeof window === 'undefined') return initialState;

    try {
      const stored = localStorage.getItem(STORAGE_KEY);
      if (stored) {
        const parsed = JSON.parse(stored);
        return {
          ...initialState,
          ...parsed,
          isLoading: false,
        };
      }
    } catch (e) {
      console.error('Failed to load auth state from storage:', e);
    }
    return { ...initialState, isLoading: false };
  }

  function saveToStorage(state: AuthState): void {
    if (typeof window === 'undefined') return;

    try {
      localStorage.setItem(
        STORAGE_KEY,
        JSON.stringify({
          user: state.user,
          tokens: state.tokens,
          sessionId: state.sessionId,
          isAuthenticated: state.isAuthenticated,
        })
      );
    } catch (e) {
      console.error('Failed to save auth state to storage:', e);
    }
  }

  return {
    subscribe,

    init: () => {
      const storedState = loadFromStorage();
      set(storedState);

      if (storedState.tokens) {
        api.setAuthToken(storedState.tokens.accessToken);
      }

      return storedState;
    },

    login: async (tenantId: string, email: string, password: string) => {
      update((state) => ({ ...state, isLoading: true }));

      try {
        const response = await loginApi(tenantId, { email, password });

        const newState: AuthState = {
          user: response.user,
          tokens: response.tokens,
          sessionId: response.sessionId,
          isAuthenticated: true,
          isLoading: false,
        };

        set(newState);
        saveToStorage(newState);
        api.setAuthToken(response.tokens.accessToken);

        return { success: true, user: response.user };
      } catch (error) {
        const errorMessage = error instanceof Error ? error.message : 'Login failed';
        update((state) => ({ ...state, isLoading: false }));
        return { success: false, error: errorMessage };
      }
    },

    logout: async () => {
      const state = get({ subscribe });
      if (state.user && state.sessionId) {
        try {
          await logoutApi(state.user.id, state.sessionId);
        } catch (e) {
          console.error('Logout API call failed:', e);
        }
      }

      set({ ...initialState, isLoading: false });
      api.clearAuthToken();
      localStorage.removeItem(STORAGE_KEY);

      if (typeof window !== 'undefined') {
        window.location.href = '/login';
      }
    },

    refreshTokens: async () => {
      const state = get({ subscribe });
      if (!state.tokens?.refreshToken) return false;

      try {
        const newTokens = await refreshTokenApi(state.tokens.refreshToken);
        update((s) => {
          const newState = {
            ...s,
            tokens: newTokens,
          };
          saveToStorage(newState);
          api.setAuthToken(newTokens.accessToken);
          return newState;
        });
        return true;
      } catch (e) {
        console.error('Token refresh failed:', e);
        return false;
      }
    },

    clear: () => {
      set({ ...initialState, isLoading: false });
      api.clearAuthToken();
      localStorage.removeItem(STORAGE_KEY);
    },
  };
}

// Create the auth store instance
const auth = createAuthStore();

// Derived stores for convenience
const isAuthenticated = derived(auth, ($auth) => $auth.isAuthenticated);
const currentUser = derived(auth, ($auth) => $auth.user);
const authTokens = derived(auth, ($auth) => $auth.tokens);
const authLoading = derived(auth, ($auth) => $auth.isLoading);

export { auth, isAuthenticated, currentUser, authTokens, authLoading };
export type { User, TokenPair, AuthState } from '../api/auth';
