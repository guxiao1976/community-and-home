# AI 软件开发团队设计方案

## 概述

将 Ralph（自主循环）、OpenSpec（规范驱动）、子 Claude 实例（按服务拆分）、Harness Pipeline（质量门禁）、知识持久化（集体记忆）整合为统一的多层 AI 开发团队，实现从需求到交付的全自动化闭环。

---

## 一、三层架构

```
┌───────────────────────────────────────────────────────┐
│  第 0 层：战略层 — 人类                                  │
│  定义产品愿景、审核 OpenSpec、决策架构、最终合并           │
└────────────────────┬──────────────────────────────────┘
                     │ OpenSpec Proposal
                     ▼
┌───────────────────────────────────────────────────────┐
│  第 1 层：编排层 — 全局 Claude                           │
│  需求分析 → 架构设计 → 任务拆分 → 分发 → 集成 → Review    │
│  工具：Skills + Workflows + Agent/Workflow 工具         │
└────────────────────┬──────────────────────────────────┘
                     │ 分发任务到子 Claude / Ralph
                     ▼
┌───────────────────────────────────────────────────────┐
│  第 2 层：执行层 — 子 Claude + Ralph 循环                │
│  每个服务独立 Agent + 独立 Worktree，读 Memory 避坑       │
│  流程：领任务 → 读设计+记忆 → 写代码 → 自测 → 写 CHANGELOG │
└───────────────────────────────────────────────────────┘
         ↑ 经验回流 ↑
┌───────────────────────────────────────────────────────┐
│  知识层：集体记忆（.harness/memory/ + services/*/memory/） │
│  自动捕获 → 审核 → 检索 → 避免再犯                        │
└───────────────────────────────────────────────────────┘
```

---

## 二、角色化智能体设计（混合式 C）

### 设计原则

- 需求分析和架构设计**永远全局**（需要看到业务全貌）
- 开发和测试**按服务并行**（单服务串行，跨服务并行）
- Review 和集成验证**永远全局**（需要全局视角）

### 完整流水线

```
所有功能类型:

  阶段1: 需求分析
    1 个全局 Agent → 产出 OpenSpec proposal + 验收场景
    （读取：CLAUDE.md + 现有 specs/ + MEMORY.md）

  阶段2: 架构设计
    1 个全局 Agent → 产出 design.md + Proto 定义 + 任务拆分
    （读取：阶段1 产出 + 各服务 design.md + MEMORY.md）

                ↓
         ┌─ 判断：涉及几个服务？ ─┐
         ↓                       ↓
      单服务                   跨服务
         ↓                       ↓
  阶段3: Dev(该服务)         阶段3: [Dev(svc1), Dev(svc2), Dev(web)]
         ↓                  ← 并行，各自独立 Worktree
  阶段4: Test(该服务)        阶段4: [Test(svc1), Test(svc2), Test(web)]
         ↓                  ← 并行
         ↓                       ↓
         └─── 阶段5: Review(全局) ───┘
              阶段6: 集成验证(全局，跨服务时)
```

### 各阶段 Agent 职责

#### 需求分析师（Requirement Analyst）

| 属性 | 值 |
|------|-----|
| 模型 | `deepseek-v4-pro`（深度推理，理解模糊需求并转化为精确规格） |
| 输入 | 用户需求描述（一句话或一段话） |
| 产出 | `openspec/changes/<name>/proposal.md` + `specs/<capability>/spec.md` |
| 必读 | 根 CLAUDE.md、现有 specs/、全局 MEMORY.md |
| 隔离 | 无（纯分析，不写代码） |

#### 架构设计师（Architect）

| 属性 | 值 |
|------|-----|
| 模型 | `deepseek-v4-pro`（架构决策、服务边界判断、接口契约设计） |
| 输入 | 阶段1 产出的 proposal + specs |
| 产出 | `openspec/changes/<name>/design.md` + `tasks.md` + Proto 变更 |
| 必读 | 阶段1 产出、各服务 design.md、全局 MEMORY.md |
| 关键决策 | 功能归属哪个服务？接口契约？数据模型？是否破坏性变更？ |
| 隔离 | 无 |

#### 开发工程师（Developer） × N 服务

