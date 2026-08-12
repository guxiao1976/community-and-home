---
triggers: ["非阻塞", "仅告警", "不阻塞", "registered_user", "授权", "grant", "对账", "backfill", "重试", "不变量"]
service: all
severity: should-follow
type: pitfall
status: active
created: 2026-08-12
updated: 2026-08-12
---

# 非阻塞（仅告警）的跨服务授权写入必须有重试/对账

## 为什么会有这条经验

当注册等一次性入口把「分配基角色」这类业务不变量写成非阻塞（失败仅 Errorf 不返回错误）时，必须考虑对账/重试/backfill 机制。原因：

- 这类入口（如 CreateUser）只在首次触发，后续正常流程不再重入
- 若依赖的下游（如 permission-service）在故障窗口不可用，授权静默丢失且永远不会自愈，用户永久缺失基角色
- 与同模块其他路径（如 JoinCommunity 授权失败→fatal+补偿）的失败哲学不一致，容易让人误以为「非阻塞=可安全丢弃」

## 怎么做

要么提供幂等重试队列/独立对账任务扫描「缺基角色」的用户补发；要么在文档/CHANGELOG 中明确记录该不变量存在暂时性缺失窗口及恢复方式。

## 怎么验证

- 涉及用户服务 CreateUser assignRegisteredUser（create_user_logic.go）
- 检查一次性入口的非阻塞写入是否有对账/重试机制或明确记录的恢复方式

## 关联经验

[[best-effort-compensation-must-log]] [[auto-grant-unverified-grant-confers-scope-level0]]
