# 用户注册与登录业务逻辑详解

> 完整的认证流程、数据库设计和服务交互说明  
> 创建时间: 2026-07-11  
> 基于代码分析

---

## 一、架构概览

### 服务拆分

```
┌─────────────────┐
│   前端/客户端    │
└────────┬────────┘
         │
         ↓
┌─────────────────┐
│ auth-service    │ ← 认证服务（API + RPC）
│  - API: 8881    │
│  - RPC: 8083    │
└────────┬────────┘
         │
         ├──→ Redis (验证码存储、角色缓存)
         ├──→ auth 数据库 (凭证表)
         └──→ user-service RPC (用户数据)
                  │
                  ↓
         ┌─────────────────┐
         │ user-service    │ ← 用户服务（RPC）
         │  - RPC: 8082    │
         └────────┬────────┘
                  │
                  └──→ user 数据库 (用户基础信息、角色)
```

---

## 二、数据库设计

### 2.1 数据库划分

**两个独立数据库**：

1. **`auth` 数据库** — auth-service 使用
   - 存储：认证凭证（账号密码）

2. **`user` 数据库** — user-service 使用
   - 存储：用户基础信息、社区关系、角色等

### 2.2 auth 数据库表结构

#### `auth_credential` 表（认证凭证表）

```sql
CREATE TABLE auth_credential (
    id              BIGINT NOT NULL AUTO_INCREMENT,
    user_id         BIGINT NOT NULL COMMENT '用户ID（关联 user.user_base.id）',
    identity_type   VARCHAR(20) NOT NULL COMMENT '认证类型：phone/email/wechat',
    identifier      VARCHAR(255) NOT NULL COMMENT '身份标识（AES加密的手机号/邮箱）',
    credential      VARCHAR(255) NOT NULL COMMENT '凭证（bcrypt 哈希后的密码）',
    created_time    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_time    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    PRIMARY KEY (id),
    UNIQUE INDEX uk_identity (identity_type, identifier),
    INDEX idx_user (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='认证凭证表';
```

**核心字段说明**：
- `identifier`: **AES 加密**的手机号（使用 .env 中统一的 AES_KEY）
- `credential`: **bcrypt 哈希**后的密码
- `identity_type`: 目前主要是 "phone"

---

### 2.3 user 数据库表结构

#### `user_base` 表（用户基础信息）

```sql
CREATE TABLE user_base (
    id              BIGINT NOT NULL COMMENT 'Snowflake ID',
    phone           VARCHAR(255) NULL COMMENT '手机号（AES加密）',
    nickname        VARCHAR(50) NULL COMMENT '昵称',
    avatar_url      VARCHAR(255) NULL COMMENT '头像URL',
    real_name       VARCHAR(50) NULL COMMENT '真实姓名',
    id_card_number  VARCHAR(255) NULL COMMENT '身份证号（AES加密）',
    gender          TINYINT NULL COMMENT '性别：1-男 2-女',
    birth_date      DATE NULL COMMENT '出生日期',
    credit_score    INT DEFAULT 100 COMMENT '信用分',
    preferences     JSON NULL COMMENT '用户偏好',
    status          TINYINT DEFAULT 1 COMMENT '状态：1-正常 2-禁用',
    created_time    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_time    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    delete_time     DATETIME NULL COMMENT '删除时间',
    
    PRIMARY KEY (id),
    INDEX idx_phone (phone),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户基础信息表';
```

#### `user_community_membership` 表（用户-小区关系）

```sql
CREATE TABLE user_community_membership (
    id              BIGINT NOT NULL AUTO_INCREMENT,
    user_id         BIGINT NOT NULL COMMENT '用户ID',
    community_id    BIGINT NOT NULL COMMENT '小区ID',
    bind_status     TINYINT NOT NULL DEFAULT 1 COMMENT '1-有效 0-已退出',
    join_time       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    leave_time      DATETIME NULL,
    created_time    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_time    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    PRIMARY KEY (id),
    UNIQUE INDEX uk_user_community (user_id, community_id),
    INDEX idx_community (community_id, bind_status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户-小区成员关系';
```

