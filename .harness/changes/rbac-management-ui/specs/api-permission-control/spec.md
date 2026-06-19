# Spec: 接口权限控制 (API Permission Control)

## 一、能力概述

后端 REST API 统一权限拦截，防止前端绕过菜单和按钮权限控制直接调用 API。提供权限中间件集成方案，支持外部请求统一拦截和内部 REST 二次校验。

---

## 二、功能需求

### 2.1 权限中间件

**功能描述**：HTTP 中间件，拦截所有 REST API 请求，调用 `permission-service.CheckPermission` 校验权限。

**实现位置**：各服务的 `internal/middleware/permissionmiddleware.go`

**中间件逻辑**：
```go
package middleware

import (
    "context"
    "net/http"
    "github.com/zeromicro/go-zero/rest"
    "github.com/zeromicro/go-zero/core/logx"
    permissionclient "community-common/v2/pkg/clients/permission"
    pb "api-proto/api/permission/v1"
)

func PermissionMiddleware(permClient permissionclient.Client) rest.Middleware {
    return func(next http.HandlerFunc) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
            // 1. 提取 user_id（从 JWT 中间件注入的 context）
            userId, ok := r.Context().Value("user_id").(int64)
            if !ok || userId == 0 {
                http.Error(w, `{"code":60003,"msg":"未登录或 Token 无效"}`, http.StatusUnauthorized)
                return
            }
            
            // 2. 调用 CheckPermission
            resp, err := permClient.CheckPermission(r.Context(), &pb.CheckPermReq{
                UserId:   userId,
                Action:   r.Method,
                Resource: r.URL.Path,
            })
            if err != nil {
                logx.Errorf("CheckPermission error: %v", err)
                http.Error(w, `{"code":60004,"msg":"权限检查失败"}`, http.StatusInternalServerError)
                return
            }
            
            // 3. 权限校验失败
            if !resp.Allowed {
                logx.Infof("Permission denied: user=%d, path=%s", userId, r.URL.Path)
                http.Error(w, `{"code":60003,"msg":"无权限访问"}`, http.StatusForbidden)
                return
            }
            
            // 4. 权限通过，继续执行
            next(w, r)
        }
    }
}
```

**使用示例**：
```go
// internal/handler/routes.go
func RegisterHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
    // 全局中间件
    server.Use(middleware.JWTMiddleware(serverCtx.Config.Auth.AccessSecret))
    server.Use(middleware.PermissionMiddleware(serverCtx.PermissionClient))
    
    // 注册路由
    server.AddRoutes(
        rest.WithMiddlewares(
            []rest.Middleware{},
            []rest.Route{
                {Method: http.MethodGet, Path: "/api/v1/user/users", Handler: user.ListUsersHandler(serverCtx)},
                {Method: http.MethodPost, Path: "/api/v1/user/users", Handler: user.CreateUserHandler(serverCtx)},
            }...,
        ),
        rest.WithPrefix("/api/v1/user"),
    )
}
```

---

### 2.2 白名单机制

**功能描述**：部分 API 无需权限校验（如登录、注册、公开接口）。

**实现方式**：中间件跳过白名单路径。

**白名单配置**：
```go
var whiteList = map[string]bool{
    "/api/v1/auth/login":    true,
    "/api/v1/auth/register": true,
    "/api/v1/auth/refresh":  true,
    "/api/v1/health":        true,
}

func PermissionMiddleware(permClient permissionclient.Client) rest.Middleware {
    return func(next http.HandlerFunc) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
            // 白名单直接放行
            if whiteList[r.URL.Path] {
                next(w, r)
                return
            }
            
            // 其他逻辑...
        }
    }
}
```

**配置化白名单**（可选增强）：
```yaml
# etc/userservice.yaml
PermissionWhitelist:
  - /api/v1/auth/login
  - /api/v1/auth/register
  - /api/v1/health
```

---

### 2.3 APISIX Gateway 统一拦截

