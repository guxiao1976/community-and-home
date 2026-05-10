# State Contracts: Web PC Admin Frontend

**Feature**: 002-web-pc-admin  
**Date**: 2026-05-03  
**Status**: Complete

## Overview

This document defines the Pinia store contracts for state management in the PC admin application. All stores use composition-style API with TypeScript for type safety.

## Store Architecture

### Store Organization

```
web/pc/src/stores/
├── auth.ts          # Authentication state (tokens, user info)
├── permission.ts    # Permission state (roles, menus, buttons)
├── app.ts           # App state (loading, breadcrumb, sidebar)
├── division.ts      # Division tree cache
└── user.ts          # User management state
```

---

## 1. Auth Store

**Purpose**: Manages authentication state, tokens, and user session.

**File**: `web/pc/src/stores/auth.ts`

### State

```typescript
export const useAuthStore = defineStore('auth', () => {
  // State
  const user = ref<User | null>(null);
  const accessToken = ref<string | null>(null);
  const refreshToken = ref<string | null>(null);
  const tokenExpiry = ref<number>(0);

  // Computed
  const isAuthenticated = computed(() => !!accessToken.value);
  const isTokenExpiring = computed(() => {
    if (!tokenExpiry.value) return false;
    const now = Date.now();
    const fiveMinutes = 5 * 60 * 1000;
    return tokenExpiry.value - now < fiveMinutes;
  });

  // Actions
  const login = async (phone: string, password: string): Promise<void> => {
    const response = await identityApi.login({ phone, password });
    setTokens(response.accessToken, response.refreshToken, response.expiresIn);
    user.value = response.user;
  };

  const loginWithSms = async (phone: string, smsCode: string): Promise<void> => {
    const response = await identityApi.loginWithSms({ phone, smsCode });
    setTokens(response.accessToken, response.refreshToken, response.expiresIn);
    user.value = response.user;
  };

  const register = async (data: RegisterRequest): Promise<void> => {
    const response = await identityApi.register(data);
    setTokens(response.accessToken, response.refreshToken, response.expiresIn);
    user.value = response.user;
  };

  const logout = async (): Promise<void> => {
    try {
      await identityApi.logout();
    } finally {
      clearTokens();
      user.value = null;
    }
  };

  const refreshAccessToken = async (): Promise<void> => {
    if (!refreshToken.value) {
      throw new Error('No refresh token available');
    }
    const response = await identityApi.refreshToken({ refreshToken: refreshToken.value });
    setTokens(response.accessToken, response.refreshToken, response.expiresIn);
  };

  const setTokens = (access: string, refresh: string, expiresIn: number): void => {
    accessToken.value = access;
    refreshToken.value = refresh;
    tokenExpiry.value = Date.now() + expiresIn * 1000;
    
    sessionStorage.setItem('accessToken', access);
    localStorage.setItem('refreshToken', refresh);
    localStorage.setItem('tokenExpiry', tokenExpiry.value.toString());
  };

  const clearTokens = (): void => {
    accessToken.value = null;
    refreshToken.value = null;
    tokenExpiry.value = 0;
    
    sessionStorage.removeItem('accessToken');
    localStorage.removeItem('refreshToken');
    localStorage.removeItem('tokenExpiry');
  };

  const restoreSession = (): void => {
    const access = sessionStorage.getItem('accessToken');
    const refresh = localStorage.getItem('refreshToken');
    const expiry = localStorage.getItem('tokenExpiry');
    
    if (access && refresh && expiry) {
      accessToken.value = access;
      refreshToken.value = refresh;
      tokenExpiry.value = parseInt(expiry, 10);
    }
  };

  return {
    // State
    user,
    accessToken,
    refreshToken,
    tokenExpiry,
    // Computed
    isAuthenticated,
    isTokenExpiring,
    // Actions
    login,
    loginWithSms,
    register,
    logout,
    refreshAccessToken,
    setTokens,
    clearTokens,
    restoreSession
  };
}, {
  persist: {
    paths: ['user']  // Only persist user info, not tokens
  }
});
```

### Usage

```typescript
// In components
const authStore = useAuthStore();

// Login
await authStore.login('13800000000', 'Admin@123456');

// Check authentication
if (authStore.isAuthenticated) {
  // User is logged in
}

// Get current user
const currentUser = authStore.user;

// Logout
await authStore.logout();
```

---

## 2. Permission Store

**Purpose**: Manages user permissions, roles, and menu access.

**File**: `web/pc/src/stores/permission.ts`

### State