#### `user_membership_role` 表（用户角色）

```sql
CREATE TABLE user_membership_role (
    id              BIGINT NOT NULL AUTO_INCREMENT,
    user_id         BIGINT NOT NULL COMMENT '用户ID',
    membership_id   BIGINT NULL COMMENT '小区成员关系ID，商家为 NULL',
    community_id    BIGINT NOT NULL DEFAULT 0 COMMENT '小区ID，0=全局角色(商家)',
    role_code       VARCHAR(30) NOT NULL COMMENT '角色编码：owner/property_admin/merchant等',
    verf_status     TINYINT NOT NULL DEFAULT 0 COMMENT '认证状态：0-未认证 1-待审 2-已通过 3-已驳回 4-已过期',
    verified_at     DATETIME NULL COMMENT '认证通过时间',
    expires_at      DATETIME NULL COMMENT '过期时间，NULL=永久有效',
    created_time    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_time    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    PRIMARY KEY (id),
    UNIQUE INDEX uk_member_role (membership_id, role_code),
    UNIQUE INDEX uk_user_community_role (user_id, community_id, role_code),
    INDEX idx_community_role (community_id, role_code, verf_status),
    INDEX idx_expires (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户角色表';
```

---

## 三、用户注册流程

### 3.1 注册流程图

```
用户输入手机号 + 验证码 + 密码
         ↓
auth-service API (POST /api/auth/register)
         ↓
auth-service RPC (RegisterLogic)
         ↓
┌────────────────────────────────────────────┐
│ 1. 校验短信验证码（Redis）                  │
│    key: sms:code:{phone}                   │
│    验证码：123456（固定）                   │
└────────────────────────────────────────────┘
         ↓
┌────────────────────────────────────────────┐
│ 2. 调用 user-service.CreateUser (RPC)     │
│    → user 数据库插入 user_base             │
│    → 返回 user_id (Snowflake ID)          │
└────────────────────────────────────────────┘
         ↓
┌────────────────────────────────────────────┐
│ 3. RSA 解密前端加密的密码                   │
│    → bcrypt 哈希密码                       │
│    → AES 加密手机号                        │
└────────────────────────────────────────────┘
         ↓
┌────────────────────────────────────────────┐
│ 4. 写入 auth_credential 表                 │
│    user_id: {user_id}                      │
│    identity_type: "phone"                  │
│    identifier: AES(phone)                  │
│    credential: bcrypt(password)            │
└────────────────────────────────────────────┘
         ↓
┌────────────────────────────────────────────┐
│ 5. 如果写入失败 → Saga 补偿                │
│    调用 user-service 删除刚创建的用户       │
└────────────────────────────────────────────┘
         ↓
┌────────────────────────────────────────────┐
│ 6. 获取用户角色（新用户通常为空 roles=[]） │
│    Cache-Aside 模式：Redis → RPC           │
└────────────────────────────────────────────┘
         ↓
┌────────────────────────────────────────────┐
│ 7. 签发 JWT Token                          │
│    - Access Token (AT): 15分钟有效         │
│      包含：user_id, roles, jti, exp        │
│    - Refresh Token (RT): 15天有效          │
│      包含：user_id, jti, exp               │
└────────────────────────────────────────────┘
         ↓
返回给前端: {
  access_token: "...",
  refresh_token: "...",
  user_id: "...",
  expires_in: 900
}
```

### 3.2 注册核心代码逻辑

**文件**: `services/auth-service/rpc/internal/logic/auth/registerlogic.go`

