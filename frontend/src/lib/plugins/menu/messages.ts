import type { PluginMessages } from '$lib/core/types';

export const messages: PluginMessages = {
  handlers: {
    'module:selected': {
      handle: async (message) => {
        console.log('Module selected:', message.payload);
        return { received: true, type: 'module:selected' };
      }
    },
    
    'menu:toggle': {
      handle: async (message) => {
        console.log('Menu toggle:', message.payload);
        return { received: true, type: 'menu:toggle' };
      }
    }
  },

  schemas: {
    'module:selected': {
      type: 'menu:module:selected',
      payload: {
        type: 'object',
        properties: {
          moduleId: { type: 'string' },
          moduleName: { type: 'string' }
        },
        required: ['moduleId']
      },
      description: 'Notification that a module was selected'
    },
    
    'menu:toggle': {
      type: 'menu:toggle',
      payload: {
        type: 'object',
        properties: {
          expanded: { type: 'boolean' },
          category: { type: 'string' }
        }
      },
      description: 'Notification that menu category was toggled'
    }
  }
};
