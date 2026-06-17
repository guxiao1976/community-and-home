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
notices (1) ────< (N) notice_attachments
community_contacts (独立)
lost_found_items (独立)
```

### 表结构

#### notices（通知公告）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT PK | Snowflake ID |
| community_id | BIGINT NOT NULL | 小区ID |
| title | VARCHAR(200) | 标题 |
| content | TEXT | 富文本正文 |
| role | VARCHAR(20) | 发布角色：community/committee/property/grid_officer |
| publisher | VARCHAR(100) | 发布单位/人名称 |
| publisher_id | BIGINT | 发布人ID |
| is_pinned | TINYINT | 是否置顶 |
| published_at | DATETIME | 发布时间 |
| created_at / updated_at | DATETIME | 自动维护 |
| deleted_at | DATETIME | 软删除标记 |

索引：`idx_community(community_id, deleted_at)`, `idx_published(community_id, published_at DESC, deleted_at)`

#### notice_attachments（通知附件）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT PK | Snowflake ID |
| notice_id | BIGINT | 关联通知 |
| file_name | VARCHAR(200) | 文件名 |
| file_url | VARCHAR(500) | 文件URL |
| file_size | BIGINT | 字节数 |

索引：`idx_notice(notice_id)`

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
