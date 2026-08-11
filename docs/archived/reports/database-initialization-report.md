# 数据库初始化完成报告

## 执行时间
2026-07-12

## 问题诊断

### 原因分析
1. **auth-service 缺少 migration 脚本** - 没有创建 `auth_credential` 表的初始化脚本
2. **user-service 的 migration 脚本未执行** - 有迁移脚本但未运行
3. **001/002 脚本是增量迁移** - 假设已有旧表结构，不适用于全新安装

## 解决方案

### 1. 创建的文件

#### auth-service
```
services/auth-service/migration/001_initial.sql
services/auth-service/README.md
```

#### user-service
```
services/user-service/migration/000_initial_schema.sql
services/user-service/README.md
```

### 2. 执行的 SQL 脚本

#### auth 数据库
```sql
-- 创建登录凭证表
CREATE TABLE auth_credential (
    id              BIGINT NOT NULL AUTO_INCREMENT,
    user_id         BIGINT NOT NULL,
    identity_type   VARCHAR(20) NOT NULL,
    identifier      VARCHAR(255) NOT NULL,  -- AES 加密的手机号
    credential      VARCHAR(255) NOT NULL,  -- bcrypt 密文
    created_time    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_time    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE INDEX uk_identity (identity_type, identifier),
    INDEX idx_user (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

**执行结果**: ✅ 成功

#### user 数据库
```sql
-- 创建 5 张核心表
1. user_base                    -- 用户基础表
2. user_community_membership    -- 用户-小区成员关系表
3. user_membership_role         -- 角色表
4. user_certification           -- 认证记录表
5. user_residence               -- 房屋明细表
```

**执行结果**: ✅ 成功

### 3. 验证结果

#### auth 数据库
```bash
$ docker exec mysql mysql -uroot -proot123456 -e "USE auth; SHOW TABLES;"
Tables_in_auth
auth_credential
```

#### user 数据库
```bash
$ docker exec mysql mysql -uroot -proot123456 -e "USE user; SHOW TABLES;"
Tables_in_user
user_base
user_certification
user_community_membership
user_membership_role
user_residence
```

## 数据库状态总览

| 数据库 | 表数量 | 状态 | 说明 |
|--------|--------|------|------|
| auth | 1 | ✅ 就绪 | 登录凭证表已创建 |
| user | 5 | ✅ 就绪 | 用户服务核心表已创建 |

## 表结构概览

### auth.auth_credential
- **用途**: 存储用户登录凭证
- **安全特性**: 
  - 手机号 AES-256 加密
  - 密码 bcrypt 加盐哈希
- **记录数**: 0（全新库）

### user 数据库（5 张表）
| 表名 | 用途 | 引擎 | 记录数 |
|------|------|------|--------|
| user_base | 用户基础信息 | InnoDB | 0 |
| user_community_membership | 用户-小区关系 | InnoDB | 0 |
| user_membership_role | 角色信息 | InnoDB | 0 |
| user_certification | 认证记录 | InnoDB | 0 |
| user_residence | 房屋信息 | InnoDB | 0 |

## 后续操作建议

### 1. 测试数据准备
```bash
# 如需创建测试数据，可以使用以下脚本
# services/user-service/migration/seed_test_data.sql
```

### 2. 增量迁移（可选）
```bash
# 如果需要执行其他增量 migration
docker exec -i mysql mysql -uroot -proot123456 user < services/user-service/migration/003_add_address_fields.sql
docker exec -i mysql mysql -uroot -proot123456 user < services/user-service/migration/004_add_moderation_status.sql
```

### 3. 服务启动验证
```bash
# 启动 auth-service
cd services/auth-service/rpc
go run authservice.go -f etc/authservice.yaml

# 启动 user-service
cd services/user-service/rpc
go run userservice.go -f etc/userservice.yaml
```

### 4. 功能测试
```bash
# 测试用户注册
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "13812345678",
    "sms_code": "123456",
    "nickname": "测试用户",
    "device_id": "test_device_001"
  }'

# 测试用户登录
curl -X POST http://localhost:8080/api/auth/login-sms \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "13812345678",
    "sms_code": "123456",
    "device_id": "test_device_001"
  }'
```

## 文档更新

### 新增文档
- ✅ `services/auth-service/README.md` - auth-service 数据库说明
- ✅ `services/user-service/README.md` - user-service 数据库说明

### 新增 Migration 脚本
- ✅ `services/auth-service/migration/001_initial.sql`
- ✅ `services/user-service/migration/000_initial_schema.sql`

## 注意事项

1. **密码安全**: MySQL root 密码是 `root123456`，生产环境请修改
2. **数据加密**: 手机号和身份证号使用 AES-256 加密，密钥在 `.env` 文件中
3. **雪花ID**: user 数据库的所有表使用雪花算法生成主键，需要在代码中实现
4. **Redis 配置**: Token 和验证码存储在 Redis，确保 Redis 服务正常运行

## 快速命令参考

```bash
# 查看 auth 数据库
docker exec mysql mysql -uroot -proot123456 -e "USE auth; SHOW TABLES;"
docker exec mysql mysql -uroot -proot123456 -e "USE auth; DESCRIBE auth_credential;"

# 查看 user 数据库
docker exec mysql mysql -uroot -proot123456 -e "USE user; SHOW TABLES;"
docker exec mysql mysql -uroot -proot123456 -e "USE user; DESCRIBE user_base;"

# 查看表数据
docker exec mysql mysql -uroot -proot123456 -e "USE auth; SELECT COUNT(*) FROM auth_credential;"
docker exec mysql mysql -uroot -proot123456 -e "USE user; SELECT COUNT(*) FROM user_base;"
```

## 状态标记

- ✅ auth 数据库初始化完成
- ✅ user 数据库初始化完成
- ✅ 表结构验证通过
- ✅ 索引创建完成
- ⏸️ 测试数据待添加
- ⏸️ 服务启动测试待执行
