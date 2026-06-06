# Uni-app Vue 3 Mobile Framework Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Initialize a standard Uni-app (Vue 3 + TypeScript + Vite) mobile front-end framework in `web/mobile/`, integrated with existing `web/common/` shared library and backend APIs.

**Architecture:** Uni-app Vue 3 with pages-based routing (`pages.json`), Pinia state management, uni-ui component library, lossless-json for Snowflake ID parsing. Reuses `web/common/` types, constants, and auth utilities via path alias `@common`. Each page is a self-contained directory under `src/pages/` following Uni-app conventions.

**Tech Stack:** Uni-app 3.x (Vue 3), Vite 5.x, TypeScript 5.x, Pinia, uni-ui, lossless-json, dayjs, SCSS

---

## File Structure (post-implementation)

```
web/mobile/
├── package.json                # Dependencies & scripts
├── vite.config.ts              # Vite config with uni-app plugin + proxy
├── tsconfig.json               # TS project references
├── tsconfig.app.json           # App TS config with @ and @common aliases
├── index.html                  # Entry HTML (Uni-app H5 mode)
├── pages.json                  # Uni-app page routing, tabBar, globalStyle
├── manifest.json               # Uni-app app manifest (H5 + mini-program)
├── uni.scss                    # Global SCSS variables (theme colors, spacing)
├── .env.development            # Dev environment variables
├── src/
│   ├── main.ts                 # Vue 3 entry: createSSRApp + Pinia install
│   ├── App.vue                 # Root component with lifecycle
│   ├── env.d.ts                # Vite env type declarations
│   ├── pages/                  # Page components (Uni-app convention)
│   │   ├── index/
│   │   │   └── index.vue       # Home/tab page placeholder
│   │   ├── discover/
│   │   │   └── discover.vue    # Discover tab placeholder
│   │   └── mine/
│   │       └── mine.vue        # My profile tab placeholder
│   ├── components/             # Shared components (empty initially)
│   ├── stores/                 # Pinia stores
│   │   └── user.ts             # User auth store (token, profile, login/logout)
│   ├── api/                    # API call layer
│   │   └── identity.ts         # Auth API (login, register, refreshToken)
│   ├── utils/                  # Local utilities
│   │   └── request.ts          # Axios instance + lossless-json + interceptors
│   ├── static/                 # Static assets (logo, icons, images)
│   │   └── logo.png            # Placeholder app logo
│   └── uni_modules/            # uni-ui components (auto-installed)
└── .gitignore                  # Extended gitignore for Uni-app
```

---

## Pre-flight Checks

Before any code, verify the environment:

- [ ] **Check 1: Node.js version**
  ```bash
  node -v
  ```
  Expected: v18+ (Uni-app Vue 3 requires Node >= 18)

- [ ] **Check 2: Existing mobile directory**
  ```bash
  ls -la web/mobile/
  ```
  Expected: only `.gitignore` present — directory is safe to populate

- [ ] **Check 3: web/common/ is accessible**
  ```bash
  ls web/common/types/identity.ts web/common/constants/config.ts web/common/utils/auth.ts
  ```
  Expected: all three files exist

---

### Task 1: Initialize Uni-app Vue 3 project

**Files:**
- Create: `web/mobile/` (all scaffolded files)
- Modify: `web/mobile/.gitignore`

- [ ] **Step 1: Scaffold Uni-app Vue 3 + TypeScript project**

  ```bash
  cd /home/jiaoxh/my-project/community-home/web/mobile && rm -f .gitignore && npx create-uni@latest --template default-ts .
  ```

  When prompted, select "Y" to overwrite existing directory.
  This generates the standard Uni-app Vite + Vue 3 + TypeScript template including:
  - `package.json`, `vite.config.ts`, `tsconfig.json`
  - `pages.json`, `manifest.json`, `uni.scss`
  - `src/main.ts`, `src/App.vue`, `src/pages/index/index.vue`
  - `index.html`

- [ ] **Step 2: Verify scaffold produced key files**

  ```bash
  test -f web/mobile/package.json && test -f web/mobile/vite.config.ts && test -f web/mobile/pages.json && test -f web/mobile/src/main.ts && echo "ALL_KEY_FILES_PRESENT" || echo "MISSING_FILES"
  ```
  Expected: `ALL_KEY_FILES_PRESENT`

- [ ] **Step 3: Restore and extend .gitignore**

  Write `web/mobile/.gitignore`:
  ```
  # Build outputs
  dist/
  .uni/
  unpackage/

  # Dependencies
  node_modules/

  # IDE
  .vscode/
  .idea/

  # OS
  .DS_Store
  Thumbs.db

  # Env
  .env.local
  .env.*.local

  # Logs
  *.log
  npm-debug.log*

  # Zone.Identifier (Windows WSL)
  *:Zone.Identifier

  # Uni-app generated
  src/uni_modules/
  src/manifest.json
  ```

  Run: `git -C /home/jiaoxh/my-project/community-home add web/mobile/.gitignore`