**功能描述**：在 API Gateway 层统一拦截外部请求，调用权限服务。

**实现方式**：APISIX 插件（OpenResty Lua）

**插件逻辑**（伪代码）：
```lua
-- apisix/plugins/permission.lua
local http = require "resty.http"
local cjson = require "cjson"

function _M.access(conf, ctx)
    -- 1. 提取 JWT 中的 user_id
    local user_id = ctx.var.jwt_claim_user_id
    if not user_id then
        return 401, {code = 60003, msg = "未登录"}
    end
    
    -- 2. 调用 permission-service.CheckPermission
    local httpc = http.new()
    local res, err = httpc:request_uri("http://permission-service:8084/CheckPermission", {
        method = "POST",
        body = cjson.encode({
            user_id = user_id,
            action = ngx.var.request_method,
            resource = ngx.var.uri
        }),
        headers = {["Content-Type"] = "application/json"}
    })
    
    if not res or res.status ~= 200 then
        return 500, {code = 60004, msg = "权限检查失败"}
    end
    
    local data = cjson.decode(res.body)
    if not data.allowed then
        return 403, {code = 60003, msg = "无权限访问"}
    end
    
    -- 3. 权限通过，继续转发
end
```

**APISIX 路由配置**：
```yaml
# apisix/routes.yaml
routes:
  - uri: /api/v1/*
    plugins:
      jwt-auth: {}
      permission: {}  # 自定义权限插件
    upstream:
      nodes:
        user-service:8080: 1
        permission-service:8080: 1
```

**注意**：APISIX 配置属于基础设施层，**本次需求不实现**，仅在设计中预留接口。

---

### 2.4 内部 gRPC 权限策略

**策略**：信任域内调用，不重复校验。

**理由**：
- gRPC 调用发生在服务间，调用方已通过 REST 权限校验
- 避免性能损耗（每次 gRPC 调用重复校验）
- 简化实现（gRPC 拦截器复杂度高）

**边界**：
- 外部请求 → REST API → 权限中间件校验 → 调用内部 gRPC（不校验）
- 内部 gRPC 调用 → 信任域，不校验

**例外**：
- 如果服务直接暴露 gRPC 端口到外网（不推荐），需要 gRPC 拦截器校验
- 本项目所有 gRPC 服务仅在内网暴露，外部请求统一经过 REST API Gateway

---

## 三、数据模型

### 3.1 CheckPermission 请求

```protobuf
// api-proto/api/permission/v1/permission.proto

message CheckPermReq {
  int64 user_id = 1 [jstype = JS_STRING];
  string action = 2;       // HTTP 方法（GET / POST / PUT / DELETE）
  string resource = 3;     // API 路径（如 /api/v1/user/users）
}

message CheckPermResp {
  common.v1.BaseResp base_resp = 1;
  bool allowed = 2;        // true=允许访问 false=拒绝访问
}
```

### 3.2 权限匹配规则

**sys_permission 表字段**：
- `path`: `/api/v1/user/users`（API 路径）
- `code`: `user:read:api`（权限码）
- `type`: 3（API 权限）

**匹配逻辑**：
1. 提取请求的 `action` 和 `resource`（如 `GET /api/v1/user/users`）
2. 查询 `sys_permission` 表，匹配 `path` 字段
3. 获取该 API 对应的权限码（如 `user:read:api`）
4. 检查用户是否拥有该权限码（从 Redis 缓存或 DB 查询）

**路径匹配策略**：
- 精确匹配：`/api/v1/user/users` 匹配 `/api/v1/user/users`
- 动态参数：`/api/v1/user/users/:id` 匹配 `/api/v1/user/users/123`（使用正则或路径解析）