```typescript
export const usePermissionStore = defineStore('permission', () => {
  // State
  const permissions = ref<string[]>([]);
  const menus = ref<Permission[]>([]);
  const roles = ref<Role[]>([]);

  // Computed
  const menuTree = computed(() => buildMenuTree(menus.value));

  // Actions
  const loadPermissions = async (): Promise<void> => {
    const authStore = useAuthStore();
    if (!authStore.user) {
      throw new Error('User not authenticated');
    }
    
    const response = await identityApi.getUserPermissions(authStore.user.id);
    permissions.value = response.permissions;
    menus.value = response.menus;
  };

  const loadRoles = async (): Promise<void> => {
    const response = await identityApi.getRoles();
    roles.value = response.list;
  };

  const hasPermission = (permission: string | string[]): boolean => {
    if (Array.isArray(permission)) {
      return permission.some(p => permissions.value.includes(p));
    }
    return permissions.value.includes(permission);
  };

  const hasAllPermissions = (permissionList: string[]): boolean => {
    return permissionList.every(p => permissions.value.includes(p));
  };

  const hasRole = (roleCode: string): boolean => {
    return roles.value.some(r => r.code === roleCode);
  };

  const canAccessRoute = (route: RouteLocationNormalized): boolean => {
    if (!route.meta.permission) return true;
    return hasPermission(route.meta.permission as string);
  };

  const filterMenusByPermission = (menuList: Permission[]): Permission[] => {
    return menuList.filter(menu => {
      if (menu.code && !hasPermission(menu.code)) {
        return false;
      }
      if (menu.children) {
        menu.children = filterMenusByPermission(menu.children);
      }
      return true;
    });
  };

  const clearPermissions = (): void => {
    permissions.value = [];
    menus.value = [];
    roles.value = [];
  };

  return {
    // State
    permissions,
    menus,
    roles,
    // Computed
    menuTree,
    // Actions
    loadPermissions,
    loadRoles,
    hasPermission,
    hasAllPermissions,
    hasRole,
    canAccessRoute,
    filterMenusByPermission,
    clearPermissions
  };
});
```

### Usage

```typescript
// In components
const permissionStore = usePermissionStore();

// Load permissions
await permissionStore.loadPermissions();

// Check permission
if (permissionStore.hasPermission('identity:user:create')) {
  // Show create button
}

// Check multiple permissions (OR)
if (permissionStore.hasPermission(['identity:user:create', 'identity:user:update'])) {
  // Has at least one permission
}

// Check multiple permissions (AND)
if (permissionStore.hasAllPermissions(['identity:user:view', 'identity:user:update'])) {
  // Has all permissions
}

// In templates with custom directive
<el-button v-permission="'identity:user:create'">Create User</el-button>
```

---

## 3. App Store

**Purpose**: Manages global app state (loading, breadcrumb, sidebar, theme).

**File**: `web/pc/src/stores/app.ts`

### State

```typescript
export const useAppStore = defineStore('app', () => {
  // State
  const loading = ref(false);
  const breadcrumb = ref<Breadcrumb[]>([]);
  const sidebarCollapsed = ref(false);
  const theme = ref<'light' | 'dark'>('light');

  // Actions
  const setLoading = (value: boolean): void => {
    loading.value = value;
  };

  const updateBreadcrumb = (route: RouteLocationNormalized): void => {
    if (route.meta.breadcrumb) {
      breadcrumb.value = route.meta.breadcrumb as Breadcrumb[];
    } else {
      const matched = route.matched.filter(r => r.meta?.title);
      breadcrumb.value = matched.map(r => ({
        title: r.meta.title as string,
        to: r.path === route.path ? undefined : r.path
      }));
    }
  };

  const toggleSidebar = (): void => {
    sidebarCollapsed.value = !sidebarCollapsed.value;
  };

  const setSidebarCollapsed = (value: boolean): void => {
    sidebarCollapsed.value = value;
  };

  const setTheme = (value: 'light' | 'dark'): void => {
    theme.value = value;
    document.documentElement.setAttribute('data-theme', value);
  };

  return {
    // State
    loading,
    breadcrumb,
    sidebarCollapsed,
    theme,
    // Actions
    setLoading,
    updateBreadcrumb,
    toggleSidebar,
    setSidebarCollapsed,
    setTheme
  };
}, {
  persist: {
    paths: ['sidebarCollapsed', 'theme']
  }
});
```

### Usage