- [ ] **Step 4: Install base dependencies**

  ```bash
  cd /home/jiaoxh/my-project/community-home/web/mobile && npm install
  ```
  Expected: installs successfully, no errors

---

### Task 2: Configure TypeScript with path aliases

**Files:**
- Create: `web/mobile/tsconfig.app.json`
- Modify: `web/mobile/tsconfig.json`

- [ ] **Step 1: Write tsconfig.json with project references**

  Read current `web/mobile/tsconfig.json` to check what scaffolding produced, then replace with:

  `web/mobile/tsconfig.json`:
  ```json
  {
    "files": [],
    "references": [
      { "path": "./tsconfig.app.json" }
    ]
  }
  ```

- [ ] **Step 2: Write tsconfig.app.json**

  `web/mobile/tsconfig.app.json`:
  ```json
  {
    "compilerOptions": {
      "target": "ES2020",
      "module": "ESNext",
      "moduleResolution": "bundler",
      "strict": true,
      "noUnusedLocals": true,
      "noUnusedParameters": true,
      "noFallthroughCasesInSwitch": true,
      "jsx": "preserve",
      "baseUrl": ".",
      "paths": {
        "@/*": ["./src/*"],
        "@common/*": ["../common/*"]
      },
      "types": ["@dcloudio/types"]
    },
    "include": ["src/**/*.ts", "src/**/*.tsx", "src/**/*.vue", "src/**/*.d.ts"]
  }
  ```

- [ ] **Step 3: Verify TypeScript resolves aliases**

  ```bash
  cd /home/jiaoxh/my-project/community-home/web/mobile && npx tsc --noEmit --project tsconfig.app.json 2>&1 | head -20
  ```
  Expected: may have some pre-existing template errors (pages.json not found etc.), but should NOT show "Cannot find module '@common/...'" errors if we haven't imported them yet.

---

### Task 3: Install runtime dependencies

**Files:**
- Modify: `web/mobile/package.json`

- [ ] **Step 1: Install Pinia (state management)**

  ```bash
  cd /home/jiaoxh/my-project/community-home/web/mobile && npm install pinia
  ```
  Expected: installs successfully

- [ ] **Step 2: Install Axios + lossless-json (API client)**

  ```bash
  cd /home/jiaoxh/my-project/community-home/web/mobile && npm install axios lossless-json
  ```
  Expected: installs successfully

- [ ] **Step 3: Install dayjs (date formatting)**

  ```bash
  cd /home/jiaoxh/my-project/community-home/web/mobile && npm install dayjs
  ```
  Expected: installs successfully

- [ ] **Step 4: Install uni-ui (official component library)**

  ```bash
  cd /home/jiaoxh/my-project/community-home/web/mobile && npm install @dcloudio/uni-ui
  ```
  Expected: installs successfully

- [ ] **Step 5: Install SCSS support (if not already)**

  ```bash
  cd /home/jiaoxh/my-project/community-home/web/mobile && npm install -D sass
  ```
  Expected: installs successfully

- [ ] **Step 6: Verify package.json has all deps**

  ```bash
  cd /home/jiaoxh/my-project/community-home/web/mobile && node -e "
  const p = require('./package.json');
  const deps = {...p.dependencies, ...p.devDependencies};
  const needed = ['pinia', 'axios', 'lossless-json', 'dayjs', '@dcloudio/uni-ui', 'sass'];
  const missing = needed.filter(n => !deps[n]);
  if (missing.length) { console.log('MISSING:', missing.join(', ')); process.exit(1); }
  else console.log('ALL_DEPS_PRESENT');
  "
  ```
  Expected: `ALL_DEPS_PRESENT`

---

### Task 4: Configure Vite with uni-app plugin, path aliases, and API proxy

**Files:**
- Modify: `web/mobile/vite.config.ts`

- [ ] **Step 1: Rewrite vite.config.ts**

  Read the scaffolding output first, then replace with:

  `web/mobile/vite.config.ts`:
  ```typescript
  import { defineConfig } from 'vite';
  import uni from '@dcloudio/vite-plugin-uni';
  import { fileURLToPath, URL } from 'node:url';

  // https://vitejs.dev/config/
  export default defineConfig({
    plugins: [
      uni(),
    ],
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url)),
        '@common': fileURLToPath(new URL('../common', import.meta.url)),
      },
    },
    server: {
      port: 3004,
      host: '0.0.0.0',
      proxy: {
        // Dev mode: direct to Go service API ports
        // Production: APISIX :9080 handles routing + JWT
        '/api/auth': {
          target: 'http://127.0.0.1:8881',
          changeOrigin: true,
        },
        '/api/users': {
          target: 'http://127.0.0.1:8882',
          changeOrigin: true,
        },
        '/api/files': {
          target: 'http://127.0.0.1:8884',
          changeOrigin: true,
        },
        '/api/masterdata': {
          target: 'http://127.0.0.1:8889',
          changeOrigin: true,
        },
        '/api/moderation': {
          target: 'http://127.0.0.1:8890',
          changeOrigin: true,
        },
        '/api/v1': {
          target: 'http://127.0.0.1:8891',
          changeOrigin: true,
        },
      },
    },
  });
  ```

