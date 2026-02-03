import type { PluginDefinition, PluginContext } from '$lib/core/types';
import { manifest } from './manifest';
import { api } from './api';
import { messages } from './messages';
import { stores } from './stores';
import { routes } from './routes';

const menuPlugin: PluginDefinition = {
  manifest,
  api,
  messages,
  stores,
  routes,

  async setup(context: PluginContext) {
    context.logger.info('Menu plugin initializing...');
  },

  async teardown() {
    console.log('Menu plugin cleaning up...');
  }
};

export default menuPlugin;
