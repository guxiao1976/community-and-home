---
triggers: ["axios", "AxiosResponse", "拦截器", "解包", "as unknown as any", "request.get", "request.post", "装饰性泛型", "type-safety", "双断言"]
service: all
severity: should-follow
type: pitfall
status: active
created: 2026-08-12
updated: 2026-08-12
---

# axios 泛型(AxiosResponse)与拦截器解包(data)不符导致逐处双断言

## 为什么会有这条经验

request.ts 响应拦截器在运行时已解包 `data` 返回，但 axios 的泛型仍把 `request.get<T>` / `request.post<T>` 标注为 `Promise<AxiosResponse<T>>`，调用方为拿到真正的数据类型只能 `res as unknown as X`（或 `as unknown as any`）双断言，泛型 `T` 沦为装饰性。

## 怎么做

1. 定义类型化 request 包装：`request<T>(): Promise<T>`（拦截器成功路径返回 `data as T`），各 api 层 `const res = await request.post<CommunityMembership>(...)` 直接得到 `CommunityMembership`，消除逐处断言
2. 新增代码禁止继续 `as unknown as any`；catch 用 `unknown` 收窄

## 怎么验证

- 变更文件 grep `as unknown as` / `as any` 数量不增加
- 调用方不再需要把 AxiosResponse 强转

## 触发场景

- 编写/修改 api 层调用、见到 `request.get<any>` + 双断言模式时
