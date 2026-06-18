import type { ModuleConfig } from '../types';
import {
  Setting,
  Location,
  OfficeBuilding,
  ChatDotSquare,
  Stamp,
  RefreshRight,
  DataAnalysis,
  Search,
  Download
} from '@element-plus/icons-vue';

export const masterdataModule: ModuleConfig = {
  name: 'masterdata',

  menu: {
    path: '/masterdata',
    title: '主数据管理',
    icon: Setting,
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
        icon: OfficeBuilding
      },
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
        path: '/masterdata/approval-center',
        title: '审核中心',
        icon: Stamp
      },
      {
        path: '/masterdata/deleted-recovery',
        title: '删除数据恢复',
        icon: RefreshRight
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
      },
      {
        path: '/masterdata/amap-sync',
        title: '小区数据获取',
        icon: Download
      }
    ]
  },

  routes: [
    {
      path: 'masterdata/divisions',
      name: 'Divisions',
      component: () => import('@/views/division/Index.vue'),
      meta: { title: '行政区划', icon: 'Location', requiresAuth: true }
    },
    {
      path: 'masterdata/grassroots',
      name: 'Grassroots',
      component: () => import('@/views/grassroots/Index.vue'),
      meta: { title: '基层组织', icon: 'OfficeBuilding', requiresAuth: true }
    },
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
    {
      path: 'masterdata/residential-areas/review',
      name: 'ResidentialAreaReview',
      component: () => import('@/views/residential-areas/Review.vue'),
      meta: { title: '小区审核', icon: 'DocumentChecked', requiresAuth: true }
    },
    {
      path: 'masterdata/sensitive-words',
      name: 'SensitiveWordList',
      component: () => import('@/views/sensitive-words/List.vue'),
      meta: { title: '敏感词管理', icon: 'Warning', requiresAuth: true }
    },
    {
      path: 'masterdata/configs',
      name: 'ConfigList',
      component: () => import('@/views/config/List.vue'),
      meta: { title: '系统配置', icon: 'Setting', requiresAuth: true }
    },
    {
      path: 'masterdata/approval-center',
      name: 'ApprovalCenter',
      component: () => import('@/views/approval-center/Index.vue'),
      meta: { title: '审核中心', icon: 'Stamp', requiresAuth: true }
    },
    {
      path: 'masterdata/deleted-recovery',
      name: 'DeletedRecovery',
      component: () => import('@/views/deleted-recovery/Index.vue'),
      meta: { title: '删除数据恢复', icon: 'RefreshRight', requiresAuth: true }
    },
    {
      path: 'masterdata/statistics/division-counts',
      name: 'DivisionCounts',
      component: () => import('@/views/statistics/division-counts/Index.vue'),
      meta: { title: '小区数据统计', icon: 'DataAnalysis', requiresAuth: true }
    },
    {
      path: 'masterdata/amap-sync',
      name: 'AMapSync',
      component: () => import('@/views/amap-sync/Index.vue'),
      meta: { title: '小区数据获取', icon: 'Download', requiresAuth: true }
    },
    {
      path: 'masterdata/query',
      name: 'MasterdataQuery',
      component: () => import('@/views/masterdata-query/Index.vue'),
      meta: { title: '数据查询', icon: 'Search', requiresAuth: true }
    }
  ]
};
