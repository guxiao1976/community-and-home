import type { ModuleConfig } from '../types';
import { Monitor } from '@element-plus/icons-vue';

export const monitoringModule: ModuleConfig = {
  name: 'monitoring',

  menu: {
    path: '/monitoring',
    title: '运行监控',
    icon: Monitor,
    children: [
      {
        path: '/monitoring/health',
        title: '服务健康',
        icon: Monitor
      }
    ]
  },

  routes: [
    {
      path: 'monitoring/health',
      name: 'HealthDashboard',
      component: () => import('@/views/monitoring/HealthDashboard.vue'),
      meta: { title: '服务健康', icon: 'Monitor', requiresAuth: true }
    }
  ]
};
