-- AI Model Service Database Schema
-- 创建时间: 2026-05-14
-- 说明: AI 模型统一管理服务数据库

-- 创建数据库
CREATE DATABASE IF NOT EXISTS ai_model_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE ai_model_db;

-- 1. 模型配置表
CREATE TABLE am_model_config (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    model_name VARCHAR(100) NOT NULL UNIQUE COMMENT '模型标识：claude-opus-4/gpt-4/qwen',
    model_type VARCHAR(20) NOT NULL COMMENT '模型类型：cloud/local',
    provider VARCHAR(50) NOT NULL COMMENT '提供商：claude/openai/ollama',
    display_name VARCHAR(200) COMMENT '显示名称',
    description TEXT COMMENT '模型描述',
    capabilities JSON COMMENT '能力列表：["text_generation","classification"]',

    -- 连接配置
    endpoint VARCHAR(500) COMMENT 'API端点',
    api_key_id BIGINT COMMENT '关联的API Key ID（外部模型）',
    timeout INT DEFAULT 60000 COMMENT '超时时间(ms)',
    max_retries INT DEFAULT 3 COMMENT '最大重试次数',

    -- 成本配置
    cost_per_1k_input_tokens DECIMAL(10, 6) DEFAULT 0 COMMENT '每1K输入token成本(USD)',
    cost_per_1k_output_tokens DECIMAL(10, 6) DEFAULT 0 COMMENT '每1K输出token成本(USD)',

    -- 配额和优先级
    priority INT DEFAULT 100 COMMENT '优先级：数字越小优先级越高',
    daily_quota INT DEFAULT 0 COMMENT '每日调用配额（0=不限制）',
    monthly_quota INT DEFAULT 0 COMMENT '每月调用配额（0=不限制）',

    -- 状态
    status TINYINT DEFAULT 1 COMMENT '状态：1=启用，0=禁用',
    health_status VARCHAR(20) DEFAULT 'unknown' COMMENT '健康状态：healthy/degraded/unhealthy/unknown',
    last_health_check TIMESTAMP NULL COMMENT '最后健康检查时间',

    created_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    delete_time TIMESTAMP NULL DEFAULT NULL,

    INDEX idx_provider (provider),
    INDEX idx_status (status),
    INDEX idx_priority (priority),
    INDEX idx_delete_time (delete_time)
) COMMENT '模型配置表';

-- 2. API密钥管理表
CREATE TABLE am_api_key (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    key_name VARCHAR(100) NOT NULL COMMENT '密钥名称',
    provider VARCHAR(50) NOT NULL COMMENT '提供商：claude/openai/ollama',
    api_key VARCHAR(500) NOT NULL COMMENT 'API密钥（AES加密存储）',
    masked_key VARCHAR(50) COMMENT '脱敏显示：sk-ant-***abc123',

    -- 配额和限制
    daily_quota INT DEFAULT 0 COMMENT '每日配额（0=无限制）',
    monthly_quota INT DEFAULT 0 COMMENT '每月配额（0=无限制）',
    priority INT DEFAULT 100 COMMENT '优先级：数字越小优先级越高',

    -- 状态
    status TINYINT DEFAULT 1 COMMENT '状态：1=启用，0=禁用',
    failure_count INT DEFAULT 0 COMMENT '连续失败次数',
    last_used_time TIMESTAMP NULL COMMENT '最后使用时间',
    last_failure_time TIMESTAMP NULL COMMENT '最后失败时间',

    created_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    delete_time TIMESTAMP NULL DEFAULT NULL,

    INDEX idx_provider (provider),
    INDEX idx_status (status),
    INDEX idx_priority (priority DESC),
    INDEX idx_delete_time (delete_time)
) COMMENT 'API密钥管理表';

-- 3. 健康检查记录表
CREATE TABLE am_health_check (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    model_id BIGINT NOT NULL COMMENT '模型配置ID',
    model_name VARCHAR(100) NOT NULL COMMENT '模型名称',
    status VARCHAR(20) NOT NULL COMMENT 'healthy/degraded/unhealthy',
    latency_ms BIGINT COMMENT '响应时间(ms)',
    error_msg TEXT COMMENT '错误信息',
    check_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_model_id (model_id),
    INDEX idx_model_time (model_name, check_time DESC),
    INDEX idx_check_time (check_time DESC)
) COMMENT '健康检查记录表';

-- 4. 调用日志表
CREATE TABLE am_call_log (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    request_id VARCHAR(64) COMMENT '请求ID（用于追踪）',
    model_id BIGINT NOT NULL COMMENT '模型配置ID',
    model_name VARCHAR(100) NOT NULL COMMENT '模型名称',
    caller_service VARCHAR(100) COMMENT '调用方服务名',

    -- 请求信息
    method VARCHAR(50) COMMENT '调用方法：CallModel/CallModelBatch',
    prompt_length INT COMMENT '提示词长度',

    -- Token和成本
    input_tokens INT DEFAULT 0,
    output_tokens INT DEFAULT 0,
    total_tokens INT DEFAULT 0,
    cost DECIMAL(10, 6) DEFAULT 0 COMMENT '本次调用成本(USD)',

    -- 性能
    latency_ms BIGINT COMMENT '耗时(ms)',
    success TINYINT DEFAULT 1 COMMENT '是否成功：1=成功，0=失败',
    error_code VARCHAR(50) COMMENT '错误码',
    error_msg TEXT COMMENT '错误信息',

    created_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_model_id (model_id),
    INDEX idx_model_time (model_name, created_time DESC),
    INDEX idx_caller (caller_service),
    INDEX idx_request_id (request_id),
    INDEX idx_created_time (created_time DESC),
    INDEX idx_success (success)
) COMMENT '调用日志表';

