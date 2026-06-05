---
triggers: ["proto", "int64", "jstype", "JS_STRING", "Snowflake", "ID", "精度丢失", "序列化"]
service: api-proto
severity: must-follow
status: active
created: 2026-06-05
updated: 2026-06-05
---

# Proto int64 字段必须加 jstype=JS_STRING

## 为什么会有这条经验
Snowflake 生成 19 位 ID，超过 JavaScript `Number.MAX_SAFE_INTEGER`（约 16 位）。
前端 JSON 解析时 int64 数字精度丢失，导致 ID 在请求/响应中不一致。

## 怎么做
1. Proto 中所有 int64 ID 字段加 `[jstype = JS_STRING]` 注解
2. Go 端 REST API 类型中 int64 ID 字段使用 `json:"...,string"` 标签
3. 前端所有 ID 字段 TypeScript 类型为 `string`，axios 使用 `lossless-json` 解析器

## 怎么验证
- `cd api-proto && make breaking-check` 会检测缺失的注解
- 前端请求-响应 ID 一致性测试
- 注意：`repeated int64` 字段也需要加 `[jstype = JS_STRING]`

## 关联经验
- [[grpc-only-comms]]
