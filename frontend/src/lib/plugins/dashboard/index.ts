// src/lib/plugins/dashboard/index.ts

import type { PluginDefinition, PluginContext } from '$lib/core/types';
import { writable, derived, type Writable } from 'svelte/store';

// ============================================================================
// Types
// ============================================================================

interface Widget {
  id: string;
  type: string;
  title: string;
  config?: Record<string, unknown>;
}

interface WidgetPosition {
  id: string;
  x: number;
  y: number;
  w: number;
  h: number;
}

interface Layout {
  columns: number;
  widgets: WidgetPosition[];
}

interface DashboardPreferences {
  theme: 'light' | 'dark' | 'auto';
  refreshInterval: number;
  compactMode: boolean;
}

// ============================================================================
// Stores
// ============================================================================

const widgets: Writable<Widget[]> = writable([
  { id: 'stats', type: 'statistics', title: 'Statistics' },
  { id: 'chart', type: 'chart', title: 'Activity Chart' },
  { id: 'recent', type: 'list', title: 'Recent Items' }
]);

const layout: Writable<Layout | null> = writable({
  columns: 3,
  widgets: [
    { id: 'stats', x: 0, y: 0, w: 2, h: 1 },
    { id: 'chart', x: 2, y: 0, w: 1, h: 2 },
    { id: 'recent', x: 0, y: 1, w: 2, h: 1 }
  ]
});

const preferences: Writable<DashboardPreferences> = writable({
  theme: 'auto',
  refreshInterval: 30,
  compactMode: false
});

const isLoading = writable(false);

const activeWidgets = derived(
  [widgets, layout],
  ([$widgets, $layout]) => {
    if (!$layout) return [];
    return $layout.widgets
      .map(w => {
        const widget = $widgets.find(widget => widget.id === w.id);
        return widget ? { ...widget, position: w } : null;
      })
      .filter((w): w is Widget & { position: WidgetPosition } => w !== null);
  }
);

// ============================================================================
// Plugin Definition
// ============================================================================

