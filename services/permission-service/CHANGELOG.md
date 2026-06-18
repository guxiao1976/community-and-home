# CHANGELOG — permission-service

## 2026-06-04 — W8: conf.MustLoad → configx.MustLoad 迁移

### 做了什么
- `rpc/permissionservice.go`：将 `conf.MustLoad` 替换为 `configx.MustLoad`，导入路径从 `go-zero/core/conf` 改为 `community-common/v2/pkg/configx`

### 为什么
根 CLAUDE.md 全局硬规则要求所有服务入口使用 `configx.MustLoad`（支持 `${VAR}` 环境变量展开）。API 层已在 C7 修复中完成迁移，本次补齐 RPC 层。

### 影响
- Proto: 无
- 调用方: 无
- 数据库: 无

## 2026-06-04 — C7 JWT Secret 硬编码修复

### 做了什么
- `api/etc/perm-api.yaml`：将硬编码的 `AccessSecret` 替换为环境变量 `${JWT_ACCESS_SECRET}`
- `api/perm.go`：将 `conf.MustLoad` 替换为 `configx.MustLoad`，确保 `${VAR}` 展开

### 为什么
审计发现 JWT Secret 占位符明文存放在配置中，不符合安全规范。改为环境变量后，密钥由根 `.env` 统一注入。

### 影响
- Proto: 无
- 调用方: 无
- 数据库: 无

---

## 2026-06-04 — 全局公约与设计文档

### 做了什么
- `CLAUDE.md` 新增 `## 全局公约` 章节，引用根 CLAUDE.md
- 首次创建设计文档 `docs/design.md`（数据库设计、权限检查流程、缓存策略、13 个 RPC 等）
- 添加 `CHANGELOG.md`（本文件）

### 为什么
项目规范化——补齐设计文档，子 Claude 启动时能感知全局架构规则和本服务设计决策。

### 已知问题（待修复）
- `CheckPermission` 中 `Expire` TTL 使用纳秒值，go-redis v9 可能未正确设置
- `GetDataScopes` 缓存 write-only，未真正加速读取
- `UpdateRole` 用 `KEYS *` 全量扫描，大规模部署需优化

### 影响
- Proto: 无
- 调用方: 无
- 数据库: 无
