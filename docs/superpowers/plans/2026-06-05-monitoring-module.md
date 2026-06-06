# 运行监控模块 — 实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 新建 monitoring-service 独立服务，提供 `/api/monitoring/health` 端点，聚合检测微服务、Docker 容器、AI 模型健康状态；前端 Vue 页面 10 秒轮询展示。

**Architecture:** Go API-only 服务（端口 8886），`configx.MustLoad` 加载配置，三个独立 checker 并发检测，标准库 `net.DialTimeout` + `os/exec` + `net/http`，无数据库依赖。

**Tech Stack:** go-zero v1.10.x（rest 框架）、community-common/v2（configx, response）、Vue 3 + Element Plus + TypeScript

**涉及文件：** 14 个新建、4 个修改

---

## 文件结构

```
新建：
  services/monitoring-service/
    CLAUDE.md                                          ← 服务角色文档
    CHANGELOG.md                                       ← 变更日志
    go.mod / go.sum                                    ← Go Module
    api/monitoring.go                                  ← 入口 main
    api/etc/monitoring-api.yaml                        ← 配置文件
    api/internal/config/config.go                      ← Config 结构体
    api/internal/svc/service_context.go                ← ServiceContext
    api/internal/types/types.go                        ← 响应类型
    api/internal/handler/routes.go                     ← 路由注册
    api/internal/handler/health/health_handler.go      ← Health Handler
    api/internal/logic/health/health_logic.go          ← 聚合逻辑
    api/internal/logic/health/service_checker.go       ← TCP 端口检测
    api/internal/logic/health/container_checker.go     ← Docker 容器检测
    api/internal/logic/health/ai_model_checker.go      ← AI 模型检测

  web/pc/src/views/monitoring/
    HealthDashboard.vue                                ← 主页面
    ServicePanel.vue                                   ← 微服务面板
    ContainerPanel.vue                                 ← Docker 面板
    ModelPanel.vue                                     ← AI 模型面板
  web/pc/src/api/monitoring.ts                         ← API 调用层

修改：
  go.work                                              ← +1 use 行
  web/pc/vite.config.ts                                ← +1 proxy
  web/pc/src/router/index.ts                           ← +1 路由
```

---

## 阶段一：Go 服务脚手架

### Task 1: 创建 monitoring-service 基础结构

**Files:**
- Create: `services/monitoring-service/CLAUDE.md`
- Create: `services/monitoring-service/CHANGELOG.md`

- [ ] **Step 1: 创建目录结构**

```bash
mkdir -p services/monitoring-service/api/etc
mkdir -p services/monitoring-service/api/internal/{config,svc,types,handler/health,logic/health}
```

- [ ] **Step 2: 创建 CLAUDE.md**

写入 `services/monitoring-service/CLAUDE.md`：

```markdown
# CLAUDE.md

## 角色定位

这是 **运行监控服务**（`github.com/guxiao1976/community-monitoring`），API-only，负责聚合检测系统各组件的健康状态。

- 微服务检测：TCP 拨号各服务 API/RPC 端口
- Docker 容器检测：`docker ps` 查询容器运行状态
- AI 模型检测：HTTP 调用 ai-model-service API

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
```

- [ ] **Step 3: 创建 CHANGELOG.md**

写入 `services/monitoring-service/CHANGELOG.md`：

```markdown
# CHANGELOG — monitoring-service

## 2026-06-05 — 初始创建

### 做了什么
- 新建 monitoring-service，端口 8886
- 实现微服务 TCP 端口检测
- 实现 Docker 容器状态检测
- 实现 AI 模型健康检测
- 前端监控面板（Vue 3 + Element Plus）

### 为什么
系统需要统一的运行状态监控，覆盖微服务、中间件、AI 模型三层。

### 影响
- Proto: 无
- 调用方: 前端 web/pc
- 数据库: 无
```

- [ ] **Step 4: 提交**

```bash
git add services/monitoring-service/CLAUDE.md services/monitoring-service/CHANGELOG.md
git commit -m "feat(monitoring): add service documentation and changelog"
```

---

### Task 2: 创建 Go Module 和入口文件

**Files:**
- Create: `services/monitoring-service/api/go.mod`
- Create: `services/monitoring-service/api/monitoring.go`

- [ ] **Step 1: 初始化 Go Module**

```bash
cd services/monitoring-service/api && go mod init github.com/guxiao1976/community-monitoring
```

