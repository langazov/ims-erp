import type { PluginRoutes } from '$lib/core/types';

export const routes: PluginRoutes = {
  basePath: '/modules',

  routes: [
    {
      path: '/',
      component: null as any,
      meta: {
        title: 'Modules',
        description: 'Installed modules and navigation',
        icon: 'menu',
        order: 0
      }
    },
    {
      path: '/:id',
      component: null as any,
      meta: {
        title: 'Module Details',
        description: 'View module information',
        order: 1
      }
    }
  ],

  navigation: [
    {
      id: 'modules',
      label: 'Modules',
      path: '/modules',
      icon: 'grid',
      order: 0
    }
  ]
};
