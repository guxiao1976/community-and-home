---
triggers: ["接口响应", "资源包装", "返回形状", "回填", "api封装", "membership", "data.membership", "request.post", "响应结构", "mock", "假绿", "envelope", "joinCommunity"]
service: web/mobile
severity: should-follow
type: pitfall
status: active
created: 2026-08-17
updated: 2026-08-17
last_applied: null
apply_count: 0
---

# 前端 api 封装须按 REST 响应对象形状解出命名资源字段（勿把 {membership} 当资源返回）

## 为什么会有这条经验

后端 REST 响应经 responsex 单层包装 `{code,message,data:{...}}`，request.ts 拦截器成功路径解包 `data` 后返回的仍是「响应对象」，不是资源本身。如 JoinCommunity 的 REST 类型为 `{membership:{...}}`，拦截器返回 `{membership}`，而 `joinCommunity()` 若直接 `return res as unknown as CommunityMembership`，调用方取 `res.id` 恒为 undefined——「membershipId 回填」等新功能在生产静默失效，全靠后续兜底（getUserMemberships）掩盖，且单测在错误层级 mock 导致假绿（mock request.post 直接 resolve 裸资源对象、或 mock api 函数本身，均未镜像真实 `{membership:{...}}` envelope）。

## 怎么做

1. api 封装按后端 REST 类型（types.go）解出命名资源字段：`const data = res as { membership: CommunityMembership }; return data.membership;`，返回资源本身
2. 调用方消费返回值前，确认返回的是资源对象还是 `{命名字段}` 包装对象（对照后端 types.JoinCommunityResp 结构）
3. 单测 mock 必须镜像真实响应形状（`{membership:{...}}`），禁止直接 resolve 裸资源对象；对「回填/取字段」类逻辑单独断言真实 envelope 下不 undefined

## 怎么验证

- 对每个返回单资源/列表资源的 api 函数，比对后端 REST types 的字段结构，确认返回类型与实际解包形状一致
- grep mock 是否含 envelope 包装；grep `res as unknown as X` 的调用方是否实际消费了不存在于包装上的字段

## 关联经验

[[api-response-single-wrap]] [[axios-interceptor-unwrap-typing-mismatch]] [[verify-api-before-calling]]