-- 5. 使用统计表（按天汇总）
CREATE TABLE am_usage_statistics (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    model_id BIGINT NOT NULL COMMENT '模型配置ID',
    model_name VARCHAR(100) NOT NULL COMMENT '模型名称',
    stat_date DATE NOT NULL COMMENT '统计日期',

    -- 调用统计
    total_calls INT DEFAULT 0,
    success_calls INT DEFAULT 0,
    failed_calls INT DEFAULT 0,

    -- Token统计
    total_tokens BIGINT DEFAULT 0,
    input_tokens BIGINT DEFAULT 0,
    output_tokens BIGINT DEFAULT 0,

    -- 成本统计
    total_cost DECIMAL(10, 2) DEFAULT 0,

    -- 性能统计
    avg_latency_ms BIGINT DEFAULT 0 COMMENT '平均耗时(ms)',
    p95_latency_ms BIGINT DEFAULT 0 COMMENT 'P95耗时(ms)',

    created_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uk_model_date (model_id, stat_date),
    INDEX idx_model_name (model_name),
    INDEX idx_stat_date (stat_date DESC)
) COMMENT '使用统计表';

-- 6. 提示词模板表
CREATE TABLE am_prompt_template (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    template_id VARCHAR(100) NOT NULL UNIQUE COMMENT '模板ID',
    template_name VARCHAR(200) NOT NULL COMMENT '模板名称',
    category VARCHAR(50) COMMENT '分类：classification/moderation/qa',
    content TEXT NOT NULL COMMENT '模板内容，支持{{variable}}变量',
    variables JSON COMMENT '变量列表：["words","categories"]',

    -- 版本管理
    version VARCHAR(20) DEFAULT 'v1.0',
    is_active TINYINT DEFAULT 1 COMMENT '是否激活：1=激活，0=未激活',

    -- 使用统计
    usage_count INT DEFAULT 0 COMMENT '使用次数',

    description TEXT COMMENT '模板说明',

    created_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    delete_time TIMESTAMP NULL DEFAULT NULL,

    INDEX idx_category (category),
    INDEX idx_active (is_active),
    INDEX idx_delete_time (delete_time)
) COMMENT '提示词模板表';

-- 7. 成本预警配置表
CREATE TABLE am_cost_alert_config (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    daily_limit DECIMAL(10, 2) DEFAULT 0 COMMENT '日成本上限（0=不限制）',
    monthly_limit DECIMAL(10, 2) DEFAULT 0 COMMENT '月成本上限（0=不限制）',
    alert_threshold DECIMAL(5, 2) DEFAULT 0.8 COMMENT '预警阈值：0.8表示达到80%时预警',

    -- 通知配置
    enable_notification TINYINT DEFAULT 1 COMMENT '是否启用通知',
    notification_channels JSON COMMENT '通知渠道：["log","dingtalk","email"]',

    status TINYINT DEFAULT 1 COMMENT '状态：1=启用，0=禁用',

    created_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) COMMENT '成本预警配置表';

-- 8. 预警记录表
CREATE TABLE am_alert_record (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    alert_type VARCHAR(20) NOT NULL COMMENT '预警类型：daily/monthly/quota',
    threshold DECIMAL(10, 2) NOT NULL COMMENT '阈值',
    actual_value DECIMAL(10, 2) NOT NULL COMMENT '实际值',
    message TEXT COMMENT '预警消息',

    -- 关联信息
    model_id BIGINT COMMENT '关联的模型ID（可选）',
    model_name VARCHAR(100) COMMENT '模型名称',

    -- 处理状态
    status VARCHAR(20) DEFAULT 'pending' COMMENT '状态：pending/notified/resolved',
    notified_time TIMESTAMP NULL COMMENT '通知时间',
    resolved_time TIMESTAMP NULL COMMENT '解决时间',

    alert_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_alert_time (alert_time DESC),
    INDEX idx_status (status),
    INDEX idx_model_id (model_id)
) COMMENT '预警记录表';

-- 插入初始数据

-- 默认成本预警配置
INSERT INTO am_cost_alert_config (daily_limit, monthly_limit, alert_threshold, notification_channels)
VALUES (100.00, 3000.00, 0.8, '["log"]');

-- 敏感词分类模板
INSERT INTO am_prompt_template (
    template_id, template_name, category, content, variables, version
) VALUES (
    'sensitive_word_classify', '敏感词分类模板', 'classification',
    '请对以下敏感词进行分类，分类包括：政治敏感、色情低俗、暴力恐怖、违法犯罪、人身攻击、广告营销、其他。

要求：
1. 严格按照JSON数组格式返回
2. 每个词必须分配一个分类
3. 格式：[{"word": "词1", "category": "分类1", "confidence": 0.95}, ...]

敏感词列表：
{{words}}',
    '["words"]', 'v1.0'
);
