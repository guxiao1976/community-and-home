# .harness 目录清理计划

**审计日期**: 2026-06-19
**审计人**: Claude (Owner Agent)

---

## 📋 审计摘要

| 类别 | 总数 | 活跃 | 未使用 | 过期 | 待定 |
|------|:---:|:---:|:---:|:---:|:---:|
| Skills | 12 | 8 | 4 | 0 | 0 |
| Loop Runs | 21 | - | - | 18 | 0 |
| Improvement Plans | 18 | 2 | 0 | 1 | 15 |
| Changes | 2 | 1 | 0 | 1 | 0 |

---

## ❌ 建议删除（过期/未使用）

### 1. 未使用的 Skills（0 次引用）

#### ❌ `.harness/skills/adaptive-review.md`
- **状态**: 功能已集成到 `owner-agent.md` §HITL 置信度自适应审查
- **理由**: Owner Agent 内联执行，不需要独立 Skill
- **操作**: 删除

#### ❌ `.harness/skills/agent-serial-mode.md`
- **状态**: Workflow 降级方案，从未使用
- **理由**: Workflow 工具稳定，降级未触发过
- **操作**: 移至 `docs/archived/skills/` 备用

#### ❌ `.harness/skills/triage.md`
- **状态**: 问题分类器，被 `harness-tasks.sh` 的自动分类替代
- **理由**: 脚本化方案更可靠（Python + 规则引擎）
- **操作**: 删除

#### ❌ `.harness/skills/unit-test-write.md`
- **状态**: 测试编写指导，未被引用
- **理由**: TDD 步骤已内建到 `harness-pipeline.js` Generator
- **操作**: 移至 `docs/archived/skills/` 备用

---

### 2. 过期的 Loop Runs（保留最近 3 天）

#### ❌ `.harness/loop-runs/run-2026-06-16-*.md`（9 个文件）
- **理由**: 6 月 16 日的日志，已超过 3 天
- **操作**: 移至 `.harness/loop-runs/_archive/2026-06-16/`

#### ❌ `.harness/loop-runs/run-2026-06-17-*.md`（4 个文件）
- **理由**: 6 月 17 日的日志，已超过 2 天
- **操作**: 移至 `.harness/loop-runs/_archive/2026-06-17/`

#### ❌ `.harness/loop-runs/run-2026-06-18-*.md`（3 个文件）
- **理由**: 6 月 18 日的日志，保留最后 1 个
- **操作**: 移至 `.harness/loop-runs/_archive/2026-06-18/`

**保留**: 6 月 19 日的 5 个文件（当天日志）

---

### 3. 过期的 Changes

#### ❌ `.harness/changes/dry-run-2026-06-09.md`
- **状态**: 6 月 9 日的演练记录
- **理由**: 演练完成，已归档到 INDEX.md
- **操作**: 移至 `.harness/changes/_archive/`

#### ❌ `.harness/changes/moderation-integration-retro.md`
- **状态**: moderation-integration 的回顾记录
- **理由**: 回顾完成，移入对应变更目录
- **操作**: 移至 `.harness/changes/moderation-integration/retro.md`

#### ⚠️ `.harness/changes/moderation-integration/`
- **状态**: 只有 request.md + proposal.md，无 design.md/tasks.md
- **理由**: 变更未完成（卡在阶段 1）
- **操作**: 
  - 方案 A: 如果不再继续 → 移至 `_archive/`
  - 方案 B: 如果继续 → 补全 design.md/tasks.md

---

### 4. Improvement Plans 重复项

#### ❌ `.harness/improvement-plans/P0-memory-index.md`
- **状态**: 与 `P0-memory-index-PROGRESS.md` 重复
- **理由**: PROGRESS 版本是最新的工作版本
- **操作**: 删除旧版本

---

## ⚠️ 待评估（需人工确认）

### Improvement Plans 状态不明（15 个）

