# 流水线检视与完善 skill 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 创建「流水线检视与完善」能力——一份权威原则标尺（harness-design-principles.md）+ 一个可重复执行的检视 skill（pipeline-review.md）。

**Architecture:** 两个纯 Markdown 产物，互为配套：原则文档是「检视标尺」（16 条原则 + 环节映射 + 目录规范），skill 是「检视流程」（用标尺跑 5 个维度 + 正向/负向测试 + 闭环）。不改任何 JS/shell 引擎代码。

**Tech Stack:** Markdown（`.harness/docs/` 与 `.harness/skills/`），验证用 `grep`/`bash`。

## Global Constraints

- 原则文档路径固定为 `.harness/docs/harness-design-principles.md`
- skill 路径固定为 `.harness/skills/pipeline-review.md`（与 CLAUDE.md 的 Skill 发现机制一致，frontmatter 需含 `name` 和 `description`）
- 本次只写文档，不修改 `workflows/`、`scripts/`、`skills/qa/` 等引擎代码（引擎与策略分离，见原则第 16 条）
- 16 条原则的编号、名称、核心表述必须与规格 `docs/superpowers/specs/2026-08-13-pipeline-review-skill-design.md` 完全一致
- 提交信息格式：`feat(harness): ...` + `Co-Authored-By: Claude <noreply@anthropic.com>`

---

### Task 1: 创建权威原则文档 `harness-design-principles.md`

**Files:**
- Create: `.harness/docs/harness-design-principles.md`

**Interfaces:**
- Consumes: 规格 `docs/superpowers/specs/2026-08-13-pipeline-review-skill-design.md` 第 3 节（16 条原则）、第 2 节（映射表 + 目录规范）
- Produces: 一份含 16 条原则、环节→skills/tools 映射、目录结构规范三部分的文档。Task 2 的 skill 引用它作为「标尺」；Task 3 的索引引用它的路径。

- [ ] **Step 1: 写原则文档**（完整内容如下）

````markdown
# Harness 设计原则

> 权威原则清单 · 流水线检视（pipeline-review skill）的标尺 · 收拢自 owner-agent.md / CLAUDE.md / pipeline-architecture.md

## 一、16 条设计原则

| # | 原则 | 核心 |
|---|---|---|
| 1 | 知识分层管理 | L1 常驻 / L2 阶段触发 / L3 按需查询，Just-enough Context |
| 2 | 按需加载 | 不全局扫描，按阶段触发 |
| 3 | 主 agent 只调度 | 不做具体事务，开发功能切到子 Claude |
| 4 | 各阶段文档交接 | HANDOFF / SESSION-START 文档 |
| 5 | 验证优先 | 任何状态声称前有 FRESH 证据 |
| 6 | TDD 强制 | 新功能先写失败测试 |
| 7 | 根因分析 | QA FAIL 触发，不做症状修复 |
| 8 | 多视角审查 | 安全/规范/业务三视角，结论取最小值 |
| 9 | 质量门禁强制 | QA FAIL 阻塞提交，Review CRITICAL 阻塞合并 |
| 10 | dispatch 分级强制 | 每个功能先走 dispatch（S/M/L → 轻量Pipeline/Pipeline/OpenSpec） |
| 11 | 确定性门禁优先于模型承诺 | 验证阶梯 L1-L5，L4（模型判断）不能假装 L1（确定性） |
| 12 | builder/checker 工具级硬隔离 | 检查 agent 物理上只有读权限，不能靠 prompt 约束 |
| 13 | 闭环验证 + 预算自主 + stall detection | 每个变更立即验证；cap 预算；相同失败指纹两次 = 停止+升级 |
| 14 | scope/evidence 守门 | 交付前检查真实 git diff 是否超出计划范围 |
| 15 | 设计先行（design-first） | 流水线自身改动，先更新设计文档再改代码 |
| 16 | 引擎与策略分离（engine vs policy） | 项目独有内容放配置（yml/json/md），禁止硬编码进引擎（js/sh） |

## 二、环节 → skills/tools 映射