const dashboardPlugin: PluginDefinition = {
  // Manifest
  manifest: {
    id: 'dashboard',
    name: 'Dashboard',
    version: '1.0.0',
    description: 'Main dashboard with customizable widgets and layouts',
    author: 'Your Team',
    
    permissions: [
      'storage:read',
      'storage:write',
      'plugins:communicate',
      'routes:register'
    ],
    
    lifecycle: {
      onEnable: () => console.log('[Dashboard] Plugin enabled'),
      onDisable: () => console.log('[Dashboard] Plugin disabled'),
      onError: (error) => console.error('[Dashboard] Plugin error:', error)
    },
    
    configSchema: {
      type: 'object',
      properties: {
        defaultLayout: {
          type: 'string',
          title: 'Default Layout',
          description: 'The default dashboard layout',
          enum: ['grid', 'list', 'compact'],
          default: 'grid'
        },
        refreshInterval: {
          type: 'number',
          title: 'Refresh Interval',
          description: 'Data refresh interval in seconds',
          default: 30
        },
        maxWidgets: {
          type: 'number',
          title: 'Maximum Widgets',
          description: 'Maximum number of widgets allowed',
          default: 12
        }
      }
    },
    
    priority: 100,
    enabled: true
  },

  // API
  api: {
    version: '1.0.0',
    
    methods: {
      getWidgets: {
        handler: async () => {
          let currentWidgets: Widget[] = [];
          widgets.subscribe(w => currentWidgets = w)();
          return currentWidgets;
        },
        description: 'Get available dashboard widgets',
        outputSchema: {
          type: 'array',
          items: {
            type: 'object',
            properties: {
              id: { type: 'string' },
              type: { type: 'string' },
              title: { type: 'string' }
            }
          }
        }
      },
      
      getLayout: {
        handler: async (input: { userId?: string }) => {
          let currentLayout: Layout | null = null;
          layout.subscribe(l => currentLayout = l)();
          return currentLayout;
        },
        inputSchema: {
          type: 'object',
          properties: {
            userId: { type: 'string' }
          }
        },
        description: 'Get dashboard layout for a user'
      },
      
      saveLayout: {
        handler: async (input: { layout: Layout }) => {
          layout.set(input.layout);
          return { success: true };
        },
        permissions: ['storage:write'],
        description: 'Save dashboard layout'
      },
      
      addWidget: {
        handler: async (input: { widget: Widget; position?: WidgetPosition }) => {
          widgets.update(w => [...w, input.widget]);
          
          if (input.position) {
            layout.update(l => {
              if (!l) return l;
              return {
                ...l,
                widgets: [...l.widgets, input.position!]
              };
            });
          }
          
          return { success: true, widgetId: input.widget.id };
        },
        description: 'Add a new widget to the dashboard'
      },
      
      removeWidget: {
        handler: async (input: { widgetId: string }) => {
          widgets.update(w => w.filter(widget => widget.id !== input.widgetId));
          layout.update(l => {
            if (!l) return l;
            return {
              ...l,
              widgets: l.widgets.filter(w => w.id !== input.widgetId)
            };
          });
          return { success: true };
        },
        description: 'Remove a widget from the dashboard'
      },
      
      getPreferences: {
        handler: async () => {
          let prefs: DashboardPreferences | null = null;
          preferences.subscribe(p => prefs = p)();
          return prefs;
        },
        description: 'Get dashboard preferences'
      },
      
      updatePreferences: {
        handler: async (input: Partial<DashboardPreferences>) => {
          preferences.update(p => ({ ...p, ...input }));
          return { success: true };
        },
        description: 'Update dashboard preferences'
      }
    },
    
    docs: {
      title: 'Dashboard API',
      description: 'API for managing dashboard widgets, layouts, and preferences',
      methods: {
        getWidgets: {
          summary: 'Retrieve available widgets',
          examples: [{
            title: 'Get all widgets',
            input: {},
            output: [{ id: 'stats', type: 'statistics', title: 'Statistics' }]
          }]
        },
        getLayout: {
          summary: 'Get the current dashboard layout',
          examples: [{
            title: 'Get layout',
            input: {},
            output: { columns: 3, widgets: [{ id: 'stats', x: 0, y: 0, w: 2, h: 1 }] }
          }]
        }
      }
    }
  },

  // Messages
  messages: {
    handlers: {
      'widget:refresh': {
        handle: async (message) => {
          const { widgetId } = message.payload as { widgetId: string };
          console.log(`[Dashboard] Refreshing widget: ${widgetId}`);
          
          // Simulate refresh
          isLoading.set(true);
          await new Promise(resolve => setTimeout(resolve, 500));
          isLoading.set(false);
          
          return { refreshed: true, widgetId, timestamp: Date.now() };
        },
        priority: 10
      },
      
      'layout:changed': {
        handle: async (message) => {
          const { newLayout } = message.payload as { newLayout: Layout };
          console.log('[Dashboard] Layout changed:', newLayout);
          layout.set(newLayout);
        }
      },
      
      'preferences:update': {
        handle: async (message) => {
          const updates = message.payload as Partial<DashboardPreferences>;
          preferences.update(p => ({ ...p, ...updates }));
        }
      }
    },
    
    schemas: {
      'widget:refresh': {
        type: 'dashboard:widget:refresh',
        payload: {
          type: 'object',
          properties: {
            widgetId: { type: 'string' }
          },
          required: ['widgetId']
        },
        response: {
          type: 'object',
          properties: {
            refreshed: { type: 'boolean' },
            widgetId: { type: 'string' },
            timestamp: { type: 'number' }
          }
        },
        expectsResponse: true,
        description: 'Request a widget to refresh its data'
      },
      
      'layout:changed': {
        type: 'dashboard:layout:changed',
        payload: {
          type: 'object',
          properties: {
            newLayout: { type: 'object' }
          }
        },
        description: 'Notification that the dashboard layout has changed'
      },
      
      'widget:added': {
        type: 'dashboard:widget:added',
        payload: {
          type: 'object',
          properties: {
            widgetId: { type: 'string' },
            widgetType: { type: 'string' },
            position: { type: 'object' }
          }
        },
        description: 'Notification that a widget was added to the dashboard'
      },
      
      'widget:removed': {
        type: 'dashboard:widget:removed',
        payload: {
          type: 'object',
          properties: {
            widgetId: { type: 'string' }
          }
        },
        description: 'Notification that a widget was removed from the dashboard'
      }
    },
    
    subscriptions: [
      {
        type: 'analytics:*',
        handler: {
          handle: (message) => {
            console.log('[Dashboard] Analytics event received:', message.type, message.payload);
          }
        }
      },
      {
        type: 'user:preferences:changed',
        source: 'settings',
        handler: {
          handle: (message) => {
            console.log('[Dashboard] User preferences changed, may need to update');
          }
        }
      }
    ]
  },

  // Routes
  routes: {
    basePath: '/dashboard',
    
    routes: [
      {
        path: '/',
        component: null as any, // Would be DashboardPage component
        meta: {
          title: 'Dashboard',
          description: 'Main dashboard view',
          icon: 'dashboard',
          order: 1
        }
      },
      {
        path: '/widgets',
        component: null as any, // Would be WidgetSettings component
        meta: {
          title: 'Widget Settings',
          description: 'Manage dashboard widgets',
          order: 2,
          permissions: ['storage:write']
        },
        guards: [
          {
            canActivate: (context) => {
              return context.user?.permissions.includes('dashboard:manage') ?? true;
            },
            redirectTo: '/dashboard'
          }
        ]
      },
      {
        path: '/layout',
        component: null as any, // Would be LayoutEditor component
        meta: {
          title: 'Layout Editor',
          description: 'Edit dashboard layout',
          order: 3
        }
      }
    ],
    
    navigation: [
      {
        id: 'dashboard-main',
        label: 'Dashboard',
        path: '/dashboard',
        icon: 'home',
        order: 1,
        children: [
          {
            id: 'dashboard-widgets',
            label: 'Widgets',
            path: '/dashboard/widgets',
            order: 1
          },
          {
            id: 'dashboard-layout',
            label: 'Layout',
            path: '/dashboard/layout',
            order: 2
          }
        ]
      }
    ]
  },

  // Stores
  stores: {
    readable: {
      activeWidgets,
      isLoading: { subscribe: isLoading.subscribe }
    },
    
    writable: {
      widgets,
      layout,
      preferences
    },
    
    schemas: {
      widgets: {
        type: 'array',
        items: {
          type: 'object',
          properties: {
            id: { type: 'string' },
            type: { type: 'string' },
            title: { type: 'string' }
          }
        }
      },
      layout: {
        type: 'object',
        properties: {
          columns: { type: 'number' },
          widgets: {
            type: 'array',
            items: {
              type: 'object',
              properties: {
                id: { type: 'string' },
                x: { type: 'number' },
                y: { type: 'number' },
                w: { type: 'number' },
                h: { type: 'number' }
              }
            }
          }
        }
      },
      preferences: {
        type: 'object',
        properties: {
          theme: { type: 'string', enum: ['light', 'dark', 'auto'] },
          refreshInterval: { type: 'number' },
          compactMode: { type: 'boolean' }
        }
      }
    }
  },

  // Setup
  async setup(context: PluginContext) {
    const { core, logger, storage } = context;
    
    logger.info('Dashboard plugin initializing...');
    
    // Load saved preferences from storage
    const savedPrefs = await storage.get<DashboardPreferences>('preferences');
    if (savedPrefs) {
      preferences.set(savedPrefs);
      logger.debug('Loaded saved preferences');
    }
    
    // Load saved layout
    const savedLayout = await storage.get<Layout>('layout');
    if (savedLayout) {
      layout.set(savedLayout);
      logger.debug('Loaded saved layout');
    }
    
    // Subscribe to preference changes to persist them
    preferences.subscribe(async (prefs) => {
      await storage.set('preferences', prefs);
    });
    
    // Subscribe to layout changes to persist them
    layout.subscribe(async (l) => {
      if (l) {
        await storage.set('layout', l);
      }
    });
    
    // Subscribe to core events
    core.events.on('user:login', (user) => {
      logger.info('User logged in, loading user-specific dashboard');
    });
    
    core.events.on('user:logout', () => {
      logger.info('User logged out, resetting dashboard');
    });
    
    // Subscribe to messages from other plugins
    core.messages.subscribe('analytics:data-updated', {
      handle: (message) => {
        logger.debug('Analytics data updated:', message.payload);
        // Could trigger widget refresh here
      }
    });
    
    logger.info('Dashboard plugin initialized');
  },

  // Teardown
  async teardown() {
    console.log('[Dashboard] Plugin cleaning up...');
    // Cleanup subscriptions, timers, etc.
  }
};

export default dashboardPlugin;

// Export stores for direct usage in Svelte components
export { widgets, layout, preferences, isLoading, activeWidgets };
