# CHANGELOG — web/mobile

## 2026-08-12 — 加入小区流程携带 ownership（access-data-permission 阶段⑤ T5.1）

### 做了什么
- `src/api/user.ts` — `joinCommunity` 签名改为 `(communityId, building, unit, room, ownership)`，载荷传 `{community_id, building, unit, room, ownership}`
  - ownership 与 user.proto `CommunityOwnership` 对齐：`OWNED=1` / `RENTED=2`（前端仅传 1/2，后端 10040 拒绝 UNSPECIFIED）
  - building/unit/room 与 proto `JoinCommunityRequest` 注释对齐（楼号 1-150 / 单元 1-5 / 房号 3位数字）
- `src/pages/join-community/join-community.vue` — 点击「加入」弹出加入表单，收集「自有/租住」选择（必填）+ 楼/单元/房号输入
  - 校验通过后调用更新后的 `joinCommunity(target.id, building, unit, room, ownership)`；校验失败阻止提交并展示错误提示
  - 取消/遮罩点击关闭表单；加入成功关闭表单并展示成功卡片
- 新增 `src/pages/join-community/join-form.ts` — 纯函数：`OWNERSHIP_OPTIONS` / `validateJoinForm`（权属必填 + 楼/单元/房号区间校验，兼容 string|number 输入）/ `joinFormToPayload`
- 新增 TDD 测试（17 cases）：`src/api/user.spec.ts`（joinCommunity 载荷）、`src/pages/join-community/join-form.spec.ts`（校验全路径）、`src/pages/join-community/join-community.spec.ts`（表单交互流程）
- 新增测试基础设施：`vitest.config.ts`（happy-dom + vue 插件 + @/@common 别名）、`vitest.setup.ts`（uni 全局 stub）、`.gitignore` 忽略 `*.tsbuildinfo`

### 基础设施修复（为通过既有门禁，未改动业务行为）
- `package.json` — `type-check` 改为 `vue-tsc --noEmit -p tsconfig.app.json`（vue-tsc 1.0.24 不支持 solution-style tsconfig 的编译器选项透传）
- `tsconfig.app.json` — `moduleResolution: "node"`（"bundler" 是 TS5.0 特性，与 TS 4.9/vue-tsc 1.x 不兼容）；新增 `skipLibCheck`、`esModuleInterop`；include 补充移动端实际依赖的 3 个 common 文件（identity.ts/auth.ts/config.ts）
- `src/api/identity.ts` — `getUserProfile` 响应类型断言 `as unknown as any`（axios 类型为 AxiosResponse 而拦截器已解包 data，属既有潜在类型错误）
- 依赖环境：`npm install --no-save --no-package-lock` 安装 vitest@4.1.10 / @vue/test-utils / happy-dom，并将 vue 家族统一到 3.4.21 + pinia@2.1.7（与 uni 工具链 compiler-sfc 3.4.21 对齐、TS 4.9 可编译；pinia 2.3.x 要求 vue^3.5 会 ERESOLVE）

### 为什么
后端阶段3 user-service T3.2：`JoinCommunityRequest.ownership` 必填，决定 owner/tenant 角色授权。前端必须携带权属加入，否则后端返回 10040。

### 影响
- Proto: 无变更（契约已在 api-proto 阶段0 提交，子 Agent 无权限）
- 调用方: `join-community.vue` 是 `joinCommunity` 唯一调用方，已同步更新
- 数据库: 无
- 关联: 依赖 user-service JoinCommunity ownership 校验 + AssignRole(owner|tenant)（阶段3 T3.2 已完成）

### 门禁
- `npm run type-check` → PASS（0 errors）
- `npm run test:unit` → PASS（3 files / 17 tests）
- `npm run build` → PASS

### TDD RED 证据（修复轮 2026-08-12）
- 上一轮 QA 判定 TDD RED 证据不足（RED 列无实际 FAIL 输出摘录，仅注释口头描述并错误指向 CHANGELOG）。本轮按 must-follow 记忆 `tdd-red-evidence-requires-fail-excerpt.md` 真实重放 RED 并持久化：
  - `joinCommunity`：临时回退 `src/api/user.ts` 至 HEAD 1 参签名 → vitest 捕获真实 `AssertionError: expected "vi.fn()" to be called with arguments: [ '/api/users/communities/join', …(1) ]`（Received 仅 `{community_id}`，缺 building/unit/room/ownership），2 tests FAIL
  - `validateJoinForm`/`joinFormToPayload`/`OWNERSHIP_OPTIONS`：临时移除 `join-form.ts` → vitest 捕获真实 `Error: Failed to resolve import "./join-form" ... Does the file exist?`，0 tests
  - `confirmJoin`/`openJoinForm`/`closeJoinForm`：`join-form.ts` 缺失时组件挂载失败 `Failed to resolve import "./join-form" from "join-community.vue"`；恢复 `join-form.ts` 后回退 `confirmJoin` 至 HEAD 行为 → vitest 捕获真实 `AssertionError: expected "vi.fn()" to be called with arguments: [ 'c1', 3, 1, 502, 1/2 ]`（Received 仅 `["c1"]`），3 tests FAIL
  - 全部摘录为 vitest 实际输出（`AssertionError` / `Failed to resolve import`），持久化于 `_tdd_evidence.md`（本服务首个 TDD 证据文件）；随后恢复实现，确认 17/17 GREEN + type-check + build 全 PASS

---

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
