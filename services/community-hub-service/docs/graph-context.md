# 知识图谱上下文 — community-hub-service

> 自动生成于 2026-06-09 22:16:29 | 数据源: Neo4j 知识图谱 | 每次 `graph-sync.sh` 后刷新

## 服务标识

| 属性 | 值 |
|------|-----|
| 名称 | community-hub-service |
| 语言 | go |
| 端口 (gRPC) | 8087 |
| 端口 (API)  | 8887 |

## 服务依赖

| 依赖服务 | 依赖类型 |
|---------|---------|
| master-data-service | gRPC |

## 被依赖方

无服务依赖本服务

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

## gRPC 接口

| RPC 方法 | 输入消息 | 输出消息 |
|---------|---------|---------|
| CreateLostFound | CreateLostFoundRequest | CreateLostFoundResponse |
| CreateNotice | CreateNoticeRequest | CreateNoticeResponse |
| DeleteNotice | DeleteNoticeRequest | DeleteNoticeResponse |
| GetLostFound | GetLostFoundRequest | GetLostFoundResponse |
| GetNotice | GetNoticeRequest | GetNoticeResponse |
| ListContacts | ListContactsRequest | ListContactsResponse |
| ListLostFound | ListLostFoundRequest | ListLostFoundResponse |
| ListNotices | ListNoticesRequest | ListNoticesResponse |
| ResolveLostFound | ResolveLostFoundRequest | ResolveLostFoundResponse |
| UpdateNotice | UpdateNoticeRequest | UpdateNoticeResponse |
| UpsertContacts | UpsertContactsRequest | UpsertContactsResponse |

## 数据库表

| 表名 | 列 |
|------|-----|
| community_contacts | updated_at (datetime), created_at (datetime), sort_order (bigint), phone (varchar), name (varchar), category (varchar), community_id (bigint), id (bigint) |
| lost_found_items | deleted_at (datetime), updated_at (datetime), created_at (datetime), publisher_id (bigint), status (varchar), contact_phone (varchar), image_urls (varchar), description (varchar), title (varchar), type (varchar), community_id (bigint), id (bigint) |
| notice_attachments | created_at (datetime), file_size (bigint), file_url (varchar), file_name (varchar), notice_id (bigint), id (bigint) |
| notices | deleted_at (datetime), updated_at (datetime), created_at (datetime), published_at (datetime), is_pinned (bigint), publisher (varchar), role (varchar), content (varchar), title (varchar), community_id (bigint), id (bigint) |

## 前端消费方

| 方法 | URL | 文件 |
|------|-----|------|
| GET | /api/community/contacts | web/mobile/src/api/community.ts |
| GET | /api/community/notices | web/mobile/src/api/community.ts |

## 实体血缘（Proto → Go → DB）

| Proto 消息 | Go 结构体 | 数据库表 |
|-----------|----------|---------|
| ContactEntry | ContactEntry | - |
| LostFoundItem | LostFoundItem | lost_found_items |
| Notice | Notice | notices |
| NoticeAttachment | NoticeAttachment | notice_attachments |

---
*此文件由 graph-sync.sh 自动生成，请勿手动编辑。*
