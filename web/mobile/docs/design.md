# 移动端设计文档 (web/mobile/)

## 设计概述

社区家园移动端，基于 Uni-app (Vue 3) 框架，支持 H5 浏览器访问，后续可扩展至微信小程序等多端。

### 技术选型理由

| 决策 | 选择 | 原因 |
|------|------|------|
| 框架 | Uni-app (Vue 3) | 跨端能力（H5→小程序），与 PC 端同为 Vue 3 生态 |
| 构建 | Vite 5.2 | Uni-app 官方 Vite 插件，HMR 极快 |
| UI 组件 | uni-ui | DCloud 官方，跨平台兼容性最好，easycom 自动导入 |
| 状态管理 | Pinia 2.1.7 | 与 PC 端一致，Composition API 风格 |
| 类型系统 | TypeScript | 与 PC 端和 `web/common/` 保持类型安全 |
| JSON 解析 | lossless-json | Snowflake ID (19位 int64 → string) 精度保护 |
| 样式 | SCSS | Uni-app 原生支持，全局变量体系 |

## 页面结构

### Tab 页面（底部导航栏）

```
┌──────────┐  ┌──────────┐  ┌──────────┐
│   首页    │  │   发现    │  │   我的    │
│  (index)  │  │(discover) │  │  (mine)   │
└──────────┘  └──────────┘  └──────────┘
```

### 子页面（需注册到 pages.json）

后续开发中按需添加，例如：

| 页面路径 | 标题 | TabBar |
|----------|------|:------:|
| `pages/login/login` | 登录 | ❌ |
| `pages/agreement/agreement` | 协议注册 | ❌ |
| `pages/join-community/join-community` | 加入小区 | ❌ |
| `pages/notice/notice` | 通知首页 | ✅ |
| `pages/notice-detail/notice-detail` | 通知详情 | ❌ |
| `pages/notice-browse/notice-browse` | 通知浏览（30 天窗口） | ❌ |

新增页面只需：
1. 在 `src/pages/` 下创建 `页面名/页面名.vue`
2. 在 `src/pages.json` 的 `pages` 数组中注册路径

### 登录 / 注册流程（现行实现）

> 2026-08 信息架构改造后的实际流程。早期设计（登录页内嵌协议勾选 + 独立 `pages/register/register` 注册页）已废弃，现改为「短信登录 + 未注册一键注册」：

```
登录页 pages/login/login
  ├─ 无协议勾选区：canSubmit 仅依赖「手机号合法 + 验证码合法」
  ├─ 提交 → loginWithSms(phone, smsCode, deviceId)
  │    ├─ 成功 → handleAuthSuccess(loginRes, { onCompleted })
  │    │         （保存 token + 拉 profile + 小区判断跳转；
  │    │          submitting 由 onCompleted 在跳转完成后复位，防双提交）
  │    └─ 失败 err.code === 50001（未注册；code 缺失时回退 msg 含 50001/未注册）
  │         → saveRegPending({phone, smsCode, deviceId, nickname}) → navigateTo 协议页
  └─ 其他错误 → 复位 submitting 并停止（拦截器已 toast）
         ↓
协议页 pages/agreement/agreement
  ├─ 展示《社区家园使用协议》正文 + 「已阅并同意」勾选 + 确认注册按钮
  ├─ 确认注册 → register({phone, smsCode, nickname, deviceId})
  │    ├─ 成功 → clearRegPending() + handleAuthSuccess(regRes, { onCompleted })（自动登录一次）
  │    └─ 失败 → toast 并保留临时数据可重试
  └─ onLoad 读不到临时数据 → 提示「注册信息已失效」并返回
```

**reg-pending 跨页一次性注册数据（`src/utils/reg-pending.ts`）**：
- 唯一契约源：`REG_PENDING_KEY` 与 `RegPending` 结构收敛到共享模块，两端（login.vue / agreement.vue）import 引用，禁止各自内联 magic string
- 载体：**模块级内存变量为主**（页面栈内导航直接传递，不落持久化）；仅 H5 环境镜像到 `window.sessionStorage`（`{data, expiresAt}`，**TTL 5 分钟**，过期即清）
- **绝不写入 localStorage** — 一次性 smsCode 持久化残留会被共享设备复用，属安全风险
- sessionStorage 访问一律 try/catch 容错（隐私模式 / 被禁用时降级内存态）

## 数据流

```
  User Action
      │
      ▼
  Page Component (pages/*.vue)
      │
      ├─→ Pinia Store (stores/*.ts) ──→ 缓存/响应式计算
      │
      └─→ API Layer (api/*.ts)
            │
            └─→ Axios Instance (utils/request.ts)
                  │
                  ├─ lossless-json parse (Snowflake → string)
                  ├─ Authorization header (Bearer token)
                  └─ 401 → Token refresh interceptor
                        │
                        ▼
                  Go Backend (via Vite proxy / APISIX)
```

### 状态管理设计

```
stores/
  user.ts           # 用户认证状态
    - user: User | null
    - accessToken: string | null
    - refreshToken: string | null
    + isLoggedIn (computed)
    + userId, nickname, avatar (computed)
    + setAuth(loginResponse)
    + setUser(user)
    + logout()
```