**实现示例**：
```go
func (l *CheckPermissionLogic) CheckPermission(in *pb.CheckPermReq) (*pb.CheckPermResp, error) {
    // 1. 查询用户所有角色
    userRoles, err := l.svcCtx.UserRoleModel.FindActiveByUserId(ctx, in.UserId)
    if err != nil {
        return nil, err
    }
    
    // 2. 系统角色直接放行
    for _, ur := range userRoles {
        role, _ := l.svcCtx.RoleModel.FindOne(ctx, ur.RoleId)
        if role.IsSystem == 1 {
            return &pb.CheckPermResp{
                BaseResp: &pb.BaseResp{Code: 0, Msg: "success"},
                Allowed:  true,
            }, nil
        }
    }
    
    // 3. 构建权限码集合（Redis 缓存）
    cacheKey := fmt.Sprintf("perm:user:%d", in.UserId)
    permissionCodes, err := l.svcCtx.Redis.SMembers(ctx, cacheKey).Result()
    if err != nil || len(permissionCodes) == 0 {
        // 缓存未命中，从 DB 查询
        permissionCodes = l.fetchUserPermissionsFromDB(in.UserId)
        // 写入缓存
        l.svcCtx.Redis.SAdd(ctx, cacheKey, permissionCodes)
        l.svcCtx.Redis.Expire(ctx, cacheKey, 30*time.Minute)
    }
    
    // 4. 查询 API 对应的权限码
    permission, err := l.svcCtx.PermissionModel.FindByPath(ctx, in.Resource)
    if err != nil {
        // API 未配置权限，默认拒绝
        return &pb.CheckPermResp{
            BaseResp: &pb.BaseResp{Code: 0, Msg: "success"},
            Allowed:  false,
        }, nil
    }
    
    // 5. 检查用户是否拥有该权限码
    for _, code := range permissionCodes {
        if code == permission.Code {
            return &pb.CheckPermResp{
                BaseResp: &pb.BaseResp{Code: 0, Msg: "success"},
                Allowed:  true,
            }, nil
        }
    }
    
    // 6. 无权限
    return &pb.CheckPermResp{
        BaseResp: &pb.BaseResp{Code: 0, Msg: "success"},
        Allowed:  false,
    }, nil
}
```

---

## 四、接口清单

### 4.1 后端 gRPC 接口

**服务**：`permission-service`

| RPC | 请求 | 响应 | 说明 |
|-----|------|------|------|
| `CheckPermission` | `CheckPermReq` | `CheckPermResp` | 已存在，校验用户是否有权访问指定 API |

**性能指标**：
- P99 < 100ms（依赖 Redis 缓存）
- QPS > 1000（单实例）
- 缓存命中率 > 95%

---

## 五、业务规则

### 5.1 权限校验流程

```
请求 → JWT 中间件（提取 user_id） → 权限中间件
                                          │
                    ┌─────────────────────┴──────────────────────┐
                    │                                             │
                白名单？                                     CheckPermission(user_id, action, resource)
                    │                                             │
                  Yes                                      ┌──────┴──────┐
                    │                                      │             │
                  放行                                  Allowed      Denied
                                                           │             │
                                                         放行         403
```

### 5.2 系统角色特权

- 系统角色（`is_system=1`）在 `CheckPermission` 中直接返回 `allowed=true`
- 不查询权限表，不匹配权限码
- 天然拥有全部 API 访问权限

### 5.3 缓存策略

**Redis 缓存**：
- Key：`perm:user:{user_id}`
- Type：Set
- Value：用户所有权限码（如 `["user:read:api", "user:create:api"]`）
- TTL：30 分钟

**缓存失效**：
- `AssignRole` / `RevokeRole`：失效单个用户缓存（`DEL perm:user:{user_id}`）
- `UpdateRole`：失效所有用户缓存（`KEYS perm:user:*` + 批量 `DEL`）

**缓存重建**：
- `CheckPermission` 缓存未命中时，从 DB 查询用户所有角色的权限码
- 写入 Redis Set，设置 TTL 30 分钟

### 5.4 API 未配置权限

**场景**：新增 API 接口，但未在 `sys_permission` 表中配置权限。

