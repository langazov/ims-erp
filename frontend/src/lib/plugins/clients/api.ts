import type { PluginAPI } from '$lib/core/types';

export const api: PluginAPI = {
  version: '1.0.0',

  methods: {
    getClients: {
      handler: async (input, context) => {
        return { success: true, clients: [] };
      },
      description: 'Get all clients'
    },

    getClient: {
      handler: async (input: { id: string }, context) => {
        return { success: true, client: null };
      },
      description: 'Get a client by ID'
    },

    createClient: {
      handler: async (input, context) => {
        return { success: true, client: null };
      },
      description: 'Create a new client'
    },

    updateClient: {
      handler: async (input: { id: string; data: unknown }, context) => {
        return { success: true, client: null };
      },
      description: 'Update a client'
    },

    deleteClient: {
      handler: async (input: { id: string }, context) => {
        return { success: true };
      },
      description: 'Delete a client'
    }
  },

  docs: {
    title: 'Clients API',
    description: 'API for managing clients',
    methods: {
      getClients: {
        summary: 'Retrieve all clients',
        examples: [{
          title: 'Get all clients',
          input: {},
          output: { success: true, clients: [] }
        }]
      }
    }
  }
};