- [ ] **Step 2: Verify Vite config parses correctly**

  ```bash
  cd /home/jiaoxh/my-project/community-home/web/mobile && node -e "
  const { createServer } = require('vite');
  console.log('Vite module loaded successfully');
  "
  ```
  Expected: no error. If this fails because vite.config.ts references `@dcloudio/vite-plugin-uni` which requires the build context, that's fine — skip this check. Just verify the file syntax by running:
  ```bash
  cd /home/jiaoxh/my-project/community-home/web/mobile && npx tsc --noEmit vite.config.ts 2>&1 | head -5
  ```

---

### Task 5: Configure Uni-app pages.json (routing + tabBar + globalStyle)

**Files:**
- Modify: `web/mobile/pages.json`
- Modify: `web/mobile/manifest.json`

- [ ] **Step 1: Read and rewrite pages.json**

  Read the scaffolded `web/mobile/pages.json` first to see what template generated, then replace:

  `web/mobile/pages.json`:
  ```json
  {
    "easycom": {
      "autoscan": true,
      "custom": {
        "^uni-(.*)": "@dcloudio/uni-ui/lib/uni-$1/uni-$1.vue"
      }
    },
    "pages": [
      {
        "path": "src/pages/index/index",
        "style": {
          "navigationBarTitleText": "首页"
        }
      },
      {
        "path": "src/pages/discover/discover",
        "style": {
          "navigationBarTitleText": "发现"
        }
      },
      {
        "path": "src/pages/mine/mine",
        "style": {
          "navigationBarTitleText": "我的"
        }
      }
    ],
    "globalStyle": {
      "navigationBarTextStyle": "white",
      "navigationBarTitleText": "社区家园",
      "navigationBarBackgroundColor": "#4A90D9",
      "backgroundColor": "#F5F5F5"
    },
    "tabBar": {
      "color": "#999999",
      "selectedColor": "#4A90D9",
      "backgroundColor": "#FFFFFF",
      "borderStyle": "black",
      "list": [
        {
          "pagePath": "src/pages/index/index",
          "text": "首页",
          "iconPath": "static/tabbar/home.png",
          "selectedIconPath": "static/tabbar/home-active.png"
        },
        {
          "pagePath": "src/pages/discover/discover",
          "text": "发现",
          "iconPath": "static/tabbar/discover.png",
          "selectedIconPath": "static/tabbar/discover-active.png"
        },
        {
          "pagePath": "src/pages/mine/mine",
          "text": "我的",
          "iconPath": "static/tabbar/mine.png",
          "selectedIconPath": "static/tabbar/mine-active.png"
        }
      ]
    }
  }
  ```

- [ ] **Step 2: Configure manifest.json for H5**

  Read scaffolded `web/mobile/manifest.json` first, then update the `h5` section:

  In `web/mobile/manifest.json`, ensure the `h5` key has:
  ```json
  {
    "h5": {
      "title": "社区家园",
      "router": {
        "mode": "hash",
        "base": "/"
      },
      "devServer": {
        "port": 3004
      }
    }
  }
  ```

  Merge with existing manifest content — only replace the `h5` section, keep other platform configs (mp-weixin, etc.) as-is.

