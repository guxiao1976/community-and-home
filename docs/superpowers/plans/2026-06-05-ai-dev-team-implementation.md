# AI 软件开发团队 — 实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将方案文档 `docs/specs/ai-dev-team-design.md` 落地为可运转的 AI 开发团队基础设施

**Architecture:** 5 个阶段按依赖顺序推进 —— 先建记忆系统（知识底座），再标准化 Agent Prompt（角色定义），接着统一 Pipeline/Skills（编排层），然后建立 OpenSpec↔Ralph 桥接（工具链），最后端到端实战验证

**Tech Stack:** Markdown 文件 + JavaScript Workflow + Claude Code Skills/Agent 工具

**涉及文件：** 6 个新建、6 个修改

---

## 文件结构

```
新建：
  .harness/memory/MEMORY.md                        ← 全局经验索引
  .harness/memory/proto-jstype.md                   ← 初始经验：Snowflake ID 序列化
  .harness/memory/grpc-only-comms.md                ← 初始经验：gRPC 通信约束
  docs/agents/REQUIREMENT_AGENT.md                 ← 需求分析师 prompt
  docs/agents/ARCHITECT_AGENT.md                   ← 架构设计师 prompt
  .harness/skills/openspec-to-ralph.md              ← OpenSpec → Ralph 桥接 Skill

修改：
  docs/agents/DEV_AGENT.md                         ← 增加记忆读写指令 + 模型指定
  docs/agents/TEST_AGENT.md                        ← 增加记忆读写指令 + 模型指定
  docs/agents/MAIN_AGENT.md                        ← 增加阶段1/2 Agent 分发 + 记忆指令
  .harness/workflows/harness-pipeline.js            ← 记忆集成 + 维度统一
  .harness/skills/dispatch.md                       ← 记忆加载步骤
  .harness/skills/review.md                         ← 维度数量修正（7 → 8）
```

---

## 阶段一：记忆系统底座

### Task 1: 创建 .harness/memory/ 目录和 MEMORY.md 索引

**Files:**
- Create: `.harness/memory/MEMORY.md`

- [ ] **Step 1: 创建 memory 目录**

```bash
mkdir -p .harness/memory
```

- [ ] **Step 2: 创建 MEMORY.md 索引文件**

写入 `.harness/memory/MEMORY.md`：

```markdown
# 全局经验索引

> Agent 启动时读取本文件，根据当前任务上下文精读相关记忆文件。
> 格式：`- [标题](文件.md) — 适用范围, 严重程度, 触发关键词`

## 必须遵守 (must-follow)

- [Proto int64 字段必须加 jstype=JS_STRING](proto-jstype.md) — api-proto, must-follow, `proto int64 jstype JS_STRING Snowflake`
- [服务间通信仅通过 gRPC](grpc-only-comms.md) — all, must-follow, `gRPC 服务间调用 直连数据库`

## 应该遵守 (should-follow)


## 参考信息 (info)

```

- [ ] **Step 3: 提交**

```bash
git add .harness/memory/MEMORY.md
git commit -m "feat: create global memory index for AI dev team"
```

---

### Task 2: 从现有规范提取初始经验

**Files:**
- Create: `.harness/memory/proto-jstype.md`
- Create: `.harness/memory/grpc-only-comms.md`

- [ ] **Step 1: 创建 proto-jstype.md**

写入 `.harness/memory/proto-jstype.md`：

```markdown
---
triggers: ["proto", "int64", "jstype", "JS_STRING", "Snowflake", "ID", "精度丢失", "序列化"]
service: api-proto
severity: must-follow
status: active
created: 2026-06-05
updated: 2026-06-05
---

# Proto int64 字段必须加 jstype=JS_STRING

## 为什么会有这条经验
Snowflake 生成 19 位 ID，超过 JavaScript `Number.MAX_SAFE_INTEGER`（约 16 位）。
前端 JSON 解析时 int64 数字精度丢失，导致 ID 在请求/响应中不一致。

## 怎么做
1. Proto 中所有 int64 ID 字段加 `[jstype = JS_STRING]` 注解
2. Go 端 REST API 类型中 int64 ID 字段使用 `json:"...,string"` 标签
3. 前端所有 ID 字段 TypeScript 类型为 `string`，axios 使用 `lossless-json` 解析器

## 怎么验证
- `cd api-proto && make breaking-check` 会检测缺失的注解
- 前端请求-响应 ID 一致性测试
- 注意：`repeated int64` 字段也需要加 `[jstype = JS_STRING]`

## 关联经验
- [[grpc-only-comms]]
```

