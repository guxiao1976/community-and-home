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

## 四、流程衔接规则（harness-pipeline vs superpowers）

按「产物类型 + 需求明确度」选择流程：

| 任务类型 | 需求 | 产物 | 流程 |
|---|---|---|---|
| 业务功能开发 | 明确（有 spec/tasks） | Go/proto/前端代码 | `harness-pipeline.js`（dispatch → S/M/L → Workflow/OpenSpec） |
| harness 自身开发 | 需探索（无 spec） | 文档/skill/配置 | superpowers（brainstorming → writing-plans → 执行） |
| 运维收尾 | 明确 | 小改动/补测试/归档 | Owner 内联 + 门禁（harness-checks） |

**边界**：harness 自身开发若需求已明确（如「修一行正则」），也可走 dispatch S 级；业务功能若需探索需求，先 brainstorm 再走 harness-pipeline。核心判据：**产物是否业务代码 + 需求是否已有 spec**。
