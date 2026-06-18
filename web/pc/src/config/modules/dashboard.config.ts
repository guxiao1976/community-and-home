import type { ModuleConfig } from '../types';
import { HomeFilled } from '@element-plus/icons-vue';

export const dashboardModule: ModuleConfig = {
  name: 'dashboard',

  menu: {
    path: '/dashboard',
    title: '仪表板',
    icon: HomeFilled
  },

  routes: [
    {
      path: 'dashboard',
      name: 'Dashboard',
      component: () => import('@/views/dashboard/Index.vue'),
      meta: {
        title: '首页',
        icon: 'House',
        requiresAuth: true
      }
    }
  ]
};