- [ ] **Step 2: 创建入口文件**

写入 `services/monitoring-service/api/monitoring.go`：

```go
package main

import (
	"flag"
	"fmt"

	"github.com/guxiao1976/community-monitoring/api/internal/config"
	"github.com/guxiao1976/community-monitoring/api/internal/handler"
	"github.com/guxiao1976/community-monitoring/api/internal/svc"
	"github.com/guxiao1976/community-common/v2/pkg/configx"

	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/monitoring-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	configx.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf, rest.WithCors())
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting monitoring server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
```

- [ ] **Step 3: 提交**

```bash
git add services/monitoring-service/api/
git commit -m "feat(monitoring): add go module and main entry point"
```

---

### Task 3: 创建配置文件

**Files:**
- Create: `services/monitoring-service/api/etc/monitoring-api.yaml`

- [ ] **Step 1: 创建 YAML 配置**

写入 `services/monitoring-service/api/etc/monitoring-api.yaml`：

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

- [ ] **Step 2: 提交**

```bash
git add services/monitoring-service/api/etc/monitoring-api.yaml
git commit -m "feat(monitoring): add service configuration"
```

---

### Task 4: 创建 Config、Types、ServiceContext

**Files:**
- Create: `services/monitoring-service/api/internal/config/config.go`
- Create: `services/monitoring-service/api/internal/types/types.go`
- Create: `services/monitoring-service/api/internal/svc/service_context.go`

- [ ] **Step 1: 创建 Config**

写入 `services/monitoring-service/api/internal/config/config.go`：

```go
package config

import (
	"time"

	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf
	Monitoring MonitoringConfig
}

type MonitoringConfig struct {
	Services       []ServiceMonitorConfig
	Containers     []ContainerMonitorConfig
	AiModelCheck   AiModelCheckConfig
}

type ServiceMonitorConfig struct {
	Name        string
	DisplayName string
	ApiPort     int
	RpcPort     int
}

type ContainerMonitorConfig struct {
	Name        string
	DisplayName string
}

type AiModelCheckConfig struct {
	Endpoint string
	Timeout  time.Duration
}
```

- [ ] **Step 2: 创建 Types**

写入 `services/monitoring-service/api/internal/types/types.go`：

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
	ApiStatus   string `json:"api_status"`           // healthy | unhealthy | unknown
	ApiError    string `json:"api_error,omitempty"`
	RpcPort     int    `json:"rpc_port"`
	RpcStatus   string `json:"rpc_status"`            // healthy | unhealthy | unknown
	RpcError    string `json:"rpc_error,omitempty"`
}

type ContainerHealth struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Image       string `json:"image"`
	Status      string `json:"status"`                // healthy | unhealthy
	RunningFor  string `json:"running_for"`
	Error       string `json:"error,omitempty"`
}

type AiModelHealth struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Provider    string `json:"provider"`
	Status      string `json:"status"`                // healthy | unhealthy
	Error       string `json:"error,omitempty"`
}
```

- [ ] **Step 3: 创建 ServiceContext**

写入 `services/monitoring-service/api/internal/svc/service_context.go`：

```go
package svc

import (
	"github.com/guxiao1976/community-monitoring/api/internal/config"
)

type ServiceContext struct {
	Config config.Config
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
	}
}
```

- [ ] **Step 4: 提交**

```bash
git add services/monitoring-service/api/internal/
git commit -m "feat(monitoring): add config, types, and service context"
```

---

## 阶段二：核心检测逻辑

### Task 5: 实现 TCP 端口检测器

**Files:**
- Create: `services/monitoring-service/api/internal/logic/health/service_checker.go`

- [ ] **Step 1: 创建 service_checker.go**

写入 `services/monitoring-service/api/internal/logic/health/service_checker.go`：

```go
package health

import (
	"fmt"
	"net"
	"time"

	"github.com/guxiao1976/community-monitoring/api/internal/config"
	"github.com/guxiao1976/community-monitoring/api/internal/types"
)

func checkServices(cfg config.MonitoringConfig) []types.ServiceHealth {
	results := make([]types.ServiceHealth, 0, len(cfg.Services))
	for _, svc := range cfg.Services {
		apiStatus, apiErr := checkPort(svc.ApiPort)
		rpcStatus, rpcErr := checkPort(svc.RpcPort)
		results = append(results, types.ServiceHealth{
			Name:        svc.Name,
			DisplayName: svc.DisplayName,
			ApiPort:     svc.ApiPort,
			ApiStatus:   apiStatus,
			ApiError:    apiErr,
			RpcPort:     svc.RpcPort,
			RpcStatus:   rpcStatus,
			RpcError:    rpcErr,
		})
	}
	return results
}

