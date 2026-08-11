---
triggers: ["migration", "数据库变更", "DDL", "ALTER TABLE", "sql", "column", "schema", "1054", "unknown column"]
service: all
type: pitfall
severity: must-follow
status: active
created: 2026-06-08
updated: 2026-06-08
last_applied: null
apply_count: 0
---

# Migration 文件提交后必须执行

## 为什么会有这条经验

加入小区功能新增了 `user_community_membership` 表的 `building`/`unit`/`room` 列，migration SQL 文件写好了也提交了，但**没有真正在数据库上执行**。加入小区时后端报 `Error 1054: unknown column 'building'`，前端弹窗 "rpc error"。

## 怎么做

1. **Migration 必须三步闭环**：写 SQL → 提交 → **执行**
2. 执行命令示例（MySQL）：
   ```bash
   docker exec mysql mysql -u root -p<password> <db> -e "ALTER TABLE ..."
   ```
3. 执行后必须验证：
   ```bash
   docker exec mysql mysql -u root -p<password> <db> -e "DESCRIBE <table>;"
   ```
4. 任何涉及数据库 schema 变更的 Task，完成标准必须包含「Migration 已执行并验证」

## 怎么验证

- 后端日志无 `unknown column` 错误
- 相关功能正常（如加入小区不报错）
- `DESCRIBE` 确认新列存在

## 关联经验
- [[pre-commit-checks]] — gate 目前不检查 DB schema，需人工确认
- [[verify-api-before-calling]]
