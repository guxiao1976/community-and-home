-- Migration 003: notices → content_posts 通用化 + scope 多小区 + 附件演化 + Kafka 待推列
--
-- ⚠️ 003 为一次性 RENAME 迁移（R4）：MySQL 无 `RENAME TABLE IF EXISTS`，本迁移仅执行一次，勿重跑；
--    重跑报错为预期（与 001/002 幂等风格不同属有意为之，见 design.md §数据模型）。
-- ⚠️ 部分失败恢复（V5）：003 为单脚本内串联（RENAME → 多次 ALTER → CREATE TABLE → MODIFY）。
--    若中途某条 ALTER 失败，表已 RENAME 但后续语句未执行，处于「已 RENAME、缺新列」的半完成态——
--    禁止直接重跑完整脚本（RENAME 已发生会报「表不存在/重名」）。恢复指引：
--    先 `RENAME TABLE content_posts TO notices`（或按失败语句前状态手动对齐表结构）回到可重入状态，
--    再修复后重跑完整 003。

USE community_hub_db;

-- 1. notices → content_posts RENAME + content→text + section_code/status/attachment_count（REQ-CPB-1）
RENAME TABLE notices TO content_posts;
ALTER TABLE content_posts
    CHANGE COLUMN content `text` TEXT NOT NULL COMMENT '一段文字（图文发布核心，原 content，D1；反引号包裹防解析歧义，评审 I7；REST wire 仍以 content 键输出，见 design.md §REST wire 兼容）',
    ADD COLUMN section_code VARCHAR(30) NOT NULL DEFAULT 'notice' COMMENT '板块：notice=通知/repair=维修保修/...（D11）',
    ADD COLUMN status TINYINT NOT NULL DEFAULT 0 COMMENT '全生命周期+审核结果：0=draft 1=submitted 2=approved 3=rejected 4=withdrawn（REVISION，权威契约）',
    ADD COLUMN attachment_count INT NOT NULL DEFAULT 0 COMMENT '附件计数（审核完整性判定载体，D15）';

-- 2. published_at / community_id 去 NOT NULL（REQ-CPB-1；D1 迁移语义，审核锚定；弃用列不写入）
ALTER TABLE content_posts MODIFY published_at DATETIME DEFAULT NULL COMMENT '审核锚定：本期 submit 即置 NOW()（隐式通过 D16）；消费者上线后按审核结果覆盖（D27）';
ALTER TABLE content_posts MODIFY community_id BIGINT DEFAULT NULL COMMENT '弃用：范围关联单源 content_post_scope（兼容期保留列，不写入，REUSE:notice-D1）';

-- 3. Kafka at-least-once 待推标记（D20：落库待推 + 定时重推；推送失败不阻塞发布但登记 + 可观测）
ALTER TABLE content_posts
    ADD COLUMN kafka_push_status TINYINT NOT NULL DEFAULT 0 COMMENT '0=无待推 1=pending-push 2=已推(ack)',
    ADD COLUMN kafka_push_retries INT NOT NULL DEFAULT 0 COMMENT '重推次数',
    ADD COLUMN kafka_push_last_error VARCHAR(500) DEFAULT NULL COMMENT '最近一次推送错误摘要（可观测）',
    ADD COLUMN kafka_pushed_at DATETIME DEFAULT NULL COMMENT '成功推送时间';

-- 4. content_post_scope 多小区关联（REQ-CPB-2，复用 notice_scope 模式）
CREATE TABLE IF NOT EXISTS content_post_scope (
    post_id      BIGINT NOT NULL COMMENT 'content_posts.id',
    community_id BIGINT NOT NULL COMMENT '目标小区（md_residential_area.id，代表小区或村）',
    created_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (post_id, community_id),              -- 唯一约束（同一小区只一条）
    KEY idx_scope_community (community_id, post_id)   -- 读路径：列表/跑马灯按 community_id 先过滤
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='内容帖-小区范围关联（多小区发布单源；纯关联表仅 created_at——显式偏离编码规范 §3.1 的 updated_at/deleted_at，理由见 design.md §字段约束补充）';

-- 5. notice_attachments → content_post_attachments + post_id/review_status/file_id/file_type（REQ-CPB-3）
RENAME TABLE notice_attachments TO content_post_attachments;
ALTER TABLE content_post_attachments
    CHANGE COLUMN notice_id post_id BIGINT NOT NULL COMMENT '关联 content_posts.id（post_id 全链一致，无 main_id 别名）',
    ADD COLUMN review_status TINYINT NOT NULL DEFAULT 1 COMMENT '附件级审核：0=pending 1=approved 2=rejected（本期默认 approved，D14）',
    ADD COLUMN file_id BIGINT DEFAULT 0 COMMENT 'file-service 文件ID（重生预签名 URL 的权威载体，D14/REUSE:notice-D24）；兼容期存量行 0，读路径回退 stored file_url',
    ADD COLUMN file_type VARCHAR(20) DEFAULT NULL COMMENT '白名单校验通过的文件类型（扩展名，自 FileInfo 回读）';
ALTER TABLE content_post_attachments MODIFY file_url VARCHAR(1024) NOT NULL COMMENT '存量回退用 stored URL（新行写占位空串 ''，file_id 为权威重生载体；加宽防 ERROR 1406）';
