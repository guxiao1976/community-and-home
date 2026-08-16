# 知识图谱上下文 — community-hub-service

> 自动生成于 2026-08-16 10:53:32 | 数据源: Neo4j 知识图谱 | 每次 `graph-sync.sh` 后刷新

## 服务标识

| 属性 | 值 |
|------|-----|
| 名称 | community-hub-service |
| 语言 | go |
| 端口 (gRPC) | 8088 |
| 端口 (API)  | 8887 |

## 服务依赖

| 依赖服务 | 依赖类型 |
|---------|---------|
| file-service | gRPC |
| master-data-service | gRPC |
| moderation-service | gRPC |
| permission-service | gRPC |
| user-service | gRPC |

## 被依赖方

| 消费方 | 依赖类型 |
|---------|---------|
| moderation-service | None |

## REST API 路由

| 方法 | 路径 |
|------|------|
| Post | /api/community/contacts |
| Get | /api/community/contacts |
| Get | /api/community/lostfound |
| Post | /api/community/lostfound |
| Get | /api/community/lostfound/:id |
| Post | /api/community/lostfound/:id/resolve |
| Get | /api/community/notices |
| Post | /api/community/notices |
| Delete | /api/community/notices/:id |
| Put | /api/community/notices/:id |
| Get | /api/community/notices/:id |
| Get | /api/community/notices/marquee |
| Get | /api/community/notices/publish-permission |

## gRPC 接口

| RPC 方法 | 输入消息 | 输出消息 |
|---------|---------|---------|
| CreateContentPost | CreateContentPostRequest | CreateContentPostResponse |
| CreateLostFound | CreateLostFoundRequest | CreateLostFoundResponse |
| CreateLostFound | CreateLostFoundRequest | CreateLostFoundResponse |
| CreateNotice | CreateNoticeRequest | CreateNoticeResponse |
| DeleteContentPost | DeleteContentPostRequest | DeleteContentPostResponse |
| DeleteNotice | DeleteNoticeRequest | DeleteNoticeResponse |
| GetContentPost | GetContentPostRequest | GetContentPostResponse |
| GetLostFound | GetLostFoundRequest | GetLostFoundResponse |
| GetLostFound | GetLostFoundRequest | GetLostFoundResponse |
| GetMarqueeNotices | GetMarqueeNoticesRequest | GetMarqueeNoticesResponse |
| GetNotice | GetNoticeRequest | GetNoticeResponse |
| GetPublishPermission | GetPublishPermissionRequest | GetPublishPermissionResponse |
| ListContacts | ListContactsRequest | ListContactsResponse |
| ListContacts | ListContactsRequest | ListContactsResponse |
| ListContentPosts | ListContentPostsRequest | ListContentPostsResponse |
| ListLostFound | ListLostFoundRequest | ListLostFoundResponse |
| ListLostFound | ListLostFoundRequest | ListLostFoundResponse |
| ListNotices | ListNoticesRequest | ListNoticesResponse |
| ResolveLostFound | ResolveLostFoundRequest | ResolveLostFoundResponse |
| ResolveLostFound | ResolveLostFoundRequest | ResolveLostFoundResponse |
| UpdateContentPost | UpdateContentPostRequest | UpdateContentPostResponse |
| UpdateLostFoundModerationStatus | UpdateModerationStatusRequest | UpdateModerationStatusResponse |
| UpdateLostFoundModerationStatus | UpdateModerationStatusRequest | UpdateModerationStatusResponse |
| UpdateNotice | UpdateNoticeRequest | UpdateNoticeResponse |
| UpdateNoticeModerationStatus | UpdateModerationStatusRequest | UpdateModerationStatusResponse |
| UpsertContacts | UpsertContactsRequest | UpsertContactsResponse |
| UpsertContacts | UpsertContactsRequest | UpsertContactsResponse |

## 数据库表

| 表名 | 列 |
|------|-----|
| community_contacts | updated_at (datetime), created_at (datetime), sort_order (bigint), phone (varchar), name (varchar), category (varchar), community_id (bigint), id (bigint) |
| content_post_attachments | created_at (datetime), file_type (varchar), file_id (bigint), review_status (bigint), file_size (bigint), file_url (varchar), file_name (varchar), post_id (bigint), id (bigint) |
| content_post_scope | created_at (datetime), community_id (bigint), post_id (bigint) |
| content_posts | deleted_at (datetime), updated_at (datetime), created_at (datetime), moderation_time (nullable), moderation_status (bigint), kafka_pushed_at (nullable), kafka_push_last_error (nullable), kafka_push_retries (bigint), kafka_push_status (bigint), attachment_count (bigint), status (bigint), section_code (varchar), published_at (nullable), is_pinned (bigint), publisher_id (bigint) ... |
| lost_found_items | deleted_at (datetime), updated_at (datetime), created_at (datetime), moderation_time (nullable), moderation_status (bigint), publisher_id (bigint), status (varchar), contact_phone (varchar), image_urls (varchar), description (varchar), title (varchar), type (varchar), community_id (bigint), id (bigint) |
| notice_attachments | created_at (datetime), file_size (bigint), file_url (varchar), file_name (varchar), notice_id (bigint), id (bigint) |
| notices | deleted_at (datetime), updated_at (datetime), created_at (datetime), moderation_time (nullable), moderation_status (bigint), published_at (datetime), is_pinned (bigint), publisher_id (bigint), publisher (varchar), role (varchar), content (varchar), title (varchar), community_id (bigint), id (bigint) |

## 前端消费方

| 方法 | URL | 文件 |
|------|-----|------|
| GET | /api/community/lostfound | web/mobile/src/api/community.ts |
| GET | /api/community/contacts | web/mobile/src/api/community.ts |
| GET | /api/community/notices | web/mobile/src/api/community.ts |

## 实体血缘（Proto → Go → DB）

| Proto 消息 | Go 结构体 | 数据库表 |
|-----------|----------|---------|
| ContactEntry | ContactEntry | - |
| ContentPost | ContentPost | content_posts |
| ContentPostAttachment | ContentPostAttachment | content_post_attachments |
| LostFoundItem | LostFoundItem | lost_found_items |
| Notice | Notice | notices |
| NoticeAttachment | NoticeAttachment | notice_attachments |

---
*此文件由 graph-sync.sh 自动生成，请勿手动编辑。*
