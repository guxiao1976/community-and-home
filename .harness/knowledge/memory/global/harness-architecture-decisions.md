---
triggers: ["harness", ".harness", "驾驭工程", "目录结构", "架构决策", "四支柱"]
service: all
type: decision
severity: should-follow
status: active
created: 2026-06-10
updated: 2026-06-10
last_applied: null
apply_count: 0
---

# Harness 架构决策记录

## 为什么会有这条经验

2026-06-10 基于 Anthropic/OpenAI 的 Harness Engineering 方法论，结合本项目的 8+ 微服务 monorepo 实际情况，从零搭建了 `.harness/` 体系。过程中做了多个关键架构决策，记录在此便于后续维护者理解"为什么这样设计"。

## 决策记录

### D1: .harness/ vs .claude/ 的分工

**决策**: `.claude/` 只放配置（settings.local.json），通过符号链接指向 `.harness/skills/`、`.harness/knowledge/memory/`、`.harness/workflows/`。

**原因**: `.claude/` 是 Claude Code 的原生目录，CC 自动发现 skills/slash commands、memory 索引、workflows。`.harness/` 是我们自己的驾驭工程体系（rules/changes/knowledge），CC 不认识。用链接让两者共存，CC 能发现，我们能独立管理。

### D2: 四支柱的物理布局

**决策**: `rules/` + `skills/` + `knowledge/`(含 memory+业务知识) + `changes/`，平级放在 `.harness/` 下。

**原因**: Rules 是稳定约束（很少变），Skills 是可复用 SOP（流程性），Knowledge 是静态知识（已沉淀），Changes 是动态管道（进行中）。生命周期不同，平级合理。

### D3: 子服务不需要自己的 .harness/

**决策**: 各服务的 CLAUDE.md 引用父 `.harness/rules/`，但不复制规则、不建子 harness。

**原因**: 子 Claude 是 leaf node——接收任务、执行、返回。父 `.harness/` 提供统一的规则/技能/知识/变更管理。子服务只需要知道"去哪里找"而非"自己有一套"。保持了"集中调度/分布执行"的架构。

### D4: Owner Agent 是薄编排层

**决策**: ~165 行，五模块（角色/索引/职责/调度/原则）。只做编排，不重复定义已被 skills/rules/knowledge 覆盖的内容。

**原因**: 项目的 Java 单体 Owner Agent 建议 ~400 行自包含。但我们的 rules/skills/knowledge/changes 已独立建好，Owner Agent 只需要告诉 Agent"在什么阶段、调哪个 Skill、产出什么、门禁是什么"。引用而非重复。

### D5: 7 阶段流水线而非 10 阶段

**决策**: 工具选择 → 需求分析 → 需求评审 → 架构设计 → Proto变更 → 编码+测试 → 集成验证+归档。

**原因**: 文章的 10 阶段（含独立 CI/部署/用户确认阶段）为 Java 单体设计。我们的微服务不需要独立 CI 阶段（go test + harness-checks 即 CI）、不需要每次变更都部署。编码和测试合并因 go test 同包执行。但保留了精确的失败路由表。

### D6: 脚本归属 Skill（OpenClaw 模式）

**决策**: QA 脚本（harness-checks.sh、check-proto-ts-align.sh）放在 `skills/qa/scripts/` 下，作为 qa Skill 的私有执行体。共享基础设施（graph-sync.sh、graph-query.sh）留在 `scripts/`。

**原因**: OpenClaw 模式——Skill 是目录，内含 SKILL.md + scripts/ + assets/。提高了内聚性。谁用归谁，共享的另放。

## 关联经验

- [[pre-commit-checks]] — QA 机械化检查是代码提交的门禁
- [[grpc-only-comms]] — 架构约束写入 `.harness/rules/`
- [[proto-jstype]] — Snowflake 规范写入 `.harness/rules/`
