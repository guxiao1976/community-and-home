# Data Model: Identity and Masterdata Microservices

**Feature**: 001-identity-masterdata  
**Date**: 2026-04-13  
**Database**: MySQL 8.0  
**Character Set**: utf8mb4  
**Collation**: utf8mb4_unicode_ci

## Database Schema Overview

Two independent database schemas:
- **identity_db**: Authentication and authorization (auth_* tables)
- **masterdata_db**: Master data management (md_* tables)

---

## Identity Service Schema (identity_db)

### auth_user

User accounts for all platform users (backend staff and homeowners).

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | User ID |
| phone | VARCHAR(20) | UNIQUE, NOT NULL | Phone number (login identifier) |
| password_hash | VARCHAR(255) | NOT NULL | bcrypt hashed password |
| nickname | VARCHAR(50) | NULL | Display name |
| avatar_url | VARCHAR(255) | NULL | Profile picture URL |
| user_type | TINYINT | NOT NULL | 1=Backend, 2=Homeowner |
| status | TINYINT | NOT NULL, DEFAULT 1 | 1=Active, 2=Disabled, 3=Locked |
| verification_status | TINYINT | NOT NULL, DEFAULT 0 | 0=Unverified, 1=Verified, 2=Rejected |
| scope_id | BIGINT | NULL, INDEX | Administrative scope (NULL=headquarters) |
| last_login_time | TIMESTAMP | NULL | Last successful login |
| last_login_ip | VARCHAR(45) | NULL | Last login IP address |
| created_time | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP | Creation timestamp |
| updated_time | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | Last update timestamp |
| delete_time | TIMESTAMP | NULL, INDEX | Soft delete timestamp |

**Indexes**:
- PRIMARY KEY (id)
- UNIQUE KEY uk_phone (phone)
- KEY idx_scope (scope_id)
- KEY idx_delete (delete_time)
- KEY idx_status (status)

**Validation Rules**:
- phone: Must match Chinese mobile format (11 digits starting with 1)
- password_hash: bcrypt with cost factor 10
- user_type: Must be 1 or 2
- status: Must be 1, 2, or 3

---

### auth_role

Role definitions for RBAC.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | Role ID |
| name | VARCHAR(50) | NOT NULL | Role name |
| code | VARCHAR(50) | UNIQUE, NOT NULL | Role code (e.g., SUPER_ADMIN) |
| description | VARCHAR(255) | NULL | Role description |
| is_system | TINYINT | NOT NULL, DEFAULT 0 | 1=System role (cannot delete) |
| sort_order | INT | NOT NULL, DEFAULT 0 | Display order |
| status | TINYINT | NOT NULL, DEFAULT 1 | 1=Active, 2=Disabled |
| created_by | BIGINT | NOT NULL | Creator user ID |
| created_time | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP | Creation timestamp |
| updated_time | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | Last update timestamp |
| delete_time | TIMESTAMP | NULL, INDEX | Soft delete timestamp |

**Indexes**:
- PRIMARY KEY (id)
- UNIQUE KEY uk_code (code)
- KEY idx_delete (delete_time)

**Validation Rules**:
- code: Must be uppercase alphanumeric with underscores
- is_system: System roles (SUPER_ADMIN) cannot be deleted or disabled

---

### auth_permission

Permission definitions (menu and button level).

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | Permission ID |
| parent_id | BIGINT | NULL, INDEX | Parent permission ID (NULL=root) |
| name | VARCHAR(50) | NOT NULL | Permission name |
| code | VARCHAR(100) | UNIQUE, NOT NULL | Permission code (e.g., user:create) |
| type | TINYINT | NOT NULL | 1=Menu, 2=Button |
| path | VARCHAR(255) | NULL | Frontend route path (for menus) |
| icon | VARCHAR(50) | NULL | Icon name (for menus) |
| sort_order | INT | NOT NULL, DEFAULT 0 | Display order |
| status | TINYINT | NOT NULL, DEFAULT 1 | 1=Active, 2=Disabled |
| created_time | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP | Creation timestamp |
| updated_time | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | Last update timestamp |

**Indexes**:
- PRIMARY KEY (id)
- UNIQUE KEY uk_code (code)
- KEY idx_parent (parent_id)
- KEY idx_type (type)

**Validation Rules**:
- code: Format `module:action` (e.g., user:create, role:update)
- type: 1=Menu (has path), 2=Button (no path)
- parent_id: Must reference existing permission or NULL

---

### auth_role_permission

Role-permission mapping (many-to-many).

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | Mapping ID |
| role_id | BIGINT | NOT NULL, INDEX | Role ID |
| permission_id | BIGINT | NOT NULL, INDEX | Permission ID |
| created_time | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP | Creation timestamp |

