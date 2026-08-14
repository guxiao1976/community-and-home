# 数据表规划方案

**制定日期**: 2026-06-24  
**项目**: community-and-home  
**目标**: 将所有数据表按微服务领域正确归属到对应的数据库

---

## 📊 现状分析

### 1. 当前数据库

```
✅ 已创建的数据库：
  - user              (用户服务)
  - auth              (认证服务)
  - permission        (权限服务)
  - community_hub_db  (社区服务)
  - moderation_db     (审核服务)
  - ai_model_db       (AI模型服务)
  - masterdata_db     (主数据服务)
  - default_db        (可能包含历史遗留表)
```

### 2. 微服务列表

```
services/
├── user-service/           - 用户服务
├── auth-service/           - 认证服务
├── permission-service/     - 权限服务
├── community-hub-service/  - 社区服务
├── file-service/           - 文件服务
├── moderation-service/     - 审核服务
├── master-data-service/    - 主数据服务
├── ai-model-service/       - AI模型服务
└── monitoring-service/     - 监控服务
```

---

## 📋 数据表规划方案

### 方案 A: 按 Model 文件分析的表归属

基于代码扫描和微服务领域划分：

#### 1. user 数据库

**服务**: user-service  
**职责**: 用户基本信息、用户资料、用户认证信息

**应包含的表**:

| 表名 | 说明 | Model 文件 |
|------|------|-----------|
| `users` 或 `user_base` | 用户基本信息 | user_base.go |
| `user_profiles` | 用户资料 | - |
| `user_certifications` | 用户认证信息 | user_certification.go |
| `user_community_memberships` | 用户社区成员关系 | user_community_membership.go |
| `user_membership_roles` | 用户成员角色 | user_membership_role.go |
| `user_residences` | 用户居住信息 | user_residence.go |
| `user_settings` | 用户设置 | - |

**表结构示例**:
```sql
CREATE TABLE users (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  phone VARCHAR(20) UNIQUE,
  nickname VARCHAR(50),
  avatar VARCHAR(255),
  status TINYINT,
  created_at TIMESTAMP,
  updated_at TIMESTAMP
);

CREATE TABLE user_certifications (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT,
  real_name VARCHAR(50),
  id_card VARCHAR(18),
  certification_type VARCHAR(20),
  status TINYINT,
  certified_at TIMESTAMP
);
```

---

#### 2. auth 数据库

**服务**: auth-service  
**职责**: 认证、授权、令牌管理

**应包含的表**:

| 表名 | 说明 | Model 文件 |
|------|------|-----------|
| `auth_credentials` | 用户凭证（密码等） | authcredential.go |
| `auth_tokens` | 访问令牌 | - |
| `refresh_tokens` | 刷新令牌 | - |
| `login_logs` | 登录日志 | - |
| `oauth_clients` | OAuth 客户端 | - |
| `oauth_access_tokens` | OAuth 令牌 | - |

**表结构示例**:
```sql
CREATE TABLE auth_credentials (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT UNIQUE,
  password_hash VARCHAR(255),
  salt VARCHAR(64),
  created_at TIMESTAMP,
  updated_at TIMESTAMP
);

CREATE TABLE auth_tokens (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT,
  access_token VARCHAR(512) UNIQUE,
  token_type VARCHAR(20) DEFAULT 'Bearer',
  expires_at TIMESTAMP,
  created_at TIMESTAMP
);

CREATE TABLE login_logs (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT,
  login_ip VARCHAR(45),
  login_device VARCHAR(100),
  login_status TINYINT,
  created_at TIMESTAMP
);
```

---

#### 3. permission 数据库

**服务**: permission-service  
**职责**: 权限、角色、RBAC 管理

**应包含的表**:

| 表名 | 说明 | Model 文件 |
|------|------|-----------|
| `permissions` | 权限定义 | permission.go |
| `roles` | 角色定义 | - |
| `role_permissions` | 角色权限关联 | - |
| `user_roles` | 用户角色关联 | - |
| `resource_permissions` | 资源权限 | - |
| `permission_groups` | 权限组 | - |
| `rel_*` | 关系表 | rel.go |

