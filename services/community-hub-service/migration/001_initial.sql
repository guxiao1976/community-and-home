-- 001_initial.sql — Community Hub Service 数据库初始化

CREATE DATABASE IF NOT EXISTS community_hub_db
  DEFAULT CHARACTER SET utf8mb4
  DEFAULT COLLATE utf8mb4_unicode_ci;

USE community_hub_db;

-- 通知公告
CREATE TABLE IF NOT EXISTS notices (
    id          BIGINT PRIMARY KEY,
    community_id BIGINT NOT NULL,
    title       VARCHAR(200) NOT NULL,
    content     TEXT NOT NULL,
    role        VARCHAR(20) NOT NULL,
    publisher   VARCHAR(100) NOT NULL,
    publisher_id BIGINT DEFAULT NULL,
    is_pinned   TINYINT DEFAULT 0,
    published_at DATETIME NOT NULL,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at  DATETIME DEFAULT NULL,
    INDEX idx_community (community_id, deleted_at),
    INDEX idx_published (community_id, published_at DESC, deleted_at)
);

-- 通知附件
CREATE TABLE IF NOT EXISTS notice_attachments (
    id        BIGINT PRIMARY KEY,
    notice_id BIGINT NOT NULL,
    file_name VARCHAR(200) NOT NULL,
    file_url  VARCHAR(500) NOT NULL,
    file_size BIGINT DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_notice (notice_id)
);

-- 便民联络
CREATE TABLE IF NOT EXISTS community_contacts (
    id          BIGINT PRIMARY KEY,
    community_id BIGINT NOT NULL,
    category    VARCHAR(30) NOT NULL,
    name        VARCHAR(100) NOT NULL,
    phone       VARCHAR(20) NOT NULL,
    sort_order  INT DEFAULT 0,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_community (community_id)
);

-- 寻失互助
CREATE TABLE IF NOT EXISTS lost_found_items (
    id            BIGINT PRIMARY KEY,
    community_id  BIGINT NOT NULL,
    type          VARCHAR(10) NOT NULL,
    title         VARCHAR(200) NOT NULL,
    description   TEXT,
    image_urls    JSON DEFAULT NULL,
    contact_phone VARCHAR(20),
    status        VARCHAR(20) DEFAULT 'active',
    publisher_id  BIGINT NOT NULL,
    created_at    DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at    DATETIME DEFAULT NULL,
    INDEX idx_community_type (community_id, type, status, deleted_at),
    INDEX idx_created (community_id, created_at DESC)
);