```typescript
// In components
const appStore = useAppStore();

// Show loading
appStore.setLoading(true);
try {
  await fetchData();
} finally {
  appStore.setLoading(false);
}

// Toggle sidebar
appStore.toggleSidebar();

// Get breadcrumb
const breadcrumb = appStore.breadcrumb;
```

---

## 4. Division Store

**Purpose**: Caches administrative division tree for performance.

**File**: `web/pc/src/stores/division.ts`

### State

```typescript
export const useDivisionStore = defineStore('division', () => {
  // State
  const divisionTree = ref<AdministrativeDivision[]>([]);
  const divisionMap = ref<Map<number, AdministrativeDivision>>(new Map());
  const lastFetchTime = ref<number>(0);

  // Computed
  const isCacheValid = computed(() => {
    const now = Date.now();
    const fiveMinutes = 5 * 60 * 1000;
    return now - lastFetchTime.value < fiveMinutes;
  });

  // Actions
  const loadDivisions = async (forceRefresh = false): Promise<void> => {
    if (!forceRefresh && isCacheValid.value && divisionTree.value.length > 0) {
      return; // Use cached data
    }

    const response = await masterdataApi.getDivisions();
    divisionTree.value = response;
    buildDivisionMap(response);
    lastFetchTime.value = Date.now();
  };

  const buildDivisionMap = (divisions: AdministrativeDivision[]): void => {
    divisionMap.value.clear();
    const traverse = (nodes: AdministrativeDivision[]) => {
      nodes.forEach(node => {
        divisionMap.value.set(node.id, node);
        if (node.children) {
          traverse(node.children);
        }
      });
    };
    traverse(divisions);
  };

  const getDivisionById = (id: number): AdministrativeDivision | undefined => {
    return divisionMap.value.get(id);
  };

  const getDivisionPath = (id: number): string => {
    const division = getDivisionById(id);
    if (!division) return '';

    const path: string[] = [division.name];
    let current = division;
    
    while (current.parentId !== 0) {
      const parent = getDivisionById(current.parentId);
      if (!parent) break;
      path.unshift(parent.name);
      current = parent;
    }

    return path.join(' > ');
  };

  const getDivisionsByLevel = (level: DivisionLevel): AdministrativeDivision[] => {
    return Array.from(divisionMap.value.values()).filter(d => d.level === level);
  };

  const getChildDivisions = (parentId: number): AdministrativeDivision[] => {
    return Array.from(divisionMap.value.values()).filter(d => d.parentId === parentId);
  };

  const clearCache = (): void => {
    divisionTree.value = [];
    divisionMap.value.clear();
    lastFetchTime.value = 0;
  };

  return {
    // State
    divisionTree,
    divisionMap,
    lastFetchTime,
    // Computed
    isCacheValid,
    // Actions
    loadDivisions,
    getDivisionById,
    getDivisionPath,
    getDivisionsByLevel,
    getChildDivisions,
    clearCache
  };
});
```

### Usage

```typescript
// In components
const divisionStore = useDivisionStore();

// Load divisions (uses cache if valid)
await divisionStore.loadDivisions();

// Force refresh
await divisionStore.loadDivisions(true);

// Get division by ID
const division = divisionStore.getDivisionById(123);

// Get full path
const path = divisionStore.getDivisionPath(123);
// Returns: "Guangdong > Shenzhen > Nanshan"

// Get all provinces
const provinces = divisionStore.getDivisionsByLevel(DivisionLevel.Province);

// Get children
const cities = divisionStore.getChildDivisions(provinceId);
```

---

## 5. User Store

**Purpose**: Manages user list state and filters for user management pages.

**File**: `web/pc/src/stores/user.ts`

### State

