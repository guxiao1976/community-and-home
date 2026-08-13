# 变更追溯索引

> 记录所有通过 OpenSpec 流程完成的需求/功能开发，按时间倒序排列

## 格式说明

```
## YYYY-MM-DD — <变更名>

**路径**: [直接Edit / Dev Agent / OpenSpec]
**状态**: [进行中 / 已完成 / 已归档]
**涉及服务**: service-a, service-b
**关联**: [PR #123] [Issue #456]

[一句话描述]

详见: [.harness/changes/<change-name>/](./<change-name>/)
```

---

## 2026-08-13 — 访问控制与数据权限（access-control）

**路径**: OpenSpec → N×Workflow（7 服务）
**状态**: ✅ 已完成（7/7）
**涉及服务**: permission-service, auth-service, user-service, community-hub-service, master-data-service, web/pc, web/mobile
**关联**: api-proto c86dc84

端限制登录准入（sys_role.platforms + 50007）+ 当前小区应用状态（user_app_state + 10015）+ 板块发布配额（sys_section_quota + 80007）+ 成员约束（每户≤6 + 10014）+ 同屋互见（SameHouseInfo）。需求 1（角色分层）/3（数据权限统一模型）由前置 access-data-permission 交付，本变更只做需求 2/4/5/6。同时作为「检验开发流水线」的真实 L 级样本，暴露并沉淀 6+ 改进点：RED 证据分诊 + 变异测试 stub（harness-pipeline-fix）、web/pc 122 条历史 TS 错误清理（web-pc-debt-cleanup）。

详见: [.harness/changes/access-control/](./access-control/)

---

## 2026-08-12 — 数据权限核心（access-data-permission）

**路径**: OpenSpec → 3×Workflow + Owner 集成验收
**状态**: ✅ 全部完成（阶段①-⑥）
**涉及服务**: permission-service, master-data-service, user-service, community-hub-service, web/mobile
**关联**: proto commit 031f4e4 + c245c09

数据权限统一模型全量落地：scope 三态 + 能力分层(min_verf_level) + registered_user 基角色 + 祖先链解析 + 发布校验。Wave1=阶段①②(permission T1.1-T1.8 / master-data T2.1-T2.3)，Wave2=阶段③④⑤(user T3.1-T3.4 / community-hub T4.0-T4.8 / web-mobile T5.1)+ 阶段⑥ 集成验收(T6.1 九项矩阵全绿)。验收修复 2 个真实缺陷：permission 种子缺口(owner/tenant 发布+读权限/选举绑定) + master-data scope 缓存 O(N²) 加载(keyset 分页+预热,656k 行 27min→4s)。

详见: [.harness/changes/access-data-permission/](./access-data-permission/)

---

## 2026-06-18 — 审核服务管线配置化

**路径**: OpenSpec
**状态**: 已归档
**涉及服务**: moderation-service
**关联**: commit 84eadb2

实现内容审核的管线配置化功能，支持动态配置审核策略。

详见: [.harness/changes/moderation-pipeline-config/](./moderation-pipeline-config/)

---

## 2026-06-16 — AI 模型服务增强

**路径**: Dev Agent
**状态**: 已归档
**涉及服务**: ai-model-service
**关联**: 无

为 AI 模型服务添加连接测试和模板管理功能。

详见: [.harness/changes/ai-model-enhancements/](./ai-model-enhancements/)
