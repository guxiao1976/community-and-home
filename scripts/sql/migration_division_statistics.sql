-- Division Statistics Table
-- Pre-computed daily statistics for community/village counts per division
CREATE TABLE IF NOT EXISTS md_division_statistics (
    id              BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '统计记录ID',
    division_id     BIGINT NOT NULL COMMENT '区划ID',
    level           TINYINT NOT NULL COMMENT '1=省 2=市 3=区县 4=街道',
    community_count INT NOT NULL DEFAULT 0 COMMENT '小区数',
    village_count   INT NOT NULL DEFAULT 0 COMMENT '村数',
    total_count     INT NOT NULL DEFAULT 0 COMMENT '合计',
    stat_date       DATE NOT NULL COMMENT '统计日期',
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    UNIQUE KEY uk_division_date (division_id, stat_date),
    KEY idx_level_date (level, stat_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='区划社区村数量统计（每日快照）';
