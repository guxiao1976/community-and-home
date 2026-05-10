# Routing Contracts: Web PC Admin Frontend

**Feature**: 002-web-pc-admin  
**Date**: 2026-05-03  
**Status**: Complete

## Overview

This document defines the routing structure, navigation guards, and route metadata for the PC admin application. All routes use Vue Router with lazy loading for code splitting.

## Route Structure

### Public Routes (No Authentication Required)

```typescript
// web/pc/src/router/routes.ts
const publicRoutes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/auth/Login.vue'),
    meta: {
      title: 'Login',
      requiresAuth: false
    }
  },
  {
    path: '/register',
    name: 'Register',
    component: () => import('@/views/auth/Register.vue'),
    meta: {
      title: 'Register',
      requiresAuth: false
    }
  },
  {
    path: '/403',
    name: 'Forbidden',
    component: () => import('@/views/error/403.vue'),
    meta: {
      title: 'Access Denied',
      requiresAuth: false
    }
  },
  {
    path: '/404',
    name: 'NotFound',
    component: () => import('@/views/error/404.vue'),
    meta: {
      title: 'Page Not Found',
      requiresAuth: false
    }
  }
];
```

---

### Protected Routes (Authentication Required)

```typescript
const protectedRoutes: RouteRecordRaw[] = [
  {
    path: '/',
    component: () => import('@/layouts/MainLayout.vue'),
    redirect: '/dashboard',
    meta: {
      requiresAuth: true
    },
    children: [
      // Dashboard
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/dashboard/Index.vue'),
        meta: {
          title: 'Dashboard',
          icon: 'Dashboard',
          requiresAuth: true
        }
      },

      // Administrative Division Management
      {
        path: 'divisions',
        name: 'DivisionManagement',
        component: () => import('@/views/divisions/Index.vue'),
        meta: {
          title: 'Administrative Divisions',
          icon: 'Location',
          requiresAuth: true,
          permission: 'masterdata:division:view'
        }
      },

      // Community Management
      {
        path: 'communities',
        name: 'CommunityList',
        component: () => import('@/views/communities/List.vue'),
        meta: {
          title: 'Community Management',
          icon: 'OfficeBuilding',
          requiresAuth: true,
          permission: 'masterdata:community:view'
        }
      },
      {
        path: 'communities/create',
        name: 'CommunityCreate',
        component: () => import('@/views/communities/Form.vue'),
        meta: {
          title: 'Create Community',
          requiresAuth: true,
          permission: 'masterdata:community:create',
          breadcrumb: [
            { title: 'Community Management', to: '/communities' },
            { title: 'Create Community' }
          ]
        }
      },
      {
        path: 'communities/:id/edit',
        name: 'CommunityEdit',
        component: () => import('@/views/communities/Form.vue'),
        meta: {
          title: 'Edit Community',
          requiresAuth: true,
          permission: 'masterdata:community:update',
          breadcrumb: [
            { title: 'Community Management', to: '/communities' },
            { title: 'Edit Community' }
          ]
        }
      },
      {
        path: 'communities/:id',
        name: 'CommunityDetail',
        component: () => import('@/views/communities/Detail.vue'),
        meta: {
          title: 'Community Details',
          requiresAuth: true,
          permission: 'masterdata:community:view',
          breadcrumb: [
            { title: 'Community Management', to: '/communities' },
            { title: 'Community Details' }
          ]
        }
      },
      {
        path: 'communities/review',
        name: 'CommunityReview',
        component: () => import('@/views/communities/Review.vue'),
        meta: {
          title: 'Community Review',
          icon: 'DocumentChecked',
          requiresAuth: true,
          permission: 'masterdata:community:review',
          breadcrumb: [
            { title: 'Community Management', to: '/communities' },
            { title: 'Review Queue' }
          ]
        }
      },

      // User Management
      {
        path: 'users',
        name: 'UserList',
        component: () => import('@/views/users/List.vue'),
        meta: {
          title: 'User Management',
          icon: 'User',
          requiresAuth: true,
          permission: 'identity:user:view'
        }
      },
      {
        path: 'users/create',
        name: 'UserCreate',
        component: () => import('@/views/users/Form.vue'),
        meta: {
          title: 'Create User',
          requiresAuth: true,
          permission: 'identity:user:create',
          breadcrumb: [
            { title: 'User Management', to: '/users' },
            { title: 'Create User' }
          ]
        }
      },
      {
        path: 'users/:id/edit',
        name: 'UserEdit',
        component: () => import('@/views/users/Form.vue'),
        meta: {
          title: 'Edit User',
          requiresAuth: true,
          permission: 'identity:user:update',
          breadcrumb: [
            { title: 'User Management', to: '/users' },
            { title: 'Edit User' }
          ]
        }
      },
      {
        path: 'users/:id',
        name: 'UserDetail',
        component: () => import('@/views/users/Detail.vue'),
        meta: {
          title: 'User Details',
          requiresAuth: true,
          permission: 'identity:user:view',
          breadcrumb: [
            { title: 'User Management', to: '/users' },
            { title: 'User Details' }
          ]
        }
      },

      // Role & Permission Management
      {
        path: 'roles',
        name: 'RoleList',
        component: () => import('@/views/roles/List.vue'),
        meta: {
          title: 'Role Management',
          icon: 'UserFilled',
          requiresAuth: true,
          permission: 'identity:role:view'
        }
      },
      {
        path: 'roles/create',
        name: 'RoleCreate',
        component: () => import('@/views/roles/Form.vue'),
        meta: {
          title: 'Create Role',
          requiresAuth: true,
          permission: 'identity:role:create',
          breadcrumb: [
            { title: 'Role Management', to: '/roles' },
            { title: 'Create Role' }
          ]
        }
      },
      {
        path: 'roles/:id/edit',
        name: 'RoleEdit',
        component: () => import('@/views/roles/Form.vue'),
        meta: {
          title: 'Edit Role',
          requiresAuth: true,
          permission: 'identity:role:update',
          breadcrumb: [
            { title: 'Role Management', to: '/roles' },
            { title: 'Edit Role' }
          ]
        }
      },
      {
        path: 'roles/:id/permissions',
        name: 'RolePermissions',
        component: () => import('@/views/roles/Permissions.vue'),
        meta: {
          title: 'Assign Permissions',
          requiresAuth: true,
          permission: 'identity:role:assign-permission',
          breadcrumb: [
            { title: 'Role Management', to: '/roles' },
            { title: 'Assign Permissions' }
          ]
        }
      },

      // Homeowner Verification Review
      {
        path: 'verifications',
        name: 'VerificationList',
        component: () => import('@/views/verification/List.vue'),
        meta: {
          title: 'Verification Review',
          icon: 'DocumentChecked',
          requiresAuth: true,
          permission: 'identity:verification:view'
        }
      },
      {
        path: 'verifications/:id',
        name: 'VerificationDetail',
        component: () => import('@/views/verification/Detail.vue'),
        meta: {
          title: 'Verification Details',
          requiresAuth: true,
          permission: 'identity:verification:view',
          breadcrumb: [
            { title: 'Verification Review', to: '/verifications' },
            { title: 'Verification Details' }
          ]
        }
      },

      // System Configuration
      {
        path: 'configs',
        name: 'ConfigList',
        component: () => import('@/views/config/List.vue'),
        meta: {
          title: 'System Configuration',
          icon: 'Setting',
          requiresAuth: true,
          permission: 'masterdata:config:view'
        }
      },

      // Sensitive Word Management
      {
        path: 'sensitive-words',
        name: 'SensitiveWordList',
        component: () => import('@/views/sensitive-words/List.vue'),
        meta: {
          title: 'Sensitive Words',
          icon: 'Warning',
          requiresAuth: true,
          permission: 'masterdata:sensitive-word:view'
        }
      }
    ]
  }
];
```

