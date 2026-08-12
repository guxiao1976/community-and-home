---
triggers: ["FindByRoleId", "assign_time", "rel_user_role", "1054", "Unknown column", "select *", "need_human", "列不存在", "缓存失效", "invalidateRoleCache"]
service: permission-service
type: pitfall
severity: must-follow
status: active
created: 2026-08-12
updated: 2026-08-12
last_applied: 2026-08-12
apply_count: 1
---

# FindByRoleId 查询不存在的 assign_time 列 → MySQL 1054（need_human 修复）

## 为什么会有这条经验

`model/rel.go` `defaultRelUserRoleModel.FindByRoleId` 原 SELECT 显式列清单引用了 `rel_user_role` 表**不存在**的 `assign_time` 列。live 库实测 `rel_user_role` 仅 `id/user_id/role_id/scope_type/scope_id/status/verified_at/expires_at/created_at`，无 `assign_time`。运行时必然 `MySQL 1054 Unknown column 'assign_time'`。

**影响链路**：`rpc updaterolelogic.go invalidateRoleCache` 调用 `UserRoleModel.FindByRoleId` → 1054 报错 → 只打 Error 日志后 return，不删任何用户的 `perm:user:{userId}` Hash 缓存 → **角色权限变更后授权残留最长 30 分钟（安全漏洞）**。

## 如何应用

1. **查询列清单必须与 live 表结构一致**：新增/修改 model 的 SELECT 列清单前，先用 `DESC <table>` 或 `information_schema.COLUMNS` 核对真实列，杜绝引用臆想列名
2. **用 `select *` 代替脆弱列清单**：`RelUserRole` 以 `db` tag 映射，go-zero sqlx 支持 `select *`，与 `FindByUserId` 一致；避免列名漂移
3. **缓存失效路径不能静默吞错**：`invalidateRoleCache` 内 FindByRoleId 失败只打日志就 return 会放大安全影响，修复后 1054 不再阻断，逐个 DEL `perm:user:{userId}` 正常执行
4. **疑似列不存在 → 先查真实表结构**，不要靠直觉写列名

## 相关改动

- `model/rel.go` FindByRoleId：SELECT `id, user_id, role_id, scope_type, scope_id, assign_time` → `select *`（与 FindByUserId 一致）
- 回归测试：`TestRelUserRoleModel_FindByRoleId`（sqlmock 断言 SQL 为 `select *`，不含 assign_time）

## 相关记忆

- [[migration-must-execute]] — 迁移/表结构变更必须真实执行，臆想列名会 1054
