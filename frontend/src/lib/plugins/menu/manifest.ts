import type { PluginManifest } from '$lib/core/types';

export const manifest: PluginManifest = {
  id: 'menu',
  name: 'Modules',
  version: '1.0.0',
  description: 'Module menu and navigation for accessing installed plugins',
  author: 'IMS Team',
  icon: 'grid',

  permissions: [
    'storage:read',
    'routes:register'
  ],

  lifecycle: {
    onEnable: () => console.log('Menu plugin enabled'),
    onDisable: () => console.log('Menu plugin disabled')
  },

  priority: 1,
  enabled: true
};
