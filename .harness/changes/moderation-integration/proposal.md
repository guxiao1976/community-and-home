# Proposal: 内容审核全链路集成

## 概述

将社区平台所有内容发布点接入 moderation-service 的合规审核管线（AC 引擎 + 大模型），实现机器审核→人工审核的完整闭环，并开发人工审核管理界面。采用异步 Redis List 驱动架构。

## 影响范围（服务×模块矩阵）

| 服务 | 模块 | 变更类型 | 说明 |
|------|------|:---:|------|
| api-proto | moderation/v1 | 新增 RPC | CreateAuditLog, ListReview, GetReviewDetail |
| api-proto | community/v1 | 新增 RPC | UpdateNoticeModerationStatus, UpdateLostFoundModerationStatus |
| api-proto | user/v1 | 新增 RPC | UpdateUserModerationStatus |
| moderation-service | rpc/consumer | 新增 | Redis BRPOP 消费者 |
| moderation-service | rpc/logic | 新增+修改 | 4 个新 RPC 实现 + SubmitReview 重写 |
| moderation-service | api/handler | 新增 | review REST API (list/detail/submit) |
| moderation-service | model | 修改 | mod_audit_log 增强 (FindList/FindOne/UpdateResult/UpdateReview) |
| community-hub-service | rpc/logic | 修改 | Notice/LostFound 发布接入审核 |
| community-hub-service | model | 修改 | 审核回调 handler + UpdateModerationStatus |
| user-service | rpc/logic | 修改 | CreateUser/SubmitCertification 接入审核 |
| user-service | model | 修改 | 审核回调 handler + UpdateModerationStatus |
| web/pc | views+components | 新增 | 人工审核界面 (4 个组件) |

## 功能列表

### P0（必须）
- [x] 内容发布异步接入审核管线（notice/lost_found/certification/nickname）
- [x] Redis List 任务队列（LPUSH/BRPOP）
- [x] 机器审核结果回写 mod_audit_log
- [x] 人工审核界面（按板块/状态筛选、详情、通过/不通过）
- [x] SubmitReview RPC 实现（从 stub 改为完整实现）

### P1（应该）
- [x] 内容表 moderation_status 字段（4 表 DDL）
- [x] 消费者回调业务服务更新 moderation_status
- [x] 审核 REST API (list/detail/submit)
- [x] 图片审核接口预留

### P2（可以）
- [ ] 时间范围筛选
- [ ] 审核统计面板
- [ ] 用户通知（审核结果推送）

## 约束条件
- 服务间通信仅 gRPC（etcd 发现）
- Proto 变更由全局 Claude 执行
- Snowflake ID → jstype=JS_STRING
- 业务服务不直连 moderation_db
- 密钥在 .env，服务入口用 configx.MustLoad

## 风险与假设

| 风险 | 缓解 |
|------|------|
| Redis 宕机导致审核任务丢失 | 消费者启动时检查残留 pending；task_id 可去重 |
| 大模型超时 | 消费者独立 context 超时 60s，超时入人审 |
| 业务服务回调失败 | 重试 3 次，最终写日志+告警 |
| gRPC 4MB 上限 | 审核内容 ≤10000 字符，图片传 URL |

## 追溯矩阵

| 需求 | 功能 | 服务 |
|------|------|------|
| 所有发布内容审核 | 异步审核管线接入 | moderation, community-hub, user-service |
| 机审不通过推送人审 | need_review 标记 + 人审队列 | moderation-service |
| 记录审核类型+结果 | mod_audit_log 增强 + moderation_status | moderation, community-hub, user-service |
| 人工审核界面 | 按板块/状态筛选+详情+操作 | web/pc, moderation-service (REST API) |
| 内容表审核结果列 | moderation_status 字段 | community-hub, user-service |
| 图片审核预留 | CheckImage 管线预留 | moderation-service |
