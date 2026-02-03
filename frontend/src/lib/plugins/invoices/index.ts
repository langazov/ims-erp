import type { PluginDefinition } from '$lib/core/types';
import { manifest } from './manifest';

const invoicesPlugin: PluginDefinition = {
  manifest,

  async setup(context) {
    context.logger.info('invoices plugin initializing...');
  },

  async teardown() {
    console.log('invoices plugin cleaning up...');
  }
};

export default invoicesPlugin;
