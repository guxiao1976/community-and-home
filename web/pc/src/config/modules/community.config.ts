// 社区管理模块 — 核心业务域：行政区划、基层组织、住宅小区、数据统计、数据查询
// 从原 masterdata.config.ts 拆分领域数据部分

import type { ModuleConfig } from '../types'
import type { RouteRecordRaw } from 'vue-router'
import { Location, OfficeBuilding, DataAnalysis, Search } from '@element-plus/icons-vue'

// ==================== 菜单配置 ====================

export const communityMenu = {
  path: '/community',
  title: '社区管理',
  icon: OfficeBuilding,
  children: [
    {
      path: '/masterdata/divisions',
      title: '行政区划',
      icon: Location
    },
    {
      path: '/masterdata/grassroots',
      title: '基层组织',
      icon: OfficeBuilding
    },
    {
      path: '/masterdata/residential-areas',
      title: '住宅小区',
      icon: OfficeBuilding,
      permission: 'community:read'
    },
    {
      path: '/masterdata/statistics/division-counts',
      title: '小区数据统计',
      icon: DataAnalysis
    },
    {
      path: '/masterdata/query',
      title: '数据查询',
      icon: Search
    }
  ]
}

// ==================== 路由配置 ====================

export const communityRoutes: RouteRecordRaw[] = [
  // ---- 行政区划 ----
  {
    path: 'masterdata/divisions',
    name: 'Divisions',
    component: () => import('@/views/division/Index.vue'),
    meta: { title: '行政区划', icon: 'Location', requiresAuth: true }
  },

  // ---- 基层组织 ----
  {
    path: 'masterdata/grassroots',
    name: 'Grassroots',
    component: () => import('@/views/grassroots/Index.vue'),
    meta: { title: '基层组织', icon: 'OfficeBuilding', requiresAuth: true }
  },

  // ---- 住宅小区 ----
  {
    path: 'masterdata/residential-areas',
    name: 'ResidentialAreas',
    component: () => import('@/views/residential-areas/List.vue'),
    meta: { title: '住宅小区', icon: 'OfficeBuilding', requiresAuth: true }
  },
  {
    path: 'masterdata/residential-areas/create',
    name: 'ResidentialAreaCreate',
    component: () => import('@/views/residential-areas/Form.vue'),
    meta: { title: '新建小区', requiresAuth: true, hidden: true }
  },
  {
    path: 'masterdata/residential-areas/:id/edit',
    name: 'ResidentialAreaEdit',
    component: () => import('@/views/residential-areas/Form.vue'),
    meta: { title: '编辑小区', requiresAuth: true, hidden: true }
  },
  {
    path: 'masterdata/residential-areas/:id',
    name: 'ResidentialAreaDetail',
    component: () => import('@/views/residential-areas/Detail.vue'),
    meta: { title: '小区详情', requiresAuth: true, hidden: true }
  },

  // ---- 小区数据统计 ----
  {
    path: 'masterdata/statistics/division-counts',
    name: 'DivisionCounts',
    component: () => import('@/views/statistics/division-counts/Index.vue'),
    meta: { title: '小区数据统计', icon: 'DataAnalysis', requiresAuth: true }
  },

  // ---- 数据查询 ----
  {
    path: 'masterdata/query',
    name: 'MasterdataQuery',
    component: () => import('@/views/masterdata-query/Index.vue'),
    meta: { title: '数据查询', icon: 'Search', requiresAuth: true }
  }
]

// ---- 模块导出 ----
export const communityModule: ModuleConfig = {
  name: 'community',
  menu: communityMenu,
  routes: communityRoutes
}
