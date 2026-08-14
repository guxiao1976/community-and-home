# Harness Pipeline Architecture

> 开发流水线架构设计文档 · 最后更新 2026-08-14

---

## 1. 概述

Harness Pipeline 是 Community-Home 项目的自动化质量保障系统，通过**三层质量保障体系**实现从代码提交到生产部署的全流程质量管控。

### 1.1 设计目标

| 目标 | 实现方式 |
|------|---------|
| **零人工规范检查** | 机械化检查自动化 15+6 项规范验证 |
| **快速反馈循环** | TDD + 增量检查，秒级反馈 |
| **质量门禁强制** | QA FAIL 阻塞提交，Review CRITICAL 阻塞合并 |
| **记忆驱动开发** | 自动注入历史经验，避免重复踩坑 |
| **多视角审查** | 3 个 Reviewer 并行，覆盖 9 个维度 |

### 1.2 核心原则

1. **验证优先** (Verification-Before-Completion) - 任何状态声称前必须有 FRESH 证据
2. **TDD 强制** (Test-Driven Development) - 新功能必须先写失败测试
3. **记忆注入** (Memory-Driven) - 编码前自动搜索并应用相关经验
4. **根因分析** (Systematic Debugging) - QA FAIL 时触发根因分析，不做症状修复
5. **多视角审查** (Multi-Perspective Review) - 安全架构/规范工程/设计业务三视角并行

---

## 2. 三层质量保障体系

```
┌─────────────────────────────────────────────────────────────┐
│  Layer 1: 机械化检查层 (Mechanized Checks)                    │
│  ├─ Go 服务: 18 项检查 (harness-checks.sh)                   │
│  ├─ 前端服务: 6 项检查 (harness-checks-frontend.sh)          │
│  └─ 确定性验证层: deterministic-rules.yml（blocker 标记）    │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│  Layer 2: 工作流编排层 (Workflow Orchestration)               │
│  ├─ Generator Agent (TDD + Memory)                          │
│  ├─ QA Agent (Verification-Before-Completion)               │
│  ├─ Debug Agent (Root Cause Analysis)                       │
│  └─ Review Agent (3 Perspectives × 9 Dimensions)            │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│  Layer 3: 全局协调层 (Global Coordination)                   │
│  ├─ dispatch 分级 (S→轻量Pipeline / M→Pipeline / L→OpenSpec) │
│  ├─ 子 Agent 派发与验收                                      │
│  ├─ Proto 变更管控                                           │
│  └─ 跨服务并行调度                                           │
└─────────────────────────────────────────────────────────────┘
```

> 三条「7-8 月演进」补充（详见 `pipeline-evolution.md` Phase 12/14/17）：
> - **确定性验证层**（Phase 12）：`config/deterministic-rules.yml` 把编译/测试/覆盖率等确定性检查显式化，`blocker` 标记 FAIL 阻断 AI 判断，避免「把模型判断当确定性验证」。
> - **dispatch 统一入口 + 工作量分级**（Phase 14）：所有开发任务先走 dispatch 分级（S/M/L），禁止绕过入口直接开发（CLAUDE.md 硬约束 #7）。
> - **流水线检视**（Phase 17）：`pipeline-review` skill + `harness-design-principles.md`（16 条原则标尺），对三层体系做持续检视与进化闭环。

---

## 3. Layer 1: 机械化检查层

### 3.1 Go 服务检查项 (18 项)

| # | 检查项 | 目的 | 阻塞级别 |
|---|--------|------|:---:|
| 1 | `go build ./...` | 编译检查 | FAIL |
| 2 | `go vet ./...` | 静态分析 | FAIL |
| 3 | `go test ./...` | 单元测试 + 0/0 检测 | FAIL |
| 4 | Proto int64 jstype | Snowflake ID 前端兼容 | FAIL |
| 5 | json:",string" | API 响应 ID 序列化 | FAIL |
| 6 | 跨服务 DB 导入 | 服务边界隔离 | FAIL |
| 7 | 错误码格式 | 5 位错误码 + 常量化 | WARN |
| 8 | 硬编码密钥 | 安全性检查 | FAIL |
| 9 | 知识图谱新鲜度 | 上下文同步 | FAIL |
| 10 | CLAUDE.md 结构化数据 | 避免重复维护 | WARN |
| 11 | Proto→TS 对齐 | 前后端类型一致性 | FAIL |
| 12 | API Logic TODO 桩 | 避免空实现上线 | FAIL |
| 13 | Response 单层包装 | 避免双层嵌套 | WARN |
| 14 | Benchmark 回归 | 性能守护 | WARN |
| 15 | API 冒烟测试 | 新增路由连通性 | WARN |