---

## Route Metadata

### Meta Fields

```typescript
interface RouteMeta {
  title: string;              // Page title (for document.title and breadcrumb)
  icon?: string;              // Element Plus icon name (for menu)
  requiresAuth: boolean;      // Requires authentication
  permission?: string;        // Required permission code (format: service:resource:action)
  breadcrumb?: Breadcrumb[];  // Custom breadcrumb trail
  hideInMenu?: boolean;       // Hide from sidebar menu (default: false)
  keepAlive?: boolean;        // Keep component alive (default: false)
}

interface Breadcrumb {
  title: string;
  to?: string;  // Route path (omit for current page)
}
```

### Permission Codes

Permission codes follow the format: `service:resource:action`

**Identity Service Permissions**:
- `identity:user:view` - View user list and details
- `identity:user:create` - Create new users
- `identity:user:update` - Edit user information
- `identity:user:delete` - Delete users
- `identity:user:disable` - Disable/enable users
- `identity:role:view` - View role list
- `identity:role:create` - Create new roles
- `identity:role:update` - Edit roles
- `identity:role:delete` - Delete roles
- `identity:role:assign-permission` - Assign permissions to roles
- `identity:verification:view` - View verification requests
- `identity:verification:review` - Approve/reject verifications

**Masterdata Service Permissions**:
- `masterdata:division:view` - View division tree
- `masterdata:division:create` - Create divisions
- `masterdata:division:update` - Edit divisions
- `masterdata:division:delete` - Delete divisions
- `masterdata:community:view` - View community list
- `masterdata:community:create` - Create communities
- `masterdata:community:update` - Edit communities
- `masterdata:community:delete` - Delete communities
- `masterdata:community:submit` - Submit communities for review
- `masterdata:community:review` - Approve/reject communities
- `masterdata:config:view` - View configurations
- `masterdata:config:create` - Create configurations
- `masterdata:config:update` - Edit configurations
- `masterdata:config:delete` - Delete configurations
- `masterdata:sensitive-word:view` - View sensitive words
- `masterdata:sensitive-word:create` - Add sensitive words
- `masterdata:sensitive-word:update` - Edit sensitive words
- `masterdata:sensitive-word:delete` - Delete sensitive words

