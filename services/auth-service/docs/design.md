# 「我的小区我的家」— 认证中心设计方案

## 一、定位

`auth-service` 是社区平台的**唯一登录与令牌管理中台**，负责身份认证、Token 签发/验证/注销全生命周期管理。

### 1.1 做什么

| 职责 | 说明 |
|------|------|
| 用户注册 | 短信验证码校验 → 调用 user-service 创建用户 → 写入凭证 → 签发 Token |
| 账密登录 | RSA 解密手机号+密码 → bcrypt 校验 → 签发 Token |
| 短信验证码登录 | 验证码校验 → 查凭证 → 签发 Token |
| Token 签发 | AT（15min）+ RT（15day）双 Token，JWT 携带用户角色信息 |
| Token 刷新 | RT 轮换（防泄露），刷新时重新拉取角色信息 |
| Token 验证 | 供 API Gateway 校验 AT 有效性（签名+过期+黑名单） |
| 主动注销 | AT 加入黑名单，清除 RT，可选全设备强踢 |
| 短信验证码 | 发送、存储（Redis TTL）、校验 |
| 角色缓存 | Redis 缓存用户已认证角色（5 分钟 TTL），减少 gRPC 穿透 | 
| RSA 公钥 | 对外提供 RSA 公钥，客户端加密手机号/密码后传输 |

### 1.2 不做什么

| 不负责 | 归属 |
|--------|------|
| 用户身份管理（基本信息、小区归属、角色认证） | user-service |
| 权限判定（角色能做什么操作） | API Gateway 配置文件 + permission-service |
| 小区信息管理 | community-service |
| 文件上传/存储 | file-service |
| 登录日志（时间/IP/设备） | auth-service `auth_login_log`（未来） |

### 1.3 核心设计原则

- **认权携带，判定分离**：Auth 签发 JWT 时携带用户已认证的角色列表（携带），但不判定角色能做什么操作（判定在 API Gateway / permission-service）
- **成员身份进 JWT，资源范围走业务服务**：JWT 中的 `c`（community_id）是**成员身份范围**——"我在小区 X 拥有角色 Y"。广告投放范围（"我的广告在哪些小区可见"）是资源可见性规则，由对应业务服务（ad-service）在查询时动态判定，不进 JWT
- **AT 短、RT 长**：AT 15 分钟过期，RT 15 天过期，兼顾安全性与用户体验
- **RT 轮换防泄露**：刷新 AT 时必须同时轮换 RT，旧 RT 立即失效
- **AT 即时撤销**：注销时 AT 加入 Redis 黑名单，TTL 对齐 AT 过期时间
- **角色缓存优先，穿透兜底**：从 Redis 读取角色缓存（TTL 5 分钟），未命中时穿透 gRPC 拉取并回填。user-service 在角色变更时主动 `DEL` 缓存 key，保证最多 5 分钟生效；敏感操作走 CheckAccess 实时校验兜底

---

## 二、JWT Token 设计

### 2.1 AT（Access Token）

**签名算法**：HMAC-SHA256  
**有效期**：15 分钟  
**用途**：客户端携带在 `Authorization: Bearer <AT>` 中访问业务接口