```go
func (l *RegisterLogic) Register(in *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
    // 1. 校验短信验证码
    codeKey := fmt.Sprintf("sms:code:%s", in.Phone)
    storedCode, err := l.svcCtx.RedisClient.Get(l.ctx, codeKey).Result()
    if err == redis.Nil {
        return &authv1.RegisterResponse{
            Base: responsex.NewBaseRespWithError(50002, "验证码已过期"),
        }, nil
    }
    if storedCode != in.SmsCode {
        return &authv1.RegisterResponse{
            Base: responsex.NewBaseRespWithError(50003, "验证码错误"),
        }, nil
    }

    // 2. 调用 User Service 创建用户
    createUserResp, err := l.svcCtx.UserServiceRpc.CreateUser(l.ctx, &userv1.CreateUserRequest{
        Phone:    in.Phone,
        Nickname: in.Nickname,
        UserType: 1, // 默认居民
    })
    userId := createUserResp.UserId

    // 3. RSA 解密密码 → bcrypt 哈希
    plainPassword, err := crypto.DecryptRSA(in.EncryptedPassword)
    bcryptHash, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)

    // AES 加密手机号
    encryptedPhone, err := crypto.EncryptAES(in.Phone)

    // 4. 写入 auth_credential
    _, err = l.svcCtx.AuthCredentialModel.Insert(l.ctx, &model.AuthCredential{
        UserId:       userId,
        IdentityType: "phone",
        Identifier:   encryptedPhone,  // AES 加密
        Credential:   bcryptHash,       // bcrypt 哈希
    })
    
    if err != nil {
        // Saga 补偿：删除用户
        l.compensateUser(userId)
        return error
    }

    // 5. 获取角色（新用户 roles=[]）
    roles, _ := getUserRolesWithCache(l.ctx, l.svcCtx, userId)

    // 6. 签发 JWT Token
    at, rt := l.generateTokens(userId, roles)
    
    return &authv1.RegisterResponse{
        AccessToken:  at,
        RefreshToken: rt,
        UserId:       userId,
    }, nil
}
```

---

## 四、用户登录流程

### 4.1 短信登录流程图

```
用户输入手机号 + 验证码
         ↓
auth-service API (POST /api/auth/login/sms)
         ↓
auth-service RPC (LoginSmsLogic)
         ↓
┌────────────────────────────────────────────┐
│ 1. 校验短信验证码（Redis）                  │
│    key: sms:code:{phone}                   │
└────────────────────────────────────────────┘
         ↓
┌────────────────────────────────────────────┐
│ 2. AES 加密手机号                          │
│    encryptedPhone = AES(phone)             │
└────────────────────────────────────────────┘
         ↓
┌────────────────────────────────────────────┐
│ 3. 查询 auth_credential 表                 │
│    WHERE identity_type='phone'             │
│      AND identifier=encryptedPhone         │
│    → 获取 user_id                          │
└────────────────────────────────────────────┘
         ↓ 未找到
┌────────────────────────────────────────────┐
│ 返回错误："该手机号未注册，请先注册"        │
│ 错误码: 50001                              │
└────────────────────────────────────────────┘
         ↓ 找到
┌────────────────────────────────────────────┐
│ 4. 调用 user-service.GetUserByPhone (RPC) │
│    → 获取用户信息                          │
│    → 检查 user.status                      │
└────────────────────────────────────────────┘
         ↓ status == 2
┌────────────────────────────────────────────┐
│ 返回错误："账号已被禁用"                    │
│ 错误码: 50005                              │
└────────────────────────────────────────────┘
         ↓ status == 1
┌────────────────────────────────────────────┐
│ 5. 获取用户角色（Cache-Aside）             │
│    Redis key: user:roles:{user_id}         │
│    Cache Miss → 调用 permission-service    │
│    → 写回 Redis (TTL: 3600s)              │
└────────────────────────────────────────────┘
         ↓
┌────────────────────────────────────────────┐
│ 6. 签发 JWT Token                          │
│    - Access Token: 包含 user_id + roles    │
│    - Refresh Token: 仅包含 user_id         │
└────────────────────────────────────────────┘
         ↓
返回给前端: {
  access_token: "...",
  refresh_token: "...",
  user_id: "..."
}
```

