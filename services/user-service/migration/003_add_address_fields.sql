-- 003_add_address_fields.sql
-- 为 user_community_membership 表增加楼号/单元号/房号字段

ALTER TABLE user_community_membership
  ADD COLUMN building INT DEFAULT 0 COMMENT '楼号',
  ADD COLUMN unit INT DEFAULT 0 COMMENT '单元号',
  ADD COLUMN room INT DEFAULT 0 COMMENT '房号';

-- 索引加速地址唯一性查询
CREATE INDEX idx_community_address
  ON user_community_membership(community_id, building, unit, room);
