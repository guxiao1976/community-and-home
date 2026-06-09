# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 角色定位

这是 **社区枢纽服务**（`github.com/guxiao1976/community-hub`），社区平台的小区信息汇聚中心。提供 **RPC + REST API 双层** 服务，负责通知公告发布、便民联络信息维护、寻失互助等社区内容场景。

## 启动步骤

在开始任何工作前，必须先阅读 `docs/graph-context.md` 获取最新的上下文子图。
该文件由 Neo4j 知识图谱自动生成，包含本服务的接口、依赖、数据库表、前端消费方等完整信息。

## 服务结构数据

本服务的 gRPC 接口、REST 路由、数据库表、依赖关系等结构化信息
由 Neo4j 知识图谱自动生成，详见 [docs/graph-context.md](docs/graph-context.md)。
**该文件是以上数据的唯一事实源，请勿在本文档中重复维护。**

## 关键规则

1. **禁止直接修改 `api-proto/`** — 如需修改 Community Hub 的 proto 接口，告知用户切换到全局 Claude
2. **禁止直接访问其他服务的数据库** — 与其他服务交互必须通过 gRPC
3. **所有 int64 ID 字段使用 `json:",string"` 标签** — 确保 JS 端 Snowflake ID 精度不丢失
4. **RPC 响应第一个字段必须是 `BaseResp base`**
5. **业务错误码范围 08xxxx** — 见下方错误码表

## 全局公约

本项目所有微服务遵守统一的架构规范。与本服务相关的关键约束：

1. **Proto 在 api-proto/ 中统一管理** — 本服务的 gRPC 接口定义在 `api-proto/api/community/v1/`，修改 proto 需告知用户切换到全局 Claude
2. **服务间通信仅通过 gRPC** — 禁止直连其他服务数据库
3. **设计文档在 docs/design.md** — 数据库设计、业务流程、API 设计等
4. **全局公约详见根 [CLAUDE.md](../../CLAUDE.md)** — 包含 Proto 管理规范、Snowflake ID 规范、go.work 联调规则、错误码规范等

## 常用命令

```bash
# 构建
go build ./...

# 测试
go test ./...

# 运行 RPC 服务（端口 8087）
cd rpc && go run communityhub.go -f etc/communityhub.yaml

# 运行 REST API（端口 8887）
cd api && go run communityhub.go -f etc/communityhub-api.yaml
```

## 架构

### 分层结构

```
rpc/                          # gRPC 服务
  communityhub.go             # 入口：package main（注册 3 个 Service）
  internal/
    config/config.go          # 配置结构体（RpcServerConf + DataSource）
    server/communityhubserver.go  # gRPC Server 实现（3 个 Service）
    svc/servicecontext.go     # 依赖注入（4 个 DB Model）
    logic/
      notice/                 # 通知公告业务逻辑（6个文件）
      contact/                # 便民联络业务逻辑（2个文件 + helper）
      lostfound/              # 寻失互助业务逻辑（4个文件 + helper）

api/                          # REST API 网关
  communityhub.go             # 入口：package main
  internal/
    config/config.go          # 配置结构体（RestConf + CommunityHubRpc）
    svc/servicecontext.go     # 依赖注入（3 个 gRPC Client）
    handler/routes.go         # HTTP 路由注册（/api/community/*）
    handler/notice/           # 通知 Handler
    handler/contact/          # 联络 Handler
    handler/lostfound/        # 寻失 Handler
    logic/                    # 业务逻辑：类型转换 + gRPC 代理
    types/types.go            # HTTP 请求/响应类型

model/                        # 数据模型（go-zero sqlx 风格）
  notice.go                   # 通知公告
  notice_attachment.go        # 通知附件
  community_contact.go        # 便民联络
  lost_found_item.go          # 寻失互助
```

### 错误码

范围 08xxxx（社区枢纽服务）：

| 错误码 | 常量 | 说明 |
|--------|------|------|
| 080001 | 通知不存在 | 查询/更新/删除不存在的通知 |
| 080002 | 无发布权限 | 未认证对应社区角色 |
| 080003 | 寻失发布次数已达上限 | 超过年度发布限制 |
| 080004 | 联络/寻失记录不存在 | 查询不存在的记录 |
| 080005 | 参数无效 | 参数校验失败 |


