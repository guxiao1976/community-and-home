---
triggers: ["唯一索引", "迁移", "ALTER", "duplicate", "重复数据", "阻塞部署", "ADD UNIQUE", "uk_user_role_scope"]
type: pitfall
severity: should-follow
service: all
status: active
created: 2026-08-12
updated: 2026-08-12
apply_count: 0
---
# 对既有数据加唯一索引必须先查重，否则 ALTER 失败阻塞部署

## 为什么会有这条经验

access-data-permission 阶段① 的 `migration/001_scope_three_state.sql` 对 `rel_user_role` 加 `UNIQUE KEY uk_user_role_scope (user_id, role_id, scope_type, scope_id)`，但**没有先查重**。若表内已有历史重复行（同一 user+role+scope 多条），MySQL `ALTER TABLE ... ADD UNIQUE` 会直接报 `Duplicate entry` 失败，阻塞整个部署流水线。开发阶段无存量数据所以没暴露，但生产环境有历史数据时必炸。

## 怎么做

- 加唯一索引前先执行查重 SQL：
  ```sql
  SELECT user_id, role_id, scope_type, scope_id, COUNT(*)
  FROM <table>
  GROUP BY user_id, role_id, scope_type, scope_id
  HAVING COUNT(*) > 1;
  ```
- 有重复行时：先清洗（保留最新/有效行，删除其余），再 `ADD UNIQUE`。
- 迁移脚本用 guard（`IF NOT EXISTS` 风格 / information_schema 检查索引是否存在）保证幂等可重跑。
- 迁移必须在编码前执行到本地 DB 验证（见 [[migration-must-execute]]）。

## 触发场景

- 任何 `ALTER TABLE ... ADD UNIQUE KEY` / `ADD INDEX` 于已有数据表。
- 幂等分配依赖唯一索引（AssignRole/Join 自动授权）时。

## 关联经验

- [[migration-must-execute]] — 迁移必须执行到本地 DB 后再编码
- [[insert-ignore-swallows-errors]] — 唯一索引幂等语义的配套处理