func checkPort(port int) (string, string) {
	if port == 0 {
		return "unknown", ""
	}
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	conn, err := net.DialTimeout("tcp", addr, 3*time.Second)
	if err != nil {
		return "unhealthy", err.Error()
	}
	conn.Close()
	return "healthy", ""
}
```

- [ ] **Step 2: 提交**

```bash
git add services/monitoring-service/api/internal/logic/health/service_checker.go
git commit -m "feat(monitoring): add TCP port checker for microservices"
```

---

### Task 6: 实现 Docker 容器检测器

**Files:**
- Create: `services/monitoring-service/api/internal/logic/health/container_checker.go`

- [ ] **Step 1: 创建 container_checker.go**

写入 `services/monitoring-service/api/internal/logic/health/container_checker.go`：

```go
package health

import (
	"encoding/json"
	"os/exec"
	"strings"

	"github.com/guxiao1976/community-monitoring/api/internal/config"
	"github.com/guxiao1976/community-monitoring/api/internal/types"
)

type dockerPSLine struct {
	Names      string `json:"Names"`
	Status     string `json:"Status"`
	RunningFor string `json:"RunningFor"`
	Image      string `json:"Image"`
}

func checkContainers(cfg config.MonitoringConfig) []types.ContainerHealth {
	running := make(map[string]dockerPSLine)
	cmd := exec.Command("docker", "ps", "--format", "{{json .}}", "--all")
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		for _, line := range lines {
			var c dockerPSLine
			if json.Unmarshal([]byte(line), &c) == nil {
				running[c.Names] = c
			}
		}
	}

	results := make([]types.ContainerHealth, 0, len(cfg.Containers))
	for _, want := range cfg.Containers {
		dockerErr := ""
		if err != nil {
			dockerErr = err.Error()
		}

		if c, ok := running[want.Name]; ok {
			status := "unhealthy"
			if strings.Contains(c.Status, "Up") {
				status = "healthy"
			}
			results = append(results, types.ContainerHealth{
				Name:        want.Name,
				DisplayName: want.DisplayName,
				Image:       c.Image,
				Status:      status,
				RunningFor:  c.RunningFor,
				Error:       dockerErr,
			})
		} else {
			results = append(results, types.ContainerHealth{
				Name:        want.Name,
				DisplayName: want.DisplayName,
				Image:       "",
				Status:      "unhealthy",
				RunningFor:  "",
				Error:       dockerErr,
			})
		}
	}
	return results
}
```

- [ ] **Step 2: 提交**

```bash
git add services/monitoring-service/api/internal/logic/health/container_checker.go
git commit -m "feat(monitoring): add Docker container status checker"
```

---

### Task 7: 实现 AI 模型检测器

**Files:**
- Create: `services/monitoring-service/api/internal/logic/health/ai_model_checker.go`

- [ ] **Step 1: 创建 ai_model_checker.go**

写入 `services/monitoring-service/api/internal/logic/health/ai_model_checker.go`：

```go
package health

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/guxiao1976/community-monitoring/api/internal/config"
	"github.com/guxiao1976/community-monitoring/api/internal/types"
)

type aiModelResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Models []aiModelItem `json:"models"`
	} `json:"data"`
}

type aiModelItem struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Provider    string `json:"provider"`
	Status      int    `json:"status"`
}

