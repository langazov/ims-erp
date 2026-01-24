// src/lib/core/message-bus.ts

import type {
  Message,
  MessageHandler,
  MessageBusAPI,
  SendOptions,
  RequestOptions,
  SubscribeOptions,
  Unsubscribe
} from './types';

interface Subscription {
  type: string;
  handler: MessageHandler;
  options: SubscribeOptions;
  pluginId?: string;
}

export interface MessageBus extends MessageBusAPI {
  createScoped(pluginId: string): MessageBusAPI;
  getHistory(type?: string, limit?: number): Message[];
  clearHistory(): void;
}

function generateId(): string {
  return Math.random().toString(36).substring(2, 15) + 
         Math.random().toString(36).substring(2, 15);
}

export function createMessageBus(): MessageBus {
  const subscriptions: Map<string, Set<Subscription>> = new Map();
  const pendingRequests: Map<string, {
    resolve: (value: unknown) => void;
    reject: (error: Error) => void;
    timeout: ReturnType<typeof setTimeout>;
  }> = new Map();
  const messageHistory: Message[] = [];
  const MAX_HISTORY = 1000;

  function matchType(pattern: string, type: string): boolean {
    if (pattern === '*') return true;
    if (pattern === type) return true;
    
    const regexPattern = pattern
      .replace(/\./g, '\\.')
      .replace(/\*/g, '.*');
    return new RegExp(`^${regexPattern}$`).test(type);
  }

  function getSubscriptions(type: string): Subscription[] {
    const result: Subscription[] = [];
    
    for (const [pattern, subs] of subscriptions) {
      if (matchType(pattern, type)) {
        result.push(...subs);
      }
    }
    
    return result.sort((a, b) => (b.options.priority ?? 0) - (a.options.priority ?? 0));
  }

  function createMessage<T>(
    type: string,
    payload: T,
    source: string,
    options?: SendOptions
  ): Message<T> {
    return {
      id: generateId(),
      type,
      source,
      target: options?.target,
      payload,
      timestamp: Date.now(),
      meta: options?.meta
    };
  }

  function addToHistory(message: Message): void {
    messageHistory.push(message);
    if (messageHistory.length > MAX_HISTORY) {
      messageHistory.shift();
    }
  }

  async function dispatch(message: Message): Promise<unknown> {
    const subs = getSubscriptions(message.type);
    let result: unknown;

    for (const sub of subs) {
      if (sub.options.source) {
        const sources = Array.isArray(sub.options.source) 
          ? sub.options.source 
          : [sub.options.source];
        if (!sources.includes(message.source) && !sources.includes('*')) {
          continue;
        }
      }

      if (message.target && sub.pluginId && message.target !== sub.pluginId) {
        continue;
      }

      if (sub.handler.filter && !sub.handler.filter(message)) {
        continue;
      }

      try {
        result = await sub.handler.handle(message);
      } catch (error) {
        console.error(`Error in message handler for ${message.type}:`, error);
      }
    }

    return result;
  }

  const bus: MessageBus = {
    send<T>(type: string, payload: T, options?: SendOptions): void {
      const message = createMessage(type, payload, 'core', options);
      addToHistory(message);
      dispatch(message);
    },

    async request<TReq, TRes>(
      type: string,
      payload: TReq,
      options?: RequestOptions
    ): Promise<TRes> {
      const correlationId = generateId();
      const timeout = options?.timeout ?? 30000;

      const message = createMessage(type, payload, 'core', options);
      message.correlationId = correlationId;

      return new Promise<TRes>((resolve, reject) => {
        const timeoutHandle = setTimeout(() => {
          pendingRequests.delete(correlationId);
          reject(new Error(`Request timeout for ${type}`));
        }, timeout);

        pendingRequests.set(correlationId, {
          resolve: resolve as (value: unknown) => void,
          reject,
          timeout: timeoutHandle
        });

        addToHistory(message);
        dispatch(message).then(result => {
          const pending = pendingRequests.get(correlationId);
          if (pending) {
            clearTimeout(pending.timeout);
            pendingRequests.delete(correlationId);
            pending.resolve(result);
          }
        });
      });
    },

    subscribe(
      type: string,
      handler: MessageHandler,
      options: SubscribeOptions = {}
    ): Unsubscribe {
      const subscription: Subscription = { type, handler, options };

      if (!subscriptions.has(type)) {
        subscriptions.set(type, new Set());
      }
      subscriptions.get(type)!.add(subscription);

      return () => {
        subscriptions.get(type)?.delete(subscription);
        if (subscriptions.get(type)?.size === 0) {
          subscriptions.delete(type);
        }
      };
    },

    async once<T>(type: string, options?: SubscribeOptions): Promise<Message<T>> {
      return new Promise((resolve) => {
        const unsubscribe = this.subscribe(
          type,
          {
            handle: (message) => {
              unsubscribe();
              resolve(message as Message<T>);
            }
          },
          options
        );
      });
    },

    createScoped(pluginId: string): MessageBusAPI {
      return {
        send: <T>(type: string, payload: T, options?: SendOptions) => {
          const message = createMessage(type, payload, pluginId, options);
          addToHistory(message);
          dispatch(message);
        },

        request: async <TReq, TRes>(
          type: string,
          payload: TReq,
          options?: RequestOptions
        ): Promise<TRes> => {
          const correlationId = generateId();
          const timeout = options?.timeout ?? 30000;

          const message = createMessage(type, payload, pluginId, options);
          message.correlationId = correlationId;

          return new Promise<TRes>((resolve, reject) => {
            const timeoutHandle = setTimeout(() => {
              pendingRequests.delete(correlationId);
              reject(new Error(`Request timeout for ${type}`));
            }, timeout);

            pendingRequests.set(correlationId, {
              resolve: resolve as (value: unknown) => void,
              reject,
              timeout: timeoutHandle
            });

            addToHistory(message);
            dispatch(message).then(result => {
              const pending = pendingRequests.get(correlationId);
              if (pending) {
                clearTimeout(pending.timeout);
                pendingRequests.delete(correlationId);
                pending.resolve(result);
              }
            });
          });
        },

        subscribe: (
          type: string,
          handler: MessageHandler,
          options: SubscribeOptions = {}
        ): Unsubscribe => {
          const subscription: Subscription = { 
            type, 
            handler, 
            options,
            pluginId 
          };

          if (!subscriptions.has(type)) {
            subscriptions.set(type, new Set());
          }
          subscriptions.get(type)!.add(subscription);

          return () => {
            subscriptions.get(type)?.delete(subscription);
          };
        },

        once: async <T>(type: string, options?: SubscribeOptions): Promise<Message<T>> => {
          return new Promise((resolve) => {
            const unsubscribe = bus.createScoped(pluginId).subscribe(
              type,
              {
                handle: (message) => {
                  unsubscribe();
                  resolve(message as Message<T>);
                }
              },
              options
            );
          });
        }
      };
    },

    getHistory(type?: string, limit = 100): Message[] {
      let history = [...messageHistory];
      
      if (type) {
        history = history.filter(m => matchType(type, m.type));
      }
      
      return history.slice(-limit);
    },

    clearHistory(): void {
      messageHistory.length = 0;
    }
  };

  return bus;
}