- [ ] **Step 2: 创建 grpc-only-comms.md**

写入 `.harness/memory/grpc-only-comms.md`：

```markdown
---
triggers: ["gRPC", "服务间通信", "直连数据库", "HTTP调用", "跨服务"]
service: all
severity: must-follow
status: active
created: 2026-06-05
updated: 2026-06-05
---

# 服务间通信仅通过 gRPC，禁止直连数据库

## 为什么会有这条经验
moderation-service 曾直接读取 masterdata_db 数据库，绕过 master-data-service 的 gRPC 接口。
这导致数据逻辑分散、耦合紧密、难以维护。这是架构债务。

## 怎么做
1. 所有服务间数据访问必须通过 gRPC 接口
2. 接口定义在 `api-proto/` 中统一管理
3. 禁止在服务中配置其他服务的数据库连接
4. 如需其他服务的数据，调用其 gRPC 接口，而非直接查库

## 怎么验证
- 检查服务配置文件中是否有其他服务的数据库 DSN
- `grep -r "mysql://" services/<name>/` 应只包含本服务的数据库
- gRPC 客户端初始化在 `internal/svc/servicecontext.go` 中

## 关联经验
- [[proto-jstype]]
```

- [ ] **Step 3: 更新 MEMORY.md 索引**

更新 `.harness/memory/MEMORY.md`，确保两条经验在"必须遵守"列表中有条目。

- [ ] **Step 4: 提交**

```bash
git add .harness/memory/
git commit -m "feat: extract initial memories from CLAUDE.md rules"
```

---

## 阶段二：Agent Prompt 标准化

### Task 3: 创建需求分析师 Agent Prompt

**Files:**
- Create: `docs/agents/REQUIREMENT_AGENT.md`

- [ ] **Step 1: 创建 REQUIREMENT_AGENT.md**

写入 `docs/agents/REQUIREMENT_AGENT.md`：

````markdown
# Requirement Analyst Agent

## 角色

你是**需求分析师**，负责将用户的模糊需求转化为精确的、可验收的 OpenSpec 规格文档。

你**不写代码**，只做分析和文档产出。

## 模型

使用 `deepseek-v4-pro`（深度推理模型），因为需求分析需要理解模糊描述、识别隐含约束、发现边界条件。

## 启动上下文

按顺序读取以下文件：

1. 根 `CLAUDE.md` — 了解项目架构、服务划分、全局约束
2. `openspec/specs/` — 现有规格文档，了解已有功能
3. 相关服务的 `docs/design.md` — 了解现有数据模型和业务流程
4. `.harness/memory/MEMORY.md` — **读取经验索引**，精读相关记忆文件，避免提出已知不可行的方案

## 输入

用户的一句话或一段话需求描述。

## 产出

在 `openspec/changes/<change-name>/` 下创建：

```
openspec/changes/<change-name>/
  .openspec.yaml          # schema: spec-driven, created: YYYY-MM-DD
  proposal.md             # 为什么做、做什么、影响哪些服务、风险评估
  specs/
    <capability-1>/
      spec.md             # GIVEN/WHEN/THEN 验收场景
    <capability-2>/
      spec.md             # （如有多个独立功能）
```

## proposal.md 模板

```markdown
# Proposal: <功能名称>

## 为什么做
<1-2 段说明业务背景和用户价值>

## 做什么
<功能概述，1-2 段>

## 影响范围
| 服务 | 变更类型 | 说明 |
|------|:---:|------|
| xxx-service | 新增 API | ... |
| web/pc | 新增页面 | ... |

## 风险评估
- <风险1：可能性 + 影响 + 缓解措施>
- <风险2>

## 验收标准
- <高层验收点1>
- <高层验收点2>
```

## spec.md 模板

