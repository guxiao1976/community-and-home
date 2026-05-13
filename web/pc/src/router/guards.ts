// Navigation guards for authentication and permission checks

import router from './index';
import { useAuthStore } from '@/stores/auth';
import { usePermissionStore } from '@/stores/permission';
import { logger } from '@/utils/logger';

router.beforeEach(async (to, from, next) => {
  logger.routeChange(from.fullPath, to.fullPath);
  logger.debug('Route meta', {
    toName: to.name,
    toPath: to.path,
    requiresAuth: to.meta.requiresAuth,
    title: to.meta.title
  });

  // Set page title
  document.title = to.meta.title
    ? `${to.meta.title} - 社区家园管理平台`
    : '社区家园管理平台';

  // Check if route requires authentication
  if (to.meta.requiresAuth !== false) {
    const authStore = useAuthStore();
    const authenticated = authStore.isAuthenticated;
    logger.debug('Auth check', { authenticated, toPath: to.path });

    if (!authenticated) {
      logger.warn('Not authenticated, redirecting to login', { from: from.fullPath, to: to.fullPath });
      next({
        path: '/login',
        query: { redirect: to.fullPath }
      });
      return;
    }

    // Check route-level permission
    const permission = (to.meta as any).permission as string | undefined;
    if (permission) {
      const permissionStore = usePermissionStore();

      // If permissions not loaded yet, wait for them
      if (!permissionStore.isLoaded && authStore.user?.id) {
        try {
          await permissionStore.loadUserPermissionsAndMenus(authStore.user.id);
        } catch {
          // If loading fails, allow through to avoid blocking all navigation
          next();
          return;
        }
      }

      if (!permissionStore.hasPermission(permission)) {
        logger.warn('Permission denied', { permission, toPath: to.fullPath });
        next('/403');
        return;
      }
    }
  }

  logger.info('Navigation allowed', { to: to.fullPath });
  next();
});

router.afterEach(() => {
  window.scrollTo(0, 0);
});