**策略**：默认拒绝访问（`allowed=false`）。

**理由**：安全优先，防止权限遗漏导致的安全风险。

**例外**：白名单接口跳过权限校验。

### 5.5 动态路径匹配

**场景**：`/api/v1/user/users/:id` 匹配 `/api/v1/user/users/123`。

**实现方式**：
- 方案 A：权限表 `path` 字段存储正则表达式（如 `^/api/v1/user/users/\d+$`）
- 方案 B：权限表 `path` 字段存储路径模板（如 `/api/v1/user/users/:id`），匹配时去掉参数

**推荐**：方案 B（路径模板），简单直观。

**实现示例**：
```go
// 去掉路径中的动态参数
func normalizePath(path string) string {
    // /api/v1/user/users/123 → /api/v1/user/users/:id
    // 简化实现：正则替换数字为 :id
    re := regexp.MustCompile(`/\d+`)
    return re.ReplaceAllString(path, "/:id")
}

// CheckPermission 中使用
normalizedPath := normalizePath(in.Resource)
permission, err := l.svcCtx.PermissionModel.FindByPath(ctx, normalizedPath)
```

---

## 六、中间件集成方案

### 6.1 集成步骤

**Step 1**：在服务 `svc.ServiceContext` 中注入 `PermissionClient`

```go
// internal/svc/servicecontext.go
type ServiceContext struct {
    Config           config.Config
    PermissionClient permissionclient.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
    permConn := zrpc.MustNewClient(c.PermissionRpc)
    return &ServiceContext{
        Config:           c,
        PermissionClient: permissionclient.NewClient(permConn),
    }
}
```

**Step 2**：在配置文件中添加 `PermissionRpc`

```yaml
# etc/userservice.yaml
Name: user.api
Host: 0.0.0.0
Port: 8080

Auth:
  AccessSecret: your-secret-key

PermissionRpc:
  Etcd:
    Hosts:
      - 127.0.0.1:2379
    Key: permission.rpc
```

**Step 3**：注册权限中间件

```go
// internal/handler/routes.go
func RegisterHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
    // 全局中间件
    server.Use(middleware.JWTMiddleware(serverCtx.Config.Auth.AccessSecret))
    server.Use(middleware.PermissionMiddleware(serverCtx.PermissionClient))
    
    // 注册路由
    // ...
}
```

### 6.2 分阶段推进

**P0**（示例服务）：
- `user-service`：集成权限中间件，验证可行性

**P1**（核心服务）：
- `auth-service`：登录/注册接口在白名单
- `permission-service`：角色/权限管理接口需要权限校验

**P2**（其他服务）：
- `moderation-service`、`community-hub-service` 等逐步集成

**P3**（全覆盖）：
- 所有服务集成完成，形成统一权限体系

---

## 七、错误处理

### 7.1 错误码定义

| HTTP 状态码 | 业务错误码 | 说明 |
|-----------|----------|------|
| 401 | 60003 | 未登录或 Token 无效 |
| 403 | 60003 | 无权限访问 |
| 500 | 60004 | 权限检查失败（服务异常） |

### 7.2 错误响应格式

```json
{
  "code": 60003,
  "msg": "无权限访问",
  "data": null
}
```

### 7.3 权限校验失败审计

**日志记录**：
```go
logx.Infof("Permission denied: user=%d, path=%s, method=%s", userId, r.URL.Path, r.Method)
```

**审计日志字段**：
- `user_id`：用户ID
- `path`：请求路径
- `method`：HTTP 方法
- `timestamp`：时间戳
- `ip`：客户端 IP

**用途**：
- 安全审计：发现恶意访问尝试
- 权限调试：排查权限配置错误

---

## 八、性能优化

### 8.1 Redis 缓存

**缓存内容**：
- Key：`perm:user:{user_id}`
- Value：用户所有权限码（Set）
- TTL：30 分钟

