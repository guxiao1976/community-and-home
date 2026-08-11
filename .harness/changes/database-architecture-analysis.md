# 数据库架构问题分析报告

**问题**: 按照微服务架构，每个微服务都应该有独立的数据库，但用户管理（user-service）为什么没有单独的数据库？

**分析日期**: 2026-06-24  
**项目**: community-and-home

---

## 📊 当前数据库状况

### 1. 现有数据库

从 MySQL 查询结果：

```
业务数据库：
  - default_db       (表数量: 待确认)
  - masterdata_db    (表数量: 待确认)

系统数据库：
  - mysql
  - information_schema
  - performance_schema
  - sys
```

### 2. 微服务列表

从项目结构看，有以下微服务：

```
services/
├── user-service/          - 用户服务 ⚠️
├── auth-service/          - 认证服务 ⚠️
├── permission-service/    - 权限服务 ⚠️
├── community-hub-service/ - 社区服务
├── file-service/          - 文件服务
├── moderation-service/    - 审核服务
├── master-data-service/   - 主数据服务 ✅ (有 masterdata_db)
├── ai-model-service/      - AI 模型服务
└── monitoring-service/    - 监控服务
```

---

## 🔍 问题分析

### 问题 1: 缺失的数据库

按照微服务标准，应该有但**缺失**的数据库：

| 服务 | 应有数据库 | 实际状态 | 影响 |
|------|-----------|---------|------|
| **user-service** | `user_db` 或 `user` | ❌ 缺失 | 用户数据存储位置不明确 |
| **auth-service** | `auth_db` 或 `auth` | ❌ 缺失 | 认证数据存储位置不明确 |
| **permission-service** | `permission_db` 或 `permission` | ❌ 缺失 | 权限数据存储位置不明确 |
| **community-hub-service** | `community_hub_db` | ❌ 缺失 | 社区数据存储位置不明确 |
| **file-service** | `file_db` | ❌ 缺失 | 文件元数据存储位置不明确 |
| **moderation-service** | `moderation_db` | ❌ 缺失 | 审核数据存储位置不明确 |
| **ai-model-service** | `ai_model_db` | ❌ 缺失 | AI 模型数据存储位置不明确 |

### 问题 2: 可能的原因

#### 原因 A: 所有表都在 default_db 中（反模式）

**可能情况**:
```
default_db/
├── users            (user-service 的表)
├── auth_tokens      (auth-service 的表)
├── permissions      (permission-service 的表)
├── posts            (community-service 的表)
└── ...              (其他服务的表混在一起)
```

**问题**:
- ❌ 违反微服务数据独立性原则
- ❌ 服务间耦合严重
- ❌ 难以独立部署和扩展
- ❌ 数据库成为单点故障

#### 原因 B: 数据库未创建（开发阶段）

**可能情况**:
- 项目处于早期开发阶段
- 数据库创建脚本未执行
- docker-compose.yml 只创建了 default_db

**证据**:
```yaml
# docker-compose.yml
environment:
  MYSQL_DATABASE: default_db  # 只创建了一个数据库
```

#### 原因 C: Identity 服务模式

**可能情况**:
- user-service, auth-service, permission-service 共享一个 `identity_db`
- 将用户、认证、权限视为一个 Identity 领域

**是否合理**:
- ⚠️ 部分合理（这三个服务关联性强）
- ❌ 但违反了微服务独立性原则

---

## 🎯 标准微服务数据库架构

### 应该是什么样？

```
微服务                     独立数据库                  数据内容
────────────────────────────────────────────────────────────────
user-service        →      user_db              用户基本信息
                           ├── users            用户表
                           ├── user_profiles    用户资料
                           └── user_settings    用户设置

auth-service        →      auth_db              认证信息
                           ├── auth_tokens      令牌
                           ├── refresh_tokens   刷新令牌
                           └── login_logs       登录日志

permission-service  →      permission_db        权限信息
                           ├── permissions      权限
                           ├── roles            角色
                           └── role_permissions 角色权限关联

community-hub       →      community_hub_db     社区数据
                           ├── posts            帖子
                           ├── comments         评论
                           └── likes            点赞

file-service        →      file_db              文件元数据
                           ├── files            文件信息
                           └── file_chunks      分块信息

moderation-service  →      moderation_db        审核数据
                           ├── review_tasks     审核任务
                           └── audit_logs       审核日志

ai-model-service    →      ai_model_db          AI 模型
                           ├── models           模型信息
                           └── predictions      预测记录

master-data         →      masterdata_db        主数据
                           ├── regions          地区
                           ├── categories       分类
                           └── configs          配置
```

---

## 💡 建议方案

### 方案 1: 创建完整的数据库（推荐）

**步骤**:

1. **创建数据库 SQL**:

```sql
-- 创建各服务的数据库
CREATE DATABASE IF NOT EXISTS user_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE IF NOT EXISTS auth_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE IF NOT EXISTS permission_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE IF NOT EXISTS community_hub_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE IF NOT EXISTS file_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE IF NOT EXISTS moderation_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE IF NOT EXISTS ai_model_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 创建各服务的用户（可选，更安全）
CREATE USER IF NOT EXISTS 'user_service'@'%' IDENTIFIED BY 'user_pass_123';
CREATE USER IF NOT EXISTS 'auth_service'@'%' IDENTIFIED BY 'auth_pass_123';
-- ... 其他服务

-- 授权
GRANT ALL PRIVILEGES ON user_db.* TO 'user_service'@'%';
GRANT ALL PRIVILEGES ON auth_db.* TO 'auth_service'@'%';
-- ... 其他服务

FLUSH PRIVILEGES;
```