```markdown
# <Capability Name> Specification

## Purpose
<功能目的，至少 50 字符>

## Requirements

### Requirement: <需求名称>
The system SHALL <行为描述，必须含 SHALL 或 MUST>.

#### Scenario: <场景名称>
- **GIVEN** <初始状态>
- **WHEN** <条件或触发>
- **THEN** <预期结果>
- **AND** <附加结果>
```

## 关键规则

1. 每个 Requirement 至少有一个 Scenario
2. 使用 RFC 2119 关键词：SHALL, MUST, SHOULD, MAY
3. Spec 描述**行为契约**，不描述实现细节
4. 不要猜测——不确定时标注 `[NEEDS CLARIFICATION: 具体问题]` 并列出选项
5. 涉及多个服务时，明确各服务的职责边界
6. 读取 MEMORY.md 后如果发现相关经验，在 proposal 中引用

## 禁止

- 写代码
- 设计数据库表结构（那是架构师的工作）
- 猜测技术实现细节
````

- [ ] **Step 2: 提交**

```bash
git add docs/agents/REQUIREMENT_AGENT.md
git commit -m "feat: add requirement analyst agent prompt"
```

---

### Task 4: 创建架构设计师 Agent Prompt

**Files:**
- Create: `docs/agents/ARCHITECT_AGENT.md`

- [ ] **Step 1: 创建 ARCHITECT_AGENT.md**

写入 `docs/agents/ARCHITECT_AGENT.md`：

````markdown
# Architect Agent

## 角色

你是**架构设计师**，负责将需求规格转化为技术设计方案和可执行的任务清单。

你需要做出技术决策：功能归属、接口契约、数据模型、Proto 变更。你**不写业务代码**。

## 模型

使用 `deepseek-v4-pro`（深度推理模型），因为架构决策需要全局视角、服务边界判断、接口契约设计能力。

## 启动上下文

按顺序读取以下文件：

1. 阶段1 产出 — `openspec/changes/<name>/proposal.md` + 所有 `specs/*/spec.md`
2. 受影响服务的 `docs/design.md` — 了解现有数据模型和接口
3. 根 `CLAUDE.md` — 了解全局架构约束
4. `api-proto/api/` — 现有 Proto 定义，避免重复或冲突
5. `.harness/memory/MEMORY.md` — **读取经验索引**，精读相关记忆，利用已有决策

## 输入

