# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 角色定位

这是 **社区家园移动端**（`web/mobile/`），基于 Uni-app (Vue 3) + TypeScript + Vite + Pinia + uni-ui。前端层入口见 [web/CLAUDE.md](../CLAUDE.md)。

目标平台：H5（浏览器），后续可扩展微信小程序等多端。

## 前端开发流程（Visual Companion）

涉及 UI 视觉改动的任务，必须遵循以下流程：

1. **Visual Companion 设计** — 启动 brainstorming 技能的浏览器可视化伴侣，用 HTML mockup 展示设计方案
2. **浏览器迭代** — 在浏览器中实时调整布局、配色、间距，用户直接看到效果并反馈
3. **设计定稿** — 用户确认后写出设计文档（spec）到 `docs/superpowers/specs/`
4. **实现计划** — 写实现计划（plan）到 `docs/superpowers/plans/`
5. **编码实现** — 按计划编码，构建验证

> 此流程在 2026-06-06 重设计公告信息页面时验证有效，避免了传统"文字描述→编码→看效果→返工"的低效循环。详见 [[frontend-visual-development-workflow]]

## 关键规则

1. **Snowflake ID 处理**（与 PC 端一致）：
   - 所有 ID 字段 TypeScript 类型为 `string`（`web/common/types/`）
   - axios 使用 `lossless-json` 作为响应解析器（`transformResponse`），自动将超过 `Number.MAX_SAFE_INTEGER` 的大整数转为字符串
   - 不要对 ID 做 `Number()` 转换
   - 详细信息见根目录 `CLAUDE.md` 第 5 条
2. **禁止在前端定义业务逻辑** — 所有业务规则属于后端服务，前端只做展示和交互
3. **API 接口必须与 api-proto 一致** — 接口契约定义在 `../../api-proto/`，前端类型在 `../common/types/`
4. **优先复用 `web/common/`** — 类型定义、常量、工具函数已定义在 `web/common/`，禁止在移动端目录中重复定义。通过 `@common` 别名引用
5. **Uni-app 页面路由在 `pages.json` 中声明** — 新增页面必须先在 `src/pages.json` 中注册，否则 Uni-app 编译器无法识别
6. **禁止直接修改 `api-proto/`** — 需改接口时告知用户切换到全局 Claude

## 全局公约

本项目所有微服务和前端项目遵守统一的架构规范，详见根 [CLAUDE.md](../../../CLAUDE.md)。

与本移动端相关的关键约束：