**Indexes**:
- PRIMARY KEY (id)
- UNIQUE KEY uk_role_permission (role_id, permission_id)
- KEY idx_role (role_id)
- KEY idx_permission (permission_id)

**Foreign Keys**:
- FOREIGN KEY (role_id) REFERENCES auth_role(id) ON DELETE CASCADE
- FOREIGN KEY (permission_id) REFERENCES auth_permission(id) ON DELETE CASCADE

---

### auth_user_role

User-role mapping (many-to-many).

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | Mapping ID |
| user_id | BIGINT | NOT NULL, INDEX | User ID |
| role_id | BIGINT | NOT NULL, INDEX | Role ID |
| created_time | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP | Creation timestamp |

**Indexes**:
- PRIMARY KEY (id)
- UNIQUE KEY uk_user_role (user_id, role_id)
- KEY idx_user (user_id)
- KEY idx_role (role_id)

**Foreign Keys**:
- FOREIGN KEY (user_id) REFERENCES auth_user(id) ON DELETE CASCADE
- FOREIGN KEY (role_id) REFERENCES auth_role(id) ON DELETE CASCADE

---

### auth_property_unit

Property units within communities.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | Property unit ID |
| community_id | BIGINT | NOT NULL, INDEX | Community ID (from masterdata) |
| building | VARCHAR(50) | NOT NULL | Building number/name |
| unit | VARCHAR(50) | NOT NULL | Unit number |
| floor | VARCHAR(20) | NULL | Floor number |
| area | DECIMAL(10,2) | NULL | Area in square meters |
| property_type | TINYINT | NOT NULL | 1=Residential, 2=Commercial, 3=Mixed |
| status | TINYINT | NOT NULL, DEFAULT 1 | 1=Active, 2=Inactive |
| created_time | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP | Creation timestamp |
| updated_time | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | Last update timestamp |
| delete_time | TIMESTAMP | NULL, INDEX | Soft delete timestamp |

**Indexes**:
- PRIMARY KEY (id)
- UNIQUE KEY uk_community_building_unit (community_id, building, unit)
- KEY idx_community (community_id)
- KEY idx_delete (delete_time)

---

### auth_property_binding

User-property unit bindings.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | Binding ID |
| user_id | BIGINT | NOT NULL, INDEX | User ID |
| property_unit_id | BIGINT | NOT NULL, INDEX | Property unit ID |
| is_primary | TINYINT | NOT NULL, DEFAULT 0 | 1=Primary user, 0=Secondary |
| bind_status | TINYINT | NOT NULL, DEFAULT 1 | 1=Active, 2=Pending, 3=Revoked |
| bind_time | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP | Binding timestamp |
| revoke_time | TIMESTAMP | NULL | Revocation timestamp |
| revoked_by | BIGINT | NULL | Revoker user ID |
| created_time | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP | Creation timestamp |
| updated_time | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | Last update timestamp |

**Indexes**:
- PRIMARY KEY (id)
- UNIQUE KEY uk_user_property (user_id, property_unit_id)
- KEY idx_user (user_id)
- KEY idx_property (property_unit_id)
- KEY idx_status (bind_status)

**Foreign Keys**:
- FOREIGN KEY (user_id) REFERENCES auth_user(id) ON DELETE CASCADE
- FOREIGN KEY (property_unit_id) REFERENCES auth_property_unit(id) ON DELETE CASCADE

**Validation Rules**:
- Only one primary user per property unit
- Primary user cannot be revoked without transferring primary status

---

### auth_homeowner_verification

Homeowner identity verification requests.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | Verification ID |
| user_id | BIGINT | NOT NULL, INDEX | User ID |
| property_unit_id | BIGINT | NOT NULL, INDEX | Property unit ID |
| document_urls | TEXT | NOT NULL | JSON array of document URLs in MinIO |
| real_name | VARCHAR(50) | NOT NULL | Real name on documents |
| id_card_number | VARCHAR(18) | NOT NULL | ID card number |
| verification_status | TINYINT | NOT NULL, DEFAULT 0 | 0=Pending, 1=Approved, 2=Rejected |
| reviewer_id | BIGINT | NULL | Reviewer user ID |
| review_time | TIMESTAMP | NULL | Review timestamp |
| review_notes | VARCHAR(500) | NULL | Reviewer notes/rejection reason |
| submit_time | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP | Submission timestamp |
| created_time | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP | Creation timestamp |
| updated_time | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | Last update timestamp |