**表结构示例**:
```sql
CREATE TABLE permissions (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  code VARCHAR(100) UNIQUE,
  name VARCHAR(100),
  description TEXT,
  resource_type VARCHAR(50),
  action VARCHAR(50),
  created_at TIMESTAMP
);

CREATE TABLE roles (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  code VARCHAR(50) UNIQUE,
  name VARCHAR(100),
  description TEXT,
  level INT,
  created_at TIMESTAMP
);

CREATE TABLE role_permissions (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  role_id BIGINT,
  permission_id BIGINT,
  created_at TIMESTAMP,
  UNIQUE KEY uk_role_permission (role_id, permission_id)
);

CREATE TABLE user_roles (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT,
  role_id BIGINT,
  scope_type VARCHAR(50),
  scope_id BIGINT,
  created_at TIMESTAMP,
  UNIQUE KEY uk_user_role_scope (user_id, role_id, scope_type, scope_id)
);
```

---

#### 4. community_hub_db 数据库

**服务**: community-hub-service  
**职责**: 社区、帖子、评论、互动

**应包含的表**:

| 表名 | 说明 | Model 文件 |
|------|------|-----------|
| `communities` | 社区（小区）| - |
| `posts` | 帖子 | - |
| `comments` | 评论 | - |
| `likes` | 点赞 | - |
| `favorites` | 收藏 | - |
| `shares` | 分享 | - |
| `topics` | 话题 | - |
| `tags` | 标签 | - |
| `post_tags` | 帖子标签关联 | - |

**表结构示例**:
```sql
CREATE TABLE communities (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  name VARCHAR(100),
  description TEXT,
  region_code VARCHAR(20),
  address VARCHAR(255),
  member_count INT DEFAULT 0,
  status TINYINT,
  created_at TIMESTAMP,
  updated_at TIMESTAMP
);

CREATE TABLE posts (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  community_id BIGINT,
  user_id BIGINT,
  title VARCHAR(255),
  content TEXT,
  images JSON,
  view_count INT DEFAULT 0,
  like_count INT DEFAULT 0,
  comment_count INT DEFAULT 0,
  status TINYINT,
  created_at TIMESTAMP,
  updated_at TIMESTAMP
);

CREATE TABLE comments (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  post_id BIGINT,
  user_id BIGINT,
  parent_id BIGINT DEFAULT 0,
  content TEXT,
  like_count INT DEFAULT 0,
  status TINYINT,
  created_at TIMESTAMP
);
```

---

#### 5. moderation_db 数据库

**服务**: moderation-service  
**职责**: 内容审核、敏感词过滤、审核任务

**应包含的表**:

| 表名 | 说明 | Model 文件 |
|------|------|-----------|
| `moderation_tasks` | 审核任务 | - |
| `moderation_rules` | 审核规则 | - |
| `sensitive_words` | 敏感词库 | - |
| `audit_logs` | 审核日志 | - |
| `moderation_results` | 审核结果 | - |
| `pipelines` | 审核管道 | - |
| `pipeline_layers` | 管道层级 | - |

**表结构示例**:
```sql
CREATE TABLE moderation_tasks (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  content_type VARCHAR(50),
  content_id BIGINT,
  content TEXT,
  pipeline_id BIGINT,
  status VARCHAR(20),
  result VARCHAR(20),
  confidence FLOAT,
  labels JSON,
  created_at TIMESTAMP,
  reviewed_at TIMESTAMP
);

CREATE TABLE sensitive_words (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  word VARCHAR(100),
  category VARCHAR(50),
  level TINYINT,
  action VARCHAR(20),
  created_at TIMESTAMP
);

CREATE TABLE audit_logs (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  task_id BIGINT,
  reviewer_id BIGINT,
  action VARCHAR(50),
  reason TEXT,
  created_at TIMESTAMP
);
```

---

#### 6. ai_model_db 数据库

**服务**: ai-model-service  
**职责**: AI 模型管理、模板、预测记录

