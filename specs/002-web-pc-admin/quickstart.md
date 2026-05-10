# Quick Start Guide: Web PC Admin Frontend

**Feature**: 002-web-pc-admin  
**Date**: 2026-05-03  
**Status**: Complete

## Overview

This guide provides step-by-step instructions for setting up and developing the PC admin frontend for the Community & Home Management Platform.

## Prerequisites

- **Node.js**: 18.0+ or 20.0+ (LTS recommended)
- **pnpm**: 8.0+ (preferred) or npm 9.0+
- **Git**: 2.30+
- **IDE**: VS Code with Volar extension (recommended)
- **Backend Services**: Identity API (port 8888) and Masterdata API (port 8889) running

## Project Setup

### 1. Initialize Project

```bash
# Navigate to project root
cd /home/jiaoxh/my-code/community-and-home

# Create web directory structure
mkdir -p web/pc web/mobile web/common

# Navigate to PC project
cd web/pc

# Initialize Vite + Vue3 + TypeScript project
pnpm create vite . --template vue-ts

# Install dependencies
pnpm install
```

### 2. Install Core Dependencies

```bash
# UI Framework
pnpm add element-plus @element-plus/icons-vue

# HTTP Client
pnpm add axios

# State Management
pnpm add pinia pinia-plugin-persistedstate

# Routing
pnpm add vue-router@4

# Utilities
pnpm add dayjs lodash-es

# Dev Dependencies
pnpm add -D @types/lodash-es
pnpm add -D unplugin-vue-components unplugin-auto-import
pnpm add -D sass
pnpm add -D @vitejs/plugin-vue
pnpm add -D vite-plugin-compression
pnpm add -D rollup-plugin-visualizer
```

### 3. Install Testing Dependencies

```bash
# Unit Testing
pnpm add -D vitest @vue/test-utils @pinia/testing happy-dom

# E2E Testing
pnpm add -D @playwright/test
pnpm exec playwright install
```

### 4. Configure Vite

Create `vite.config.ts`:

```typescript
import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';
import { fileURLToPath, URL } from 'node:url';
import AutoImport from 'unplugin-auto-import/vite';
import Components from 'unplugin-vue-components/vite';
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers';
import compression from 'vite-plugin-compression';
import { visualizer } from 'rollup-plugin-visualizer';

export default defineConfig({
  plugins: [
    vue(),
    AutoImport({
      imports: ['vue', 'vue-router', 'pinia'],
      resolvers: [ElementPlusResolver()],
      dts: 'src/auto-imports.d.ts'
    }),
    Components({
      resolvers: [ElementPlusResolver()],
      dts: 'src/components.d.ts'
    }),
    compression({
      algorithm: 'gzip',
      ext: '.gz'
    }),
    visualizer({
      open: false,
      gzipSize: true,
      brotliSize: true
    })
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
      '@common': fileURLToPath(new URL('../common', import.meta.url))
    }
  },
  server: {
    port: 3000,
    proxy: {
      '/api/identity': {
        target: 'http://localhost:8888',
        changeOrigin: true
      },
      '/api/masterdata': {
        target: 'http://localhost:8889',
        changeOrigin: true
      }
    }
  },
  build: {
    rollupOptions: {
      output: {
        manualChunks: {
          'element-plus': ['element-plus'],
          'element-icons': ['@element-plus/icons-vue'],
          'vue-vendor': ['vue', 'vue-router', 'pinia']
        }
      }
    },
    chunkSizeWarningLimit: 1000
  }
});
```

### 5. Configure TypeScript

Create `tsconfig.json`:

```json
{
  "compilerOptions": {
    "target": "ES2020",
    "useDefineForClassFields": true,
    "module": "ESNext",
    "lib": ["ES2020", "DOM", "DOM.Iterable"],
    "skipLibCheck": true,

    /* Bundler mode */
    "moduleResolution": "bundler",
    "allowImportingTsExtensions": true,
    "resolveJsonModule": true,
    "isolatedModules": true,
    "noEmit": true,
    "jsx": "preserve",

    /* Linting */
    "strict": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "noFallthroughCasesInSwitch": true,

    /* Path Aliases */
    "baseUrl": ".",
    "paths": {
      "@/*": ["./src/*"],
      "@common/*": ["../common/*"]
    }
  },
  "include": ["src/**/*.ts", "src/**/*.d.ts", "src/**/*.tsx", "src/**/*.vue"],
  "references": [{ "path": "./tsconfig.node.json" }]
}
```

