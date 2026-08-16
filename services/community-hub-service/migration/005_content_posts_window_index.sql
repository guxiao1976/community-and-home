-- Migration 005: content_posts 窗口过滤索引（mobile-homepage-content-revamp Task 1.2 / REQ-NTW-6 / ADR-3）
--
-- 30 天窗口谓词：WHERE scope.community_id=? AND status=approved AND published_at >= ? AND published_at <= ?
-- ... ORDER BY is_pinned DESC, published_at DESC。
-- 既有 idx_published(community_id, published_at DESC, deleted_at) 因 community_id 为弃用 NULL 列
-- 无法服务 scope JOIN 后的过滤/排序（REVISION #5），故新增本索引。
--
-- 索引列序 (status, is_pinned, published_at)：等值 status + 等值 is_pinned（低基数）+ 范围 published_at，
-- 同时覆盖 ORDER BY 前导列 is_pinned（等值列在前、范围列在后，读取天然有序，减少 filesort）。
--
-- 幂等守卫：MySQL 8.0 无 `ADD INDEX IF NOT EXISTS`，经 information_schema.statistics 检查存在性，
-- 已存在则仅输出提示不重复 ALTER（重复执行安全）。
-- 纯增量（additive）索引：不影响缺省（无 since_days）非窗口调用（REQ-NTW-6 场景 2）。
--
-- SEE: [[migration-must-execute]] — 执行 + SHOW INDEX 验证归 Task 3.1 Owner 运维验证
USE community_hub_db;

SET @db = DATABASE();
SET @idx_exists = (SELECT COUNT(*) FROM information_schema.statistics
    WHERE table_schema = @db AND table_name = 'content_posts' AND index_name = 'idx_status_pinned_published');
SET @sql = IF(@idx_exists = 0,
    'ALTER TABLE content_posts ADD INDEX idx_status_pinned_published (status, is_pinned, published_at)',
    'SELECT ''idx_status_pinned_published already exists''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
