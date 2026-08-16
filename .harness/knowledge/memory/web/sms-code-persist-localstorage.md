---
triggers: ["smsCode", "短信验证码", "一次性验证码", "localStorage", "sessionStorage", "持久化", "残留", "共享设备", "OTP", "setStorageSync", "reg_pending"]
service: all
severity: must-follow
type: pitfall
status: active
created: 2026-08-16
updated: 2026-08-16
last_applied: null
apply_count: 0
---

# 一次性短信验证码禁止落 localStorage 持久化残留

## 为什么会有这条经验

登录/注册流程中的短信验证码（smsCode）是一次性敏感数据，只能使用一次。若前端用 `uni.setStorageSync` / `localStorage.setItem` 持久化暂存（如登录页 50001 未注册时暂存 {phone, smsCode} 跳协议页），H5 下会写入 localStorage 且长期残留。共享设备上验证码 + 手机号可被复用，形成账号风险；即使验证码用过一次，残留值也会误导后续校验。

## 怎么做

1. **一次性验证码仅走内存态载体**：模块级变量为主载体，页面栈内导航（login → agreement）直接可读，不落任何持久化
2. **仅 H5 镜像到 sessionStorage 并加 TTL**：如需跨刷新保留，`sessionStorage.setItem` 写入 `{data, expiresAt}` 信封，TTL（如 5 分钟）过期即整体清除并返回 null，`save` 先清旧再写
3. **绝不写 localStorage**：`uni.setStorageSync` 在 H5 即映射 localStorage，禁止用于敏感一次性数据
4. **sessionStorage 访问一律 try/catch 容错**：隐私模式 / 被禁用时降级内存态，不抛出
5. 清除路径（`clear`）必须同时清内存态 + session 镜像

案例：`web/mobile/src/utils/reg-pending.ts` — 模块级内存变量为主载体 + 仅 H5 sessionStorage 镜像（TTL 5 分钟）+ localStorage 零触碰，login.vue / agreement.vue 共享该契约源。

## 怎么验证

- grep 前端登录/注册相关 `setStorageSync` / `localStorage.setItem`，确认一次性验证码未落持久化
- 检查跨页共享模块是否写 sessionStorage 且带 TTL 信封，过期的数据被清除而非残留
- 共享设备场景：验证码使用一次后刷新页面，无残留可复用

## 关联经验

[[frontend-cross-page-storage-contract]] [[cross-page-sensitive-temp-data-storage]]
