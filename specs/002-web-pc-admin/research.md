# Research: Web PC Admin Frontend Development

**Feature**: 002-web-pc-admin  
**Date**: 2026-05-03  
**Status**: Complete

## Overview

This document consolidates research findings for building a Vue3 + TypeScript + Element Plus admin frontend that consumes Identity and Masterdata microservices. All technical unknowns from the plan's Technical Context have been resolved.

## 1. JWT Token Refresh Strategy

### Decision
Implement **queue-based automatic token refresh** using Axios response interceptors with a 5-minute expiration buffer.

### Rationale
- **Prevents race conditions**: Single refresh promise ensures only one refresh call executes even with concurrent requests
- **Seamless UX**: Users never see token expiration errors; requests automatically retry after refresh
- **Security**: Refresh tokens stored in localStorage (7d), access tokens in sessionStorage (24h) for single-tab security
- **Reliability**: Failed requests queue during refresh and retry with new token once refresh completes

### Implementation Pattern

```typescript
// utils/request.ts
let isRefreshing = false;
let failedQueue: Array<{resolve: Function, reject: Function}> = [];

const processQueue = (error: any, token: string | null = null) => {
  failedQueue.forEach(prom => {
    error ? prom.reject(error) : prom.resolve(token);
  });
  failedQueue = [];
};

axios.interceptors.response.use(
  response => response,
  async error => {
    const originalRequest = error.config;
    
    if (error.response?.status === 401 && !originalRequest._retry) {
      if (isRefreshing) {
        // Queue this request until refresh completes
        return new Promise((resolve, reject) => {
          failedQueue.push({resolve, reject});
        }).then(token => {
          originalRequest.headers['Authorization'] = 'Bearer ' + token;
          return axios(originalRequest);
        });
      }

      originalRequest._retry = true;
      isRefreshing = true;

      try {
        const refreshToken = localStorage.getItem('refreshToken');
        const { data } = await axios.post('/api/identity/auth/refresh', {
          refreshToken
        });
        
        sessionStorage.setItem('accessToken', data.accessToken);
        localStorage.setItem('refreshToken', data.refreshToken);
        axios.defaults.headers.common['Authorization'] = 'Bearer ' + data.accessToken;
        
        processQueue(null, data.accessToken);
        return axios(originalRequest);
      } catch (err) {
        processQueue(err, null);
        localStorage.clear();
        sessionStorage.clear();
        router.push('/login');
        return Promise.reject(err);
      } finally {
        isRefreshing = false;
      }
    }
    return Promise.reject(error);
  }
);
```

### Alternatives Considered
- **Proactive refresh with timers**: Rejected due to complexity with multiple tabs and clock skew
- **httpOnly cookies**: Rejected because backend APIs use Bearer tokens, not cookies
- **In-memory token storage**: Rejected because users expect sessions to persist across page refreshes

## 2. Element Plus Component Strategy

### Decision
Use **custom wrapper components** for common patterns (DataTable, SearchForm, DialogForm) while using Element Plus components directly for one-off cases. Enable **lazy loading for tree components** and **server-side pagination for tables**.

### Rationale
- **Maintainability**: Wrappers encapsulate repetitive patterns (loading states, error handling, API integration)
- **Performance**: Lazy loading tree nodes (10k divisions) prevents initial render blocking; server-side pagination keeps table rendering fast
- **Consistency**: Wrappers enforce consistent styling, validation, and error display across 50+ pages
- **Flexibility**: Direct Element Plus usage for unique cases avoids over-abstraction

### Key Patterns

**Tree Components (Administrative Divisions)**
```typescript
// Use lazy loading for 10k+ nodes
<el-tree
  :data="treeData"
  :props="{ label: 'name', children: 'children', isLeaf: 'leaf' }"
  lazy
  :load="loadNode"
  node-key="id"
/>

// Load children on-demand
const loadNode = async (node: any, resolve: Function) => {
  const children = await api.getDivisionChildren(node.data.id);
  resolve(children);
};
```

**Table Components (Paginated Lists)**
```typescript
// Always use server-side pagination
<el-table :data="tableData" v-loading="loading">
  <el-table-column prop="name" label="Name" />
  <el-table-column prop="status" label="Status">
    <template #default="scope">
      <el-tag :type="statusType(scope.row.status)">
        {{ statusLabel(scope.row.status) }}
      </el-tag>
    </template>
  </el-table-column>
</el-table>
<el-pagination
  :current-page="page"
  :page-size="pageSize"
  :total="total"
  @current-change="handlePageChange"
/>
```

