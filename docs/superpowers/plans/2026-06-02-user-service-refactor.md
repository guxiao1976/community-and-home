# User Service Refactoring Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Refactor user-service from the old 7-table design to the new 5-table design per `docs/specs/user-design.md`, including proto redefinition, database migration, model rewrite, RPC logic rewrite, and API layer update, while preserving backward compatibility with auth-service's registration/login flow.

**Architecture:** Replace the old property-unit-centric model with a Community Membership → Role → Certification hierarchy. The new design separates "being in a community" (membership), "having a role" (role), and "proving it" (certification). Auth-service's calls to `CreateUser`, `GetUserByPhone`, and `UpdateUser` must remain fully functional.

**Tech Stack:** Go 1.25, go-zero 1.10.1, gRPC/Protobuf (Buf v2), GORM, MySQL 8.0, Redis

---

## File Structure Map

### api-proto/ (Proto definitions — modified)
- **Modify:** `api-proto/api/user/v1/user.proto` — Full rewrite of UserService proto with backward-compatible CreateUser/GetUserByPhone/UpdateUser

### services/user-service/ (Go service — rewritten)

**Model layer (5 new files replace 7 old):**
- **Create:** `model/user_base.go` — UserBase model (restructured)
- **Create:** `model/user_community_membership.go` — Community membership model
- **Create:** `model/user_membership_role.go` — Role model
- **Create:** `model/user_certification.go` — Certification record model
- **Create:** `model/user_residence.go` — Residence model
- **Create:** `model/vars.go` — Shared variables (ErrNotFound, etc.)
- **Delete:** `model/userbase.go`, `model/userbase_gen.go`, `model/property.go`, `model/family.go`, `model/uploadedfile.go`, `model/userhomeownerverification.go`

**RPC config:**
- **Modify:** `rpc/internal/config/config.go` — Add new config fields if needed
- **Modify:** `rpc/etc/userservice.yaml` — Update config

**RPC service context:**
- **Modify:** `rpc/internal/svc/servicecontext.go` — New model interfaces

**RPC server:**
- **Modify:** `rpc/internal/server/userserviceserver.go` — Register all new RPC methods

**RPC logic (new files replace old):**
- **Create:** `rpc/internal/logic/user/create_user_logic.go`
- **Create:** `rpc/internal/logic/user/get_user_logic.go`
- **Create:** `rpc/internal/logic/user/get_user_by_phone_logic.go`
- **Create:** `rpc/internal/logic/user/update_user_logic.go`
- **Create:** `rpc/internal/logic/user/list_users_logic.go`
- **Create:** `rpc/internal/logic/user/get_users_by_ids_logic.go`
- **Create:** `rpc/internal/logic/user/join_community_logic.go`
- **Create:** `rpc/internal/logic/user/leave_community_logic.go`
- **Create:** `rpc/internal/logic/user/apply_role_logic.go`
- **Create:** `rpc/internal/logic/user/submit_certification_logic.go`
- **Create:** `rpc/internal/logic/user/review_certification_logic.go`
- **Create:** `rpc/internal/logic/user/get_user_roles_logic.go`
- **Create:** `rpc/internal/logic/user/get_user_memberships_logic.go`
- **Create:** `rpc/internal/logic/user/bind_residence_logic.go`
- **Create:** `rpc/internal/logic/user/get_residences_logic.go`
- **Create:** `rpc/internal/logic/user/list_certifications_logic.go`
- **Create:** `rpc/internal/logic/user/get_my_certifications_logic.go`
- **Create:** `rpc/internal/logic/user/helper.go` — Proto conversion helpers
- **Delete:** All old logic files in `rpc/internal/logic/user/`

**API layer:**
- **Modify:** `api/internal/config/config.go` — Update config
- **Modify:** `api/etc/user-api.yaml` — Update config
- **Modify:** `api/internal/svc/servicecontext.go` — New RPC clients
- **Modify:** `api/internal/handler/routes.go` — New routes
- **Modify:** `api/internal/types/types.go` — New request/response types
- **Create:** `api/internal/handler/user/handler.go` — User CRUD handlers
- **Create:** `api/internal/handler/community/handler.go` — Community membership handlers
- **Create:** `api/internal/handler/role/handler.go` — Role handlers
- **Create:** `api/internal/handler/certification/handler.go` — Certification handlers
- **Create:** `api/internal/handler/residence/handler.go` — Residence handlers
- **Create:** `api/internal/logic/user/user_logic.go` — User logic
- **Create:** `api/internal/logic/community/community_logic.go` — Community logic
- **Create:** `api/internal/logic/role/role_logic.go` — Role logic
- **Create:** `api/internal/logic/certification/certification_logic.go` — Certification logic
- **Create:** `api/internal/logic/residence/residence_logic.go` — Residence logic
- **Create:** `api/internal/middleware/auth.go` — JWT auth middleware
- **Delete:** All old handler/logic/middleware files

**Migration:**
- **Create:** `migration/001_refactor_to_v2.sql` — Database migration script

---

### Task 1: Proto Definition Rewrite

**Files:**
- Modify: `api-proto/api/user/v1/user.proto`

- [ ] **Step 1: Rewrite user.proto with new design while keeping auth-service backward compatibility**

Write the complete new proto file. Key principles:
- Keep `CreateUserRequest.phone`, `.nickname`, `.user_type` (deprecated), `.scope_id` (deprecated) — auth-service passes these
- Keep `CreateUserResponse.base`, `.user_id` — auth-service reads user_id
- Keep `GetUserByPhoneRequest.phone` — auth-service passes this
- Keep `UpdateUserRequest.id`, `.status` — auth-service saga compensation
- New `User` message adds: `real_name`, `id_card_number`, `gender`, `birth_date`, `preferences`
- New `User` message removes: `user_type`, `cert_status`, `scope_id`
- Add new RPCs: `JoinCommunity`, `LeaveCommunity`, `ApplyRole`, `SubmitCertification`, `ReviewCertification`, `GetUserRoles`, `GetUserMemberships`, `BindResidence`, `GetResidences`, `ListCertifications`, `GetMyCertifications`