阶段1（需求分析师）产出的 proposal.md + specs/*/spec.md。

## 产出

### 1. design.md

```markdown
# Design: <功能名称>

## 服务归属决策
| 功能 | 归属服务 | 理由 |
|------|---------|------|
| <功能点1> | <服务名> | <决策依据> |

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
<关键流程描述，复杂流程用流程图>

## Proto 变更
| 文件 | 变更类型 | 说明 |
|------|:---:|------|
| api/xxx/v1/xxx.proto | 新增 Message | ... |

## 安全考虑
- <安全点1>
- <安全点2>
```

### 2. tasks.md（按服务分组）

```markdown
# Tasks: <功能名称>

## 全局 / Proto
- [ ] 0.1 修改 api-proto/api/<svc>/v1/<file>.proto
- [ ] 0.2 cd api-proto && make generate && make lint && make breaking-check

## <服务名1>
- [ ] 1.1 <任务描述>
- [ ] 1.2 <任务描述>

## <服务名2>
- [ ] 2.1 <任务描述>
```

### 3. 各服务的 fix_plan.md

从 tasks.md 拆分，写入各服务目录：
`services/<name>/.ralph/fix_plan.md`

## 关键规则

1. **服务归属决策**：优先考虑数据所有权——谁拥有数据，谁提供接口
2. **Proto 变更**：由全局 Claude 执行，不在 tasks.md 中分发给子 Claude
3. **任务拆分粒度**：每个任务 1-4 小时工作量，1-5 个文件，可独立测试
4. **依赖顺序**：基础设施 → 核心逻辑 → 辅助功能 → 前端
5. **破坏性变更**：必须在 design.md 中明确标注并评估影响范围
6. 读取 MEMORY.md 后如发现相关架构决策，引用到 design.md 中

## 禁止

- 写业务代码
- 跳过 Proto 变更直接影响评估
- 在不了解现有设计的情况下做决策
````

- [ ] **Step 2: 提交**

```bash
git add docs/agents/ARCHITECT_AGENT.md
git commit -m "feat: add architect agent prompt"
```

---

### Task 5: 更新 Dev Agent Prompt（增加记忆读写）

**Files:**
- Modify: `docs/agents/DEV_AGENT.md`

- [ ] **Step 1: 在角色定义后增加模型和记忆指令**

在 `docs/agents/DEV_AGENT.md` 的角色说明后（"## 角色" 段落之后），插入以下内容：

```markdown
## 模型

使用 `deepseek-v4-flash`（高效执行模型），开发任务需要快速迭代、批量写代码，速度和成本优先。

## 记忆系统

### 启动时加载经验
在开始任何开发工作前，按顺序执行：
1. 读取 `.harness/memory/MEMORY.md`（全局经验索引）
2. 读取 `services/<本服务>/.harness/memory/MEMORY.md`（服务特有经验，如果存在）
3. 根据当前任务的上下文关键词，精读匹配的记忆文件
4. **在开发过程中主动应用这些经验，避免重复已知错误**

### 遇到新问题时记录经验
当遇到以下情况，自动创建记忆文件：
1. `go build` 失败且根因不是拼写错误（新模式失败）
2. `go test` 失败且根因不是测试逻辑错误（环境/依赖/配置问题）
3. 你自己发现了一个非显而易见的技术约束

记忆文件格式见 `.harness/memory/MEMORY.md` 中的说明。写入后更新对应的 MEMORY.md 索引。
```

- [ ] **Step 2: 在"开发阶段"步骤中增加记忆加载步骤**

找到 Dev Agent 的"开发阶段"（Phase 1: Understand task 附近），在"理解任务"之前插入：

```markdown
0. **加载经验**：读取 MEMORY.md 索引 → 精读相关记忆文件
```

- [ ] **Step 3: 提交**

```bash
git add docs/agents/DEV_AGENT.md
git commit -m "feat: add memory system and model spec to dev agent"
```

---

### Task 6: 更新 Test Agent Prompt（增加记忆读写）

**Files:**
- Modify: `docs/agents/TEST_AGENT.md`

- [ ] **Step 1: 增加记忆读写指令**

在 `docs/agents/TEST_AGENT.md` 的角色说明后，插入：

```markdown
## 模型

使用 `deepseek-v4-flash`（高效执行模型），测试任务需要运行构建/测试/覆盖率检查，速度优先。

## 记忆系统

### 启动时加载经验
1. 读取 `.harness/memory/MEMORY.md`（全局经验索引）
2. 读取 `services/<目标服务>/.harness/memory/MEMORY.md`（服务特有经验）
3. 根据测试目标，精读相关记忆文件中的"怎么验证"章节
4. 在测试报告中引用相关经验

### 测试失败时记录经验
当 QA 返回 FAIL 时，分析根本原因并创建记忆文件：
1. 确定根因（不是表面错误，是为什么产生这个错误）
2. 写入 `.harness/memory/<slug>.md`（全局）或 `services/<svc>/.harness/memory/<slug>.md`（服务特有）
3. 记忆文件必须包含：原因、复现步骤、修复方案、验证方法
4. 更新对应的 MEMORY.md 索引

如果已有相关记忆文件，更新它（增加新的复现场景或关联经验）而非创建新文件。
```

- [ ] **Step 2: 提交**

```bash
git add docs/agents/TEST_AGENT.md
git commit -m "feat: add memory system and model spec to test agent"
```

---

### Task 7: 更新 Main Agent Prompt（增加阶段1/2）

**Files:**
- Modify: `docs/agents/MAIN_AGENT.md`

- [ ] **Step 1: 增加需求分析和架构设计阶段的分发指令**

在 Main Agent 的 "Phases" 部分（Phase 1: Receive requirements 之前），插入新的阶段0（需求分析 + 架构设计）：

```markdown

## 完整流水线模式

当用户说"全流程"/"完整流水线"/"new feature"时，按以下阶段执行：

### Phase 0: 需求分析与架构设计

#### 0.1 需求分析
- 读取 `docs/agents/REQUIREMENT_AGENT.md` 获取需求分析师 prompt
- 分发 Requirement Analyst Agent，输入为用户需求
- Agent 产出 `openspec/changes/<name>/proposal.md` + `specs/`
- 将产出呈现给用户审核

#### 0.2 架构设计（用户批准 proposal 后）
- 读取 `docs/agents/ARCHITECT_AGENT.md` 获取架构师 prompt
- 分发 Architect Agent，输入为阶段1产出
- Agent 产出 `design.md` + `tasks.md` + 各服务的 `fix_plan.md`
- 将设计呈现给用户审核

#### 0.3 Proto 变更（如果有）
- **Main Agent 亲自执行** Proto 变更（子 Claude 禁止修改 api-proto/）
- `cd api-proto && make generate && make lint && make breaking-check`
- 记录到 `api-proto/CHANGELOG.md`

### Phase 1-3: 执行（原有流程）
按原有 DEV_AGENT → TEST_AGENT 流程执行，但增加：
- 并行模式：跨服务功能时，多个 Dev Agent 并行（使用 Worktree 隔离）

## 记忆系统

### 启动时加载
在开始任何工作前，读取 `.harness/memory/MEMORY.md` 了解历史经验。

### 运行时记录
当遇到以下情况，创建或更新记忆文件：
- Ralph 熔断器跳闸 → 分析原因 → 写入 memory/
- 集成测试失败 → 分析根因 → 写入 memory/
- 用户纠正了 Agent 的错误 → 记录纠正内容
```

- [ ] **Step 2: 提交**

```bash
git add docs/agents/MAIN_AGENT.md
git commit -m "feat: add requirement/architect phases and memory to main agent"
```

---

## 阶段三：Pipeline 与 Skills 统一

### Task 8: 更新 harness-pipeline.js（记忆集成 + 维度统一）

**Files:**
- Modify: `.harness/workflows/harness-pipeline.js`

- [ ] **Step 1: 在 Generator prompt 中增加记忆加载指令**

在 `generatorPrompt` 函数中，在 "按顺序读取以下文件" 的列表末尾，增加：

```javascript
// 在 generatorPrompt 函数中，文件读取列表末尾添加：
`5. **读取 .harness/memory/MEMORY.md（服务：${serviceDir}/.harness/memory/MEMORY.md）** — 加载历史经验，避免重复已知错误
6. 根据任务关键词，精读匹配的记忆文件内容`
```

- [ ] **Step 2: 在 QA prompt 中增加记忆写入指令**

在 `qaPrompt` 函数末尾，增加：

```javascript
// 在 qaPrompt 函数中，return 语句之前添加：
`
## 记忆记录
如果 QA 判定为 FAIL，你必须：
1. 分析根本原因（不是表面错误信息）
2. 检查 .harness/memory/MEMORY.md 是否已有相关经验
3. 如果有 → 更新该记忆文件（增加新的复现场景）
4. 如果没有 → 创建新的记忆文件到 .harness/memory/<slug>.md
5. 更新 MEMORY.md 索引
记忆文件格式：参见 .harness/memory/MEMORY.md 中的说明
`
```

- [ ] **Step 3: 在 Reviewer prompt 中统一维度为 8 项**

确认 `reviewPrompt` 函数中的维度列表是 8 项（架构一致性、设计一致性、规范遵守、代码质量、安全性、可复用性、测试覆盖、变更完整性），与 `reviewers/code-reviewer/CLAUDE.md` 对齐。

- [ ] **Step 4: 在 Reviewer prompt 中增加记忆写入指令**

在 `reviewPrompt` 函数末尾，增加与 QA 类似的记忆记录指令：

```javascript
`
## 记忆记录
如果 Review 发现 CRITICAL 问题，你必须：
1. 判断这是否是一个新的规范/踩坑（而非一次性的代码错误）
2. 如果是新经验 → 创建记忆文件到 .harness/memory/<slug>.md
3. 如果已有相关记忆 → 更新它
4. 更新 MEMORY.md 索引
`
```

- [ ] **Step 5: 提交**

```bash
git add .harness/workflows/harness-pipeline.js
git commit -m "feat: integrate memory system and unify dimensions in harness pipeline"
```

---

### Task 9: 更新 dispatch.md（增加记忆加载步骤）

**Files:**
- Modify: `.harness/skills/dispatch.md`

- [ ] **Step 1: 在标准 dispatch prompt 中增加记忆加载**

在 dispatch.md 的"启动上下文"部分（读取 CLAUDE.md、docs/design.md、CHANGELOG.md 之后），增加：

```markdown
4. **读取 .harness/memory/MEMORY.md**（全局经验索引），根据任务关键词精读相关记忆文件
5. 读取 `services/<service>/.harness/memory/MEMORY.md`（服务特有经验，如果存在）
```

- [ ] **Step 2: 提交**

```bash
git add .harness/skills/dispatch.md
git commit -m "feat: add memory loading step to dispatch skill"
```

---

### Task 10: 修正 review.md 维度数量

**Files:**
- Modify: `.harness/skills/review.md`

- [ ] **Step 1: 将 "7 维度" 改为 "8 维度"**

查找 `review.md` 中的 "7-dimension review" 或 "7 维度"，改为 "8 维度"，与 `reviewers/code-reviewer/CLAUDE.md` 的 8 项对齐。

- [ ] **Step 2: 提交**

```bash
git add .harness/skills/review.md
git commit -m "fix: correct review dimension count from 7 to 8"
```

---

## 阶段四：OpenSpec ↔ Ralph 桥接

### Task 11: 创建 openspec-to-ralph Skill

**Files:**
- Create: `.harness/skills/openspec-to-ralph.md`

- [ ] **Step 1: 创建 openspec-to-ralph.md**

写入 `.harness/skills/openspec-to-ralph.md`：

````markdown
# OpenSpec to Ralph Bridge

## 触发

"生成执行计划" / "openspec to ralph" / "导出任务到 Ralph" / "创建 fix_plan"

## 功能

将 OpenSpec 的 `tasks.md` 按服务拆分，为每个服务生成 Ralph 可执行的 `fix_plan.md`。

## 流程

### Step 1: 定位 OpenSpec change

从用户输入或上下文中确定 change 名称，读取：
```
openspec/changes/<change-name>/
  tasks.md       ← 任务清单
  design.md      ← 技术设计（了解服务归属）
```

### Step 2: 按服务拆分任务

解析 `tasks.md`，识别每个任务的服务归属：
- 任务前缀如 `[proto]` / `[global]` → 全局 Claude 执行
- `## <服务名>` 章节 → 归属该服务
- 任务描述中包含服务名 → 归属对应服务

### Step 3: 为每个服务生成 fix_plan.md

对每个受影响的服务，生成 `services/<name>/.ralph/fix_plan.md`：

```markdown
# Fix Plan: <change-name> — <service-name>

> 来源：openspec/changes/<change-name>/tasks.md
> 生成时间：YYYY-MM-DD HH:MM
> 关联设计：openspec/changes/<change-name>/design.md

## 前置阅读
1. 服务 CLAUDE.md
2. docs/design.md
3. CHANGELOG.md
4. .harness/memory/MEMORY.md ← 加载经验

## 任务清单

- [ ] <task-id> <任务描述>
  - 关联 spec: `<spec-ref>`
  - 验收标准: <criteria>

- [ ] <task-id> <任务描述>
  ...
```

### Step 4: 全局任务保留

Proto 变更等全局任务保留在原地，提示用户由全局 Claude 执行：

```
⚠️ 以下任务需要全局 Claude 执行（子 Claude 不能修改 api-proto/）：
- [ ] 0.1 修改 api-proto/api/<svc>/v1/<file>.proto
- [ ] 0.2 cd api-proto && make generate && make lint && make breaking-check
```

## 输出

```
✅ 已为 N 个服务生成执行计划：
  - services/<svc1>/.ralph/fix_plan.md (M 个任务)
  - services/<svc2>/.ralph/fix_plan.md (K 个任务)
  - web/pc/.ralph/fix_plan.md (P 个任务)

⚠️ 1 个全局任务需要亲自执行（Proto 变更）

下一步：切换到各服务目录，运行 ralph start
```
````

- [ ] **Step 2: 提交**

```bash
git add .harness/skills/openspec-to-ralph.md
git commit -m "feat: add openspec-to-ralph bridge skill"
```

---

## 阶段五：端到端实战验证

### Task 12: 选取功能进行全流程验证

**无文件变更**（验证任务，不产生代码）

- [ ] **Step 1: 选取验证目标**

选取 `openspec/changes/submission-record/`（已存在提案，最完整），或一个新的小功能。

- [ ] **Step 2: 按全流程执行**

```
阶段1: 用 REQUIREMENT_AGENT 生成 OpenSpec proposal → 审核
阶段2: 用 ARCHITECT_AGENT 生成 design.md + tasks.md + fix_plan.md → 审核
阶段3: 用 openspec-to-ralph skill 导出 → 检查 fix_plan.md 是否正确
阶段4: 用 harness-pipeline 执行开发+测试（单服务）
阶段5: 验证记忆系统是否自动记录了经验
```

- [ ] **Step 3: 记录摩擦点**

在验证过程中记录以下信息：
1. 哪个阶段 Agent 产出不符合预期？Prompt 如何调整？
2. 记忆系统是否自动触发了捕获？记忆内容质量如何？
3. 任务拆分粒度是否合适？有没有过大或过小的任务？
4. 有没有流程步骤遗漏或不必要的？

- [ ] **Step 4: 迭代优化**

根据摩擦点更新相关 Agent prompt 和 Skill。

---

## 阶段六：持续运营

### Task 13: 建立记忆维护机制

**Files:**
- Create: `.harness/memory/MAINTENANCE.md`

- [ ] **Step 1: 创建维护指南**

写入 `.harness/memory/MAINTENANCE.md`：

```markdown
# 记忆系统维护指南

## 日常

- 每次 Agent 执行完毕后，检查是否有新记忆被创建（`git status .harness/memory/`）
- 新记忆 status: draft → 快速浏览，确认内容合理 → 改为 status: active

## 每周

- 审查 status: active 的记忆文件：
  - 是否有类似经验重复出现？ → 合并
  - 是否有经验已经不再适用？ → 改为 status: superseded
  - triggers 关键词是否准确？ → 补充或修正

## 每月

- 分析 superseded 记忆，总结模式（什么类型的经验容易过时？）
- 将高频触发的经验提升为 CLAUDE.md 中的硬规则
- 清理 3 个月以上的 superseded 记忆（git rm）

## 命令速查

```bash
# 查看所有记忆
ls -la .harness/memory/ services/*/.harness/memory/

