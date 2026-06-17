-- Migration 004: add moderation fields to user-service content tables
ALTER TABLE user_base
    ADD COLUMN nickname_moderation_status TINYINT NOT NULL DEFAULT 0
    COMMENT '0=待审核 1=机器通过 2=机器不通过 3=人审通过 4=人审不通过';

ALTER TABLE user_certification
    ADD COLUMN moderation_status TINYINT NOT NULL DEFAULT 0
    COMMENT '0=待审核 1=机器通过 2=机器不通过 3=人审通过 4=人审不通过',
    ADD COLUMN moderation_time DATETIME NULL COMMENT '审核时间';
