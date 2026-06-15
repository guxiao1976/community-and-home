---
triggers: ["API 响应", "response.Success", "双层嵌套", "BaseResponse", "goctl Response", "axios 拦截器", "response format", "Logic 返回"]
service: all
severity: must-follow
status: active
created: 2026-06-12
updated: 2026-06-12
---

# API 响应必须单层包装

## 为什么会有这条经验

ai-model-service 的 REST API 出现双层嵌套：`{code:0, data: {code:0, data: {实际数据}}}`。根因是 goctl 生成的 Response 类型嵌入了 `BaseResponse`（含 `code, msg, data`），Logic 返回这些类型后，Handler 又用 `response.Success(w, resp)` 包了一层 `Body{code, msg, data}`。前端 axios 拦截器剥掉外层 data 后，拿到的仍是 `{code, 0, data: ...}`，实际业务数据还要再取一层 `.data`。

导致前端出现大量 `(res.data || res)` 兼容代码，且创建后的数据在列表中不显示、测试结果反馈 `undefined` 等一系列 bug。

## 怎么做

1. **Logic 返回纯业务数据**（struct/pointer），不嵌 `BaseResponse`
2. **Handler 统一用 `response.Success(w, data)` 包一层**
3. goctl 生成的 `*Response` 类型只用于 swagger 文档生成，不作为 Logic 返回值

```go
// ✅ 正确：Logic 返回业务数据
func (l *MyLogic) DoSomething(req *types.XxxRequest) (*MyData, error) { ... }

// ❌ 错误：Logic 返回 goctl Response 类型（嵌 BaseResponse）
func (l *MyLogic) DoSomething(req *types.XxxRequest) (*types.XxxResponse, error) { ... }
```

## 怎么验证

- 用 curl 调用任意 REST API，响应体 `data` 字段的值本身不应再含 `code` 字段。即 `jq '.data.code'` 应为 `null`
- 前端代码中不应出现 `(res as any)?.data || res` 这种兜底写法

## 关联经验

- [[api-calling-verify-route]]
