import type { PluginDefinition } from '$lib/core/types';
import { manifest } from './manifest';

const settingsPlugin: PluginDefinition = {
  manifest,

  async setup(context) {
    context.logger.info('settings plugin initializing...');
  },

  async teardown() {
    console.log('settings plugin cleaning up...');
  }
};

export default settingsPlugin;
