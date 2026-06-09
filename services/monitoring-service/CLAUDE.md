# CLAUDE.md

## 角色定位

这是 **运行监控服务**（`github.com/guxiao1976/community-monitoring`），API-only，负责聚合检测系统各组件的健康状态。

- 微服务检测：TCP 拨号各服务 API/RPC 端口
- Docker 容器检测：`docker ps` 查询容器运行状态
- AI 模型检测：HTTP 调用 ai-model-service API

## 启动步骤

在开始任何工作前，必须先阅读 `docs/graph-context.md` 获取最新的上下文子图。
该文件由 Neo4j 知识图谱自动生成，包含本服务的接口、依赖、数据库表、前端消费方等完整信息。

## 关键规则

1. **无状态服务** — 不依赖数据库，不依赖其他微服务
2. **只做探测和聚合** — 不修改任何外部系统状态
3. **配置驱动** — 检测目标全在 YAML 中配置，增减服务无需改代码
4. **禁止修改 api-proto/** — 本服务无 Proto

## 全局公约

- 服务间通信仅通过 gRPC（本服务不调用其他服务，仅探测）
- 响应格式：`{code:0, msg:"success", data:<payload>}`
- 全局公约详见根 [CLAUDE.md](../../CLAUDE.md)

## 架构

```
monitoring-service/
  api/
    monitoring.go                  ← 入口
    etc/monitoring-api.yaml        ← 配置
    internal/
      config/config.go             ← Config 结构体
      svc/service_context.go       ← ServiceContext
      types/types.go               ← 响应类型
      handler/routes.go            ← 路由注册
      handler/health/              ← Health Handler
      logic/health/                ← 核心检测逻辑
```

## 常用命令

```bash
# 构建
cd api && go build -o monitoring .

# 运行（端口 8886）
cd api && go run monitoring.go -f etc/monitoring-api.yaml

# 测试
go test ./...
```

## 依赖

- `github.com/zeromicro/go-zero v1.10.1` — REST 框架
- `github.com/guxiao1976/community-common/v2` — configx, response
