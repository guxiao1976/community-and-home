# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 角色定位

这是 **认证服务**（`github.com/guxiao1976/community-auth`），社区平台的认证中心。**RPC-only** 服务（无 REST 层），提供 AT（Access Token）+ RT（Refresh Token）双 Token 认证机制。

## 启动步骤

在开始任何工作前，必须先阅读 `docs/graph-context.md` 获取最新的上下文子图。
该文件由 Neo4j 知识图谱自动生成，包含本服务的接口、依赖、数据库表、前端消费方等完整信息。

## 服务结构数据

本服务的 gRPC 接口、REST 路由、数据库表、依赖关系等结构化信息
由 Neo4j 知识图谱自动生成，详见 [docs/graph-context.md](docs/graph-context.md)。
**该文件是以上数据的唯一事实源，请勿在本文档中重复维护。**

## 关键规则

1. **禁止直接修改 `api-proto/`** — 如需修改 AuthService 的 proto 接口，告知用户切换到全局 Claude
2. **禁止直接访问其他服务的数据库** — 获取/操作用户数据必须通过 user-service 的 gRPC 接口
3. **Token 安全**：RT 使用 Redis 持久化、轮换时删除旧 RT、注销时 AT 加入黑名单（黑名单写入失败必须返回错误，不可静默忽略）。**RefreshToken 必须先拉取角色再旋转 RT**，防止角色拉取失败时旧 RT 已被销毁导致用户被迫重新登录
4. **密码安全**：密码使用 Bcrypt 哈希，传输中使用 RSA 加密

## 设计文档

详细设计方案见 [docs/design.md](docs/design.md)（JWT 设计、认证流程、鉴权路径、角色缓存、数据库、安全机制等）。

## 全局公约

本项目所有微服务遵守统一的架构规范。与本服务相关的关键约束：

1. **Proto 在 api-proto/ 中统一管理** — 本服务的 gRPC 接口定义在 `api-proto/api/auth/v1/`，修改 proto 需告知用户切换到全局 Claude
2. **服务间通信仅通过 gRPC** — 调用 user-service 走 gRPC（etcd 服务发现），禁止直连 user 数据库
3. **全局公约详见根 [CLAUDE.md](../../CLAUDE.md)** — 包含 Proto 管理规范、Snowflake ID 规范、go.work 联调规则、错误码规范等

## 常用命令

```bash
# 构建
go build ./...

# 测试
go test ./...

# 运行 RPC 服务
cd rpc && go run authservice.go
```

## 架构

### 分层结构

```
rpc/                          # gRPC 服务（唯一入口）
  authservice.go              # 入口：package main（初始化 AES+RSA）
  auth/authservice.go         # 客户端代理封装
  internal/
    config/config.go          # 配置结构体（含 JWT Secret、RSA/AES key）
    server/authserviceserver.go  # gRPC Server 实现
    svc/servicecontext.go     # 依赖注入（DB, Redis, UserServiceRpc 客户端）
    logic/                    # 业务逻辑
```

### Token 机制要点

- **AT**（Access Token）：JWT，15 分钟过期，payload 含 `user_id, jti, exp`
- **RT**（Refresh Token）：JWT，15 天过期，payload 含 `jti, user_id, device_id`
- **RT 轮换**：刷新时删除旧 RT 并发新 RT，防止 RT 泄露后被复用
- **AT 黑名单**：注销时 AT 加入 Redis 黑名单直到过期
- **强制所有设备下线**：Logout 时可选清除所有 RT