**应包含的表**:

| 表名 | 说明 | Model 文件 |
|------|------|-----------|
| `ai_models` | AI 模型 | - |
| `model_versions` | 模型版本 | - |
| `model_templates` | 模型模板 | - |
| `predictions` | 预测记录 | - |
| `training_jobs` | 训练任务 | - |
| `model_metrics` | 模型指标 | - |

**表结构示例**:
```sql
CREATE TABLE ai_models (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  name VARCHAR(100),
  type VARCHAR(50),
  provider VARCHAR(50),
  config JSON,
  status VARCHAR(20),
  created_at TIMESTAMP,
  updated_at TIMESTAMP
);

CREATE TABLE model_templates (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  name VARCHAR(100),
  model_id BIGINT,
  template TEXT,
  variables JSON,
  created_at TIMESTAMP
);

CREATE TABLE predictions (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  model_id BIGINT,
  input JSON,
  output JSON,
  confidence FLOAT,
  latency_ms INT,
  created_at TIMESTAMP
);
```

---

#### 7. masterdata_db 数据库

**服务**: master-data-service  
**职责**: 主数据、地区、分类、配置

**当前已有的表**:
```
✅ md_pipeline_test_work_item
✅ md_pipeline_test_work_record
✅ pipeline_test_work_items
✅ pipeline_test_work_records
```

**应包含的表**:

| 表名 | 说明 | Model 文件 |
|------|------|-----------|
| `regions` | 地区数据 | - |
| `provinces` | 省份 | - |
| `cities` | 城市 | - |
| `districts` | 区县 | - |
| `categories` | 分类 | - |
| `configs` | 系统配置 | - |
| `dictionaries` | 数据字典 | - |
| `residential_areas` | 小区数据 | - |

**表结构示例**:
```sql
CREATE TABLE regions (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  code VARCHAR(20) UNIQUE,
  name VARCHAR(100),
  parent_code VARCHAR(20),
  level TINYINT,
  full_name VARCHAR(255),
  created_at TIMESTAMP
);

CREATE TABLE residential_areas (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  name VARCHAR(100),
  region_code VARCHAR(20),
  address VARCHAR(255),
  building_count INT,
  household_count INT,
  property_company VARCHAR(100),
  created_at TIMESTAMP,
  updated_at TIMESTAMP
);
```

---

#### 8. file_db 数据库（建议创建）

**服务**: file-service  
**职责**: 文件元数据、存储信息

**应包含的表**:

| 表名 | 说明 | 备注 |
|------|------|------|
| `files` | 文件元数据 | - |
| `file_chunks` | 文件分块 | 大文件分片上传 |
| `file_access_logs` | 访问日志 | - |
| `storage_providers` | 存储提供商 | OSS/S3等 |

**表结构示例**:
```sql
CREATE TABLE files (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  filename VARCHAR(255),
  file_path VARCHAR(500),
  file_size BIGINT,
  mime_type VARCHAR(100),
  md5_hash VARCHAR(32),
  storage_provider VARCHAR(50),
  bucket VARCHAR(100),
  uploaded_by BIGINT,
  status TINYINT,
  created_at TIMESTAMP
);
```

---

## 🔄 数据迁移方案

### Phase 1: 调查现状（1天）

```bash
# 1. 列出 default_db 中的所有表
docker exec -i mysql mysql -uroot -proot123456 -e "SHOW TABLES FROM default_db;"

# 2. 查看每个表的结构
docker exec -i mysql mysql -uroot -proot123456 -e "SHOW CREATE TABLE default_db.table_name\G"

# 3. 确认表的数据量
docker exec -i mysql mysql -uroot -proot123456 -e "
  SELECT 
    TABLE_NAME, 
    TABLE_ROWS 
  FROM information_schema.TABLES 
  WHERE TABLE_SCHEMA='default_db';
"
```

### Phase 2: 生成迁移脚本（1天）

为每个表生成迁移 SQL：