### 3.2 前端服务检查项 (6 项)

| # | 检查项 | 目的 | 阻塞级别 |
|---|--------|------|:---:|
| 1 | `vue-tsc --noEmit` / `tsc --noEmit` | 类型检查 | FAIL |
| 2 | `vitest run` / `npm test` | 单元测试 + 0/0 检测 | FAIL |
| 3 | `vite build` / `npm run build` | 构建检查 | FAIL |
| 4 | 硬编码密钥 | API key/token 检测 | FAIL |
| 5 | 调试残留 | console.log/debugger | WARN |
| 6 | 类型安全 | `as any` 逃逸统计 | WARN |

### 3.3 检查脚本设计

**差分扫描模式** (默认):
- 仅检查 `git diff` 中的变更文件
- 适用于增量开发，秒级反馈

**全量扫描模式** (`--full`):
- 检查整个服务目录
- 适用于发布前全面验证

**JSON 输出模式** (`--json`):
- 结构化输出，供 Pipeline 解析
- 格式：`{timestamp, service, results:[{check, status, detail}], summary:{pass, fail, warn}}`

---

## 4. Layer 2: 工作流编排层

### 4.1 harness-pipeline.js 架构

```javascript
export const meta = {
  name: 'harness-pipeline',
  description: 'Generator → QA → (Debug) → Reviewer 闭环',
  phases: [
    { title: 'Develop', detail: 'Generator 实现/修复代码' },
    { title: 'QA', detail: 'QA Agent 编译+测试验证' },
    { title: 'Debug', detail: '根因分析 (QA FAIL 时触发)' },
    { title: 'Review', detail: '3 视角并行审查' },
  ],
}
```

### 4.2 Generator Agent (开发 Agent)

**职责**: 实现功能代码，遵循 TDD 和记忆驱动开发

**上下文加载** (三层按需):
```
L1: 服务上下文 (~350 lines)
    └─ CLAUDE.md + design.md + graph-context.md + CHANGELOG.md

L2: 任务上下文
    └─ .harness/changes/<change>/design.md + tasks.md

L3: 经验记忆 (按需)
    └─ memory-index-query.sh --union <keywords>
```

**TDD 编码纪律** (feature/bug 类型强制):
1. **RED** - 先写失败测试，必须看到 FAIL 输出
2. **GREEN** - 最小实现通过测试
3. **REFACTOR** - 清理代码，保持全绿

**记忆驱动流程**:
1. 提取任务关键词
2. 查询记忆索引 (`.harness/knowledge/memory/.memory-index.json`)
3. 读取命中的记忆文件
4. 在代码中标记 `// SEE: [[memory-slug]]`
5. 输出记忆应用报告

**产出**:
- 实现代码 + 测试
- CHANGELOG.md 更新
- 记忆应用报告

### 4.3 QA Agent (质量保障 Agent)

**职责**: 验证代码质量，执行机械化检查

**Verification-Before-Completion 纪律**:
```
铁律: NO COMPLETION CLAIMS WITHOUT FRESH VERIFICATION EVIDENCE

声称任何状态前必须:
1. IDENTIFY - 哪条命令能证明？
2. RUN - 执行完整命令 (fresh, 非缓存)
3. READ - 读完整输出、exit code、失败数
4. VERIFY - 输出是否确认了声称？
```

**验证步骤**:
1. 运行机械化检查 (`harness-checks.sh --service <name> --json`)
2. 运行 `go build ./...` / `npm run build`
3. 运行 `go vet ./...` / `npm run type-check`
4. 运行 `go test ./...` / `npm run test:unit`
5. 检查 TDD 证据 (是否有 RED→GREEN 证明)
6. 检查测试覆盖 (新增函数是否有测试)
7. 写入 `_qa.md` (FRESH 覆盖，不追加)
8. 输出 VERDICT (PASS/FAIL)

**产出**:
- `_qa.md` (包含机械化检查结果 + TDD 证据检查)
- VERDICT + failures 详情

