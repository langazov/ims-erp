import { writable, derived, get } from 'svelte/store';
import { browser } from '$app/environment';
import { auth } from './auth';

export type WSStatus = 'disconnected' | 'connecting' | 'connected' | 'error';

export interface WSEvent {
  type: string;
  payload: unknown;
  tenantId: string;
  timestamp: string;
}

type EventHandler = (event: WSEvent) => void;

function createWebSocketStore() {
  const status = writable<WSStatus>('disconnected');
  const lastEvent = writable<WSEvent | null>(null);
  
  let ws: WebSocket | null = null;
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  let reconnectDelay = 1000;
  const handlers = new Map<string, Set<EventHandler>>();

  function connect() {
    if (!browser) return;
    const authState = get(auth);
    if (!authState.isAuthenticated || !authState.tokens) return;

    const WS_URL = import.meta.env.VITE_WS_URL || 'ws://localhost:8080';
    const token = authState.tokens.accessToken;
    const tenantId = authState.user?.tenantId || '';

    try {
      status.set('connecting');
      ws = new WebSocket(`${WS_URL}/ws/events?token=${token}&tenantId=${tenantId}`);

      ws.onopen = () => {
        status.set('connected');
        reconnectDelay = 1000;
      };

      ws.onmessage = (event) => {
        try {
          const data: WSEvent = JSON.parse(event.data);
          lastEvent.set(data);
          // Route to registered handlers
          const typeHandlers = handlers.get(data.type);
          if (typeHandlers) typeHandlers.forEach(h => h(data));
          const wildcardHandlers = handlers.get('*');
          if (wildcardHandlers) wildcardHandlers.forEach(h => h(data));
        } catch (e) {
          console.warn('[WS] Failed to parse event:', e);
        }
      };

      ws.onclose = () => {
        status.set('disconnected');
        ws = null;
        scheduleReconnect();
      };

      ws.onerror = () => {
        status.set('error');
      };
    } catch (e) {
      status.set('error');
      scheduleReconnect();
    }
  }

  function scheduleReconnect() {
    if (reconnectTimer) clearTimeout(reconnectTimer);
    reconnectTimer = setTimeout(() => {
      reconnectDelay = Math.min(reconnectDelay * 2, 30000);
      connect();
    }, reconnectDelay);
  }

  function disconnect() {
    if (reconnectTimer) clearTimeout(reconnectTimer);
    if (ws) {
      ws.close();
      ws = null;
    }
    status.set('disconnected');
  }

  function on(eventType: string, handler: EventHandler): () => void {
    if (!handlers.has(eventType)) handlers.set(eventType, new Set());
    handlers.get(eventType)!.add(handler);
    return () => handlers.get(eventType)?.delete(handler);
  }

  function off(eventType: string, handler: EventHandler) {
    handlers.get(eventType)?.delete(handler);
  }

  return {
    status: { subscribe: status.subscribe },
    lastEvent: { subscribe: lastEvent.subscribe },
    connect,
    disconnect,
    on,
    off,
  };
}

export const wsStore = createWebSocketStore();
