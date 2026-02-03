import type { PluginManifest } from '$lib/core/types';

export const manifest: PluginManifest = {
  id: 'users',
  name: 'Users',
  version: '1.0.0',
  description: 'User management and access control',
  author: 'IMS Team',
  icon: 'user',
  enabled: true,
  priority: 10
};