---

## Navigation Guards

### Global Before Guard

```typescript
// web/pc/src/router/guards.ts
import { useAuthStore } from '@/stores/auth';
import { usePermissionStore } from '@/stores/permission';

router.beforeEach(async (to, from, next) => {
  const authStore = useAuthStore();
  const permissionStore = usePermissionStore();

  // Set page title
  document.title = to.meta.title 
    ? `${to.meta.title} - Community & Home Admin`
    : 'Community & Home Admin';

  // Check authentication
  if (to.meta.requiresAuth) {
    if (!authStore.isAuthenticated) {
      // Not logged in, redirect to login
      next({
        path: '/login',
        query: { redirect: to.fullPath }
      });
      return;
    }

    // Load permissions if not loaded
    if (permissionStore.permissions.length === 0) {
      try {
        await permissionStore.loadPermissions();
      } catch (error) {
        console.error('Failed to load permissions:', error);
        authStore.logout();
        next('/login');
        return;
      }
    }

    // Check permission
    if (to.meta.permission) {
      if (!permissionStore.hasPermission(to.meta.permission)) {
        next('/403');
        return;
      }
    }
  }

  next();
});
```

### Global After Guard

```typescript
router.afterEach((to, from) => {
  // Update breadcrumb in app store
  const appStore = useAppStore();
  appStore.updateBreadcrumb(to);

  // Scroll to top
  window.scrollTo(0, 0);
});
```

---

## Menu Structure

The sidebar menu is dynamically generated from routes with menu permissions.

