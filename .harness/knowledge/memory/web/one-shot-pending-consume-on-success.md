---
triggers: ["pending", "一次性", "清除", "消费", "retry", "跨页", "applyRole", "待加入", "pendingCommunityId", "pending-join"]
service: web/mobile
severity: should-follow
type: pitfall
status: active
created: 2026-08-17
updated: 2026-08-17
last_applied: null
apply_count: 0
---

# 一次性跨页待处理数据应在异步操作成功后消费/清除，而非读取即清

## 为什么会有这条经验

跨页一次性状态（如 `communityStore.pendingCommunityId`）若在发起异步请求前就同步清除，请求失败（网络/校验/后端拒绝）后重试上下文丢失，用户需重走完整流程（如 join-community→join-choice）。

## 怎么做

1. 在异步操作成功回调（.then / 成功后）再消费/清除一次性 pending；失败时保留状态可重试
2. 「读取即清」仅适用于确不需要重试的场景
3. 自测：pending 消费时机断言「申请失败后 pending 仍在」

案例：`web/mobile/src/pages/my/my.vue` `applyForRole` 的 `pendingCommunityId` 改为申请成功后清除（`usedPending` 标记 + `.then` 内 clear）。

## 怎么验证

- 变更中一次性 pending 的清除点位于异步成功回调内
- 失败路径测试断言 pending 未清除（可重试）
