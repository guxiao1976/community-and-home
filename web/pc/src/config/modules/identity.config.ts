import type { ModuleConfig } from '../types';
import { User, UserFilled, Document } from '@element-plus/icons-vue';

export const identityModule: ModuleConfig = {
  name: 'identity',

  menu: {
    path: '/users',
    title: '用户管理',
    icon: User,
    children: [
      {
        path: '/users/list',
        title: '用户列表',
        icon: User,
        permission: 'user:view'
      },
      {
        path: '/users/verifications',
        title: '实名审核',
        icon: Document,
        permission: 'verification:view'
      },
      {
        path: '/roles',
        title: '角色管理',
        icon: UserFilled,
        permission: 'role:view'
      }
    ]
  },

  routes: [
    {
      path: 'users/list',
      name: 'UserList',
      component: () => import('@/views/users/List.vue'),
      meta: { title: '用户列表', icon: 'User', requiresAuth: true, permission: 'user:view' }
    },
    {
      path: 'users/create',
      name: 'UserCreate',
      component: () => import('@/views/users/Form.vue'),
      meta: { title: '创建用户', requiresAuth: true, hidden: true, permission: 'user:create' }
    },
    {
      path: 'users/:id/edit',
      name: 'UserEdit',
      component: () => import('@/views/users/Form.vue'),
      meta: { title: '编辑用户', requiresAuth: true, hidden: true, permission: 'user:update' }
    },
    {
      path: 'users/:id',
      name: 'UserDetail',
      component: () => import('@/views/users/Detail.vue'),
      meta: { title: '用户详情', requiresAuth: true, hidden: true, permission: 'user:view' }
    },
    {
      path: 'users/verifications',
      name: 'VerificationList',
      component: () => import('@/views/verification/List.vue'),
      meta: { title: '实名审核', icon: 'DocumentChecked', requiresAuth: true, permission: 'verification:view' }
    },
    {
      path: 'roles',
      name: 'RoleList',
      component: () => import('@/views/roles/List.vue'),
      meta: { title: '角色管理', icon: 'UserFilled', requiresAuth: true, permission: 'role:view' }
    },
    {
      path: 'roles/:id/permissions',
      name: 'RolePermissions',
      component: () => import('@/views/roles/Permissions.vue'),
      meta: { title: '权限配置', requiresAuth: true, hidden: true, permission: 'role:permission' }
    }
  ]
};