### 6. Create Directory Structure

```bash
# Create source directories
mkdir -p src/{api,views,components,stores,router,utils,types,styles,assets}
mkdir -p src/views/{auth,divisions,communities,users,roles,verification,config,sensitive-words,dashboard,error}
mkdir -p src/components/{layout,common,business}

# Create test directories
mkdir -p tests/{unit,e2e}
```

## Development Workflow

### 1. Start Development Server

```bash
# Terminal 1: Start backend services (if not running)
cd services/identity/api
go run identity.go -f etc/identity-api.yaml

# Terminal 2: Start Masterdata API
cd services/masterdata/api
go run masterdata.go -f etc/masterdata-api.yaml

# Terminal 3: Start frontend dev server
cd web/pc
pnpm dev
```

Access the application at `http://localhost:3000`

### 2. Login with Default Credentials

- **Phone**: `13800000000`
- **Password**: `Admin@123456`
- **Role**: Super Administrator

### 3. Development Commands

```bash
# Run dev server with HMR
pnpm dev

# Type check
pnpm vue-tsc --noEmit

# Build for production
pnpm build

# Preview production build
pnpm preview

# Run unit tests
pnpm test:unit

# Run E2E tests
pnpm test:e2e

# Run all tests
pnpm test

# Lint code
pnpm lint

# Format code
pnpm format
```

## Project Structure

```
web/pc/
├── src/
│   ├── api/                    # API client functions
│   │   ├── identity.ts         # Identity service API
│   │   ├── masterdata.ts       # Masterdata service API
│   │   └── index.ts
│   ├── views/                  # Page components
│   │   ├── auth/               # Login, Register
│   │   ├── dashboard/          # Dashboard
│   │   ├── divisions/          # Division management
│   │   ├── communities/        # Community management
│   │   ├── users/              # User management
│   │   ├── roles/              # Role & permission management
│   │   ├── verification/       # Verification review
│   │   ├── config/             # System configuration
│   │   ├── sensitive-words/    # Sensitive word management
│   │   └── error/              # Error pages (403, 404)
│   ├── components/             # Reusable components
│   │   ├── layout/             # Layout components
│   │   ├── common/             # Common UI components
│   │   └── business/           # Business components
│   ├── stores/                 # Pinia stores
│   │   ├── auth.ts
│   │   ├── permission.ts
│   │   ├── app.ts
│   │   ├── division.ts
│   │   └── user.ts
│   ├── router/                 # Vue Router
│   │   ├── index.ts
│   │   ├── routes.ts
│   │   └── guards.ts
│   ├── utils/                  # Utility functions
│   │   ├── request.ts          # Axios instance
│   │   ├── validation.ts       # Form validation
│   │   ├── permission.ts       # Permission helpers
│   │   └── format.ts           # Data formatting
│   ├── types/                  # TypeScript types
│   │   ├── views.d.ts
│   │   └── components.d.ts
│   ├── styles/                 # Global styles
│   │   ├── variables.scss
│   │   ├── mixins.scss
│   │   └── global.scss
│   ├── App.vue                 # Root component
│   └── main.ts                 # Entry point
├── public/                     # Static assets
├── tests/                      # Tests
│   ├── unit/
│   └── e2e/
├── index.html
├── vite.config.ts
├── tsconfig.json
├── package.json
└── README.md
```

## Common Development Tasks

### Create a New Page

1. **Create view component**:
```bash
# Example: Create user list page
touch src/views/users/List.vue
```

2. **Add route**:
```typescript
// src/router/routes.ts
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
}
```

3. **Create API functions**:
```typescript
// src/api/identity.ts
export const getUsers = (params: UserFilter) => 
  request.get<ApiResponse<PageResponse<User>>>('/api/identity/users', { params });
```

4. **Implement component**:
```vue
<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { getUsers } from '@/api/identity';

const users = ref<User[]>([]);
const loading = ref(false);

const loadUsers = async () => {
  loading.value = true;
  try {
    const response = await getUsers({ page: 1, pageSize: 20 });
    users.value = response.list;
  } finally {
    loading.value = false;
  }
};

onMounted(() => {
  loadUsers();
});
</script>

<template>
  <div class="user-list">
    <el-table :data="users" v-loading="loading">
      <el-table-column prop="phone" label="Phone" />
      <el-table-column prop="nickname" label="Nickname" />
    </el-table>
  </div>
</template>
```