| 属性 | 值 |
|------|-----|
| 模型 | `deepseek-v4-flash`（高效执行、批量写代码、速度快成本低） |
| 输入 | 本服务的 fix_plan.md（从 tasks.md 拆分而来） |
| 产出 | 代码实现 + 单元测试 + CHANGELOG.md |
| 必读 | 服务 CLAUDE.md → docs/design.md → CHANGELOG.md → MEMORY.md |
| 隔离 | Worktree（每服务一独立分支，互不干扰） |
| 自检 | `go build ./... && go test ./... && go vet ./...` |

#### 测试工程师（QA Engineer） × N 服务

| 属性 | 值 |
|------|-----|
| 模型 | `deepseek-v4-flash`（高效执行构建/测试/覆盖率检查） |
| 输入 | Dev Agent 的产出 + 验收标准 |
| 产出 | `_qa.md` + PASS/FAIL 判定 |
| 步骤 | go build → go vet → go test → 覆盖率检查 → 功能验证 |
| 只读 | 不修改代码，只输出报告 |

#### 代码审查员（Reviewer）

| 属性 | 值 |
|------|-----|
| 模型 | `deepseek-v4-pro`（全局视角、多维度判断、识别隐蔽问题） |
| 输入 | 所有服务的 Dev + Test 产出 |
| 产出 | `_review.md` + PASS/FAIL 判定 |
| 维度 | 架构一致性、设计一致性、规范遵守、代码质量、安全性、可复用性、测试覆盖、变更完整性 |
| 只读 | 不修改代码 |

---

## 三、知识持久化（集体记忆）

### 设计原则

- 统一结构，不按类型拆分
- 语义触发词驱动检索
- 自动捕获 + 人工审核高风险项
- 每次 Agent 启动自动加载

### 记忆文件结构

```markdown
---
triggers: ["关键词1", "关键词2", "关键词3"]   # AI 语义匹配钥匙
service: <服务名 | all | api-proto>          # 适用范围
severity: must-follow | should-follow | info # 严重程度
status: active | draft | superseded          # 生命周期
created: 2026-06-05
updated: 2026-06-05
---

# <标题>

## 为什么会有这条经验
<原因>

## 怎么做
<具体操作>

## 怎么验证
<检查方法>

## 关联经验
- [[other-memory-name]]
```

### 存储位置

```
.harness/memory/MEMORY.md           ← 全局经验索引
.harness/memory/proto-jstype.md
.harness/memory/grpc-only-comms.md

services/<name>/.harness/memory/    ← 服务特有经验
```

### 自动捕获触发点

| 触发条件                    | 捕获者            | 产物                           |
| ----------------------- | -------------- | ---------------------------- |
| QA Agent 返回 FAIL        | Test Agent     | memory 文件（含根因 + 复现步骤 + 修复方案） |
| Reviewer 发现 CRITICAL    | Reviewer Agent | memory 文件（含规范 + 正确做法）        |
| Ralph 熔断器跳闸             | 全局 Claude      | memory 文件（含原因 + 修复方案）        |
| Proto breaking check 失败 | 全局 Claude      | memory 文件（含变更规则）             |
| `go build` 失败（新模式）      | Dev Agent      | memory 文件                    |
| 集成测试失败                  | 全局 Claude      | memory 文件                    |
| 你纠正 Agent 错误            | Agent          | memory 文件                    |

### 记忆生命周期

```
触发捕获 → Agent 草拟（status: draft）
  → severity: info → 自动生效
  → severity: must-follow → needs-review，等待你确认
  → 你审核通过 → status: active
  → 类似场景触发 → 经验有效
  → 规则已过时 → status: superseded（保留历史）
```

### Agent 启动时的记忆加载

```
Agent 启动流程:
  1. 读取 CLAUDE.md（角色定义）
  2. 读取 docs/design.md（业务理解）
  3. 读取 CHANGELOG.md（历史变更）
  4. 读取 MEMORY.md 索引 → 根据 triggers 精读相关 memory → 避坑
  5. 开始执行任务
```

### 与 Harness Pipeline 集成

```
Harness Pipeline:
  Generator → 读取 MEMORY.md → 避免已知错误
      ↓
  QA 失败 → Test Agent 分析根因 → 自动写入 memory/
      ↓
  Generator（修复轮）→ 读到刚写入的 memory → 精准修复
      ↓
  Review 完成 → Reviewer 总结新发现 → 写入 memory/
```

**闭环效果**：同一个坑，第一次踩了记录，第二次 Agent 启动时就能看到并避免。

---

## 四、核心工具集成

### Ralph — 自主执行循环

