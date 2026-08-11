// 审核中心模块 — 统一审核入口：实名审核、内容审核、住宅小区审核
// 汇聚原散落在 identity、moderation、masterdata 中的审核功能

import type { ModuleConfig } from '../types'
import type { RouteRecordRaw } from 'vue-router'
import { DocumentChecked, View, Monitor } from '@element-plus/icons-vue'

// ==================== 菜单配置 ====================

export const reviewCenterMenu = {
  path: '/review-center',
  title: '审核中心',
  icon: DocumentChecked,
  children: [
    {
      path: '/users/verifications',
      title: '实名审核',
      icon: DocumentChecked
    },
    {
      path: '/moderation/review',
      title: '内容审核',
      icon: View,
      permission: 'moderation:read'
    },
    {
      path: '/masterdata/residential-areas/review',
      title: '住宅小区审核',
      icon: Monitor
    }
  ]
}

// ==================== 路由配置 ====================

export const reviewCenterRoutes: RouteRecordRaw[] = [
  // ---- 实名审核 (从 identity 移入) ----
  {
    path: 'users/verifications',
    name: 'VerificationList',
    component: () => import('@/views/verification/List.vue'),
    meta: { title: '实名审核', icon: 'DocumentChecked', requiresAuth: true }
  },

  // ---- 内容审核 (从 moderation 移入) ----
  {
    path: 'moderation/review',
    name: 'ManualReview',
    component: () => import('@/views/moderation/ManualReview.vue'),
    meta: { title: '人工审核', requiresAuth: true }
  },
  {
    path: 'moderation/config-test',
    name: 'ModerationConfigTest',
    component: () => import('@/views/moderation/ModerationConfigTest.vue'),
    meta: { title: '配置测试', requiresAuth: true, hidden: true }
  },

  // ---- 住宅小区审核 (从 masterdata 移入) ----
  {
    path: 'masterdata/residential-areas/review',
    name: 'ResidentialAreaReview',
    component: () => import('@/views/residential-areas/Review.vue'),
    meta: { title: '住宅小区审核', icon: 'DocumentChecked', requiresAuth: true }
  }
]

// ---- 模块导出 ----
export const reviewCenterModule: ModuleConfig = {
  name: 'review-center',
  menu: reviewCenterMenu,
  routes: reviewCenterRoutes
}
