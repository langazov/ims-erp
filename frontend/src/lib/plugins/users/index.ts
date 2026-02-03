import type { PluginDefinition } from '$lib/core/types';
import { manifest } from './manifest';

const usersPlugin: PluginDefinition = {
  manifest,

  async setup(context) {
    context.logger.info('Users plugin initializing...');
  },

  async teardown() {
    console.log('Users plugin cleaning up...');
  }
};

export default usersPlugin;
