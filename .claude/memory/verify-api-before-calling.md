---
name: verify-api-before-calling
description: 前端调用 API 前必须验证后端路由存在，禁止静默吞错
metadata:
  type: feedback
---

前端每新增一个 API 调用，必须先确认后端路由已实现且可用，不能假设"后端应该有了"。

**验证步骤：**
1. 查看对应服务的 `api/internal/handler/routes.go`，确认 Method + Path 已注册
2. 查看 `api/internal/types/types.go`，确认请求/响应结构体的 JSON tag 与前端一致
3. 启动服务后用 curl 测试端点可达（HTTP 200/401/400 均可，不能是 404）

**禁止静默吞错：**
```typescript
// ❌ 坏模式：catch 为空，调用方不知道失败了
try { await someApi(); } catch {}

// ✅ 好模式：至少打日志
try { await someApi(); } catch (e) { console.warn('[module] API failed:', e); }
```

**Why:** 2026-06-06 发现 `getUserProfile()` 调用 `/api/users/profile`，该路由根本不存在，但 catch 块为空，
导致 userStore.user 永远为 null，所有依赖 isLoggedIn 的页面显示"点击登录"。
排查链路长（前端 → API → 路由 → 逻辑 → gRPC），如果一个环节验证过就能避免。

**How to apply:**
- 新增前端 API 调用时，grep 后端 routes.go 确认路由存在
- 后端新增路由后，curl 验证端点可达
- 代码评审时检查 catch 块是否为空
