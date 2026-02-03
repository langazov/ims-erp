import type { PluginDefinition } from '$lib/core/types';
import { manifest } from './manifest';

const documentsPlugin: PluginDefinition = {
  manifest,

  async setup(context) {
    context.logger.info('documents plugin initializing...');
  },

  async teardown() {
    console.log('documents plugin cleaning up...');
  }
};

export default documentsPlugin;
