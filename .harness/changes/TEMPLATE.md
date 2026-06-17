# summary.md 模板

> 每个变更的 Single Source of Truth。创建变更时复制此模板到 `.harness/changes/<change-name>/summary.md`。
> 每个阶段完成后立即更新，确保全流程可追溯。

---

# <变更名称>

**创建时间**: YYYY-MM-DD
**Owner**: 
**状态**: 🟡 进行中 / ✅ 已完成 / ❌ 已废弃

## 阶段追踪

| # | 阶段 | 状态 | 评审轮次 | 结论 | 备注 |
|---|------|:---:|:---:|------|------|
| 0 | 工具选择 | ✅ | — | <工具名> | |
| 1 | 需求分析 | | — | | |
| 2 | 需求评审 | | /3 | | |
| 3 | 架构设计 | | — | | |
| 4 | Proto 变更 | | — | | 如不涉及填 N/A |
| 5 | 编码+测试 | | /2 | | |
| 6 | 集成验证 | | — | | |

## 关键决策

| 日期 | 决策 | 原因 |
|------|------|------|

## QA 摘要

| 服务 | QA 轮次 | 测试数 | 覆盖率 | 结论 |
|------|:---:|:---:|:---:|------|
| | | | | |

## Review 摘要

| 服务 | 审查轮次 | CRITICAL | WARNING | 结论 |
|------|:---:|:---:|:---:|------|
| | | | | |

## 例外 & 未解决问题

| 事项 | 严重度 | 处理方式 |
|------|:---:|------|

## 产物索引

| 类型 | 路径 |
|------|------|
| proposal | `.harness/changes/<name>/proposal.md` |
| specs | `.harness/changes/<name>/specs/*/spec.md` |
| design | `.harness/changes/<name>/design.md` |
| tasks | `.harness/changes/<name>/tasks.md` |
| 需求评审 | `.harness/changes/<name>/review/spec_review_v*.md` |
| QA | `services/<name>/_qa.md` |
| Review | `services/<name>/_review.md` (v1/v2/...) |
| CHANGELOG | 各服务 `CHANGELOG.md` |
