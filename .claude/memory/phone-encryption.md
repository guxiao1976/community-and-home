---
triggers: ["手机号", "phone", "加密", "AES", "解密", "EncryptedString", "显示", "乱码"]
service: all
severity: must-follow
status: active
created: 2026-06-08
updated: 2026-06-08
last_applied: null
apply_count: 0
---

# 手机号 AES 加密存储，读取时必须解密

## 为什么会有这条经验

用户服务（user-service）在创建用户时调用 `crypto.AESEncrypt(phone)` 加密手机号后存入数据库。但 `toProtoUser` 中直接返回了 `u.Phone`（密文），导致前端显示乱码。直到加上 `crypto.AESDecrypt(phone)` 才正常。

## 怎么做

1. **写入时加密**：`crypto.AESEncrypt(phone)` → 密文入库
2. **读取时解密**：`crypto.AESDecrypt(u.Phone)` → 明文返回前端
3. **解密失败兜底**：如果解密失败（密钥不匹配、已是明文），返回原始值
4. 所有读取用户的路径（`toProtoUser`、`toProtoUsers`、列表查询等）都必须解密

## 怎么验证

- 前端「我的」页面手机号显示正常（11 位数字，非乱码）
- 前端登录/注册流程手机号展示正常
- API 响应中 `phone` 字段为明文

## 注意事项

- user-service 使用 go-zero sqlx（非 GORM），`model.EncryptedString` 自动解密仅 GORM 可用，sqlx 需手动调用 `crypto.AESDecrypt`
- 解密依赖 `AES_KEY` 环境变量，确保服务启动时 `.env` 已加载

## 关联经验
- [[proto-jstype]]
- [[verify-api-before-calling]]
