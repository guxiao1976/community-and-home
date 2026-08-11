---
triggers: ['Redis', '缓存', '权限', 'permission', 'KEYS', '失效', '缓存一致性']
type: guideline
severity: must-follow
service: permission-service
status: active
created: 2026-08-09
updated: 2026-08-09
apply_count: 0
---

# 权限缓存一致性 — 修改角色/权限时必须刷新 Redis 缓存

修改角色或权限时（UpdateRole、AssignRole、RevokeRole），必须批量刷新 Redis 缓存（KEYS perm:user:* → DEL），确保缓存与 MySQL 一致。防止旧权限缓存导致的越权访问。

## 参考

详见 `services/permission-service/CLAUDE.md` 和 `services/permission-service/docs/design.md`。
