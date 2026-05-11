// Vue Router instance

import { createRouter, createWebHistory } from 'vue-router';
import type { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/auth/Login.vue'),
    meta: {
      title: '登录',
      requiresAuth: false
    }
  },
  {
    path: '/register',
    name: 'Register',
    component: () => import('@/views/auth/Register.vue'),
    meta: {
      title: '注册',
      requiresAuth: false
    }
  },
  {
    path: '/',
    component: () => import('@/components/layout/MainLayout.vue'),
    redirect: '/dashboard',
    meta: {
      requiresAuth: true
    },
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/dashboard/Index.vue'),
        meta: {
          title: '首页',
          icon: 'House',
          requiresAuth: true
        }
      },
      {
        path: 'masterdata/divisions',
        name: 'Divisions',
        component: () => import('@/views/division/Index.vue'),
        meta: {
          title: '行政区划',
          icon: 'Location',
          requiresAuth: true
        }
      },
      {
        path: 'masterdata/grassroots',
        name: 'Grassroots',
        component: () => import('@/views/grassroots/Index.vue'),
        meta: {
          title: '基层组织',
          icon: 'OfficeBuilding',
          requiresAuth: true
        }
      },
      {
        path: 'masterdata/residential-areas',
        name: 'ResidentialAreas',
        component: () => import('@/views/residential-areas/List.vue'),
        meta: {
          title: '住宅小区',
          icon: 'OfficeBuilding',
          requiresAuth: true
        }
      },
      {
        path: 'masterdata/residential-areas/create',
        name: 'ResidentialAreaCreate',
        component: () => import('@/views/residential-areas/Form.vue'),
        meta: {
          title: '新建小区',
          requiresAuth: true,
          hidden: true
        }
      },
      {
        path: 'masterdata/residential-areas/:id/edit',
        name: 'ResidentialAreaEdit',
        component: () => import('@/views/residential-areas/Form.vue'),
        meta: {
          title: '编辑小区',
          requiresAuth: true,
          hidden: true
        }
      },
      {
        path: 'masterdata/residential-areas/:id',
        name: 'ResidentialAreaDetail',
        component: () => import('@/views/residential-areas/Detail.vue'),
        meta: {
          title: '小区详情',
          requiresAuth: true,
          hidden: true
        }
      },
      {
        path: 'masterdata/residential-areas/review',
        name: 'ResidentialAreaReview',
        component: () => import('@/views/residential-areas/Review.vue'),
        meta: {
          title: '小区审核',
          icon: 'DocumentChecked',
          requiresAuth: true
        }
      },
      {
        path: 'masterdata/sensitive-words',
        name: 'SensitiveWordList',
        component: () => import('@/views/sensitive-words/List.vue'),
        meta: {
          title: '敏感词管理',
          icon: 'Warning',
          requiresAuth: true
        }
      },
      {
        path: 'masterdata/configs',
        name: 'ConfigList',
        component: () => import('@/views/config/List.vue'),
        meta: {
          title: '系统配置',
          icon: 'Setting',
          requiresAuth: true
        }
      },
      {
        path: 'masterdata/approval-center',
        name: 'ApprovalCenter',
        component: () => import('@/views/approval-center/Index.vue'),
        meta: {
          title: '审核中心',
          icon: 'Stamp',
          requiresAuth: true
        }
      },
      {
        path: 'masterdata/deleted-recovery',
        name: 'DeletedRecovery',
        component: () => import('@/views/deleted-recovery/Index.vue'),
        meta: {
          title: '删除数据恢复',
          icon: 'RefreshRight',
          requiresAuth: true
        }
      },
      {
        path: 'masterdata/statistics/division-counts',
        name: 'DivisionCounts',
        component: () => import('@/views/statistics/division-counts/Index.vue'),
        meta: {
          title: '小区数据统计',
          icon: 'DataAnalysis',
          requiresAuth: true
        }
      },
      {
        path: 'masterdata/amap-sync',
        name: 'AMapSync',
        component: () => import('@/views/amap-sync/Index.vue'),
        meta: {
          title: '小区数据获取',
          icon: 'Download',
          requiresAuth: true
        }
      },
      {
        path: 'masterdata/query',
        name: 'MasterdataQuery',
        component: () => import('@/views/masterdata-query/Index.vue'),
        meta: {
          title: '数据查询',
          icon: 'Search',
          requiresAuth: true
        }
      },
      {
        path: 'users/list',
        name: 'UserList',
        component: () => import('@/views/users/List.vue'),
        meta: {
          title: '用户列表',
          icon: 'User',
          requiresAuth: true
        }
      },
      {
        path: 'users/create',
        name: 'UserCreate',
        component: () => import('@/views/users/Form.vue'),
        meta: {
          title: '创建用户',
          requiresAuth: true,
          hidden: true
        }
      },
      {
        path: 'users/:id/edit',
        name: 'UserEdit',
        component: () => import('@/views/users/Form.vue'),
        meta: {
          title: '编辑用户',
          requiresAuth: true,
          hidden: true
        }
      },
      {
        path: 'users/:id',
        name: 'UserDetail',
        component: () => import('@/views/users/Detail.vue'),
        meta: {
          title: '用户详情',
          requiresAuth: true,
          hidden: true
        }
      },
      {
        path: 'users/verifications',
        name: 'VerificationList',
        component: () => import('@/views/verification/List.vue'),
        meta: {
          title: '实名审核',
          icon: 'DocumentChecked',
          requiresAuth: true
        }
      },
      {
        path: 'roles',
        name: 'RoleList',
        component: () => import('@/views/roles/List.vue'),
        meta: {
          title: '角色管理',
          icon: 'UserFilled',
          requiresAuth: true
        }
      },
      {
        path: 'roles/:id/permissions',
        name: 'RolePermissions',
        component: () => import('@/views/roles/Permissions.vue'),
        meta: {
          title: '权限配置',
          requiresAuth: true,
          hidden: true
        }
      }
    ]
  },
  {
    path: '/403',
    name: 'Forbidden',
    component: () => import('@/views/error/403.vue'),
    meta: {
      title: '访问被拒绝',
      requiresAuth: false
    }
  },
  {
    path: '/404',
    name: 'NotFound',
    component: () => import('@/views/error/404.vue'),
    meta: {
      title: '页面不存在',
      requiresAuth: false
    }
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/404'
  }
];

const router = createRouter({
  history: createWebHistory(),
  routes
});

export default router;
