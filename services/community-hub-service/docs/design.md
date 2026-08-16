# Community Hub Service — 设计文档

## 概述

社区枢纽服务（community-hub-service）是社区平台的小区信息汇聚中心，负责通知公告发布、便民联络信息维护、寻失互助等社区内容场景。

## 设计决策

### 1. 双层架构

同时提供 RPC 和 REST API 两种接入方式：
- **gRPC (8087)**：服务间通信，注册到 etcd (`community-hub.rpc`)
- **REST API (8887)**：前端接入，通过 gRPC 客户端代理到 RPC 层

### 2. 三层 Service 独立注册

虽然三个 Service（Notice/Contact/LostFound）共用同一个 MySQL 实例和 ServiceContext，但在 gRPC 层面作为独立的 Service 注册，遵循 proto 定义的分层结构。未来可独立拆分为不同服务。

### 3. Snowflake ID

所有表的主键使用 Snowflake 分布式唯一 ID 生成。由 `common/pkg/snowflake` 提供，自动检测节点 ID（K8s POD_IP → 本地 IP → 随机）。

### 4. 软删除

notices 和 lost_found_items 表使用软删除（deleted_at），community_contacts 使用硬删除（Upsert 语义：先删后插）。

## 数据库设计

### ER 图

```
content_posts (1) ────< (N) content_post_scope     多小区范围关联
content_posts (1) ────< (N) content_post_attachments
community_contacts (独立)
lost_found_items (独立)
```

### 表结构

#### content_posts（通用图文发布，原 notices RENAME，Migration 003）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT PK | Snowflake ID |
| community_id | BIGINT DEFAULT NULL | 弃用：范围关联单源 content_post_scope（兼容期保留列，不写入） |
| title | VARCHAR(200) | 标题 |
| text | TEXT | 正文（原 content 改名；REST wire 仍以 content 键输出） |
| role | VARCHAR(20) | 发布角色：community/committee/property/grid_officer（RBAC→映射派生） |
| publisher | VARCHAR(100) | 展示名（取用户真实档案，禁请求体信任） |
| publisher_id | BIGINT | 发布人ID（JWT 派生） |
| is_pinned | TINYINT | 是否置顶 |
| published_at | DATETIME DEFAULT NULL | 审核锚定：本期 submit 即置 NOW() |
| section_code | VARCHAR(30) DEFAULT 'notice' | 板块：notice=通知/repair=维修保修/... |
| status | TINYINT DEFAULT 0 | 全生命周期+审核结果：0=draft 1=submitted 2=approved 3=rejected 4=withdrawn |
| attachment_count | INT DEFAULT 0 | 附件计数（审核完整性判定载体） |
| moderation_status / moderation_time | 兼容期保留 | 逐步过渡到 status + 附件级 |
| kafka_push_status | TINYINT DEFAULT 0 | 0=无待推 1=pending-push 2=已推(ack) |
| kafka_push_retries | INT DEFAULT 0 | 重推次数 |
| kafka_push_last_error | VARCHAR(500) NULL | 最近一次推送错误摘要（可观测） |
| kafka_pushed_at | DATETIME NULL | 成功推送时间 |
| created_at / updated_at | DATETIME | 自动维护 |
| deleted_at | DATETIME | 软删除标记 |

索引：`idx_community(community_id, deleted_at)`（deprecated 兼容期保留）、`idx_published(community_id, published_at DESC, deleted_at)`（deprecated 兼容期保留）、`idx_status_pinned_published(status, is_pinned, published_at)`（migration 005，通知 30 天窗口过滤读路径：等值 status/is_pinned + 范围 published_at，覆盖 ORDER BY 前导列）；新读路径走 content_post_scope 索引。

#### content_post_scope（内容帖-小区范围关联，多小区发布单源）

| 字段 | 类型 | 说明 |
|------|------|------|
| post_id | BIGINT NOT NULL | 关联 content_posts.id（复合 PK） |
| community_id | BIGINT NOT NULL | 目标小区（复合 PK + idx_scope_community 读索引） |
| created_at | DATETIME | 默认 CURRENT_TIMESTAMP |

复合 PK `(post_id, community_id)`；索引 `idx_scope_community(community_id, post_id)`。纯关联表仅 created_at（显式偏离 §3.1 时间三件套）；撤回不删行（主表软删表达）。

#### content_post_attachments（内容帖附件，原 notice_attachments RENAME）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT PK | Snowflake ID |
| post_id | BIGINT | 关联 content_posts.id（原 notice_id 改名，post_id 全链一致） |
| file_name | VARCHAR(200) | 文件名 |
| file_url | VARCHAR(1024) | 存量回退用 stored URL（新行占位空串，file_id 为权威重生载体） |
| file_size | BIGINT | 字节数 |
| review_status | TINYINT DEFAULT 1 | 附件级审核：0=pending 1=approved 2=rejected（本期默认 approved） |
| file_id | BIGINT DEFAULT 0 | file-service 文件ID（重生预签名 URL 权威载体）；兼容期存量行 0 |
| file_type | VARCHAR(20) NULL | 白名单校验通过的文件类型（扩展名） |
| created_at | DATETIME | 自动维护 |

索引：`idx_notice(post_id)`（RENAME 后自动改指 post_id）。

#### community_contacts（便民联络）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT PK | Snowflake ID |
| community_id | BIGINT | 小区ID |
| category | VARCHAR(30) | 类别：water/electricity/gas/unicom/mobile/telecom/police |
| name | VARCHAR(100) | 显示名称 |
| phone | VARCHAR(20) | 电话号码 |
| sort_order | INT | 排序 |

索引：`idx_community(community_id)`

#### lost_found_items（寻失互助）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT PK | Snowflake ID |
| community_id | BIGINT | 小区ID |
| type | VARCHAR(10) | lost/found |
| title | VARCHAR(200) | 标题 |
| description | TEXT | 描述 |
| image_urls | JSON | 图片URL数组 |
| contact_phone | VARCHAR(20) | 联系电话 |
| status | VARCHAR(20) | active/resolved |
| publisher_id | BIGINT | 发布人ID |
| created_at / updated_at | DATETIME | 自动维护 |
| deleted_at | DATETIME | 软删除 |

索引：`idx_community_type(community_id, type, status, deleted_at)`, `idx_created(community_id, created_at DESC)`

## 业务流程

### 通知发布

1. 接收 CreateNoticeRequest（community_id, title, content, role, publisher）
2. 生成 Snowflake ID
3. 写入 notices 表
4. 返回通知 ID

> 注意：当前版本不包含角色校验（user-service 集成）和审核对接（moderation-service），后续补充。

### 便民联络维护

1. UpsertContacts：先删除 community_id 下所有旧数据，再逐条插入新数据
2. ListContacts：按 community_id 查询，按 sort_order 排序

### 寻失互助

1. CreateLostFound：生成 Snowflake ID，status 默认 "active"
2. ListLostFound：按 community_id 分页查询，可选按 type 过滤，created_at DESC
3. ResolveLostFound：将 status 更新为 "resolved"

## 接口契约

详见 `api-proto/api/community/v1/community.proto`。

## 缓存策略

当前版本不使用 Redis 缓存。数据量较小，直接查询 MySQL 即可。后续视负载情况考虑引入 Redis 缓存热门通知。

## 安全机制

- 通知发布需社区角色认证（当前版本未实现，后续接入 user-service）
- 寻失互助发布需登录（当前版本未实现）
- 所有发布内容经过 moderation-service 审核后公开可见（当前版本未实现）
- 便民联络仅管理员可维护，普通用户只读（当前版本未实现权限校验）
