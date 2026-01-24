// src/lib/core/state-manager.ts

import { writable, get, type Readable, type Writable } from 'svelte/store';
import type { StateManagerAPI, Unsubscribe } from './types';

export interface StateManager extends StateManagerAPI {
  /** Get all state keys */
  keys(): string[];
  
  /** Check if state exists */
  has(key: string): boolean;
  
  /** Delete state */
  delete(key: string): void;
  
  /** Clear all state */
  clear(): void;
  
  /** Get all stores */
  getAllStores(): Map<string, Readable<unknown>>;
}

export function createStateManager(): StateManager {
  const stores: Map<string, Writable<unknown>> = new Map();
  const subscribers: Map<string, Set<(value: unknown) => void>> = new Map();

  function getOrCreateStore<T>(key: string, initialValue?: T): Writable<T> {
    if (!stores.has(key)) {
      stores.set(key, writable(initialValue));
    }
    return stores.get(key) as Writable<T>;
  }

  function notifySubscribers(key: string, value: unknown): void {
    const subs = subscribers.get(key);
    if (subs) {
      subs.forEach(handler => handler(value));
    }
  }

  return {
    get<T>(key: string): T | undefined {
      const store = stores.get(key);
      if (!store) return undefined;
      return get(store) as T;
    },

    set<T>(key: string, value: T): void {
      const store = getOrCreateStore<T>(key, value);
      store.set(value);
      notifySubscribers(key, value);
    },

    subscribe<T>(key: string, handler: (value: T) => void): Unsubscribe {
      if (!subscribers.has(key)) {
        subscribers.set(key, new Set());
      }
      subscribers.get(key)!.add(handler as (value: unknown) => void);

      // Also subscribe to the store
      const store = getOrCreateStore<T>(key);
      const unsubStore = store.subscribe((value) => {
        handler(value as T);
      });

      return () => {
        subscribers.get(key)?.delete(handler as (value: unknown) => void);
        unsubStore();
      };
    },

    getStore<T>(key: string): Readable<T> {
      const store = getOrCreateStore<T>(key);
      return { subscribe: store.subscribe } as Readable<T>;
    },

    keys(): string[] {
      return Array.from(stores.keys());
    },

    has(key: string): boolean {
      return stores.has(key);
    },

    delete(key: string): void {
      stores.delete(key);
      subscribers.delete(key);
    },

    clear(): void {
      stores.clear();
      subscribers.clear();
    },

    getAllStores(): Map<string, Readable<unknown>> {
      const result = new Map<string, Readable<unknown>>();
      for (const [key, store] of stores) {
        result.set(key, { subscribe: store.subscribe });
      }
      return result;
    }
  };
}

// src/lib/core/event-emitter.ts

import type { EventEmitterAPI, EventHandler, Unsubscribe } from './types';

export interface EventEmitter extends EventEmitterAPI {
  /** Remove all listeners for an event */
  removeAllListeners(event?: string): void;
  
  /** Get listener count for an event */
  listenerCount(event: string): number;
  
  /** Get all event names */
  eventNames(): string[];
}

export function createEventEmitter(): EventEmitter {
  const listeners: Map<string, Set<EventHandler>> = new Map();
  const onceListeners: Map<string, Set<EventHandler>> = new Map();

  return {
    on(event: string, handler: EventHandler): Unsubscribe {
      if (!listeners.has(event)) {
        listeners.set(event, new Set());
      }
      listeners.get(event)!.add(handler);

      return () => {
        listeners.get(event)?.delete(handler);
        if (listeners.get(event)?.size === 0) {
          listeners.delete(event);
        }
      };
    },

    off(event: string, handler: EventHandler): void {
      listeners.get(event)?.delete(handler);
      onceListeners.get(event)?.delete(handler);
    },

    emit(event: string, data?: unknown): void {
      // Regular listeners
      const eventListeners = listeners.get(event);
      if (eventListeners) {
        eventListeners.forEach(handler => {
          try {
            handler(data);
          } catch (error) {
            console.error(`Error in event handler for ${event}:`, error);
          }
        });
      }

      // Once listeners
      const onceEventListeners = onceListeners.get(event);
      if (onceEventListeners) {
        onceEventListeners.forEach(handler => {
          try {
            handler(data);
          } catch (error) {
            console.error(`Error in once handler for ${event}:`, error);
          }
        });
        onceListeners.delete(event);
      }

      // Wildcard listeners
      const wildcardListeners = listeners.get('*');
      if (wildcardListeners) {
        wildcardListeners.forEach(handler => {
          try {
            handler({ event, data });
          } catch (error) {
            console.error(`Error in wildcard handler:`, error);
          }
        });
      }
    },

    removeAllListeners(event?: string): void {
      if (event) {
        listeners.delete(event);
        onceListeners.delete(event);
      } else {
        listeners.clear();
        onceListeners.clear();
      }
    },

    listenerCount(event: string): number {
      return (listeners.get(event)?.size ?? 0) + (onceListeners.get(event)?.size ?? 0);
    },

    eventNames(): string[] {
      const events = new Set([...listeners.keys(), ...onceListeners.keys()]);
      return Array.from(events);
    }
  };
}
