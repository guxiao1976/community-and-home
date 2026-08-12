---
triggers: ["补偿恢复", "回滚恢复", "UpdateBindStatus", "time.Time{}", "零值", "DATETIME", "strict 模式"]
service: all
severity: should-follow
type: pitfall
status: active
created: 2026-08-12
updated: 2026-08-12
---

# 补偿/恢复路径严禁用 time.Time{} 零值写入 MySQL DATETIME

## 为什么会有这条经验

Go 的 `time.Time{}` 零值序列化为 `0001-01-01 00:00:00`，超出 MySQL DATETIME 范围（min 1000-01-01）。在 strict SQL mode 下该 UPDATE 直接报错，若错误被 `_ =` 丢弃则补偿静默失败；非 strict 模式写入 '0000-00-00' 垃圾值。

错误示例：`UpdateBindStatus(ctx, id, active, time.Time{})`（恢复 bind_status 时把零值写进 leave_time）。

## 怎么做

恢复语义应写 NULL——`sql.NullTime{}` / 参数为 NULL，或为模型增加专门的重置方法（如 `RestoreActive(ctx,id)` 只更新 bind_status 和 updated_at，不动 leave_time）。

注意：单测若用内存 mock 模型（直接存 struct 字段）会掩盖此问题，因为 mock 不经过真实驱动序列化；补真实 sqlmock 用例验证恢复 SQL 的参数。

## 怎么验证

- 涉及用户服务 LeaveCommunity 撤销失败补偿（leave_community_logic.go）
- 检查恢复路径 SQL 参数是否为 NULL 而非 0001-01-01

## 关联经验

[[best-effort-compensation-must-log]]