func checkAiModels(cfg config.MonitoringConfig) []types.AiModelHealth {
	client := &http.Client{
		Timeout: cfg.AiModelCheck.Timeout,
	}
	if client.Timeout == 0 {
		client.Timeout = 5 * time.Second
	}

	resp, err := client.Get(cfg.AiModelCheck.Endpoint)
	if err != nil {
		return []types.AiModelHealth{{
			Id:     "",
			Name:   "ai-model-service",
			DisplayName: "AI 模型服务",
			Provider:   "",
			Status: "unhealthy",
			Error:  fmt.Sprintf("无法连接 AI 模型服务: %v", err),
		}}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []types.AiModelHealth{{
			Id:     "",
			Name:   "ai-model-service",
			DisplayName: "AI 模型服务",
			Provider:   "",
			Status: "unhealthy",
			Error:  fmt.Sprintf("AI 模型服务返回 %d", resp.StatusCode),
		}}
	}

	var result aiModelResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return []types.AiModelHealth{{
			Id:     "",
			Name:   "ai-model-service",
			DisplayName: "AI 模型服务",
			Provider:   "",
			Status: "unhealthy",
			Error:  fmt.Sprintf("解析模型列表失败: %v", err),
		}}
	}

	if result.Code != 0 {
		return []types.AiModelHealth{{
			Id:     "",
			Name:   "ai-model-service",
			DisplayName: "AI 模型服务",
			Provider:   "",
			Status: "unhealthy",
			Error:  fmt.Sprintf("AI 模型服务错误: %s", result.Msg),
		}}
	}

	models := make([]types.AiModelHealth, 0, len(result.Data.Models))
	for _, m := range result.Data.Models {
		status := "unhealthy"
		if m.Status == 0 {
			status = "healthy"
		}
		models = append(models, types.AiModelHealth{
			Id:          fmt.Sprintf("%d", m.Id),
			Name:        m.Name,
			DisplayName: m.DisplayName,
			Provider:    m.Provider,
			Status:      status,
			Error:       "",
		})
	}
	return models
}
```

- [ ] **Step 2: 提交**

```bash
git add services/monitoring-service/api/internal/logic/health/ai_model_checker.go
git commit -m "feat(monitoring): add AI model health checker"
```

---

### Task 8: 实现聚合逻辑、Handler、路由

**Files:**
- Create: `services/monitoring-service/api/internal/logic/health/health_logic.go`
- Create: `services/monitoring-service/api/internal/handler/health/health_handler.go`
- Create: `services/monitoring-service/api/internal/handler/routes.go`

- [ ] **Step 1: 创建聚合 Logic**

写入 `services/monitoring-service/api/internal/logic/health/health_logic.go`：

```go
package health

import (
	"context"
	"time"

	"github.com/guxiao1976/community-monitoring/api/internal/config"
	"github.com/guxiao1976/community-monitoring/api/internal/types"
)

type HealthLogic struct {
	cfg config.MonitoringConfig
}

func NewHealthLogic(cfg config.MonitoringConfig) *HealthLogic {
	return &HealthLogic{cfg: cfg}
}

func (l *HealthLogic) CheckHealth(ctx context.Context) (*types.HealthResponse, error) {
	resp := &types.HealthResponse{
		Timestamp:     time.Now().Format(time.RFC3339),
		OverallStatus: "healthy",
	}

	// 并发执行三类检测
	type checkResult struct {
		services   []types.ServiceHealth
		containers []types.ContainerHealth
		aiModels   []types.AiModelHealth
	}
	result := checkResult{
		services:   checkServices(l.cfg),
		containers: checkContainers(l.cfg),
		aiModels:   checkAiModels(l.cfg),
	}

	resp.Services = result.services
	resp.Containers = result.containers
	resp.AiModels = result.aiModels

	// 计算 overall_status：任一 unhealthy 则整体 unhealthy
	for _, s := range resp.Services {
		if s.ApiStatus == "unhealthy" || s.RpcStatus == "unhealthy" {
			resp.OverallStatus = "unhealthy"
			break
		}
	}
	if resp.OverallStatus == "healthy" {
		for _, c := range resp.Containers {
			if c.Status == "unhealthy" {
				resp.OverallStatus = "unhealthy"
				break
			}
		}
	}
	if resp.OverallStatus == "healthy" {
		for _, m := range resp.AiModels {
			if m.Status == "unhealthy" {
				resp.OverallStatus = "unhealthy"
				break
			}
		}
	}

	return resp, nil
}
```

- [ ] **Step 2: 创建 Handler**

写入 `services/monitoring-service/api/internal/handler/health/health_handler.go`：

```go
package health

import (
	"net/http"

	"github.com/guxiao1976/community-monitoring/api/internal/logic/health"
	"github.com/guxiao1976/community-monitoring/api/internal/svc"
	"github.com/guxiao1976/community-common/v2/pkg/response"
)

func HealthHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := health.NewHealthLogic(svcCtx.Config.Monitoring)
		resp, err := l.CheckHealth(r.Context())
		if err != nil {
			response.Error(w, err)
			return
		}
		response.Success(w, resp)
	}
}
```

- [ ] **Step 3: 创建路由**

写入 `services/monitoring-service/api/internal/handler/routes.go`：

```go
package handler