**Indexes**:
- PRIMARY KEY (id)
- KEY idx_user (user_id)
- KEY idx_property (property_unit_id)
- KEY idx_status (verification_status)
- KEY idx_submit_time (submit_time)

**Foreign Keys**:
- FOREIGN KEY (user_id) REFERENCES auth_user(id) ON DELETE CASCADE
- FOREIGN KEY (property_unit_id) REFERENCES auth_property_unit(id) ON DELETE CASCADE

**Validation Rules**:
- document_urls: JSON array, max 9 URLs
- id_card_number: 18-digit Chinese ID card format
- Only one pending verification per user at a time

---

### auth_family

Family/household records.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | Family ID |
| property_unit_id | BIGINT | NOT NULL, INDEX | Property unit ID |
| family_head_id | BIGINT | NOT NULL, INDEX | Family head user ID |
| family_name | VARCHAR(100) | NULL | Family name |
| status | TINYINT | NOT NULL, DEFAULT 1 | 1=Active, 2=Inactive |
| created_time | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP | Creation timestamp |
| updated_time | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | Last update timestamp |
| delete_time | TIMESTAMP | NULL, INDEX | Soft delete timestamp |

**Indexes**:
- PRIMARY KEY (id)
- UNIQUE KEY uk_property (property_unit_id)
- KEY idx_head (family_head_id)
- KEY idx_delete (delete_time)

**Foreign Keys**:
- FOREIGN KEY (property_unit_id) REFERENCES auth_property_unit(id) ON DELETE CASCADE
- FOREIGN KEY (family_head_id) REFERENCES auth_user(id) ON DELETE CASCADE

---

### auth_family_member

Family member records.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | Member ID |
| family_id | BIGINT | NOT NULL, INDEX | Family ID |
| user_id | BIGINT | NULL, INDEX | User ID (NULL if not registered) |
| name | VARCHAR(50) | NOT NULL | Member name |
| relationship | VARCHAR(20) | NOT NULL | Relationship to family head |
| phone | VARCHAR(20) | NULL | Phone number |
| id_card_number | VARCHAR(18) | NULL | ID card number |
| birth_date | DATE | NULL | Birth date |
| gender | TINYINT | NULL | 1=Male, 2=Female |
| created_time | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP | Creation timestamp |
| updated_time | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | Last update timestamp |
| delete_time | TIMESTAMP | NULL, INDEX | Soft delete timestamp |

**Indexes**:
- PRIMARY KEY (id)
- KEY idx_family (family_id)
- KEY idx_user (user_id)
- KEY idx_delete (delete_time)

**Foreign Keys**:
- FOREIGN KEY (family_id) REFERENCES auth_family(id) ON DELETE CASCADE
- FOREIGN KEY (user_id) REFERENCES auth_user(id) ON DELETE SET NULL

---

### auth_uploaded_file

File upload tracking (MinIO references).

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | File ID |
| user_id | BIGINT | NOT NULL, INDEX | Uploader user ID |
| entity_type | VARCHAR(50) | NOT NULL | Entity type (e.g., homeowner_verification) |
| entity_id | BIGINT | NOT NULL, INDEX | Entity ID |
| file_name | VARCHAR(255) | NOT NULL | Original file name |
| file_path | VARCHAR(500) | NOT NULL | MinIO object path |
| file_size | BIGINT | NOT NULL | File size in bytes |
| file_type | VARCHAR(50) | NOT NULL | MIME type |
| bucket_name | VARCHAR(100) | NOT NULL | MinIO bucket name |
| upload_time | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP | Upload timestamp |
| created_time | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP | Creation timestamp |

**Indexes**:
- PRIMARY KEY (id)
- KEY idx_user (user_id)
- KEY idx_entity (entity_type, entity_id)
- KEY idx_upload_time (upload_time)

**Foreign Keys**:
- FOREIGN KEY (user_id) REFERENCES auth_user(id) ON DELETE CASCADE

**Validation Rules**:
- file_size: Max 5MB (5242880 bytes)
- file_type: Only image/jpeg, image/png

---

## Masterdata Service Schema (masterdata_db)

### md_administrative_division

Five-tier administrative division hierarchy.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | Division ID |
| parent_id | BIGINT | NULL, INDEX | Parent division ID (NULL=root) |
| level | TINYINT | NOT NULL | 1=Province, 2=City, 3=District, 4=Street, 5=Community |
| name | VARCHAR(100) | NOT NULL | Division name |
| code | VARCHAR(20) | UNIQUE, NOT NULL | Administrative code |
| path | VARCHAR(500) | NOT NULL, INDEX | Materialized path (e.g., /1/23/456/) |
| sort_order | INT | NOT NULL, DEFAULT 0 | Display order |
| status | TINYINT | NOT NULL, DEFAULT 1 | 1=Active, 2=Inactive |
| created_by | BIGINT | NOT NULL | Creator user ID |
| created_time | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP | Creation timestamp |
| updated_time | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | Last update timestamp |
| delete_time | TIMESTAMP | NULL, INDEX | Soft delete timestamp |

