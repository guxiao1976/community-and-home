// Navigation guards for authentication and permission checks

import router from './index';
import { useAuthStore } from '@/stores/auth';
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
    // Default to requiring auth unless explicitly set to false
    const authStore = useAuthStore();
    const authenticated = authStore.isAuthenticated;
    logger.debug('Auth check', { authenticated, toPath: to.path });

    if (!authenticated) {
      // Not logged in, redirect to login
      logger.warn('Not authenticated, redirecting to login', { from: from.fullPath, to: to.fullPath });
      next({
        path: '/login',
        query: { redirect: to.fullPath }
      });
      return;
    }

    // TODO: Check permissions when permission store is implemented
    // const permissionStore = usePermissionStore();
    // if (to.meta.permission && !permissionStore.hasPermission(to.meta.permission)) {
    //   next('/403');
    //   return;
    // }
  }

  logger.info('Navigation allowed', { to: to.fullPath });
  next();
});

router.afterEach(() => {
  // Scroll to top after navigation
  window.scrollTo(0, 0);
});
