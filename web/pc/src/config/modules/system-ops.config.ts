// 系统运维模块 — 运维/配置类功能：敏感词、系统配置、删除恢复、小区获取、运行监控
// 从原 masterdata.config.ts 系统配置部分 + monitoring.config.ts 合并

import type { ModuleConfig } from '../types'
import type { RouteRecordRaw } from 'vue-router'
import { Setting, ChatDotSquare, RefreshRight, Download, Monitor } from '@element-plus/icons-vue'

// ==================== 菜单配置 ====================

export const systemOpsMenu = {
  path: '/system-ops',
  title: '系统运维',
  icon: Setting,
  children: [
    {
      path: '/masterdata/sensitive-words',
      title: '敏感词管理',
      icon: ChatDotSquare
    },
    {
      path: '/masterdata/configs',
      title: '系统配置',
      icon: Setting
    },
    {
      path: '/masterdata/deleted-recovery',
      title: '删除数据恢复',
      icon: RefreshRight
    },
    {
      path: '/masterdata/amap-sync',
      title: '小区数据获取',
      icon: Download
    },
    {
      path: '/monitoring/health',
      title: '运行监控',
      icon: Monitor
    }
  ]
}

// ==================== 路由配置 ====================

export const systemOpsRoutes: RouteRecordRaw[] = [
  // ---- 敏感词管理 ----
  {
    path: 'masterdata/sensitive-words',
    name: 'SensitiveWordList',
    component: () => import('@/views/sensitive-words/List.vue'),
    meta: { title: '敏感词管理', icon: 'Warning', requiresAuth: true }
  },

  // ---- 系统配置 ----
  {
    path: 'masterdata/configs',
    name: 'ConfigList',
    component: () => import('@/views/config/List.vue'),
    meta: { title: '系统配置', icon: 'Setting', requiresAuth: true }
  },

  // ---- 删除数据恢复 ----
  {
    path: 'masterdata/deleted-recovery',
    name: 'DeletedRecovery',
    component: () => import('@/views/deleted-recovery/Index.vue'),
    meta: { title: '删除数据恢复', icon: 'RefreshRight', requiresAuth: true }
  },

  // ---- 小区数据获取 ----
  {
    path: 'masterdata/amap-sync',
    name: 'AMapSync',
    component: () => import('@/views/amap-sync/Index.vue'),
    meta: { title: '小区数据获取', icon: 'Download', requiresAuth: true }
  },

  // ---- 运行监控 (从 monitoring 移入) ----
  {
    path: 'monitoring/health',
    name: 'HealthDashboard',
    component: () => import('@/views/monitoring/HealthDashboard.vue'),
    meta: { title: '运行监控', icon: 'Monitor', requiresAuth: true }
  }
]

// ---- 模块导出 ----
export const systemOpsModule: ModuleConfig = {
  name: 'system-ops',
  menu: systemOpsMenu,
  routes: systemOpsRoutes
}
