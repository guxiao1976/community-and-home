-- Migration 002: add moderation fields to community-hub content tables
ALTER TABLE notices
    ADD COLUMN moderation_status TINYINT NOT NULL DEFAULT 0
    COMMENT '0=待审核 1=机器通过 2=机器不通过 3=人审通过 4=人审不通过',
    ADD COLUMN moderation_time DATETIME NULL COMMENT '审核时间';

ALTER TABLE lost_found_items
    ADD COLUMN moderation_status TINYINT NOT NULL DEFAULT 0
    COMMENT '0=待审核 1=机器通过 2=机器不通过 3=人审通过 4=人审不通过',
    ADD COLUMN moderation_time DATETIME NULL COMMENT '审核时间';
