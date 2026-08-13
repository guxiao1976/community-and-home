# 知识图谱上下文 — file-service

> 自动生成于 2026-08-13 07:57:07 | 数据源: Neo4j 知识图谱 | 每次 `graph-sync.sh` 后刷新

## 服务标识

| 属性 | 值 |
|------|-----|
| 名称 | file-service |
| 语言 | go |
| 端口 (gRPC) | 8085 |
| 端口 (API)  | 8884 |

## 服务依赖

| 依赖服务 | 依赖类型 |
|---------|---------|
| master-data-service | gRPC |

## 被依赖方

无服务依赖本服务

## REST API 路由

| 方法 | 路径 |
|------|------|
| Get | /api/files/ |
| Delete | /api/files/:id |
| Get | /api/files/:id |
| Post | /api/files/confirm |
| Post | /api/files/upload-url |

## gRPC 接口

| RPC 方法 | 输入消息 | 输出消息 |
|---------|---------|---------|
| ConfirmUpload | ConfirmUploadRequest | ConfirmUploadResponse |
| DeleteFile | DeleteFileRequest | DeleteFileResponse |
| GetFileUrl | GetFileUrlRequest | GetFileUrlResponse |
| GetUploadUrl | GetUploadUrlRequest | GetUploadUrlResponse |
| ListFiles | ListFilesRequest | ListFilesResponse |

## 数据库表

无数据库表

## 前端消费方

> ⚠️ 未匹配到服务特定路由。列出所有前端 API 调用：

