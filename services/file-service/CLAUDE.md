# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 角色定位

这是 **文件服务**（`github.com/guxiao1976/community-file`），提供 MinIO 对象存储的统一文件上传、下载、删除和列表功能。**RPC + API 双层** 服务，gRPC 供其他业务服务调用，REST API 供前端直接调用。

## 启动步骤

在开始任何工作前，必须先阅读 `docs/graph-context.md` 获取最新的上下文子图。
该文件由 Neo4j 知识图谱自动生成，包含本服务的接口、依赖、数据库表、前端消费方等完整信息。

## 服务结构数据

本服务的 gRPC 接口、REST 路由、数据库表、依赖关系等结构化信息
由 Neo4j 知识图谱自动生成，详见 [docs/graph-context.md](docs/graph-context.md)。
**该文件是以上数据的唯一事实源，请勿在本文档中重复维护。**

## 关键规则

1. **禁止直接修改 `api-proto/`** — 如需修改 FileService 接口定义，告知用户切换到全局 Claude
2. **客户端直传模式** — 上传使用预签名 URL，客户端直传 MinIO，文件流不经过本服务
3. **MinIO 操作失败不阻塞 DB 操作** — 删除时优先保证 DB 元数据一致性，MinIO 操作失败仅记录日志
4. **复用 common/pkg/minio** — 下载/删除使用 common 封装，预签名上传使用原始 minio-go 客户端

## 全局公约

本项目所有微服务遵守统一的架构规范。与本服务相关的关键约束：

1. **Proto 在 api-proto/ 中统一管理** — 本服务的 gRPC 接口定义在 `api-proto/api/file/v1/`，修改 proto 需告知用户切换到全局 Claude
2. **服务间通信仅通过 gRPC** — 被其他服务调用走 gRPC（etcd 服务发现），REST API 仅用于前端上传下载
3. **全局公约详见根 [CLAUDE.md](../../CLAUDE.md)** — 包含 Proto 管理规范、go.work 联调规则等

## 常用命令

```bash
# 构建
go build ./...

# 测试
go test ./...

# 运行 RPC 服务
cd rpc && go run fileservice.go

# 运行 API 服务
cd api && go run file.go
```

## 架构

### 分层结构

```
file-service/
  api/                          # REST API 层（port 8884）
    file.go                     # 入口：package main
    etc/file-api.yaml           # 配置（JWT 密钥、etcd 发现 file.rpc）
    internal/
      config/config.go          # Config 结构体
      svc/servicecontext.go     # 依赖注入（gRPC 客户端）
      middleware/auth.go        # JWT user_id 提取
      handler/
        routes.go               # 路由注册（5 个路由，均 JWT 保护）
        file/handler.go         # 5 个 HTTP handler
      logic/file/               # 5 个 logic 文件（转调 gRPC）
      types/types.go            # 请求/响应类型
  rpc/                          # gRPC 服务（port 8085）
    fileservice.go              # 入口：package main
    file/fileservice.go         # 客户端代理封装
    etc/fileservice.yaml        # 配置（MinIO 凭证、etcd）
    internal/
      config/config.go          # 配置结构体
      server/fileserviceserver.go  # gRPC Server 实现
      svc/servicecontext.go     # 依赖注入（DB, MinIO, RawMinio）
      logic/file/               # 业务逻辑（5 个文件）
  model/                        # GORM 数据模型
    file.go                     # File 结构体
    filemodel.go                # FileModel 接口和实现
```

### 上传流程

```
客户端                    file-service              MinIO
  │                           │                      │
  │ ① GetUploadUrl() ────────▶│                      │
  │ ◀── upload_url, key ──────│                      │
  │                           │                      │
  │ ② PUT upload_url ──────────────────────────────▶│
  │ ◀── 200 OK ─────────────────────────────────────│
  │                           │                      │
  │ ③ ConfirmUpload(key) ────▶│                      │
  │                           │ ④ insert metadata    │
  │ ◀── FileInfo ──────────────│                      │
```


