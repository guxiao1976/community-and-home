# 变更索引

> 最后更新：2026-06-09
>
> 格式：`| 日期 | 变更 | Plan | Spec | CHANGELOG | Review | 状态 |`

## 基础设施变更

| 日期 | 变更 | 产出 | 状态 |
|------|------|------|:---:|
| 2026-06-09 | Harness Pipeline Dry Run | [报告](dry-run-2026-06-09.md) | ✅ 完成 |
| 2026-06-09 | Memory frontmatter 统一 + type 字段 | `.harness/knowledge/memory/` (7 文件统一) | ✅ 完成 |
| 2026-06-09 | changes/ 索引目录建立 | [INDEX.md](INDEX.md) + [README.md](README.md) | ✅ 完成 |

## 有完整追溯链的变更

| 日期 | 变更 | Plan | Spec | CHANGELOG | 状态 |
|------|------|:---:|:---:|:---:|:---:|
| 2026-06-09 | 系统配置 Redis 化 | [plan](../docs/superpowers/plans/2026-06-09-system-config-redis.md) | [spec](../docs/superpowers/specs/2026-06-09-system-config-redis-design.md) | *(进行中)* | 🟡 进行中 |
| 2026-06-07 | 我的页面 & 加入小区流程 | [plan](../docs/superpowers/plans/2026-06-07-my-page-and-join-flow.md) | [spec](../docs/superpowers/specs/2026-06-07-my-page-and-join-flow-design.md) | [api-proto](../api-proto/CHANGELOG.md) | ✅ 已完成 |
| 2026-06-06 | 公告信息页重设计 | [plan](../docs/superpowers/plans/2026-06-06-notice-page-redesign.md) | [spec](../docs/superpowers/specs/2026-06-06-notice-page-redesign.md) | [web/mobile](../web/mobile/CHANGELOG.md) | ✅ 已完成 |
| 2026-06-06 | 移动端社区功能 | [plan](../docs/superpowers/plans/2026-06-06-mobile-community-feature.md) | — | [web/mobile](../web/mobile/CHANGELOG.md) | ✅ 已完成 |
| 2026-06-06 | Uni-app 移动端框架 | [plan](../docs/superpowers/plans/2026-06-06-uni-app-mobile-framework.md) | — | [web/mobile](../web/mobile/CHANGELOG.md) | ✅ 已完成 |
| 2026-06-05 | AI 开发团队建设 | [plan](../docs/superpowers/plans/2026-06-05-ai-dev-team-implementation.md) | — | 多服务（见下） | ✅ 已完成 |
| 2026-06-05 | 监控模块 | [plan](../docs/superpowers/plans/2026-06-05-monitoring-module.md) | — | [monitoring-service](../services/monitoring-service/CHANGELOG.md) | ✅ 已完成 |
| 2026-06-02 | 用户服务重构 | [plan](../docs/superpowers/plans/2026-06-02-user-service-refactor.md) | — | [user-service](../services/user-service/CHANGELOG.md) | ✅ 已完成 |

## 仅有 CHANGELOG 的变更（缺 Plan/Spec）

这些变更是 Harness 体系建立前或快速修复产生的，缺少需求分析和设计阶段文档。

| 日期 | 变更 | CHANGELOG | 服务 | 备注 |
|------|------|-----------|------|------|
| 2026-06-06 | community-hub-service 初始化 | [CHANGELOG](../services/community-hub-service/CHANGELOG.md) | community-hub-service | 新服务创建 |
| 2026-06-04 | C1/C2 架构债务修复（HTTP→gRPC） | [CHANGELOG](../services/moderation-service/CHANGELOG.md) | moderation-service | gRPC 合规改造 |
| 2026-06-04 | C8 SMS 验证码绕过修复 | [CHANGELOG](../services/auth-service/CHANGELOG.md) | auth-service | 安全修复 |
| 2026-06-04 | W8 conf.MustLoad → configx.MustLoad | [CHANGELOG](../services/permission-service/CHANGELOG.md) | 多服务 | 全局公约迁移 |
| 2026-06-04 | 错误码 6位→5位统一 | [common](../common/CHANGELOG.md) / [user-service](../services/user-service/CHANGELOG.md) | common, user-service | Breaking change |

## 全局架构变更（Proto / Common）

| 日期 | 变更 | CHANGELOG | 影响范围 |
|------|------|-----------|---------|
| 2026-06-07 | JoinCommunityRequest 增加地址字段 | [api-proto](../api-proto/CHANGELOG.md) | user-service, web/mobile |
| 2026-06-04 | 错误码 v2.1.0（6位→5位） | [common](../common/CHANGELOG.md) | 所有服务 |

## 状态说明

| 标记 | 含义 |
|:---:|------|
| 🟡 进行中 | 有 plan/spec，尚未全部完成 |
| ✅ 已完成 | CHANGELOG 已记录，功能已交付 |
| ⚠️ 缺 Review | 已完成但无评审记录 |
| ❌ 已废弃 | 计划取消或方案变更 |
