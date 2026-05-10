-- Masterdata Service Seed Data
-- Database: masterdata_db
-- Description: Initial data including national administrative divisions

USE masterdata_db;

-- ============================================================
-- Seed Data: md_administrative_division
-- Description: Five-tier administrative division hierarchy (sample data)
-- Note: This is a subset. Full national data should be imported from official sources.
-- ============================================================

-- Level 1: Provinces (示例：北京、上海、广东)
INSERT INTO md_administrative_division (id, parent_id, level, name, code, path, sort_order, status, created_by) VALUES
(1, NULL, 1, '北京市', '110000', '/1/', 1, 1, 1),
(2, NULL, 1, '上海市', '310000', '/2/', 2, 1, 1),
(3, NULL, 1, '广东省', '440000', '/3/', 3, 1, 1);

-- Level 2: Cities under Beijing
INSERT INTO md_administrative_division (id, parent_id, level, name, code, path, sort_order, status, created_by) VALUES
(11, 1, 2, '北京市', '110100', '/1/11/', 1, 1, 1);

-- Level 2: Cities under Shanghai
INSERT INTO md_administrative_division (id, parent_id, level, name, code, path, sort_order, status, created_by) VALUES
(21, 2, 2, '上海市', '310100', '/2/21/', 1, 1, 1);

-- Level 2: Cities under Guangdong
INSERT INTO md_administrative_division (id, parent_id, level, name, code, path, sort_order, status, created_by) VALUES
(31, 3, 2, '广州市', '440100', '/3/31/', 1, 1, 1),
(32, 3, 2, '深圳市', '440300', '/3/32/', 2, 1, 1),
(33, 3, 2, '珠海市', '440400', '/3/33/', 3, 1, 1);

-- Level 3: Districts under Beijing
INSERT INTO md_administrative_division (id, parent_id, level, name, code, path, sort_order, status, created_by) VALUES
(111, 11, 3, '东城区', '110101', '/1/11/111/', 1, 1, 1),
(112, 11, 3, '西城区', '110102', '/1/11/112/', 2, 1, 1),
(113, 11, 3, '朝阳区', '110105', '/1/11/113/', 3, 1, 1),
(114, 11, 3, '海淀区', '110108', '/1/11/114/', 4, 1, 1);

-- Level 3: Districts under Shanghai
INSERT INTO md_administrative_division (id, parent_id, level, name, code, path, sort_order, status, created_by) VALUES
(211, 21, 3, '黄浦区', '310101', '/2/21/211/', 1, 1, 1),
(212, 21, 3, '徐汇区', '310104', '/2/21/212/', 2, 1, 1),
(213, 21, 3, '浦东新区', '310115', '/2/21/213/', 3, 1, 1);

-- Level 3: Districts under Guangzhou
INSERT INTO md_administrative_division (id, parent_id, level, name, code, path, sort_order, status, created_by) VALUES
(311, 31, 3, '越秀区', '440104', '/3/31/311/', 1, 1, 1),
(312, 31, 3, '天河区', '440106', '/3/31/312/', 2, 1, 1),
(313, 31, 3, '海珠区', '440105', '/3/31/313/', 3, 1, 1);

-- Level 3: Districts under Shenzhen
INSERT INTO md_administrative_division (id, parent_id, level, name, code, path, sort_order, status, created_by) VALUES
(321, 32, 3, '福田区', '440304', '/3/32/321/', 1, 1, 1),
(322, 32, 3, '南山区', '440305', '/3/32/322/', 2, 1, 1),
(323, 32, 3, '宝安区', '440306', '/3/32/323/', 3, 1, 1);

-- Level 4: Streets under Chaoyang District (Beijing)
INSERT INTO md_administrative_division (id, parent_id, level, name, code, path, sort_order, status, created_by) VALUES
(1131, 113, 4, '建国门街道', '110105001', '/1/11/113/1131/', 1, 1, 1),
(1132, 113, 4, '朝外街道', '110105002', '/1/11/113/1132/', 2, 1, 1),
(1133, 113, 4, '三里屯街道', '110105003', '/1/11/113/1133/', 3, 1, 1);

-- Level 4: Streets under Haidian District (Beijing)
INSERT INTO md_administrative_division (id, parent_id, level, name, code, path, sort_order, status, created_by) VALUES
(1141, 114, 4, '中关村街道', '110108001', '/1/11/114/1141/', 1, 1, 1),
(1142, 114, 4, '学院路街道', '110108002', '/1/11/114/1142/', 2, 1, 1);

-- Level 4: Streets under Pudong New District (Shanghai)
INSERT INTO md_administrative_division (id, parent_id, level, name, code, path, sort_order, status, created_by) VALUES
(2131, 213, 4, '陆家嘴街道', '310115001', '/2/21/213/2131/', 1, 1, 1),
(2132, 213, 4, '张江镇', '310115002', '/2/21/213/2132/', 2, 1, 1);

-- Level 4: Streets under Tianhe District (Guangzhou)
INSERT INTO md_administrative_division (id, parent_id, level, name, code, path, sort_order, status, created_by) VALUES
(3121, 312, 4, '天河南街道', '440106001', '/3/31/312/3121/', 1, 1, 1),
(3122, 312, 4, '珠江新城街道', '440106002', '/3/31/312/3122/', 2, 1, 1);

-- Level 4: Streets under Nanshan District (Shenzhen)
INSERT INTO md_administrative_division (id, parent_id, level, name, code, path, sort_order, status, created_by) VALUES
(3221, 322, 4, '南头街道', '440305001', '/3/32/322/3221/', 1, 1, 1),
(3222, 322, 4, '科技园街道', '440305002', '/3/32/322/3222/', 2, 1, 1);