**优化效果**：
- 缓存命中率 > 95%
- 权限检查耗时从 50ms 降至 < 5ms

### 8.2 批量失效优化

**问题**：`UpdateRole` 使用 `KEYS perm:user:*` 全量扫描，生产环境可能阻塞 Redis。

**优化方案**：
- 方案 A：使用 `SCAN` 替代 `KEYS`（游标迭代，不阻塞）
- 方案 B：维护反向索引（`role:users:{role_id}` → Set of user_id），精确失效
- 方案 C：延长 TTL 至 5 分钟，减少失效频率

**推荐**：方案 A（短期）+ 方案 B（长期）

### 8.3 权限树预加载

**问题**：每次 `CheckPermission` 都查询 `sys_permission` 表。

**优化**：
- 服务启动时，预加载所有 API 权限到内存（Map）
- Key：`path`（如 `/api/v1/user/users`）
- Value：`PermissionInfo`（含 `code`）

**实现示例**：
```go
type PermissionCache struct {
    mu          sync.RWMutex
    permissions map[string]*model.SysPermission
}

func (c *PermissionCache) Load() {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    perms, _ := l.svcCtx.PermissionModel.FindAll(ctx)
    c.permissions = make(map[string]*model.SysPermission, len(perms))
    for _, p := range perms {
        c.permissions[p.Path] = p
    }
}

func (c *PermissionCache) Get(path string) *model.SysPermission {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.permissions[path]
}
```

---

## 九、测试场景

### 9.1 单元测试

**中间件测试**：
- 白名单路径跳过校验
- 无 Token 请求返回 401
- 无权限请求返回 403
- 有权限请求放行

**CheckPermission 测试**：
- 系统角色直接放行
- 缓存命中返回正确结果
- 缓存未命中查 DB 并重建缓存
- 动态路径匹配正确

### 9.2 集成测试

- E2E 流程：登录 → 获取 Token → 调用有权限 API（200）→ 调用无权限 API（403）
- 权限变更：撤销角色 → 缓存失效 → 下次请求重建缓存 → API 请求被拒绝
- 系统角色：owner 角色访问任意 API，全部放行

### 9.3 性能测试

- 权限检查 P99 < 100ms（100 并发）
- 缓存命中率 > 95%（1000 次请求）
- Redis 批量失效耗时 < 500ms（1000 个用户）

### 9.4 安全测试

- 尝试伪造 Token（无效签名）→ 401
- 尝试访问无权限 API → 403
- 尝试绕过中间件（直接调用 Handler）→ 403（中间件在 Server 层注册）

---

## 十、依赖与约束

### 10.1 依赖

| 依赖 | 说明 |
|------|------|
| `permission-service` | 提供 `CheckPermission` 接口 |
| `auth-service` | 提供 JWT 中间件（注入 `user_id`） |
| Redis | 权限缓存 |
| etcd | 服务发现（gRPC 调用） |

### 10.2 约束

- 权限中间件必须在 JWT 中间件之后注册（依赖 `user_id`）
- 内部 gRPC 调用不校验权限（信任域）
- API 未配置权限时默认拒绝访问（安全优先）

---

## 十一、追溯

### 11.1 需求来源

- 用户原话："权限管理要实现接口级的管理，不能仅仅是前端是否展示相关按钮或者内容，这样知道api时，容易绕过去"
- Brainstorming 决策点 4：接口级权限控制 → 选择"外部请求统一拦截 + 内部 REST 二次校验"

### 11.2 关联 Spec

- `menu-permission-control/spec.md`：前端权限控制（可被绕过，本 Spec 兜底）
- `permission-management/spec.md`：依赖权限树数据（`sys_permission` 表）
- `role-management/spec.md`：依赖角色数据

### 11.3 记忆触发

- [[grpc-only-comms]]：服务间通信仅通过 gRPC
- [[pre-commit-checks]]：提交前运行 `harness-checks.sh`
