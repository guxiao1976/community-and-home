# 知识图谱上下文 — user-service

> 自动生成于 2026-08-12 15:15:21 | 数据源: Neo4j 知识图谱 | 每次 `graph-sync.sh` 后刷新

## 服务标识

| 属性 | 值 |
|------|-----|
| 名称 | user-service |
| 语言 | go |
| 端口 (gRPC) | 8082 |
| 端口 (API)  | 8882 |

## 服务依赖

| 依赖服务 | 依赖类型 |
|---------|---------|
| master-data-service | gRPC |
| moderation-service | gRPC |
| permission-service | gRPC |

## 被依赖方

| 消费方 | 依赖类型 |
|---------|---------|
| auth-service | None |

## REST API 路由

| 方法 | 路径 |
|------|------|
| Post | /api/users |
| Get | /api/users |
| Delete | /api/users/:id |
| Put | /api/users/:id |
| Get | /api/users/:id |
| Get | /api/users/certifications |
| Post | /api/users/certifications |
| Post | /api/users/communities/join |
| Post | /api/users/communities/leave |
| Get | /api/users/communities/memberships |
| Get | /api/users/profile |
| Post | /api/users/residences/bind |
| Get | /api/users/roles |
| Post | /api/users/roles/apply |
| Get | /api/verifications |
| Post | /api/verifications/:id/review |

## gRPC 接口

| RPC 方法 | 输入消息 | 输出消息 |
|---------|---------|---------|
| ApplyRole | ApplyRoleRequest | ApplyRoleResponse |
| BatchUpdateUsers | BatchUpdateUsersRequest | BatchUpdateUsersResponse |
| BindResidence | BindResidenceRequest | BindResidenceResponse |
| CheckAccess | CheckAccessRequest | CheckAccessResponse |
| CreateUser | CreateUserRequest | CreateUserResponse |
| GetMyCertifications | GetMyCertificationsRequest | GetMyCertificationsResponse |
| GetResidences | GetResidencesRequest | GetResidencesResponse |
| GetUser | GetUserRequest | GetUserResponse |
| GetUserByPhone | GetUserByPhoneRequest | GetUserResponse |
| GetUserMemberships | GetUserMembershipsRequest | GetUserMembershipsResponse |
| GetUserRoles | GetUserRolesRequest | GetUserRolesResponse |
| GetUsersByIds | GetUsersByIdsRequest | GetUsersByIdsResponse |
| JoinCommunity | JoinCommunityRequest | JoinCommunityResponse |
| LeaveCommunity | LeaveCommunityRequest | LeaveCommunityResponse |
| ListCertifications | ListCertificationsRequest | ListCertificationsResponse |
| ListUsers | ListUsersRequest | ListUsersResponse |
| ReviewCertification | ReviewCertificationRequest | ReviewCertificationResponse |
| SubmitCertification | SubmitCertificationRequest | SubmitCertificationResponse |
| UpdateUser | UpdateUserRequest | UpdateUserResponse |
| UpdateUserModerationStatus | UpdateModerationStatusRequest | UpdateModerationStatusResponse |

## 数据库表

| 表名 | 列 |
|------|-----|
| user_base | deleted_at (nullable), updated_at (datetime), created_at (datetime), nickname_moderation_status (bigint), delete_time (nullable), updated_time (datetime), created_time (datetime), preferences (nullable), credit_score (bigint), status (bigint), birth_date (nullable), gender (nullable), id_card_number (nullable), real_name (nullable), avatar_url (nullable) ... |
| user_certification | moderation_time (nullable), moderation_status (bigint), submit_time (datetime), review_notes (nullable), review_time (nullable), reviewer_id (nullable), status (bigint), document_urls (nullable), user_id (bigint), role_id (bigint), id (bigint) |
| user_community_membership | updated_at (datetime), created_at (datetime), room (bigint), unit (bigint), building (bigint), updated_time (datetime), created_time (datetime), leave_time (nullable), join_time (datetime), bind_status (bigint), community_id (bigint), user_id (bigint), id (bigint) |
| user_membership_role | updated_time (datetime), created_time (datetime), expires_at (nullable), verified_at (nullable), verf_status (bigint), role_code (varchar), community_id (bigint), membership_id (nullable), user_id (bigint), id (bigint) |
| user_residence | updated_at (datetime), created_at (datetime), updated_time (datetime), created_time (datetime), end_date (nullable), start_date (nullable), is_primary (bigint), room (varchar), unit (varchar), building (varchar), house_id (varchar), user_id (bigint), membership_id (bigint), id (bigint) |

## 前端消费方

| 方法 | URL | 文件 |
|------|-----|------|
| GET | /api/verifications | web/pc/src/api/identity.ts |
| GET | /api/users/roles | web/mobile/src/api/user.ts |
| POST | /api/users/roles/apply | web/mobile/src/api/user.ts |
| POST | /api/users/residences/bind | web/mobile/src/api/user.ts |
| POST | /api/users/communities/leave | web/mobile/src/api/user.ts |
| GET | /api/users/communities/memberships | web/mobile/src/api/user.ts |
| POST | /api/users/communities/join | web/mobile/src/api/user.ts |
| POST | /api/users | web/pc/src/api/identity.ts |
| GET | /api/users | web/pc/src/api/identity.ts |
| GET | /api/users/profile | web/mobile/src/api/identity.ts |

## 实体血缘（Proto → Go → DB）

| Proto 消息 | Go 结构体 | 数据库表 |
|-----------|----------|---------|
| CommunityMembership | CommunityMembership | - |
| Residence | Residence | - |

---
*此文件由 graph-sync.sh 自动生成，请勿手动编辑。*
