# User Service - 用户服务

## 数据库初始化

### 前置条件
- MySQL 8.0 容器运行中
- 数据库 `user` 已创建

### 执行初始化脚本

```bash
# 初始化 user 数据库表结构（全新安装）
docker exec -i mysql mysql -uroot -proot123456 user < services/user-service/migration/000_initial_schema.sql
```

### 验证
```bash
# 查看所有表
docker exec mysql mysql -uroot -proot123456 -e "USE user; SHOW TABLES;"

# 查看表结构
docker exec mysql mysql -uroot -proot123456 -e "USE user; DESCRIBE user_base;"
```

## Migration 脚本列表

| 文件 | 说明 | 状态 |
|------|------|------|
| `000_initial_schema.sql` | 创建 5 张核心表（全新安装） | ✅ 已执行 |
| `001_refactor_to_v2.sql` | 旧表迁移脚本（仅用于从旧版本升级） | ⏸️ 跳过 |
| `002_snowflake_ids.sql` | 雪花ID迁移（仅用于从旧版本升级） | ⏸️ 跳过 |
| `003_add_address_fields.sql` | 增量字段（可选） | ⏸️ 待执行 |
| `004_add_moderation_status.sql` | 审核状态字段（可选） | ⏸️ 待执行 |

## 数据库表结构

### 1. user_base — 用户基础表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT | 用户ID（雪花算法生成） |
| phone | VARCHAR(255) | 手机号（AES加密存储） |
| nickname | VARCHAR(100) | 昵称 |
| avatar_url | VARCHAR(500) | 头像URL |
| real_name | VARCHAR(50) | 真实姓名（认证后回填） |
| id_card_number | VARCHAR(255) | 身份证号（AES加密） |
| gender | TINYINT | 性别：1-男 2-女 |
| birth_date | DATE | 出生日期 |
| status | TINYINT | 状态：1-正常 2-禁用 |
| credit_score | INT | 信用分（默认100） |
| preferences | JSON | 用户偏好 |
| created_time | DATETIME | 创建时间 |
| updated_time | DATETIME | 更新时间 |
| delete_time | DATETIME | 软删除时间 |

**索引**：
- PRIMARY KEY (id)
- UNIQUE INDEX idx_phone (phone)
- INDEX idx_status (status)

### 2. user_community_membership — 用户-小区成员关系表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT | 雪花算法生成 |
| user_id | BIGINT | 用户ID |
| community_id | BIGINT | 小区ID |
| bind_status | TINYINT | 1-有效 0-已退出 |
| join_time | DATETIME | 加入时间 |
| leave_time | DATETIME | 退出时间 |
| created_time | DATETIME | 创建时间 |
| updated_time | DATETIME | 更新时间 |

**索引**：
- PRIMARY KEY (id)
- UNIQUE INDEX uk_user_community (user_id, community_id)
- INDEX idx_community (community_id, bind_status)

### 3. user_membership_role — 角色表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT | 雪花算法生成 |
| user_id | BIGINT | 用户ID（冗余，分片键） |
| membership_id | BIGINT | 小区成员关系ID（商家为NULL） |
| community_id | BIGINT | 小区ID（0=全局角色） |
| role_code | VARCHAR(30) | 角色编码 |
| verf_status | TINYINT | 认证状态：0-未认证 1-待审 2-已通过 3-已驳回 4-已过期 |
| verified_at | DATETIME | 认证通过时间 |
| expires_at | DATETIME | 过期时间（NULL=永久有效） |
| created_time | DATETIME | 创建时间 |
| updated_time | DATETIME | 更新时间 |

**角色类型**：
- owner: 业主
- tenant: 租户
- grid_worker: 网格员
- community_admin: 社区管理员
- property_admin: 物业管理员
- committee: 业委会成员
- merchant: 商家

**索引**：
- PRIMARY KEY (id)
- UNIQUE INDEX uk_member_role (membership_id, role_code)
- UNIQUE INDEX uk_user_community_role (user_id, community_id, role_code)
- INDEX idx_community_role (community_id, role_code, verf_status)
- INDEX idx_expires (expires_at)

### 4. user_certification — 认证记录表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT | 雪花算法生成 |
| role_id | BIGINT | 角色ID |
| user_id | BIGINT | 用户ID（冗余） |
| document_urls | TEXT | 证明材料URL（JSON数组） |
| status | TINYINT | 审核状态：1-待审核 2-已通过 3-已驳回 |
| reviewer_id | BIGINT | 审核人ID |
| review_time | DATETIME | 审核时间 |
| review_notes | VARCHAR(500) | 审核备注 |
| submit_time | DATETIME | 提交时间 |

**索引**：
- PRIMARY KEY (id)
- INDEX idx_role (role_id)
- INDEX idx_user (user_id)
- INDEX idx_status (status)

### 5. user_residence — 房屋明细表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT | 雪花算法生成 |
| membership_id | BIGINT | 小区成员关系ID |
| user_id | BIGINT | 用户ID（冗余，分片键） |
| house_id | VARCHAR(50) | 房屋ID（如 1-2-301） |
| building | VARCHAR(20) | 楼号 |
| unit | VARCHAR(20) | 单元号 |
| room | VARCHAR(20) | 房号 |
| is_primary | TINYINT | 主房产：1-是 0-否 |
| start_date | DATE | 入住/合同开始日期 |
| end_date | DATE | 搬离/合同结束日期 |
| created_time | DATETIME | 创建时间 |
| updated_time | DATETIME | 更新时间 |

**索引**：
- PRIMARY KEY (id)
- UNIQUE INDEX uk_member_house (membership_id, house_id)
- INDEX idx_user_id (user_id)

## 设计特性

- **雪花ID**：所有表主键使用雪花算法生成，非 AUTO_INCREMENT，支持分库分表
- **加密存储**：手机号、身份证号使用 AES-256 加密
- **分片策略**：按 user_id 分片，同一用户数据落在同一分片
- **角色认证**：统一认证流程，支持 7 种角色类型
- **软删除**：user_base 使用 delete_time 实现软删除