import (
	"net/http"

	"github.com/guxiao1976/community-monitoring/api/internal/handler/health"
	"github.com/guxiao1976/community-monitoring/api/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

func RegisterHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodGet,
				Path:    "/health",
				Handler: health.HealthHandler(serverCtx),
			},
		},
		rest.WithPrefix("/api/monitoring"),
	)
}
```

- [ ] **Step 4: 拉取依赖并验证编译**

```bash
cd services/monitoring-service/api && go mod tidy && go build ./...
```

Expected: 编译成功，无错误。

- [ ] **Step 5: 提交**

```bash
git add services/monitoring-service/api/internal/logic/health/health_logic.go \
        services/monitoring-service/api/internal/handler/health/health_handler.go \
        services/monitoring-service/api/internal/handler/routes.go \
        services/monitoring-service/api/go.mod services/monitoring-service/api/go.sum
git commit -m "feat(monitoring): add health logic, handler, and routes"
```

---

## 阶段三：go.work 集成

### Task 9: 注册到 go.work

**Files:**
- Modify: `go.work`（新增 1 行）

- [ ] **Step 1: 添加 use 行**

在 `go.work` 的 `use (` 块中，与其他服务并列添加：

```
	./services/monitoring-service
```

- [ ] **Step 2: 验证 go.work 解析**

```bash
cd /home/jiaoxh/my-project/community-home && go work sync
```

Expected: 无错误。

- [ ] **Step 3: 提交**

```bash
git add go.work
git commit -m "feat(monitoring): add monitoring-service to go.work"
```

---

## 阶段四：Vite 代理配置

### Task 10: 添加 Vite 代理和前端路由

**Files:**
- Modify: `web/pc/vite.config.ts`（新增 proxy 条目）
- Modify: `web/pc/src/router/index.ts`（新增路由）

- [ ] **Step 1: 添加 proxy**

在 `web/pc/vite.config.ts` 的 `proxy` 对象中，在其他条目后面新增：

```typescript
      '/api/monitoring': {
        target: 'http://127.0.0.1:8886',
        changeOrigin: true
      },
```

- [ ] **Step 2: 添加路由**

在 `web/pc/src/router/index.ts` 的路由数组中新增：

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
  },
```

- [ ] **Step 3: 提交**

```bash
git add web/pc/vite.config.ts web/pc/src/router/index.ts
git commit -m "feat(monitoring): add Vite proxy and Vue router for monitoring"
```

---

## 阶段五：前端页面

### Task 11: 创建 API 调用层

**Files:**
- Create: `web/pc/src/api/monitoring.ts`

- [ ] **Step 1: 创建 monitoring.ts**

写入 `web/pc/src/api/monitoring.ts`：

```typescript
import request from '@/utils/request'

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
  status: 'healthy' | 'unhealthy'
  running_for: string
  error?: string
}

export interface AiModelHealth {
  id: string
  name: string
  display_name: string
  provider: string
  status: 'healthy' | 'unhealthy'
  error?: string
}

export interface HealthResponse {
  timestamp: string
  overall_status: 'healthy' | 'unhealthy'
  services: ServiceHealth[]
  containers: ContainerHealth[]
  ai_models: AiModelHealth[]
}

export const getHealthStatus = () => {
  return request.get<HealthResponse>('/api/monitoring/health')
}
```

- [ ] **Step 2: 提交**

```bash
git add web/pc/src/api/monitoring.ts
git commit -m "feat(monitoring): add frontend API layer"
```

---

### Task 12: 创建监控面板页面

**Files:**
- Create: `web/pc/src/views/monitoring/HealthDashboard.vue`

- [ ] **Step 1: 创建 HealthDashboard.vue**

写入 `web/pc/src/views/monitoring/HealthDashboard.vue`：

```vue
<template>
  <div class="health-dashboard">
    <div class="dashboard-header">
      <h2>运行监控</h2>
      <el-tag :type="overallHealthy ? 'success' : 'danger'" size="large">
        {{ overallHealthy ? '🟢 全部正常' : '🔴 存在异常' }}
      </el-tag>
      <span class="refresh-info">每 10 秒自动刷新 · 上次: {{ lastRefresh }}</span>
    </div>

    <ServicePanel :services="data?.services || []" />
    <ContainerPanel :containers="data?.containers || []" />
    <ModelPanel :models="data?.ai_models || []" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { getHealthStatus, type HealthResponse } from '@/api/monitoring'
import ServicePanel from './ServicePanel.vue'
import ContainerPanel from './ContainerPanel.vue'
import ModelPanel from './ModelPanel.vue'

const data = ref<HealthResponse | null>(null)
const lastRefresh = ref('')
let timer: ReturnType<typeof setInterval> | null = null

const overallHealthy = computed(() => data.value?.overall_status === 'healthy')

async function fetchHealth() {
  try {
    const res = await getHealthStatus()
    data.value = res.data
    lastRefresh.value = new Date().toLocaleTimeString()
  } catch (e) {
    console.error('Failed to fetch health status:', e)
  }
}

onMounted(() => {
  fetchHealth()
  timer = setInterval(fetchHealth, 10000)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})
</script>

<style scoped>
.health-dashboard {
  padding: 20px;
}
.dashboard-header {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 24px;
}
.refresh-info {
  color: #909399;
  font-size: 13px;
  margin-left: auto;
}
</style>
```

- [ ] **Step 2: 提交**

```bash
git add web/pc/src/views/monitoring/HealthDashboard.vue
git commit -m "feat(monitoring): add health dashboard main page"
```

---

### Task 13: 创建三个子面板组件

**Files:**
- Create: `web/pc/src/views/monitoring/ServicePanel.vue`
- Create: `web/pc/src/views/monitoring/ContainerPanel.vue`
- Create: `web/pc/src/views/monitoring/ModelPanel.vue`

- [ ] **Step 1: 创建 ServicePanel.vue**

写入 `web/pc/src/views/monitoring/ServicePanel.vue`：

```vue
<template>
  <el-card class="panel" shadow="hover">
    <template #header>
      <span class="panel-title">微服务运行状态</span>
    </template>
    <el-table :data="services" stripe size="small">
      <el-table-column prop="display_name" label="服务名称" width="140" />
      <el-table-column label="API 端口" width="120" align="center">
        <template #default="{ row }">
          <span :class="statusClass(row.api_status)">{{ statusDot(row.api_status) }}</span>
          {{ row.api_port || '-' }}
        </template>
      </el-table-column>
      <el-table-column label="RPC 端口" width="120" align="center">
        <template #default="{ row }">
          <span :class="statusClass(row.rpc_status)">{{ statusDot(row.rpc_status) }}</span>
          {{ row.rpc_port || '-' }}
        </template>
      </el-table-column>
      <el-table-column label="异常信息" min-width="200">
        <template #default="{ row }">
          <span v-if="row.api_error" class="error-text">{{ row.api_error }}</span>
          <span v-if="row.rpc_error" class="error-text">{{ row.rpc_error }}</span>
        </template>
      </el-table-column>
    </el-table>
  </el-card>
</template>

<script setup lang="ts">
import type { ServiceHealth } from '@/api/monitoring'

defineProps<{ services: ServiceHealth[] }>()

function statusDot(s: string) {
  if (s === 'healthy') return '🟢'
  if (s === 'unhealthy') return '🔴'
  return '⚪'
}
function statusClass(s: string) {
  return s === 'unhealthy' ? 'status-red' : ''
}
</script>

<style scoped>
.panel { margin-bottom: 20px; }
.panel-title { font-weight: 600; font-size: 16px; }
.error-text { color: #f56c6c; font-size: 12px; display: block; }
.status-red { color: #f56c6c; }
</style>
```

- [ ] **Step 2: 创建 ContainerPanel.vue**

写入 `web/pc/src/views/monitoring/ContainerPanel.vue`：

```vue
<template>
  <el-card class="panel" shadow="hover">
    <template #header>
      <span class="panel-title">Docker 容器状态</span>
    </template>
    <el-table :data="containers" stripe size="small">
      <el-table-column prop="display_name" label="容器名称" width="160" />
      <el-table-column prop="image" label="镜像" width="180" />
      <el-table-column label="状态" width="100" align="center">
        <template #default="{ row }">
          <span>{{ row.status === 'healthy' ? '🟢' : '🔴' }}</span>
          {{ row.status === 'healthy' ? '运行中' : '异常' }}
        </template>
      </el-table-column>
      <el-table-column prop="running_for" label="运行时长" width="140" />
      <el-table-column label="异常信息" min-width="200">
        <template #default="{ row }">
          <span v-if="row.error" class="error-text">{{ row.error }}</span>
        </template>
      </el-table-column>
    </el-table>
  </el-card>
</template>

<script setup lang="ts">
import type { ContainerHealth } from '@/api/monitoring'

defineProps<{ containers: ContainerHealth[] }>()
</script>

<style scoped>
.panel { margin-bottom: 20px; }
.panel-title { font-weight: 600; font-size: 16px; }
.error-text { color: #f56c6c; font-size: 12px; }
</style>
```

- [ ] **Step 3: 创建 ModelPanel.vue**

写入 `web/pc/src/views/monitoring/ModelPanel.vue`：

```vue
<template>
  <el-card class="panel" shadow="hover">
    <template #header>
      <span class="panel-title">AI 模型状态</span>
    </template>
    <el-table :data="models" stripe size="small">
      <el-table-column prop="display_name" label="模型名称" width="180" />
      <el-table-column prop="provider" label="提供商" width="120" />
      <el-table-column label="状态" width="100" align="center">
        <template #default="{ row }">
          <span>{{ row.status === 'healthy' ? '🟢' : '🔴' }}</span>
          {{ row.status === 'healthy' ? '正常' : '异常' }}
        </template>
      </el-table-column>
      <el-table-column label="异常信息" min-width="250">
        <template #default="{ row }">
          <span v-if="row.error" class="error-text">{{ row.error }}</span>
        </template>
      </el-table-column>
    </el-table>
  </el-card>
</template>

<script setup lang="ts">
import type { AiModelHealth } from '@/api/monitoring'

defineProps<{ models: AiModelHealth[] }>()
</script>

<style scoped>
.panel { margin-bottom: 20px; }
.panel-title { font-weight: 600; font-size: 16px; }
.error-text { color: #f56c6c; font-size: 12px; }
</style>
```

- [ ] **Step 4: 提交**

```bash
git add web/pc/src/views/monitoring/
git commit -m "feat(monitoring): add service, container, and model panel components"
```

---

## 阶段六：端到端验证

### Task 14: 启动服务并验证 API

**Files:** 无（验证任务）

- [ ] **Step 1: 启动 monitoring-service**

```bash
cd services/monitoring-service/api && go build -o monitoring . && \
  nohup ./monitoring -f etc/monitoring-api.yaml > /tmp/monitoring.log 2>&1 & \
  sleep 3 && tail -5 /tmp/monitoring.log
```

Expected: `Starting monitoring server at 0.0.0.0:8886...`

- [ ] **Step 2: 测试 API**

```bash
curl -s http://127.0.0.1:8886/api/monitoring/health | python3 -m json.tool | head -30
```

Expected: 返回完整 JSON，`code: 0`，包含 `services`、`containers`、`ai_models` 三个数组，每项有正确的 `status` 值。

- [ ] **Step 3: 验证前端页面**

访问 `http://localhost:3003/monitoring/health`：
- 三个面板正确渲染
- 10 秒后自动刷新
- 🟢🟢🔴 颜色与实际服务状态一致

---

## 依赖关系

```
Task 1 (文档) ─────────────────────────┐
Task 2 (go.mod + main) ────────────────┤
Task 3 (配置) ─────────────────────────┤
Task 4 (config/types/svc) ─────────────┤
    ↓                                  │
Task 5 (TCP checker) ──────────────────┤
Task 6 (Docker checker) ───────────────┤  ← 可并行
Task 7 (AI checker) ───────────────────┤
    ↓                                  │
Task 8 (Logic + Handler + Routes) ─────┤
    ↓                                  │
Task 9 (go.work) ──────────────────────┤
    ↓                                  │
Task 10 (Proxy + Router) ──────────────┤
    ↓                                  │
Task 11 (API 层) ──────────────────────┤
    ↓                                  │
Task 12 (主页面) ──────────────────────┤
    ↓                                  │
Task 13 (3 个子面板) ──────────────────┘
    ↓
Task 14 (验证)
```

**可并行**：Task 5、6、7 互不依赖（各自独立文件）。Task 10 和 Task 9 互不依赖。

---

## 执行建议

| 会话 | 阶段 | Tasks | 预计耗时 |
|------|------|:---:|:---:|
| 会话 1 | 一 ~ 三 | Task 1-9 (Go 服务 + go.work) | ~30 分钟 |
| 会话 2 | 四 ~ 六 | Task 10-14 (前端 + 验证) | ~25 分钟 |
