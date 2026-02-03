import type { PluginDefinition } from '$lib/core/types';
import { manifest } from './manifest';

const productsPlugin: PluginDefinition = {
  manifest,

  async setup(context) {
    context.logger.info('products plugin initializing...');
  },

  async teardown() {
    console.log('products plugin cleaning up...');
  }
};

export default productsPlugin;
