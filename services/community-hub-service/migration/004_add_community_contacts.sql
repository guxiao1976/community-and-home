-- Migration 004: community_contacts 幂等补表（mobile-homepage-content-revamp Task 1.1 / REQ-CLP-2）
--
-- 背景：运行库缺 community_contacts 表导致 ListContacts 报 `Table doesn't exist`。
-- 001_initial.sql 为该表 schema 单源（本文件 DDL 与 001 / model/community_contact.go 完全对齐）；
-- 004 是对「001 已应用后运行库仍缺表」的幂等补救（CREATE TABLE IF NOT EXISTS）。
--
-- 边界声明（REQ-CLP-2 场景 5）：若运行库表已存在但结构漂移（如缺列/列类型不一致），
-- IF NOT EXISTS 不自动修复，需人工订正（不静默掩盖，见 design.md §数据模型）。
-- 不预置种子数据（REQ-CLP-2 场景 3 空态，D4：运营后续维护）。
--
-- SEE: [[migration-must-execute]] — 执行 + DESCRIBE 验证归 Task 3.1 Owner 运维验证
USE community_hub_db;

CREATE TABLE IF NOT EXISTS community_contacts (
    id          BIGINT PRIMARY KEY COMMENT 'Snowflake ID',
    community_id BIGINT NOT NULL COMMENT '小区ID',
    category    VARCHAR(30) NOT NULL COMMENT '类别：water/electricity/gas/unicom/mobile/telecom/police',
    name        VARCHAR(100) NOT NULL COMMENT '显示名称',
    phone       VARCHAR(20) NOT NULL COMMENT '电话号码',
    sort_order  INT DEFAULT 0 COMMENT '排序',
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_community (community_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='便民联络（001 单源，004 幂等补救运行库缺失，不预置种子）';
