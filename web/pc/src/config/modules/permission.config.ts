// 权限管理模块菜单配置
// 添加角色、权限、资源管理菜单

import type { RouteRecordRaw } from 'vue-router'
import type { MenuItemConfig } from '../types'

// 菜单配置
export const permissionMenu: MenuItemConfig = {
  path: '/system',
  title: '系统管理',
  icon: 'Setting',
  permission: 'system:view',
  children: [
    {
      path: '/system/users',
      title: '用户管理',
      icon: 'User',
      permission: 'system:user:view'
    },
    {
      path: '/system/roles',
      title: '角色管理',
      icon: 'UserFilled',
      permission: 'system:role:view'
    },
    {
      path: '/system/permissions',
      title: '权限管理',
      icon: 'Lock',
      permission: 'system:permission:view'
    }
  ]
}

// 路由配置
export const permissionRoutes: RouteRecordRaw[] = [
  {
    path: 'system',
    meta: {
      title: '系统管理'
    },
    children: [
      {
        path: 'users',
        name: 'SystemUsers',
        component: () => import('@/views/system/UserManagement.vue'),
        meta: {
          title: '用户管理',
          permissions: ['system:user:view']
        }
      },
      {
        path: 'roles',
        name: 'SystemRoles',
        component: () => import('@/views/system/RoleManagement.vue'),
        meta: {
          title: '角色管理',
          permissions: ['system:role:view']
        }
      },
      {
        path: 'permissions',
        name: 'SystemPermissions',
        component: () => import('@/views/roles/Permissions.vue'),
        meta: {
          title: '权限管理',
          permissions: ['system:permission:view']
        }
      }
    ]
  },
  {
    path: 'roles',
    meta: {
      title: '角色管理'
    },
    children: [
      {
        path: '',
        name: 'RolesList',
        component: () => import('@/views/roles/List.vue'),
        meta: {
          title: '角色列表',
          permissions: ['system:role:view']
        }
      },
      {
        path: ':id/permissions',
        name: 'RolePermissions',
        component: () => import('@/views/roles/Permissions.vue'),
        meta: {
          title: '角色权限配置',
          permissions: ['system:role:permission']
        }
      }
    ]
  },
  {
    path: 'users/:id/roles',
    name: 'UserRoles',
    component: () => import('@/views/users/UserRoles.vue'),
    meta: {
      title: '用户角色',
      permissions: ['system:user:role']
    }
  }
]

// 模块导出
export const permissionModule = {
  menu: permissionMenu,
  routes: permissionRoutes
}