-- Level 5: Communities under Zhongguancun Street (Beijing Haidian)
INSERT INTO md_administrative_division (id, parent_id, level, name, code, path, sort_order, status, created_by) VALUES
(11411, 1141, 5, '中关村东路社区', '110108001001', '/1/11/114/1141/11411/', 1, 1, 1),
(11412, 1141, 5, '中关村西区社区', '110108001002', '/1/11/114/1141/11412/', 2, 1, 1),
(11413, 1141, 5, '科育社区', '110108001003', '/1/11/114/1141/11413/', 3, 1, 1);

-- Level 5: Communities under Lujiazui Street (Shanghai Pudong)
INSERT INTO md_administrative_division (id, parent_id, level, name, code, path, sort_order, status, created_by) VALUES
(21311, 2131, 5, '陆家嘴社区', '310115001001', '/2/21/213/2131/21311/', 1, 1, 1),
(21312, 2131, 5, '东昌社区', '310115001002', '/2/21/213/2131/21312/', 2, 1, 1);

-- Level 5: Communities under Zhujiang New Town Street (Guangzhou Tianhe)
INSERT INTO md_administrative_division (id, parent_id, level, name, code, path, sort_order, status, created_by) VALUES
(31221, 3122, 5, '珠江新城社区', '440106002001', '/3/31/312/3122/31221/', 1, 1, 1),
(31222, 3122, 5, '华明社区', '440106002002', '/3/31/312/3122/31222/', 2, 1, 1);

-- Level 5: Communities under Keji Park Street (Shenzhen Nanshan)
INSERT INTO md_administrative_division (id, parent_id, level, name, code, path, sort_order, status, created_by) VALUES
(32221, 3222, 5, '科技园社区', '440305002001', '/3/32/322/3222/32221/', 1, 1, 1),
(32222, 3222, 5, '高新社区', '440305002002', '/3/32/322/3222/32222/', 2, 1, 1);

-- ============================================================
-- Seed Data: md_residential_area
-- Description: Sample residential area data (approved status)
-- ============================================================

INSERT INTO md_residential_area (id, county_id, street_id, community_div_id, code, name, address, area, population, community_type, submission_status, submitter_id, submit_time, reviewer_id, review_time) VALUES
(1, 114, 1141, 11411, 'RA-110108001001', '中关村东路社区居委会', '北京市海淀区中关村东路', 0.5, 5000, 1, 2, 1, NOW(), 1, NOW()),
(2, 114, 1141, 11412, 'RA-110108001002', '中关村西区社区居委会', '北京市海淀区中关村西区', 0.6, 6000, 1, 2, 1, NOW(), 1, NOW()),
(3, 213, 2131, 21311, 'RA-310115001001', '陆家嘴社区居委会', '上海市浦东新区陆家嘴', 0.8, 8000, 1, 2, 1, NOW(), 1, NOW()),
(4, 312, 3122, 31221, 'RA-440106002001', '珠江新城社区居委会', '广州市天河区珠江新城', 1.0, 10000, 1, 2, 1, NOW(), 1, NOW()),
(5, 322, 3222, 32221, 'RA-440305002001', '科技园社区居委会', '深圳市南山区科技园', 0.7, 7000, 1, 2, 1, NOW(), 1, NOW());

-- ============================================================
-- Seed Data: md_district_economic_data
-- Description: Sample economic data for districts (2024)
-- ============================================================

INSERT INTO md_district_economic_data (division_id, year, population, gdp, per_capita_income, unemployment_rate, data_source, created_by) VALUES
(113, 2024, 3500000, 850000.00, 120000.00, 2.5, '北京市统计局', 1),
(114, 2024, 4000000, 1200000.00, 150000.00, 2.3, '北京市统计局', 1),
(213, 2024, 5600000, 1500000.00, 130000.00, 2.8, '上海市统计局', 1),
(312, 2024, 1800000, 600000.00, 95000.00, 3.2, '广州市统计局', 1),
(322, 2024, 2000000, 800000.00, 110000.00, 2.9, '深圳市统计局', 1);

-- ============================================================
-- Seed Data: md_configuration
-- Description: Default platform configuration
-- ============================================================

INSERT INTO md_configuration (module, config_key, config_value, value_type, description, is_public, approval_status, created_by) VALUES
('认证管理', 'sms.rate_limit', '{"max_per_hour": 5, "max_per_day": 10}', 'json', '短信验证码发送频率限制', 0, 2, 1),
('认证管理', 'password.min_length', '8', 'number', '密码最小长度', 0, 2, 1),
('认证管理', 'password.require_special_char', 'true', 'boolean', '密码是否需要特殊字符', 0, 2, 1),
('认证管理', 'jwt.access_token_expire', '7200', 'number', 'JWT访问令牌过期时间（秒）', 0, 2, 1),
('认证管理', 'jwt.refresh_token_expire', '604800', 'number', 'JWT刷新令牌过期时间（秒）', 0, 2, 1),
('基础数据', 'community.auto_approve', 'false', 'boolean', '社区数据是否自动审批', 0, 2, 1),
('物业管理', 'verification.max_documents', '9', 'number', '业主认证最大文档数量', 0, 2, 1),
('物业管理', 'file.max_size', '5242880', 'number', '文件上传最大大小（字节）', 0, 2, 1);

-- ============================================================
-- Seed Data: md_sensitive_word
-- Description: Sample sensitive words for content moderation
-- ============================================================

INSERT INTO md_sensitive_word (word, category, severity, action, status, created_by) VALUES
('测试敏感词1', 'test', 1, 1, 1, 1),
('测试敏感词2', 'test', 2, 2, 1, 1),
('测试敏感词3', 'test', 3, 3, 1, 1);
