# 流水线改造计划：对齐设计文档

> 目标：让现有流水线符合 `harness-design-principles.md`（16 条原则 + 目录三层规范 + 引擎与策略分离）
> 规模：L 级 · 分 5 阶段 9 任务 · 每任务独立可验收
> 日期：2026-08-13

## 背景

首次 pipeline-review 检视发现流水线与设计文档存在 5 类 gap：服务名硬编码、文档目录混乱、配置漂移、脚本冗余、文档滞后。本计划按「低风险优先、依赖先行」拆解整改。

## 执行原则

1. 每任务独立提交，可独立回滚
2. 改造前先备份（改动面大的任务）
3. 引擎改动优先走「设计先行」：先更新规则文档，再改代码
4. 迁移类任务用 `git mv` 保留历史

---

## 阶段 1：文档目录重组（迁移）

### 任务 1.1：docs/ 分类梳理

- **目标**：把 `docs/` 43 个文件分为「活跃核心」与「历史存档」两类。
- **改动**：产出分类清单（不改动文件）。
- **验收**：清单覆盖全部 43 文件，无遗漏、无重复。
- **依赖**：无

### 任务 1.2：历史存档迁移到 `docs/_archive/`

- **目标**：把历史存档移出活跃区，`docs/` 只留活跃文档。
- **改动**：
  - 迁移：`PHASE*-COMPLETE.md`、`FINAL-*-SUMMARY.md`、`ULTIMATE-SUMMARY.md`、`problem-*`（8 个）、`p0-improvements-*`、`*-report.md`（qa/regression/integration/comprehensive/final）、`phase1-implementation-summary.md`、`MOCK-SETUP-GUIDE.md`、`before-after-visualization.md`、`harness-engineering-comparison.md`、`improvement-summary.md`、`skills-evaluation-and-integration.md`、`anthropic-feature-request-circuit-breaker.md`、`claude-md-addition-circuit-breaker.md`
  - 保留：`pipeline-*.md`（4）、`harness-design-principles.md`、`pipeline-review-report-*.md`、`README.md`、`tool-usage-rules.md`、`circuit-breaker.md`、`circuit-breaker-integration.md`、`change-directory-structure.md`
- **验收**：`docs/` 仅剩活跃文档（≤15 个），`docs/_archive/` 含全部历史文档，`git mv` 保留历史。
- **依赖**：任务 1.1

### 任务 1.3：更新 README 索引

- **目标**：`README.md` 反映重组后的目录，历史存档标注 `_archive/`。
- **验收**：README 索引与实际文件一致，无死链。
- **依赖**：任务 1.2

---

## 阶段 2：引擎与策略分离（原则 16 落地）

### 任务 2.1：服务名映射外提到 registry

- **目标**：消除 `harness-pipeline.js` / `harness-pipeline-core.js` 硬编码的中文服务名映射，改为从 `registry/services.json` 读。
- **改动**：
  1. 修 `registry/services.json` 的 `displayName`（当前错填 "CLAUDE.md"），补全正确中文名
  2. 在 core.js 加 `loadServiceDisplayName()`（读 registry，内存缓存）
  3. 删除 harness-pipeline.js:626 和 core.js:99 的硬编码映射对象，替换为函数调用
- **验收**：`grep -rn "'community-hub-service': '社区枢纽服务'" .harness/workflows/` 返回 0；pipeline 跑通（正向样例）。
- **依赖**：无（与阶段 1 可并行）
- **风险**：改 core.js 影响所有 pipeline 调用，需正向测试验证。

### 任务 2.2：harness-checks.sh 项目特定检查标注为策略

- **目标**：明确 `harness-checks.sh` 里 Snowflake/jstype/5位错误码/跨服务 DB 检查是「项目策略」，而非「通用引擎」。
- **改动**：在 harness-checks.sh 头部加注释标注「以下检查为 community 项目特有策略（Snowflake ID / 错误码 / 跨服务），迁移到其他项目需替换为对应策略」，并在 `config/` 下建 `project-policies.md` 说明这些策略的归属。
- **验收**：注释 + `config/project-policies.md` 存在，说明清晰。
- **依赖**：无

---

## 阶段 3：配置漂移消除

### 任务 3.1：quality-gates.yml 收敛

- **目标**：消除「63 条规则 vs 3 个 gate 函数」的配置漂移。
- **改动**：
  1. 盘点 quality-gates.yml 每条规则：哪些被 gate-engine.js 执行、哪些未接入
  2. 未接入的规则：明确标注 `status: not-implemented` + 原因（或删除，若确认不需要）
  3. 在 gate-engine.js 头部同步更新「规则映射」注释，使 yml 与 js 一一对应
- **验收**：`grep -c "not-implemented" quality-gates.yml` 与未接入规则数一致；gate-engine.js 注释准确。
- **依赖**：无

---

## 阶段 4：脚本清理

### 任务 4.1：废弃/重复脚本归档

- **目标**：清理 `scripts/` 里废弃、重复的脚本。
- **改动**：归档以下脚本到 `scripts/_archive/`（确认无引用后）：
  - `build-pipeline.sh`、`build-pipeline-new.sh`（旧构建脚本）
  - `harness-gate-check.sh`（被 v2 取代）
  - `run-all-checks.sh`、`test-p0-improvements.sh`、`verify-skills-integration.sh`、`workflow-fallback.sh`、`complete-feature.sh`（历史任务脚本）
- **验收**：归档前 `grep -rn` 确认无脚本引用；归档后 `scripts/` 只剩活跃脚本。
- **依赖**：无

---

## 阶段 5：文档补写

### 任务 5.1：architecture/flow/patterns 补写 7-8 月演进

- **目标**：把 evolution.md 已记录的 Phase 12-17，同步到 architecture（架构变化）、flow（流程变化）、patterns（新模式）。
- **改动**：
  - `pipeline-architecture.md`：补 dispatch 分级、确定性验证层、检视机制到架构图/三层体系
  - `pipeline-flow-complete.md`：补 dispatch 入口、pipeline-review 检视流程
  - `pipeline-patterns.md`：补 Incident 模式、检视模式、引擎策略分离模式
- **验收**：三文档提到 dispatch/pipeline-review/Incident，与 evolution.md 一致。
- **依赖**：阶段 1（文档目录已重组）

---

## 执行顺序与依赖

```
阶段1(1.1→1.2→1.3)  阶段2(2.1,2.2)  阶段3(3.1)  阶段4(4.1)  阶段5(5.1)
      ↓                    ↓              ↓           ↓           ↓
   独立可先行          2.1 需正向测试    独立        独立       依赖阶段1
```

- 阶段 1、2、3、4 相互独立，可并行推进
- 阶段 5 依赖阶段 1（文档目录重组完成后再补写）
- 风险最高的是任务 2.1（改 core.js），需正向测试护航
