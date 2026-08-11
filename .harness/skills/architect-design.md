# architect-design

## 触发条件

将需求规格转化为技术设计方案和可执行的任务清单。触发词：`架构设计`、`技术设计`、`写 design`、`拆分任务`、`出任务清单`、`设计方案`。

## 角色

你是架构设计师 — 做技术决策：功能归属、接口契约、数据模型、Proto 变更。**不写业务代码**。

## 前置条件

必须先完成 `requirement-analysis`，已有 `proposal.md` + `specs/*/spec.md`。

## 执行步骤

### Step 1: 加载上下文

按顺序：
1. 阶段 1 产出 — `.harness/changes/<name>/proposal.md` + 所有 `specs/*/spec.md`
2. 受影响服务的 `docs/design.md` — 现有数据模型和接口
3. 根 `CLAUDE.md` + `.harness/rules/工程结构.md` — 全局架构约束
4. `api-proto/api/` — 现有 Proto 定义，避免重复或冲突
5. `.harness/knowledge/memory/MEMORY.md` — 精读相关架构决策记忆

### Step 1.5: 记忆注入（设计阶段预防 — 防 Review 被动发现）

**在写 design.md 之前**，主动搜索并注入经验记忆。不做到 Review 时才被动发现遗漏。

1. **提取技术关键词** — 从 proposal + specs 中提取：服务边界/gRPC/Proto/数据库/缓存/安全/Migration/性能/ID生成 等
2. **两级匹配搜索** — 用关键词精确匹配 MEMORY.md 索引中的 triggers（优先），不足再正文降权搜索
3. **按设计章节归类** — 将匹配的记忆分配到 design.md 对应章节：
   | 记忆 | 适用章节 | 设计影响 |
   |------|---------|---------|
   | [[memory-slug]] | 数据模型 | 主键用 BIGINT Snowflake + `json:",string"` |
   | [[memory-slug]] | 接口设计 | gRPC 消息大小限制，大列表必须分页 |
   | [[memory-slug]] | 安全考虑 | 输入校验/参数化查询 |
4. **高风险检查** — 涉及数据迁移/权限变更/Proto破坏性变更/性能敏感路径 的 Task，**必须**有记忆引用
5. **输出注入报告**：
   ```
   记忆注入报告: 匹配 N 个, 注入 M 个, 不适用 K 个
   - [[slug-1]] → 数据模型 §X — <设计决策>
   - [[slug-2]] → 接口设计 §Y — <设计决策>
   ```

### Step 2: 服务归属决策

对每个功能点，判断归属服务：

```
决策原则：
1. 谁拥有数据 → 谁提供接口
2. 谁负责该业务领域 → 谁处理逻辑
3. 是否涉及前端 → 哪些页面
```

产出归属表：

| 功能 | 归属服务 | 理由 |
|------|---------|------|

### Step 3: 产出 design.md

写入 `.harness/changes/<name>/design.md`：