```sql
-- 示例：迁移 users 表
-- 从 default_db 迁移到 user

-- 1. 创建表结构
CREATE TABLE user.users LIKE default_db.users;

-- 2. 复制数据
INSERT INTO user.users SELECT * FROM default_db.users;

-- 3. 验证
SELECT COUNT(*) FROM user.users;
SELECT COUNT(*) FROM default_db.users;

-- 4. 确认无误后删除原表
-- DROP TABLE default_db.users;
```

### Phase 3: 执行迁移（2-3天）

**迁移顺序**（按依赖关系）:

1. **无依赖的表先迁移**:
   - masterdata_db: regions, categories, configs
   - user: users

2. **有外键依赖的表**:
   - user: user_certifications, user_settings
   - auth: auth_credentials, auth_tokens
   - permission: permissions, roles

3. **关联表最后迁移**:
   - permission: role_permissions, user_roles
   - community: posts, comments

### Phase 4: 更新应用代码（1天）

更新服务配置，确保连接正确的数据库：

```yaml
# 已经正确配置，无需修改
# services/user-service/api/etc/user-api.yaml
DataSource: root:root123456@tcp(mysql:3306)/user?charset=utf8mb4
```

### Phase 5: 测试验证（1天）

```bash
# 1. 启动所有服务
docker-compose up -d

# 2. 测试各服务 API
curl http://localhost:8001/api/users
curl http://localhost:8002/api/auth/login

# 3. 检查数据一致性
# 对比迁移前后的数据条数
```

---

## 📊 迁移检查清单

### 迁移前检查

- [ ] 备份所有数据库
- [ ] 确认各服务配置正确
- [ ] 生成完整的迁移脚本
- [ ] 在测试环境验证

### 迁移中检查

- [ ] 按顺序执行迁移
- [ ] 验证每个表的数据完整性
- [ ] 记录迁移日志
- [ ] 出现问题立即回滚

### 迁移后检查

- [ ] 验证所有服务可正常启动
- [ ] 测试核心功能
- [ ] 验证数据一致性
- [ ] 删除 default_db 中的旧表（谨慎）

---

## 🎯 推荐执行方案

### 方案 A: 立即全量迁移（推荐用于开发环境）

**适用**: 开发环境，数据量小，可以停服

**步骤**:
1. 停止所有服务
2. 备份数据
3. 执行迁移脚本
4. 验证数据
5. 启动服务测试

**时间**: 1天

### 方案 B: 渐进式迁移（推荐用于生产环境）

**适用**: 生产环境，不能停服

**步骤**:
1. 双写模式（同时写入新旧数据库）
2. 灰度迁移（逐步切换读取）
3. 数据对账
4. 完全切换
5. 清理旧数据

**时间**: 1-2周

---

## 📄 生成的文件

1. **规划方案**: `.harness/changes/database-table-planning.md` (本文档)
2. **迁移脚本**: `scripts/migrate-tables.sql` (待生成)
3. **验证脚本**: `scripts/verify-migration.sh` (待生成)

---

## 💡 后续建议

### 1. 表命名规范

统一表命名风格：
- ✅ 推荐：`users`, `auth_tokens`, `role_permissions`
- ❌ 避免：`tb_users`, `t_users`, 混合大小写

### 2. 外键约束

谨慎使用外键：
- ✅ 微服务间不使用外键（通过 API 关联）
- ⚠️ 同服务内可以使用外键

### 3. 数据库用户权限

为每个服务创建独立的数据库用户：
```sql
CREATE USER 'user_service'@'%' IDENTIFIED BY 'password';
GRANT ALL PRIVILEGES ON user.* TO 'user_service'@'%';
```

### 4. 定期备份

每个数据库独立备份策略：
```bash
mysqldump -u root -p user > backup/user_$(date +%Y%m%d).sql
```

---

**状态**: ⚠️ 方案已制定，等待执行  
**优先级**: P1  
**预计工作量**: 5-7天（开发环境）/ 1-2周（生产环境）

---

**下一步**: 
1. 确认当前 default_db 中有哪些表
2. 根据实际情况调整本方案
3. 生成具体的迁移 SQL 脚本