| 环节 | skill | tool |
|---|---|---|
| 入口分级 | `dispatch.md` | Skill 工具 |
| 工具选择 | `select-tool.md` | Owner 内联 |
| 需求分析 | `requirement-analyst` | Agent（子 agent） |
| 需求评审 | `review.md` | Agent（3 子 agent 并行） |
| 架构设计 | `architecture-designer` | Agent（子 agent） |
| 编码+测试 | `harness-pipeline.js` | Workflow（N×并行） |
| QA / Review | qa / review skill | Workflow 内部 |
| 集成验证 | `pipeline-flow-complete.md` | 内联 + 脚本 |

## 三、harness 目录结构规范（三层）

| 层 | 目录 | 职责 |
|---|---|---|
| 引擎层（通用，可移植） | `workflows/`、`scripts/`、`skills/`、`linters/`、`validators/`、`templates/`、`tools/`、`tests/`、`agents/` | 只放通用逻辑，禁止硬编码项目特定内容 |
| 策略层（项目独有） | `config/`、`registry/`、`rules/`、`knowledge/`、`changes/`、`tasks/` | 所有项目独有的规则、配置、知识 |
| 运行时层（运行产物） | `logs/`、`loop-runs/`、`backups/` | 运行产生，明确 gitignore 边界 |

**现状缺口**：`scripts/`（harness-checks 硬编码 Snowflake）、`skills/`（dispatch 硬编码服务名）、`agents/`（owner-agent 混入项目职责）需逐步把项目策略外提到 `config/` / `registry/` / `rules/`。
````

- [ ] **Step 2: 验证原则文档结构完整**

Run:
```bash
grep -c '^| [0-9]' .harness/docs/harness-design-principles.md   # 期望 16（16 条原则）
grep -c '^| 入口分级' .harness/docs/harness-design-principles.md # 期望 1（映射表存在）
grep -c '引擎层（通用' .harness/docs/harness-design-principles.md # 期望 1（目录规范存在）
```
Expected: 依次输出 `16`、`1`、`1`。

- [ ] **Step 3: 提交**

```bash
git add .harness/docs/harness-design-principles.md
git commit -m "feat(harness): 新增权威设计原则文档（16 条原则 + 环节映射 + 目录规范）

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

### Task 2: 创建检视 skill `pipeline-review.md`

**Files:**
- Create: `.harness/skills/pipeline-review.md`

**Interfaces:**
- Consumes: Task 1 产出的 `.harness/docs/harness-design-principles.md`（作为检视标尺）
- Produces: 一个可被 Skill 工具加载的 skill，定义 5 个检视维度 + 正向/负向测试 + 闭环。执行方式为 Owner 内联跑命令，不派发子 agent。

- [ ] **Step 1: 写 skill 文件**（完整内容如下）

````markdown
---
name: pipeline-review
description: 流水线检视与完善 —— 定期或按需检查开发流水线自身健康度（文档新鲜度、应用率、门禁健康、进化机制、原则符合性），跑样例验链路，产出报告并将问题回 BACKLOG。
---

# 流水线检视与完善

> 用 `.harness/docs/harness-design-principles.md` 作为标尺，逐项检查流水线自身健康。触发方式：定期（harness-loop，建议每周）或手动 `/pipeline-review`。

## Step 1: 加载标尺

读 `.harness/docs/harness-design-principles.md`（16 条原则 + 环节映射 + 目录规范）。

## Step 2: 检视 5 个维度

### 维度 1 · 文档新鲜度（相对判定，非绝对阈值）

```bash
# 最近流水线代码改动时间（workflows/scripts/skills 的最近 commit）
code_ts=$(git log -1 --format=%ct -- .harness/workflows .harness/scripts .harness/skills)
# 各 harness 文档最后更新时间
for f in .harness/docs/pipeline-*.md .harness/docs/harness-design-principles.md .harness/agents/owner-agent.md; do
  doc_ts=$(git log -1 --format=%ct -- "$f")
  [ "$doc_ts" -lt "$code_ts" ] && echo "滞后: $f"
