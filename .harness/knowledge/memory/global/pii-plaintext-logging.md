---
triggers: ["日志", "手机号", "phone", "PII", "明文", "脱敏", "Infof", "加密", "AES"]
service: all
severity: should-follow
type: pitfall
status: active
created: 2026-08-12
updated: 2026-08-12
---

# 加密落库的敏感字段禁止明文打日志

## 为什么会有这条经验

create_user_logic.go 里 `l.Infof("CreateUser success, userId=%d, phone=%s", userId, in.Phone)` 把客户端明文手机号写进日志，而数据库里 phone 是 AES 加密存储——落库加密但日志泄密，两头不一致。

## 怎么做

1. 凡要求加密落库的敏感字段（手机号/身份证号），日志一律脱敏（只打后 4 位或打加密值）
2. 严禁直接打客户端明文
3. 安全审查固定检查项：加密字段明文进日志 = WARNING

## 怎么验证

- grep 变更代码中的 `Infof`/`Errorf`，确认敏感字段已脱敏
- 日志系统检索无明文手机号

## 关联经验

[[phone-encryption]]