```json
{
  "user_id": 1234567890,
  "jti": "1234567890-1717430400123456789",
  "roles": [
    {"r": "owner", "c": 1001},
    {"r": "committee", "c": 1001}
  ],
  "exp": 1717431300,
  "iat": 1717430400
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `user_id` | int64 | 用户 ID（雪花算法生成） |
| `jti` | string | Token 唯一 ID，格式 `{user_id}-{unix_nano}`，用于黑名单精确匹配 |
| `roles` | array | 用户已认证的角色列表（仅 `verf_status=2` 的角色） |
| `roles[].r` | string | `role_code`：owner / tenant / grid_worker / community_admin / property_admin / committee / merchant |
| `roles[].c` | int64 | `community_id`：角色所属小区。merchant 为 0（全局角色） |
| `exp` | int64 | 过期时间（Unix 秒） |
| `iat` | int64 | 签发时间（Unix 秒） |

**字段命名精简说明**：`r`/`c` 而非 `role_code`/`community_id`，节省 JWT 体积。典型用户 1-2 个角色，额外开销约 50-100 字节，远低于 HTTP Header 8KB 限制。

### 2.3 JWT 体积与上限

| 约束 | 值 | 说明 |
|------|-----|------|
| HTTP Header 上限（Nginx 默认） | 4KB 单 header / 8KB 总 | 最常见瓶颈 |
| 实践建议 | **≤ 2KB**（含签名+Header） | 留余量给其他 headers |
| 每个 role 条目 `{"r":"owner","c":1001}` | ~28 字节 | 2KB 约可容纳 70 个条目 |
| JWT roles 条目上限 | **50 个** | 超限时省略 `c`，Gateway 降级为 CheckAccess |

**`c` 字段语义**：`c` 是**成员身份范围**（"我在小区 X 拥有角色 Y"），不是资源可见范围。`merchant` 角色 `c=0`（全局角色），广告投放范围由 ad-service 的 `ad_merchant_scope` 表独立管理，不进 JWT。

**极端情况兜底**：若角色条目超过 50 个上限（如网格员跨 100+ 小区），该角色条目的 `c` 省略，Gateway 检测到缺失 `c` 时自动降级为 CheckAccess RPC 实时校验。

### 2.2 RT（Refresh Token）

**签名算法**：HMAC-SHA256（独立密钥）  
**有效期**：15 天  
**用途**：AT 过期后换取新的 AT+RT 对，不直接用于业务鉴权

```json
{
  "user_id": 1234567890,
  "device_id": "web_chrome_abc123",
  "jti": "1234567890-1717430400123456789",
  "exp": 1718726400,
  "iat": 1717430400
}
```

| 字段 | 说明 |
|------|------|
| `user_id` | 用户 ID |
| `device_id` | 设备标识（多端登录管理） |
| `jti` | Token 唯一 ID，Redis 持久化键值的校验凭据 |
| `exp` | 过期时间（15 天） |
| `iat` | 签发时间 |

**RT 不携带 roles**：RT 仅用于换取新 Token，不暴露给业务接口。角色信息仅在签发新 AT 时实时拉取。

---

## 三、认证流程

### 3.1 账密登录

```
客户端 → auth-service.Login(encrypted_phone, encrypted_password, device_id)

1. RSA 私钥解密 phone + password
2. AES 加密 phone → 作为 identifier 查询 auth_credential
3. bcrypt 校验 password
4. 调用 user-service.GetUserByPhone(phone) → 获取 user_id、status
5. 校验 status != 2（禁用）
6. 获取已认证角色（Cache-Aside）：
   a. Redis GET auth:roles:{user_id} → 命中，直接返回
   b. 未命中 → gRPC GetUserRoles(user_id, verf_status=2) → Redis SET auth:roles:{user_id} EX 300
7. 构建 JWT claims（含 roles）→ 签发 AT + RT
8. Redis 持久化 RT → 加入设备集合
9. 返回 AT, RT, expires_at, user_id
```

### 3.2 短信验证码登录

```
客户端 → auth-service.LoginSms(phone, sms_code, device_id)

1. 从 Redis 校验短信验证码（校验通过立即删除，防重放）
2. AES 加密 phone → 查询 auth_credential
3. 调用 user-service.GetUserByPhone(phone)
4. 获取已认证角色（优先 Redis 缓存，未命中穿透 gRPC）
5. 签发 AT（含 roles）+ RT
6. Redis 持久化 RT
```

### 3.3 注册

```
客户端 → auth-service.Register(phone, sms_code, encrypted_password, nickname, device_id)

1. 校验短信验证码
2. 调用 user-service.CreateUser(phone, nickname) → 获取 user_id
3. RSA 解密 password → bcrypt 哈希 → 写入 auth_credential
   ⚠️ credential 写入失败 → Saga 补偿：调用 user-service.UpdateUser(status=3) 软删除
