# 设计文档 — monitoring-service

## 1. 服务定位

运行监控服务，API-only，负责聚合检测系统各组件的健康状态。无状态、无数据库依赖、不修改外部系统状态。

## 2. 数据模型

本服务无数据库。所有检测结果在内存中计算，每次请求实时探测。

| 概念 | 说明 |
|------|------|
| 微服务健康 | TCP 拨号各服务 API/RPC 端口，超时 = 不健康 |
| Docker 容器 | `docker ps` 查询容器运行状态，exited = 不健康 |
| AI 模型 | HTTP 调用 ai-model-service API，HTTP 200 = 健康 |

## 3. API 接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/monitoring/health` | 返回所有组件的健康状态聚合 |

响应格式：`{code:0, msg:"success", data:{services:[...], containers:[...], ai_models:[...]}}`

## 4. 配置

```yaml
# services/monitoring-service/api/etc/monitoring.yaml
Name: monitoring.api
Host: 0.0.0.0
Port: 8886

# 检测目标配置
MonitorTargets:
  Services:       # TCP dial targets
    - Name: user-service
      Host: 127.0.0.1
      Port: 8080
    # ... more services
  Containers:     # docker ps names
    - mysql
    - redis
    - etcd
  AiModels:       # HTTP health check URLs
    - http://127.0.0.1:8890/api/ai-model/health
```

## 5. 架构决策

- **为什么是无状态**: 监控数据实时性要求高，缓存旧结果反而误导。每次请求现场探测
- **为什么不走 gRPC**: 监控面向运维人员，直接 HTTP 访问更简单（浏览器可直接查看）
- **为什么不依赖数据库**: 避免"监控系统本身需要被监控"的循环依赖

## 6. 技术栈

- 语言：Go
- 框架：go-zero (rest.MustNewServer + configx.MustLoad)
- 检测方法：`net.DialTimeout` (TCP), `os/exec` (docker ps), `net/http` (AI model)

## 7. 实际与设计差异

（首次创建，无差异）
