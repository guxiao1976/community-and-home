-- 添加 model_id 列到 am_api_key 表
ALTER TABLE am_api_key
ADD COLUMN model_id BIGINT DEFAULT 0 COMMENT '关联的模型配置ID' AFTER id,
ADD INDEX idx_model_id (model_id);

-- 更新现有记录：根据 provider 匹配第一个对应的模型
UPDATE am_api_key ak
LEFT JOIN am_model_config mc ON ak.provider = mc.provider
SET ak.model_id = COALESCE(mc.id, 0)
WHERE ak.model_id = 0;