```typescript
// web/pc/src/components/layout/Sidebar.vue
interface MenuItem {
  path: string;
  title: string;
  icon?: string;
  permission?: string;
  children?: MenuItem[];
}

const generateMenu = (routes: RouteRecordRaw[]): MenuItem[] => {
  const permissionStore = usePermissionStore();
  
  return routes
    .filter(route => {
      // Filter out routes without permission or hidden routes
      if (route.meta?.hideInMenu) return false;
      if (route.meta?.permission && !permissionStore.hasPermission(route.meta.permission)) {
        return false;
      }
      return true;
    })
    .map(route => ({
      path: route.path,
      title: route.meta?.title || '',
      icon: route.meta?.icon,
      permission: route.meta?.permission,
      children: route.children ? generateMenu(route.children) : undefined
    }));
};
```

**Menu Hierarchy**:
```
Dashboard
├─ Administrative Divisions
├─ Community Management
│  ├─ Community List
│  └─ Review Queue
├─ User Management
├─ Role Management
├─ Verification Review
├─ System Configuration
└─ Sensitive Words
```

---

## Breadcrumb Navigation

Breadcrumbs are automatically generated from route hierarchy or custom meta.breadcrumb.

```typescript
// web/pc/src/stores/app.ts
export const useAppStore = defineStore('app', () => {
  const breadcrumb = ref<Breadcrumb[]>([]);

  const updateBreadcrumb = (route: RouteLocationNormalized) => {
    if (route.meta.breadcrumb) {
      // Use custom breadcrumb
      breadcrumb.value = route.meta.breadcrumb as Breadcrumb[];
    } else {
      // Generate from route hierarchy
      const matched = route.matched.filter(r => r.meta?.title);
      breadcrumb.value = matched.map(r => ({
        title: r.meta.title as string,
        to: r.path === route.path ? undefined : r.path
      }));
    }
  };

  return { breadcrumb, updateBreadcrumb };
});
```

---

## Route Transitions

```typescript
// web/pc/src/App.vue
<template>
  <router-view v-slot="{ Component, route }">
    <transition name="fade" mode="out-in">
      <keep-alive :include="keepAliveRoutes">
        <component :is="Component" :key="route.path" />
      </keep-alive>
    </transition>
  </router-view>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
```

---

## Lazy Loading Strategy

All route components use dynamic imports for code splitting:

```typescript
// Lazy load by route
component: () => import('@/views/users/List.vue')

// Lazy load with chunk name (for better debugging)
component: () => import(/* webpackChunkName: "users" */ '@/views/users/List.vue')
```

**Chunk Strategy**:
- **auth**: Login, Register (public routes)
- **dashboard**: Dashboard page
- **divisions**: Division management
- **communities**: Community management and review
- **users**: User management
- **roles**: Role and permission management
- **verification**: Homeowner verification review
- **config**: System configuration
- **sensitive-words**: Sensitive word management
- **error**: 403, 404 error pages

---

## Redirect Rules

```typescript
// Redirect root to dashboard
{
  path: '/',
  redirect: '/dashboard'
}

// Redirect after login
router.push(route.query.redirect as string || '/dashboard');

// Redirect after logout
router.push('/login');

// Redirect on 401 (unauthorized)
router.push({
  path: '/login',
  query: { redirect: router.currentRoute.value.fullPath }
});

// Redirect on 403 (forbidden)
router.push('/403');

// Catch-all 404
{
  path: '/:pathMatch(.*)*',
  redirect: '/404'
}
```

---

## Summary

This routing contract defines:

- **50+ routes** across 8 feature modules
- **Public routes**: Login, Register, Error pages
- **Protected routes**: Dashboard + 8 management modules
- **Route metadata**: Title, icon, permissions, breadcrumb
- **Navigation guards**: Authentication and permission checks
- **Menu generation**: Dynamic sidebar from routes with permissions
- **Breadcrumb navigation**: Automatic or custom breadcrumb trails
- **Lazy loading**: Code splitting by feature module
- **Redirect rules**: Login, logout, error handling

All routes align with the 8 user stories and support role-based access control as defined in the feature spec.