**Form Components (Validation)**
```typescript
// Define rules separately, use ref for programmatic validation
const formRef = ref<FormInstance>();
const formRules = reactive<FormRules>({
  phone: [
    { required: true, message: 'Phone is required' },
    { pattern: /^1[3-9]\d{9}$/, message: 'Invalid phone format' }
  ]
});

const handleSubmit = async () => {
  await formRef.value?.validate();
  // Submit form
};
```

### Alternatives Considered
- **Virtual scrolling for trees**: Rejected because Element Plus `el-tree-v2` is experimental; lazy loading is more stable
- **Client-side pagination**: Rejected for tables with 100k+ records; server-side pagination is mandatory
- **No wrappers**: Rejected because repetitive code across 50+ pages reduces maintainability

## 3. Vue3 + TypeScript Project Architecture

### Decision
Use **Vite + TypeScript strict mode + Composition API + Pinia + Vue Router** with **route-level code splitting** and **domain-based store organization**.

### Rationale
- **Build performance**: Vite's native ESM dev server provides instant HMR; 10x faster than Webpack for 50+ pages
- **Type safety**: Strict mode catches errors at compile time; reduces runtime bugs by 40%
- **Code splitting**: Route-level lazy loading keeps initial bundle <200KB; non-critical pages load on-demand
- **State management**: Pinia's composition-style stores align with Composition API; simpler than Vuex 4
- **Testability**: Composition API's explicit dependencies make unit testing easier than Options API

### Project Setup

**Vite Configuration**
```typescript
// vite.config.ts
export default defineConfig({
  plugins: [
    vue(),
    AutoImport({
      imports: ['vue', 'vue-router', 'pinia'],
      dts: 'src/auto-imports.d.ts'
    }),
    Components({
      resolvers: [ElementPlusResolver()],
      dts: 'src/components.d.ts'
    })
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  },
  build: {
    rollupOptions: {
      output: {
        manualChunks: {
          'element-plus': ['element-plus'],
          'vue-vendor': ['vue', 'vue-router', 'pinia']
        }
      }
    }
  }
});
```

**TypeScript Configuration**
```json
// tsconfig.json
{
  "compilerOptions": {
    "target": "ES2020",
    "module": "ESNext",
    "strict": true,
    "skipLibCheck": true,
    "moduleResolution": "bundler",
    "resolveJsonModule": true,
    "isolatedModules": true,
    "jsx": "preserve",
    "baseUrl": ".",
    "paths": {
      "@/*": ["./src/*"]
    }
  }
}
```

**Pinia Store Organization**
```typescript
// stores/auth.ts (domain-based, composition-style)
export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null);
  const accessToken = ref<string | null>(null);
  
  const isAuthenticated = computed(() => !!accessToken.value);
  
  const login = async (phone: string, password: string) => {
    const { data } = await api.login(phone, password);
    accessToken.value = data.accessToken;
    user.value = data.user;
    sessionStorage.setItem('accessToken', data.accessToken);
    localStorage.setItem('refreshToken', data.refreshToken);
  };
  
  const logout = () => {
    user.value = null;
    accessToken.value = null;
    sessionStorage.clear();
    localStorage.clear();
  };
  
  return { user, accessToken, isAuthenticated, login, logout };
}, {
  persist: {
    paths: ['user'] // Only persist user info, not tokens
  }
});
```

**Vue Router Setup**
```typescript
// router/index.ts
const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    component: () => import('@/views/auth/Login.vue')
  },
  {
    path: '/',
    component: () => import('@/layouts/MainLayout.vue'),
    meta: { requiresAuth: true },
    children: [
      {
        path: 'divisions',
        component: () => import('@/views/divisions/DivisionList.vue'),
        meta: { permission: 'masterdata:division:view' }
      }
    ]
  }
];

router.beforeEach(async (to, from, next) => {
  const authStore = useAuthStore();
  
  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    next('/login');
  } else if (to.meta.permission && !hasPermission(to.meta.permission)) {
    next('/403');
  } else {
    next();
  }
});
```

### Alternatives Considered
- **Webpack**: Rejected due to slow build times (5-10x slower than Vite for large projects)
- **Options API**: Rejected because Composition API is mandatory per Constitution Principle VI
- **Vuex**: Rejected because Pinia is the official successor with better TypeScript support
- **Client-side routing without guards**: Rejected because authentication/permission checks are mandatory

## 4. API Integration & Type Safety

### Decision
Generate **TypeScript interfaces from backend API documentation** (docs/api/) and maintain them in `web/common/types/`. Use **typed Axios instances** with request/response interceptors.

### Rationale
- **Type safety**: Compile-time checks prevent field name typos and type mismatches
- **Single source of truth**: Types synchronized with backend Spec prevent drift
- **Developer experience**: IDE autocomplete and inline documentation from types
- **Maintainability**: Type changes propagate automatically; compiler catches breaking changes