```
用途：驱动单个服务的任务执行
输入：services/<name>/.ralph/fix_plan.md（从 OpenSpec tasks.md 拆分而来）
机制：每轮启动全新 AI 实例，执行一个任务，自测，标记完成
      读到 MEMORY.md 避坑，失败后写入新 memory
```

**关键配置**：

- `ALLOWED_TOOLS` 白名单对齐项目实际工具链（go, buf, docker, python3 等）
- 超时 30 分钟（复杂任务需要）
- 熔断器阈值：3 次无进展 / 5 次相同错误

### OpenSpec — 规范驱动

```
用途：定义"做什么"（非"怎么做"）
流程：
  Draft (proposal) → Review (你审核) → Apply (AI 实现) → Archive (合并到 specs/)
  阶段1（需求分析 Agent）产出 proposal
  阶段2（架构设计 Agent）产出 design.md + tasks.md
```

### 子 Claude 实例 — 服务专属开发

```
每个服务配备：
  CLAUDE.md       → 角色 + 全局公约引用 + Ralph 执行模式
  docs/design.md  → 数据模型 + 业务流程 + 接口契约
  CHANGELOG.md    → 变更历史
  .harness/memory/ → 服务特有经验
```

### Harness Pipeline — 质量门禁

```
现有：Generator → QA → Reviewer（最多 3 轮修复循环）
增强点：
  - Generator 启动时读 MEMORY.md
  - QA 失败自动写入 memory/
  - Reviewer 发现 CRITICAL 自动写入 memory/
```

### Workflow 工具 — 并行编排

```
用途：跨服务功能时并行调度多个 Dev/Test Agent
示例：submission-record 功能涉及 master-data + web + proto
  phase('Plan')     → 任务拆分
  phase('Develop')  → parallel([Dev(master-data), Dev(web)]) + 全局改 Proto
  phase('Test')     → parallel([Test(master-data), Test(web)])
  phase('Review')   → 全局 Reviewer
  phase('Integrate')→ 全局集成验证
```

---

## 五、以"提交记录功能"为例的完整流程

### Step 1：你提出需求

```
"我需要在 master-data-service 中增加提交记录追踪功能，前端也要有对应页面"
```

### Step 2：阶段1 — 需求分析

```
输入：你的一句话需求
过程：
  1. 读取根 CLAUDE.md（了解项目架构）
  2. 读取 master-data-service/docs/design.md（了解现有模型）
  3. 读取 web/pc 相关代码（了解前端现状）
  4. 读取全局 MEMORY.md（了解历史经验）
输出：
  openspec/changes/submission-record/
    proposal.md           ← 为什么做、影响哪些服务
    specs/submission-record/spec.md   ← GIVEN/WHEN/THEN 验收场景
    specs/submission-record-ui/spec.md
```

### Step 3：阶段2 — 架构设计

```
输入：阶段1 的 proposal + specs
过程：
  1. 决定新增表 md_submission_record 的 DDL
  2. 决定 API 接口签名（Create/List/Get）
  3. 决定是否需要 Proto 变更（REST-only 则不需要）
  4. 拆分任务到各服务
输出：
  openspec/changes/submission-record/
    design.md             ← DDL、数据模型、API 设计、业务流程图
    tasks.md              ← 按服务分组的所有实现任务

同时生成各服务的 fix_plan.md：
  services/master-data-service/.ralph/fix_plan.md   ← 后端任务清单
  web/pc/.ralph/fix_plan.md                         ← 前端任务清单
```

### Step 4：阶段3 — 并行开发（2 个 Dev Agent）

```
并行启动：
  Dev Agent(master-data-service):
    工作目录：services/master-data-service/(Worktree 隔离)
    fix_plan.md:
      - [ ] 1.1 创建 DDL migration
      - [ ] 1.2 创建 GORM 模型 SubmissionRecord
      - [ ] 1.3 实现 CreateSubmissionRecord 业务逻辑
      - [ ] 1.4 实现 ListSubmissionRecords（分页+筛选）
      - [ ] 1.5 实现 GetSubmissionRecord 详情查询
      - [ ] 1.6 添加 REST API handler
      - [ ] 1.7 编写单元测试
      - [ ] 1.8 更新 CHANGELOG.md

  Dev Agent(web/pc):
    工作目录：web/pc/(Worktree 隔离)
    fix_plan.md:
      - [ ] 2.1 创建提交记录列表页面
      - [ ] 2.2 创建提交记录详情组件
      - [ ] 2.3 API 接口对接
      - [ ] 2.4 编写组件测试
      - [ ] 2.5 更新 CHANGELOG.md

每个 Dev Agent 启动时：
  1. 读 CLAUDE.md
  2. 读 docs/design.md
  3. 读 CHANGELOG.md
  4. 读 MEMORY.md ← 载入历史经验
  5. 逐项执行 fix_plan.md
  6. 每完成一项，自测（go build + go test / npm run build + test）
```

