# CHANGELOG — web/mobile

## 2026-06-06 — 验证码登录/注册页面 + RSA 加密

### 做了什么
- 实现 `src/pages/login/login.vue`：验证码登录/注册一体化页面
  - 手机号 + 短信验证码 + 协议复选框
  - 先尝试 LoginSms → 失败则自动注册（昵称="用户"+手机尾号4位）
  - SMS 按钮最小 30s 冷却（成功后延长至 60s）
- 实现 `src/utils/crypto.ts`：RSA 加密工具（Web Crypto API，RSA-OAEP + SHA-256，与后端 `common/pkg/crypto/rsa.go` 对齐）
- 实现 `src/utils/device.ts`：设备 ID 生成器（localStorage 持久化）
- 更新 `src/api/identity.ts`：集成 RSA 加密，修复 API 路径（`sms/send`、`token/refresh`）
- 更新 `src/pages/mine/mine.vue`：未登录状态点击跳转登录页
- 更新 `src/pages.json`：注册登录页路由

### 为什么
移动端需要验证码登录/注册入口，RSA 加密是后端 API 的硬性要求（手机号必须加密传输）。

### 影响
- Proto: 无变更
- 调用方: 无
- 数据库: 无
- 关联: 依赖 auth-service 的 `/api/auth/sms/send`、`/api/auth/login/sms`、`/api/auth/register`、`/api/auth/public-key` 接口

## 2026-06-06 — 初始化 Uni-app Vue 3 移动端框架

### 做了什么
- 使用 Uni-app (Vue 3) + Vite + TypeScript 初始化项目脚手架
- 配置 TypeScript 路径别名（`@/*`、`@common/*`，引用 `web/common/`）
- 安装核心依赖：Pinia 2.1.7（状态管理）、Axios（HTTP 客户端）、lossless-json（Snowflake ID 解析）、dayjs（日期格式化）、uni-ui（官方组件库）
- 配置 Vite：uni 插件、API 代理（6 个后端服务）、开发端口 3004
- 配置 `pages.json`：三 Tab 路由（首页/发现/我的）+ TabBar + globalStyle
- 配置 `manifest.json`：H5 模式（hash 路由、端口 3004）
- 编写 `uni.scss`：完整的主题变量体系（品牌色、文字色、背景色、间距、字号、阴影）
- 实现 `src/utils/request.ts`：Axios 实例 + lossless-json Snowflake ID 解析 + Token 自动刷新（并发刷新保护）
- 实现 `src/stores/user.ts`：Pinia 用户状态管理（token、profile、login/logout）
- 实现 `src/api/identity.ts`：认证 API 封装（login、loginWithSms、register、refreshToken、getUserProfile、sendSmsCode）
- 实现三个 Tab 占位页面：首页（含用户状态引用）、发现页、我的（登录/未登录状态切换）
- 创建 `CLAUDE.md`（子 Claude 配置）、`docs/design.md`（设计文档）、`CHANGELOG.md`
- 更新父级 `web/CLAUDE.md` 添加 mobile 子实例索引

### 为什么
社区家园需要移动端访问入口，Uni-app (Vue 3) 提供跨端能力（H5+小程序），复用 `web/common/` 共享类型和工具函数保证与 PC 端一致。

### 影响
- Proto: 无变更（使用现有 API）
- 调用方: 无
- 数据库: 无
- 关联: 框架初始化，为后续移动端功能开发提供基础设施
