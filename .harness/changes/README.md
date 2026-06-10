# 变更管理目录

> 本目录是跨服务变更的统一索引，连接 proposal → plan → spec → implementation → CHANGELOG → review 的全追溯链。

## 目录用途

- **索引** (`INDEX.md`)：所有已完成/进行中变更的入口表。每条记录链接到实际工件位置
- **不迁移**：本目录只做索引，不复制工件。各阶段产物留在原位置（`docs/superpowers/`, 各服务 `CHANGELOG.md` 等）

## 为什么不统一放在这里

项目是 8+ 微服务的 monorepo，变更产物天然分布在各处：
- 需求计划 → `docs/superpowers/plans/`
- 设计规范 → `docs/superpowers/specs/`
- 实现记录 → 各服务 `CHANGELOG.md`
- 评审报告 → 各服务 `_review.md` / `_qa.md`
- Proto 变更 → `api-proto/CHANGELOG.md`

`changes/` 的职责是**连接这些分散的工件**，而非重新集中它们。

## 使用方式

- **新增需求时**：在 `INDEX.md` 顶部添加一条记录，链接到 plan/spec 文件
- **完成阶段时**：更新对应记录的 CHANGELOG 和 review 链接
- **回溯时**：从 INDEX.md 找到需求 → 追踪到 plan → spec → CHANGELOG → review
- **子 Agent 启动时**：读取本索引了解近期变更上下文

## 关联

- 规则体系：`.harness/knowledge/memory/MEMORY.md`
- 需求模板：`docs/requirements/TEMPLATE.md`
- Harness 管线：`.harness/workflows/harness-pipeline.js`
