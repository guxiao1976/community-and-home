---
triggers: ["logout", "退出登录", "缓存", "user_phone", "共享设备", "跨账号", "清理", "storage", "clearTokens", "兜底字段"]
service: web/mobile
severity: should-follow
type: pitfall
status: active
created: 2026-08-17
updated: 2026-08-17
last_applied: null
apply_count: 0
---

# 退出登录需清理登录期写入的本地缓存字段（如 user_phone），防共享设备跨账号残留

## 为什么会有这条经验

登录/注册成功时写入本地 storage 的兜底字段（如 `user_phone`，经 auth-flow handleAuthSuccess 写入）在退出登录时（`userStore.logout()` → `clearTokens()`）**不会**被清除，因为 `clearTokens()` 只清 token 相关 key。共享设备上：用户 A 退出 → 用户 B 登录，若 B 的 profile 缺失/脱敏且登录流程未写入兜底（或写入失败），displayPhone 等会显示上一账号 A 的真实手机号，构成信息泄漏/串号。

## 怎么做

1. 登录期写入的敏感兜底缓存（手机号等）应在 `logout()` 中一并 `uni.removeStorageSync` 清除
2. 将「缓存字段清单」收敛到共享模块（如 reg-pending 契约源模式），logout 与登录写入同一来源，避免各自内联 magic string
3. 自测：退出登录后断言兜底 key 已清空

案例：`web/mobile/src/stores/user.ts` `logout()` 增加 `uni.removeStorageSync('user_phone')`（与 auth-flow handleAuthSuccess 的写入对应）。

## 怎么验证

- 退出登录用例断言兜底 key 已清空（my.spec.ts 退出用例）
- grep 确认 logout 清理清单与登录写入清单一致
