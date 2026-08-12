---
triggers: ["补偿", "rollback", "静默", "_ =", "UpdateBindStatus", "有成员无scope", "幂等", "补偿恢复"]
service: all
severity: should-follow
type: pitfall
status: active
created: 2026-08-12
updated: 2026-08-12
---

# 多步写操作的补偿错误不可静默丢弃（`_ =` 必须记日志）

## 为什么会有这条经验

user-service JoinCommunity/LeaveCommunity 在「写 membership + 调 permission 授权/撤销」两步操作中，第二步失败时用 `_ = UpdateBindStatus(...)` 做 best-effort 补偿；补偿结果被静默丢弃，若补偿也失败，将产生「有成员无 scope」反例且无任何日志。

## 怎么做

1. 补偿（回滚）操作的错误**必须记录日志**（Errorf），不得 `_ =` 丢弃——补偿失败比原失败更需要被看见（不变量可能被破坏）
2. 补偿可返回「补偿失败」告警信息或结构化到响应，至少写日志包含上下文（userId/communityId/原错误/补偿错误）
3. 涉及跨服务幂等依赖（如 AssignRole/RevokeRole 靠唯一索引/INSERT IGNORE）时，额外评估下游对非唯一键错误是否返回真实 error，否则补偿永远不会被触发

## 怎么验证

- grep 变更代码中 `_ = model.UpdateXxx` / `_ = repo.Xxx` 出现在补偿/回滚路径
- 「不留 X」类不变量（如「有成员必有 scope」）依赖补偿生效的场景

## 关联经验

[[insert-ignore-swallows-errors]] [[verify-api-before-calling]]
