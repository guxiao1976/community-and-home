# 正确模式：跨服务调用

## 核心原则

**服务间通信必须通过 gRPC，禁止直接访问其他服务的数据库。**

## 为什么

1. **服务边界** — 直接 DB 访问破坏服务封装，造成紧耦合
2. **演进能力** — 数据库 schema 变更会破坏所有直接访问的服务
3. **安全性** — 绕过服务层的权限控制和审计
4. **可测试性** — 无法 mock 其他服务的行为

## ❌ 错误模式

### 场景：Auth Service 需要获取用户信息

```go
// services/auth-service/api/internal/logic/login_logic.go

import (
    "github.com/guxiao1976/community-user/model"  // ❌ 跨服务导入
)

type LoginLogic struct {
    ctx    context.Context
    svcCtx *svc.ServiceContext
}

func (l *LoginLogic) Login(req *types.LoginRequest) (*types.LoginResponse, error) {
    // ❌ 直接访问 user-service 的数据库
    user, err := l.svcCtx.UserModel.FindOneByMobile(l.ctx, req.Mobile)
    if err != nil {
        return nil, err
    }
    
    // 业务逻辑...
    return &types.LoginResponse{...}, nil
}
```

**问题**：
- 导入了 `community-user/model` 包
- 直接调用 `UserModel`（GORM/sqlx 模型）
- 绕过了 user-service 的业务逻辑层

## ✅ 正确模式

### 步骤 1：确保 svcCtx 中有 RPC 客户端

```go
// services/auth-service/api/internal/svc/service_context.go

import (
    "github.com/zeromicro/go-zero/zrpc"
    "github.com/guxiao1976/community-user/userclient"  // ✅ 导入 RPC 客户端
)

type ServiceContext struct {
    Config   config.Config
    UserRpc  userclient.User  // ✅ RPC 客户端字段
}

func NewServiceContext(c config.Config) *ServiceContext {
    return &ServiceContext{
        Config:  c,
        UserRpc: userclient.NewUser(zrpc.MustNewClient(c.UserRpc)),  // ✅ 初始化 RPC 客户端
    }
}
```

### 步骤 2：配置文件中添加 RPC 配置

```yaml
# services/auth-service/api/etc/auth-api.yaml

Name: auth-api
Host: 0.0.0.0
Port: 8081

# ✅ RPC 客户端配置
UserRpc:
  Etcd:
    Hosts:
      - 127.0.0.1:2379
    Key: user.rpc
```

### 步骤 3：Config 结构体中添加字段

```go
// services/auth-service/api/internal/config/config.go

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
    rest.RestConf
    UserRpc zrpc.RpcClientConf  // ✅ RPC 配置字段
}
```

### 步骤 4：Logic 中调用 RPC

```go
// services/auth-service/api/internal/logic/login_logic.go

import (
    "github.com/guxiao1976/community-user/userclient"  // ✅ 导入 RPC 客户端
    userpb "github.com/guxiao/community-and-home/api-proto/gen/go/user/v1"  // ✅ 导入 Proto
)

func (l *LoginLogic) Login(req *types.LoginRequest) (*types.LoginResponse, error) {
    // ✅ 通过 RPC 调用 user-service
    userResp, err := l.svcCtx.UserRpc.GetUserByMobile(l.ctx, &userpb.GetUserByMobileReq{
        Mobile: req.Mobile,
    })
    if err != nil {
        return nil, errx.Wrap(err, "调用 UserRpc.GetUserByMobile 失败")
    }
    
    // ✅ 使用 RPC 响应中的数据
    if !verifyPassword(req.Password, userResp.User.Password) {
        return nil, errx.NewCodeError(40001, "密码错误")
    }
    
    // 生成 Token...
    return &types.LoginResponse{
        Token:  token,
        UserId: userResp.User.UserId,
    }, nil
}
```

## 完整示例参考

### 参考实现 1：Auth Service 调用 User Service

**文件**: `services/auth-service/api/internal/logic/verify_token_logic.go:28-35`