### Step 5：阶段4 — 并行测试（2 个 Test Agent）

```
Test Agent(master-data-service):
  - go build ./... → 通过
  - go vet ./... → 通过
  - go test ./... → 覆盖率检查
  - 按 spec.md 验收场景验证
  - 输出 _qa.md + PASS/FAIL

Test Agent(web/pc):
  - npm run build → 通过
  - npm run test:unit → 通过
  - 按 spec.md 验收场景验证
  - 输出 _qa.md + PASS/FAIL

任何 FAIL：
  → Test Agent 分析根因 → 自动写入 memory/
  → 返回阶段3，Dev Agent 修复（读到刚写入的 memory）
```

### Step 6：阶段5 — 全局代码审查

```
Reviewer Agent（全局）:
  - 读取所有服务的 Dev 产出 + Test 报告
  - 8 维度审查（架构一致性、代码质量、安全性等）
  - 输出 _review.md + PASS/FAIL

任何 CRITICAL：
  → Reviewer 写入 memory/
  → 返回阶段3
```

### Step 7：阶段6 — 集成验证（全局 Claude）

```
- 启动 docker-compose 中间件
- 各服务 go build，go.work 解析本地依赖
- 端到端测试（提交记录创建 → 查询 → 前端展示）
- 确认所有改动服务的 CHANGELOG.md 已更新
- 生成集成报告
```

### Step 8：交付

```
全局 Claude 向你汇报：
  - N 个服务改动完成
  - QA/Review/集成验证全部通过
  - 新增 N 条经验到 memory/
  - 等待你的最终审核 → 合并
```

---

## 六、实施路线图

### 第一步：修复基础设施（立即）

- [x] 修复 `.ralphrc` ALLOWED_TOOLS（已完成）
- [x] 重置 Ralph 熔断器（已完成）
- [ ] 创建 `.harness/memory/` 目录和初始 `MEMORY.md`
- [ ] 从现有文档中提取初始经验（CLAUDE.md 硬规则 + 已知踩坑）

### 第二步：标准化 Agent Prompt

- [ ] 更新 `docs/agents/` 中 5 个 Agent 的 prompt（增加记忆读写指令）
- [ ] 统一 `harness-pipeline.js` 与 Agent 定义的维度/流程
- [ ] 创建需求分析师 Agent prompt（当前缺失）

### 第三步：OpenSpec ↔ Ralph 桥接

- [ ] 创建 `openspec-to-ralph` Skill（tasks.md → fix_plan.md 自动转换）
- [ ] 为现有 OpenSpec changes 生成 Ralph 执行计划

### 第四步：端到端实战

- [ ] 选一个真实功能（如 submission-record），完整跑通全流程
- [ ] 记录流程中的摩擦点，迭代优化 prompt

### 第五步：持续运营

- [ ] 积累 memory 经验库
- [ ] 建立常见问题模式库
- [ ] 定期审查 superseded 记忆，清理过时内容

---

## 七、团队角色映射

| 传统角色 | AI 对应 | 模型 | 说明 |
|----------|--------|:---:|------|
| 产品经理 | 你 + OpenSpec | — | 定义需求、审核 proposal |
| 需求分析师 | Requirement Analyst Agent | `deepseek-v4-pro` | 产出 OpenSpec proposal + specs |
| 架构师 | Architect Agent | `deepseek-v4-pro` | 产出 design.md + 任务拆分 |
| 后端开发 | Dev Agent (各服务) | `deepseek-v4-flash` | 服务代码实现 |
| 前端开发 | Dev Agent (web/pc) | `deepseek-v4-flash` | 前端页面和组件 |
| QA 工程师 | QA Agent | `deepseek-v4-flash` | 构建验证 + 测试 |
| Code Reviewer | Reviewer Agent | `deepseek-v4-pro` | 多维度代码审查 |
| DevOps/CI | Ralph + Workflow 编排 | — | 自动化执行和调度 |
| 技术文档 | 各 Agent 产出 | — | design.md / CHANGELOG.md / spec.md |
| 集体记忆 | .harness/memory/ 系统 | — | 经验积累和自动避坑 |