4. 获取已认证角色（优先 Redis 缓存，未命中穿透 gRPC。新用户 roles=[]）
5. 签发 AT（含 roles）+ RT
6. Redis 持久化 RT
```

### 3.4 Token 刷新

```
客户端 → auth-service.RefreshToken(refresh_token)

1. 解析 RT → 提取 user_id, device_id, jti
2. Redis GET auth:rt:{user_id}:{device_id}
   → 不存在或值 != jti → 403（已注销或已被旋转）
3. Lua 原子旋转：删除旧 RT → 写入新 RT（重置 TTL）
4. 重新拉取角色（优先 Redis 缓存，未命中穿透 gRPC）
5. 生成新 AT（含最新 roles）+ 新 RT
6. 返回新 Token 对

关键设计：刷新时重新拉取角色，保证角色变更在最多 5 分钟（缓存 TTL）内生效。
```

### 3.5 注销

```
客户端 → auth-service.Logout(access_token, user_id, device_id, kick_all_devices?)

1. 解析 AT → 提取 jti, exp
2. AT 加入黑名单：Redis SET auth:at:blacklist:{jti} = 1, TTL = exp - now
3. Redis DEL auth:rt:{user_id}:{device_id}（删除当前设备 RT）
4. IF kick_all_devices:
     SMEMBERS auth:rt:{user_id}:devices → 批量 DEL 所有设备 RT
```

### 3.6 Token 验证（API Gateway 调用）

```
API Gateway → auth-service.ValidateToken(access_token)

1. 解析 JWT → 验证签名 + 过期时间
2. 提取 jti → 检查 Redis 黑名单 auth:at:blacklist:{jti}
3. 返回 valid, user_id, expires_at

API Gateway 自行从 JWT 解析 roles 进行常规鉴权（见 §4.1）。
```

---

## 四、鉴权路径

### 4.1 路径一：JWT 角色解析（常规操作）

```
请求到达 API Gateway
  ├── 调用 ValidateToken 验证 AT 有效性（签名+过期+黑名单）
  ├── 从 JWT payload 解析 roles: [{"r":"owner","c":1001}]
  ├── 匹配路由配置: POST /api/communities/:cid/posts → allow_roles=[owner,tenant]
  │   → 检查 roles 中是否有 role_code IN allow_roles 且 community_id == :cid
  ├── 放行 / 403
  └── 转发到业务服务（注入 user_id, current_role, community_id 到 Metadata）
```

**适用**：所有常规 CRUD 操作（发帖、聊天、查信息等），要求低延迟。

### 4.2 路径二：CheckAccess 实时校验（敏感操作）

```
API Gateway → user-service.CheckAccess(user_id, role_codes, community_id)
  → 实时查询 DB，确保角色未被撤销（即使 AT 未过期）
  → allowed=true/false
```

**适用**：审核认证、发通知、管理操作等敏感操作。弥补 JWT 角色快照最多 15 分钟延迟的不足。

### 4.3 路径对比

| 维度 | JWT 角色解析 | CheckAccess RPC |
|------|:---:|:---:|
| 延迟 | ~0ms（本地解析） | ~5ms（gRPC + DB） |
| 角色实时性 | AT 签发时快照（≤15min） | 实时 |
| 适用场景 | 常规业务操作 | 敏感操作 |
| 依赖 | 无（自包含） | user-service |

### 4.4 路径三：资源范围查询（广告等场景）

**JWT 不携带资源可见范围**。当业务需要判定"某资源是否对当前小区可见"时，由业务服务在查询时自行判定：

```
请求 GET /api/communities/1001/ads
  → Gateway: JWT 解析 roles → 用户是 1001 的 owner → 允许访问本小区
  → ad-service: 
      SELECT ad.* FROM ad
      JOIN ad_merchant_scope s ON ad.merchant_id = s.merchant_id
      WHERE (s.scope_type='community' AND s.scope_id=1001)
         OR (s.scope_type='county' AND s.scope_id = (SELECT county_id FROM community WHERE id=1001))
         OR (s.scope_type='city'   AND s.scope_id = (SELECT city_id FROM community WHERE id=1001))
