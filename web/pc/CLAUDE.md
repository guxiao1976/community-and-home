# CLAUDE.md

## 角色定位

这是 **Vue 3 管理后台**（`web/pc/`），Element Plus + TypeScript + Pinia + Vue Router。前端层入口见 [web/CLAUDE.md](../CLAUDE.md)。

## 关键规则

1. **Snowflake ID 处理**：
   - 所有 ID 字段 TypeScript 类型为 `string`（`web/common/types/`）
   - axios 使用 `lossless-json` 作为响应解析器（`transformResponse`）
   - 不再使用 `Number(route.params.id)`，直接使用 `route.params.id as string`
   - 详细信息见根目录 `CLAUDE.md` 第 5 条
2. **禁止在前端定义业务逻辑** — 所有业务规则属于后端服务
3. **API 接口必须与 api-proto 一致** — 接口契约定义在 `../../api-proto/`
4. **API 响应直接使用** — 后端遵循单层包装（`.harness/rules/项目编码规范.md §9`），axios 拦截器已剥掉外层 data，`res` 直接就是业务数据。**不需要** `res.data` 再取一层
5. **Vue 模板避免嵌套 `{{ }}`** — `{{` 在模板插值内会被解析为新插值起始。要展示花括号字面量，拆为 `{'{' + v + '}'}`。详见 `.harness/knowledge/memory/vue-template-nested-interpolation.md`
6. **API 调用禁止静默吞错** — catch 块至少要 `console.error`，关键操作需要 `ElMessage.error` 提示用户

## 架构

```
web/pc/
  src/
    api/           # API 调用层（identity.ts, masterdata.ts, aimodel.ts）
    stores/        # Pinia 状态管理（auth, permission, user, division）
    views/         # 页面组件
    components/    # 通用组件（business/ 业务组件）
    utils/         # 工具函数（request.ts — axios 实例 + lossless-json）
    router/        # Vue Router 配置
  tests/
    unit/          # Vitest 单元测试
    e2e/           # Playwright E2E 测试
```

## 常用命令

```bash
npm run dev           # Vite dev server (port 3003)
npm run build         # vue-tsc + vite build
npm run lint          # vue-tsc type check
npm run test:unit     # Vitest
npm run test:e2e      # Playwright
```
