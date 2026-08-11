# Auth Service - 认证服务

## 数据库初始化

### 前置条件
- MySQL 8.0 容器运行中
- 数据库 `auth` 已创建

### 执行初始化脚本

```bash
# 初始化 auth 数据库表结构
docker exec -i mysql mysql -uroot -proot123456 auth < services/auth-service/migration/001_initial.sql
```

### 验证
```bash
# 查看表
docker exec mysql mysql -uroot -proot123456 -e "USE auth; SHOW TABLES;"

# 查看表结构
docker exec mysql mysql -uroot -proot123456 -e "USE auth; DESCRIBE auth_credential;"
```

## Migration 脚本列表

| 文件 | 说明 | 状态 |
|------|------|------|
| `001_initial.sql` | 创建 `auth_credential` 登录凭证表 | ✅ 已执行 |

## 表结构说明

### auth_credential — 登录凭证表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT | 主键（AUTO_INCREMENT） |
| user_id | BIGINT | 用户ID（FK → user_base.id） |
| identity_type | VARCHAR(20) | 登录方式：phone / sms / wechat |
| identifier | VARCHAR(255) | 标识符（AES 加密的手机号） |
| credential | VARCHAR(255) | 凭证（bcrypt 密文） |
| created_time | DATETIME | 创建时间 |
| updated_time | DATETIME | 更新时间 |

**索引**：
- PRIMARY KEY (id)
- UNIQUE INDEX uk_identity (identity_type, identifier)
- INDEX idx_user (user_id)

**安全特性**：
- identifier: AES-256 加密存储的手机号（不存明文）
- credential: bcrypt 加盐哈希的密码
- 支持多种登录方式，同一 identity_type 下 identifier 唯一
