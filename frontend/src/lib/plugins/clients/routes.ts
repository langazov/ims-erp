import type { PluginRoutes } from '$lib/core/types';

export const routes: PluginRoutes = {
  basePath: '/clients',

  routes: [
    {
      path: '/',
      component: null as any,
      meta: {
        title: 'Clients',
        description: 'Manage clients',
        icon: 'users',
        order: 2
      }
    },
    {
      path: '/new',
      component: null as any,
      meta: {
        title: 'New Client',
        description: 'Create a new client',
        order: 1
      }
    },
    {
      path: '/:id',
      component: null as any,
      meta: {
        title: 'Client Details',
        description: 'View client details',
        order: 3
      }
    },
    {
      path: '/:id/edit',
      component: null as any,
      meta: {
        title: 'Edit Client',
        description: 'Edit client information',
        order: 4
      }
    }
  ],

  navigation: [
    {
      id: 'clients',
      label: 'Clients',
      path: '/clients',
      icon: 'users',
      order: 2
    }
  ]
};
