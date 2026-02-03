import type { PluginDefinition } from '$lib/core/types';
import { manifest } from './manifest';

const inventoryPlugin: PluginDefinition = {
  manifest,

  async setup(context) {
    context.logger.info('inventory plugin initializing...');
  },

  async teardown() {
    console.log('inventory plugin cleaning up...');
  }
};

export default inventoryPlugin;
