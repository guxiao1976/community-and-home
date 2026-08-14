# Harness 整体架构

> **harness = 开发自动化平台**（最高层设计文档）。描述整个系统的子系统、协作、自检与目录。
> 定位：本文件是 harness 架构的**唯一总纲**，替代散落的分子系统描述作为架构入口。

---

## 1. 系统总览

harness 是一个「用 AI 开发 AI 系统」的自动化平台，由 **6 大子系统** 协作构成：

```
                    ┌─────────────────────────────────────┐
                    │          Owner Agent（总调度员）      │
                    │  入口分级 · 启动流水线 · HITL 决策    │
                    └──────────────┬──────────────────────┘
                                   │ 触发
        ┌──────────────┬───────────┼────────────┬─────────────┐
        ▼              ▼           ▼            ▼             ▼
   ┌─────────┐   ┌──────────┐ ┌────────┐  ┌──────────┐  ┌──────────┐
   │开发流水线│   │ 知识管理  │ │上下文管理│ │ 门禁体系  │  │任务/变更  │
   │spec-    │   │ memory/  │ │ L1/L2/L3│ │ checks/  │  │ BACKLOG/ │
   │pipeline │   │ graph/   │ │ 分层    │ │ gate/    │  │ changes/ │
   │harness- │   │ CHANGELOG│ │         │ │ pre-commit│  │          │
   │pipeline │   └──────────┘ └────────┘  └──────────┘  └──────────┘
   └─────────┘                                             
            └─────────────── 自检机制（harness-self-check）────────┐
                                                                  ▼
                                            保证上述所有子系统一致、可用、高质量
```

---

## 2. 子系统

### 2.1 开发流水线（详参 pipeline-architecture.md）

| 层 | 脚本 | 职责 |
|----|------|------|
| **全流程自动化** | `harness-spec-pipeline.js` | 0-6 阶段自动编排 + 每阶段 HITL 暂停 |
| **编码流水线** | `harness-pipeline.js` | 阶段 5 编码：Generator → QA → (Debug) → Review |

### 2.2 知识管理

| 资源 | 路径 | 用途 |
|------|------|------|
| 经验记忆 | `.harness/knowledge/memory/MEMORY.md` | 踩坑经验（must/should-follow），`knowledge-load.sh` 注入流水线 |
| 知识图谱 | `.harness/knowledge/INDEX.md` + `graph-context.md` | 架构/业务/数据模型，Neo4j 自动生成 |
| 变更追溯 | `.harness/changes/INDEX.md` | 历史变更 + OpenSpec 产物 |
| 开发日志 | `docs/devlog/` | 每日记录 |

### 2.3 上下文管理

Owner Agent 三层加载（`owner-agent.md` §2）：
- **L1 会话常驻**：CLAUDE.md + owner-agent + 编码规范（≤600 行）
- **L2 阶段触发**：进入阶段才加载对应 skill（dispatch/review/requirement-analysis 等）
- **L3 按需查询**：knowledge/INDEX → design/graph-context，Agent 自主查阅

### 2.4 门禁体系

| 门禁 | 触发 | 查什么 |
|------|------|--------|
| `harness-checks.sh` | QA 阶段 | 服务代码 18 项（build/test/规范/安全） |
| `gate-engine.js` | 流水线内 | 阶段门禁（评审投票/轮次/verify） |
| `pre-commit` | 每次提交 | Memory 索引 + pipeline 重建 + **harness 自检** |
| `harness-self-check.sh` | 改动 harness 时 | **harness 自身一致性**（见 §4） |

### 2.5 任务/变更管理

- `BACKLOG.md`：待办单一数据源，`harness-tasks.sh` 管理
- `.harness/changes/<change>/`：每个变更的 OpenSpec 产物（proposal/spec/design/tasks/summary）

### 2.6 自检机制（meta-CI）

对 harness 自身的 CI（区别于 harness-checks 对服务代码）：
- **脚本**：`.harness/scripts/harness-self-check.sh`
- **触发**：pre-commit（改动 workflows/skills/scripts/agents 时）+ pipeline-review（定期）+ 手动
- **职责**：查「harness 内部一致」（见 §4）

---

## 3. 子系统协作

