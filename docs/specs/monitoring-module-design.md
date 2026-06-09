# 运行监控模块 — 后端设计

## 概述

新建独立的 `monitoring-service`（API-only），聚合检查微服务端口、Docker 容器、AI 模型三类健康状态，前端 10 秒轮询。独立服务便于未来扩展日志、告警、指标等监控能力。

---

## 一、服务定位

```
monitoring-service（新建）
  API-only（无 RPC 层）
  端口：8886
  不依赖其他服务（纯外部探测）
  无数据库（只读探测，无状态）
```

---

## 二、API 设计

### 端点

```
GET /api/monitoring/health
```

无需认证，无请求参数。

### 响应格式

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "timestamp": "2026-06-05T23:00:00+08:00",
    "overall_status": "healthy",
    "services": [
      {
        "name": "user-service",
        "display_name": "用户服务",
        "api_port": 8882,
        "api_status": "healthy",
        "api_error": "",
        "rpc_port": 8082,
        "rpc_status": "healthy",
        "rpc_error": ""
      }
    ],
    "containers": [
      {
        "name": "mysql",
        "display_name": "MySQL 8.0",
        "image": "mysql:8.0",
        "status": "healthy",
        "running_for": "5 days",
        "error": ""
      }
    ],
    "ai_models": [
      {
        "id": "11",
        "name": "Claude Opus 4.7",
        "display_name": "Opus 4.7 Pro",
        "provider": "openai",
        "status": "healthy",
        "error": ""
      }
    ]
  }
}
```

### 状态枚举

| 值 | 含义 | 前端展示 |
|----|------|:---:|
| `healthy` | 正常 | 🟢 |
| `unhealthy` | 异常（不可达/未运行/超时） | 🔴 |
| `unknown` | 未配置检测 | ⚪ |

`overall_status`：全部 healthy → `healthy`，任一 unhealthy → `unhealthy`。

---

## 三、检测实现

### 3.1 微服务检测

**方式**：TCP 拨号（`net.DialTimeout`），3 秒超时。并发检测所有端口。

```go
func checkPort(host string, port int) (string, string) {
    if port == 0 {
        return "unknown", ""  // 无此端口不检测
    }
    addr := fmt.Sprintf("%s:%d", host, port)
    conn, err := net.DialTimeout("tcp", addr, 3*time.Second)
    if err != nil {
        return "unhealthy", err.Error()
    }
    conn.Close()
    return "healthy", ""
}
```

检测目标（YAML 配置，可随时增减）：

| 服务 | API 端口 | RPC 端口 |
|------|:---:|:---:|
| user-service | 8882 | 8082 |
| auth-service | 8881 | 8083 |
| permission-service | 8883 | 8084 |
| file-service | 8884 | 8085 |
| master-data-service | 8889 | 8087 |
| moderation-service | 8890 | 0 |
| ai-model-service | 8891 | 8080 |

`RpcPort: 0` → 跳过该端口检测，返回 `unknown`。

### 3.2 Docker 容器检测

**方式**：执行 `docker ps --format '{{json .}}' --all`，解析 JSON。

```go
type dockerContainer struct {
    Names      string `json:"Names"`
    Status     string `json:"Status"`
    RunningFor string `json:"RunningFor"`
    Image      string `json:"Image"`
}

func checkContainers(watchList []string) []ContainerHealth {
    cmd := exec.Command("docker", "ps", "--format", "{{json .}}", "--all")
    output, err := cmd.Output()
    if err != nil {
        // docker 不可用，所有容器标记 unhealthy
    }
    // 按行解析 JSON，用 watchList 过滤
    // Status 含 "Up" → healthy
}
```

**需要 docker 组权限**：运行 `monitoring-service` 的用户需在 `docker` 组中，或使用 `sudo`。

监控容器（YAML 配置）：

```yaml
Containers:
  - Name: mysql
    DisplayName: MySQL 8.0
  - Name: etcd
    DisplayName: etcd v3.5
  - Name: redis
    DisplayName: Redis 7
  - Name: apisix
    DisplayName: APISIX
  - Name: minio
    DisplayName: MinIO