**Indexes**:
- PRIMARY KEY (id)
- UNIQUE KEY uk_code (code)
- KEY idx_parent (parent_id)
- KEY idx_level (level)
- KEY idx_path (path)
- KEY idx_delete (delete_time)

**Validation Rules**:
- level: Must be 1-5
- parent_id: Level 1 must have NULL parent, others must have parent of level-1
- path: Auto-generated trigger on insert/update

---

### md_community

Community/village master data.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | Community ID |
| division_id | BIGINT | NOT NULL, INDEX | Administrative division ID (level 5) |
| name | VARCHAR(100) | NOT NULL | Community name |
| address | VARCHAR(255) | NOT NULL | Full address |
| area | DECIMAL(10,2) | NULL | Area in square kilometers |
| population | INT | NULL | Population count |
| community_type | TINYINT | NOT NULL | 1=Residential, 2=Village, 3=Mixed |
| submission_status | TINYINT | NOT NULL, DEFAULT 0 | 0=Draft, 1=Submitted, 2=Approved, 3=Rejected |
| submitter_id | BIGINT | NOT NULL | Submitter user ID |
| submit_time | TIMESTAMP | NULL | Submission timestamp |
| reviewer_id | BIGINT | NULL | Reviewer user ID |
| review_time | TIMESTAMP | NULL | Review timestamp |
| review_notes | VARCHAR(500) | NULL | Review notes/rejection reason |
| created_time | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP | Creation timestamp |
| updated_time | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | Last update timestamp |
| delete_time | TIMESTAMP | NULL, INDEX | Soft delete timestamp |

**Indexes**:
- PRIMARY KEY (id)
- KEY idx_division (division_id)
- KEY idx_status (submission_status)
- KEY idx_submitter (submitter_id)
- KEY idx_delete (delete_time)

**Foreign Keys**:
- FOREIGN KEY (division_id) REFERENCES md_administrative_division(id) ON DELETE RESTRICT

**Validation Rules**:
- division_id: Must reference level 5 (community) division
- submission_status: 0→1 (submit), 1→2 (approve), 1→3 (reject)
- Only headquarters can approve/reject (reviewer must have scope_id=NULL)

---

### md_district_economic_data

County/district economic and population data.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | Data ID |
| division_id | BIGINT | NOT NULL, INDEX | Administrative division ID (level 3) |
| year | INT | NOT NULL | Data year |
| population | BIGINT | NULL | Total population |
| gdp | DECIMAL(15,2) | NULL | GDP in million CNY |
| per_capita_income | DECIMAL(10,2) | NULL | Per capita income in CNY |
| unemployment_rate | DECIMAL(5,2) | NULL | Unemployment rate (%) |
| data_source | VARCHAR(255) | NULL | Data source |
| created_by | BIGINT | NOT NULL | Creator user ID |
| created_time | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP | Creation timestamp |
| updated_time | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | Last update timestamp |

**Indexes**:
- PRIMARY KEY (id)
- UNIQUE KEY uk_division_year (division_id, year)
- KEY idx_division (division_id)
- KEY idx_year (year)

**Foreign Keys**:
- FOREIGN KEY (division_id) REFERENCES md_administrative_division(id) ON DELETE CASCADE

**Validation Rules**:
- division_id: Must reference level 3 (district) division
- year: Must be between 2000 and current year + 1

---

### md_configuration

Platform-wide configurable parameters.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | Configuration ID |
| module | VARCHAR(50) | NOT NULL, INDEX | Module name (e.g., auth, masterdata) |
| config_key | VARCHAR(100) | NOT NULL | Configuration key |
| config_value | TEXT | NOT NULL | Configuration value (JSON) |
| value_type | VARCHAR(20) | NOT NULL | string/number/boolean/json |
| description | VARCHAR(255) | NULL | Configuration description |
| is_public | TINYINT | NOT NULL, DEFAULT 0 | 1=Public (visible to all), 0=Internal |
| approval_status | TINYINT | NOT NULL, DEFAULT 0 | 0=Draft, 1=Pending, 2=Approved |
| created_by | BIGINT | NOT NULL | Creator user ID |
| created_time | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP | Creation timestamp |
| updated_time | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | Last update timestamp |
| delete_time | TIMESTAMP | NULL, INDEX | Soft delete timestamp |

