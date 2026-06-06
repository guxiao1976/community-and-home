---
name: api-alignment-pitfalls
description: API 对接的三个常见坑——路径对齐、响应解包、JSON字段命名，移动端子Claude必读
metadata:
  type: reference
---

# API 对接三大坑

移动端对接后端 API 时，以下三个问题反复出现，根源是前端代码与后端契约未对齐。

## 坑1：API 路径不一致

**问题**：前端随意命名路径，与后端 routes.go 定义的不一致。

**正确做法**：写 API 调用前，必须查看 `services/auth-service/api/internal/handler/routes.go` 确认路径：
```
rest.WithPrefix("/api/auth") + Path: "/sms/send"      → POST /api/auth/sms/send
rest.WithPrefix("/api/auth") + Path: "/login/sms"     → POST /api/auth/login/sms
rest.WithPrefix("/api/auth") + Path: "/token/refresh" → POST /api/auth/token/refresh
```

## 坑2：响应拦截器解包层级不一致

**问题**：PC 端 `request.ts` 在 `code === 0` 时返回 `data`（已解包），移动端曾返回 `response.data`（整包），导致 API 调用方拿到的数据多一层嵌套。

**正确做法**：移动端 `request.ts` 的响应拦截器必须与 PC 端对齐：
```typescript
if (code === 0) {
  return data !== undefined ? data : response.data;  // 解包 data 字段
}
```
调用方直接 `return res`，不要 `return res.data`（会双重解包导致 undefined）。

## 坑3：JSON 字段名不匹配

**问题**：后端 types.go 的 JSON tag 是驼峰 `json:"publicKey"`，前端用了下划线 `data.public_key`，取到 undefined。

**正确做法**：以 `services/auth-service/api/internal/types/types.go` 中的 `json:"..."` tag 为准：
```go
PublicKey string `json:"publicKey"`   // 前端取 data.publicKey
```

**Why:** 后端 go-zero 默认用 camelCase 序列化，但具体取决于 struct tag。永远以 tag 为准。

**How to apply:** 每次新增 API 调用前，查看对应服务的 `types.go` 确认请求/响应字段名；修改后端响应格式时，同步更新前端类型定义。

## 坑4：Proto 返回的 CommunityMembership 不含小区名称

**问题**：`GetUserMemberships` 的 proto 定义只有 `community_id` 等 ID 字段，没有 `community_name`。
前端 store 的 `loadMemberships()` 如果简单覆盖，会把加入时设置的小区名清空。

**正确做法**：
1. 前端 `CommunityMembership` 接口使用 snake_case（`community_id`），匹配 protojson 输出
2. `loadMemberships()` 用 `existingMap` 保留已有的 `communityName`/`address`
3. 兼容处理：`m.community_id || m.communityId` 双读，防止前后端不同时上线

**Why:** 2026-06-06 发现加入小区后刷新页面显示"小区 undefined"，根因是 protojson 输出 `community_id`（蛇形），
但前端接口定义为 `communityId`（驼峰），取值为 undefined，拼接出"小区 undefined"。

**How to apply:** 
- Proto 定义字段使用 snake_case（如 `community_id`），没有显式 `json_name` 时 protojson 保持 snake_case
- 前端接口字段名必须与 API 实际输出一致，以 `types.go` 的 `json:"..."` tag 或 proto 字段名为准
- 当 API 不返回完整数据时，store 的 load 方法应合并已有数据而非全量覆盖
