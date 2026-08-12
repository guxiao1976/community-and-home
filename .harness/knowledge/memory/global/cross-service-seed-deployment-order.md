---
triggers: ["种子", "seed", "init_permissions", "部署顺序", "依赖", "rel_user_role", "系统身份", "静默失败", "migration", "脚本"]
service: all
severity: should-follow
type: pitfall
status: active
created: 2026-08-12
updated: 2026-08-12
---

# 消费方业务流依赖生产方脚本种子（非 migration）必须纳入部署编排

## 为什么会有这条经验

跨服务业务流依赖生产方的「脚本种子」（非 migration 文件，如 permission-service `init_permissions.sql` 的 rel_user_role user_id=0 系统审核身份 global 授权）时，必须保证该种子已在目标环境执行，否则消费方会**静默失败**：数据权限校验被拒 → 业务静默降级/丢弃，且消费方日志显示成功，极难排查。

注意：`scripts/init-databases.sh` 只 CREATE DATABASE，不执行任何种子/建表。

## 怎么做

部署编排显式执行生产方种子脚本并加自检（如 SELECT 校验种子存在），或在消费方对拒判分支记录 WARN 日志。案例：community-hub 审核回调依赖 permission 种子（rel_user_role user_id=0 global），种子缺失时所有 moderation_status 更新被静默拒绝。

## 怎么验证

- 部署脚本是否显式执行生产方种子脚本
- 是否有种子存在性自检（SELECT 校验）
- 消费方拒判分支是否有 WARN 日志

## 关联经验

[[migration-must-execute]] [[rpc-callback-must-check-response-base]]
