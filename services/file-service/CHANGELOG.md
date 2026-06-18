# CHANGELOG — file-service

## 2026-06-04 — W8: conf.MustLoad → configx.MustLoad 迁移

### 做了什么
- `rpc/fileservice.go`：将 `conf.MustLoad` 替换为 `configx.MustLoad`，导入路径从 `go-zero/core/conf` 改为 `community-common/v2/pkg/configx`

### 为什么
根 CLAUDE.md 全局硬规则要求所有服务入口使用 `configx.MustLoad`（支持 `${VAR}` 环境变量展开）。API 层已在 C7 修复中完成迁移，本次补齐 RPC 层。

### 影响
- Proto: 无
- 调用方: 无
- 数据库: 无

## 2026-06-04 — C7 JWT Secret 硬编码修复

### 做了什么
- `api/etc/file-api.yaml`：将硬编码的 `AccessSecret`（`changeme`）替换为环境变量 `${JWT_ACCESS_SECRET}`
- `api/file.go`：将 `conf.MustLoad` 替换为 `configx.MustLoad`，确保 `${VAR}` 展开

### 为什么
审计发现 JWT Secret 占位符明文存放在配置中，不符合安全规范。改为环境变量后，密钥由根 `.env` 统一注入。

### 影响
- Proto: 无
- 调用方: 无
- 数据库: 无

---

## 2026-06-04 — W10 MinIO 凭证迁移

### 做了什么
- `rpc/etc/fileservice.yaml`：MinIO `AccessKey`（`admin`）和 `SecretKey`（`admin123`）替换为 `${MINIO_ACCESS_KEY}` 和 `${MINIO_SECRET_KEY}` 环境变量引用
- 根 `.env.example` 新增 `MINIO_ACCESS_KEY`、`MINIO_SECRET_KEY` 占位符

### 为什么
MinIO 管理员凭证硬编码在配置文件中构成安全风险，迁移到环境变量由根 `.env` 统一管理。

### 影响
- 配置: `fileservice.yaml` Minio.AccessKey / Minio.SecretKey 字段变更
- 部署: 需在 `.env` 中设置 `MINIO_ACCESS_KEY` / `MINIO_SECRET_KEY`（默认值与 docker-compose.yml 一致）
- 数据库: 无变更

---

## 2026-06-04 — 全局公约与设计文档

### 做了什么
- `CLAUDE.md` 新增 `## 全局公约` 章节，引用根 CLAUDE.md
- 首次创建设计文档 `docs/design.md`（数据库设计、上传/下载流程、5 个 RPC 等）
- 添加 `CHANGELOG.md`（本文件）

### 为什么
项目规范化——补齐设计文档，子 Claude 启动时能感知全局架构规则和本服务设计决策。

### 影响
- Proto: 无
- 调用方: 无
- 数据库: 无
