# 流水线检视报告 2026-08-13

> 首次 pipeline-review 运行 · 检视标尺 `.harness/docs/harness-design-principles.md`

## 一、总体结论

流水线**基本健康**，但暴露了 4 类真实问题：文档滞后、应用率不足、门禁一处漏检、以及检视 skill 自身的两处缺陷。正向测试验证了 harness-pipeline.js 链路通畅。

## 二、5 维检视结果

### 维度 1 · 文档新鲜度 ❌
- ❌ 4 个 `pipeline-*.md`（architecture/evolution/flow/patterns）滞后——最后提交 07-10，7-8 月的演进（dispatch 分级、TDD 证据校验、Incident 机制、graph_freshness 修复）均未写回。
- ⚠️ **误报**：`harness-design-principles.md` 被误判滞后——它和 `pipeline-review` skill 是同一次交付（文档 09:50、代码 09:51），判定命令把同批产物也对比了。

### 维度 2 · 应用率 ❌
- 今日开发任务 #4（`a1adfa3` 审核可见性门禁）、#6（`6de00cc` 补单测）均 Owner 内联，未走 harness-pipeline.js Workflow，也未走 dispatch 分级入口。

### 维度 3 · 门禁健康 ⚠️
- ✅ pre-commit hook 已挂载（symlink 到 `.harness/scripts/git-hooks/pre-commit`）。
- ⚠️ 门禁日志 5 天未更新（`.harness/logs/gates/` 最近 08-08）。
- ❌ **hardcoded secrets 检查漏检**：`check_hardcoded_secrets` 的正则 `password\s*[:=]\s*"..."` 中 `[:=]` 只匹配单个 `:` 或 `=`，漏掉 Go 短变量声明 `password := "..."`。
- ⚠️ 配置漂移检查方法不精确：skill 里用 grep 函数名统计 gate-engine.js 实现数，但命名不统一导致无法准确对比 quality-gates.yml 规则数。

### 维度 4 · 进化机制运转 ⚠️
- Incident 3 条（08-10、08-11×2）< 5 条阈值，未触发 evolve-pipeline。

### 维度 5 · 原则符合性 ❌
- 违反原则 3（主 agent 只调度）：#4/#6 Owner 内联写代码。
- 违反原则 10（dispatch 分级强制）：#4/#6 未走 dispatch 入口。
- 违反原则 15（设计先行）：graph_freshness 修复（`85fe311`）直接改代码，未先更新文档。

## 三、测试结果

### 正向（验链路）✅
harness-pipeline.js 完整跑通：Generator → QA（16 PASS / 0 FAIL）→ Reviewer（3/3 PASS），1 轮通过。**链路通畅**。

### 负向（验门禁承重）✅ + ❌
- ✅ 注入违规后 go build/vet/test 正确拦截（13 PASS / 3 FAIL），门禁承重。
- ❌ hardcoded secrets 检查漏检 `:=` 模式（见维度 3）。

### 样例任务选型缺陷 ⚠️
正向测试选的「加 `// Deprecated:` 注释」样例不安全——`Deprecated:` 是 Go 机器可读标记（godoc 标灰、staticcheck SA1019），会误导活跃类型。Reviewer 捕获此问题并产出 3 条 memory 建议（合并为 1 条：禁止用 `// Deprecated:` 做自测/占位注释）。

## 四、待办问题清单（需回 BACKLOG）

| # | 问题 | 严重度 | 类型 |
|---|---|---|---|
| 1 | 4 个 pipeline-*.md 文档滞后（7-8 月演进未写回） | WARN | 文档债 |
| 2 | hardcoded secrets 检查漏检 `:=` 模式 | WARN | 门禁缺陷 |
| 3 | #4/#6 未走 pipeline（呼应 task-2026-08-13-001 衔接规则） | WARN | 应用率 |
| 4 | skill 维度 1 判定命令误报（同批产物对比） | WARN | skill 缺陷 |
| 5 | 样例任务选型：禁用 Deprecated 做自测标记 | INFO | 经验 |

## 五、结论

本次检视的**价值已显现**：skill 首次运行即抓出 4 类真实问题 + 自身 2 处缺陷 + 1 条安全经验。下一步：
1. 把 §四 的 5 条问题登记回 BACKLOG。
2. 落地 memory 建议（Deprecated 注释 pitfall）。
3. 维度 1 的 4 个文档滞后是最大的一块债务，建议优先补写。
