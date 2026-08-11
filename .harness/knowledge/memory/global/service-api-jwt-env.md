---
triggers: ["JWT_ACCESS_SECRET", "env", "启动", "API layer", "panic", "secret"]
type: pitfall
severity: should-follow
service: all
status: active
created: 2026-06-17
updated: 2026-08-09
apply_count: 0
---
部分服务 API 层启动时要求 `JWT_ACCESS_SECRET` 环境变量存在且长度 ≥ 8，否则 panic：`secret's length can't be less than 8`。

**Why:** 配置文件（如 `file-api.yaml`、`perm-api.yaml`）使用 `${JWT_ACCESS_SECRET}` 占位符，go-zero 框架在注册 JWT 路由时校验 secret 长度。Docker Compose 启动时会自动注入该变量，但 `go run` 直接启动时不会。

**How to apply:** 启动服务前先 `export JWT_ACCESS_SECRET=$(grep JWT_ACCESS_SECRET .env | cut -d= -f2)`。

**相关服务:** file-service API (8884)、permission-service API (8883)
