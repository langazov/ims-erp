import type { PluginDefinition } from '$lib/core/types';
import { manifest } from './manifest';
import { api } from './api';
import { messages } from './messages';
import { stores } from './stores';
import { routes } from './routes';

const clientsPlugin: PluginDefinition = {
  manifest,
  api,
  messages,
  stores,
  routes,

  async setup(context) {
    context.logger.info('Clients plugin initializing...');
  },

  async teardown() {
    console.log('Clients plugin cleaning up...');
  }
};

export default clientsPlugin;
