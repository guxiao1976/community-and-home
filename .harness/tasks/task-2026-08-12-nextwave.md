---
id: task-2026-08-12-nextwave
title: "access-data-permission 后续 Wave：user/community-hub/web-mobile/集成验收"
service: multi
type: feature
priority: P0
status: open
computed_score: 100.00
source: human
source_detail: "Wave 1 完成交接，阶段 3/4/5/6 待执行（见 .harness/changes/access-data-permission/HANDOFF-NEXT-WAVE.md）"
created: 2026-08-12
blocks: []
blocked_by: []
assigned_run: ""
completed: ""
outcome: ""
---

# access-data-permission 后续 Wave：阶段 3/4/5/6

## 背景

Wave 1（阶段① 数据权限核心）已完整交付并提交（根 4c31166 / master-data 49d476e+a42955e / api-proto c245c09）。permission-service 的 AssertPublishScope RPC、master-data 的 ResolveScopeAncestors 均已就绪，后续 Wave 在此基础上执行编排/消费/前端/集成。

## 执行范围（详见 tasks.md + HANDOFF-NEXT-WAVE.md）

- **阶段 3 · user-service**（编排，非权威）：T3.1 CreateUser 自动分配 registered_user / T3.2 JoinCommunity ownership 自动授权 / T3.3 LeaveCommunity 撤销 / T3.4 门禁
- **阶段 4 · community-hub-service**（消费方）：T4.0 前置 JWT 通道 / T4.1 publisher_id 取 JWT / T4.2-4.4 AssertPublishScope 挂载(lostfound/notice/contacts) / T4.5 moderation 系统身份回调 / T4.6 读列表按 GetDataScopes 过滤 / T4.7 错误码 080006 / T4.8 门禁
- **阶段 5 · web/mobile**：T5.1 joinCommunity 携带 ownership
- **阶段 6 · 集成验收**：T6.1 跨服务端到端验收矩阵 / T6.2 收尾

## 关键约束

- taskType: feature（强制 TDD），阶段 3/4 可并行（无依赖）
- 子 Claude 禁止改 api-proto/（契约已提交 031f4e4 + c245c09）
- community-hub 不得直连 master-data 做 scope 解析，祖先链仅经 permission ResolveScopeAncestors 消费
- 错误码映射：permission 060007 → community-hub 080006
- JWT claim 键统一 user_id，消费用 util.ExtractUserID

## 执行方式

按 CLAUDE.md dispatch Skill 分级（跨服务 → L 级 OpenSpec 已完成，直接进入 N×Workflow 并行），每服务一个 harness-pipeline.js Workflow。