# 查找特定主题的记忆
grep -rl "关键词" .harness/memory/ services/*/.harness/memory/

# 查看记忆状态统计
grep -r "status:" .harness/memory/ | sort | uniq -c

# 列出 superseded 记忆
grep -rl "status: superseded" .harness/memory/
```
```

- [ ] **Step 2: 提交**

```bash
git add .harness/memory/MAINTENANCE.md
git commit -m "docs: add memory system maintenance guide"
```

---

## 依赖关系

```
Phase 1 (Task 1-2)  ─────────────────────────────┐
     ↓                                             │
Phase 2 (Task 3-7)  ← depends on Phase 1          │
     ↓                                             │
Phase 3 (Task 8-10) ← depends on Phase 2          │
     ↓                                             │
Phase 4 (Task 11)   ← depends on Phase 2          │
     ↓                                             │
Phase 5 (Task 12)   ← depends on Phase 1-4 ───────┤
                                                    │
Phase 6 (Task 13)   ← no hard dependency ──────────┘
```

Phase 1-4 必须按顺序执行。Phase 5 依赖前 4 个阶段。Phase 6 可在 Phase 1 完成后随时开始。

---

## 执行建议

建议分三次会话完成：

| 会话 | 阶段 | 预计耗时 |
|------|------|:---:|
| 会话 1 | Phase 1 (Task 1-2) + Phase 2 (Task 3-7) | ~30 分钟 |
| 会话 2 | Phase 3 (Task 8-10) + Phase 4 (Task 11) | ~20 分钟 |
| 会话 3 | Phase 5 (Task 12) — 全流程验证 | ~45 分钟 |

Phase 6 在验证通过后作为日常习惯建立。