2. **更新服务配置**:

```yaml
# services/user-service/api/etc/user-api.yaml
DataSource: root:root123456@tcp(mysql:3306)/user_db?charset=utf8mb4&parseTime=true

# services/auth-service/rpc/etc/authservice.yaml
DataSource: root:root123456@tcp(mysql:3306)/auth_db?charset=utf8mb4&parseTime=true

# services/permission-service/rpc/etc/permissionservice.yaml
DataSource: root:root123456@tcp(mysql:3306)/permission_db?charset=utf8mb4&parseTime=true
```

3. **更新 docker-compose.yml**:

```yaml
mysql:
  image: mysql:8.0
  environment:
    MYSQL_ROOT_PASSWORD: root123456
    # 不设置 MYSQL_DATABASE，手动创建多个数据库
  volumes:
    - ./data/mysql-data:/var/lib/mysql
    - ./scripts/init-databases.sql:/docker-entrypoint-initdb.d/init.sql
```

4. **创建初始化脚本**:

```bash
# scripts/init-databases.sql
CREATE DATABASE IF NOT EXISTS user_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE IF NOT EXISTS auth_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
-- ... 其他数据库
```

### 方案 2: Identity 服务整合（折中）

如果 user、auth、permission 关联性很强，可以合并：

```sql
CREATE DATABASE IF NOT EXISTS identity_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- identity_db 包含：
-- - users (用户)
-- - auth_tokens (认证)
-- - permissions (权限)
-- - roles (角色)
```

**适用场景**:
- user、auth、permission 服务经常一起部署
- 它们之间有大量联合查询
- 团队认为拆分过细

**仍需创建的数据库**:
- community_hub_db
- file_db
- moderation_db
- ai_model_db

---

## 🚨 当前问题的影响

### 1. 违反微服务原则

| 原则 | 现状 | 影响 |
|------|-----|------|
| 数据独立性 | ❌ 违反 | 服务耦合 |
| 独立部署 | ❌ 困难 | 无法独立上线 |
| 故障隔离 | ❌ 缺失 | 一个服务故障影响全部 |
| 技术异构 | ❌ 受限 | 无法使用不同数据库 |

### 2. 运维问题

- ❌ 无法对单个服务的数据库进行独立备份
- ❌ 无法对单个服务的数据库进行独立优化
- ❌ 无法对单个服务的数据库进行独立扩展
- ❌ 数据库成为性能瓶颈

### 3. 开发问题

- ❌ 服务间可能直接访问其他服务的表
- ❌ 表结构变更影响范围大
- ❌ 数据迁移复杂

---

## ✅ 行动计划

### Phase 1: 调查现状（立即）

```bash
# 1. 检查 default_db 中有哪些表
docker exec -i mysql mysql -uroot -proot123456 -e "SHOW TABLES FROM default_db;"

# 2. 检查各服务的配置文件
find services -name "*.yaml" -exec grep -H "DataSource" {} \;

# 3. 确认表的归属
docker exec -i mysql mysql -uroot -proot123456 default_db -e "SHOW TABLES;"
```

### Phase 2: 创建数据库（1天）

1. 编写 `scripts/create-databases.sql`
2. 执行 SQL 创建数据库
3. 验证数据库创建成功

### Phase 3: 迁移数据（2-3天）

1. 将 default_db 中的表按服务归属迁移到各自的数据库
2. 更新服务配置文件
3. 测试服务连接

### Phase 4: 更新部署配置（1天）

1. 更新 docker-compose.yml
2. 更新初始化脚本
3. 文档更新

---

## 📋 检查清单

- [ ] 确认 default_db 中有哪些表
- [ ] 确认各服务当前使用的数据库
- [ ] 创建缺失的数据库
- [ ] 迁移表到对应的数据库
- [ ] 更新服务配置
- [ ] 测试服务连接
- [ ] 更新部署脚本
- [ ] 更新文档

---

## 🎓 总结

### 问题回答

**Q: 用户管理属于哪个微服务？**  
**A**: 属于 `user-service` 微服务

**Q: 为什么没有单独的数据库？**  
**A**: 可能原因：
1. 项目早期开发阶段，未创建完整的数据库
2. 所有表都放在 default_db 中（反模式）
3. 配置遗漏

### 正确做法

✅ **应该创建独立的数据库**:
- `user_db` - 用户服务
- `auth_db` - 认证服务
- `permission_db` - 权限服务
- 或者 `identity_db` - 整合用户、认证、权限

✅ **遵循微服务数据独立性原则**:
- 每个服务有自己的数据库
- 服务间通过 API 通信，不直接访问数据库
- 数据独立备份、扩展、优化

---

**状态**: ⚠️ 需要修复  
**优先级**: P1（影响架构合理性）  
**预计工作量**: 3-5天  

---

**建议**: 立即执行 Phase 1 调查现状，确认问题后执行数据库创建和迁移。
