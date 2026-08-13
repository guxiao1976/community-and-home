# 流水线检视与完善 skill 设计

> 日期：2026-08-13
> 状态：设计定稿（待评审）
> 作者：全局架构协调层（Owner）

## 1. 背景与目的

数据权限 Wave 开发期间，暴露出开发流水线本身的三个缺口：

1. **应用不严**：S 级 / 完善任务没有强制走流水线，执行者（含 Owner）倾向于「这太小了，内联就行」而绕过。
2. **文档滞后**：流水线设计文档（`pipeline-architecture/evolution/flow/patterns`）最后更新停在 2026-06-22，之后 7-8 月的演进（dispatch 分级、TDD 证据校验、Incident 机制、graph_freshness 修复等）均未写回。
3. **进化空转**：Incident 记录仅 3 条（阈值 5 条），`pipeline-evolution-v2` 仍 pending，`evolve-pipeline.sh` 未真正运转，「检视→进化」闭环空转。

本设计的目标：把「检视流水线」从一个临时动作固化为**标准化、可重复、可度量**的 skill，并建立检视所需的「原则标尺」。

## 2. 产物

两个产物，一个 skill 一个标尺：

| 产物 | 位置 | 作用 |
|---|---|---|
| `harness-design-principles.md` | `.harness/docs/` | **检视标尺**——把散落在 owner-agent.md / CLAUDE.md / pipeline-architecture.md 的原则收拢成权威清单，含「环节 → skills/tools 映射」 |
| `pipeline-review` skill | `.harness/skills/pipeline-review.md` | **检视流程**——用标尺 + 5 个维度做检视，含测试步骤和判定标准 |

**原则文档除 16 条原则外，还包含一个「环节 → skills/tools 映射」章节**，收拢当前散落在 owner-agent.md（阶段表六元组）和 dispatch.md（S/M/L 路由）两处的映射，作为权威操作手册：

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

## 3. 16 条设计原则（检视标尺）

前 10 条从现有文档提炼，第 11-14 条来自业界 harness / loop engineering 最佳实践，第 15-16 条为本次设计新增。

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

## 4. 检视维度与判定标准

skill 的「检视步骤」逐项跑 5 个维度，每个维度有可度量判定标准。

### 4.1 文档新鲜度（相对判定，非绝对阈值）

**目的**：确保 harness 文档**准确反映当前实现**，而非「常更新」。

**范围**：仅 harness 自身文档（`pipeline-*.md`、`harness-design-principles.md`、`owner-agent.md`、各 skill 的 `SKILL.md`、`docs/INDEX.md`）。**不检视**工程文档（服务 design.md、specs、CHANGELOG、graph-context）。

**判定**（相对，非「30 天」绝对阈值）：

| 条件 | 结论 |
|---|---|
| `文档最后更新时间 ≥ 最近流水线代码改动时间` | ✅ 准确 |
| `文档最后更新时间 < 最近流水线代码改动时间` | ❌ 滞后（代码改了，文档没跟上） |

流水线稳定长期不改时，文档未更新属「准确」而非「滞后」，不误报。

### 4.2 应用率

**检查**：抽查最近 N 个已完成开发/完善任务，是否走了 dispatch 分级 + pipeline。

**判定**：应用率 ≥ 90%；S 级允许内联但必须有 dispatch 分级记录 + 门禁。

### 4.3 门禁健康（4 个子检查）

| 子检查 | 方法 | 判定 |
|---|---|---|
| 机械化检查 | 各服务跑 `harness-checks` | 0 FAIL；WARN 有记录且递减 |
| 配置漂移 | 对比 `quality-gates.yml` 规则数 vs `gate-engine.js` 执行数 | 无「定义了没执行」漂移（或明确标注未接入） |
| pre-commit 生效 | 确认 hook 挂载 + 抽查拦截 | hook 挂载且有效 |
| 门禁日志 | 查 `.harness/logs/gates/` 更新时间 | 日志在记录（有留痕） |

### 4.4 进化机制运转

**检查**：`logs/incidents/` 条数 + 处理状态 + evolve-pipeline 是否达阈值。

**判定**：Incident 处理率 100%；达阈值（≥5 条）则已触发进化。

### 4.5 原则符合性（元检视）

**检查**：对照 16 条原则 + 环节→skills/tools 映射，抽查近期 devlog / changes / commit，确认各环节用了正确的 skill/tool。

**判定**：无违反（原则 + 映射一致）；违反已记录为 Incident。

## 5. 测试步骤

「测试」分正向和负向两半。

**正向 —— 验链路通畅**：跑一个最小、无害、可逆的样例任务（如给某服务 model 加一行 `// Deprecated`），用 `harness-pipeline.js` 走 Generator → QA → Reviewer，验证不卡死、产出 PASS。

**负向 —— 验门禁承重**：故意注入 FAIL（如改坏一个测试），验证 QA 门禁正确拦截（返回 FAIL，非假装成功）。

**判定**：正向全链路走通 + 负向被门禁正确拦截，才算「流水线功能自测」通过。

## 6. 闭环

```
检视 + 测试
   ├─ 产出报告：.harness/docs/pipeline-review-report-<日期>.md
   ├─ 每个问题 → harness-tasks.sh 建任务回 BACKLOG
   ├─ Incident 达阈值（≥5 条）→ 触发 evolve-pipeline.sh
   └─ 报告归档，作为下次检视基线
```

## 7. 触发方式

- **自动**：复用 `harness-loop` 机制，定期（建议每周）触发。
- **手动**：随时 `/pipeline-review`。

## 8. 实现范围（后续 writing-plans 阶段细化）

本设计只定「做什么、标准是什么」，不展开实现细节。实现阶段需确定：

1. `harness-design-principles.md` 的精确章节结构。
2. `pipeline-review` skill 的精确步骤序列（含每个维度对应的命令/脚本）。
3. 样例任务的选型（哪个服务、什么最小改动最安全）。
4. 报告模板。
