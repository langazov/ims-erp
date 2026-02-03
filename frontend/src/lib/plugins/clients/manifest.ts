import type { PluginManifest } from '$lib/core/types';

export const manifest: PluginManifest = {
  id: 'clients',
  name: 'Clients',
  version: '1.0.0',
  description: 'Client management for maintaining customer relationships',
  author: 'IMS Team',
  icon: 'users',

  permissions: [
    'storage:read',
    'storage:write',
    'routes:register'
  ],

  lifecycle: {
    onEnable: () => console.log('Clients plugin enabled'),
    onDisable: () => console.log('Clients plugin disabled')
  },

  priority: 10,
  enabled: true
};
