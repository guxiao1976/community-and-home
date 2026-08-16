---
triggers: ["跨页", "一次性", "临时数据", "敏感数据", "内存态", "模块级", "变量", "sessionStorage", "localStorage", "持久化", "reg_pending", "暂存", "navigateTo"]
service: all
severity: should-follow
type: guideline
status: active
created: 2026-08-16
updated: 2026-08-16
last_applied: null
apply_count: 0
---

# 跨页一次性敏感数据优先走内存态载体，而非持久化

## 为什么会有这条经验

页面栈内导航（如 login → agreement）需要传递一次性的敏感临时数据（验证码、手机号、设备指纹等）。若直接用 `uni.setStorageSync` / localStorage 持久化传递，数据会跨会话残留，扩大暴露面并产生共享设备复用风险。此前跨页一次性数据默认走持久化是惯性做法，缺少「先内存态、后按需镜像」的分层决策。

## 怎么做

1. **模块级内存变量为主载体**：跨页一次性数据放模块级变量，页面栈内导航直接可读，不落任何持久化
2. **仅在确需跨刷新保留时才镜像 sessionStorage**：带 TTL 信封（`{data, expiresAt}`），过期即清，且访问容错 try/catch
3. **H5 下绝不 localStorage**：`uni.setStorageSync` 在 H5 映射 localStorage，禁止用于敏感一次性数据
4. **非 H5 环境统一走内存态**：小程序等无 window 的环境直接内存传递
5. 按「内存态优先 → session 镜像降级 → localStorage 禁止」的顺序决策，把敏感性纳入存储介质选型

案例：`web/mobile/src/utils/reg-pending.ts` — 模块级内存变量为主载体（页面栈内导航可读），仅 H5 镜像到 sessionStorage（TTL 5 分钟），非 H5 走内存态，localStorage 零触碰。

## 怎么验证

- 跨页导航数据传递路径：内存态变量存在且可读，非必要不落持久化
- 只有跨刷新场景才出现 sessionStorage，且带 TTL 信封
- 敏感数据（验证码/手机号/指纹）无 localStorage 残留

## 关联经验

[[sms-code-persist-localstorage]] [[frontend-cross-page-storage-contract]]