### Add a New API Endpoint

1. **Define types in common**:
```typescript
// web/common/types/identity.d.ts
export interface NewEntity {
  id: number;
  name: string;
}
```

2. **Create API function**:
```typescript
// web/pc/src/api/identity.ts
export const getNewEntity = (id: number) =>
  request.get<ApiResponse<NewEntity>>(`/api/identity/new-entity/${id}`);
```

3. **Use in component**:
```typescript
import { getNewEntity } from '@/api/identity';

const entity = await getNewEntity(123);
```

### Add Permission Check

1. **In template**:
```vue
<el-button v-permission="'identity:user:create'">
  Create User
</el-button>
```

2. **In script**:
```typescript
import { usePermissionStore } from '@/stores/permission';

const permissionStore = usePermissionStore();

if (permissionStore.hasPermission('identity:user:create')) {
  // Show create button
}
```

3. **In route**:
```typescript
{
  path: 'users/create',
  meta: {
    permission: 'identity:user:create'
  }
}
```

## Testing

### Unit Test Example

```typescript
// tests/unit/components/UserList.spec.ts
import { mount } from '@vue/test-utils';
import { createTestingPinia } from '@pinia/testing';
import UserList from '@/views/users/List.vue';

describe('UserList', () => {
  it('loads users on mount', async () => {
    const wrapper = mount(UserList, {
      global: {
        plugins: [createTestingPinia({ stubActions: false })]
      }
    });

    await wrapper.vm.$nextTick();
    expect(wrapper.find('.el-table').exists()).toBe(true);
  });
});
```

### E2E Test Example

```typescript
// tests/e2e/auth.spec.ts
import { test, expect } from '@playwright/test';

test('login flow', async ({ page }) => {
  await page.goto('http://localhost:3000/login');
  
  await page.fill('input[name="phone"]', '13800000000');
  await page.fill('input[name="password"]', 'Admin@123456');
  await page.click('button[type="submit"]');
  
  await expect(page).toHaveURL('http://localhost:3000/dashboard');
});
```

## Troubleshooting

### Backend API Not Responding

**Problem**: API calls return 404 or connection refused

**Solution**:
1. Verify backend services are running on ports 8888 and 8889
2. Check Vite proxy configuration in `vite.config.ts`
3. Test API directly: `curl http://localhost:8888/api/identity/health`

### Token Refresh Not Working

**Problem**: User gets logged out unexpectedly

**Solution**:
1. Check token expiry in localStorage: `localStorage.getItem('tokenExpiry')`
2. Verify refresh token is valid: `localStorage.getItem('refreshToken')`
3. Check Axios interceptor in `src/utils/request.ts`

### Permission Denied Errors

**Problem**: User sees 403 errors or missing menu items

**Solution**:
1. Verify user has correct roles assigned
2. Check permission store: `permissionStore.permissions`
3. Verify permission codes match backend: `identity:user:view`

### Build Errors

**Problem**: TypeScript errors during build

**Solution**:
1. Run type check: `pnpm vue-tsc --noEmit`
2. Check `tsconfig.json` configuration
3. Verify all imports have correct types

## Next Steps

1. **Implement Authentication** (User Story 1)
   - Create login/register pages
   - Implement token management
   - Set up Axios interceptors

2. **Implement Division Management** (User Story 2)
   - Create division tree component
   - Implement CRUD operations
   - Add lazy loading for large trees

3. **Implement Community Management** (User Story 3)
   - Create community list and forms
   - Implement submission workflow
   - Add review interface for headquarters

4. **Continue with remaining user stories** (4-8)

## Resources

- **Vue 3 Documentation**: https://vuejs.org/
- **Element Plus Documentation**: https://element-plus.org/
- **Pinia Documentation**: https://pinia.vuejs.org/
- **Vue Router Documentation**: https://router.vuejs.org/
- **Vite Documentation**: https://vitejs.dev/
- **Backend API Documentation**: `/docs/api/README.md`

## Support

For questions or issues:
1. Check backend API documentation in `docs/api/`
2. Review feature spec in `specs/002-web-pc-admin/spec.md`
3. Review implementation plan in `specs/002-web-pc-admin/plan.md`
4. Check constitution for development guidelines in `.specify/memory/constitution.md`
