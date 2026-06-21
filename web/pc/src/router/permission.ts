/**
 * 路由权限守卫
 * 根据用户权限动态过滤路由和菜单
 */

import type { Router, RouteRecordRaw } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { ElMessage } from 'element-plus'

/**
 * 检查用户是否有访问该路由的权限
 */
export function hasPermission(permissions: string[], route: RouteRecordRaw): boolean {
  // 如果路由没有定义权限要求，默认允许访问
  if (!route.meta?.permissions || route.meta.permissions.length === 0) {
    return true
  }

  // 检查用户是否拥有路由要求的任一权限
  const requiredPermissions = route.meta.permissions as string[]
  return requiredPermissions.some(permission => permissions.includes(permission))
}

/**
 * 根据权限过滤路由
 */
export function filterRoutes(routes: RouteRecordRaw[], permissions: string[]): RouteRecordRaw[] {
  const filteredRoutes: RouteRecordRaw[] = []

  for (const route of routes) {
    const tmp = { ...route }

    // 检查当前路由是否有权限
    if (hasPermission(permissions, tmp)) {
      // 递归过滤子路由
      if (tmp.children) {
        tmp.children = filterRoutes(tmp.children, permissions)
      }
      filteredRoutes.push(tmp)
    }
  }

  return filteredRoutes
}

/**
 * 生成动态路由
 */
export function generateRoutes(routes: RouteRecordRaw[], permissions: string[]): RouteRecordRaw[] {
  return filterRoutes(routes, permissions)
}

/**
 * 安装路由守卫
 */
export function setupPermissionGuard(router: Router) {
  router.beforeEach(async (to, from, next) => {
    const userStore = useUserStore()

    // 白名单路由（登录页、404等）
    const whiteList = ['/login', '/404', '/403']
    if (whiteList.includes(to.path)) {
      next()
      return
    }

    // 检查用户是否已登录
    if (!userStore.token) {
      ElMessage.warning('请先登录')
      next({ path: '/login', query: { redirect: to.fullPath } })
      return
    }

    // 检查用户信息是否已加载
    if (!userStore.userInfo) {
      try {
        await userStore.getUserInfo()
      } catch (error) {
        ElMessage.error('获取用户信息失败')
        next({ path: '/login' })
        return
      }
    }

    // 检查用户权限是否已加载
    if (!userStore.permissions || userStore.permissions.length === 0) {
      try {
        await userStore.getUserPermissions()
      } catch (error) {
        ElMessage.error('获取用户权限失败')
        next({ path: '/login' })
        return
      }
    }

    // 检查路由权限
    if (to.meta?.permissions && Array.isArray(to.meta.permissions)) {
      const hasRoutePermission = (to.meta.permissions as string[]).some(permission =>
        userStore.permissions.includes(permission)
      )

      if (!hasRoutePermission) {
        ElMessage.warning('您没有访问该页面的权限')
        next({ path: '/403' })
        return
      }
    }

    next()
  })
}

/**
 * 动态添加路由
 */
export function addDynamicRoutes(router: Router, routes: RouteRecordRaw[], permissions: string[]) {
  // 过滤出有权限的路由
  const accessRoutes = generateRoutes(routes, permissions)

  // 添加到路由器
  accessRoutes.forEach(route => {
    router.addRoute(route)
  })

  return accessRoutes
}