### 4.4 Debug Agent (根因分析 Agent)

**职责**: QA FAIL 时触发，分析根本原因

**Systematic Debugging 流程**:
```
Phase 1: Root Cause Investigation
  ├─ 仔细阅读错误信息
  ├─ 复现问题
  ├─ 检查最近变更 (git diff)
  └─ 追溯依赖链

Phase 2: Evidence Collection
  ├─ 提取关键日志
  ├─ 检查相关配置
  └─ 验证假设

Phase 3: Root Cause Confirmation
  └─ 一句话根因描述 + 置信度

Phase 4: Fix Suggestions
  └─ 精确到文件:行号的修复建议
```

**产出**:
- `rootCause` (一句话描述)
- `confidence` (high/medium/low)
- `evidence` (证据链)
- `fixSuggestions` (修复建议列表)

### 4.5 Review Agent (代码审查 Agent)

**职责**: 多视角审查代码，覆盖 9 个维度

**三视角并行**:

| 视角 | 维度 | 关注点 |
|------|------|--------|
| **安全架构** | 架构一致性(#1)、安全性(#5)、变更完整性(#8) | Proto/gRPC 规范、服务边界、跨服务隔离、硬编码密钥、SQL 注入、输入校验、CHANGELOG 完整性 |
| **规范工程** | 规范遵循(#3)、复用性(#6)、测试覆盖(#7)、记忆遵守(#9) | Snowflake ID 序列化、错误码格式、API 响应格式、代码复用、测试覆盖、记忆引用准确性 |
| **设计业务** | 设计一致性(#2)、代码质量(#4)、Migration(#8部分) | 与 design.md 一致性、数据模型正确性、业务流程正确性、边界条件处理、错误处理完善性、Migration 安全性 |

**记忆遵守检查** (规范工程视角):
- M1: 收集代码中的 `// SEE: [[memory-slug]]` 引用
- M2: 验证引用准确性 (slug 存在、代码遵守指导)
- M3: 检查遗漏的记忆 (关键词匹配 triggers)
- M4: 建议新记忆 (模式性问题 + 可复用 + 未覆盖)

**VERDICT 规则**:
- PASS - 负责维度中无 CRITICAL 问题
- FAIL - 存在 ≥1 个 CRITICAL

**产出**:
- `_review_security-arch.md`
- `_review_standards-eng.md`
- `_review_design-biz.md`
- VERDICT + criticalCount + memorySuggestions

### 4.6 Pipeline 循环逻辑

```
Iteration 1:
  Generator (feature) → QA
    ├─ PASS → Review (3 视角并行)
    │   ├─ 2/3 PASS → SUCCESS
    │   └─ <2/3 PASS → Iteration 2 (fixContext = Review CRITICAL)
    └─ FAIL → Debug → Iteration 2 (fixContext = QA failures)

Iteration 2-3:
  Generator (debt) → QA
    ├─ PASS → Review
    │   └─ (同上)
    └─ FAIL → Debug → Iteration 3

Iteration 4+:
  ESCALATE (超过最大轮次，升级给用户)
```

**最大轮次**:
- QA-Debug 循环: 3 轮
- Review 循环: 2 轮

---

## 5. Layer 3: 全局协调层

### 5.1 Owner Agent 角色

**定位**: 纯编排器，不亲自做需求分析和架构设计

**核心职责**:
1. **路径选择** - 收到需求后立即判断路径
2. **子 Agent 派发** - 需求分析/架构设计/编码 → 启动独立子 Agent
3. **产出验收** - 读产出文件摘要做验收
4. **Go/No-Go 裁决** - HITL 确认点暂停
5. **Proto 变更** - 硬性规则，由 Owner 亲自执行
6. **质量把关** - 确保每个变更走完 QA + Review
7. **文档与知识维护** - 代码变更 → CHANGELOG，新坑 → memory/
8. **任务与 BACKLOG** - QA/Review 问题 → tasks/

### 5.2 路径选择规则（dispatch 统一入口 + 工作量分级）

**收到任何改动需求后，必须走 dispatch 入口做 S/M/L 工作量分级** (硬性第一步，CLAUDE.md 约束 #7)。

| 分级 | 判定条件 | 执行方式 | QA | Review |
|------|---------|---------|:---:|:---:|
| **S（轻量）** | 单服务单文件 ≤20 行，不涉及 Proto/common，不新增公开 API | 轻量 Pipeline（`workload:"S"`） | ✅ 18 项 | ❌ 跳过 |
| **M（单服务）** | 单服务代码改动，非 S 非 L | Pipeline（默认全流程） | ✅ 18 项 | 按 taskType |
| **L（跨服务）** | 跨 2+ 服务 / 涉及 Proto/common / 新增公开 API / 架构决策 | OpenSpec → N×Pipeline | ✅ 每服务 | ✅ 每服务 3 视角 |
| **跳过** | 纯文案/注释/配置值，无需编译验证 | Edit → build | ❌ | ❌ |

> 原「直接 Edit / Dev Agent」路径已废弃——它们绕过 Pipeline，导致未 QA 的代码直达用户。

### 5.3 OpenSpec 六阶段流程

| # | 阶段 | 执行方式 | 产出 | 门禁 |
|---|------|:---:|------|------|
| 0 | **工具选择** | Owner 内联 | `request.md` | 选对工具 |
| 1 | **需求分析** | 子 Agent (requirement-analyst) | `proposal.md` + `specs/*.md` | 追溯表全✅ + Self-Review PASS |
| 2 | **需求评审** | 3 子 Agent 并行 (coverage/structure/clarity) | `review/spec_review_*.md` ×3 | 2/3 APPROVED |
| 3 | **架构设计** | 子 Agent (architecture-designer) | `design.md` + `tasks.md` | 记忆注入 + 零占位符 + TDD 步骤 |
| 4 | **Proto 变更** | Owner 内联 | `api-proto/` + make ci | lint+breaking 全过 |
| 5 | **编码+测试** | N×Workflow 并行 (harness-pipeline.js) | 代码 + `_qa.md` + `_review.md` | 每服务 QA PASS + Review 2/3 PASS |
| 6 | **集成归档** | Owner 内联 | 移动 QA/Review 到 impl/ + INDEX + summary | 全链路通过 |

### 5.3.1 双层流水线总览（命名统一）

**当前流水线分两层，命名全仓统一**：

| 层 | 名称 | 脚本 | 职责 |
|----|------|------|------|
| **全流程自动化** | `spec-pipeline` | `harness-spec-pipeline.js` | 0-6 阶段自动编排 + 每阶段 HITL 暂停 |
| **编码流水线** | `harness-pipeline` | `harness-pipeline.js` | 阶段 5 编码：Generator → QA → (Debug) → Reviewer |

- **spec-pipeline**：规范驱动全流程（dispatch → 需求分析 → 评审 → 架构 → Proto → 编码 → 归档），每阶段末 `need_input` 暂停等用户，`resumeFromRunId` 续跑。
- **harness-pipeline**：spec-pipeline 阶段 5 委托的编码流水线（每服务一个，QA 18 项 + Review）。
- 旧概念「广义 OpenSpec vs 狭义 Pipeline」已由「spec-pipeline vs harness-pipeline」取代（更精确）。

**需求分析、设计是 spec-pipeline 的阶段 1-3，由它自动编排（子 Agent 在干净上下文执行）**。

### 5.3.1.1 全流程自动化（spec-pipeline）

> **`harness-spec-pipeline.js` 将广义全流程自动化编排**（2026-08-14 真实 L 级端到端验证通过）。

- **输入**：`Workflow({scriptPath:".harness/workflows/harness-spec-pipeline.js", args:{change, task}})`。
- **编排**：0 dispatch → 1 需求分析（澄清+分析）→ 2 需求评审 → 3 架构设计 → 4 Proto → 5 编码（委托 harness-pipeline.js）→ 6 集成归档。
- **HITL**：每阶段末 `need_input` 暂停，Owner 用 AskUserQuestion 问用户，`resumeFromRunId + resumeState(ctx) + resumeWith(decisions)` 续跑（SOP 见 owner-agent.md「如何 resume」）。
- **门禁**：沙箱内纯逻辑门禁（评审投票/轮次）；文件级门禁由子 Agent 完整环境跑 harness-checks + Owner 事后验证（沙箱无 fs 的架构限制）。
- **状态**：`pipeline-state.json` 存 ctx（沙箱无 fs 实际经 args.resumeState 传回）。
- **验证**：真实需求（角色列表排序）完整跑通 0→6，修复 6+ bug（门禁 fs/proto_done/svcName/反引号等）。

### 5.3.2 各阶段工具与技能

| 阶段 | 工具/子 Agent | 技能 | 产出 |
|------|--------------|------|------|
| 需求分析 | `requirement-analyst` 子 Agent | `superpowers:brainstorming`（交互澄清）+ `AskUserQuestion`（结构化提问） | `proposal.md` + `specs/*/spec.md` + `.change.yaml` |
| 需求评审 | 3 Reviewer 子 Agent 并行（coverage/structure/clarity） | `review.md` 模式一 | `review/spec_review_*.md` ×3 |
| 架构设计 | `architecture-designer` 子 Agent | `superpowers:writing-plans`（bite-sized 任务拆分） | `design.md` + `tasks.md` |
| 编码+测试 | 狭义 Pipeline（`harness-pipeline.js`） | Generator/QA/Debug/Review prompt | 代码 + `_qa.md` + `_review.md` |
| Proto 变更 | Owner 内联（api-proto + make ci） | — | proto + 生成代码 |

**⚠️ 硬性约束**：需求分析/设计必须由子 Agent 在干净上下文执行，Owner 只验收产出摘要——禁止在 Owner 主对话里 inline 拍脑袋出方案（绕过子 Agent 隔离会污染上下文、丢失验收门禁）。

**调度策略**:
```
Proto 变更 (Owner, 先做)
  │
  ├─ 并行组 1: 无依赖的微服务 (同时启动)
  │   Workflow({serviceDir: "services/moderation-service", ...})
  │   Workflow({serviceDir: "services/community-hub-service", ...})
  │
  ├─ 并行组 2: 前端 (与后端无依赖，可与组1同时)
  │   Workflow({serviceDir: "web/pc", ...})
  │   Workflow({serviceDir: "web/mobile", ...})
  │
  └─ 依赖服务 (等上游完成后)
      Workflow({serviceDir: "services/user-service", ...})
```

**任务提取**: 从 `tasks.md` 按服务分组，每个 Workflow 只传属于自己的 task 描述

**回退传播**: 上游服务 FAIL → 依赖它的下游服务等待修复后重试

### 5.5 HITL 置信度自适应审查

Pipeline 返回的 `confidence` 评分 (0.0-1.0) 基于：
- 迭代次数
- Review 一致性
- Memory 匹配数
- QA 一次性通过率

**审查深度规则**:

| 置信度 | 审查深度 | 操作 |
|:---:|---|------|
| ≥ 0.80 | 摘要审查 | 读 QA summary + review summary，确认无异常 |
| 0.50–0.79 | 抽查 | 随机抽取 max(2, totalFiles×30%) 个变更文件全文阅读 |
| < 0.50 | 全文审查 | 阅读全部变更文件，建议暂停并要求人工确认 |

---

## 6. 工具链与依赖

### 6.1 核心脚本

| 脚本 | 用途 | 调用方 |
|------|------|--------|
| `harness-checks.sh` | Go 服务机械化检查 | QA Agent / Owner |
| `harness-checks-frontend.sh` | 前端服务机械化检查 | QA Agent / Owner |
| `check-proto-ts-align.sh` | Proto→TS 对齐检查 | harness-checks.sh (check #11) |
| `harness-pipeline.js` | 工作流编排脚本 | Owner (Workflow 工具) |
| `gate-engine.js` | QA/Review/verify 门禁（validateGate） | core.js 阶段 5 |
| `harness-smoke.sh` | 运行时冒烟测试 | Owner (阶段 6) |
| `memory-index-build.sh` | 记忆索引构建 | Generator (编码前) |
| `memory-index-query.sh` | 记忆索引查询 | Generator (编码前) |
| `graph-sync.sh` | 知识图谱同步 | Owner (定期 / QA 提示) |

### 6.2 外部依赖

**必需工具**:
- Go 1.21+
- Node.js 18+ (前端服务)
- git
- bash 4.0+
- jq (JSON 解析)

**可选工具**:
- protoc + buf (Proto 变更时)
- neo4j (知识图谱同步)
- docker + docker-compose (中间件)

### 6.3 集成点

```
┌──────────────┐
│  git hooks   │ (pre-commit: harness-checks.sh)
└──────┬───────┘
       │
┌──────▼───────┐
│  Workflow    │ harness-pipeline.js
└──────┬───────┘
       │
┌──────▼───────┐
│  CI/CD       │ (GitHub Actions / GitLab CI)
└──────────────┘
```

---

## 7. 性能与扩展性

### 7.1 性能指标

| 指标 | 目标 | 实际 |
|------|:---:|:---:|
| 增量检查 (Go) | <10s | ~5-8s |
| 增量检查 (前端) | <30s | ~15-25s |
| 全量检查 (Go) | <60s | ~30-50s |
| Pipeline 完整循环 (单服务) | <10min | ~5-8min |
| 跨服务并行调度 (3 服务) | <15min | ~10-12min |

### 7.2 扩展点

**新增机械化检查**:
1. 在 `harness-checks.sh` 添加 `check_<name>()` 函数
2. 在 `main()` 中调用
3. 更新 `SKILL.md` 中的检查项表格
4. 更新本文档

**新增 Review 维度**:
1. 在 `review.md` 添加维度定义
2. 在 `harness-pipeline.js` 的 `REVIEW_LENSES` 添加视角
3. 调整 `reviewLensPrompt()` 分工表

**新增 Agent 类型**:
1. 在 `.harness/agents/subagents/` 添加 Agent 定义
2. 在 `owner-agent.md` 更新调度表
3. 在 `harness-pipeline.js` 添加 prompt 函数 (如需)

---

## 8. 质量度量

### 8.1 流水线健康指标

| 指标 | 计算方式 | 目标 |
|------|---------|:---:|
| **QA 一次通过率** | PASS / (PASS + FAIL) | ≥80% |
| **Review 通过率** | 2/3 PASS / total | ≥70% |
| **平均迭代次数** | sum(iterations) / total | ≤2.0 |
| **Memory 命中率** | hits / searches | ≥50% |
| **TDD 证据完整率** | with_RED_evidence / new_functions | ≥90% |

### 8.2 代码质量指标

| 指标 | 来源 | 目标 |
|------|------|:---:|
| **机械化检查 PASS 率** | harness-checks.sh | 100% |
| **测试覆盖率** | go test -cover | ≥60% |
| **CRITICAL 问题密度** | Review / KLOC | <1.0 |
| **记忆应用率** | applied / total | ≥60% |

---

## 9. 已知限制与改进方向

### 9.1 当前限制

1. **前端检查覆盖不足** - 仅 6 项，缺少 E2E 测试、a11y 检查
2. **Benchmark 回归非阻塞** - 性能退化仅 WARN，未强制回退
3. **API 冒烟测试需运行时** - 服务未启动时跳过
4. **Memory 索引手动构建** - 未自动化，需定期运行脚本
5. **跨服务依赖分析手动** - Owner 需人工判断并行组

### 9.2 改进方向

**短期** (1-2 周):
- [ ] 前端 E2E 测试集成 (Playwright)
- [ ] Memory 索引自动构建 (pre-commit hook)
- [ ] API 冒烟测试增强 (自动启动服务)

**中期** (1-2 月):
- [ ] Benchmark 回归阻塞策略 (>50% 退化 FAIL)
- [ ] 跨服务依赖自动分析 (从 design.md 提取)
- [ ] Review 维度权重配置化

**长期** (3-6 月):
- [ ] Pipeline 性能可视化看板
- [ ] 质量趋势分析 (周/月报告)
- [ ] 自适应检查 (根据变更类型调整检查项)

---

## 10. 参考资源

| 资源 | 路径 |
|------|------|
| QA 技能定义 | `.harness/skills/qa/SKILL.md` |
| Review 技能定义 | `.harness/skills/review.md` |
| Owner Agent 调度规则 | `.harness/agents/owner-agent.md` |
| 编码规范 | `.harness/rules/项目编码规范.md` |
| Proto 管理规范 | `.harness/rules/Proto管理规范.md` |
| 工具使用规则 | `.harness/docs/tool-usage-rules.md` |
| Pipeline 演进日志 | `.harness/docs/pipeline-evolution.md` |
| Pipeline 最佳实践 | `.harness/docs/pipeline-patterns.md` |

---

**文档维护**: 重大架构变更时更新本文档，记录变更原因到 `pipeline-evolution.md`
