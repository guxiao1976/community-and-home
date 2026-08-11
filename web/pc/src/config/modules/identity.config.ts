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
        path: '/users/admin',
        title: '管理员管理',
        icon: UserFilled
        // permission: 'identity:admin-user:list'  // 临时注释，开发环境使用
      },
      {
        path: '/users/regular',
        title: '普通用户管理',
        icon: User
        // permission: 'identity:regular-user:list'  // 临时注释，开发环境使用
      },
      {
        path: '/users/verifications',
        title: '实名审核',
        icon: Document
        // permission: 'verification:view'  // 临时注释，开发环境使用
      },
      {
        path: '/roles',
        title: '角色管理',
        icon: UserFilled
        // permission: 'role:view'  // 临时注释，开发环境使用
      }
    ]
  },

  routes: [
    {
      path: 'users/admin',
      name: 'AdminUserList',
      component: () => import('@/views/identity/admin-user/AdminUserList.vue'),
      meta: { title: '管理员管理', icon: 'UserFilled', requiresAuth: true}
    },
    {
      path: 'users/regular',
      name: 'RegularUserList',
      component: () => import('@/views/identity/regular-user/RegularUserList.vue'),
      meta: { title: '普通用户管理', icon: 'User', requiresAuth: true}
    },
    {
      path: 'users/list',
      name: 'UserList',
      component: () => import('@/views/users/List.vue'),
      meta: { title: '用户列表', icon: 'User', requiresAuth: true}
    },
    {
      path: 'users/create',
      name: 'UserCreate',
      component: () => import('@/views/users/Form.vue'),
      meta: { title: '创建用户', requiresAuth: true, hidden: true}
    },
    {
      path: 'users/:id/edit',
      name: 'UserEdit',
      component: () => import('@/views/users/Form.vue'),
      meta: { title: '编辑用户', requiresAuth: true, hidden: true}
    },
    {
      path: 'users/:id',
      name: 'UserDetail',
      component: () => import('@/views/users/Detail.vue'),
      meta: { title: '用户详情', requiresAuth: true, hidden: true}
    },
    {
      path: 'users/verifications',
      name: 'VerificationList',
      component: () => import('@/views/verification/List.vue'),
      meta: { title: '实名审核', icon: 'DocumentChecked', requiresAuth: true}
    },
    {
      path: 'roles',
      name: 'RoleList',
      component: () => import('@/views/roles/List.vue'),
      meta: { title: '角色管理', icon: 'UserFilled', requiresAuth: true}
    },
    {
      path: 'roles/:id/permissions',
      name: 'RolePermissions',
      component: () => import('@/views/roles/Permissions.vue'),
      meta: { title: '权限配置', requiresAuth: true, hidden: true}
    }
  ]
};