**Indexes**:
- PRIMARY KEY (id)
- UNIQUE KEY uk_module_key (module, config_key)
- KEY idx_module (module)
- KEY idx_status (approval_status)
- KEY idx_delete (delete_time)

**Validation Rules**:
- config_key: Alphanumeric with dots (e.g., sms.rate_limit)
- value_type: Must be string/number/boolean/json
- approval_status: Changes require headquarters approval

---

### md_sensitive_word

Sensitive word list for content moderation.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | Word ID |
| word | VARCHAR(100) | UNIQUE, NOT NULL | Sensitive word |
| category | VARCHAR(50) | NOT NULL | Category (e.g., political, violence) |
| severity | TINYINT | NOT NULL | 1=Low, 2=Medium, 3=High |
| action | TINYINT | NOT NULL | 1=Warn, 2=Block, 3=Review |
| status | TINYINT | NOT NULL, DEFAULT 1 | 1=Active, 2=Inactive |
| created_by | BIGINT | NOT NULL | Creator user ID |
| created_time | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP | Creation timestamp |
| updated_time | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | Last update timestamp |

**Indexes**:
- PRIMARY KEY (id)
- UNIQUE KEY uk_word (word)
- KEY idx_category (category)
- KEY idx_status (status)

**Validation Rules**:
- word: Trimmed, lowercase for matching
- severity: 1-3
- action: 1-3

---

### md_audit_log

Audit log for all significant data changes.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | Log ID |
| user_id | BIGINT | NOT NULL, INDEX | User who made the change |
| entity_type | VARCHAR(50) | NOT NULL, INDEX | Entity type (table name) |
| entity_id | BIGINT | NOT NULL, INDEX | Entity ID |
| action | VARCHAR(20) | NOT NULL | CREATE/UPDATE/DELETE |
| old_value | TEXT | NULL | Old value (JSON) |
| new_value | TEXT | NULL | New value (JSON) |
| ip_address | VARCHAR(45) | NOT NULL | Client IP address |
| user_agent | VARCHAR(255) | NULL | Client user agent |
| created_time | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP, INDEX | Action timestamp |

**Indexes**:
- PRIMARY KEY (id)
- KEY idx_user (user_id)
- KEY idx_entity (entity_type, entity_id)
- KEY idx_action (action)
- KEY idx_created_time (created_time)

**Partitioning**:
- Partition by RANGE on created_time (monthly partitions)
- Retain 24 months of data, archive older partitions

**Validation Rules**:
- action: Must be CREATE/UPDATE/DELETE
- old_value: NULL for CREATE, required for UPDATE/DELETE
- new_value: Required for CREATE/UPDATE, NULL for DELETE

---

## Entity Relationships

### Identity Service

```
auth_user (1) ----< (N) auth_user_role (N) >---- (1) auth_role
auth_role (1) ----< (N) auth_role_permission (N) >---- (1) auth_permission
auth_user (1) ----< (N) auth_property_binding (N) >---- (1) auth_property_unit
auth_user (1) ----< (N) auth_homeowner_verification >---- (1) auth_property_unit
auth_property_unit (1) ----< (1) auth_family (1) >---- (1) auth_user (family_head)
auth_family (1) ----< (N) auth_family_member
auth_user (1) ----< (N) auth_uploaded_file
```

### Masterdata Service

```
md_administrative_division (1) ----< (N) md_administrative_division (parent-child)
md_administrative_division (1) ----< (N) md_community
md_administrative_division (1) ----< (N) md_district_economic_data
```

### Cross-Service References

```
auth_property_unit.community_id -> md_community.id (via RPC, not FK)
auth_user.scope_id -> md_administrative_division.id (via RPC, not FK)
```

---

## Data Migration Strategy

1. **Initial Setup**: Run DDL scripts to create all tables
2. **Seed Data**: Insert default roles, permissions, and super admin user
3. **Administrative Divisions**: Import national 5-tier hierarchy from official data
4. **Indexes**: Create all indexes after bulk data import for performance
5. **Triggers**: Create triggers for path maintenance and audit logging

---

## Performance Considerations

- **Indexes**: All foreign keys, scope_id, delete_time, and frequently queried columns indexed
- **Partitioning**: audit_log partitioned by month for query performance
- **Caching**: Hot tables (roles, permissions, divisions) cached in Redis
- **Soft Delete**: Queries must filter `WHERE delete_time IS NULL` (indexed)
- **Materialized Path**: Enables efficient subtree queries without recursion
