# .harness/ 目录说明

> **本目录 = Harness 开发自动化平台**。所有与流水线、门禁、知识、规范相关的文件都集中在这里。
> 架构总纲见 [`docs/harness-architecture.md`](docs/harness-architecture.md)；本文件只回答**「哪个文件夹放什么、坏了怎么清理」**。

## 📁 目录总览（14 个顶层目录）

| 文件夹                        | 职责                                   | 关键文件                                                                                            |
| -------------------------- | ------------------------------------ | ----------------------------------------------------------------------------------------------- |
| [`agents/`](agents/)       | **执行者定义**：Owner + 子 Agent + Prompt 源 | `owner-agent.md`（调度总纲）、`prompts/{generator,qa,review,debug}.js`（Prompt 源）                 |
| [`skills/`](skills/)       | **方法指导**：可复用流程技能                     | `dispatch.md`（统一入口）、`qa/`、`verify-before-deliver/`                                              |
| [`workflows/`](workflows/) | **流水线编排**（Workflow 脚本）               | `harness-pipeline.js`（**已构建产物**）、`harness-spec-pipeline.js`、`gate-engine.js`                    |
| [`scripts/`](scripts/)     | **工具脚本**（门禁/自检/知识加载）                 | `harness-checks.sh`（QA）、`harness-self-check.sh`（meta-CI）、`knowledge-load.sh`、`harness-tasks.sh` |
| [`config/`](config/)       | **机器可读配置**（yml）                      | `quality-gates.yml`、`deterministic-rules.yml`                                                   |
| [`rules/`](rules/)         | **人工硬性规范**（md）                       | `项目编码规范.md`、`Proto管理规范.md`、`工程结构.md`、`Git治理规范.md`                                               |
| [`registry/`](registry/)   | **服务注册**                             | `services.json`                                                                                 |
| [`knowledge/`](knowledge/) | **知识库**：记忆 + 架构/业务/图谱                | `memory/MEMORY.md`、`INDEX.md`、`business-flows.md`                                               |
| [`changes/`](changes/)     | **变更追溯**：OpenSpec 产物                 | `INDEX.md`、`<change-name>/`                                                                     |
| [`tasks/`](tasks/)         | **任务管理**                             | `BACKLOG.md`、`task-*.md`                                                                        |
| [`docs/`](docs/)           | **平台文档** + **统一归档家**                 | `harness-architecture.md`、`pipeline-architecture.md`、`_archive/`                                |
| [`logs/`](logs/)           | **运行日志**                             | `gates/*.json`、`incidents/*.yml`、`judgments/*.json`、`usage/`（脚本调用打点，gitignore）、`pipeline/`（metrics，gitignore）|
| [`loop-runs/`](loop-runs/) | **循环运行记录**（保留近 7 天）                  | `run-*.md`、`_archive/<日期>/`                                                                     |
| [`tools/`](tools/)         | **小型工具源码**（二进制不入库）                   | `go-ast-checker/`（main.go，脚本每次 `go build` 生成）                                                   |

> 根目录还有 `.graph_last_attempt` / `.graph_last_sync` —— 知识图谱运行时状态，自动维护，勿手动编辑。

## 🧭 边界约定（避免再混淆）