```markdown
# Design: <功能名称>

## 服务归属决策
| 功能 | 归属服务 | 理由 |
|------|---------|------|

## 数据模型
### 新增表：<table_name>
```sql
CREATE TABLE <table_name> (
    id BIGINT PRIMARY KEY,
    ...
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```
### 索引设计
- idx_xxx (column1, column2) — <用途>

## 接口设计
### <ServiceName>.<MethodName>
- **输入**：<字段列表>
- **输出**：<字段列表>
- **错误码**：<5位错误码>

## 业务流程
<关键流程描述>

## Proto 变更
| 文件 | 变更类型 | 说明 |
|------|:---:|------|

## 安全考虑
- <安全点>

## 记忆引用（设计阶段预防性注入，Step 1.5 产出）
| 记忆 | 适用章节 | 设计决策 |
|------|---------|---------|
| [[memory-slug]] | <数据模型/接口设计/安全/...> | <该记忆如何影响本设计的具体决策> |
```

### Step 4: 产出 tasks.md

写入 `.harness/changes/<name>/tasks.md`，按服务分组：

**参考 writing-plans 原则**（`superpowers:writing-plans`）：独立子任务、精确文件路径、TDD 步骤、零占位符。

```markdown
# Tasks: <功能名称>

> **对执行 Agent 的指令**: 每个 Task 独立可测，按 TDD 执行（先写测试→看失败→写实现→看通过）。精确到文件路径。

## 全局 / Proto（由全局 Claude 执行）

### Task 0.1: 定义 <MethodName> Proto
- **文件**: `api-proto/api/<svc>/v1/<file>.proto`
- [ ] 添加 message 定义（Request/Response）
- [ ] 为 int64 ID 字段标注 `[jstype = JS_STRING]`
- [ ] 定义 RPC 方法

### Task 0.2: 生成代码+CI
- [ ] `cd api-proto && make generate`
- [ ] `make lint` → 确认 0 errors
- [ ] `make breaking-check` → 确认无破坏性变更

## <服务名1>

### Task 1.1: <功能模块> — 数据模型
- **创建**: `services/<name>/internal/model/<entity>.go`
- **修改**: `services/<name>/internal/svc/service_context.go`
- [ ] 定义 Model struct（Snowflake ID 用 int64 + json:",string"）
- [ ] 注册 Migration
- [ ] **TDD**: 无逻辑代码可不写测试

### Task 1.2: <功能模块> — Logic 层
- **创建**: `services/<name>/internal/logic/<action>logic.go`
- **创建**: `services/<name>/internal/logic/<action>logic_test.go`
- [ ] **RED**: 写 table-driven tests（正常路径+边界+错误路径）
- [ ] **确认 RED**: `go test -run TestXxx` → 看到 FAIL
- [ ] **GREEN**: 实现 Logic 函数（刚好过测试）
- [ ] **确认 GREEN**: `go test -run TestXxx` → 看到 PASS
- [ ] **REFACTOR**: 清理重复/命名，保持测试绿

### Task 1.3: <功能模块> — Handler 注册
- **修改**: `services/<name>/internal/handler/routes.go`
- [ ] 注册路由/Handler

## <服务名2>
...

## 前端

### Task 3.1: <页面/组件>
- **创建**: `web/<pc|mobile>/src/views/<Page>.vue`
- **修改**: `web/<pc|mobile>/src/router/index.ts`
- [ ] 创建页面/组件
- [ ] 注册路由
```

### Step 5: Tasks Self-Review（产出后自检）

在完成 tasks.md 后逐项检查：

1. **占位符扫描** — 搜索 `<任务描述>`、`TBD`、`TODO`。每个任务必须具体到文件路径
2. **TDD 覆盖** — 每个含新增逻辑的 Task 是否包含 RED→GREEN 步骤？
3. **依赖顺序** — 数据模型 → Logic → Handler，基础设施 → 核心 → 辅助？
4. **独立可测** — 每个 Task 能否独立完成和测试？还是必须等其他 Task？
5. **记忆引用** — 提取 tasks.md 中的技术关键词（gRPC/Proto/Migration/缓存/安全 等），搜索 `.harness/knowledge/memory/MEMORY.md` 索引：
   - must-follow 记忆被遗漏 → 🔴 在对应 Task 描述中补充 `// SEE: [[memory-slug]]` 引用
   - 高风险 Task（数据迁移/Proto 变更/安全相关）**必须**标注相关记忆引用
   - 输出检查结果：`记忆引用检查: N 个相关记忆, M 个已引用, K 个遗漏已补充`

发现问题 → 就地修复。

### Step 6: 产出各服务 fix_plan.md

从 tasks.md 拆分，写入各服务目录 `services/<name>/.ralph/fix_plan.md`。

用 `openspec-to-ralph` Skill 执行此步骤。

## 关键规则

1. **服务归属**：谁拥有数据，谁提供接口
2. **Proto 变更标记**：涉及 Proto 的任务归到「全局 / Proto」组，不分发给子 Claude
3. **任务粒度**：每个 Task 拆到独立可测的子步骤（2-5 分钟/步骤），参考 writing-plans 的 bite-sized 原则
4. **TDD 内建**：每个含逻辑代码的 Task 默认包含 RED→GREEN→REFACTOR 步骤，不只是"写代码"
5. **精确路径**：每个 Task 列出明确的文件路径（创建/修改），零占位符
6. **依赖顺序**：基础设施 → 核心逻辑 → 辅助功能 → 前端
7. **破坏性变更**：必须在 design.md 中明确标注并评估影响范围
8. **记忆引用**：如发现相关架构决策记忆，引用到 design.md 中 `[[memory-slug]]`
9. **Snowflake ID**：所有新表主键用 BIGINT + Snowflake 生成，Go 端用 `json:",string"`
10. **错误码**：新功能使用 5 位错误码 `XXYYY`（XX=服务中心），用 errx 命名常量

## 反例

| ❌ 错误 | ✅ 正确 |
|---------|--------|
| 所有功能归到一个服务 | 按数据所有权和业务领域分拆 |
| tasks.md 中写 "实现登录" | 拆分为 "JWT 签发逻辑 / 验证码校验 / 登录 API / 前端页面" |
| Proto 变更分发给子 Claude | 标记到「全局 / Proto」组，由全局 Claude 执行 |
| 跳过 design.md 直接写任务 | 先设计数据模型和接口契约，再拆任务 |

## 产出物

```
.harness/changes/<name>/
├── proposal.md          ← 阶段1产出
├── specs/*/spec.md      ← 阶段1产出
├── design.md            ← 本阶段产出
├── tasks.md             ← 本阶段产出
└── (各服务 fix_plan.md)  ← 由 openspec-to-ralph Skill 生成
```

## 下一步

- 全局 / Proto 任务 → 由全局 Claude 执行
- 各服务任务 → 用 `dispatch` Skill（统一入口，自动 S/M/L 分级路由到 `harness-pipeline` Workflow）

## 关联

- 需求分析：`.harness/skills/requirement-analysis.md`
- 派发 Skill：`.harness/skills/dispatch.md`
- OpenSpec→Ralph：`.harness/skills/openspec-to-ralph.md`
- **互补**: `Skill("superpowers:writing-plans")` — 更细粒度的实现计划编写（TDD 步骤 + 精确文件路径），tasks.md 格式已对齐 writing-plans 原则
- 工程结构：`.harness/rules/工程结构.md`
- 项目编码规范：`.harness/rules/项目编码规范.md`