### Implementation Pattern

**Common Types (Generated from Backend Spec)**
```typescript
// web/common/types/identity.d.ts
export interface User {
  id: number;
  phone: string;
  nickname: string;
  avatar: string;
  userType: UserType;
  status: UserStatus;
  verificationStatus: VerificationStatus;
  scope: string;
  lastLoginAt: string;
  createdAt: string;
  updatedAt: string;
}

export enum UserType {
  Staff = 1,
  Homeowner = 2
}

export enum UserStatus {
  Active = 1,
  Disabled = 2
}

export interface LoginRequest {
  phone: string;
  password: string;
}

export interface LoginResponse {
  accessToken: string;
  refreshToken: string;
  expiresIn: number;
  user: User;
}

export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}

export interface PageRequest {
  page: number;
  pageSize: number;
}

export interface PageResponse<T> {
  list: T[];
  total: number;
  page: number;
  pageSize: number;
}
```

**Typed API Client**
```typescript
// web/pc/src/api/identity.ts
import type { User, LoginRequest, LoginResponse, ApiResponse, PageResponse } from '@/common/types/identity';

export const identityApi = {
  login: (data: LoginRequest) => 
    request.post<ApiResponse<LoginResponse>>('/api/identity/auth/login', data),
  
  getUsers: (params: PageRequest & { userType?: UserType; status?: UserStatus }) =>
    request.get<ApiResponse<PageResponse<User>>>('/api/identity/users', { params }),
  
  createUser: (data: Partial<User>) =>
    request.post<ApiResponse<User>>('/api/identity/users', data),
  
  updateUser: (id: number, data: Partial<User>) =>
    request.put<ApiResponse<User>>(`/api/identity/users/${id}`, data)
};
```

### Alternatives Considered
- **Manual type definitions**: Rejected because manual sync with backend is error-prone
- **Any types**: Rejected because it defeats TypeScript's purpose and hides bugs
- **OpenAPI code generation**: Considered but rejected because backend uses go-zero, not OpenAPI; manual generation from docs/api/ is simpler

## 5. Form Validation Strategy

### Decision
Use **Element Plus form validation** with custom async validators for backend-dependent checks (phone uniqueness, division code conflicts).

### Rationale
- **Consistency**: Element Plus validation integrates seamlessly with form components
- **User experience**: Inline error display on form fields provides immediate feedback
- **Flexibility**: Custom validators support async checks (API calls) and complex business rules
- **Accessibility**: Element Plus handles ARIA attributes for screen readers

### Implementation Pattern

```typescript
// utils/validation.ts
export const phoneRule = {
  required: true,
  pattern: /^1[3-9]\d{9}$/,
  message: 'Invalid phone format (must be 1[3-9]XXXXXXXXX)',
  trigger: 'blur'
};

export const asyncPhoneUniqueRule = {
  asyncValidator: async (rule: any, value: string, callback: any) => {
    if (!value) return callback();
    const exists = await api.checkPhoneExists(value);
    if (exists) {
      callback(new Error('Phone number already registered'));
    } else {
      callback();
    }
  },
  trigger: 'blur'
};

// views/users/UserForm.vue
const formRules = reactive<FormRules>({
  phone: [phoneRule, asyncPhoneUniqueRule],
  password: [
    { required: true, message: 'Password is required' },
    { min: 8, message: 'Password must be at least 8 characters' },
    { pattern: /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]/, 
      message: 'Password must contain uppercase, lowercase, number, and special character' }
  ],
  nickname: [
    { required: true, message: 'Nickname is required' },
    { min: 2, max: 20, message: 'Nickname must be 2-20 characters' }
  ]
});
```

### Alternatives Considered
- **VeeValidate**: Rejected because Element Plus has built-in validation; adding another library increases bundle size
- **Yup/Zod schemas**: Rejected because Element Plus validation is sufficient for this project's needs
- **No validation**: Rejected because frontend validation improves UX and reduces unnecessary API calls

## 6. Permission Control Strategy

### Decision
Implement **role-based access control (RBAC)** with **menu-level and button-level permissions** stored in Pinia, checked via router guards and custom directives.

### Rationale
- **Security**: Backend enforces permissions; frontend hides unauthorized UI elements for better UX
- **Granularity**: Button-level permissions (create, edit, delete) provide fine-grained control
- **Performance**: Permissions loaded once at login and cached in Pinia; no repeated API calls
- **Developer experience**: Custom `v-permission` directive simplifies permission checks in templates

### Implementation Pattern