```go
func (l *VerifyTokenLogic) VerifyToken(req *types.VerifyTokenRequest) (*types.VerifyTokenResponse, error) {
    // 解析 Token 获取 userId
    claims, err := l.parseToken(req.Token)
    if err != nil {
        return nil, errx.Wrap(err, "Token 解析失败")
    }
    
    // ✅ 通过 RPC 验证用户状态
    userResp, err := l.svcCtx.UserRpc.GetUserInfo(l.ctx, &userpb.GetUserInfoReq{
        UserId: claims.UserId,
    })
    if err != nil {
        return nil, errx.Wrap(err, "获取用户信息失败")
    }
    
    if userResp.User.Status != 1 {
        return nil, errx.NewCodeError(40003, "用户已被禁用")
    }
    
    return &types.VerifyTokenResponse{Valid: true}, nil
}
```

### 参考实现 2：Moderation Service 调用 Community Hub Service

**文件**: `services/moderation-service/rpc/internal/logic/review_logic.go:42-50`

```go
func (l *ReviewContentLogic) ReviewContent(in *pb.ReviewContentReq) (*pb.ReviewContentResp, error) {
    // 审核逻辑...
    result := l.moderateText(in.Content)
    
    // ✅ 审核通过后，通过 RPC 通知 community-hub-service
    if result.Approved {
        _, err := l.svcCtx.CommunityHubRpc.ApproveContent(l.ctx, &hubpb.ApproveContentReq{
            ContentId:   in.ContentId,
            ContentType: in.ContentType,
        })
        if err != nil {
            return nil, errx.Wrap(err, "调用 CommunityHubRpc.ApproveContent 失败")
        }
    }
    
    return &pb.ReviewContentResp{Approved: result.Approved}, nil
}
```

## RPC vs. DB 对比表

| 维度 | 直接 DB 访问 | RPC 调用 |
|------|------------|---------|
| 服务边界 | ❌ 破坏封装 | ✅ 尊重边界 |
| 业务逻辑 | ❌ 绕过服务层 | ✅ 经过服务层 |
| 权限控制 | ❌ 绕过 | ✅ 由服务层控制 |
| 演进能力 | ❌ schema 变更破坏调用方 | ✅ Proto 兼容性检查 |
| 可测试性 | ❌ 需要真实 DB | ✅ 可 mock RPC |
| 性能 | ⚡ 更快（无网络开销） | 🐢 稍慢（网络开销） |
| 推荐场景 | 服务内部 | 服务间通信 |

## 常见问题

### Q1: 如果 user-service 没有提供我需要的 RPC 接口怎么办？

**A**: 在 `api-proto/api/user/v1/user.proto` 中添加新的 RPC 方法，然后：

1. `cd api-proto && make generate` — 生成代码
2. 在 `user-service/rpc/internal/logic/` 中实现新方法
3. 在调用方使用新的 RPC 方法

### Q2: 性能考虑？RPC 会不会太慢？

**A**: 
- RPC 调用的网络开销通常在 1-5ms（局域网）
- 如果性能敏感，考虑在调用方增加缓存（Redis）
- **禁止为了性能直接访问数据库** — 架构约束优先于性能优化

### Q3: 如果是批量查询（N+1 问题）怎么办？

**A**: 在 Proto 中设计批量接口：

```protobuf
rpc BatchGetUsers(BatchGetUsersReq) returns (BatchGetUsersResp);

message BatchGetUsersReq {
  repeated int64 user_ids = 1 [(gogoproto.jstype) = JS_STRING];
}
```

## 相关文档

- [项目编码规范 §1 — 服务间通信](./../rules/项目编码规范.md#1-服务间通信)
- [工程结构 — 服务分层](../../rules/工程结构.md)
- [Proto 管理规范](../../rules/Proto管理规范.md)
- [Memory: grpc-only-comms](../../knowledge/memory/grpc-only-comms.md)

## 检查清单

修复跨服务 DB 导入违规时，确认以下步骤：

- [ ] 移除跨服务的 `model` 包导入
- [ ] 在 `svc/service_context.go` 中添加 RPC 客户端字段
- [ ] 在 `config/config.go` 中添加 RPC 配置字段
- [ ] 在 YAML 配置文件中添加 RPC 配置（Etcd Key）
- [ ] 在 Logic 中调用 RPC 方法替代直接 DB 访问
- [ ] 处理 RPC 错误（使用 `errx.Wrap`）
- [ ] 运行 `bash .harness/skills/qa/scripts/harness-checks.sh --service <name>` 验证
