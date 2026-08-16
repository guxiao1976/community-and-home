# CHANGELOG — file-service

## 2026-08-16 — 通用图文发布附件安全重构（content-post-generalization，Task 2.1-2.6）

> 任务类型：chore（非纯字段映射，含 L1/L2 校验逻辑函数，已按 TDD 留 RED 证据）

### 做了什么
- **错误码登记**：`ErrCodeUnsupportedFileType=70004` / `ErrCodeFileSizeExceeded=70005` 落到 4 个 errcode.go（rpc/internal/errx、rpc/internal/logic/file、api/internal/logic、api/internal/logic/file）；70001-70003 不重编号
- **L1 快速拒绝（GetUploadUrl）**：新建 `internal/guard` 包——白名单 {png,jpg,jpeg,gif,pdf,doc,docx}、禁止集 {exe,bat,sh,cmd,com,msi,apk,js,vbs,ps1,py,pl,php}、zip/rar 扩展名层全拒、10MB 硬上限（=10MB 放行）；预签名 URL 生成前调 `ValidateFileNameForEntityType` + `ValidateFileSize`（不触碰 MinIO）
- **L2 magic-bytes 嗅探（ConfirmUpload）**：`internal/guard/magic.go` `SniffType`——png/jpg/gif/pdf 魔数；doc=OLE2/CFB（D0 CF 11 E0 A1 B1 1A E1）+`WordDocument` 流（UTF-16LE）、docx=ZIP（PK 头）+`word/document.xml` 部件特判；msi/xls/ppt 改 doc、xlsx/pptx 改 docx、通用 zip/rar 改 docx → 070004；ConfirmUpload 回读 MinIO 对象前 64KB，嗅探类型与声明扩展名一致才放行（jpg/jpeg 等价），不一致 → 070004
- **落库**：`File.FileType`（嗅探规范扩展名）+ `Confirmed=true`；`toProtoFile` 填充 `file_type`/`confirmed`
- **模型扩展 + Migration 002**：`uploaded_file` 增 `file_type VARCHAR(20)` / `confirmed TINYINT NOT NULL DEFAULT 1`（存量行免嗅探即 confirmed）；`FileModel.Insert` 含新列；`File.UserID/EntityID` json tag 补 `,string`（Snowflake 规范，QA 门禁）
- **entity_type 覆盖机制**：`RegisterEntityTypeOverride` 基线上追加允许类型；禁止集不可放行、10MB 硬上限不可放宽（REQ-CAS-4 不变量）；GetUploadUrl 按 `in.EntityType` 查覆盖（content_posts 附件本期走全局基线）
- 新增测试：`internal/guard/whitelist_test.go`、`internal/guard/magic_test.go`、`rpc/internal/logic/file/getuploadurllogic_test.go`、`rpc/internal/logic/file/confirmuploadlogic_test.go`、`model/filemodel_test.go`（RED→GREEN，全绿）

### 为什么
社区图文发布组件（content_posts）附件需要白名单 + magic-bytes 内容校验，防伪装扩展名/容器子类型绕过，封堵上传投毒面（REQ-CAS-1/2/3/4）。

### 影响
- Proto: 无（FileInfo 已含 file_type/confirmed，已完成架构评审）
- 调用方: 无（REST wire 不变；RPC 响应 FileInfo 新增 2 字段，非破坏）
- 数据库: `migration/002_file_guard.sql`（执行由 Owner 验证）
- 前端: `web/common/types/file.d.ts` 已同步，本服务无需改前端类型

---

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