| 方法 | URL | 文件 |
|------|-----|------|
| POST | /api/users/roles/apply | web/mobile/src/api/user.ts |
| POST | /api/perm/user-roles | web/pc/src/api/identity.ts |
| POST | /api/perm/permissions/auto-discover | web/pc/src/api/identity.ts |
| POST | /api/masterdata/approval/batch-review | web/pc/src/api/masterdata.ts |
| POST | /api/masterdata/divisions/batch-submit | web/pc/src/api/masterdata.ts |
| POST | /api/masterdata/residential-areas/batch-submit | web/pc/src/api/masterdata.ts |
| POST | /api/masterdata/sensitive-words/batch-submit | web/pc/src/api/masterdata.ts |
| POST | /api/users/residences/bind | web/mobile/src/api/user.ts |
| POST | /api/v1/model/call | web/pc/src/api/aimodel.ts |
| POST | /api/moderation/image/check | web/pc/src/api/moderation.ts |
| POST | /api/moderation/text/check | web/pc/src/api/moderation.ts |
| POST | /api/masterdata/divisions | web/pc/src/api/masterdata.ts |
| POST | /api/v1/apikey | web/pc/src/api/aimodel.ts |
| POST | /api/v1/model | web/pc/src/api/aimodel.ts |
| POST | /api/moderation/pipeline | web/pc/src/api/moderation.ts |
| POST | /api/masterdata/residential-areas | web/pc/src/api/masterdata.ts |
| POST | /api/perm/roles | web/pc/src/api/identity.ts |
| POST | /api/v1/template | web/pc/src/api/aimodel.ts |
| POST | /api/users | web/pc/src/api/identity.ts |
| POST | /api/v1/model/fetch-provider-models | web/pc/src/api/aimodel.ts |
| GET | /api/masterdata/divisions | web/pc/src/api/masterdata.ts |
| GET | /api/v1/apikeys | web/pc/src/api/aimodel.ts |
| GET | /api/v1/models | web/pc/src/api/aimodel.ts |
| GET | /api/masterdata/configurations | web/pc/src/api/masterdata.ts |
| GET | /api/community/contacts | web/mobile/src/api/community.ts |
| GET | /api/masterdata/deleted-items | web/pc/src/api/masterdata.ts |
| GET | /api/masterdata/statistics/division-counts | web/pc/src/api/masterdata.ts |
| GET | /api/masterdata/statistics/division-counts/realtime | web/pc/src/api/masterdata.ts |
| GET | /api/v1/health-checks | web/pc/src/api/aimodel.ts |
| GET | /api/monitoring/health | web/pc/src/api/monitoring.ts |
| GET | /api/community/lost-found | web/mobile/src/api/community.ts |
| GET | /api/v1/models | web/pc/src/api/aimodel.ts |
| GET | /api/v1/templates | web/pc/src/api/aimodel.ts |
| GET | /api/community/notices | web/mobile/src/api/community.ts |
| GET | /api/masterdata/approval/pending-items | web/pc/src/api/masterdata.ts |
| GET | /api/perm/permissions | web/pc/src/api/identity.ts |
| GET | /api/masterdata/residential-areas | web/pc/src/api/masterdata.ts |
| POST | /api/masterdata/residential-areas/batch | web/mobile/src/api/user.ts |
| GET | /api/moderation/review/detail | web/pc/src/api/moderation.ts |
| GET | /api/masterdata/submission-records/reviewed | web/pc/src/api/masterdata.ts |
| GET | /api/perm/roles | web/pc/src/api/identity.ts |
| GET | /api/masterdata/sensitive-words | web/pc/src/api/masterdata.ts |
| GET | /api/masterdata/amap-sync/progress | web/pc/src/api/masterdata.ts |
| GET | /api/v1/templates | web/pc/src/api/aimodel.ts |
| GET | /api/v1/statistics | web/pc/src/api/aimodel.ts |
| GET | /api/users/communities/memberships | web/mobile/src/api/user.ts |
| GET | /api/users/profile | web/mobile/src/api/identity.ts |
| GET | /api/users/roles | web/mobile/src/api/user.ts |
| GET | /api/users | web/pc/src/api/identity.ts |
| GET | /api/verifications | web/pc/src/api/identity.ts |
| POST | /api/users/communities/join | web/mobile/src/api/user.ts |
| POST | /api/users/communities/leave | web/mobile/src/api/user.ts |
| GET | /api/moderation/pipelines | web/pc/src/api/moderation.ts |
| GET | /api/moderation/review/list | web/pc/src/api/moderation.ts |
| POST | /api/auth/login | web/pc/src/api/identity.ts |
| POST | /api/auth/login | web/mobile/src/api/identity.ts |
| POST | /api/auth/login/sms | web/pc/src/api/identity.ts |
| POST | /api/auth/login/sms | web/mobile/src/api/identity.ts |
| POST | /api/auth/logout | web/pc/src/api/identity.ts |
| GET | /api/masterdata/query/residential-areas | web/pc/src/api/masterdata.ts |
| POST | /api/auth/token/refresh | web/pc/src/api/identity.ts |
| POST | /api/auth/token/refresh | web/mobile/src/api/identity.ts |
| POST | /api/auth/register | web/pc/src/api/identity.ts |
| POST | /api/auth/register | web/mobile/src/api/identity.ts |
| DELETE | /api/perm/user-roles | web/pc/src/api/identity.ts |
| POST | /api/auth/sms/send | web/pc/src/api/identity.ts |
| POST | /api/auth/sms/send | web/mobile/src/api/identity.ts |
| POST | /api/moderation/review/submit | web/pc/src/api/moderation.ts |
| POST | /api/v1/model/test-connection | web/pc/src/api/aimodel.ts |
| POST | /api/moderation/pipeline/test | web/pc/src/api/moderation.ts |
| POST | /api/v1/template/test | web/pc/src/api/aimodel.ts |
| PUT | /api/v1/apikey | web/pc/src/api/aimodel.ts |
| PUT | /api/v1/model | web/pc/src/api/aimodel.ts |
| PUT | /api/v1/template | web/pc/src/api/aimodel.ts |

## 实体血缘（Proto → Go → DB）

| Proto 消息 | Go 结构体 | 数据库表 |
|-----------|----------|---------|
| FileInfo | FileInfo | - |

---
*此文件由 graph-sync.sh 自动生成，请勿手动编辑。*
