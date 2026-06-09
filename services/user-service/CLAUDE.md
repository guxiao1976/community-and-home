# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 角色定位

这是 **用户服务**（`github.com/guxiao1976/community-user`），提供统一的用户资料管理。当前为 **RPC-only** 服务（api/ 层仅有 goctl 脚手架，未实现）。

## 启动步骤

在开始任何工作前，必须先阅读 `docs/graph-context.md` 获取最新的上下文子图。
该文件由 Neo4j 知识图谱自动生成，包含本服务的接口、依赖、数据库表、前端消费方等完整信息。

## 服务结构数据

本服务的 gRPC 接口、REST 路由、数据库表、依赖关系等结构化信息
由 Neo4j 知识图谱自动生成，详见 [docs/graph-context.md](docs/graph-context.md)。
**该文件是以上数据的唯一事实源，请勿在本文档中重复维护。**

## 关键规则

1. **禁止直接修改 `api-proto/`** — 如需修改 UserService 的 proto 接口，告知用户切换到全局 Claude
2. **禁止直接访问其他服务的数据库** — 与其他服务交互必须通过 gRPC
3. **本服务被 auth-service 调用** — 修改接口时需考虑对 auth-service 的兼容性
4. **手机号必须 AES 加密存储** — 使用 `common/model.EncryptedString` 类型
5. **写入操作前必须校验关联实体存在** — JoinCommunity/ApplyRole 前校验 user_id 存在（FindOne），ReviewCertification 前校验 reviewer_id 存在且未被禁用

## 全局公约

本项目所有微服务遵守统一的架构规范。与本服务相关的关键约束：

1. **Proto 在 api-proto/ 中统一管理** — 本服务的 gRPC 接口定义在 `api-proto/api/user/v1/`，修改 proto 需告知用户切换到全局 Claude
2. **服务间通信仅通过 gRPC** — 被 auth-service 等调用走 gRPC（etcd 服务发现），禁止直连其他数据库
3. **设计文档在 docs/design.md** — 数据库设计、业务流程、权限模型、ID 策略等见 [docs/design.md](docs/design.md)
4. **全局公约详见根 [CLAUDE.md](../../CLAUDE.md)** — 包含 Proto 管理规范、Snowflake ID 规范、go.work 联调规则等

## 设计文档

详细设计方案见 [docs/design.md](docs/design.md)（数据库设计、业务流程、权限模型、ID 策略、缓存设计等）。

## 常用命令

```bash
# 构建
go build ./...

# 测试
go test ./...

# 运行 RPC 服务
cd rpc && go run userservice.go

# 运行（指定配置文件）
cd rpc && go run userservice.go -f etc/userservice.yaml
```

## 架构

### 分层结构

```
rpc/                          # gRPC 服务（主要实现）
  userservice.go              # 入口：package main
  user/userservice.go         # 客户端代理封装（供其他服务调用）
  internal/
    config/config.go          # 配置结构体
    server/userserviceserver.go  # gRPC Server 实现
    svc/servicecontext.go     # 依赖注入（DB, Redis）
    logic/user/               # 业务逻辑（7 个文件，CRUD 操作）
model/                        # GORM 数据模型
api/                          # REST 层（仅有脚手架，未实现）
```

