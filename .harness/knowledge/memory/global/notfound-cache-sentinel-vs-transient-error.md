---
triggers: ["缓存", "not-found", "sentinel", "ErrNoRows", "瞬时错误", "缓存失效", "DB抖动", "30分钟", "min_verf_level", "冷缓存"]
type: pitfall
severity: should-follow
service: all
status: active
created: 2026-08-12
updated: 2026-08-12
apply_count: 0
---
# 缓存 not-found sentinel 前必须区分 ErrNoRows 与瞬时 DB 错误

## 为什么会有这条经验

permission-service `permissionDefMinLevel`（checkpermissionlogic.go）查询权限定义 `min_verf_level` 时，把 DB 查询**所有错误**（包括瞬时连接错误/超时）都当作 not-found，缓存 "-1" sentinel 30 分钟。冷缓存期一次 DB 抖动 → 该权限对所有用户标记为不存在 → 拒绝授权最长 30 分钟，且期间不重试回源。同理 `grantSatisfiedLevel` 缓存路径也存在此风险。

## 怎么做

- 缓存 not-found sentinel 前必须区分 `sql.ErrNoRows`（真不存在，可缓存）与瞬时错误（连接/超时，**不可缓存**，应立即返回 error 让调用方重试或 fail-open/closed 决策）。
- go-zero `QueryRowCtx` 返回 `sql.ErrNoRows` 表示无结果；其他 err 均为真实错误，不得当 not-found。
- 安全关键路径（权限判断）宁可 fail-closed（拒绝）也不要缓存错误的 not-found；但瞬时错误至少不能被持久化成 sentinel。
- 单测覆盖：`ErrNoRows` → 缓存 sentinel；`context.DeadlineExceeded` / 连接错误 → 返回 error，不写 sentinel。

## 触发场景

- 缓存 permissionDef / 权限定义 / 配置等冷缓存查询。
- 审查缓存回源逻辑是否把 `err != nil` 一律当 not-found 处理。

## 关联经验

- [[redis-cache-soft-delete]] — 缓存失效与软删除联动
- [[error-code-collision-and-namespace-alignment]] — 错误返回的语义一致性
