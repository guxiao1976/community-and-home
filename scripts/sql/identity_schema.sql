-- Identity Service Database Schema
-- Database: identity_db
-- MySQL 8.0+
-- Character Set: utf8mb4
-- Collation: utf8mb4_unicode_ci

CREATE DATABASE IF NOT EXISTS identity_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE identity_db;

-- ============================================================
-- Table: auth_user
-- Description: User accounts for all platform users (backend staff and homeowners)
-- ============================================================
CREATE TABLE IF NOT EXISTS auth_user (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT 'User ID',
    phone VARCHAR(20) NOT NULL COMMENT 'Phone number (login identifier)',
    password_hash VARCHAR(255) NOT NULL COMMENT 'bcrypt hashed password',
    nickname VARCHAR(50) NULL COMMENT 'Display name',
    avatar_url VARCHAR(255) NULL COMMENT 'Profile picture URL',
    user_type TINYINT NOT NULL COMMENT '1=Backend, 2=Homeowner',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '1=Active, 2=Disabled, 3=Locked',
    verification_status TINYINT NOT NULL DEFAULT 0 COMMENT '0=Unverified, 1=Verified, 2=Rejected',
    scope_id BIGINT NULL COMMENT 'Administrative scope (NULL=headquarters)',
    last_login_time TIMESTAMP NULL COMMENT 'Last successful login',
    last_login_ip VARCHAR(45) NULL COMMENT 'Last login IP address',
    created_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Creation timestamp',
    updated_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Last update timestamp',
    delete_time TIMESTAMP NULL COMMENT 'Soft delete timestamp',
    UNIQUE KEY uk_phone (phone),
    KEY idx_scope (scope_id),
    KEY idx_delete (delete_time),
    KEY idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='User accounts';

-- ============================================================
-- Table: auth_role
-- Description: Role definitions for RBAC
-- ============================================================
CREATE TABLE IF NOT EXISTS auth_role (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT 'Role ID',
    name VARCHAR(50) NOT NULL COMMENT 'Role name',
    code VARCHAR(50) NOT NULL COMMENT 'Role code (e.g., SUPER_ADMIN)',
    description VARCHAR(255) NULL COMMENT 'Role description',
    is_system TINYINT NOT NULL DEFAULT 0 COMMENT '1=System role (cannot delete)',
    sort_order INT NOT NULL DEFAULT 0 COMMENT 'Display order',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '1=Active, 2=Disabled',
    created_by BIGINT NOT NULL COMMENT 'Creator user ID',
    created_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Creation timestamp',
    updated_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Last update timestamp',
    delete_time TIMESTAMP NULL COMMENT 'Soft delete timestamp',
    UNIQUE KEY uk_code (code),
    KEY idx_delete (delete_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Role definitions';

-- ============================================================
-- Table: auth_permission
-- Description: Permission definitions (menu and button level)
-- ============================================================
CREATE TABLE IF NOT EXISTS auth_permission (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT 'Permission ID',
    parent_id BIGINT NULL COMMENT 'Parent permission ID (NULL=root)',
    name VARCHAR(50) NOT NULL COMMENT 'Permission name',
    code VARCHAR(100) NOT NULL COMMENT 'Permission code (e.g., user:create)',
    type TINYINT NOT NULL COMMENT '1=Menu, 2=Button',
    path VARCHAR(255) NULL COMMENT 'Frontend route path (for menus)',
    icon VARCHAR(50) NULL COMMENT 'Icon name (for menus)',
    sort_order INT NOT NULL DEFAULT 0 COMMENT 'Display order',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '1=Active, 2=Disabled',
    created_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Creation timestamp',
    updated_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Last update timestamp',
    UNIQUE KEY uk_code (code),
    KEY idx_parent (parent_id),
    KEY idx_type (type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Permission definitions';

-- ============================================================
-- Table: auth_role_permission
-- Description: Role-permission mapping (many-to-many)
-- ============================================================
CREATE TABLE IF NOT EXISTS auth_role_permission (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT 'Mapping ID',
    role_id BIGINT NOT NULL COMMENT 'Role ID',
    permission_id BIGINT NOT NULL COMMENT 'Permission ID',
    created_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Creation timestamp',
    UNIQUE KEY uk_role_permission (role_id, permission_id),
    KEY idx_role (role_id),
    KEY idx_permission (permission_id),
    CONSTRAINT fk_role_permission_role FOREIGN KEY (role_id) REFERENCES auth_role(id) ON DELETE CASCADE,
    CONSTRAINT fk_role_permission_permission FOREIGN KEY (permission_id) REFERENCES auth_permission(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Role-permission mapping';

-- ============================================================
-- Table: auth_user_role
-- Description: User-role mapping (many-to-many)
-- ============================================================
CREATE TABLE IF NOT EXISTS auth_user_role (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT 'Mapping ID',
    user_id BIGINT NOT NULL COMMENT 'User ID',
    role_id BIGINT NOT NULL COMMENT 'Role ID',
    created_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Creation timestamp',
    UNIQUE KEY uk_user_role (user_id, role_id),
    KEY idx_user (user_id),
    KEY idx_role (role_id),
    CONSTRAINT fk_user_role_user FOREIGN KEY (user_id) REFERENCES auth_user(id) ON DELETE CASCADE,
    CONSTRAINT fk_user_role_role FOREIGN KEY (role_id) REFERENCES auth_role(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='User-role mapping';

-- ============================================================
-- Table: auth_property_unit
-- Description: Property units within communities
-- ============================================================
CREATE TABLE IF NOT EXISTS auth_property_unit (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT 'Property unit ID',
    community_id BIGINT NOT NULL COMMENT 'Community ID (from masterdata)',
    building VARCHAR(50) NOT NULL COMMENT 'Building number/name',
    unit VARCHAR(50) NOT NULL COMMENT 'Unit number',
    floor VARCHAR(20) NULL COMMENT 'Floor number',
    area DECIMAL(10,2) NULL COMMENT 'Area in square meters',
    property_type TINYINT NOT NULL COMMENT '1=Residential, 2=Commercial, 3=Mixed',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '1=Active, 2=Inactive',
    created_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Creation timestamp',
    updated_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Last update timestamp',
    delete_time TIMESTAMP NULL COMMENT 'Soft delete timestamp',
    UNIQUE KEY uk_community_building_unit (community_id, building, unit),
    KEY idx_community (community_id),
    KEY idx_delete (delete_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Property units';

-- ============================================================
-- Table: auth_property_binding
-- Description: User-property unit bindings
-- ============================================================
CREATE TABLE IF NOT EXISTS auth_property_binding (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT 'Binding ID',
    user_id BIGINT NOT NULL COMMENT 'User ID',
    property_unit_id BIGINT NOT NULL COMMENT 'Property unit ID',
    is_primary TINYINT NOT NULL DEFAULT 0 COMMENT '1=Primary user, 0=Secondary',
    bind_status TINYINT NOT NULL DEFAULT 1 COMMENT '1=Active, 2=Pending, 3=Revoked',
    bind_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Binding timestamp',
    revoke_time TIMESTAMP NULL COMMENT 'Revocation timestamp',
    revoked_by BIGINT NULL COMMENT 'Revoker user ID',
    created_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Creation timestamp',
    updated_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Last update timestamp',
    UNIQUE KEY uk_user_property (user_id, property_unit_id),
    KEY idx_user (user_id),
    KEY idx_property (property_unit_id),
    KEY idx_status (bind_status),
    CONSTRAINT fk_property_binding_user FOREIGN KEY (user_id) REFERENCES auth_user(id) ON DELETE CASCADE,
    CONSTRAINT fk_property_binding_property FOREIGN KEY (property_unit_id) REFERENCES auth_property_unit(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='User-property bindings';

-- ============================================================
-- Table: auth_homeowner_verification
-- Description: Homeowner identity verification requests
-- ============================================================
CREATE TABLE IF NOT EXISTS auth_homeowner_verification (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT 'Verification ID',
    user_id BIGINT NOT NULL COMMENT 'User ID',
    property_unit_id BIGINT NOT NULL COMMENT 'Property unit ID',
    document_urls TEXT NOT NULL COMMENT 'JSON array of document URLs in MinIO',
    real_name VARCHAR(50) NOT NULL COMMENT 'Real name on documents',
    id_card_number VARCHAR(18) NOT NULL COMMENT 'ID card number',
    verification_status TINYINT NOT NULL DEFAULT 0 COMMENT '0=Pending, 1=Approved, 2=Rejected',
    reviewer_id BIGINT NULL COMMENT 'Reviewer user ID',
    review_time TIMESTAMP NULL COMMENT 'Review timestamp',
    review_notes VARCHAR(500) NULL COMMENT 'Reviewer notes/rejection reason',
    submit_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Submission timestamp',
    created_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Creation timestamp',
    updated_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Last update timestamp',
    KEY idx_user (user_id),
    KEY idx_property (property_unit_id),
    KEY idx_status (verification_status),
    KEY idx_submit_time (submit_time),
    CONSTRAINT fk_homeowner_verification_user FOREIGN KEY (user_id) REFERENCES auth_user(id) ON DELETE CASCADE,
    CONSTRAINT fk_homeowner_verification_property FOREIGN KEY (property_unit_id) REFERENCES auth_property_unit(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Homeowner verification requests';

-- ============================================================
-- Table: auth_family
-- Description: Family/household records
-- ============================================================
CREATE TABLE IF NOT EXISTS auth_family (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT 'Family ID',
    property_unit_id BIGINT NOT NULL COMMENT 'Property unit ID',
    family_head_id BIGINT NOT NULL COMMENT 'Family head user ID',
    family_name VARCHAR(100) NULL COMMENT 'Family name',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '1=Active, 2=Inactive',
    created_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Creation timestamp',
    updated_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Last update timestamp',
    delete_time TIMESTAMP NULL COMMENT 'Soft delete timestamp',
    UNIQUE KEY uk_property (property_unit_id),
    KEY idx_head (family_head_id),
    KEY idx_delete (delete_time),
    CONSTRAINT fk_family_property FOREIGN KEY (property_unit_id) REFERENCES auth_property_unit(id) ON DELETE CASCADE,
    CONSTRAINT fk_family_head FOREIGN KEY (family_head_id) REFERENCES auth_user(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Family records';

-- ============================================================
-- Table: auth_family_member
-- Description: Family member records
-- ============================================================
CREATE TABLE IF NOT EXISTS auth_family_member (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT 'Member ID',
    family_id BIGINT NOT NULL COMMENT 'Family ID',
    user_id BIGINT NULL COMMENT 'User ID (NULL if not registered)',
    name VARCHAR(50) NOT NULL COMMENT 'Member name',
    relationship VARCHAR(20) NOT NULL COMMENT 'Relationship to family head',
    phone VARCHAR(20) NULL COMMENT 'Phone number',
    id_card_number VARCHAR(18) NULL COMMENT 'ID card number',
    birth_date DATE NULL COMMENT 'Birth date',
    gender TINYINT NULL COMMENT '1=Male, 2=Female',
    created_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Creation timestamp',
    updated_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Last update timestamp',
    delete_time TIMESTAMP NULL COMMENT 'Soft delete timestamp',
    KEY idx_family (family_id),
    KEY idx_user (user_id),
    KEY idx_delete (delete_time),
    CONSTRAINT fk_family_member_family FOREIGN KEY (family_id) REFERENCES auth_family(id) ON DELETE CASCADE,
    CONSTRAINT fk_family_member_user FOREIGN KEY (user_id) REFERENCES auth_user(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Family member records';

-- ============================================================
-- Table: auth_uploaded_file
-- Description: File upload tracking (MinIO references)
-- ============================================================
CREATE TABLE IF NOT EXISTS auth_uploaded_file (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT 'File ID',
    user_id BIGINT NOT NULL COMMENT 'Uploader user ID',
    entity_type VARCHAR(50) NOT NULL COMMENT 'Entity type (e.g., homeowner_verification)',
    entity_id BIGINT NOT NULL COMMENT 'Entity ID',
    file_name VARCHAR(255) NOT NULL COMMENT 'Original file name',
    file_path VARCHAR(500) NOT NULL COMMENT 'MinIO object path',
    file_size BIGINT NOT NULL COMMENT 'File size in bytes',
    file_type VARCHAR(50) NOT NULL COMMENT 'MIME type',
    bucket_name VARCHAR(100) NOT NULL COMMENT 'MinIO bucket name',
    upload_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Upload timestamp',
    created_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Creation timestamp',
    KEY idx_user (user_id),
    KEY idx_entity (entity_type, entity_id),
    KEY idx_upload_time (upload_time),
    CONSTRAINT fk_uploaded_file_user FOREIGN KEY (user_id) REFERENCES auth_user(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='File upload tracking';