### 4.2 登录核心代码逻辑

**文件**: `services/auth-service/rpc/internal/logic/auth/loginsmslogic.go`

```go
func (l *LoginSmsLogic) LoginSms(in *authv1.LoginRequest) (*authv1.LoginResponse, error) {
    // 1. 校验短信验证码
    codeKey := fmt.Sprintf("sms:code:%s", in.Phone)
    storedCode, err := l.svcCtx.RedisClient.Get(l.ctx, codeKey).Result()
    if storedCode != in.SmsCode {
        return error("验证码错误")
    }

    // 2. AES 加密手机号
    encryptedPhone, err := crypto.EncryptAES(in.Phone)

    // 3. 查询 auth_credential
    credential, err := l.svcCtx.AuthCredentialModel.FindOneByIdentity(
        l.ctx, "phone", encryptedPhone,
    )
    if err == ErrNotFound {
        // 未注册
        return &authv1.LoginResponse{
            Base: responsex.NewBaseRespWithError(50001, "该手机号未注册，请先注册"),
        }, nil
    }

    // 4. 获取用户信息（检查状态）
    userResp, err := l.svcCtx.UserServiceRpc.GetUserByPhone(l.ctx, 
        &userv1.GetUserByPhoneRequest{Phone: in.Phone})
    
    if userResp.User.Status == 2 {
        return &authv1.LoginResponse{
            Base: responsex.NewBaseRespWithError(50005, "账号已被禁用"),
        }, nil
    }

    userId := credential.UserId

    // 5. 获取用户角色（Cache-Aside）
    roles, err := getUserRolesWithCache(l.ctx, l.svcCtx, userId)

    // 6. 签发 JWT Token
    at, rt := l.generateTokens(userId, roles)
    
    return &authv1.LoginResponse{
        AccessToken:  at,
        RefreshToken: rt,
        UserId:       userId,
    }, nil
}
```

---

## 五、关键技术点

### 5.1 敏感数据加密

**手机号加密**：
- **算法**: AES-256
- **密钥**: 统一使用 `.env` 中的 `AES_KEY`（32字节）
- **存储位置**: 
  - `auth.auth_credential.identifier` (AES加密)
  - `user.user_base.phone` (AES加密)

**密码加密**：
- **传输**: RSA 加密（前端使用公钥加密，后端私钥解密）
- **存储**: bcrypt 哈希（`auth.auth_credential.credential`）
- **不可逆**: 登录时验证，无法还原原始密码

### 5.2 验证码机制

**存储**: Redis  
**Key**: `sms:code:{phone}`  
**Value**: 验证码（当前固定为 `123456` 便于测试）  
**TTL**: 300 秒（5分钟）

**限流**: Redis  
**Key**: `sms:rate:{phone}`  
**TTL**: 60 秒  
**作用**: 同一手机号 60 秒内只能发送一次

### 5.3 角色缓存机制（Cache-Aside）

```
获取用户角色
    ↓
查 Redis: user:roles:{user_id}
    ↓ Cache Hit
返回角色列表
    ↓ Cache Miss
调用 permission-service RPC
    ↓
写回 Redis (TTL: 3600s)
    ↓
返回角色列表
```

### 5.4 JWT Token 设计

**Access Token (AT)**:
- 有效期: 15 分钟 (900秒)
- 包含: `user_id`, `roles`, `jti`, `exp`, `iat`
- 签名算法: HS256
- 密钥: `.env` 中的 `JWT_ACCESS_SECRET`

**Refresh Token (RT)**:
- 有效期: 15 天 (1296000秒)
- 包含: `user_id`, `jti`, `exp`, `iat`
- 签名算法: HS256
- 密钥: `.env` 中的 `JWT_REFRESH_SECRET`

