import type { ModuleConfig } from '../types';
import { Setting, Key, Document, DataAnalysis } from '@element-plus/icons-vue';

export const aimodelModule: ModuleConfig = {
  name: 'aimodel',

  menu: {
    path: '/aimodel',
    title: 'AI模型管理',
    icon: Setting,
    children: [
      {
        path: '/aimodel/models',
        title: '模型配置',
        icon: Setting
      },
      {
        path: '/aimodel/apikeys',
        title: 'API密钥',
        icon: Key
      },
      {
        path: '/aimodel/templates',
        title: '提示模板',
        icon: Document
      },
      {
        path: '/aimodel/statistics',
        title: '使用统计',
        icon: DataAnalysis
      }
    ]
  },

  routes: [
    {
      path: 'aimodel/models',
      name: 'ModelList',
      component: () => import('@/views/aimodel/ModelList.vue'),
      meta: { title: '模型配置', icon: 'Setting', requiresAuth: true }
    },
    {
      path: 'aimodel/models/create',
      name: 'ModelCreate',
      component: () => import('@/views/aimodel/ModelForm.vue'),
      meta: { title: '创建模型', requiresAuth: true, hidden: true }
    },
    {
      path: 'aimodel/models/:id/edit',
      name: 'ModelEdit',
      component: () => import('@/views/aimodel/ModelForm.vue'),
      meta: { title: '编辑模型', requiresAuth: true, hidden: true }
    },
    {
      path: 'aimodel/apikeys',
      name: 'ApiKeyList',
      component: () => import('@/views/aimodel/ApiKeyList.vue'),
      meta: { title: 'API密钥', icon: 'Key', requiresAuth: true }
    },
    {
      path: 'aimodel/templates',
      name: 'TemplateList',
      component: () => import('@/views/aimodel/TemplateList.vue'),
      meta: { title: '提示模板', icon: 'Document', requiresAuth: true }
    },
    {
      path: 'aimodel/statistics',
      name: 'Statistics',
      component: () => import('@/views/aimodel/Statistics.vue'),
      meta: { title: '使用统计', icon: 'DataAnalysis', requiresAuth: true }
    }
  ]
};
