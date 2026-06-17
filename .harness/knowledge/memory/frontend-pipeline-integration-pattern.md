---
triggers: "前端 管线 pipeline web/ Vue TypeScript generator QA review isFrontend vitest npm"
status: active
severity: should-follow
type: guideline
created: 2026-06-17
updated: 2026-06-17
last_applied: null
apply_count: 0
---

# 前端管线接入模式：条件分支而非独立文件

## 场景

2026-06-17 将 `web/pc/` 和 `web/mobile/` 接入 Harness 管线。最初的 Go-centric 管线（`go build/vet/test`、Proto/gRPC 检查）完全不适用于前端。

## 方案

**不创建独立的 generator-frontend.js、qa-frontend.js、review-frontend.js**。而是在现有文件中用 `isFrontend` 条件分支：

```js
const isFrontend = (SVC_DIR || '').startsWith('web/')
const buildCmd = isFrontend ? 'npm run build' : 'go build ./...'
const testCmd  = isFrontend ? 'npm run test:unit' : 'go test ./... -count=1'
```

### 三个阶段的适配

**Generator**：`generator.js`
- Go: table-driven TDD → `go test`
- 前端: vitest TDD → `npm run test:unit`，用 `@vue/test-utils mount`

**QA**：`qa.js`
- Go: `harness-checks.sh` 14 项（go build/vet/test/proto/...）
- 前端: `harness-checks-frontend.sh` 6 项（type-check/unit-test/build/secrets/debug-artifacts/type-safety）

**Review**：`review.js`
- Go: 检查 Proto/gRPC/jstype/json_string/跨服务 DB
- 前端: 检查 TypeScript 类型/no as any/no console.log/Snowflake string/XSS/ElMessage/API 契约对齐

### 前端专属硬边界

```
- 不修改 web/common/ — 那是共享层
- API 接口必须与 api-proto 一致 — 字段名和类型对齐后端 Proto 定义
- 所有 ID 字段使用 string 类型（Snowflake 精度）
```

### 前置条件

前端服务必须满足最小工具链才能通过 QA：
- `package.json` 中必须有 `build`、`test:unit`、`type-check` 三个脚本
- `test:unit` 需要 vitest + @vue/test-utils + happy-dom
- `type-check` 需要 vue-tsc

## 实践规则

1. 新增前端项目时，先用 `harness-checks-frontend.sh` 验证 6 项机械检查能跑通
2. 前端 prompt 修改与 Go prompt 修改在同一文件中进行——不要创建独立的前端 generator
3. `isFrontend` 检测基于 `SVC_DIR.startsWith('web/')`，不要引入新的 args 参数
