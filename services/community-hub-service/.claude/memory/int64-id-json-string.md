---
triggers: ['int64', 'json', 'Snowflake', 'ID', '精度', '前端']
type: guideline
severity: must-follow
service: community-hub-service
status: active
created: 2026-08-09
updated: 2026-08-09
apply_count: 0
---

# int64 ID 全部使用 json:",string" 确保前端精度

所有 int64 ID 字段在 API types.go 中使用 `json:",string"` tag，确保 Snowflake 19 位 ID 在前端 JavaScript 中不丢失精度。

## 参考

详见 `services/community-hub-service/CLAUDE.md` 和 `services/community-hub-service/docs/design.md`。