```typescript
export const useUserStore = defineStore('user', () => {
  // State
  const users = ref<User[]>([]);
  const total = ref(0);
  const currentPage = ref(1);
  const pageSize = ref(20);
  const filters = ref<UserFilter>({
    userType: undefined,
    status: undefined,
    verificationStatus: undefined,
    keyword: ''
  });

  // Actions
  const loadUsers = async (): Promise<void> => {
    const response = await identityApi.getUsers({
      page: currentPage.value,
      pageSize: pageSize.value,
      ...filters.value
    });
    users.value = response.list;
    total.value = response.total;
  };

  const setPage = (page: number): void => {
    currentPage.value = page;
  };

  const setPageSize = (size: number): void => {
    pageSize.value = size;
    currentPage.value = 1; // Reset to first page
  };

  const setFilters = (newFilters: Partial<UserFilter>): void => {
    filters.value = { ...filters.value, ...newFilters };
    currentPage.value = 1; // Reset to first page
  };

  const clearFilters = (): void => {
    filters.value = {
      userType: undefined,
      status: undefined,
      verificationStatus: undefined,
      keyword: ''
    };
    currentPage.value = 1;
  };

  const getUserById = (id: number): User | undefined => {
    return users.value.find(u => u.id === id);
  };

  const updateUserInList = (updatedUser: User): void => {
    const index = users.value.findIndex(u => u.id === updatedUser.id);
    if (index !== -1) {
      users.value[index] = updatedUser;
    }
  };

  const removeUserFromList = (id: number): void => {
    const index = users.value.findIndex(u => u.id === id);
    if (index !== -1) {
      users.value.splice(index, 1);
      total.value--;
    }
  };

  return {
    // State
    users,
    total,
    currentPage,
    pageSize,
    filters,
    // Actions
    loadUsers,
    setPage,
    setPageSize,
    setFilters,
    clearFilters,
    getUserById,
    updateUserInList,
    removeUserFromList
  };
});
```

### Usage

```typescript
// In components
const userStore = useUserStore();

// Load users with current filters
await userStore.loadUsers();

// Set filters
userStore.setFilters({ userType: UserType.Staff, status: UserStatus.Active });

// Clear filters
userStore.clearFilters();

// Pagination
userStore.setPage(2);
userStore.setPageSize(50);

// Update user in list after edit
userStore.updateUserInList(updatedUser);
```

---

## Store Persistence

### Pinia Persist Plugin

```typescript
// web/pc/src/main.ts
import { createPinia } from 'pinia';
import piniaPluginPersistedstate from 'pinia-plugin-persistedstate';

const pinia = createPinia();
pinia.use(piniaPluginPersistedstate);

app.use(pinia);
```

### Persistence Configuration

**Auth Store**:
- Persists: `user` (user info)
- Storage: localStorage
- Does NOT persist: tokens (stored separately in sessionStorage/localStorage)

**App Store**:
- Persists: `sidebarCollapsed`, `theme`
- Storage: localStorage

**Permission Store**:
- No persistence (loaded on login)

**Division Store**:
- No persistence (cached in memory with TTL)

**User Store**:
- No persistence (page-specific state)

---

## Store Initialization

### On App Mount

```typescript
// web/pc/src/App.vue
import { useAuthStore } from '@/stores/auth';
import { usePermissionStore } from '@/stores/permission';

const authStore = useAuthStore();
const permissionStore = usePermissionStore();

onMounted(async () => {
  // Restore session from storage
  authStore.restoreSession();

  // Load permissions if authenticated
  if (authStore.isAuthenticated) {
    try {
      await permissionStore.loadPermissions();
    } catch (error) {
      console.error('Failed to load permissions:', error);
      authStore.logout();
      router.push('/login');
    }
  }
});
```

### On Login

```typescript
// After successful login
await authStore.login(phone, password);
await permissionStore.loadPermissions();
router.push('/dashboard');
```

### On Logout

```typescript
// Clear all stores
await authStore.logout();
permissionStore.clearPermissions();
divisionStore.clearCache();
router.push('/login');
```

---

## Store Communication

### Cross-Store Dependencies

```typescript
// Permission store depends on auth store
const loadPermissions = async (): Promise<void> => {
  const authStore = useAuthStore();
  if (!authStore.user) {
    throw new Error('User not authenticated');
  }
  // ... load permissions
};

// Division store can be used by multiple feature stores
const communityStore = useCommunityStore();
const divisionStore = useDivisionStore();

const loadCommunityWithDivision = async (id: number) => {
  const community = await communityStore.loadCommunity(id);
  const divisionPath = divisionStore.getDivisionPath(community.divisionId);
  return { ...community, divisionPath };
};
```

---

## Summary

This state contract defines 5 Pinia stores:

1. **Auth Store**: Authentication, tokens, user session
2. **Permission Store**: Permissions, roles, menu access
3. **App Store**: Global UI state (loading, breadcrumb, sidebar, theme)
4. **Division Store**: Cached division tree with 5-minute TTL
5. **User Store**: User list state and filters

All stores use:
- **Composition API**: `defineStore` with setup function
- **TypeScript**: Strict typing for state, actions, computed
- **Selective persistence**: Only user preferences and session data
- **Clear separation**: Each store has single responsibility
- **Cross-store communication**: Stores can depend on each other

These stores support all 8 user stories and align with Constitution Principle VI (Frontend Development Specifications).
