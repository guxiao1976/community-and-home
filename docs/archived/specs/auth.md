## 项目背景

系统唯一登录与令牌管理中台。采用 AT+RT 双 Token 机制，实现业务请求的高性能无状态验证，以及登录状态的安全强管控。遵循认权分离原则，Auth 只管身份，不管权限。

## 数据模型

- **MySQL (`db_auth`)**:
    - `auth_credential`: `id`, `user_id`, `identity_type`, `identifier(RSA密文)`, `credential(bcrypt密文)`
- **Redis**:
    - RT 存储: `auth:rt:{user_id}:{device_id}` = `{jti}`, TTL=15天
    - AT 黑名单: `auth:at:blacklist:{jti}` = `1`, TTL=AT剩余过期时间

## 接口清单

- `Login(LoginReq) returns (LoginResp)`
- `RefreshToken(RefreshTokenReq) returns (LoginResp)`
- `Logout(LogoutReq) returns (BaseResp)`

## 核心逻辑流

### 1. 登录签发

1. 接收账密（RSA 密文传输），解密后校验 bcrypt。
2. 校验通过，生成**极简 AT** (Payload 仅含 `user_id`, `jti`, `exp`，有效期 15 分钟)。
3. 生成 RT (Payload 含 `jti`, `user_id`, `device_id`，有效期 15 天)。
4. **Redis 持久化 RT**：`SET auth:rt:{user_id}:{device_id} {rt_jti} EX 1296000` (15天)。
5. 返回 AT 和 RT 给客户端。

### 2. AT 过期无感刷新 (核心难点)

1. 客户端携带 RT 请求 `RefreshToken` 接口。
2. 服务端解析 RT 获取 `user_id`, `device_id`, `jti`。
3. 去 Redis 查询 `auth:rt:{user_id}:{device_id}` 的值。
    - **如果查不到或值不等于 jti**：说明已注销或已被旋转机制淘汰，返回 403，强制重新登录。
    - **如果查到且匹配**：确认 RT 合法。
4. **RT 旋转机制 (防泄露)**：
    - 生成全新的 AT 和 新的 RT (new_jti)。
    - **删除旧 RT**：`DEL auth:rt:{user_id}:{device_id}`
    - **写入新 RT**：`SET auth:rt:{user_id}:{device_id} {new_jti} EX 1296000` (重置 15 天 TTL)。
5. 返回新的双 Token 给客户端。

### 3. 主动注销 (封号/退出登录)

1. 解析当前请求的 AT，提取 `jti` 和剩余过期时间 `remaining_ttl`。
2. **拉黑当前 AT**：`SET auth:at:blacklist:{jti} 1 EX {remaining_ttl}`，防止注销后 AT 在 15 分钟内仍可用。
3. **清除当前设备的 RT**：`DEL auth:rt:{user_id}:{device_id}`。AT 过期后，黑客即使拿着旧 RT 也无法刷新。
4. (可选) 强踢全设备：`KEYS auth:rt:{user_id}:*` 全部删除。

### 网关协同逻辑

API 网关拦截业务请求时：

1. 验证 AT 签名和 `exp`。
2. (可选高安全模式) 检查 `auth:at:blacklist:{jti}` 是否存在。
3. 验证通过，**只提取 `user_id`**，注入到下游微服务的 HTTP Header 或 gRPC Metadata 中，由权限中心处理后续鉴权。