done
```

判定：任何文档 `doc_ts < code_ts` → FAIL（代码改了文档没跟上）；无输出 → PASS。

### 维度 2 · 应用率

抽查 `.harness/tasks/` 最近 N 个已完成任务（completed），确认是否走了 dispatch 分级 + pipeline。判定：应用率 ≥ 90%；S 级内联需有 dispatch 分级记录 + 门禁。

### 维度 3 · 门禁健康（4 子项）

1. **机械化检查**：`bash .harness/skills/qa/scripts/harness-checks.sh --service <各服务>` → 0 FAIL；WARN 有记录且递减
2. **配置漂移**：对比 `config/quality-gates.yml` 规则数 vs `workflows/gate-engine.js` 执行数 → 无「定义了没执行」漂移
3. **pre-commit 生效**：确认 `.git/hooks/pre-commit` 挂载且能拦截
4. **门禁日志**：`ls -la .harness/logs/gates/` → 最近有记录

### 维度 4 · 进化机制运转

```bash
ls .harness/logs/incidents/*.yml | grep -v _template | wc -l   # Incident 条数
```

判定：Incident 处理率 100%；达阈值（≥5 条）则已触发 `evolve-pipeline.sh`。

### 维度 5 · 原则符合性（元检视）

对照 16 条原则 + 环节→skills/tools 映射，抽查近期 `docs/devlog/`、`.harness/changes/`、git log，确认各环节用了正确 skill/tool、无「主 agent 越界写代码」等违规。

## Step 3: 测试（正向 + 负向）

**正向**：跑一个最小无害样例任务（给某服务 model 加一行 `// Deprecated`），用 `harness-pipeline.js` 走 Generator→QA→Reviewer，验证不卡死、产出 PASS。

**负向**：故意改坏一个测试，验证 QA 门禁正确拦截（返回 FAIL 而非假装成功）。

判定：正向全链路走通 + 负向被门禁正确拦截。

## Step 4: 闭环

1. 产出报告 `.harness/docs/pipeline-review-report-<日期>.md`（记录 5 维度 + 测试结果 + 结论）
2. 每个问题 → `bash .harness/scripts/harness-tasks.sh create ...` 回 BACKLOG
3. Incident 达阈值 → `bash .harness/scripts/evolve-pipeline.sh`
4. 报告归档作为下次检视基线
````

- [ ] **Step 2: 验证 skill 结构完整**

Run:
```bash
grep -c '^### 维度' .harness/skills/pipeline-review.md          # 期望 5（5 个检视维度）
grep -c '^## Step' .harness/skills/pipeline-review.md           # 期望 4（加载标尺/检视/测试/闭环）
head -5 .harness/skills/pipeline-review.md | grep -q 'name: pipeline-review' && echo "frontmatter ok"
```
Expected: 依次输出 `5`、`4`、`frontmatter ok`。

- [ ] **Step 3: 提交**

```bash
git add .harness/skills/pipeline-review.md
git commit -m "feat(harness): 新增 pipeline-review 检视 skill

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

### Task 3: 更新文档索引并验证 skill 可发现

**Files:**
- Modify: `.harness/docs/README.md`（「架构与设计」表格加一行原则文档）

**Interfaces:**
- Consumes: Task 1 的 `.harness/docs/harness-design-principles.md`、Task 2 的 `.harness/skills/pipeline-review.md`
- Produces: README 索引指向两个新产物；确认 skill 能被 Skill 工具发现（name 与 description 合法）。

- [ ] **Step 1: 在 README「核心文档」表格加一行**

在 `.harness/docs/README.md` 的「核心文档（流水线体系）」表格末尾追加：

```markdown
| [harness-design-principles.md](./harness-design-principles.md) | 权威设计原则（16 条 + 环节映射 + 目录规范） | 新成员、检视者 |
```

- [ ] **Step 2: 验证索引指向正确**

Run:
```bash
grep -c 'harness-design-principles.md' .harness/docs/README.md   # 期望 ≥1
ls -la .harness/skills/pipeline-review.md                        # 期望文件存在
```
Expected: 第一条输出 ≥1，第二条列出 skill 文件。

- [ ] **Step 3: 提交**

```bash
git add .harness/docs/README.md
git commit -m "docs(harness): 索引加入原则文档与检视 skill

Co-Authored-By: Claude <noreply@anthropic.com>"
```