```

**适用**：广告展示、内容推荐等需要按区域/小区过滤资源的场景。资源范围变化即时生效，无需等 Token 过期。

### 4.5 路径对比总览

| 路径 | 机制 | 延迟 | 实时性 | 适用场景 |
|------|------|:---:|:---:|------|
| JWT 角色解析 | 本地解析 roles | ~0ms | ≤15min | 常规 CRUD（发帖、聊天） |
| CheckAccess RPC | gRPC 实时查 DB | ~5ms | 实时 | 敏感操作（审核、管理） |
| 资源范围查询 | 业务服务自行 JOIN | 查询内联 | 实时 | 广告展示、内容推荐 |

---

## 五、角色缓存与一致性

### 5.1 缓存策略

```
读取（Cache-Aside）：
  auth-service: Redis GET auth:roles:{user_id}
    → HIT  → 直接返回（< 1ms）
    → MISS → gRPC GetUserRoles(user_id, verf_status=2)
           → Redis SET auth:roles:{user_id} = [{"r":"owner","c":1001}], EX 300
           → 返回

失效（user-service 主动删除）：
  user-service 角色状态变更 → Redis DEL auth:roles:{user_id}
  → 下次读取 MISS → 穿透拉取最新角色
```

### 5.2 缓存失效时机

| 事件 | 触发方 | 动作 | 生效延迟 |
|------|--------|------|:---:|
| 审核通过角色 | user-service ReviewCertification | `DEL auth:roles:{user_id}` | 下次刷新 |
| 审核驳回角色 | user-service ReviewCertification | `DEL auth:roles:{user_id}` | 下次刷新 |
| 管理员撤销角色 | user-service | `DEL auth:roles:{user_id}` | 下次刷新 |
| 角色过期（定时任务） | user-service Cron | `DEL auth:roles:{user_id}` | 下次刷新 |
| TTL 自然过期 | Redis | 自动删除 | ≤5 分钟 |
| DEL 操作失败 | — | TTL 兜底 | ≤5 分钟 |

### 5.3 角色变更传播时间

```
管理员审核通过角色
  → user-service 更新 DB + DEL auth:roles:{user_id}
  → 用户下次刷新 Token → Redis MISS → 穿透 gRPC → 新角色写入 JWT
  → 最长延迟：5 分钟（缓存 TTL）+ 15 分钟（旧 AT 过期）
  → 敏感操作：CheckAccess 实时查 DB，立即感知变更