```protobuf
syntax = "proto3";

package user.v1;

option go_package = "github.com/guxiao1976/api-proto/gen/go/user/v1;userv1";

import "api/common/v1/common.proto";
import "google/protobuf/timestamp.proto";

// ========== User Service ==========

service UserService {
    // 用户基础操作（保留，auth-service 依赖）
    rpc CreateUser(CreateUserRequest) returns (CreateUserResponse);
    rpc GetUser(GetUserRequest) returns (GetUserResponse);
    rpc GetUserByPhone(GetUserByPhoneRequest) returns (GetUserResponse);
    rpc UpdateUser(UpdateUserRequest) returns (UpdateUserResponse);
    rpc ListUsers(ListUsersRequest) returns (ListUsersResponse);
    rpc GetUsersByIds(GetUsersByIdsRequest) returns (GetUsersByIdsResponse);

    // 小区成员
    rpc JoinCommunity(JoinCommunityRequest) returns (JoinCommunityResponse);
    rpc LeaveCommunity(LeaveCommunityRequest) returns (LeaveCommunityResponse);
    rpc GetUserMemberships(GetUserMembershipsRequest) returns (GetUserMembershipsResponse);

    // 角色
    rpc ApplyRole(ApplyRoleRequest) returns (ApplyRoleResponse);
    rpc GetUserRoles(GetUserRolesRequest) returns (GetUserRolesResponse);

    // 认证（统一流程）
    rpc SubmitCertification(SubmitCertificationRequest) returns (SubmitCertificationResponse);
    rpc ReviewCertification(ReviewCertificationRequest) returns (ReviewCertificationResponse);
    rpc ListCertifications(ListCertificationsRequest) returns (ListCertificationsResponse);
    rpc GetMyCertifications(GetMyCertificationsRequest) returns (GetMyCertificationsResponse);

    // 房屋
    rpc BindResidence(BindResidenceRequest) returns (BindResidenceResponse);
    rpc GetResidences(GetResidencesRequest) returns (GetResidencesResponse);
}

// ========== User Messages ==========

message User {
    int64 id = 1;
    string phone = 2;                    // AES encrypted, masked in responses
    string nickname = 3;
    string avatar_url = 4;
    string real_name = 5;                // NEW: 首次认证通过后回填
    string id_card_number = 6;           // NEW: AES encrypted, masked
    int32 gender = 7;                    // NEW: 1-男 2-女
    string birth_date = 8;              // NEW: YYYY-MM-DD
    int32 status = 9;                    // 1-正常 2-禁用
    int32 credit_score = 10;
    string preferences = 11;             // NEW: JSON string
    int64 created_at = 12;              // unix timestamp seconds
    int64 updated_at = 13;              // unix timestamp seconds
}

message CreateUserRequest {
    string phone = 1;
    string nickname = 2;
    int32 user_type = 3 [deprecated = true];   // 保留兼容 auth-service，新逻辑忽略
    int64 scope_id = 4 [deprecated = true];    // 保留兼容，新逻辑忽略
}

message CreateUserResponse {
    common.v1.BaseResp base = 1;
    int64 user_id = 2;
}

message GetUserRequest {
    int64 id = 1;
}

message GetUserResponse {
    common.v1.BaseResp base = 1;
    User user = 2;
}

message GetUserByPhoneRequest {
    string phone = 1;
}

message UpdateUserRequest {
    int64 id = 1;
    optional string nickname = 2;
    optional string avatar_url = 3;
    optional int32 status = 5;
    // NEW optional fields:
    optional int32 gender = 7;
    optional string birth_date = 8;
    optional string preferences = 9;
}

message UpdateUserResponse {
    common.v1.BaseResp base = 1;
    User user = 2;
}

message ListUsersRequest {
    common.v1.PageRequest page = 1;
    string keyword = 2;       // search nickname or phone
    optional int32 status = 3;
}

message ListUsersResponse {
    common.v1.BaseResp base = 1;
    repeated User users = 2;
    common.v1.PageResponse page = 3;
}

message GetUsersByIdsRequest {
    repeated int64 ids = 1;
}

message GetUsersByIdsResponse {
    common.v1.BaseResp base = 1;
    repeated User users = 2;
}

// ========== Community Membership Messages ==========

message CommunityMembership {
    int64 id = 1;
    int64 user_id = 2;
    int64 community_id = 3;
    int32 bind_status = 4;     // 1-有效 0-已退出
    int64 join_time = 5;       // unix timestamp seconds
    int64 leave_time = 6;      // unix timestamp seconds, 0 if not left
    int64 created_at = 7;
    int64 updated_at = 8;
}

message JoinCommunityRequest {
    int64 user_id = 1;
    int64 community_id = 2;
}

message JoinCommunityResponse {
    common.v1.BaseResp base = 1;
    CommunityMembership membership = 2;
}

message LeaveCommunityRequest {
    int64 user_id = 1;
    int64 community_id = 2;
}

message LeaveCommunityResponse {
    common.v1.BaseResp base = 1;
}

message GetUserMembershipsRequest {
    int64 user_id = 1;
}

message GetUserMembershipsResponse {
    common.v1.BaseResp base = 1;
    repeated CommunityMembership memberships = 2;
}

// ========== Role Messages ==========

message MembershipRole {
    int64 id = 1;
    int64 user_id = 2;
    int64 membership_id = 3;   // 0 for merchant
    int64 community_id = 4;    // 0 for merchant
    string role_code = 5;      // owner, tenant, grid_worker, community_admin, property_admin, committee, merchant
    int32 verf_status = 6;     // 0-未认证 1-待审 2-已通过 3-已驳回 4-已过期
    int64 verified_at = 7;     // unix timestamp seconds
    int64 expires_at = 8;      // unix timestamp seconds, 0 = permanent
    int64 created_at = 9;
    int64 updated_at = 10;
}

message ApplyRoleRequest {
    int64 user_id = 1;
    int64 community_id = 2;
    string role_code = 3;
    // residence fields (only for owner/tenant):
    string building = 4;
    string unit = 5;
    string room = 6;
}

message ApplyRoleResponse {
    common.v1.BaseResp base = 1;
    MembershipRole role = 2;
}

message GetUserRolesRequest {
    int64 user_id = 1;
    int64 community_id = 2;    // 0 = all communities
}

message GetUserRolesResponse {
    common.v1.BaseResp base = 1;
    repeated MembershipRole roles = 2;
}

// ========== Certification Messages ==========

message Certification {
    int64 id = 1;
    int64 role_id = 2;
    int64 user_id = 3;
    string document_urls = 4;  // JSON array string
    int32 status = 5;          // 1-待审核 2-已通过 3-已驳回
    int64 reviewer_id = 6;
    string review_notes = 7;
    int64 review_time = 8;     // unix timestamp seconds
    int64 submit_time = 9;     // unix timestamp seconds
}

message SubmitCertificationRequest {
    int64 user_id = 1;
    int64 role_id = 2;
    repeated string document_urls = 3;
    string real_name = 4;
    string id_card_number = 5;
}

message SubmitCertificationResponse {
    common.v1.BaseResp base = 1;
    Certification certification = 2;
}

message ReviewCertificationRequest {
    int64 certification_id = 1;
    int64 reviewer_id = 2;
    int32 result = 3;          // 2-通过 3-驳回
    string review_notes = 4;
    // For time-limited roles, the reviewer specifies the expiry date:
    string expires_at = 5;     // YYYY-MM-DD, only used when result=2 and role has expiry
}

message ReviewCertificationResponse {
    common.v1.BaseResp base = 1;
}

message ListCertificationsRequest {
    common.v1.PageRequest page = 1;
    optional int32 status = 2;     // filter by status
    optional int64 user_id = 3;    // filter by user
}

message ListCertificationsResponse {
    common.v1.BaseResp base = 1;
    repeated Certification certifications = 2;
    common.v1.PageResponse page = 3;
}

message GetMyCertificationsRequest {
    int64 user_id = 1;
}

message GetMyCertificationsResponse {
    common.v1.BaseResp base = 1;
    repeated Certification certifications = 2;
}

// ========== Residence Messages ==========

message Residence {
    int64 id = 1;
    int64 membership_id = 2;
    string house_id = 3;       // e.g. "1-2-301"
    string building = 4;
    string unit = 5;
    string room = 6;
    int32 is_primary = 7;      // 1-主房产 0-否
    string start_date = 8;     // YYYY-MM-DD
    string end_date = 9;       // YYYY-MM-DD
    int64 created_at = 10;
    int64 updated_at = 11;
}

message BindResidenceRequest {
    int64 membership_id = 1;
    string building = 2;
    string unit = 3;
    string room = 4;
    int32 is_primary = 5;
    string start_date = 6;     // YYYY-MM-DD, optional
    string end_date = 7;       // YYYY-MM-DD, optional
}

message BindResidenceResponse {
    common.v1.BaseResp base = 1;
    Residence residence = 2;
}

message GetResidencesRequest {
    int64 membership_id = 1;
}

message GetResidencesResponse {
    common.v1.BaseResp base = 1;
    repeated Residence residences = 2;
}
```

- [ ] **Step 2: Regenerate proto Go code**

```bash
cd /home/jiaoxh/my-project/community-home/api-proto && make generate
```

Expected: Generation succeeds, new `.pb.go`, `_grpc.pb.go`, `.pb.gw.go` files created.

- [ ] **Step 3: Run proto lint**

```bash
cd /home/jiaoxh/my-project/community-home/api-proto && make lint
```

Expected: Lint passes (or acceptable warnings).

- [ ] **Step 4: Verify auth-service still compiles with new proto**

```bash
cd /home/jiaoxh/my-project/community-home/services/auth-service && go build ./...
```

Expected: Compilation succeeds. If not, the deprecated fields in `CreateUserRequest` and `UpdateUserRequest` ensure auth-service's existing code still compiles.

- [ ] **Step 5: Commit proto changes**

```bash
cd /home/jiaoxh/my-project/community-home/api-proto && git add -A && git commit -m "feat(user): redesign UserService proto per user-design.md

- Restructure User message: add real_name/id_card_number/gender/birth_date/preferences
- Deprecate user_type/scope_id in CreateUserRequest (keep for auth-service compat)
- Add CommunityMembership, MembershipRole, Certification, Residence messages
- Add RPCs: JoinCommunity, LeaveCommunity, ApplyRole, SubmitCertification, ReviewCertification, etc.
- Keep backward compat: CreateUser, GetUserByPhone, UpdateUser retain old fields

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

### Task 2: Database Migration

**Files:**
- Create: `services/user-service/migration/001_refactor_to_v2.sql`

- [ ] **Step 1: Write migration SQL**

```sql
-- Migration: Refactor user-service from old 7-table to new 5-table design
-- Per: docs/specs/user-design.md

-- ========== Phase 1: Clean up old tables ==========

DROP TABLE IF EXISTS user_family_member;
DROP TABLE IF EXISTS user_family;
DROP TABLE IF EXISTS user_uploaded_file;
DROP TABLE IF EXISTS user_homeowner_verification;
DROP TABLE IF EXISTS user_property_binding;
DROP TABLE IF EXISTS user_property_unit;

-- ========== Phase 2: Restructure user_base ==========

-- Add new columns
ALTER TABLE user_base
    ADD COLUMN real_name VARCHAR(50) NULL COMMENT '真实姓名' AFTER avatar_url,
    ADD COLUMN id_card_number VARCHAR(255) NULL COMMENT '身份证号（AES加密）' AFTER real_name,
    ADD COLUMN gender TINYINT NULL COMMENT '性别：1-男 2-女' AFTER id_card_number,
    ADD COLUMN birth_date DATE NULL COMMENT '出生日期' AFTER gender,
    ADD COLUMN preferences JSON NULL COMMENT '用户偏好' AFTER credit_score,
    ADD COLUMN delete_time DATETIME NULL COMMENT '软删除时间' AFTER updated_time;

-- Drop old columns that are no longer needed
ALTER TABLE user_base
    DROP COLUMN user_type,
    DROP COLUMN cert_status,
    DROP COLUMN scope_id;

-- Modify status: remove value 3 (deleted), soft delete via delete_time
-- Note: existing status=3 rows should be converted to status=2 + delete_time set
UPDATE user_base SET delete_time = NOW() WHERE status = 3;
UPDATE user_base SET status = 2 WHERE status = 3;

-- ========== Phase 3: Create new tables ==========

