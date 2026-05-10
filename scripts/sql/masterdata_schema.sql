-- Masterdata Service Database Schema
-- Database: masterdata_db
-- MySQL 8.0+
-- Character Set: utf8mb4
-- Collation: utf8mb4_unicode_ci

CREATE DATABASE IF NOT EXISTS masterdata_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE masterdata_db;

-- ============================================================
-- Table: md_administrative_division
-- Description: Five-tier administrative division hierarchy
-- ============================================================
CREATE TABLE IF NOT EXISTS md_administrative_division (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT 'Division ID',
    parent_id BIGINT NULL COMMENT 'Parent division ID (NULL=root)',
    level TINYINT NOT NULL COMMENT '1=Province, 2=City, 3=District, 4=Street, 5=Community',
    name VARCHAR(100) NOT NULL COMMENT 'Division name',
    code VARCHAR(20) NOT NULL COMMENT 'Administrative code',
    path VARCHAR(500) NOT NULL COMMENT 'Materialized path (e.g., /1/23/456/)',
    sort_order INT NOT NULL DEFAULT 0 COMMENT 'Display order',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '1=Active, 2=Inactive',
    submission_status TINYINT NOT NULL DEFAULT 0 COMMENT '0=Pending, 1=Submitted, 2=Approved, 3=Rejected',
    submission_type TINYINT NULL COMMENT '1=Create, 2=Update, 3=Delete',
    change_snapshot JSON NULL COMMENT 'Snapshot of values before modification',
    submitter_id BIGINT NULL COMMENT 'Submitter user ID',
    submit_time TIMESTAMP NULL COMMENT 'Submission timestamp',
    reviewer_id BIGINT NULL COMMENT 'Reviewer user ID',
    review_time TIMESTAMP NULL COMMENT 'Review timestamp',
    review_notes VARCHAR(500) NULL COMMENT 'Review notes/rejection reason',
    created_by BIGINT NOT NULL COMMENT 'Creator user ID',
    created_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Creation timestamp',
    updated_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Last update timestamp',
    delete_time TIMESTAMP NULL COMMENT 'Soft delete timestamp',
    UNIQUE KEY uk_code (code),
    KEY idx_parent (parent_id),
    KEY idx_level (level),
    KEY idx_path (path),
    KEY idx_delete (delete_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Administrative division hierarchy';

-- ============================================================
-- Table: md_residential_area
-- Description: Residential area (小区) master data
-- ============================================================
DROP TABLE IF EXISTS md_community;
CREATE TABLE IF NOT EXISTS md_residential_area (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT 'Residential area ID',
    county_id BIGINT NULL COMMENT 'District/county division ID (level 3)',
    street_id BIGINT NULL COMMENT 'Street division ID (level 4)',
    community_div_id BIGINT NULL COMMENT 'Community division ID (level 5)',
    code VARCHAR(100) NOT NULL COMMENT 'Residential area unique code',
    name VARCHAR(100) NOT NULL COMMENT 'Residential area name',
    address VARCHAR(255) NOT NULL COMMENT 'Full address',
    longitude DECIMAL(10,7) NULL COMMENT 'Longitude',
    latitude DECIMAL(10,7) NULL COMMENT 'Latitude',
    data_source TINYINT NOT NULL DEFAULT 0 COMMENT '0=Manual, 1=AMap API',
    area DECIMAL(10,2) NULL COMMENT 'Area in square kilometers',
    population INT NULL COMMENT 'Population count',
    community_type TINYINT NOT NULL COMMENT '1=Residential, 2=Village, 3=Mixed',
    submission_status TINYINT NOT NULL DEFAULT 0 COMMENT '0=Draft, 1=Submitted, 2=Approved, 3=Rejected',
    submitter_id BIGINT NOT NULL COMMENT 'Submitter user ID',
    submit_time TIMESTAMP NULL COMMENT 'Submission timestamp',
    reviewer_id BIGINT NULL COMMENT 'Reviewer user ID',
    review_time TIMESTAMP NULL COMMENT 'Review timestamp',
    review_notes VARCHAR(500) NULL COMMENT 'Review notes/rejection reason',
    created_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Creation timestamp',
    updated_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Last update timestamp',
    delete_time TIMESTAMP NULL COMMENT 'Soft delete timestamp',
    UNIQUE KEY uk_code (code),
    KEY idx_county (county_id),
    KEY idx_street (street_id),
    KEY idx_community_div (community_div_id),
    KEY idx_status (submission_status),
    KEY idx_submitter (submitter_id),
    KEY idx_delete (delete_time),
    CONSTRAINT fk_ra_county FOREIGN KEY (county_id) REFERENCES md_administrative_division(id) ON DELETE SET NULL,
    CONSTRAINT fk_ra_street FOREIGN KEY (street_id) REFERENCES md_administrative_division(id) ON DELETE SET NULL,
    CONSTRAINT fk_ra_community_div FOREIGN KEY (community_div_id) REFERENCES md_administrative_division(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Residential area master data';

-- ============================================================
-- Table: md_district_economic_data
-- Description: County/district economic and population data
-- ============================================================
CREATE TABLE IF NOT EXISTS md_district_economic_data (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT 'Data ID',
    division_id BIGINT NOT NULL COMMENT 'Administrative division ID (level 3)',
    year INT NOT NULL COMMENT 'Data year',
    population BIGINT NULL COMMENT 'Total population',
    gdp DECIMAL(15,2) NULL COMMENT 'GDP in million CNY',
    per_capita_income DECIMAL(10,2) NULL COMMENT 'Per capita income in CNY',
    unemployment_rate DECIMAL(5,2) NULL COMMENT 'Unemployment rate (%)',
    data_source VARCHAR(255) NULL COMMENT 'Data source',
    created_by BIGINT NOT NULL COMMENT 'Creator user ID',
    created_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Creation timestamp',
    updated_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Last update timestamp',
    UNIQUE KEY uk_division_year (division_id, year),
    KEY idx_division (division_id),
    KEY idx_year (year),
    CONSTRAINT fk_economic_data_division FOREIGN KEY (division_id) REFERENCES md_administrative_division(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='District economic data';

-- ============================================================
-- Table: md_configuration
-- Description: Platform-wide configurable parameters
-- ============================================================
CREATE TABLE IF NOT EXISTS md_configuration (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT 'Configuration ID',
    module VARCHAR(50) NOT NULL COMMENT 'Module name (e.g., auth, masterdata)',
    config_key VARCHAR(100) NOT NULL COMMENT 'Configuration key',
    config_value TEXT NOT NULL COMMENT 'Configuration value (JSON)',
    value_type VARCHAR(20) NOT NULL COMMENT 'string/number/boolean/json',
    description VARCHAR(255) NULL COMMENT 'Configuration description',
    is_public TINYINT NOT NULL DEFAULT 0 COMMENT '1=Public (visible to all), 0=Internal',
    approval_status TINYINT NOT NULL DEFAULT 0 COMMENT '0=Draft, 1=Pending, 2=Approved',
    created_by BIGINT NOT NULL COMMENT 'Creator user ID',
    created_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Creation timestamp',
    updated_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Last update timestamp',
    delete_time TIMESTAMP NULL COMMENT 'Soft delete timestamp',
    UNIQUE KEY uk_module_key (module, config_key),
    KEY idx_module (module),
    KEY idx_status (approval_status),
    KEY idx_delete (delete_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Platform configuration';

-- ============================================================
-- Table: md_sensitive_word
-- Description: Sensitive word list for content moderation
-- ============================================================
CREATE TABLE IF NOT EXISTS md_sensitive_word (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT 'Word ID',
    word VARCHAR(100) NOT NULL COMMENT 'Sensitive word',
    category VARCHAR(50) NOT NULL COMMENT 'Category (e.g., political, violence)',
    severity TINYINT NOT NULL COMMENT '1=Low, 2=Medium, 3=High',
    action TINYINT NOT NULL COMMENT '1=Warn, 2=Block, 3=Review',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '1=Active, 2=Inactive',
    created_by BIGINT NOT NULL COMMENT 'Creator user ID',
    created_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Creation timestamp',
    updated_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Last update timestamp',
    UNIQUE KEY uk_word (word),
    KEY idx_category (category),
    KEY idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Sensitive word list';

-- ============================================================
-- Table: md_audit_log
-- Description: Audit log for all significant data changes
-- ============================================================
CREATE TABLE IF NOT EXISTS md_audit_log (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT 'Log ID',
    user_id BIGINT NOT NULL COMMENT 'User who made the change',
    entity_type VARCHAR(50) NOT NULL COMMENT 'Entity type (table name)',
    entity_id BIGINT NOT NULL COMMENT 'Entity ID',
    action VARCHAR(20) NOT NULL COMMENT 'CREATE/UPDATE/DELETE',
    old_value TEXT NULL COMMENT 'Old value (JSON)',
    new_value TEXT NULL COMMENT 'New value (JSON)',
    ip_address VARCHAR(45) NOT NULL COMMENT 'Client IP address',
    user_agent VARCHAR(255) NULL COMMENT 'Client user agent',
    created_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Action timestamp',
    KEY idx_user (user_id),
    KEY idx_entity (entity_type, entity_id),
    KEY idx_action (action),
    KEY idx_created_time (created_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Audit log';

-- Table: md_submission_record
-- Tracks submission/approval lifecycle for all masterdata entities
CREATE TABLE IF NOT EXISTS md_submission_record (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT 'Record ID',
    entity_type VARCHAR(50) NOT NULL COMMENT 'Entity type (administrative_division/residential_area/configuration/sensitive_word)',
    entity_id BIGINT NOT NULL COMMENT 'Entity ID (kept even if entity is deleted)',
    entity_name VARCHAR(200) NULL COMMENT 'Entity name (snapshot at submit time)',
    entity_code VARCHAR(100) NULL COMMENT 'Entity code (snapshot at submit time)',
    submission_type TINYINT NOT NULL COMMENT '1=Create, 2=Update, 3=Delete',
    submitter_id BIGINT NOT NULL COMMENT 'Submitter user ID',
    submit_time TIMESTAMP NOT NULL COMMENT 'Submission timestamp',
    reviewer_id BIGINT NULL COMMENT 'Reviewer user ID',
    review_time TIMESTAMP NULL COMMENT 'Review timestamp',
    review_result TINYINT NOT NULL DEFAULT 0 COMMENT '0=Pending, 1=Approved, 2=Rejected, 3=Withdrawn',
    review_notes VARCHAR(500) NULL COMMENT 'Review notes/rejection reason',
    created_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Record created time',
    KEY idx_submitter (submitter_id),
    KEY idx_reviewer (reviewer_id),
    KEY idx_entity (entity_type, entity_id),
    KEY idx_result (review_result)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Submission/approval records';
