// 用户与权限模块 — RBAC 三位一体：用户管理、角色管理、权限资源
// 合并原 identity.config.ts 和 permission.config.ts

import type { ModuleConfig } from '../types'
import type { RouteRecordRaw } from 'vue-router'
import { User, UserFilled, Lock } from '@element-plus/icons-vue'

// ==================== 菜单配置 ====================

export const userPermissionMenu = {
  path: '/user-permission',
  title: '用户与权限',
  icon: UserFilled,
  children: [
    {
      path: '/users/list',
      title: '用户管理',
      icon: User,
      permission: 'user:read'
    },
    {
      path: '/roles',
      title: '角色管理',
      icon: UserFilled,
      permission: 'role:read'
    },
    {
      path: '/system/permissions',
      title: '权限资源',
      icon: Lock,
      permission: 'role:permission'
    }
  ]
}

// ==================== 路由配置 ====================

export const userPermissionRoutes: RouteRecordRaw[] = [
  // ---- 用户管理 ----
  {
    path: 'users/list',
    name: 'UserList',
    component: () => import('@/views/users/List.vue'),
    meta: { title: '用户列表', icon: 'User', requiresAuth: true }
  },
  {
    path: 'users/create',
    name: 'UserCreate',
    component: () => import('@/views/users/Form.vue'),
    meta: { title: '创建用户', requiresAuth: true, hidden: true }
  },
  {
    path: 'users/:id/edit',
    name: 'UserEdit',
    component: () => import('@/views/users/Form.vue'),
    meta: { title: '编辑用户', requiresAuth: true, hidden: true }
  },
  {
    path: 'users/:id',
    name: 'UserDetail',
    component: () => import('@/views/users/Detail.vue'),
    meta: { title: '用户详情', requiresAuth: true, hidden: true }
  },
  {
    path: 'users/:id/roles',
    name: 'UserRoles',
    component: () => import('@/views/users/UserRoles.vue'),
    meta: { title: '用户角色', permissions: ['user:assign-role'], requiresAuth: true, hidden: true }
  },
  // 保留旧路由兼容（管理员管理/普通用户管理 → 隐藏菜单但路由可用）
  {
    path: 'users/admin',
    name: 'AdminUserList',
    component: () => import('@/views/identity/admin-user/AdminUserList.vue'),
    meta: { title: '管理员管理', requiresAuth: true, hidden: true }
  },
  {
    path: 'users/regular',
    name: 'RegularUserList',
    component: () => import('@/views/identity/regular-user/RegularUserList.vue'),
    meta: { title: '普通用户管理', requiresAuth: true, hidden: true }
  },

  // ---- 角色管理 ----
  {
    path: 'roles',
    name: 'RoleList',
    component: () => import('@/views/roles/List.vue'),
    meta: { title: '角色管理', icon: 'UserFilled', requiresAuth: true }
  },
  {
    path: 'roles/:id/permissions',
    name: 'RolePermissions',
    component: () => import('@/views/roles/Permissions.vue'),
    meta: { title: '权限配置', permissions: ['role:permission'], requiresAuth: true, hidden: true }
  },

  {
    path: 'roles/:id/users',
    name: 'RoleUsers',
    component: () => import('@/views/roles/RoleUsers.vue'),
    meta: { title: '角色用户', requiresAuth: true, hidden: true }
  },

  // ---- 权限资源 ----
  {
    path: 'system/permissions',
    name: 'SystemPermissions',
    component: () => import('@/views/roles/Permissions.vue'),
    meta: { title: '权限资源', icon: 'Lock', requiresAuth: true }
  }
]

// ---- 模块导出 ----
export const userPermissionModule: ModuleConfig = {
  name: 'user-permission',
  menu: userPermissionMenu,
  routes: userPermissionRoutes
}