**数据流**（用户提需求 → 交付）：
```
用户需求
  → Owner 分级 → spec-pipeline（读 request.md/specs/tasks）
  → 子 Agent 写盘交接（proposal → design → tasks → 代码）
  → harness-pipeline 编码（QA 18 项 + Review）
  → 归档（summary + INDEX）
  全程知识注入（memory-load）+ 门禁校验（checks/gate）+ 上下文分层加载
```

**触发链**：
```
改动 harness 源码 → pre-commit → self-check（查引用/命名/文档/配置/调用链）
                     → 不通过则阻止提交（保证 harness 自身不腐化）
```

---

## 4. 自检与一致性基准

`harness-self-check.sh` 以本总纲 + 关键文档为「一致性基准」，查 5 类：

| 检查 | 基准 |
|------|------|
| 流程引用存在 | 关键脚本/文档路径有效 |
| 命名口径一致 | 无残留旧流程词（并行 N×Workflow/派发子Agent/Dev Agent） |
| 文档同步 | 核心文档均提到当前 spec-pipeline |
| 配置漂移 | quality-gates.yml 无 not-implemented 残留 |
| 调用链完整 | dispatch → spec-pipeline → harness-pipeline → gate-engine 契约在 |

---

## 5. 目录结构（.harness/ 全览）

> 每个目录的用途、边界约定与清理规则详见 [`.harness/README.md`](../../README.md)（目录说明）。

```
.harness/
├─ agents/          # 执行者定义：Owner + 子 Agent + Prompt 源（prompts/templates/）
├─ skills/          # 方法技能（dispatch/review/qa/requirement-analysis...）
├─ workflows/       # 流水线编排（spec-pipeline + harness-pipeline + gate-engine）
├─ scripts/         # 工具脚本（harness-checks/self-check/knowledge-load/graph-sync...）
├─ config/          # 机器配置（quality-gates.yml + deterministic-rules.yml）
├─ rules/           # 硬性规范（编码规范/Proto管理/工程结构/Git治理）
├─ registry/        # 服务注册（services.json）
├─ knowledge/       # 知识（memory/ + INDEX + business-flows）
├─ changes/         # 变更（每变更一个目录，含 OpenSpec 产物）
├─ tasks/           # 任务（BACKLOG + task-*.md）
├─ docs/            # 平台文档 + 统一归档家 _archive/（历史目录按原名分收）
├─ logs/            # 运行日志（gates/incidents/judgments）
├─ loop-runs/       # 循环运行记录（保留近 7 天）
└─ tools/           # 小型工具源码（二进制不入库，go build 生成）
```

> 2026-08-14 已收敛：`linters/` `validators/` `tests/` `templates/` `improvement-plans/` `workflows/prompts/` 均为无活跃引用的历史目录，整体归档至 `docs/_archive/`（详见 `.harness/README.md` 清理记录）。

---

## 6. 文档体系（本总纲与其他文档的关系）

| 文档 | 角色 | 关系 |
|------|------|------|
| **harness-architecture.md**（本文件） | **总纲** | harness 整体架构唯一入口 |
| harness-design-principles.md | 16 条设计原则 | 总纲的「设计价值观」 |
| pipeline-architecture.md | 开发流水线子系统设计 | 总纲 §2.1 的详参 |
| pipeline-flow-complete.md | 流水线流程 | 总纲 §3 数据流的详参 |
| pipeline-patterns.md / evolution.md | 模式/演进 | 子系统细节 |
| owner-agent.md | Owner 调度规则 | 总纲 §1 调度员 + §2.3 上下文 |
| docs/README.md | 文档索引 | 导航 |
| .harness/README.md | 目录说明 | §5 的详参（每目录用途 + 清理规则） |

> **替代关系**：本总纲替代「pipeline-architecture.md 作为架构入口」的角色（它降级为开发流水线子系统详参）。新增 harness 子系统（自检/知识/上下文）时，架构描述收敛到本总纲，不再散落。

---

**维护者**: Owner Agent
**一致性**: 由 harness-self-check.sh 自动校验（本文件是文档同步检查的基准之一）
**更新**: 子系统架构变化时更新对应详参 + 本总纲
