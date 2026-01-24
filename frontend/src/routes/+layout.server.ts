// src/routes/+layout.server.ts

import type { LayoutServerLoad } from './$types';
import type { PluginManifestEntry } from '$lib/core';

export const load: LayoutServerLoad = async ({ fetch, url }) => {
  const manifests: PluginManifestEntry[] = [];
  
  // ============================================================================
  // Built-in Plugins
  // ============================================================================
  // These are bundled with the application and always available
  
  const builtinPlugins: PluginManifestEntry[] = [
    {
      id: 'dashboard',
      source: { type: 'builtin', path: 'dashboard' },
      enabled: true
    },
    // Add more built-in plugins here
    // {
    //   id: 'settings',
    //   source: { type: 'builtin', path: 'settings' },
    //   enabled: true
    // },
    // {
    //   id: 'analytics',
    //   source: { type: 'builtin', path: 'analytics' },
    //   enabled: true
    // },
  ];
  
  manifests.push(...builtinPlugins);
  
  // ============================================================================
  // External Plugins (from API)
  // ============================================================================
  // Load additional plugins from your backend or a plugin registry
  
  try {
    const response = await fetch('/api/plugins/manifest');
    if (response.ok) {
      const externalManifests: PluginManifestEntry[] = await response.json();
      manifests.push(...externalManifests);
    }
  } catch (error) {
    console.warn('[Server] Failed to load external plugin manifests:', error);
  }
  
  // ============================================================================
  // Environment-specific Plugins
  // ============================================================================
  // Load plugins based on environment variables
  
  const envPlugins = process.env.EXTERNAL_PLUGINS?.split(',').filter(Boolean) ?? [];
  
  for (const pluginUrl of envPlugins) {
    try {
      // Fetch manifest from external plugin URL
      const manifestUrl = `${pluginUrl.trim()}/manifest.json`;
      const response = await fetch(manifestUrl);
      
      if (response.ok) {
        const manifest = await response.json();
        manifests.push({
          id: manifest.id,
          source: { type: 'external', url: `${pluginUrl.trim()}/index.js` },
          enabled: manifest.enabled ?? true
        });
      }
    } catch (error) {
      console.warn(`[Server] Failed to load plugin from ${pluginUrl}:`, error);
    }
  }
  
  // ============================================================================
  // User-specific Plugins
  // ============================================================================
  // In a real app, you might load user-enabled plugins from a database
  
  // const userPlugins = await getUserPlugins(locals.user?.id);
  // manifests.push(...userPlugins);
  
  // ============================================================================
  // Filter and Sort
  // ============================================================================
  
  // Remove duplicates (keep first occurrence)
  const uniqueManifests = manifests.reduce((acc, manifest) => {
    if (!acc.find(m => m.id === manifest.id)) {
      acc.push(manifest);
    }
    return acc;
  }, [] as PluginManifestEntry[]);
  
  // Filter disabled plugins
  const enabledManifests = uniqueManifests.filter(m => m.enabled !== false);
  
  console.log(`[Server] Loading ${enabledManifests.length} plugins:`, 
    enabledManifests.map(m => m.id).join(', '));
  
  return {
    pluginManifests: enabledManifests
  };
};
