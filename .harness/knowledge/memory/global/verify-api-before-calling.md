---
triggers: ["API路由", "404", "catch空", "静默吞错", "前端调用API", "routes.go", "端点验证", "HTTP 404"]
service: all
type: pitfall
severity: must-follow
status: active
created: 2026-06-06
updated: 2026-06-09
last_applied: null
apply_count: 0
---

# 前端调用 API 前必须验证路由存在，禁止静默吞错

## 为什么会有这条经验

2026-06-06 发现 `getUserProfile()` 调用 `/api/users/profile`，该路由根本不存在，但 catch 块为空，
导致 userStore.user 永远为 null，所有依赖 isLoggedIn 的页面显示"点击登录"。
排查链路长（前端 → API → 路由 → 逻辑 → gRPC），如果一个环节验证过就能避免。

## 怎么做

1. 查看对应服务的 `api/internal/handler/routes.go`，确认 Method + Path 已注册
2. 查看 `api/internal/types/types.go`，确认请求/响应结构体的 JSON tag 与前端一致
3. 启动服务后用 curl 测试端点可达（HTTP 200/401/400 均可，不能是 404）

## 怎么验证

```bash
# 确认路由已注册
grep -n "路由路径" services/<name>/api/internal/handler/routes.go

# curl 验证端点可达
curl -s -o /dev/null -w "%{http_code}" http://localhost:<port>/api/<path>
# 期望：非 404
```

## 禁止静默吞错

```typescript
// ❌ 坏模式：catch 为空，调用方不知道失败了
try { await someApi(); } catch {}

// ✅ 好模式：至少打日志
try { await someApi(); } catch (e) { console.warn('[module] API failed:', e); }
```

## 关联经验

- [[pre-commit-checks]]
- [[grpc-only-comms]]
