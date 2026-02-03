import { writable, derived } from 'svelte/store';
import type { PluginStores } from '$lib/core/types';
import type { Client, ClientFilter, ClientListResponse } from '$lib/shared/api/clients';
import * as api from '$lib/shared/api/clients';

interface ClientsState {
  clients: Client[];
  selectedClient: Client | null;
  loading: boolean;
  error: string | null;
  filter: ClientFilter;
  pagination: {
    page: number;
    pageSize: number;
    total: number;
    totalPages: number;
  };
}

const initialState: ClientsState = {
  clients: [],
  selectedClient: null,
  loading: false,
  error: null,
  filter: {},
  pagination: {
    page: 1,
    pageSize: 20,
    total: 0,
    totalPages: 0
  }
};

function createClientsStore() {
  const { subscribe, set, update } = writable<ClientsState>(initialState);

  return {
    subscribe,

    async loadClients(filter?: ClientFilter) {
      update(state => ({ ...state, loading: true, error: null }));

      try {
        const response = await api.getClients(filter);
        update(state => ({
          ...state,
          clients: response.data,
          loading: false,
          pagination: {
            page: response.page,
            pageSize: response.pageSize,
            total: response.total,
            totalPages: response.totalPages
          },
          filter: filter || state.filter
        }));
      } catch (error) {
        update(state => ({
          ...state,
          loading: false,
          error: error instanceof Error ? error.message : 'Failed to load clients'
        }));
      }
    },

    async loadClient(id: string) {
      update(state => ({ ...state, loading: true, error: null }));

      try {
        const client = await api.getClientById(id);
        update(state => ({
          ...state,
          selectedClient: client,
          loading: false
        }));
        return client;
      } catch (error) {
        update(state => ({
          ...state,
          loading: false,
          error: error instanceof Error ? error.message : 'Failed to load client'
        }));
        throw error;
      }
    },

    async createClient(data: Parameters<typeof api.createClient>[0]) {
      update(state => ({ ...state, loading: true, error: null }));

      try {
        const client = await api.createClient(data);
        update(state => ({
          ...state,
          clients: [client, ...state.clients],
          loading: false
        }));
        return client;
      } catch (error) {
        update(state => ({
          ...state,
          loading: false,
          error: error instanceof Error ? error.message : 'Failed to create client'
        }));
        throw error;
      }
    },

    async updateClient(id: string, data: Parameters<typeof api.updateClient>[1]) {
      update(state => ({ ...state, loading: true, error: null }));

      try {
        const client = await api.updateClient(id, data);
        update(state => ({
          ...state,
          clients: state.clients.map(c => c.id === id ? client : c),
          selectedClient: state.selectedClient?.id === id ? client : state.selectedClient,
          loading: false
        }));
        return client;
      } catch (error) {
        update(state => ({
          ...state,
          loading: false,
          error: error instanceof Error ? error.message : 'Failed to update client'
        }));
        throw error;
      }
    },

    async deleteClient(id: string) {
      update(state => ({ ...state, loading: true, error: null }));

      try {
        await api.deleteClient(id);
        update(state => ({
          ...state,
          clients: state.clients.filter(c => c.id !== id),
          selectedClient: state.selectedClient?.id === id ? null : state.selectedClient,
          loading: false
        }));
      } catch (error) {
        update(state => ({
          ...state,
          loading: false,
          error: error instanceof Error ? error.message : 'Failed to delete client'
        }));
        throw error;
      }
    },

    selectClient(client: Client | null) {
      update(state => ({ ...state, selectedClient: client }));
    },

    setFilter(filter: ClientFilter) {
      update(state => ({ ...state, filter }));
    },

    setPage(page: number) {
      update(state => ({ ...state, filter: { ...state.filter, page } }));
    },

    clearError() {
      update(state => ({ ...state, error: null }));
    },

    reset() {
      set(initialState);
    }
  };
}

export const clientsStore = createClientsStore();

export const stores: PluginStores = {
  readable: {
    clients: derived(clientsStore, $state => $state.clients),
    selectedClient: derived(clientsStore, $state => $state.selectedClient),
    loading: derived(clientsStore, $state => $state.loading),
    error: derived(clientsStore, $state => $state.error),
    filter: derived(clientsStore, $state => $state.filter),
    pagination: derived(clientsStore, $state => $state.pagination)
  },
  writable: {
    clients: {
      subscribe: (run) => clientsStore.subscribe(state => run(state.clients)),
      set: (value) => clientsStore.update(state => ({ ...state, clients: value })),
      update: (fn) => clientsStore.update(state => ({ ...state, clients: fn(state.clients) }))
    }
  }
};
