# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 角色定位

这是 **权限服务**（`github.com/guxiao1976/community-permission`），基于 RBAC 的权限和数据范围引擎。**API+RPC 双层**服务。

## 启动步骤

在开始任何工作前，必须先阅读 `docs/graph-context.md` 获取最新的上下文子图。
该文件由 Neo4j 知识图谱自动生成，包含本服务的接口、依赖、数据库表、前端消费方等完整信息。

## 服务结构数据

本服务的 gRPC 接口、REST 路由、数据库表、依赖关系等结构化信息
由 Neo4j 知识图谱自动生成，详见 [docs/graph-context.md](docs/graph-context.md)。
**该文件是以上数据的唯一事实源，请勿在本文档中重复维护。**

## 关键规则

1. **禁止直接修改 `api-proto/`** — 如需修改 PermissionService 的 proto 接口，告知用户切换到全局 Claude
2. **禁止直接访问其他服务的数据库**
3. **权限缓存一致性**：Redis 中缓存的权限数据需与 MySQL 保持一致，修改角色/权限时必须刷新缓存
4. **数据范围安全**：`GetDataScopes` 返回的 scope_id 列表用于 GORM 拦截器的 WHERE IN 过滤

## 全局公约

本项目所有微服务遵守统一的架构规范。与本服务相关的关键约束：

1. **Proto 在 api-proto/ 中统一管理** — 本服务的 gRPC 接口定义在 `api-proto/api/permission/v1/`，修改 proto 需告知用户切换到全局 Claude
2. **服务间通信仅通过 gRPC** — 被其他服务调用走 gRPC（etcd 服务发现），禁止直连其他数据库
3. **全局公约详见根 [CLAUDE.md](../../CLAUDE.md)** — 包含 Proto 管理规范、go.work 联调规则等

## 常用命令

```bash
# 构建
go build ./...

# 测试
go test ./...

# 运行 RPC 服务（gRPC :8084）
cd rpc && go run permissionservice.go

# 运行 API 服务（REST :8883）
cd api && go run perm.go
```

## 架构

### 分层结构

```
api/                               # REST 网关（go-zero rest.MustNewServer）
  perm.go                          # 入口：package main，端口 :8883
  internal/
    config/config.go               # 配置结构体
    handler/routes.go              # 路由注册（/api/perm/*），JWT 认证
    logic/perm/                    # 业务逻辑（代理 gRPC 调用）
    svc/servicecontext.go          # 依赖注入（gRPC 客户端）
    types/types.go                 # HTTP 请求/响应类型

rpc/                               # gRPC 服务（go-zero zrpc.MustNewServer）
  permissionservice.go             # 入口：package main，端口 :8084
  permission/permissionservice.go  # 客户端代理封装
  internal/
    config/config.go               # 配置结构体
    server/permissionserviceserver.go  # gRPC Server 实现
    svc/servicecontext.go          # 依赖注入（DB, Redis, RBAC 模型）
    logic/                         # 业务逻辑

model/                             # GORM 数据模型
```

### 核心概念

- **Role**（角色）：owner、property_admin、community_admin、grid_worker，支持 `is_system` 标记
- **Permission**（权限）：API 级别，类型 menu/button/api
- **UserRole**（用户角色）：关联 user_id + role_id + scope_type + scope_id
- **DataScope**（数据范围）：控制用户可见的数据边界