后续扩展：
- `stores/community.ts` — 社区内容（帖子、评论）
- `stores/app.ts` — 应用全局状态（网络状态、主题等）

### API 层设计

```
api/
  identity.ts       # 认证 API
    login(phone, password, deviceId) → LoginResponse
    loginWithSms(phone, smsCode, deviceId) → LoginResponse
    register(params) → LoginResponse
    refreshToken(token) → RefreshTokenResponse
    getUserProfile() → User
    sendSmsCode(phone) → void
```

所有 API 方法返回 Promise，由 Pinia action 调用并管理状态。

## 主题系统

### 设计方向

> **白色为主，淡咖色点缀，温暖、简单、松弛**

整体视觉追求干净、温馨、不拥挤。大面积留白，咖色仅用于关键交互元素（按钮、选中态、链接），让页面像一杯拿铁一样舒服。

### 色彩方案

| 变量 | 值 | 用途 |
|------|------|------|
| `$uni-color-primary` | `#B8956A` | 温暖咖色 — 按钮、选中态、链接、强调 |
| `$uni-color-primary-light` | `#D4B896` | 浅咖色 — hover、背景点缀 |
| `$uni-color-primary-dark` | `#8B6F4E` | 深咖色 — 按下态 |
| `$uni-color-success` | `#8DAF7E` | 柔和绿 — 成功/确认 |
| `$uni-color-warning` | `#E8C98E` | 柔和琥珀 — 警告 |
| `$uni-color-error` | `#D4958A` | 柔和红 — 错误/删除 |
| `$uni-color-info` | `#A6988A` | 暖灰 — 提示信息 |
| `$uni-text-color` | `#3D3226` | 暖深棕 — 主要文字 |
| `$uni-text-color-grey` | `#A6988A` | 暖灰 — 辅助文字 |
| `$uni-text-color-placeholder` | `#CCC4BA` | 浅暖灰 — 占位符 |
| `$uni-bg-color` | `#FFFFFF` | 纯白 — 页面背景 |
| `$uni-bg-color-card` | `#FAF8F5` | 暖白 — 卡片背景 |
| `$uni-bg-color-input` | `#F5F0EA` | 暖灰白 — 输入框背景 |
| `$uni-border-color` | `#E8E0D5` | 暖色 — 边框/分割线 |

### 视觉原则

- **大面积留白**：页面背景纯白，卡片和区块之间用阴影或暖色分割线区分，不用大色块
- **咖色克制**：仅用于按钮、选中态 Tab、链接文字，不铺背景
- **圆角柔和**：卡片 12rpx、按钮 48rpx（胶囊形）、输入框 12rpx
- **阴影轻盈**：`$uni-shadow-sm` 和 `$uni-shadow-base` 用于卡片浮起，不厚重
- **字号宽松**：正文 28rpx，标题 32-36rpx，行高 1.6，阅读不压迫

### 间距体系

基于 8px 基准，用于页面和组件内部间距：

| 变量 | 值 | 场景 |
|------|------|------|
| `$uni-spacing-sm` | 8px | 紧凑间距（图标与文字、标签间隙） |
| `$uni-spacing-md` | 16px | 标准间距（卡片内边距、列表项间隔） |
| `$uni-spacing-lg` | 24px | 宽松间距（板块之间） |
| `$uni-spacing-xl` | 32px | 大间距（页面水平边距） |

## 安全机制

### Token 处理

```
Token 存储：
  - Access Token → sessionStorage（关闭标签页即清除）
  - Refresh Token → localStorage（持久化，用于自动登录）

Token 刷新流程：
  请求 → 401 → 检查 refreshToken
    ├─ 有 RT → 调用 /api/auth/refresh-token
    │   ├─ 成功 → 更新 AT + RT → 重试原请求
    │   └─ 失败 → 清除 Token → 跳转登录页
    └─ 无 RT → 清除 Token → 跳转登录页

并发刷新保护：
  多个请求同时 401 → 只有第一个触发刷新 → 其他排队等待
```

### RSA 加密（待实现）

当前手机号/密码以明文传输。后续需从 PC 端迁移 RSA 加密工具：
- `web/pc/src/utils/crypto.ts` → `web/mobile/src/utils/crypto.ts`
- `web/pc/src/utils/device.ts` → `web/mobile/src/utils/device.ts`

## 与 PC 端的共享

通过 `@common` 别名直接引用 `web/common/` 下的文件：

```
web/common/
  types/identity.ts       → @common/types/identity (User, LoginRequest, ...)
  types/common.d.ts        → @common/types/common (ApiResponse, PageResponse, ...)
  types/masterdata.d.ts    → @common/types/masterdata
  constants/config.ts      → @common/constants/config (API_CONFIG, TOKEN_CONFIG)
  constants/error-codes.ts → @common/constants/error-codes
  constants/enums.ts       → @common/constants/enums (状态标签映射)
  utils/auth.ts            → @common/utils/auth (token 存取)
  utils/format.ts          → @common/utils/format (日期/数字格式化)
```

共享而非重写，保证 PC 端和移动端的类型定义、Token 处理、错误码解释完全一致。
