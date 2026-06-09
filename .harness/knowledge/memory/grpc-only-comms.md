---
triggers: ["gRPC", "服务间通信", "直连数据库", "HTTP调用", "跨服务"]
service: all
type: guideline
severity: must-follow
status: active
created: 2026-06-05
updated: 2026-06-05
last_applied: null
apply_count: 0
---

# 服务间通信仅通过 gRPC，禁止直连数据库

## 为什么会有这条经验
moderation-service 曾直接读取 masterdata_db 数据库，绕过 master-data-service 的 gRPC 接口。
这导致数据逻辑分散、耦合紧密、难以维护。这是架构债务。

## 怎么做
1. 所有服务间数据访问必须通过 gRPC 接口
2. 接口定义在 `api-proto/` 中统一管理
3. 禁止在服务中配置其他服务的数据库连接
4. 如需其他服务的数据，调用其 gRPC 接口，而非直接查库

## 怎么验证
- 检查服务配置文件中是否有其他服务的数据库 DSN
- `grep -r "mysql://" services/<name>/` 应只包含本服务的数据库
- gRPC 客户端初始化在 `internal/svc/servicecontext.go` 中

## 关联经验
- [[proto-jstype]]
