---
triggers: ["测试", "TDD", "unit test", "integration test", "PR", "提交", "改动"]
type: process
severity: must-follow
service: all
status: active
created: 2026-06-17
updated: 2026-08-09
apply_count: 0
---
## 测试纪律（硬性约束）

采用 **Testing Trophy** (Kent C. Dodds) + **Honeycomb** (Spotify) 混合模型，适配本项目的微服务 monorepo 架构。

任何代码改动提交前，必须按以下 5 层自底向上完成测试，**全部 PASS 方可交付**。

```
         ▲  E2E Tests        ← 核心业务流，≤10%
        ╱ ╲
       ▕   ▏  Contract Tests  ← API 契约（Proto breaking-check，响应格式）
      ▕     ▏
     ▕       ▏  Integration Tests ← 服务+真实DB/Redis，Mock外部gRPC
    ▕         ▏
   ▕           ▏  Unit Tests   ← 纯逻辑无IO，Table-driven
  ▕             ▏
 ▕━━━━━━━━━━━━━━━▏  Static Analysis ← lint, vet, type-check
```

### Static Analysis — 秒级反馈，CI 可自动拦截

| 语言 | 工具 | 命令 |
|------|------|------|
| Go | go vet | `go vet ./...` |
| TS/Vue | vue-tsc | `npx vue-tsc --noEmit` |
| Proto | buf lint | `cd api-proto && make lint` |

- 零配置，CI 必须跑，FAIL = 不允许合并

### Unit Tests — 纯逻辑，无 I/O，毫秒级

- **测什么**：纯函数、算法、校验逻辑、状态转换。Mock 所有外部依赖。
- **命名**：`Test<Function>_<Condition>_<ExpectedResult>`（如 `TestCheckSmallModelOnly_NilModel_ReturnsNeedReview`）
- Go：`go test ./pkg/...`，table-driven 覆盖正常/边界/错误路径
- TS：`npx vitest run` 对应 spec 文件
- **门禁**：新增函数至少 1 个 case；修改函数需确认 cover 不下降

### Integration Tests — 服务 + 真实中间件，Mock 外部 gRPC

- **测什么**：服务与 DB/Redis/Cache 的交互；HTTP handler 的请求-响应链路；RPC client 的序列化/反序列化
- **如何测**：
  - Go：构造请求 → 调 handler/logic → 验证响应 + DB 状态
  - TS：`curl` 验证 API 端点返回正确的 JSON 结构和业务数据
  - 对外部 gRPC 服务用 mock/fake
- **门禁**：所有新增/修改的 API 端点至少 1 个 curl case

### Contract Tests — API 契约不破

- **测什么**：Proto 向后兼容、HTTP 响应格式（单层 data 包装）、字段类型对齐（int64→string）
- **工具**：`buf breaking-check`、`harness-checks.sh` 的 Proto/Snowflake/响应格式检查项
- **门禁**：涉及 proto 变更必须跑 `make ci`

### E2E Tests — 核心业务流，≤10%，有真实部署

- **测什么**：前端操作 → API → 业务逻辑 → DB 往返的完整链路
- **只测**：top 3-10 核心业务流（如登录、发帖、审核）
- **工具**：curl 脚本（后端）、Playwright（前端）、手动走一遍浏览器
- **门禁**：每个需求至少验证 1 条 Happy Path E2E

### 执行要求

| 规则 | 说明 |
|------|------|
| **自底向上** | Static → Unit → Integration → Contract → E2E，任一层 FAIL 不进入下一层 |
| **证据驱动** | 不允许 "应该没问题"，必须有输出日志/curl 结果/截图佐证 |
| **记录结果** | 写入对应 `_qa.md` 或变更记录的 `## Testing` 章节 |
| **harness 收尾** | 5 层全部通过后，跑 `bash .harness/skills/qa/scripts/harness-checks.sh --service <name>` |
| **失败路由** | 编译失败 → 修代码；测试 FAIL → 修代码或测试；契约破坏 → 回退 Proto 变更 |

### 与 owner-agent 阶段 5 的关系

本纪律细化 owner-agent.md 阶段 5（编码+测试+QA+Review）中的测试环节：

- **编码步骤**：完成 Static Analysis + Unit Tests
- **QA 步骤**：完成 Integration + Contract + E2E Tests
- **门禁**：5 层全 PASS + harness PASS → 进入 Review

**Why:** 用户反馈本轮开发测试不足。行业最佳实践（Testing Trophy, Honeycomb）提供了专业分层框架，本项目据此建立硬性测试纪律。

**How to apply:** 每次改动按 5 层顺序执行，结果记录在 `_qa.md`。任何一层 FAIL 不提交。
