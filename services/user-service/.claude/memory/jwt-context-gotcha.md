---
name: user-service-jwt-context
description: go-zero JWT 中间件的 context 键名和类型——别再用 "userId" 了
metadata:
  type: reference
---

# JWT Context 取用户 ID 的正确方式

## 问题

在 user-service API 层的 logic 中，从 JWT 取 `userId` 时：

```go
// ❌ 错误：go-zero 不存 "userId" 键，永远是 nil，panic
userId := l.ctx.Value("userId").(int64)
```

## 正确方式

go-zero JWT 中间件用 JWT Payload 的**原始字段名**存入 context，JSON 数字类型为 `float64`：

```go
func getUserId(ctx context.Context) int64 {
    v := ctx.Value("user_id")
    if v == nil {
        return 0
    }
    switch n := v.(type) {
    case float64:
        return int64(n)
    case int64:
        return n
    case json.Number:
        id, _ := n.Int64()
        return id
    default:
        return 0
    }
}
```

字段名取决于 JWT 签发时的 key：auth-service 签发 JWT 用的是 `"user_id"`，所以取 `ctx.Value("user_id")`。

## 其他关键参数

- `MaxCommunities = 3`（最多加入 3 个小区）
- 重复加入同一小区：返回 `10007: "不能重复加入同一个小区"`
- 超限：返回 `10006: "最多加入 3 个小区"`

**Why:** go-zero 的 JWT 行为没有在官方文档明确说明，这个坑容易反复踩。

**How to apply:** 所有 user-service API 层新增 logic 都使用 `getUserId()`，不要直接 `ctx.Value(...)`。
