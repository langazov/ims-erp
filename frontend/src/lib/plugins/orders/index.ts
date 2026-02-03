import type { PluginDefinition } from '$lib/core/types';
import { manifest } from './manifest';

const ordersPlugin: PluginDefinition = {
  manifest,

  async setup(context) {
    context.logger.info('orders plugin initializing...');
  },

  async teardown() {
    console.log('orders plugin cleaning up...');
  }
};

export default ordersPlugin;
