# Auth Service 后端修改清单

> 由 web/pc 前端 Claude 整理。前端已完成类型对齐和 RSA 加密，以下 3 个后端问题需要修复。

## 背景

前端登录/注册请求已改为 RSA 加密传输：
- `LoginRequest`: `{ encryptedPhone, encryptedPassword, deviceId, deviceType }`
- `RegisterRequest`: `{ encryptedPhone, encryptedPassword?, smsCode, nickname, deviceId, deviceType }`
- 加密方式：RSA-OAEP + SHA-256 + Base64（对齐 `common/pkg/crypto/rsa.go` 的 `RSAEncrypt`）
- 公钥获取：`GET /api/auth/public-key`

## 问题 1（P0）：缺少公钥端点

**现状**：没有对外暴露 RSA 公钥的 API，前端无法加密。

**修改**：

### 1a. common/pkg/crypto/rsa.go — 新增 getter
```go
// GetRSAPublicKey 返回当前全局 RSA 公钥（PEM 格式）
func GetRSAPublicKey() string {
    rsaKeyMutex.RLock()
    defer rsaKeyMutex.RUnlock()
    return /* PEM 字符串，从 globalRSAPublicKey 反序列化，或直接存 PEM */
}
```
> 当前 `InitRSA` 存的是 `*rsa.PublicKey` 对象，需要额外存一份 PEM 字符串，或新增一个 getter 返回 PEM。

### 1b. api/internal/config/config.go — 新增配置字段
```go
RsaPublicKey string  // RSA 公钥（PEM，与 RPC 配置相同）
```

### 1c. api/etc/auth-api.yaml — 新增配置值
```yaml
RsaPublicKey: "-----BEGIN PUBLIC KEY-----\n...\n-----END PUBLIC KEY-----"
```

### 1d. api/internal/types/types.go — 新增响应类型
```go
type PublicKeyResp struct {
    PublicKey string `json:"public_key"`
}
```

### 1e. api/internal/logic/publickeylogic.go（新建）— 逻辑层
```go
func (l *PublicKeyLogic) GetPublicKey() (*types.PublicKeyResp, error) {
    return &types.PublicKeyResp{PublicKey: l.svcCtx.Config.RsaPublicKey}, nil
}
```

### 1f. api/internal/handler/routes.go — 注册路由
在公开路由组（`rest.WithPrefix("/api/auth")`）中添加：
```go
{
    Method:  http.MethodGet,
    Path:    "/public-key",
    Handler: PublicKeyHandler(svcCtx),
},
```

---

## 问题 2（P0）：RegisterReq 字段名不匹配

**现状**：后端 `RegisterReq` 的 phone 字段是 `json:"phone"`（明文），但前端现在发 `encryptedPhone`（RSA 密文）。

**修改**：

### 2a. api/internal/types/types.go
```go
// 改前
type RegisterReq struct {
    Phone    string `json:"phone"`
    ...
}

// 改后
type RegisterReq struct {
    EncryptedPhone string `json:"encrypted_phone"`
    ...
}
```

### 2b. api/internal/logic/registerlogic.go
```go
// 改前
resp, err := l.svcCtx.AuthRpc.Register(l.ctx, &authv1.RegisterRequest{
    Phone: req.Phone,
    ...
})

// 改后
resp, err := l.svcCtx.AuthRpc.Register(l.ctx, &authv1.RegisterRequest{
    Phone: req.EncryptedPhone,  // 字段名变更
    ...
})
```

> 注意：proto 中 `RegisterRequest.phone` 是明文，但这里 API 层收到的是 RSA 密文。需要在 API 层调用 `crypto.RSADecrypt(req.EncryptedPhone)` 解密后传入 proto。

---

## 问题 3（P1）：短信验证码未校验

**现状**：`rpc/internal/logic/auth/registerlogic.go:47` 有 TODO，验证码在 Redis 中但从未被读取比对。任意 6 位数字都能通过注册。

**修改**：

### 3a. rpc/internal/logic/auth/registerlogic.go
在 `// 1. 校验短信验证码` 处（约 L47），增加：
```go
// 从 Redis 读取验证码
codeKey := fmt.Sprintf("sms:code:%s", in.Phone)
storedCode, err := l.svcCtx.RedisClient.Get(l.ctx, codeKey).Result()
if err != nil || storedCode == "" {
    return &authv1.RegisterResponse{
        Base: responsex.NewBaseRespWithError(50004, "验证码已过期，请重新获取"),
    }, nil
}
if storedCode != in.SmsCode {
    return &authv1.RegisterResponse{
        Base: responsex.NewBaseRespWithError(50004, "验证码错误"),
    }, nil
}
// 验证通过，删除验证码（防重放）
l.svcCtx.RedisClient.Del(l.ctx, codeKey)
```

---

## 操作顺序

```
1. 修改 common/pkg/crypto/rsa.go（如需全局 Claude 协作）
2. 修改 api/internal/config + yaml（公钥配置）
3. 实现 public-key 端点（types → logic → handler → routes）
4. 修改 RegisterReq 字段名 + 解密逻辑
5. 实现短信验证码校验
6. go build ./... 验证编译
7. 重启 auth-service API + RPC
```

## 参考

- RSA 工具：`common/pkg/crypto/rsa.go`（`InitRSA`, `RSADecrypt`, `RSAEncrypt`）
- 现有端点模式：`api/internal/handler/routes.go`（RegisterHandlers）
- 登录逻辑参考：`rpc/internal/logic/auth/loginlogic.go`（RSA 解密 → bcrypt 校验）
- 前端类型：`web/common/types/identity.ts`（LoginRequest, RegisterRequest 等）