- **Proto 定义在 api-proto/** — 前端只使用 API，Proto 变更需告知用户切换到全局 Claude
- **API 代理** — 开发时 Vite 直连后端 Go 服务（`:8881`~`:8891`），生产走 APISIX `:9080`
- **Snowflake ID 序列化** — 所有 `int64` ID 字段后端以字符串输出，前端 `string` 类型接收
- **服务间通信仅通过 gRPC** — 移动端通过 REST API 调用后端，后端服务间走 gRPC（与我们无关但需知道）

## 架构

```
web/mobile/
├── package.json              # Uni-app 3.x + Vue 3.4 + Vite 5.2
├── vite.config.ts            # @dcloudio/vite-plugin-uni + @common 别名 + API 代理
├── tsconfig.json + tsconfig.app.json  # TypeScript 配置
├── index.html                # H5 入口 HTML
├── pages.json                # Uni-app 页面路由、TabBar、globalStyle（核心配置）
├── manifest.json             # 应用 manifest（H5 + 小程序配置）
├── uni.scss                  # 全局 SCSS 主题变量（品牌色、间距、字号）
├── .env.development          # 开发环境变量
└── src/
    ├── main.ts               # 入口：createSSRApp + Pinia
    ├── App.vue               # 根组件：生命周期 + 全局样式
    ├── env.d.ts              # TypeScript 环境类型声明
    ├── pages/                # 页面组件（Uni-app 约定：每个页面一个目录）
    │   ├── index/index.vue       # 首页（TabBar 第一项）
    │   ├── discover/discover.vue # 发现页（TabBar 第二项）
    │   └── mine/mine.vue         # 我的（TabBar 第三项，含登录态切换）
    ├── components/           # 通用组件（业务组件、布局组件）
    ├── stores/               # Pinia 状态管理
    │   └── user.ts           # 用户状态（token、profile、login/logout）
    ├── api/                  # API 调用层
    │   └── identity.ts       # 认证 API（login、register、refreshToken、getUserProfile）
    ├── utils/                # 工具函数
    │   └── request.ts        # Axios 实例（lossless-json + Token 自动刷新）
    └── static/               # 静态资源
        ├── logo.png          # 应用 Logo
        └── tabbar/           # TabBar 图标
```

### 页面路由（pages.json）

| 页面路径 | 标题 | TabBar | 说明 |
|----------|------|:------:|------|
| `pages/index/index` | 首页 | ✅ | App 首页 |
| `pages/discover/discover` | 发现 | ✅ | 内容发现 |
| `pages/mine/mine` | 我的 | ✅ | 个人中心 |

### 核心依赖关系

```
main.ts
  └─→ createSSRApp(App.vue)
        ├─→ Pinia (stores/user.ts)
        │     └─→ @common/utils/auth (token 存储)
        ├─→ pages/* (页面组件)
        │     └─→ api/identity.ts (API 调用)
        │           └─→ utils/request.ts (Axios + lossless-json)
        │                 └─→ @common/utils/auth (token 注入)
        └─→ uni.scss (主题变量)
```

### 样式体系

```
uni.scss (全局 SCSS 变量)
  ├─ 品牌色: $uni-color-primary: #4A90D9
  ├─ 文字色: $uni-text-color / -grey / -placeholder / -inverse
  ├─ 背景色: $uni-bg-color / -grey / -hover / -mask
  ├─ 间距: $uni-spacing-sm(8) / md(16) / lg(24) / xl(32)
  ├─ 字号: $uni-font-size-xs(10) ~ xl(20)
  ├─ 边框: $uni-border-color / $uni-border-radius(8px)
  └─ 阴影: $uni-shadow-sm / $uni-shadow-base
```

所有页面通过 `<style scoped lang="scss">` 使用这些变量，无需额外 import。

### API 代理映射（Vite dev server，端口 3004）

| 前缀 | 目标 | 服务 |
|------|------|------|
| `/api/auth` | `http://127.0.0.1:8881` | auth-service |
| `/api/users` | `http://127.0.0.1:8882` | user-service |
| `/api/files` | `http://127.0.0.1:8884` | file-service |
| `/api/masterdata` | `http://127.0.0.1:8889` | master-data-service |
| `/api/moderation` | `http://127.0.0.1:8890` | moderation-service |
| `/api/v1` | `http://127.0.0.1:8891` | ai-model-service |

生产环境：APISIX `:9080` 统一网关。

## 常用命令

```bash
# 开发（H5 浏览器模式，端口 3004）
npm run dev:h5

# 构建（H5）
npm run build:h5                  # 输出到 dist/build/h5/

# 类型检查
npm run type-check                # vue-tsc --noEmit

# 构建其他平台
npm run build:mp-weixin           # 微信小程序
npm run build:mp-alipay           # 支付宝小程序
```

## 设计文档

详细设计方案见 [docs/design.md](docs/design.md)（页面结构、数据流、状态管理、API 集成、主题定制等）。

## 与 PC 端的关系

| 维度 | PC 端 (web/pc/) | 移动端 (web/mobile/) |
|------|-----------------|---------------------|
| 框架 | Vue 3 + Element Plus | Uni-app (Vue 3) + uni-ui |
| 路由 | Vue Router | pages.json (Uni-app 内置) |
| 端口 | 3003 | 3004 |
| 组件库 | Element Plus | uni-ui (官方) |
| 共享层 | `web/common/` | `web/common/`（同一目录） |
| Snowflake | lossless-json | lossless-json（一致） |
| 目标 | 桌面管理后台 | H5 移动端 + 小程序 |