| 边界                     | 划分                                                                                                                                                |
| ---------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------- |
| **agents/ vs skills/** | `agents/` = 流水线里的**执行者**（Owner/Generator/QA/Reviewer + 他们的 Prompt）；`skills/` = **跨流水线可复用的方法技能**（dispatch/qa/review）。新增"谁在跑"→agents/；"怎么跑"→skills/ |
| **config/ vs rules/**  | `config/` = **机器执行**的 yml 配置（门禁阈值、确定性规则）；`rules/` = **人阅读/遵守**的 md 规范。策略文档归 rules/，机器参数归 config/                                                  |
| **scripts/ vs tools/** | `scripts/` = shell/单文件工具；`tools/` = 需要独立构建的源码模块（Go 等）                                                                                             |

## 🧩 重点结构说明

### 执行记录（流水线复盘数据）

**位置**：`.harness/logs/usage/` + `.harness/logs/pipeline/`（JSONL：每行一个 JSON，**gitignore 运行时产物**）

| 文件 | 记录内容 | 产生方 |
|------|---------|--------|
| `usage/knowledge-load.jsonl` | 知识检索：服务/关键词/命中记忆数 | knowledge-load.sh 调用时 |
| `usage/graph-query.jsonl` | 图谱查询：服务/成功或不可用 | graph-query.sh 调用时 |
| `usage/harness-checks.jsonl` | QA 门禁：服务/PASS/FAIL/WARN 数 | harness-checks.sh 运行时 |
| `pipeline/metrics.jsonl` | pipeline 指标：任务类型/轮次/置信度/状态 | harness-pipeline（非沙箱写文件，沙箱 console 输出 METRIC）|

**用途**：为流水线复盘提供第一手数据——哪个脚本被调用/频率、知识检索命中率、QA 通过率、pipeline 轮次分布。复盘时 `jq -s '.' <file>` 聚合分析。

> 打点由 `scripts/log-usage.sh` 统一写入；改 `logs/` 下其他记录（gates 门禁判定、incidents 事件、judgments 判断）仍入库。
>
> **开关**：`config/tracking.yml` 的 `enabled`（默认 `true`，开发流水线初期记录；成熟后置 `false` 关闭）或环境变量 `HARNESS_TRACKING=off` 临时关闭。

### skills/：自建 vs 第三方

`skills/` 下自建与第三方共存于同一目录；`.claude/skills/` 是指向它的**软链别名**（供 Skill 工具注册扫描），两路径同源、改一处全同步。**区分方式**：`ls -la` 权限位 `l` = 第三方软链，`-` = 自建文件。

**自建**（普通文件，git 跟踪）：

| 文件 | 用途 |
|------|------|
| `dispatch.md` | 开发任务统一入口 + S/M/L 分级路由 |
| `review.md` | 代码评审 |
| `qa.md` + `qa/`（SKILL.md + 6 脚本）| 机械化质量门禁（harness-checks）|
| `verify-before-deliver.md` + `verify-before-deliver/`（SKILL.md）| 交付前自验证 |
| `requirement-analysis.md` | 需求分析（spec-pipeline 阶段 1）|
| `architect-design.md` | 架构设计（spec-pipeline 阶段 3）|
| `select-tool.md` | 工具选择（owner-agent）|
| `pipeline-review.md` | 流水线自检 |
| `github.md` | GitHub 操作 |
| `openspec-to-ralph.md` | OpenSpec → spec 转换 |
| `_archive/` | 已归档模板（无引用）|

**第三方**（软链 → `.agents/skills/`，系统级加载，非项目维护）：
`canvas-design` `executing-plans` `frontend-design` `requesting-code-review` `theme-factory` `vercel-composition-patterns` `web-artifacts-builder` `web-design-guidelines` `webapp-testing` `writing-plans`

**维护规则**：
- 新增**自建** skill → 在 `skills/` 写 `.md`；需被 Skill 工具调用时用 `SKILL.md` 目录格式（如 `qa/`、`verify-before-deliver/`）
- 新增**第三方** skill → 放 `.agents/skills/`，在 `skills/` 建软链 `ln -s ../../.agents/skills/<name> <name>`，**禁止复制文件**
- **禁止**在 `.claude/skills/` 放物理文件——它是软链，物理文件会造成真冗余

### Prompt 构建体系（单一源）

`harness-pipeline.js` 由 `scripts/build-pipeline.sh` 从 `agents/prompts/{generator,qa,review,debug}.js` 打包生成（pre-commit 改动 prompts 时自动重建）。**改 Prompt 请编辑这 4 个 `.js` 后运行 `build-pipeline.sh`**。

> 2026-08-14 已收敛：原 `templates/*.md` + `*-new.js` + `build-pipeline-new.sh` 新轨契约与 core 断裂且从未接线（无任何调用点），整体归档至 `docs/_archive/prompts-new/`（git 历史可回溯），流水线回归单一 prompt 源，消除"误跑新构建生成坏流水线"风险。

```
agents/prompts/{generator,qa,review,debug}.js  ← Prompt 源（唯一）
scripts/build-pipeline.sh                      → 打包（pre-commit 触发）
harness-pipeline.js                            ← 已构建产物（Workflow 加载）
```

### 统一归档家（docs/_archive/）

**历史内容统一归档到 `docs/_archive/`，按原目录名分子目录**（`prompts/` `validators/` `tests/` `linters/` `templates/` `improvement-plans/` …），保留在 git 历史中，只是退出活跃视图。删除仅限：git-ignored 大体积备份、可再生的编译产物。

## 🧹 清理规则

| 内容                   | 保留策略                     | 归档位置                       |
| -------------------- | ------------------------ | -------------------------- |
| `loop-runs/run-*.md` | 保留近 **7 天**              | `loop-runs/_archive/<日期>/` |
| `tasks/task-*.md`    | 完成后即归档                   | `tasks/_archive/`          |
| `changes/` 一次性报告     | 不留在根目录                   | `changes/_archive/`        |
| `tools/*/` 编译产物      | **不入库**，脚本 `go build` 生成 | —                          |
| 弃用/死目录               | 无活跃引用即整体归档               | `docs/_archive/<原目录名>/`    |

## 🗑️ 清理记录

### 2026-08-14 · 目录级精简（顶层 20 → 15）

**归档到 `docs/_archive/`**（git mv，可回溯）：

- `workflows/prompts/` → `prompts/`（与 `agents/prompts/` 重复的旧副本）
- `validators/` → `validators/`（0 活跃引用）
- `tests/` → `tests/`（0 活跃引用）
- `linters/` → `linters/`（检查已硬编码进 harness-checks.sh，md 无引用）
- `templates/` → `templates/`（文档/模板/脚本混放的自包含死集群）
- `improvement-plans/` → `improvement-plans/`（空壳）
- `scripts/cleanup-harness.sh` → `scripts/_archive/`（6-19 清理的执行脚本，已消费）

### 2026-08-14 · 首次清理（释放 ~305MB）

**删除**：`backups/`（307M git-ignored 残片）、`go-ast-checker` 二进制 + 嵌套残留
**归档**：`changes/` 26 个散落报告、`loop-runs/` 28 个旧记录、`tasks/` 19 个旧任务、`skills/templates/` 7 个废弃模板、CLEANUP 旧文档

**验证**：`bash .harness/scripts/harness-self-check.sh` → 5 PASS / 0 FAIL。

---

**维护者**: Owner Agent
**一致性**: `harness-self-check.sh` 自动校验；目录结构变更时同步更新本文件 + `docs/harness-architecture.md` §5