---

## 六、服务交互图

### 6.1 注册时的服务调用

```
前端
  ↓ POST /api/auth/register
auth-service API (8881)
  ↓ gRPC: Register
auth-service RPC (8083)
  ├─→ Redis: 验证码校验
  ├─→ user-service RPC (8082): CreateUser
  │    ↓
  │   user 数据库: INSERT user_base
  │    ↓
  │   返回 user_id
  ├─→ auth 数据库: INSERT auth_credential
  └─→ Redis: 缓存角色（可选）
```

### 6.2 登录时的服务调用

```
前端
  ↓ POST /api/auth/login/sms
auth-service API (8881)
  ↓ gRPC: LoginSms
auth-service RPC (8083)
  ├─→ Redis: 验证码校验
  ├─→ auth 数据库: SELECT auth_credential
  ├─→ user-service RPC (8082): GetUserByPhone
  │    ↓
  │   user 数据库: SELECT user_base
  │    ↓
  │   返回用户信息
  ├─→ Redis: 查询角色缓存
  │    ↓ Cache Miss
  ├─→ permission-service RPC (8084): GetUserRoles
  │    ↓
  │   permission 数据库: SELECT user_membership_role
  │    ↓
  │   返回角色列表
  └─→ Redis: 写回角色缓存
```

---

## 七、错误码说明

| 错误码 | 说明 | 场景 |
|-------|------|------|
| 50001 | 该手机号未注册，请先注册 | 登录时未找到凭证 |
| 50002 | 验证码已过期 | Redis 中无验证码 |
| 50003 | 验证码错误 | 验证码不匹配 |
| 50004 | 该手机号已注册 | 注册时手机号重复 |
| 50005 | 账号已被禁用 | user.status == 2 |
| 509001 | 注册失败，请稍后重试 | 创建用户/凭证失败 |
| 509503 | 获取用户信息失败 | user-service RPC 失败 |
| 509504 | 获取用户角色失败 | permission-service RPC 失败 |

---

## 八、配置文件

### .env（统一环境变量）

```bash
# AES 加密密钥（32 字节 = AES-256）
AES_KEY=49e163155b0e2c2569af909d7d64117e

# JWT 密钥
JWT_ACCESS_SECRET=QBViQKNdUpsAdq48ClBoFxRoRwayo7MZFrLr_eu5NL8
JWT_REFRESH_SECRET=-hSf5SGq_7AQaE6GDXhALU53LhK7Er2gQMDTS9RMVL4

# MySQL 配置
MYSQL_USER=root
MYSQL_PASSWORD=root123456

# Redis 配置
REDIS_PASSWORD=123456
```

---

## 九、总结

### 核心设计原则

1. ✅ **服务职责清晰**
   - auth-service: 负责认证（登录、注册、Token）
   - user-service: 负责用户数据（基础信息、角色）
   - permission-service: 负责权限管理

2. ✅ **数据库隔离**
   - `auth` 数据库: 认证凭证
   - `user` 数据库: 用户数据

3. ✅ **安全性保障**
   - 手机号: AES 加密存储
   - 密码: bcrypt 哈希，不可逆
   - 传输: RSA 加密

4. ✅ **性能优化**
   - Redis 缓存角色（TTL: 1小时）
   - Cache-Aside 模式

5. ✅ **一致性保障**
   - Saga 补偿机制（注册失败时删除用户）

### 关键特性

- ⚠️ **短信登录不会自动注册** — 必须先注册
- ✅ **固定验证码 123456** — 方便测试
- ✅ **统一密钥管理** — 所有服务共用 .env 配置
- ✅ **完整的 Audit Trail** — 创建/更新时间记录

---

**创建时间**: 2026-07-11  
**维护者**: Kiro Development Team  
**相关文档**: 
- `services/auth-service/docs/design.md`
- `services/user-service/docs/design.md`
