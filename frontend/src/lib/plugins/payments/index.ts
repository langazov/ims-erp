import type { PluginDefinition } from '$lib/core/types';
import { manifest } from './manifest';

const paymentsPlugin: PluginDefinition = {
  manifest,

  async setup(context) {
    context.logger.info('payments plugin initializing...');
  },

  async teardown() {
    console.log('payments plugin cleaning up...');
  }
};

export default paymentsPlugin;
