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
        path: '/moderation/test',
        title: '内容审核测试',
        icon: Monitor
      }
    ]
  },

  routes: [
    {
      path: '/moderation/test',
      name: 'ModerationTest',
      component: () => import('@/views/moderation/ModerationTest.vue'),
      meta: { title: '内容审核测试', requiresAuth: true }
    }
  ]
};