- [ ] **Step 3: Create tabBar icon placeholders**

  ```bash
  mkdir -p /home/jiaoxh/my-project/community-home/web/mobile/src/static/tabbar
  ```

  Create 6 placeholder SVG icons (18x18, simple shapes). Write these files:

  `web/mobile/src/static/tabbar/home.png` — 1x1 transparent PNG placeholder:
  ```bash
  # Create minimal valid PNG files (1x1 blue pixel for active, gray for inactive)
  # We'll use a script to generate these
  cd /home/jiaoxh/my-project/community-home/web/mobile/src/static/tabbar

  # Minimal 1x1 PNG files (valid PNG binary)
  # Blue pixel for active icons
  printf '\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR\x00\x00\x00\x01\x00\x00\x00\x01\x08\x02\x00\x00\x00\x90wS\xde\x00\x00\x00\x0cIDATx\x9cc\xf8\x0f\x00\x00\x01\x01\x00\x05\x18\xd8N\x00\x00\x00\x00IEND\xaeB`\x82' > home-active.png
  cp home-active.png discover-active.png
  cp home-active.png mine-active.png

  # Gray pixel for inactive icons
  printf '\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR\x00\x00\x00\x01\x00\x00\x00\x01\x08\x02\x00\x00\x00\x90wS\xde\x00\x00\x00\x0cIDATx\x9cc\xf8\x0f\x00\x00\x01\x01\x00\x05\x18\xd8N\x00\x00\x00\x00IEND\xaeB`\x82' > home.png
  cp home.png discover.png
  cp home.png mine.png
  ```
  Note: These are minimal 1x1 PNGs. Replace with real icons later.

---

### Task 6: Configure uni.scss theme variables

**Files:**
- Modify: `web/mobile/uni.scss`

- [ ] **Step 1: Write uni.scss with theme variables**

  `web/mobile/uni.scss`:
  ```scss
  /* ========================================
   * Uni-app Global Theme Variables
   * Override uni-ui default variables here
   * ======================================== */

  /* --- Brand Colors --- */
  $uni-color-primary: #4A90D9;
  $uni-color-success: #07C160;
  $uni-color-warning: #FF976A;
  $uni-color-error: #EE0A24;
  $uni-color-info: #909399;

  /* --- Text Colors --- */
  $uni-text-color: #323233;
  $uni-text-color-grey: #969799;
  $uni-text-color-placeholder: #C8C9CC;
  $uni-text-color-inverse: #FFFFFF;

  /* --- Background Colors --- */
  $uni-bg-color: #F5F5F5;
  $uni-bg-color-grey: #F7F8FA;
  $uni-bg-color-hover: #F2F3F5;
  $uni-bg-color-mask: rgba(0, 0, 0, 0.4);

  /* --- Border --- */
  $uni-border-color: #EBEDF0;
  $uni-border-radius: 8px;

  /* --- Spacing --- */
  $uni-spacing-sm: 8px;
  $uni-spacing-md: 16px;
  $uni-spacing-lg: 24px;
  $uni-spacing-xl: 32px;

  /* --- Font Sizes --- */
  $uni-font-size-xs: 10px;
  $uni-font-size-sm: 12px;
  $uni-font-size-base: 14px;
  $uni-font-size-md: 16px;
  $uni-font-size-lg: 18px;
  $uni-font-size-xl: 20px;

  /* --- Shadows --- */
  $uni-shadow-sm: 0 1px 2px rgba(0, 0, 0, 0.05);
  $uni-shadow-base: 0 2px 8px rgba(0, 0, 0, 0.08);

  /* --- Custom App Variables --- */
  $app-primary: #4A90D9;
  $app-primary-light: #6BA5E7;
  $app-primary-dark: #3570B0;
  $app-safe-area-bottom: env(safe-area-inset-bottom, 0px);
  ```

---

### Task 7: Write Axios request utility with lossless-json

**Files:**
- Create: `web/mobile/src/utils/request.ts`

- [ ] **Step 1: Write request.ts**

  This mirrors `web/pc/src/utils/request.ts` but adapted for Uni-app (no `ElMessage`, uses `uni.showToast`):

  ```typescript
  // Axios instance with interceptors for Uni-app H5 API requests
  import axios, {
    type AxiosInstance,
    type AxiosError,
    type InternalAxiosRequestConfig,
    type AxiosResponse,
  } from 'axios';
  import { parse, isLosslessNumber, type LosslessNumber } from 'lossless-json';
  import { getAccessToken, getRefreshToken, setTokens, clearTokens } from '@common/utils/auth';
  import { TOKEN_CONFIG } from '@common/constants/config';

  // Custom JSON reviver: keep safe integers as numbers, convert large integers to strings.
  // Snowflake IDs are 19-digit int64 values that exceed JS Number.MAX_SAFE_INTEGER (2^53).
  function reviveLargeNumbers(_key: string, value: unknown): unknown {
    if (isLosslessNumber(value)) {
      const asNumber = Number((value as LosslessNumber).toString());
      if (Number.isSafeInteger(asNumber)) {
        return asNumber;
      }
      return (value as LosslessNumber).toString();
    }
    return value;
  }

  const request: AxiosInstance = axios.create({
    baseURL: import.meta.env.VITE_API_BASE_URL || '',
    timeout: 30000,
    headers: {
      'Content-Type': 'application/json',
    },
    transformResponse: [
      (data: string) => {
        if (typeof data === 'string') {
          try {
            return parse(data, reviveLargeNumbers);
          } catch {
            return data;
          }
        }
        return data;
      },
    ],
  });

  // Token refresh state — prevents concurrent refresh calls
  let isRefreshing = false;
  let failedQueue: Array<{
    resolve: (value?: unknown) => void;
    reject: (reason?: unknown) => void;
  }> = [];

  const processQueue = (error: AxiosError | null, token: string | null = null) => {
    failedQueue.forEach((promise) => {
      if (error) {
        promise.reject(error);
      } else {
        promise.resolve(token);
      }
    });
    failedQueue = [];
  };

  // --- Request Interceptor ---
  request.interceptors.request.use(
    (config: InternalAxiosRequestConfig) => {
      const token = getAccessToken();
      if (token && config.headers) {
        config.headers.Authorization = `Bearer ${token}`;
      }
      return config;
    },
    (error: AxiosError) => {
      return Promise.reject(error);
    },
  );

  // --- Response Interceptor ---
  request.interceptors.response.use(
    (response: AxiosResponse) => {
      const { code, msg, message } = response.data as Record<string, unknown>;

      // Backend uses "msg" (go-zero), frontend uses "message" — support both
      const errorMessage = (msg || message) as string | undefined;

      if (code === 0) {
        return response.data;
      }

      // Business error
      uni.showToast({
        title: errorMessage || '请求失败',
        icon: 'none',
        duration: 2000,
      });
      return Promise.reject(new Error(errorMessage || `Business error code: ${code}`));
    },
    async (error: AxiosError) => {
      const originalRequest = error.config as InternalAxiosRequestConfig & { _retry?: boolean };

      // Handle 401 — attempt token refresh
      if (error.response?.status === 401 && !originalRequest._retry) {
        const refreshToken = getRefreshToken();
        if (!refreshToken) {
          clearTokens();
          uni.reLaunch({ url: '/src/pages/login/login' });
          return Promise.reject(error);
        }

        if (isRefreshing) {
          // Another refresh is in progress — queue this request
          return new Promise((resolve, reject) => {
            failedQueue.push({ resolve, reject });
          })
            .then((token) => {
              if (originalRequest.headers) {
                originalRequest.headers.Authorization = `Bearer ${token}`;
              }
              return request(originalRequest);
            })
            .catch((err) => Promise.reject(err));
        }

        isRefreshing = true;
        originalRequest._retry = true;

        try {
          const response = await axios.post('/api/auth/refresh-token', {
            refreshToken,
          });

          const data = response.data as { data?: { accessToken: string; refreshToken: string; expiresAt: number } };
          if (data.data) {
            const { accessToken, refreshToken: newRefreshToken, expiresAt } = data.data;
            setTokens(accessToken, newRefreshToken, expiresAt);
            processQueue(null, accessToken);
            if (originalRequest.headers) {
              originalRequest.headers.Authorization = `Bearer ${accessToken}`;
            }
            return request(originalRequest);
          }

          throw new Error('Token refresh returned no data');
        } catch (refreshError) {
          processQueue(refreshError as AxiosError, null);
          clearTokens();
          uni.reLaunch({ url: '/src/pages/login/login' });
          return Promise.reject(refreshError);
        } finally {
          isRefreshing = false;
        }
      }

      // Network error
      if (!error.response) {
        uni.showToast({
          title: '网络连接失败，请检查网络',
          icon: 'none',
          duration: 2000,
        });
      }

      return Promise.reject(error);
    },
  );

  export default request;
  export type { AxiosResponse, AxiosError };
  ```

- [ ] **Step 2: Verify the file compiles**

  ```bash
  cd /home/jiaoxh/my-project/community-home/web/mobile && npx tsc --noEmit src/utils/request.ts 2>&1 | head -10
  ```
  Expected: may show "Cannot find module" for `@common/utils/auth` if we haven't configured tsconfig resolution properly — or may pass. The key check is that there are no syntax errors.

---

### Task 8: Write Pinia user store

**Files:**
- Create: `web/mobile/src/stores/user.ts`

- [ ] **Step 1: Write user store**

  `web/mobile/src/stores/user.ts`:
  ```typescript
  import { defineStore } from 'pinia';
  import { ref, computed } from 'vue';
  import {
    getAccessToken,
    getRefreshToken,
    setTokens,
    clearTokens,
    isAuthenticated,
  } from '@common/utils/auth';
  import type { User } from '@common/types/identity';

  export const useUserStore = defineStore('user', () => {
    // --- State ---
    const user = ref<User | null>(null);
    const accessToken = ref<string | null>(getAccessToken());
    const refreshToken = ref<string | null>(getRefreshToken());

    // --- Getters ---
    const isLoggedIn = computed(() => isAuthenticated() && user.value !== null);
    const userId = computed(() => user.value?.id || '');
    const nickname = computed(() => user.value?.nickname || '');
    const avatar = computed(() => user.value?.avatar || '');
    const userType = computed(() => user.value?.userType);

    // --- Actions ---
    function setAuth(loginResponse: { accessToken: string; refreshToken: string; expiresAt: number }) {
      accessToken.value = loginResponse.accessToken;
      refreshToken.value = loginResponse.refreshToken;
      setTokens(
        loginResponse.accessToken,
        loginResponse.refreshToken,
        loginResponse.expiresAt,
      );
    }

    function setUser(newUser: User) {
      user.value = newUser;
    }

    function logout() {
      user.value = null;
      accessToken.value = null;
      refreshToken.value = null;
      clearTokens();
    }

    return {
      // State
      user,
      accessToken,
      refreshToken,
      // Getters
      isLoggedIn,
      userId,
      nickname,
      avatar,
      userType,
      // Actions
      setAuth,
      setUser,
      logout,
    };
  });
  ```

- [ ] **Step 2: Verify the store compiles**

  ```bash
  cd /home/jiaoxh/my-project/community-home/web/mobile && npx tsc --noEmit src/stores/user.ts 2>&1 | head -10
  ```

---

### Task 9: Write Auth API module

**Files:**
- Create: `web/mobile/src/api/identity.ts`

- [ ] **Step 1: Write identity.ts API module**

  ```typescript
  // Identity/Auth Service API — Uni-app mobile
  import request from '@/utils/request';
  import type {
    LoginRequest,
    LoginSmsRequest,
    RegisterRequest,
    LoginResponse,
    RefreshTokenRequest,
    RefreshTokenResponse,
    User,
  } from '@common/types/identity';

  /**
   * Login with phone and password.
   * Note: For mobile, we send plain phone/password initially.
   * RSA encryption will be added later (reuse @/utils/crypto from PC).
   */
  export async function login(phone: string, password: string, deviceId: string): Promise<LoginResponse> {
    const data: LoginRequest = {
      encryptedPhone: phone,   // TODO: RSA encrypt after crypto utility is ported
      encryptedPassword: password,
      deviceId,
      deviceType: 'mobile-h5',
    };

    const res = await request.post<LoginResponse>('/api/auth/login', data);
    return res.data;
  }

  /**
   * Login with phone + SMS code.
   */
  export async function loginWithSms(phone: string, smsCode: string, deviceId: string): Promise<LoginResponse> {
    const data: LoginSmsRequest = {
      encryptedPhone: phone,   // TODO: RSA encrypt
      smsCode,
      deviceId,
      deviceType: 'mobile-h5',
    };

    const res = await request.post<LoginResponse>('/api/auth/login/sms', data);
    return res.data;
  }

  /**
   * Register new user.
   */
  export async function register(params: {
    phone: string;
    password?: string;
    smsCode: string;
    nickname: string;
    deviceId: string;
  }): Promise<LoginResponse> {
    const data: RegisterRequest = {
      encryptedPhone: params.phone,       // TODO: RSA encrypt
      encryptedPassword: params.password,
      smsCode: params.smsCode,
      nickname: params.nickname,
      deviceId: params.deviceId,
      deviceType: 'mobile-h5',
    };

    const res = await request.post<LoginResponse>('/api/auth/register', data);
    return res.data;
  }

  /**
   * Refresh access token.
   */
  export async function refreshToken(refreshToken: string): Promise<RefreshTokenResponse> {
    const data: RefreshTokenRequest = { refreshToken };
    const res = await request.post<RefreshTokenResponse>('/api/auth/refresh-token', data);
    return res.data;
  }

  /**
   * Get current user profile.
   */
  export async function getUserProfile(): Promise<User> {
    const res = await request.get<User>('/api/users/profile');
    return res.data;
  }

  /**
   * Send SMS verification code.
   */
  export async function sendSmsCode(phone: string): Promise<void> {
    await request.post('/api/auth/send-sms', { phone });
  }
  ```

- [ ] **Step 2: Verify compiles**

  ```bash
  cd /home/jiaoxh/my-project/community-home/web/mobile && npx tsc --noEmit src/api/identity.ts 2>&1 | head -10
  ```

---

### Task 10: Write App.vue and main.ts entry

**Files:**
- Modify: `web/mobile/src/main.ts`
- Modify: `web/mobile/src/App.vue`

- [ ] **Step 1: Rewrite src/main.ts**

  Read current `web/mobile/src/main.ts` first, then replace:

  ```typescript
  import { createSSRApp } from 'vue';
  import { createPinia } from 'pinia';
  import App from './App.vue';

  export function createApp() {
    const app = createSSRApp(App);
    const pinia = createPinia();

    app.use(pinia);

    return {
      app,
      pinia,
    };
  }
  ```

- [ ] **Step 2: Rewrite src/App.vue**

  Read current `web/mobile/src/App.vue` first, then replace:

  ```vue
  <script setup lang="ts">
  import { onLaunch, onShow, onHide } from '@dcloudio/uni-app';

  onLaunch(() => {
    console.log('[App] Launch — 社区家园 mobile');
  });

  onShow(() => {
    console.log('[App] Show');
  });

  onHide(() => {
    console.log('[App] Hide');
  });
  </script>

  <style lang="scss">
  /* Global app styles */
  @import '@/uni.scss';

  page {
    font-family: -apple-system, BlinkMacSystemFont, 'Helvetica Neue', Helvetica,
      'Segoe UI', Arial, 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei',
      sans-serif;
    font-size: $uni-font-size-base;
    color: $uni-text-color;
    background-color: $uni-bg-color;
    -webkit-font-smoothing: antialiased;
  }
  </style>
  ```

- [ ] **Step 3: Verify app compiles**

  ```bash
  cd /home/jiaoxh/my-project/community-home/web/mobile && npx tsc --noEmit src/main.ts src/App.vue 2>&1 | head -10
  ```

---

### Task 11: Write placeholder page components

**Files:**
- Create: `web/mobile/src/pages/discover/discover.vue`
- Create: `web/mobile/src/pages/mine/mine.vue`
- Modify: `web/mobile/src/pages/index/index.vue`

- [ ] **Step 1: Write home page (index/index.vue)**

  Read existing `web/mobile/src/pages/index/index.vue` first, then replace:

  ```vue
  <template>
    <view class="page">
      <view class="header">
        <text class="title">社区家园</text>
        <text class="subtitle">欢迎回来</text>
      </view>
      <view class="content">
        <text class="placeholder">首页内容建设中...</text>
      </view>
    </view>
  </template>

  <script setup lang="ts">
  import { useUserStore } from '@/stores/user';

  const userStore = useUserStore();
  </script>

  <style scoped lang="scss">
  .page {
    min-height: 100vh;
    padding: 32px 24px;
  }

  .header {
    margin-bottom: 32px;

    .title {
      display: block;
      font-size: 48rpx;
      font-weight: 700;
      color: $uni-text-color;
      margin-bottom: 8px;
    }

    .subtitle {
      display: block;
      font-size: 28rpx;
      color: $uni-text-color-grey;
    }
  }

  .content {
    display: flex;
    justify-content: center;
    align-items: center;
    padding: 120rpx 0;

    .placeholder {
      font-size: 28rpx;
      color: $uni-text-color-placeholder;
    }
  }
  </style>
  ```

- [ ] **Step 2: Write discover page**

  `web/mobile/src/pages/discover/discover.vue`:
  ```vue
  <template>
    <view class="page">
      <view class="header">
        <text class="title">发现</text>
      </view>
      <view class="content">
        <text class="placeholder">发现内容建设中...</text>
      </view>
    </view>
  </template>

  <script setup lang="ts">
  </script>

  <style scoped lang="scss">
  .page {
    min-height: 100vh;
    padding: 32px 24px;
  }

  .header {
    margin-bottom: 32px;

    .title {
      display: block;
      font-size: 48rpx;
      font-weight: 700;
      color: $uni-text-color;
    }
  }

  .content {
    display: flex;
    justify-content: center;
    align-items: center;
    padding: 120rpx 0;

    .placeholder {
      font-size: 28rpx;
      color: $uni-text-color-placeholder;
    }
  }
  </style>
  ```

- [ ] **Step 3: Write mine page**

  `web/mobile/src/pages/mine/mine.vue`:
  ```vue
  <template>
    <view class="page">
      <!-- Not logged in -->
      <view v-if="!userStore.isLoggedIn" class="login-prompt">
        <view class="avatar-placeholder">
          <text class="avatar-icon">👤</text>
        </view>
        <text class="login-text">点击登录</text>
      </view>

      <!-- Logged in -->
      <view v-else class="profile">
        <view class="avatar">
          <image
            v-if="userStore.avatar"
            :src="userStore.avatar"
            class="avatar-img"
            mode="aspectFill"
          />
          <text v-else class="avatar-icon">👤</text>
        </view>
        <text class="nickname">{{ userStore.nickname }}</text>
      </view>
    </view>
  </template>

  <script setup lang="ts">
  import { useUserStore } from '@/stores/user';

  const userStore = useUserStore();
  </script>

  <style scoped lang="scss">
  .page {
    min-height: 100vh;
    padding: 32px 24px;
  }

  .login-prompt {
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 120rpx 0;

    .avatar-placeholder {
      width: 120rpx;
      height: 120rpx;
      border-radius: 50%;
      background-color: $uni-bg-color-grey;
      display: flex;
      align-items: center;
      justify-content: center;
      margin-bottom: 20rpx;

      .avatar-icon {
        font-size: 60rpx;
      }
    }

    .login-text {
      font-size: 32rpx;
      color: $uni-color-primary;
    }
  }

  .profile {
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 60rpx 0;

    .avatar {
      width: 120rpx;
      height: 120rpx;
      border-radius: 50%;
      background-color: $uni-bg-color-grey;
      display: flex;
      align-items: center;
      justify-content: center;
      margin-bottom: 20rpx;
      overflow: hidden;

      .avatar-img {
        width: 100%;
        height: 100%;
      }

      .avatar-icon {
        font-size: 60rpx;
      }
    }

    .nickname {
      font-size: 36rpx;
      font-weight: 600;
      color: $uni-text-color;
    }
  }
  </style>
  ```

- [ ] **Step 4: Verify pages compile**

  ```bash
  cd /home/jiaoxh/my-project/community-home/web/mobile && npx tsc --noEmit src/pages/index/index.vue src/pages/discover/discover.vue src/pages/mine/mine.vue 2>&1 | head -20
  ```

---

### Task 12: Create env.d.ts and .env.development

**Files:**
- Create: `web/mobile/src/env.d.ts`
- Create: `web/mobile/.env.development`

- [ ] **Step 1: Write env.d.ts**

  `web/mobile/src/env.d.ts`:
  ```typescript
  /// <reference types="vite/client" />
  /// <reference types="@dcloudio/types" />

  declare module '*.vue' {
    import type { DefineComponent } from 'vue';
    const component: DefineComponent<Record<string, never>, Record<string, never>, unknown>;
    export default component;
  }

  interface ImportMetaEnv {
    readonly VITE_API_BASE_URL: string;
    readonly VITE_APP_TITLE: string;
  }

  interface ImportMeta {
    readonly env: ImportMetaEnv;
  }
  ```

- [ ] **Step 2: Write .env.development**

  `web/mobile/.env.development`:
  ```env
  # Uni-app H5 Dev Server
  VITE_API_BASE_URL=
  VITE_APP_TITLE=社区家园
  ```

---

### Task 13: Final build verification

**Files:**
- Verify all above

- [ ] **Step 1: Install all dependencies fresh**

  ```bash
  cd /home/jiaoxh/my-project/community-home/web/mobile && npm install
  ```
  Expected: no errors

- [ ] **Step 2: Run TypeScript type check**

  ```bash
  cd /home/jiaoxh/my-project/community-home/web/mobile && npx tsc --noEmit -p tsconfig.app.json 2>&1 | head -30
  ```
  Expected: errors may exist from scaffolding (Uni-app's `.vue` imports, global `uni` type). The critical check is that there are NO errors about:
  - `Cannot find module '@common/...'`
  - `Cannot find module '@/utils/request'`
  - `Cannot find module '@/stores/user'`

- [ ] **Step 3: Try H5 dev build**

  ```bash
  cd /home/jiaoxh/my-project/community-home/web/mobile && npx vite build --mode development 2>&1 | tail -20
  ```
  Expected: build succeeds with output in `dist/`

- [ ] **Step 4: Try dev server (quick smoke test)**

  ```bash
  cd /home/jiaoxh/my-project/community-home/web/mobile && timeout 5 npx vite --port 3004 2>&1 || true
  ```
  Expected: server starts, no crash. May show "ready in X ms" or similar.

---

## Self-Review Checklist

### 1. Spec Coverage
- [x] Uni-app Vue 3 framework initialized — Task 1
- [x] TypeScript with `@common` alias — Task 2
- [x] Pinia, Axios, lossless-json, dayjs, uni-ui, SCSS — Task 3
- [x] Vite config with uni-app plugin + proxy — Task 4
- [x] pages.json routing + tabBar config — Task 5
- [x] Theme variables (uni.scss) — Task 6
- [x] Request utility with Snowflake ID handling — Task 7
- [x] User auth store — Task 8
- [x] API module (identity) — Task 9
- [x] App entry + lifecycle — Task 10
- [x] Three tab pages (index, discover, mine) — Task 11
- [x] Type declarations + env — Task 12
- [x] Build verification — Task 13

### 2. Placeholder Scan
- No "TBD" or "TODO implement later" in code (only intentional TODO comments for RSA encryption which is a known future task)
- All imports reference actual files
- All paths are exact
- All code blocks are concrete

### 3. Type Consistency
- `web/common/types/identity.ts` exports `User`, `LoginRequest`, `LoginResponse`, etc. — all referenced correctly in store and API
- `web/common/utils/auth.ts` exports `getAccessToken`, `getRefreshToken`, `setTokens`, `clearTokens`, `isAuthenticated` — used consistently in store and request
- `web/common/constants/config.ts` exports `TOKEN_CONFIG` — referenced in request.ts
- `@common/*` alias defined in tsconfig.app.json and vite.config.ts
- `@/*` alias defined identically in tsconfig.app.json and vite.config.ts
- Pinia store function names match between `stores/user.ts` definition and page component usage (`isLoggedIn`, `nickname`, `avatar`)

---

## Execution Options

**Plan complete and saved.** Two execution options:

1. **Subagent-Driven (recommended)** — I dispatch a fresh subagent per task, review between tasks, fast iteration

2. **Inline Execution** — Execute tasks in this session, batch execution with checkpoints

Which approach?