| 文件 | 状态推测 | 建议 |
|------|---------|------|
| `CONTINUOUS-WORK-MODE.md` | 概念设计？ | 评估是否已实现 |
| `P0-frontend-checks.md` | 规划中？ | 检查是否已完成（frontend-checks.sh 存在）|
| `P0-hitl-adaptive-review.md` | 规划中？ | 检查是否已完成（owner-agent.md 有 HITL）|
| `P0-hitl-summary-template.md` | 规划中？ | 检查 TEMPLATE.md 是否已同步 |
| `P1-runtime-loading.md` | 待实现 | 保留或移至 Backlog |
| `P1-task-scheduling.md` | 待实现 | 检查是否已被 harness-tasks.sh 覆盖 |
| `P1-workflow-fallback.md` | 待实现 | 检查 workflow-fallback.sh 状态 |
| `P2-cost-monitoring.md` | 待实现 | 低优先级，可移至 Backlog |
| `P2-e2e-testing.md` | 待实现 | 低优先级，可移至 Backlog |
| `P2-observability.md` | 待实现 | 低优先级，可移至 Backlog |
| `P2-overview.md` | 概览文档 | 保留 |

**操作**: 逐个审查，分类为：
- ✅ 已完成 → 重命名为 `*-COMPLETED.md`
- 🔄 进行中 → 重命名为 `*-PROGRESS.md`
- 📋 待做 → 转为 `tasks/` 任务
- ❌ 废弃 → 移至 `_archive/`

---

## ✅ 保留（活跃使用）

### Skills（8 个）
- ✅ `architect-design.md` — 3 次引用
- ✅ `dispatch.md` — 10 次引用
- ✅ `github.md` — 3 次引用
- ✅ `openspec-to-ralph.md` — 1 次引用
- ✅ `qa.md` — 90 次引用
- ✅ `requirement-analysis.md` — 4 次引用
- ✅ `review.md` — 57 次引用
- ✅ `select-tool.md` — 2 次引用

### 其他目录（保持现状）
- ✅ `agents/` — 活跃使用
- ✅ `rules/` — 全局规范，必须保留
- ✅ `scripts/` — 自动化脚本，活跃使用
- ✅ `workflows/` — harness-pipeline 核心
- ✅ `knowledge/` — 知识库，持续积累
- ✅ `tasks/` — 任务管理，活跃使用

---

## 📊 清理预期收益

| 指标 | 清理前 | 清理后 | 改善 |
|------|:---:|:---:|:---:|
| Skills 数量 | 12 | 8 | -33% |
| Loop Runs | 21 | 5 | -76% |
| Improvement Plans | 18 | 待定 | 待定 |
| .harness/ 大小 | ~1.2 MB | ~800 KB | -33% |

---

## 🚀 执行建议

### 阶段 1: 安全删除（立即执行）
```bash
# 1. 删除未使用的 Skills
rm .harness/skills/adaptive-review.md
rm .harness/skills/triage.md

# 2. 归档降级方案（备用）
mkdir -p docs/archived/skills
mv .harness/skills/agent-serial-mode.md docs/archived/skills/
mv .harness/skills/unit-test-write.md docs/archived/skills/

# 3. 归档旧 loop-runs
mkdir -p .harness/loop-runs/_archive/{2026-06-16,2026-06-17,2026-06-18}
mv .harness/loop-runs/run-2026-06-16-*.md .harness/loop-runs/_archive/2026-06-16/
mv .harness/loop-runs/run-2026-06-17-*.md .harness/loop-runs/_archive/2026-06-17/
mv .harness/loop-runs/run-2026-06-18-1*.md .harness/loop-runs/_archive/2026-06-18/

# 4. 整理 changes
mkdir -p .harness/changes/_archive
mv .harness/changes/dry-run-2026-06-09.md .harness/changes/_archive/
mv .harness/changes/moderation-integration-retro.md .harness/changes/moderation-integration/retro.md

# 5. 删除重复的 improvement-plans
rm .harness/improvement-plans/P0-memory-index.md
```

### 阶段 2: 人工评估（需确认）
```bash
# 逐个审查 improvement-plans 状态
for plan in .harness/improvement-plans/P*.md; do
    echo "审查: $plan"
    # 读取内容，判断状态
done
```

### 阶段 3: 建立自动清理机制
```bash
# 添加到 harness-loop.sh 的定期清理
# 自动归档 3 天前的 loop-runs
```

---

## 📝 维护建议

### 日常清理规则
1. **Loop Runs**: 保留最近 3 天，每周自动归档
2. **Improvement Plans**: 完成后立即重命名为 `*-COMPLETED.md`
3. **Changes**: 完成后立即写 summary.md，未完成超过 30 天归档
4. **Skills**: 6 个月未使用 → 移至 `docs/archived/`

### 定期审计（每月）
```bash
bash .harness/scripts/harness-audit.sh
```

输出未使用的资源清单，人工决策保留/删除。