```

### 3.3 AI 模型检测

**方式**：HTTP GET 调用 ai-model-service 的模型列表 API。

```go
func checkAiModels(endpoint string) []AiModelHealth {
    resp, err := http.Get(endpoint)  // 5s 超时
    // HTTP 200 + code:0 → 解析 data.models[]
    // HTTP 非 200 / 超时 / 不可达 → 返回空列表 + 标记
}
```

YAML 中配置端点地址，方便未来切换：

```yaml
AiModelCheck:
  Endpoint: http://127.0.0.1:8891/api/v1/models
  Timeout: 5s
```

---

## 四、服务结构

### 4.1 目录结构

```
services/monitoring-service/
  CLAUDE.md
  docs/
    design.md
  CHANGELOG.md
  api/
    monitoring.go                     # 入口：package main
    go.mod
    go.sum
    etc/
      monitoring-api.yaml             # 配置文件
    internal/
      config/
        config.go                     # Config 结构体
      handler/
        routes.go                     # 路由注册
        health/
          health_handler.go           # GET /api/monitoring/health
      logic/
        health/
          health_logic.go             # 核心检测逻辑
          service_checker.go          # TCP 端口检测
          container_checker.go        # Docker 容器检测
          ai_model_checker.go         # AI 模型检测
      svc/
        service_context.go            # ServiceContext
      types/
        types.go                      # 响应类型定义
```

### 4.2 模块信息

```
Go Module: github.com/guxiao1976/community-monitoring
端口：8886
依赖：go-zero (rest), community-common/v2 (configx, response)
```

### 4.3 配置文件

`api/etc/monitoring-api.yaml`：

```yaml
Name: monitoring-api
Host: 0.0.0.0
Port: 8886

Monitoring:
  Services:
    - Name: user-service
      DisplayName: 用户服务
      ApiPort: 8882
      RpcPort: 8082
    - Name: auth-service
      DisplayName: 认证服务
      ApiPort: 8881
      RpcPort: 8083
    - Name: permission-service
      DisplayName: 权限服务
      ApiPort: 8883
      RpcPort: 8084
    - Name: file-service
      DisplayName: 文件服务
      ApiPort: 8884
      RpcPort: 8085
    - Name: master-data-service
      DisplayName: 主数据服务
      ApiPort: 8889
      RpcPort: 8087
    - Name: moderation-service
      DisplayName: 审核服务
      ApiPort: 8890
      RpcPort: 0
    - Name: ai-model-service
      DisplayName: AI模型服务
      ApiPort: 8891
      RpcPort: 8080
  Containers:
    - Name: mysql
      DisplayName: MySQL 8.0
    - Name: etcd
      DisplayName: etcd v3.5
    - Name: redis
      DisplayName: Redis 7
    - Name: apisix
      DisplayName: APISIX
    - Name: minio
      DisplayName: MinIO
  AiModelCheck:
    Endpoint: http://127.0.0.1:8891/api/v1/models
    Timeout: 5s
```

### 4.4 类型定义

`api/internal/types/types.go`：

```go
package types

type HealthResponse struct {
    Timestamp     string            `json:"timestamp"`
    OverallStatus string            `json:"overall_status"`
    Services      []ServiceHealth   `json:"services"`
    Containers    []ContainerHealth `json:"containers"`
    AiModels      []AiModelHealth   `json:"ai_models"`
}

type ServiceHealth struct {
    Name        string `json:"name"`
    DisplayName string `json:"display_name"`
    ApiPort     int    `json:"api_port"`
    ApiStatus   string `json:"api_status"`
    ApiError    string `json:"api_error,omitempty"`
    RpcPort     int    `json:"rpc_port"`
    RpcStatus   string `json:"rpc_status"`
    RpcError    string `json:"rpc_error,omitempty"`
}