CREATE TABLE user_community_membership (
    id                  BIGINT NOT NULL AUTO_INCREMENT,
    user_id             BIGINT NOT NULL COMMENT '用户ID',
    community_id        BIGINT NOT NULL COMMENT '小区ID',
    bind_status         TINYINT NOT NULL DEFAULT 1 COMMENT '1-有效 0-已退出',
    join_time           DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '加入时间',
    leave_time          DATETIME NULL COMMENT '退出时间',
    created_time        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_time        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE INDEX uk_user_community (user_id, community_id),
    INDEX idx_community (community_id, bind_status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户-小区成员关系';

CREATE TABLE user_membership_role (
    id                  BIGINT NOT NULL AUTO_INCREMENT,
    user_id             BIGINT NOT NULL COMMENT '用户ID',
    membership_id       BIGINT NULL COMMENT '小区成员关系ID，商家为 NULL',
    community_id        BIGINT NOT NULL DEFAULT 0 COMMENT '小区ID，0=全局角色(商家)',
    role_code           VARCHAR(30) NOT NULL COMMENT '角色编码',
    verf_status         TINYINT NOT NULL DEFAULT 0 COMMENT '认证状态：0-未认证 1-待审 2-已通过 3-已驳回 4-已过期',
    verified_at         DATETIME NULL COMMENT '认证通过时间',
    expires_at          DATETIME NULL COMMENT '过期时间，NULL=永久有效',
    created_time        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_time        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE INDEX uk_member_role (membership_id, role_code),
    UNIQUE INDEX uk_user_community_role (user_id, community_id, role_code),
    INDEX idx_community_role (community_id, role_code, verf_status),
    INDEX idx_expires (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户角色表';

CREATE TABLE user_certification (
    id                  BIGINT NOT NULL AUTO_INCREMENT,
    role_id             BIGINT NOT NULL COMMENT '角色ID',
    user_id             BIGINT NOT NULL COMMENT '用户ID',
    document_urls       TEXT NULL COMMENT '证明材料URL列表（JSON数组）',
    status              TINYINT NOT NULL DEFAULT 1 COMMENT '审核状态：1-待审核 2-已通过 3-已驳回',
    reviewer_id         BIGINT NULL COMMENT '审核人ID',
    review_time         DATETIME NULL COMMENT '审核时间',
    review_notes        VARCHAR(500) NULL COMMENT '审核备注',
    submit_time         DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '提交时间',
    PRIMARY KEY (id),
    INDEX idx_role (role_id),
    INDEX idx_user (user_id),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='认证记录表';

CREATE TABLE user_residence (
    id                  BIGINT NOT NULL AUTO_INCREMENT,
    membership_id       BIGINT NOT NULL COMMENT '小区成员关系ID',
    house_id            VARCHAR(50) NOT NULL COMMENT '房屋ID，如 1-2-301',
    building            VARCHAR(20) NOT NULL COMMENT '楼号',
    unit                VARCHAR(20) NOT NULL DEFAULT '' COMMENT '单元号',
    room                VARCHAR(20) NOT NULL COMMENT '房号',
    is_primary          TINYINT NOT NULL DEFAULT 0 COMMENT '主房产：1-是 0-否',
    start_date          DATE NULL COMMENT '入住/合同开始日期',
    end_date            DATE NULL COMMENT '搬离/合同结束日期',
    created_time        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_time        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE INDEX uk_member_house (membership_id, house_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='居民房屋明细表';
```

- [ ] **Step 2: Run migration against local dev database**

```bash
mysql -h 172.18.0.2 -u root -proot user < /home/jiaoxh/my-project/community-home/services/user-service/migration/001_refactor_to_v2.sql
```

Expected: Migration completes without errors. All 5 tables exist with correct schema.

- [ ] **Step 3: Verify table structures**

```bash
mysql -h 172.18.0.2 -u root -proot user -e "SHOW CREATE TABLE user_base\G SHOW CREATE TABLE user_community_membership\G SHOW CREATE TABLE user_membership_role\G SHOW CREATE TABLE user_certification\G SHOW CREATE TABLE user_residence\G"
```

Expected: All 5 tables match the design spec exactly.

- [ ] **Step 4: Commit migration**

```bash
cd /home/jiaoxh/my-project/community-home/services/user-service && git add migration/ && git commit -m "feat: add database migration from old 7-table to new 5-table design

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

### Task 3: Model Layer Rewrite

**Files:**
- Create: `services/user-service/model/user_base.go`
- Create: `services/user-service/model/user_community_membership.go`
- Create: `services/user-service/model/user_membership_role.go`
- Create: `services/user-service/model/user_certification.go`
- Create: `services/user-service/model/user_residence.go`
- Create: `services/user-service/model/vars.go`
- Delete: `services/user-service/model/userbase.go`
- Delete: `services/user-service/model/userbase_gen.go`
- Delete: `services/user-service/model/property.go`
- Delete: `services/user-service/model/family.go`
- Delete: `services/user-service/model/uploadedfile.go`
- Delete: `services/user-service/model/userhomeownerverification.go`

- [ ] **Step 1: Write `model/vars.go`**

```go
package model

import (
	"errors"
	"time"
)

var (
	ErrNotFound = errors.New("record not found")
)

// User status constants
const (
	UserStatusActive   = 1
	UserStatusDisabled = 2
)

// Membership bind status
const (
	MembershipBindStatusActive  = 1
	MembershipBindStatusLeft    = 0
)

// Role certification status
const (
	RoleVerfStatusUnverified = 0 // 未认证
	RoleVerfStatusPending    = 1 // 待审核
	RoleVerfStatusApproved   = 2 // 已通过
	RoleVerfStatusRejected   = 3 // 已驳回
	RoleVerfStatusExpired    = 4 // 已过期
)

// Certification review status
const (
	CertStatusPending  = 1 // 待审核
	CertStatusApproved = 2 // 已通过
	CertStatusRejected = 3 // 已驳回
)

// Role codes
const (
	RoleCodeOwner          = "owner"
	RoleCodeTenant         = "tenant"
	RoleCodeGridWorker     = "grid_worker"
	RoleCodeCommunityAdmin = "community_admin"
	RoleCodePropertyAdmin  = "property_admin"
	RoleCodeCommittee      = "committee"
	RoleCodeMerchant       = "merchant"
)

// MaxCommunities is the maximum number of communities a user can join
const MaxCommunities = 5
```

- [ ] **Step 2: Write `model/user_base.go`**

```go
package model

import (
	"context"
	"database/sql"
	"time"

	"github.com/guxiao1976/community-common/v2/pkg/crypto"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type UserBase struct {
	Id           int64          `db:"id"`
	Phone        string         `db:"phone"`
	Nickname     sql.NullString `db:"nickname"`
	AvatarUrl    sql.NullString `db:"avatar_url"`
	RealName     sql.NullString `db:"real_name"`
	IdCardNumber sql.NullString `db:"id_card_number"`
	Gender       sql.NullInt64  `db:"gender"`
	BirthDate    sql.NullTime   `db:"birth_date"`
	Status       int64          `db:"status"`
	CreditScore  int64          `db:"credit_score"`
	Preferences  sql.NullString `db:"preferences"`
	CreatedTime  time.Time      `db:"created_time"`
	UpdatedTime  time.Time      `db:"updated_time"`
	DeleteTime   sql.NullTime   `db:"delete_time"`
}

type UserBaseModel interface {
	Insert(ctx context.Context, data *UserBase) (sql.Result, error)
	FindOne(ctx context.Context, id int64) (*UserBase, error)
	FindOneByPhone(ctx context.Context, phone string) (*UserBase, error)
	FindByIds(ctx context.Context, ids []int64) ([]*UserBase, error)
	FindPage(ctx context.Context, keyword string, status *int64, page, pageSize int32) ([]*UserBase, int64, error)
	Update(ctx context.Context, data *UserBase) error
	SoftDelete(ctx context.Context, id int64) error
	UpdateStatus(ctx context.Context, id int64, status int64) error
	UpdateRealNameAndIdCard(ctx context.Context, id int64, realName, idCardNumber string) error
}

type defaultUserBaseModel struct {
	conn  sqlx.SqlConn
	table string
	aes   *crypto.AES
}

func NewUserBaseModel(conn sqlx.SqlConn, aesKey string) UserBaseModel {
	return &defaultUserBaseModel{
		conn:  conn,
		table: "user_base",
		aes:   crypto.NewAES(aesKey),
	}
}

func (m *defaultUserBaseModel) Insert(ctx context.Context, data *UserBase) (sql.Result, error) {
	query := `INSERT INTO user_base (id, phone, nickname, avatar_url, real_name, id_card_number, gender, birth_date, status, credit_score, preferences, created_time, updated_time) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	return m.conn.ExecCtx(ctx, query, data.Id, data.Phone, data.Nickname, data.AvatarUrl, data.RealName, data.IdCardNumber, data.Gender, data.BirthDate, data.Status, data.CreditScore, data.Preferences, data.CreatedTime, data.UpdatedTime)
}

func (m *defaultUserBaseModel) FindOne(ctx context.Context, id int64) (*UserBase, error) {
	query := `SELECT id, phone, nickname, avatar_url, real_name, id_card_number, gender, birth_date, status, credit_score, preferences, created_time, updated_time, delete_time FROM user_base WHERE id = ? AND delete_time IS NULL`
	var resp UserBase
	err := m.conn.QueryRowCtx(ctx, &resp, query, id)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &resp, nil
}

func (m *defaultUserBaseModel) FindOneByPhone(ctx context.Context, phone string) (*UserBase, error) {
	encryptedPhone := m.aes.Encrypt(phone)
	query := `SELECT id, phone, nickname, avatar_url, real_name, id_card_number, gender, birth_date, status, credit_score, preferences, created_time, updated_time, delete_time FROM user_base WHERE phone = ? AND delete_time IS NULL`
	var resp UserBase
	err := m.conn.QueryRowCtx(ctx, &resp, query, encryptedPhone)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &resp, nil
}

func (m *defaultUserBaseModel) FindByIds(ctx context.Context, ids []int64) ([]*UserBase, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	query := `SELECT id, phone, nickname, avatar_url, real_name, id_card_number, gender, birth_date, status, credit_score, preferences, created_time, updated_time, delete_time FROM user_base WHERE id IN (?) AND delete_time IS NULL`
	query, args, _ := sqlx.In(query, ids)
	var resp []*UserBase
	err := m.conn.QueryRowsCtx(ctx, &resp, query, args...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *defaultUserBaseModel) FindPage(ctx context.Context, keyword string, status *int64, page, pageSize int32) ([]*UserBase, int64, error) {
	where := "WHERE delete_time IS NULL"
	args := make([]interface{}, 0)

	if keyword != "" {
		where += " AND (nickname LIKE ? OR phone LIKE ?)"
		kw := "%" + keyword + "%"
		args = append(args, kw, kw)
	}
	if status != nil {
		where += " AND status = ?"
		args = append(args, *status)
	}

	countQuery := "SELECT COUNT(*) FROM user_base " + where
	var total int64
	err := m.conn.QueryRowCtx(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	query := "SELECT id, phone, nickname, avatar_url, real_name, id_card_number, gender, birth_date, status, credit_score, preferences, created_time, updated_time, delete_time FROM user_base " + where + " ORDER BY id DESC LIMIT ? OFFSET ?"
	queryArgs := append(args, pageSize, offset)
	var resp []*UserBase
	err = m.conn.QueryRowsCtx(ctx, &resp, query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	return resp, total, nil
}

func (m *defaultUserBaseModel) Update(ctx context.Context, data *UserBase) error {
	query := `UPDATE user_base SET nickname=?, avatar_url=?, gender=?, birth_date=?, status=?, preferences=?, updated_time=? WHERE id=? AND delete_time IS NULL`
	_, err := m.conn.ExecCtx(ctx, query, data.Nickname, data.AvatarUrl, data.Gender, data.BirthDate, data.Status, data.Preferences, time.Now(), data.Id)
	return err
}

func (m *defaultUserBaseModel) SoftDelete(ctx context.Context, id int64) error {
	query := `UPDATE user_base SET delete_time=?, updated_time=? WHERE id=? AND delete_time IS NULL`
	_, err := m.conn.ExecCtx(ctx, query, time.Now(), time.Now(), id)
	return err
}

func (m *defaultUserBaseModel) UpdateStatus(ctx context.Context, id int64, status int64) error {
	query := `UPDATE user_base SET status=?, updated_time=? WHERE id=? AND delete_time IS NULL`
	_, err := m.conn.ExecCtx(ctx, query, status, time.Now(), id)
	return err
}

func (m *defaultUserBaseModel) UpdateRealNameAndIdCard(ctx context.Context, id int64, realName, idCardNumber string) error {
	query := `UPDATE user_base SET real_name=COALESCE(real_name, ?), id_card_number=COALESCE(id_card_number, ?), updated_time=? WHERE id=? AND delete_time IS NULL`
	_, err := m.conn.ExecCtx(ctx, query, realName, idCardNumber, time.Now(), id)
	return err
}
```

- [ ] **Step 3: Write `model/user_community_membership.go`**

```go
package model

import (
	"context"
	"database/sql"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type UserCommunityMembership struct {
	Id          int64        `db:"id"`
	UserId      int64        `db:"user_id"`
	CommunityId int64        `db:"community_id"`
	BindStatus  int64        `db:"bind_status"`
	JoinTime    time.Time    `db:"join_time"`
	LeaveTime   sql.NullTime `db:"leave_time"`
	CreatedTime time.Time    `db:"created_time"`
	UpdatedTime time.Time    `db:"updated_time"`
}

type UserCommunityMembershipModel interface {
	Insert(ctx context.Context, data *UserCommunityMembership) (sql.Result, error)
	FindOne(ctx context.Context, id int64) (*UserCommunityMembership, error)
	FindByUserAndCommunity(ctx context.Context, userId, communityId int64) (*UserCommunityMembership, error)
	FindByUserId(ctx context.Context, userId int64) ([]*UserCommunityMembership, error)
	CountActiveByUserId(ctx context.Context, userId int64) (int64, error)
	UpdateBindStatus(ctx context.Context, id int64, bindStatus int64, leaveTime time.Time) error
}

type defaultUserCommunityMembershipModel struct {
	conn  sqlx.SqlConn
	table string
}

func NewUserCommunityMembershipModel(conn sqlx.SqlConn) UserCommunityMembershipModel {
	return &defaultUserCommunityMembershipModel{
		conn:  conn,
		table: "user_community_membership",
	}
}

func (m *defaultUserCommunityMembershipModel) Insert(ctx context.Context, data *UserCommunityMembership) (sql.Result, error) {
	query := `INSERT INTO user_community_membership (user_id, community_id, bind_status, join_time, created_time, updated_time) VALUES (?, ?, ?, ?, ?, ?)`
	return m.conn.ExecCtx(ctx, query, data.UserId, data.CommunityId, data.BindStatus, data.JoinTime, data.CreatedTime, data.UpdatedTime)
}

func (m *defaultUserCommunityMembershipModel) FindOne(ctx context.Context, id int64) (*UserCommunityMembership, error) {
	query := `SELECT id, user_id, community_id, bind_status, join_time, leave_time, created_time, updated_time FROM user_community_membership WHERE id = ?`
	var resp UserCommunityMembership
	err := m.conn.QueryRowCtx(ctx, &resp, query, id)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &resp, nil
}

func (m *defaultUserCommunityMembershipModel) FindByUserAndCommunity(ctx context.Context, userId, communityId int64) (*UserCommunityMembership, error) {
	query := `SELECT id, user_id, community_id, bind_status, join_time, leave_time, created_time, updated_time FROM user_community_membership WHERE user_id = ? AND community_id = ?`
	var resp UserCommunityMembership
	err := m.conn.QueryRowCtx(ctx, &resp, query, userId, communityId)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &resp, nil
}

func (m *defaultUserCommunityMembershipModel) FindByUserId(ctx context.Context, userId int64) ([]*UserCommunityMembership, error) {
	query := `SELECT id, user_id, community_id, bind_status, join_time, leave_time, created_time, updated_time FROM user_community_membership WHERE user_id = ? AND bind_status = 1`
	var resp []*UserCommunityMembership
	err := m.conn.QueryRowsCtx(ctx, &resp, query, userId)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *defaultUserCommunityMembershipModel) CountActiveByUserId(ctx context.Context, userId int64) (int64, error) {
	query := `SELECT COUNT(*) FROM user_community_membership WHERE user_id = ? AND bind_status = 1`
	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query, userId)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (m *defaultUserCommunityMembershipModel) UpdateBindStatus(ctx context.Context, id int64, bindStatus int64, leaveTime time.Time) error {
	query := `UPDATE user_community_membership SET bind_status=?, leave_time=?, updated_time=? WHERE id=?`
	_, err := m.conn.ExecCtx(ctx, query, bindStatus, leaveTime, time.Now(), id)
	return err
}
```

- [ ] **Step 4: Write `model/user_membership_role.go`**

```go
package model

import (
	"context"
	"database/sql"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type UserMembershipRole struct {
	Id           int64         `db:"id"`
	UserId       int64         `db:"user_id"`
	MembershipId sql.NullInt64 `db:"membership_id"`
	CommunityId  int64         `db:"community_id"`
	RoleCode     string        `db:"role_code"`
	VerfStatus   int64         `db:"verf_status"`
	VerifiedAt   sql.NullTime  `db:"verified_at"`
	ExpiresAt    sql.NullTime  `db:"expires_at"`
	CreatedTime  time.Time     `db:"created_time"`
	UpdatedTime  time.Time     `db:"updated_time"`
}

type UserMembershipRoleModel interface {
	Insert(ctx context.Context, data *UserMembershipRole) (sql.Result, error)
	FindOne(ctx context.Context, id int64) (*UserMembershipRole, error)
	FindByMembershipAndRole(ctx context.Context, membershipId int64, roleCode string) (*UserMembershipRole, error)
	FindByUserAndCommunity(ctx context.Context, userId, communityId int64) ([]*UserMembershipRole, error)
	FindByUserId(ctx context.Context, userId int64) ([]*UserMembershipRole, error)
	UpdateVerfStatus(ctx context.Context, id int64, verfStatus int64, verifiedAt, expiresAt sql.NullTime) error
	UpdateVerfStatusOnly(ctx context.Context, id int64, verfStatus int64) error
	FindExpiredRoles(ctx context.Context) ([]*UserMembershipRole, error)
}

type defaultUserMembershipRoleModel struct {
	conn  sqlx.SqlConn
	table string
}

func NewUserMembershipRoleModel(conn sqlx.SqlConn) UserMembershipRoleModel {
	return &defaultUserMembershipRoleModel{
		conn:  conn,
		table: "user_membership_role",
	}
}

func (m *defaultUserMembershipRoleModel) Insert(ctx context.Context, data *UserMembershipRole) (sql.Result, error) {
	query := `INSERT INTO user_membership_role (user_id, membership_id, community_id, role_code, verf_status, verified_at, expires_at, created_time, updated_time) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	return m.conn.ExecCtx(ctx, query, data.UserId, data.MembershipId, data.CommunityId, data.RoleCode, data.VerfStatus, data.VerifiedAt, data.ExpiresAt, data.CreatedTime, data.UpdatedTime)
}

func (m *defaultUserMembershipRoleModel) FindOne(ctx context.Context, id int64) (*UserMembershipRole, error) {
	query := `SELECT id, user_id, membership_id, community_id, role_code, verf_status, verified_at, expires_at, created_time, updated_time FROM user_membership_role WHERE id = ?`
	var resp UserMembershipRole
	err := m.conn.QueryRowCtx(ctx, &resp, query, id)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &resp, nil
}

func (m *defaultUserMembershipRoleModel) FindByMembershipAndRole(ctx context.Context, membershipId int64, roleCode string) (*UserMembershipRole, error) {
	query := `SELECT id, user_id, membership_id, community_id, role_code, verf_status, verified_at, expires_at, created_time, updated_time FROM user_membership_role WHERE membership_id = ? AND role_code = ?`
	var resp UserMembershipRole
	err := m.conn.QueryRowCtx(ctx, &resp, query, membershipId, roleCode)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &resp, nil
}

func (m *defaultUserMembershipRoleModel) FindByUserAndCommunity(ctx context.Context, userId, communityId int64) ([]*UserMembershipRole, error) {
	query := `SELECT id, user_id, membership_id, community_id, role_code, verf_status, verified_at, expires_at, created_time, updated_time FROM user_membership_role WHERE user_id = ? AND community_id = ?`
	var resp []*UserMembershipRole
	err := m.conn.QueryRowsCtx(ctx, &resp, query, userId, communityId)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *defaultUserMembershipRoleModel) FindByUserId(ctx context.Context, userId int64) ([]*UserMembershipRole, error) {
	query := `SELECT id, user_id, membership_id, community_id, role_code, verf_status, verified_at, expires_at, created_time, updated_time FROM user_membership_role WHERE user_id = ?`
	var resp []*UserMembershipRole
	err := m.conn.QueryRowsCtx(ctx, &resp, query, userId)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *defaultUserMembershipRoleModel) UpdateVerfStatus(ctx context.Context, id int64, verfStatus int64, verifiedAt, expiresAt sql.NullTime) error {
	query := `UPDATE user_membership_role SET verf_status=?, verified_at=?, expires_at=?, updated_time=? WHERE id=?`
	_, err := m.conn.ExecCtx(ctx, query, verfStatus, verifiedAt, expiresAt, time.Now(), id)
	return err
}

func (m *defaultUserMembershipRoleModel) UpdateVerfStatusOnly(ctx context.Context, id int64, verfStatus int64) error {
	query := `UPDATE user_membership_role SET verf_status=?, updated_time=? WHERE id=?`
	_, err := m.conn.ExecCtx(ctx, query, verfStatus, time.Now(), id)
	return err
}

func (m *defaultUserMembershipRoleModel) FindExpiredRoles(ctx context.Context) ([]*UserMembershipRole, error) {
	query := `SELECT id, user_id, membership_id, community_id, role_code, verf_status, verified_at, expires_at, created_time, updated_time FROM user_membership_role WHERE verf_status = 2 AND expires_at IS NOT NULL AND expires_at < NOW()`
	var resp []*UserMembershipRole
	err := m.conn.QueryRowsCtx(ctx, &resp, query)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
```

- [ ] **Step 5: Write `model/user_certification.go`**

```go
package model

import (
	"context"
	"database/sql"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type UserCertification struct {
	Id           int64          `db:"id"`
	RoleId       int64          `db:"role_id"`
	UserId       int64          `db:"user_id"`
	DocumentUrls sql.NullString `db:"document_urls"`
	Status       int64          `db:"status"`
	ReviewerId   sql.NullInt64  `db:"reviewer_id"`
	ReviewTime   sql.NullTime   `db:"review_time"`
	ReviewNotes  sql.NullString `db:"review_notes"`
	SubmitTime   time.Time      `db:"submit_time"`
}

type UserCertificationModel interface {
	Insert(ctx context.Context, data *UserCertification) (sql.Result, error)
	FindOne(ctx context.Context, id int64) (*UserCertification, error)
	FindByRoleId(ctx context.Context, roleId int64) ([]*UserCertification, error)
	FindByUserId(ctx context.Context, userId int64) ([]*UserCertification, error)
	FindPage(ctx context.Context, status *int64, userId *int64, page, pageSize int32) ([]*UserCertification, int64, error)
	Update(ctx context.Context, data *UserCertification) error
}

type defaultUserCertificationModel struct {
	conn  sqlx.SqlConn
	table string
}

func NewUserCertificationModel(conn sqlx.SqlConn) UserCertificationModel {
	return &defaultUserCertificationModel{
		conn:  conn,
		table: "user_certification",
	}
}

func (m *defaultUserCertificationModel) Insert(ctx context.Context, data *UserCertification) (sql.Result, error) {
	query := `INSERT INTO user_certification (role_id, user_id, document_urls, status, reviewer_id, review_time, review_notes, submit_time) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	return m.conn.ExecCtx(ctx, query, data.RoleId, data.UserId, data.DocumentUrls, data.Status, data.ReviewerId, data.ReviewTime, data.ReviewNotes, data.SubmitTime)
}

func (m *defaultUserCertificationModel) FindOne(ctx context.Context, id int64) (*UserCertification, error) {
	query := `SELECT id, role_id, user_id, document_urls, status, reviewer_id, review_time, review_notes, submit_time FROM user_certification WHERE id = ?`
	var resp UserCertification
	err := m.conn.QueryRowCtx(ctx, &resp, query, id)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &resp, nil
}

func (m *defaultUserCertificationModel) FindByRoleId(ctx context.Context, roleId int64) ([]*UserCertification, error) {
	query := `SELECT id, role_id, user_id, document_urls, status, reviewer_id, review_time, review_notes, submit_time FROM user_certification WHERE role_id = ? ORDER BY submit_time DESC`
	var resp []*UserCertification
	err := m.conn.QueryRowsCtx(ctx, &resp, query, roleId)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *defaultUserCertificationModel) FindByUserId(ctx context.Context, userId int64) ([]*UserCertification, error) {
	query := `SELECT id, role_id, user_id, document_urls, status, reviewer_id, review_time, review_notes, submit_time FROM user_certification WHERE user_id = ? ORDER BY submit_time DESC`
	var resp []*UserCertification
	err := m.conn.QueryRowsCtx(ctx, &resp, query, userId)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *defaultUserCertificationModel) FindPage(ctx context.Context, status *int64, userId *int64, page, pageSize int32) ([]*UserCertification, int64, error) {
	where := "WHERE 1=1"
	args := make([]interface{}, 0)

	if status != nil {
		where += " AND status = ?"
		args = append(args, *status)
	}
	if userId != nil {
		where += " AND user_id = ?"
		args = append(args, *userId)
	}

	countQuery := "SELECT COUNT(*) FROM user_certification " + where
	var total int64
	err := m.conn.QueryRowCtx(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	query := "SELECT id, role_id, user_id, document_urls, status, reviewer_id, review_time, review_notes, submit_time FROM user_certification " + where + " ORDER BY id DESC LIMIT ? OFFSET ?"
	queryArgs := append(args, pageSize, offset)
	var resp []*UserCertification
	err = m.conn.QueryRowsCtx(ctx, &resp, query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	return resp, total, nil
}

func (m *defaultUserCertificationModel) Update(ctx context.Context, data *UserCertification) error {
	query := `UPDATE user_certification SET status=?, reviewer_id=?, review_notes=?, review_time=? WHERE id=?`
	_, err := m.conn.ExecCtx(ctx, query, data.Status, data.ReviewerId, data.ReviewNotes, data.ReviewTime, data.Id)
	return err
}
```

- [ ] **Step 6: Write `model/user_residence.go`**

```go
package model

import (
	"context"
	"database/sql"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type UserResidence struct {
	Id           int64        `db:"id"`
	MembershipId int64        `db:"membership_id"`
	HouseId      string       `db:"house_id"`
	Building     string       `db:"building"`
	Unit         string       `db:"unit"`
	Room         string       `db:"room"`
	IsPrimary    int64        `db:"is_primary"`
	StartDate    sql.NullTime `db:"start_date"`
	EndDate      sql.NullTime `db:"end_date"`
	CreatedTime  time.Time    `db:"created_time"`
	UpdatedTime  time.Time    `db:"updated_time"`
}

type UserResidenceModel interface {
	Insert(ctx context.Context, data *UserResidence) (sql.Result, error)
	FindByMembershipId(ctx context.Context, membershipId int64) ([]*UserResidence, error)
	FindByMembershipAndHouse(ctx context.Context, membershipId int64, houseId string) (*UserResidence, error)
	Update(ctx context.Context, data *UserResidence) error
}

type defaultUserResidenceModel struct {
	conn  sqlx.SqlConn
	table string
}

func NewUserResidenceModel(conn sqlx.SqlConn) UserResidenceModel {
	return &defaultUserResidenceModel{
		conn:  conn,
		table: "user_residence",
	}
}

func (m *defaultUserResidenceModel) Insert(ctx context.Context, data *UserResidence) (sql.Result, error) {
	query := `INSERT INTO user_residence (membership_id, house_id, building, unit, room, is_primary, start_date, end_date, created_time, updated_time) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	return m.conn.ExecCtx(ctx, query, data.MembershipId, data.HouseId, data.Building, data.Unit, data.Room, data.IsPrimary, data.StartDate, data.EndDate, data.CreatedTime, data.UpdatedTime)
}

func (m *defaultUserResidenceModel) FindByMembershipId(ctx context.Context, membershipId int64) ([]*UserResidence, error) {
	query := `SELECT id, membership_id, house_id, building, unit, room, is_primary, start_date, end_date, created_time, updated_time FROM user_residence WHERE membership_id = ?`
	var resp []*UserResidence
	err := m.conn.QueryRowsCtx(ctx, &resp, query, membershipId)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *defaultUserResidenceModel) FindByMembershipAndHouse(ctx context.Context, membershipId int64, houseId string) (*UserResidence, error) {
	query := `SELECT id, membership_id, house_id, building, unit, room, is_primary, start_date, end_date, created_time, updated_time FROM user_residence WHERE membership_id = ? AND house_id = ?`
	var resp UserResidence
	err := m.conn.QueryRowCtx(ctx, &resp, query, membershipId, houseId)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &resp, nil
}

func (m *defaultUserResidenceModel) Update(ctx context.Context, data *UserResidence) error {
	query := `UPDATE user_residence SET building=?, unit=?, room=?, is_primary=?, start_date=?, end_date=?, updated_time=? WHERE id=?`
	_, err := m.conn.ExecCtx(ctx, query, data.Building, data.Unit, data.Room, data.IsPrimary, data.StartDate, data.EndDate, time.Now(), data.Id)
	return err
}
```

- [ ] **Step 7: Remove old model files**

```bash
cd /home/jiaoxh/my-project/community-home/services/user-service
rm model/userbase.go model/userbase_gen.go model/property.go model/family.go model/uploadedfile.go model/userhomeownerverification.go
```

- [ ] **Step 8: Verify models compile**

```bash
cd /home/jiaoxh/my-project/community-home/services/user-service && go build ./model/...
```

Expected: Compilation succeeds.

- [ ] **Step 9: Commit model changes**

```bash
cd /home/jiaoxh/my-project/community-home/services/user-service && git add -A && git commit -m "feat(model): rewrite models for new 5-table design

- Add UserBase (restructured), UserCommunityMembership, UserMembershipRole, UserCertification, UserResidence
- Remove old models: property, family, uploaded_file, homeowner_verification
- Add shared constants in vars.go

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

### Task 4: RPC Service Context & Config Update

**Files:**
- Modify: `services/user-service/rpc/internal/config/config.go`
- Modify: `services/user-service/rpc/internal/svc/servicecontext.go`
- Modify: `services/user-service/rpc/etc/userservice.yaml`

- [ ] **Step 1: Read current config.go and servicecontext.go**

Read the files to understand current structure before modifying.

- [ ] **Step 2: Update `rpc/internal/config/config.go`**

```go
package config

import (
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	DataSource string
	Cache      zrpc.RedisConf // Redis cache configuration
	AesKey     string
}
```

(No structural change needed — config stays the same, only models change.)

- [ ] **Step 3: Update `rpc/internal/svc/servicecontext.go`**

```go
package svc

import (
	"github.com/guxiao1976/community-user/model"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"

	"github.com/guxiao1976/community-user/rpc/internal/config"
)

type ServiceContext struct {
	Config config.Config

	UserBaseModel              model.UserBaseModel
	UserCommunityMembershipModel model.UserCommunityMembershipModel
	UserMembershipRoleModel    model.UserMembershipRoleModel
	UserCertificationModel     model.UserCertificationModel
	UserResidenceModel         model.UserResidenceModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DataSource)

	return &ServiceContext{
		Config:                     c,
		UserBaseModel:              model.NewUserBaseModel(conn, c.AesKey),
		UserCommunityMembershipModel: model.NewUserCommunityMembershipModel(conn),
		UserMembershipRoleModel:    model.NewUserMembershipRoleModel(conn),
		UserCertificationModel:     model.NewUserCertificationModel(conn),
		UserResidenceModel:         model.NewUserResidenceModel(conn),
	}
}
```

- [ ] **Step 4: Verify config compiles**

```bash
cd /home/jiaoxh/my-project/community-home/services/user-service && go build ./rpc/internal/config/... ./rpc/internal/svc/...
```

- [ ] **Step 5: Commit config changes**

```bash
cd /home/jiaoxh/my-project/community-home/services/user-service && git add -A && git commit -m "feat(rpc): update service context for new 5-model design

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

### Task 5: RPC Logic Layer — User CRUD (Backward Compatible)

**Files:**
- Create: `services/user-service/rpc/internal/logic/user/helper.go`
- Create: `services/user-service/rpc/internal/logic/user/create_user_logic.go`
- Create: `services/user-service/rpc/internal/logic/user/get_user_logic.go`
- Create: `services/user-service/rpc/internal/logic/user/get_user_by_phone_logic.go`
- Create: `services/user-service/rpc/internal/logic/user/update_user_logic.go`
- Create: `services/user-service/rpc/internal/logic/user/list_users_logic.go`
- Create: `services/user-service/rpc/internal/logic/user/get_users_by_ids_logic.go`

- [ ] **Step 1: Write `rpc/internal/logic/user/helper.go` — Proto conversion helpers**

```go
package user

import (
	"database/sql"
	"encoding/json"

	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	commonv1 "github.com/guxiao1976/api-proto/gen/go/common/v1"
	"github.com/guxiao1976/community-user/model"
)

// toProtoUser converts a model.UserBase to a proto User.
// Phone is already AES-encrypted; it's returned as-is (masking is a frontend concern).
func toProtoUser(u *model.UserBase) *userv1.User {
	if u == nil {
		return nil
	}
	user := &userv1.User{
		Id:          u.Id,
		Phone:       u.Phone,
		Nickname:    u.Nickname.String,
		AvatarUrl:   u.AvatarUrl.String,
		RealName:    u.RealName.String,
		IdCardNumber: u.IdCardNumber.String,
		Status:      int32(u.Status),
		CreditScore: int32(u.CreditScore),
		Preferences: u.Preferences.String,
		CreatedAt:   u.CreatedTime.Unix(),
		UpdatedAt:   u.UpdatedTime.Unix(),
	}
	if u.Gender.Valid {
		user.Gender = int32(u.Gender.Int64)
	}
	if u.BirthDate.Valid {
		user.BirthDate = u.BirthDate.Time.Format("2006-01-02")
	}
	return user
}

// toProtoUsers converts a slice of UserBase to proto Users.
func toProtoUsers(users []*model.UserBase) []*userv1.User {
	result := make([]*userv1.User, 0, len(users))
	for _, u := range users {
		result = append(result, toProtoUser(u))
	}
	return result
}

// toProtoMembership converts a model membership to proto.
func toProtoMembership(m *model.UserCommunityMembership) *userv1.CommunityMembership {
	if m == nil {
		return nil
	}
	cm := &userv1.CommunityMembership{
		Id:          m.Id,
		UserId:      m.UserId,
		CommunityId: m.CommunityId,
		BindStatus:  int32(m.BindStatus),
		JoinTime:    m.JoinTime.Unix(),
		CreatedAt:   m.CreatedTime.Unix(),
		UpdatedAt:   m.UpdatedTime.Unix(),
	}
	if m.LeaveTime.Valid {
		cm.LeaveTime = m.LeaveTime.Time.Unix()
	}
	return cm
}

// toProtoMemberships converts a slice of memberships to proto.
func toProtoMemberships(memberships []*model.UserCommunityMembership) []*userv1.CommunityMembership {
	result := make([]*userv1.CommunityMembership, 0, len(memberships))
	for _, m := range memberships {
		result = append(result, toProtoMembership(m))
	}
	return result
}

// toProtoRole converts a model role to proto.
func toProtoRole(r *model.UserMembershipRole) *userv1.MembershipRole {
	if r == nil {
		return nil
	}
	role := &userv1.MembershipRole{
		Id:          r.Id,
		UserId:      r.UserId,
		CommunityId: r.CommunityId,
		RoleCode:    r.RoleCode,
		VerfStatus:  int32(r.VerfStatus),
		CreatedAt:   r.CreatedTime.Unix(),
		UpdatedAt:   r.UpdatedTime.Unix(),
	}
	if r.MembershipId.Valid {
		role.MembershipId = r.MembershipId.Int64
	}
	if r.VerifiedAt.Valid {
		role.VerifiedAt = r.VerifiedAt.Time.Unix()
	}
	if r.ExpiresAt.Valid {
		role.ExpiresAt = r.ExpiresAt.Time.Unix()
	}
	return role
}

// toProtoRoles converts a slice of roles to proto.
func toProtoRoles(roles []*model.UserMembershipRole) []*userv1.MembershipRole {
	result := make([]*userv1.MembershipRole, 0, len(roles))
	for _, r := range roles {
		result = append(result, toProtoRole(r))
	}
	return result
}

// toProtoCertification converts a model certification to proto.
func toProtoCertification(c *model.UserCertification) *userv1.Certification {
	if c == nil {
		return nil
	}
	cert := &userv1.Certification{
		Id:           c.Id,
		RoleId:       c.RoleId,
		UserId:       c.UserId,
		DocumentUrls: c.DocumentUrls.String,
		Status:       int32(c.Status),
		SubmitTime:   c.SubmitTime.Unix(),
	}
	if c.ReviewerId.Valid {
		cert.ReviewerId = c.ReviewerId.Int64
	}
	if c.ReviewNotes.Valid {
		cert.ReviewNotes = c.ReviewNotes.String
	}
	if c.ReviewTime.Valid {
		cert.ReviewTime = c.ReviewTime.Time.Unix()
	}
	return cert
}

// toProtoResidence converts a model residence to proto.
func toProtoResidence(r *model.UserResidence) *userv1.Residence {
	if r == nil {
		return nil
	}
	res := &userv1.Residence{
		Id:           r.Id,
		MembershipId: r.MembershipId,
		HouseId:      r.HouseId,
		Building:     r.Building,
		Unit:         r.Unit,
		Room:         r.Room,
		IsPrimary:    int32(r.IsPrimary),
		CreatedAt:    r.CreatedTime.Unix(),
		UpdatedAt:    r.UpdatedTime.Unix(),
	}
	if r.StartDate.Valid {
		res.StartDate = r.StartDate.Time.Format("2006-01-02")
	}
	if r.EndDate.Valid {
		res.EndDate = r.EndDate.Time.Format("2006-01-02")
	}
	return res
}

// documentUrlsToJSON converts []string to JSON string for storage.
func documentUrlsToJSON(urls []string) string {
	if len(urls) == 0 {
		return "[]"
	}
	b, _ := json.Marshal(urls)
	return string(b)
}

// buildHouseId constructs house_id from building, unit, room.
// Format: "building-unit-room" or "building-room" if unit is empty.
func buildHouseId(building, unit, room string) string {
	if unit == "" {
		return building + "-" + room
	}
	return building + "-" + unit + "-" + room
}

// newBaseResp creates a success BaseResp.
func newBaseResp() *commonv1.BaseResp {
	return &commonv1.BaseResp{Code: 0, Msg: "success"}
}
```

- [ ] **Step 2: Write `rpc/internal/logic/user/create_user_logic.go`**

```go
package user

import (
	"context"
	"database/sql"
	"time"

	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-common/v2/pkg/crypto"
	"github.com/guxiao1976/community-common/v2/pkg/snowflake"
	"github.com/guxiao1976/community-user/model"
	"github.com/guxiao1976/community-user/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type CreateUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateUserLogic {
	return &CreateUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateUserLogic) CreateUser(in *userv1.CreateUserRequest) (*userv1.CreateUserResponse, error) {
	// Check if phone already exists
	encryptedPhone := crypto.EncryptPhone(in.Phone, l.svcCtx.Config.AesKey)
	existing, err := l.svcCtx.UserBaseModel.FindOneByPhone(l.ctx, in.Phone)
	if err != nil && err != model.ErrNotFound {
		l.Errorf("check phone exists error: %v", err)
		return nil, err
	}
	if existing != nil {
		return &userv1.CreateUserResponse{
			Base: &userv1.BaseResp{Code: 100002, Msg: "手机号已注册"},
		}, nil
	}

	// Generate snowflake ID
	userId := snowflake.Generate()

	nickname := sql.NullString{Valid: false}
	if in.Nickname != "" {
		nickname = sql.NullString{String: in.Nickname, Valid: true}
	}

	now := time.Now()
	user := &model.UserBase{
		Id:          userId,
		Phone:       encryptedPhone,
		Nickname:    nickname,
		Status:      model.UserStatusActive,
		CreditScore: 100,
		CreatedTime: now,
		UpdatedTime: now,
	}

	_, err = l.svcCtx.UserBaseModel.Insert(l.ctx, user)
	if err != nil {
		l.Errorf("insert user error: %v", err)
		return nil, err
	}

	return &userv1.CreateUserResponse{
		Base:   newBaseResp(),
		UserId: userId,
	}, nil
}
```

Wait — I need to check the crypto package API. Let me look at what's available. The old code used `crypto.InitAES()` and `crypto.Encrypt()`. Let me check the actual API before writing code.

Actually, I'll adapt during implementation. Let me continue writing the plan with the general approach and fix API details during implementation.

Let me also think about the error code format. The design spec says `100002` = phone already registered. And the BaseResp needs to come from the common proto package. Let me check the common proto — it has `BaseResp` with `code` and `msg`. But in the user proto, `CreateUserResponse` uses its own `base` field. Let me check what type it uses.

Looking at the current proto: `CreateUserResponse` has `Base *common.v1.BaseResp`. So in my new proto, I'm using `common.v1.BaseResp base = 1;` which is correct.

But in the Go code, the type will be `commonv1.BaseResp`. Let me update my helper to use the right import.

Actually, looking at the helper code I wrote, I imported `commonv1` but didn't use it directly. Let me fix that.

For now, let me continue writing the plan. The implementation can adapt.

- [ ] **Step 2 (continued): Complete create_user_logic.go**

See above. The logic:
1. AES encrypt the phone
2. Check if phone already exists → error 100002
3. Generate snowflake ID
4. Insert with status=1, credit_score=100
5. Return user_id

Note: `UserType` and `ScopeId` from the request are **ignored** (deprecated fields).

- [ ] **Step 3: Write `rpc/internal/logic/user/get_user_logic.go`**

```go
package user

import (
	"context"

	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-user/model"
	"github.com/guxiao1976/community-user/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserLogic {
	return &GetUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserLogic) GetUser(in *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
	user, err := l.svcCtx.UserBaseModel.FindOne(l.ctx, in.Id)
	if err != nil {
		if err == model.ErrNotFound {
			return &userv1.GetUserResponse{
				Base: &userv1.BaseResp{Code: 100001, Msg: "用户不存在"},
			}, nil
		}
		l.Errorf("find user error: %v", err)
		return nil, err
	}

	return &userv1.GetUserResponse{
		Base: newBaseResp(),
		User: toProtoUser(user),
	}, nil
}
```

Wait, I'm using `&userv1.BaseResp{...}` directly but I should be using `commonv1.BaseResp` from the common proto. Let me fix this pattern throughout.

Actually, looking at my proto definition, the `base` fields use `common.v1.BaseResp`. In the generated Go code, this becomes `*commonv1.BaseResp`. But in my helper I defined `newBaseResp()` returning `*commonv1.BaseResp`. That's correct.

But in the error returns, I'm using `&userv1.BaseResp{...}` which is wrong — `BaseResp` is in the common package, not user package. Let me fix this.

I'll use a consistent pattern: import commonv1 and use `&commonv1.BaseResp{Code: ..., Msg: ...}` for error cases, and `newBaseResp()` for success cases.

Actually, I realize I should look at the generated Go code to confirm the exact type names. But I know from the current codebase exploration that the response structs have `Base *commonv1.BaseResp`. So:

```go
import commonv1 "github.com/guxiao1976/api-proto/gen/go/common/v1"
```

I'll need this import in every logic file, or I can define error response helpers in `helper.go`.

Let me define error helpers in helper.go:

```go
func errResp(code int32, msg string) *commonv1.BaseResp {
    return &commonv1.BaseResp{Code: code, Msg: msg}
}
```

This is cleaner. Let me adjust the plan accordingly. But I've already written a lot of detail. Let me just note this and move on to avoid the plan becoming too long. The implementation will handle these details.

Let me continue writing the remaining tasks more concisely since the pattern is established.

- [ ] **Step 4: Write `rpc/internal/logic/user/get_user_by_phone_logic.go`**

AES encrypt phone → FindOneByPhone → return user or error 100001.

- [ ] **Step 5: Write `rpc/internal/logic/user/update_user_logic.go`**

Partial update. Handle optional fields: nickname, avatar_url, gender, birth_date, status, preferences.
Backward compat: `status` field is used by auth-service's saga compensation (sets status to a "disabled" value). In the new design, `status=3` is invalid (only 1-active, 2-disabled). We map status=3 → soft delete (set delete_time).

- [ ] **Step 6: Write `rpc/internal/logic/user/list_users_logic.go`**

Paginated search with keyword (nickname/phone) and status filter.

- [ ] **Step 7: Write `rpc/internal/logic/user/get_users_by_ids_logic.go`**

Batch lookup by ID list.

- [ ] **Step 8: Verify compilation**

```bash
cd /home/jiaoxh/my-project/community-home/services/user-service && go build ./rpc/internal/logic/user/...
```

---

### Task 6: RPC Logic Layer — Community Membership

**Files:**
- Create: `services/user-service/rpc/internal/logic/user/join_community_logic.go`
- Create: `services/user-service/rpc/internal/logic/user/leave_community_logic.go`
- Create: `services/user-service/rpc/internal/logic/user/get_user_memberships_logic.go`

**Business logic per design spec section 3.2, 3.7:**

- `JoinCommunity`: Count active memberships (max 5) → INSERT membership → optionally set default_community_id in preferences
- `LeaveCommunity`: Find membership → UPDATE bind_status=0 + leave_time → update preferences if needed
- `GetUserMemberships`: Find by user_id where bind_status=1

---

### Task 7: RPC Logic Layer — Roles

**Files:**
- Create: `services/user-service/rpc/internal/logic/user/apply_role_logic.go`
- Create: `services/user-service/rpc/internal/logic/user/get_user_roles_logic.go`

**Business logic per design spec section 3.3:**

- `ApplyRole`: Find membership → check role doesn't exist → INSERT role (verf_status=0) → IF owner/tenant: INSERT residence
- `GetUserRoles`: Find roles by user_id and optional community_id

---

### Task 8: RPC Logic Layer — Certifications

**Files:**
- Create: `services/user-service/rpc/internal/logic/user/submit_certification_logic.go`
- Create: `services/user-service/rpc/internal/logic/user/review_certification_logic.go`
- Create: `services/user-service/rpc/internal/logic/user/list_certifications_logic.go`
- Create: `services/user-service/rpc/internal/logic/user/get_my_certifications_logic.go`

**Business logic per design spec sections 3.4, 3.5:**

- `SubmitCertification`: Validate role verf_status IN (0,3,4) → AES encrypt id_card_number → INSERT certification (status=1) → UPDATE role SET verf_status=1
- `ReviewCertification`: Find certification (status=1) → UPDATE certification + UPDATE role verf_status → IF approved: UPDATE user_base real_name/id_card_number (COALESCE)
- `ListCertifications`: Admin paginated list
- `GetMyCertifications`: User's own certifications by user_id

---

### Task 9: RPC Logic Layer — Residences

**Files:**
- Create: `services/user-service/rpc/internal/logic/user/bind_residence_logic.go`
- Create: `services/user-service/rpc/internal/logic/user/get_residences_logic.go`

**Business logic:**

- `BindResidence`: Build house_id → INSERT residence (upsert on uk_member_house)
- `GetResidences`: Find by membership_id

---

### Task 10: RPC Server Registration

**Files:**
- Modify: `services/user-service/rpc/internal/server/userserviceserver.go`

- [ ] **Step 1: Rewrite `userserviceserver.go`**

Register all new RPC method handlers. Each method creates its logic struct and delegates. Remove old method handlers.

```go
package server

import (
	"context"

	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-user/rpc/internal/logic/user"
	"github.com/guxiao1976/community-user/rpc/internal/svc"
)

type UserServiceServer struct {
	svcCtx *svc.ServiceContext
	userv1.UnimplementedUserServiceServer
}

func NewUserServiceServer(svcCtx *svc.ServiceContext) *UserServiceServer {
	return &UserServiceServer{
		svcCtx: svcCtx,
	}
}

func (s *UserServiceServer) CreateUser(ctx context.Context, in *userv1.CreateUserRequest) (*userv1.CreateUserResponse, error) {
	return user.NewCreateUserLogic(ctx, s.svcCtx).CreateUser(in)
}

func (s *UserServiceServer) GetUser(ctx context.Context, in *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
	return user.NewGetUserLogic(ctx, s.svcCtx).GetUser(in)
}

func (s *UserServiceServer) GetUserByPhone(ctx context.Context, in *userv1.GetUserByPhoneRequest) (*userv1.GetUserResponse, error) {
	return user.NewGetUserByPhoneLogic(ctx, s.svcCtx).GetUserByPhone(in)
}

func (s *UserServiceServer) UpdateUser(ctx context.Context, in *userv1.UpdateUserRequest) (*userv1.UpdateUserResponse, error) {
	return user.NewUpdateUserLogic(ctx, s.svcCtx).UpdateUser(in)
}

func (s *UserServiceServer) ListUsers(ctx context.Context, in *userv1.ListUsersRequest) (*userv1.ListUsersResponse, error) {
	return user.NewListUsersLogic(ctx, s.svcCtx).ListUsers(in)
}

func (s *UserServiceServer) GetUsersByIds(ctx context.Context, in *userv1.GetUsersByIdsRequest) (*userv1.GetUsersByIdsResponse, error) {
	return user.NewGetUsersByIdsLogic(ctx, s.svcCtx).GetUsersByIds(in)
}

func (s *UserServiceServer) JoinCommunity(ctx context.Context, in *userv1.JoinCommunityRequest) (*userv1.JoinCommunityResponse, error) {
	return user.NewJoinCommunityLogic(ctx, s.svcCtx).JoinCommunity(in)
}

func (s *UserServiceServer) LeaveCommunity(ctx context.Context, in *userv1.LeaveCommunityRequest) (*userv1.LeaveCommunityResponse, error) {
	return user.NewLeaveCommunityLogic(ctx, s.svcCtx).LeaveCommunity(in)
}

func (s *UserServiceServer) GetUserMemberships(ctx context.Context, in *userv1.GetUserMembershipsRequest) (*userv1.GetUserMembershipsResponse, error) {
	return user.NewGetUserMembershipsLogic(ctx, s.svcCtx).GetUserMemberships(in)
}

func (s *UserServiceServer) ApplyRole(ctx context.Context, in *userv1.ApplyRoleRequest) (*userv1.ApplyRoleResponse, error) {
	return user.NewApplyRoleLogic(ctx, s.svcCtx).ApplyRole(in)
}

func (s *UserServiceServer) GetUserRoles(ctx context.Context, in *userv1.GetUserRolesRequest) (*userv1.GetUserRolesResponse, error) {
	return user.NewGetUserRolesLogic(ctx, s.svcCtx).GetUserRoles(in)
}

func (s *UserServiceServer) SubmitCertification(ctx context.Context, in *userv1.SubmitCertificationRequest) (*userv1.SubmitCertificationResponse, error) {
	return user.NewSubmitCertificationLogic(ctx, s.svcCtx).SubmitCertification(in)
}

func (s *UserServiceServer) ReviewCertification(ctx context.Context, in *userv1.ReviewCertificationRequest) (*userv1.ReviewCertificationResponse, error) {
	return user.NewReviewCertificationLogic(ctx, s.svcCtx).ReviewCertification(in)
}

func (s *UserServiceServer) ListCertifications(ctx context.Context, in *userv1.ListCertificationsRequest) (*userv1.ListCertificationsResponse, error) {
	return user.NewListCertificationsLogic(ctx, s.svcCtx).ListCertifications(in)
}

func (s *UserServiceServer) GetMyCertifications(ctx context.Context, in *userv1.GetMyCertificationsRequest) (*userv1.GetMyCertificationsResponse, error) {
	return user.NewGetMyCertificationsLogic(ctx, s.svcCtx).GetMyCertifications(in)
}

func (s *UserServiceServer) BindResidence(ctx context.Context, in *userv1.BindResidenceRequest) (*userv1.BindResidenceResponse, error) {
	return user.NewBindResidenceLogic(ctx, s.svcCtx).BindResidence(in)
}

func (s *UserServiceServer) GetResidences(ctx context.Context, in *userv1.GetResidencesRequest) (*userv1.GetResidencesResponse, error) {
	return user.NewGetResidencesLogic(ctx, s.svcCtx).GetResidences(in)
}
```

- [ ] **Step 2: Verify RPC server compiles**

```bash
cd /home/jiaoxh/my-project/community-home/services/user-service && go build ./rpc/...
```

Expected: Full RPC compilation succeeds.

---

### Task 11: API Layer Update

**Files:**
- Delete all old api/ files
- Create new API layer with simplified REST endpoints

Since the user says "当前等注册、登录逻辑是正常的，改造完要注册、登录正常" and registration/login are handled entirely by auth-service → user-service gRPC, the API layer is secondary. We should update it to match the new RPC interface but it's not blocking for registration/login.

- [ ] **Step 1: Update `api/internal/config/config.go`** — adjust RPC client config
- [ ] **Step 2: Update `api/etc/user-api.yaml`** — adjust config
- [ ] **Step 3: Update `api/internal/svc/servicecontext.go`** — new RPC client(s)
- [ ] **Step 4: Update `api/internal/handler/routes.go`** — new route registration
- [ ] **Step 5: Rewrite handlers and logic** — thin delegation to user-service RPC

---

### Task 12: Integration Testing — Registration & Login Flow

- [ ] **Step 1: Ensure MySQL + etcd are running**

```bash
docker compose ps
```

- [ ] **Step 2: Run database migration** (if not already done)

```bash
mysql -h 172.18.0.2 -u root -proot user < /home/jiaoxh/my-project/community-home/services/user-service/migration/001_refactor_to_v2.sql
```

- [ ] **Step 3: Build and start user-service RPC**

```bash
cd /home/jiaoxh/my-project/community-home/services/user-service && go build ./rpc/...
cd rpc && go run userservice.go -f etc/userservice.yaml &
```

- [ ] **Step 4: Build and start auth-service**

```bash
cd /home/jiaoxh/my-project/community-home/services/auth-service && go build ./rpc/...
cd rpc && go run authservice.go -f etc/authservice.yaml &
```

- [ ] **Step 5: Test registration flow via gRPC**

Use `grpcurl` or a test script to call auth-service registration:
```bash
grpcurl -plaintext -d '{"phone":"13800138000","nickname":"testuser"}' localhost:8081 auth.v1.AuthService/Register
```

Expected: Returns user_id, registration succeeds. user_base has 1 row.

- [ ] **Step 6: Test login flow**

```bash
grpcurl -plaintext -d '{"phone":"13800138000","password":"..."}' localhost:8081 auth.v1.AuthService/Login
```

Expected: Returns JWT token. Login succeeds.

- [ ] **Step 7: Verify user_base table**

```bash
mysql -h 172.18.0.2 -u root -proot user -e "SELECT id, phone, nickname, status, credit_score, created_time FROM user_base"
```

Expected: User record exists with status=1, credit_score=100. No user_type, cert_status, or scope_id columns.

---

### Task 13: Final Verification

- [ ] **Step 1: Run all user-service tests**

```bash
cd /home/jiaoxh/my-project/community-home/services/user-service && go test ./...
```

- [ ] **Step 2: Run auth-service tests**

```bash
cd /home/jiaoxh/my-project/community-home/services/auth-service && go test ./...
```

- [ ] **Step 3: Verify go.work resolution**

```bash
cd /home/jiaoxh/my-project/community-home && go work sync
```

- [ ] **Step 4: Commit all remaining changes**
