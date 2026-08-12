---
triggers: ["INSERT IGNORE", "幂等", "唯一键", "静默", "假成功", "RowsAffected", "ON DUPLICATE KEY", "AssignRole"]
type: pitfall
severity: should-follow
service: all
status: active
created: 2026-08-12
updated: 2026-08-12
apply_count: 0
---
# INSERT IGNORE 静默吞掉非唯一键错误导致假成功

## 为什么会有这条经验

用 `INSERT IGNORE` 实现唯一键幂等（如 AssignRole 重复分配）时，MySQL 会**忽略所有错误**，不仅是 `Duplicate key` —— 还包括外键约束失败（FK）、字段超长截断、NOT NULL 违反、数据类型错误等。调用方只看 RowsAffected==0 就当作"已存在"，实际可能是"写入失败被吞掉"，造成假成功。

## 怎么做

- 幂等语义仅靠 `INSERT IGNORE` 不够：非唯一键错误（FK/截断）会静默失败，调用方无法区分"重复"与"真实错误"。
- 需要区分时：改用 `INSERT ... ON DUPLICATE KEY UPDATE`（可拿到受影响行数语义）或先 `SELECT` 判断存在性再 `INSERT`，或对 `RowsAffected` 结合错误码判断。
- 如果 `INSERT IGNORE` 的 RowsAffected==0 被当作幂等成功，必须在**同一事务内**先验证目标行确实存在（查一次），否则假成功。
- model 层单测必须覆盖：重复键返回 `NewResult(0,0)`、非唯一键错误（FK）应返回 error（断言不会静默成功）。

## 触发场景

- 用 `INSERT IGNORE` 实现角色分配/授权等幂等语义时。
- 审查 `InsertIgnore` 类方法是否只断言 nil error 而忽略其他错误类型。

## 关联经验

- [[redis-cache-soft-delete]] — 授权变更后缓存失效收敛
- [[error-code-collision-and-namespace-alignment]] — 授权失败的错误码语义必须唯一
