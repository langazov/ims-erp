import type { PluginMessages } from '$lib/core/types';

export const messages: PluginMessages = {
  handlers: {
    'client:created': {
      handle: async (message) => {
        console.log('Client created:', message.payload);
        return { received: true, type: 'client:created' };
      }
    },
    
    'client:updated': {
      handle: async (message) => {
        console.log('Client updated:', message.payload);
        return { received: true, type: 'client:updated' };
      }
    },
    
    'client:deleted': {
      handle: async (message) => {
        console.log('Client deleted:', message.payload);
        return { received: true, type: 'client:deleted' };
      }
    }
  },

  schemas: {
    'client:created': {
      type: 'clients:created',
      payload: {
        type: 'object',
        properties: {
          clientId: { type: 'string' },
          name: { type: 'string' }
        },
        required: ['clientId', 'name']
      },
      description: 'Notification that a new client was created'
    },
    
    'client:updated': {
      type: 'clients:updated',
      payload: {
        type: 'object',
        properties: {
          clientId: { type: 'string' },
          changes: { type: 'object' }
        },
        required: ['clientId']
      },
      description: 'Notification that a client was updated'
    },
    
    'client:deleted': {
      type: 'clients:deleted',
      payload: {
        type: 'object',
        properties: {
          clientId: { type: 'string' }
        },
        required: ['clientId']
      },
      description: 'Notification that a client was deleted'
    }
  }
};
