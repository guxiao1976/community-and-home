-- 005_add_user_app_state.sql
-- 用户应用状态表：当前小区（账号级、跨设备一致）
-- SEE: [[migration-must-execute]]

CREATE TABLE user_app_state (
    user_id BIGINT PRIMARY KEY COMMENT '账号（雪花 ID，同 user_base.id）',
    current_community_id BIGINT NOT NULL DEFAULT 0 COMMENT '当前小区；0=未设置',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT '用户应用状态：当前小区（账号级、跨设备一致）';