type ContainerHealth struct {
    Name        string `json:"name"`
    DisplayName string `json:"display_name"`
    Image       string `json:"image"`
    Status      string `json:"status"`
    RunningFor  string `json:"running_for"`
    Error       string `json:"error,omitempty"`
}

type AiModelHealth struct {
    Id          string `json:"id"`
    Name        string `json:"name"`
    DisplayName string `json:"display_name"`
    Provider    string `json:"provider"`
    Status      string `json:"status"`
    Error       string `json:"error,omitempty"`
}
```

---

## 五、go.work 集成

根 `go.work` 新增：

```
github.com/guxiao1976/community-monitoring

use ./services/monitoring-service/api
```

---

## 六、前端集成

### Vite 代理

`web/pc/vite.config.ts` 新增：

```typescript
'/api/monitoring': {
  target: 'http://127.0.0.1:8886',
  changeOrigin: true
}
```

### API 调用

```typescript
// web/pc/src/api/monitoring.ts
import request from '@/utils/request'

export interface HealthResponse {
  timestamp: string
  overall_status: string
  services: ServiceHealth[]
  containers: ContainerHealth[]
  ai_models: AiModelHealth[]
}

export interface ServiceHealth {
  name: string
  display_name: string
  api_port: number
  api_status: 'healthy' | 'unhealthy' | 'unknown'
  api_error?: string
  rpc_port: number
  rpc_status: 'healthy' | 'unhealthy' | 'unknown'
  rpc_error?: string
}

export interface ContainerHealth {
  name: string
  display_name: string
  image: string
  status: 'healthy' | 'unhealthy' | 'unknown'
  running_for: string
  error?: string
}

export interface AiModelHealth {
  id: string
  name: string
  display_name: string
  provider: string
  status: 'healthy' | 'unhealthy' | 'unknown'
  error?: string
}

export const getHealthStatus = () => {
  return request.get<HealthResponse>('/api/monitoring/health')
}
```

### 页面组件

```
web/pc/src/views/monitoring/
  HealthDashboard.vue    ← 主页面，3 个 Panel + 定时轮询
  ServicePanel.vue       ← 微服务状态表
  ContainerPanel.vue     ← Docker 容器状态表
  ModelPanel.vue         ← AI 模型状态表
```

### 路由

```typescript
{
  path: '/monitoring',
  name: 'Monitoring',
  component: Layout,
  meta: { title: '运行监控', icon: 'Monitor' },
  children: [
    {
      path: 'health',
      name: 'HealthDashboard',
      component: () => import('@/views/monitoring/HealthDashboard.vue'),
      meta: { title: '服务健康' }
    }
  ]
}
```

### 菜单

```
运行监控          ← 一级菜单
  └── 服务健康    ← 二级菜单
```

---

## 七、端口规划总览

| 服务 | API 端口 | RPC 端口 | 新增 |
|------|:---:|:---:|:---:|
| user-service | 8882 | 8082 | |
| auth-service | 8881 | 8083 | |
| permission-service | 8883 | 8084 | |
| file-service | 8884 | 8085 | |
| **monitoring-service** | **8886** | — | ✅ 新增 |
| master-data-service | 8889 | 8087 | |
| moderation-service | 8890 | — | |
| ai-model-service | 8891 | 8080 | |

---

## 八、未来扩展点

独立服务后，可逐步增加：

| 功能 | 说明 |
|------|------|
| 日志采集 | 读取各服务日志，统计错误率 |
| 资源监控 | CPU/内存/磁盘使用率 |
| 告警通知 | 服务异常时企业微信/邮件通知 |
| 历史记录 | 写入时序数据库，展示健康趋势图 |
| gRPC 服务发现 | 从 etcd 自动获取服务列表，无需静态配置 |