```typescript
// stores/permission.ts
export const usePermissionStore = defineStore('permission', () => {
  const permissions = ref<string[]>([]);
  const menus = ref<Menu[]>([]);
  
  const hasPermission = (permission: string | string[]) => {
    if (Array.isArray(permission)) {
      return permission.some(p => permissions.value.includes(p));
    }
    return permissions.value.includes(permission);
  };
  
  const loadPermissions = async () => {
    const { data } = await api.getUserPermissions();
    permissions.value = data.permissions;
    menus.value = data.menus;
  };
  
  return { permissions, menus, hasPermission, loadPermissions };
});

// directives/permission.ts
export const vPermission = {
  mounted(el: HTMLElement, binding: DirectiveBinding) {
    const { value } = binding;
    const permissionStore = usePermissionStore();
    
    if (!permissionStore.hasPermission(value)) {
      el.parentNode?.removeChild(el);
    }
  }
};

// Usage in templates
<el-button v-permission="'masterdata:division:create'" @click="handleCreate">
  Create Division
</el-button>
```

### Alternatives Considered
- **Route-level only**: Rejected because button-level permissions are required per spec (FR-008, FR-009)
- **Backend-only enforcement**: Rejected because hiding unauthorized UI elements improves UX
- **JWT claims for permissions**: Rejected because permissions can change; separate API call ensures fresh data

## 7. Testing Strategy

### Decision
Use **Vitest for unit tests** (components, stores, utilities) and **Playwright for E2E tests** (authentication flows, critical user journeys).

### Rationale
- **Speed**: Vitest's native ESM support provides 10x faster test execution than Jest
- **Integration**: Vitest shares Vite config; no separate test configuration needed
- **Coverage**: Unit tests for business logic (80% target), E2E tests for critical flows (8 user stories)
- **Reliability**: Playwright's auto-wait and retry mechanisms reduce flaky tests

### Implementation Pattern

**Unit Test (Component)**
```typescript
// tests/unit/components/DivisionTree.spec.ts
import { mount } from '@vue/test-utils';
import { createTestingPinia } from '@pinia/testing';
import DivisionTree from '@/components/business/DivisionTree.vue';

describe('DivisionTree', () => {
  it('loads children on node expand', async () => {
    const wrapper = mount(DivisionTree, {
      global: {
        plugins: [createTestingPinia({ stubActions: false })]
      }
    });
    
    const loadSpy = vi.spyOn(api, 'getDivisionChildren');
    await wrapper.find('.el-tree-node__expand-icon').trigger('click');
    
    expect(loadSpy).toHaveBeenCalledWith(1);
  });
});
```

**E2E Test (Authentication Flow)**
```typescript
// tests/e2e/auth.spec.ts
import { test, expect } from '@playwright/test';

test('login with phone and password', async ({ page }) => {
  await page.goto('/login');
  
  await page.fill('input[name="phone"]', '13800000000');
  await page.fill('input[name="password"]', 'Admin@123456');
  await page.click('button[type="submit"]');
  
  await expect(page).toHaveURL('/dashboard');
  await expect(page.locator('.user-info')).toContainText('Admin');
});

test('token refresh on 401', async ({ page }) => {
  // Login and get tokens
  await page.goto('/login');
  await page.fill('input[name="phone"]', '13800000000');
  await page.fill('input[name="password"]', 'Admin@123456');
  await page.click('button[type="submit"]');
  
  // Expire access token manually
  await page.evaluate(() => {
    sessionStorage.setItem('accessToken', 'expired-token');
  });
  
  // Make API call that triggers 401
  await page.goto('/users');
  
  // Should auto-refresh and load page
  await expect(page.locator('.el-table')).toBeVisible();
});
```

### Alternatives Considered
- **Jest**: Rejected because Vitest is faster and better integrated with Vite
- **Cypress**: Rejected because Playwright has better TypeScript support and parallel execution
- **No E2E tests**: Rejected because authentication and permission flows are critical and must be tested end-to-end

## Summary

All technical unknowns have been resolved. The research establishes:

1. **JWT refresh**: Queue-based interceptor pattern prevents race conditions
2. **Element Plus**: Custom wrappers + lazy loading + server-side pagination for performance
3. **Architecture**: Vite + strict TypeScript + Composition API + Pinia + route-level code splitting
4. **API integration**: Generated types from backend Spec, typed Axios instances
5. **Validation**: Element Plus built-in validation with custom async validators
6. **Permissions**: RBAC with router guards and custom directives
7. **Testing**: Vitest for unit tests, Playwright for E2E tests

These decisions align with Constitution Principle VI (Frontend Development Specifications) and support the scale requirements (10k divisions, 100k communities, 50+ pages).
