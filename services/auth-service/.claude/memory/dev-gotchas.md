---
name: auth-service-dev-gotchas
description: auth-service 开发环境的关键配置和常见问题
metadata:
  type: reference
---

# auth-service 开发环境注意事项

## 1. AES_KEY 格式要求

`common/pkg/crypto/aes.go` 的 `InitAES` 要求 key 必须是 **16、24 或 32 字节**的原始字符串，不是 Base64。

- ❌ `AES_KEY=F9PXfr0tO63uqi8EUqyBTrh8Eq6Lr7CfgqshlKODzy4`（44字符 Base64 → panic）
- ✅ `AES_KEY=49e163155b0e2c2569af909d7d64117e`（32字符 hex）

## 2. RSA 密钥路径

`rpc/etc/authservice.yaml` 中 `RsaPrivateKeyPath: "etc/rsa_private.pem"` 是相对路径。启动时必须从 `rpc/` 目录执行：

```bash
cd services/auth-service/rpc && go run authservice.go -f etc/authservice.yaml
```

脚本 `scripts/start.sh` 已修正为上述路径。**不要**从 `services/auth-service/` 目录启动 RPC。

## 3. 开发阶段验证码

修改 `api/internal/logic/smssendlogic.go:45`：
```go
code := "123456" // 开发阶段固定
```
重启 auth API 后生效。正式上线前记得改回随机生成。

## 4. 验证码 Redis 存储

- Key: `sms:code:{phone}`，Value: 6位数字字符串
- TTL: 300 秒
- 限流 Key: `sms:rate:{phone}`，TTL: 60 秒

**Why:** 每次调试登录都要查验证码太麻烦，固定后效率大幅提升。

**How to apply:** 开发环境固定验证码，提交代码前还原。
