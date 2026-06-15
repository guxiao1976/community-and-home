import type { ModuleConfig } from '../types';
import { Warning, Monitor } from '@element-plus/icons-vue';

export const moderationModule: ModuleConfig = {
  name: 'moderation',

  menu: {
    path: '/moderation',
    title: '内容审核',
    icon: Warning,
    children: [
      {
        path: '/moderation/config-test',
        title: '配置测试',
        icon: Monitor
      }
    ]
  },

  routes: [
    {
      path: '/moderation/config-test',
      name: 'ModerationConfigTest',
      component: () => import('@/views/moderation/ModerationConfigTest.vue'),
      meta: { title: '配置测试', requiresAuth: true }
    }
  ]
};