```

### 5.4 安全兜底

| 场景 | 兜底机制 |
|------|------|
| user-service DEL 失败 | TTL 5 分钟后自动过期，下次读取穿透 |
| 缓存未命中 | 穿透到 gRPC（逻辑与无缓存一致） |
| 敏感操作 | CheckAccess RPC 实时查 DB，不经过缓存 |
| 紧急撤销 | 管理员调用 auth-service 强制下线（清除所有 RT） |

---

## 六、数据库

### 6.1 `auth_credential` — 登录凭证表

```sql
CREATE TABLE auth_credential (
    id              BIGINT NOT NULL AUTO_INCREMENT,
    user_id         BIGINT NOT NULL COMMENT '用户ID（FK → user_base.id）',
    identity_type   VARCHAR(20) NOT NULL COMMENT '登录方式：phone / sms / wechat',
    identifier      VARCHAR(255) NOT NULL COMMENT '标识符（AES 加密的手机号）',
    credential      VARCHAR(255) NOT NULL COMMENT '凭证（bcrypt 密文）',
    created_time    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_time    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE INDEX uk_identity (identity_type, identifier),
    INDEX idx_user (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='登录凭证表';
```

**一条用户可有多条凭证**：password 登录、短信登录分开记录，同一 `identity_type` 下 `identifier` 唯一。

### 6.2 Redis 键设计

| Key | 类型 | TTL | 说明 |
|-----|------|-----|------|
| `auth:rt:{user_id}:{device_id}` | String | 15 天 | RT jti 值，用于轮换校验 |
| `auth:rt:{user_id}:devices` | Set | — | 用户活跃设备集合 |
| `auth:at:blacklist:{jti}` | String | AT 剩余有效期 | 注销拉黑的 AT |
| `auth:roles:{user_id}` | String (JSON) | 5 分钟 | 用户已认证角色缓存，值如 `[{"r":"owner","c":1001}]` |
| `sms:code:{phone}` | String | 5 分钟 | 短信验证码 |

---

## 七、安全机制

| 机制 | 说明 |
|------|------|
| **手机号加密存储** | `auth_credential.identifier` 使用 AES-256 加密，不存明文 |
| **密码 Bcrypt** | `auth_credential.credential` 使用 bcrypt 加盐哈希 |
| **传输加密** | 客户端使用 RSA 公钥加密手机号和密码后传输，服务端 RSA 私钥解密 |
| **AT 短有效期** | 15 分钟，即使泄露影响窗口有限 |
| **RT 轮换** | 每次刷新时旧 RT 立即失效，防止 RT 泄露后被复用 |
| **AT 黑名单** | 注销时 AT 加入黑名单，TTL 对齐 AT 过期时间，即时生效 |
| **JWT 密钥隔离** | AT 和 RT 使用不同签名密钥 |
| **验证码防重放** | 校验通过后立即删除 Redis 中的验证码 |
| **Saga 补偿** | 注册时 Credential 写入失败 → 软删除已创建的 User |

---

## 八、错误码

| 错误码 | 含义 |
|--------|------|
| 0 | 成功 |
| 50001 | 登录失败（手机号或密码错误 / 未注册） |
| 50002 | Token 已过期 |
| 50003 | RT 已失效（已注销或已被旋转淘汰） |
| 50004 | 验证码错误或已过期 |
| 50005 | 账号已被禁用 |
| 50006 | Token 已被拉黑（已注销） |
| 509503 | 获取用户信息失败（user-service 不可用） |
| 509504 | 获取用户角色失败 |
| 500400 | 参数校验失败 |

---

## 九、配置

```yaml
# rpc/etc/authservice.yaml
Name: auth.rpc
ListenOn: 0.0.0.0:8083

# etcd 服务注册
Etcd:
  Hosts:
    - localhost:2379
  Key: auth.rpc

# JWT 配置
JwtAuth:
  AccessSecret: ${JWT_ACCESS_SECRET}    # AT 签名密钥，从 .env 注入
  AccessExpire: 900                      # AT 过期时间（秒），15 分钟
  RefreshSecret: ${JWT_REFRESH_SECRET}   # RT 签名密钥，从 .env 注入
  RefreshExpire: 1296000                 # RT 过期时间（秒），15 天

# RSA 加密
RsaPrivateKey: |
  -----BEGIN RSA PRIVATE KEY-----
  ...
  -----END RSA PRIVATE KEY-----
RsaPublicKey: |
  -----BEGIN RSA PUBLIC KEY-----
  ...
  -----END RSA PUBLIC KEY-----

# AES 加密（手机号存储）
AesKey: ${AES_KEY}

# MySQL
DataSource: ${AUTH_DB_DSN}

# User Service gRPC 客户端
UserServiceRpc:
  Etcd:
    Hosts:
      - localhost:2379
    Key: user.rpc
```

---

## 十、gRPC 接口

定义在 `api-proto/api/auth/v1/auth.proto`（`AuthService`）：

| RPC | 说明 | 认证 | 超时 | 调用 GetUserRoles |
|-----|------|------|------|:---:|
| `Login` | 密码登录 | PUBLIC | 3s | ✅ |
| `LoginSms` | 短信验证码登录 | PUBLIC | 5s | ✅ |
| `Register` | 注册 | PUBLIC | 5s | ✅（新用户 roles=[]） |
| `RefreshToken` | 刷新 AT+RT | PUBLIC | 2s | ✅（重新拉取） |
| `Logout` | 注销 | JWT | 2s | — |
| `ValidateToken` | 校验 AT（网关） | INTERNAL | 500ms | — |

---

## 十一、服务依赖

```
auth-service
  ├── user-service (gRPC)
  │     ├── CreateUser        注册时创建用户档案
  │     ├── GetUserByPhone    登录时获取用户信息
  │     ├── GetUserRoles      签发 Token 时获取已认证角色 ⬅ 新增调用
  │     └── UpdateUser        Saga 补偿：软删除用户
  ├── MySQL (auth 库)
  │     └── auth_credential   登录凭证表
  └── Redis
        ├── RT 持久化（auth:rt:*）
        ├── AT 黑名单（auth:at:blacklist:*）
        ├── 角色缓存（auth:roles:*，TTL 5 分钟，user-service 主动失效）
        └── 短信验证码（sms:code:*）
```

---

## 十二、实现计划

### 12.1 变更范围

| 文件 | 变更 | 说明 |
|------|------|------|
| `api-proto/api/auth/v1/auth.proto` | 修改注释 | 更新 AT payload 说明，去除"极简"描述，增加 roles |
| `rpc/internal/logic/auth/loginlogic.go` | 修改 | 签发 AT 前调用 `GetUserRoles`，JWT claims 增加 roles |
| `rpc/internal/logic/auth/loginsmslogic.go` | 修改 | 同上 |
| `rpc/internal/logic/auth/registerlogic.go` | 修改 | 签发 AT 前调用 `GetUserRoles`（新用户 roles=[]），Cache-Aside |
| `rpc/internal/logic/auth/refreshtokenlogic.go` | 修改 | 刷新时调用 `GetUserRoles` 重新拉取角色，Cache-Aside |
| `rpc/internal/svc/servicecontext.go` | 不变 | Redis 客户端已存在，无需变更 |
| `api/internal/types/types.go` | 不变 | API 层透传 proto 响应，无需变更 |

### 12.2 风险与注意事项

1. **GetUserRoles 调用失败处理**：user-service 不可用时，是拒绝登录还是签发无角色 Token？
   - **建议**：拒绝登录。无角色 Token 会导致所有请求被 Gateway 拒绝，与其让用户登录后什么也做不了，不如明确报错。
   
2. **JWT 体积增长**：每个角色约 28 字节，典型 1-2 个角色，影响可忽略。设定 50 个条目上限，超限时省略 `c` 字段，Gateway 降级为 CheckAccess。

3. **Register 后的 roles**：新注册用户无角色，`roles=[]`。这是正常的——用户需要先加入小区、申请角色、认证通过后才能获得有效角色。

4. **merchant 角色的 `c=0`**：merchant 是全局角色，不绑定具体小区。广告投放范围由 ad-service 的 `ad_merchant_scope` 表独立管理，不进 JWT。Gateway 对 merchant 角色的鉴权不检查 `c`。

5. **Proto 注释更新**：需要修改 proto 中"极简 Payload"的描述，使其与实现一致。

6. **API Gateway 同步更新**：Gateway 需要感知 `c` 缺失时的降级逻辑，以及 merchant 角色不校验 `c` 的特殊规则。

7. **Redis 角色缓存一致性**：user-service 已实现角色变更时 `DEL auth:roles:{user_id}`。auth-service 仅读取缓存，不写入。TTL 5 分钟兜底，确保即使 DEL 失败也能自愈。

### 12.3 实施步骤

```
1. 修改 proto 注释（去除"极简"描述，增加 roles 说明）
2. 抽取公共函数 `getUserRolesWithCache(ctx, userId)` — Cache-Aside：
   a. Redis GET auth:roles:{userId} → 命中返回
   b. 未命中 → gRPC GetUserRoles → Redis SET EX 300 → 返回
3. 在 loginlogic.go 中调用 `getUserRolesWithCache` + roles 写入 JWT
4. 在 loginsmslogic.go 中增加同上逻辑
5. 在 registerlogic.go 中增加同上逻辑（新用户 roles=[]）
6. 在 refreshtokenlogic.go 中增加同上逻辑（重新拉取角色）
7. 构建验证：go build ./...
8. 告知用户：API Gateway 需要同步更新，支持从 JWT 解析 roles 做常规鉴权
```
