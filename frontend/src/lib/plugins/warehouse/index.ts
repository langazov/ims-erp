import type { PluginDefinition } from '$lib/core/types';
import { manifest } from './manifest';

const warehousePlugin: PluginDefinition = {
  manifest,

  async setup(context) {
    context.logger.info('warehouse plugin initializing...');
  },

  async teardown() {
    console.log('warehouse plugin cleaning up...');
  }
};

export default warehousePlugin;
