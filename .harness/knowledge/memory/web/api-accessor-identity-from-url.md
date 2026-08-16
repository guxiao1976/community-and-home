---
triggers: ["viewer_id", "viewerId", "user-detail", "IDOR", "越权", "脱敏", "phone", "query参数", "访问者身份", "同屋互见"]
service: web/mobile
severity: must-follow
type: pitfall
status: active
created: 2026-08-16
updated: 2026-08-16
last_applied: null
apply_count: 0
---

# 禁止把 URL query 中的访问者身份传给权限/脱敏 API

## 为什么会有这条经验

前端页面从 URL query（攻击者可手工构造/深链）读取访问者身份（如 `viewer_id`）并原样传给权限敏感的掩码/脱敏 API（如 `GET /api/users/:id?viewer_id=xxx` 决定同屋手机号明文/脱敏）。若后端以该客户端参数（而非 JWT 身份）判定数据范围，即产生 IDOR/越权：构造 `viewer_id=他人` 可看他人真实手机号/房屋号。

## 怎么做

1. **前端一律从已认证会话派生访问者身份**（`userStore.userId` / JWT），绝不接受 URL 传入的 viewer/accessor 参数覆盖
2. **后端必须忽略客户端 viewer 身份，以 JWT 为准**（数据范围/脱敏决策权威在后端）
3. 前端传参仅作提示性上下文，不承载权限决策

案例：`web/mobile/src/pages/user-detail/user-detail.vue` `load()` 原 `options?.viewer_id || userStore.userId` 属反模式，已删除 URL 分支，改为 `const viewerId = userStore.userId || undefined;`。

## 怎么验证

- grep 确认无 `options?.viewer_id` / `?viewer_id=` 从 URL 读取并传给权限 API
- 后端接口以 JWT 身份判定数据范围，viewer_id 参数仅作展示提示
