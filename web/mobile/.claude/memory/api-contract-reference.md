---
name: api-contract-reference
description: 移动端依赖的 auth-service API 契约快照——路径、请求体、响应体、JSON字段名
metadata:
  type: reference
---

# Auth Service API 契约参考

## 端点清单（`services/auth-service/api/internal/handler/routes.go`）

| 方法 | 完整路径 | 说明 | 认证 |
|------|---------|------|:--:|
| POST | `/api/auth/sms/send` | 发送短信验证码 | 无 |
| POST | `/api/auth/login/sms` | 短信验证码登录 | 无 |
| POST | `/api/auth/register` | 注册 | 无 |
| POST | `/api/auth/token/refresh` | 刷新 Token | 无 |
| GET | `/api/auth/public-key` | 获取 RSA 公钥 | 无 |

## 关键字段（`services/auth-service/api/internal/types/types.go`）

- `SmsSendReq.Phone` → `json:"phone"` — 明文，不加密
- `LoginSmsReq.EncryptedPhone` → `json:"encryptedPhone"` — RSA 密文
- `RegisterReq.EncryptedPhone` → `json:"encryptedPhone"` — RSA 密文
- `RegisterReq.Nickname` → `json:"nickname"` — 明文
- `PublicKeyResp.PublicKey` → `json:"publicKey"` — **驼峰**

## SMS 验证码（`services/auth-service/api/internal/logic/smssendlogic.go`）

- Redis Key: `sms:code:{phone}`，Value: 6位数字字符串，TTL: 300s
- 限流 Key: `sms:rate:{phone}`，TTL: 60s

## RSA 加密（`common/pkg/crypto/rsa.go`）

- 算法：RSA-OAEP + SHA-256
- 编码：Base64 StdEncoding
- 前端实现：`web/mobile/src/utils/crypto.ts`（Web Crypto API）
- 公钥端点重启后可能变化，sessionStorage 缓存

**Why:** 每次调试 API 对接问题都去翻后端代码太慢，快照在这里可以快速对照。

**How to apply:** 新增 API 调用前对照此表确认路径和字段名；后端接口变更时更新此文档。
