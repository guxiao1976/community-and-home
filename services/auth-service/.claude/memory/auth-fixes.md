---
triggers: ["RSA", "公钥", "public-key", "register", "短信验证码", "RegisterReq", "加密"]
service: auth-service
type: pitfall
severity: must-follow
status: active
created: 2026-06-17
updated: 2026-08-09
apply_count: 0
---

# Auth Service 后端修改清单（历史遗留）

> 由 web/pc 前端 Claude 整理。前端已完成类型对齐和 RSA 加密，以下 3 个后端问题需要修复。

## 背景

前端登录/注册请求已改为 RSA 加密传输：
- `LoginRequest`: `{ encryptedPhone, encryptedPassword, deviceId, deviceType }`
- `RegisterRequest`: `{ encryptedPhone, encryptedPassword?, smsCode, nickname, deviceId, deviceType }`
- 加密方式：RSA-OAEP + SHA-256 + Base64（对齐 `common/pkg/crypto/rsa.go` 的 `RSAEncrypt`）
- 公钥获取：`GET /api/auth/public-key`

## 问题 1（P0）：缺少公钥端点

### 1a. common/pkg/crypto/rsa.go — 新增 getter
### 1b. api/internal/config/config.go — 新增配置字段
### 1c. api/etc/auth-api.yaml — 新增配置值
### 1d. api/internal/types/types.go — 新增响应类型
### 1e. api/internal/logic/publickeylogic.go（新建）
### 1f. api/internal/handler/routes.go — 注册路由

## 问题 2（P0）：RegisterReq 字段名不匹配

后端 `RegisterReq` 的 phone 字段是 `json:"phone"`（明文），但前端发 `encryptedPhone`（RSA 密文）。需要在 API 层 `registerlogic.go` 中调用 `crypto.RSADecrypt(req.EncryptedPhone)` 解密后传入 proto。

## 问题 3（P1）：短信验证码未校验

`rpc/internal/logic/auth/registerlogic.go:47` 有 TODO，验证码在 Redis 中但从未被读取比对。

## 参考

- RSA 工具：`common/pkg/crypto/rsa.go`
- 现有端点模式：`api/internal/handler/routes.go`
- 登录逻辑参考：`rpc/internal/logic/auth/loginlogic.go`
