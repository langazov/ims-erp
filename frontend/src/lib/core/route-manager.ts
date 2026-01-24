// src/lib/core/route-manager.ts

import { writable, get, type Readable } from 'svelte/store';
import type {
  RouteDefinition,
  RouteManagerAPI,
  NavigateOptions,
  RouteInfo,
  NavigationItem,
  Unsubscribe
} from './types';

export interface RouteManager extends RouteManagerAPI {
  getAllRoutes(): RegisteredRoute[];
  getNavigation(): NavigationItem[];
  getStore(): Readable<RegisteredRoute[]>;
  matchRoute(path: string): { route: RegisteredRoute; params: Record<string, string> } | null;
  addNavigationItem(item: NavigationItem): void;
  removeNavigationItem(id: string): void;
  getNavigationStore(): Readable<NavigationItem[]>;
  onRouteChange(handler: (route: RouteInfo) => void): Unsubscribe;
}

export interface RegisteredRoute extends RouteDefinition {
  pluginId?: string;
  fullPath: string;
}

export function createRouteManager(): RouteManager {
  const routes = writable<RegisteredRoute[]>([]);
  const navigation = writable<NavigationItem[]>([]);
  const routeChangeHandlers = new Set<(route: RouteInfo) => void>();
  
  let currentRoute: RouteInfo = {
    path: '/',
    params: {},
    query: {}
  };

  function matchPath(
    pattern: string,
    path: string
  ): Record<string, string> | null {
    const patternParts = pattern.split('/').filter(Boolean);
    const pathParts = path.split('/').filter(Boolean);

    // Check for catch-all routes
    const hasCatchAll = patternParts.some(p => p.includes('...'));
    
    if (!hasCatchAll && patternParts.length !== pathParts.length) {
      return null;
    }

    const params: Record<string, string> = {};

    for (let i = 0; i < patternParts.length; i++) {
      const pattern = patternParts[i];
      const part = pathParts[i];

      // Handle catch-all [...rest]
      if (pattern.startsWith('[...') && pattern.endsWith(']')) {
        const paramName = pattern.slice(4, -1);
        params[paramName] = pathParts.slice(i).join('/');
        return params;
      }

      // Handle optional catch-all [[...rest]]
      if (pattern.startsWith('[[...') && pattern.endsWith(']]')) {
        const paramName = pattern.slice(5, -2);
        params[paramName] = pathParts.slice(i).join('/') || '';
        return params;
      }

      // Handle dynamic segment [param] or :param
      if (pattern.startsWith('[') && pattern.endsWith(']')) {
        const paramName = pattern.slice(1, -1);
        if (part === undefined) return null;
        params[paramName] = part;
      } else if (pattern.startsWith(':')) {
        const paramName = pattern.slice(1);
        if (part === undefined) return null;
        params[paramName] = part;
      } else if (pattern !== part) {
        return null;
      }
    }

    return params;
  }

  function normalizeFullPath(basePath: string, routePath: string): string {
    const base = basePath.replace(/\/$/, '');
    const route = routePath.startsWith('/') ? routePath : `/${routePath}`;
    return route === '/' ? base || '/' : `${base}${route}`;
  }

  function sortRoutes(routes: RegisteredRoute[]): RegisteredRoute[] {
    return routes.sort((a, b) => {
      // Static routes before dynamic
      const aIsDynamic = a.fullPath.includes('[') || a.fullPath.includes(':');
      const bIsDynamic = b.fullPath.includes('[') || b.fullPath.includes(':');
      
      if (aIsDynamic !== bIsDynamic) {
        return aIsDynamic ? 1 : -1;
      }
      
      // Longer paths first (more specific)
      return b.fullPath.length - a.fullPath.length;
    });
  }

  return {
    register(newRoutes: RouteDefinition[], basePath = '', pluginId?: string): void {
      routes.update(current => {
        const updated = [...current];
        
        for (const route of newRoutes) {
          const fullPath = normalizeFullPath(basePath, route.path);
          const registered: RegisteredRoute = {
            ...route,
            fullPath,
            pluginId
          };
          
          const existingIndex = updated.findIndex(r => r.fullPath === fullPath);
          if (existingIndex >= 0) {
            updated[existingIndex] = registered;
          } else {
            updated.push(registered);
          }
          
          // Register child routes recursively
          if (route.children) {
            this.register(route.children, fullPath, pluginId);
          }
        }
        
        return sortRoutes(updated);
      });
    },

    unregister(paths: string[]): void {
      routes.update(current => 
        current.filter(r => !paths.includes(r.fullPath))
      );
    },

    async navigate(path: string, options?: NavigateOptions): Promise<void> {
      // In a real SvelteKit app, use goto from $app/navigation
      // This is a fallback for non-SvelteKit contexts or SSR
      if (typeof window !== 'undefined') {
        const url = new URL(path, window.location.origin);
        
        if (options?.replace) {
          window.history.replaceState(options?.state ?? null, '', url);
        } else {
          window.history.pushState(options?.state ?? null, '', url);
        }
        
        // Update current route
        currentRoute = {
          path: url.pathname,
          params: {},
          query: Object.fromEntries(url.searchParams)
        };
        
        // Notify handlers
        routeChangeHandlers.forEach(handler => handler(currentRoute));
      }
    },

    getCurrentRoute(): RouteInfo {
      if (typeof window !== 'undefined') {
        const url = new URL(window.location.href);
        return {
          path: url.pathname,
          params: {},
          query: Object.fromEntries(url.searchParams)
        };
      }
      return currentRoute;
    },

    getAllRoutes(): RegisteredRoute[] {
      return get(routes);
    },

    getNavigation(): NavigationItem[] {
      return get(navigation);
    },

    getStore(): Readable<RegisteredRoute[]> {
      return { subscribe: routes.subscribe };
    },

    getNavigationStore(): Readable<NavigationItem[]> {
      return { subscribe: navigation.subscribe };
    },

    matchRoute(path: string) {
      const allRoutes = get(routes);
      
      for (const route of allRoutes) {
        const params = matchPath(route.fullPath, path);
        if (params !== null) {
          return { route, params };
        }
      }
      
      return null;
    },

    addNavigationItem(item: NavigationItem): void {
      navigation.update(current => {
        const existing = current.findIndex(n => n.id === item.id);
        if (existing >= 0) {
          current[existing] = item;
          return [...current];
        }
        
        // Insert in order
        const inserted = [...current, item];
        return inserted.sort((a, b) => (a.order ?? 0) - (b.order ?? 0));
      });
    },

    removeNavigationItem(id: string): void {
      navigation.update(current => current.filter(n => n.id !== id));
    },

    onRouteChange(handler: (route: RouteInfo) => void): Unsubscribe {
      routeChangeHandlers.add(handler);
      return () => routeChangeHandlers.delete(handler);
    }
  };
}
