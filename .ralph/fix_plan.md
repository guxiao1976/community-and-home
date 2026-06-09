# Ralph Fix Plan — CRITICAL Issues

## Phase 1: Proto 修复

- [x] **1.1 masterdata.proto repeated int64 ids 加 jstype**: `api-proto/api/masterdata/v1/masterdata.proto:77,154` — `repeated int64 ids = 1;` → `repeated int64 ids = 1 [jstype = JS_STRING];`
- [x] **1.2 masterdata.proto 添加 BaseResp**: 所有 Response message 添加 `common.v1.BaseResp base = 1;`，import common.proto，替换裸 `bool success`
- [x] **1.3 ai-model-service proto 添加 BaseResp**: 所有 Response 添加 BaseResp，import common.proto
- [x] **1.4 moderation-service proto 添加 BaseResp**: 所有 Response 添加 BaseResp，import common.proto

## Phase 2: gRPC 通信合规修复

- [x] **2.1 moderation-service 去掉直连 masterdata_db**: 改为通过 gRPC 调用 master-data-service，移除直接数据库连接
- [x] **2.2 moderation-service 去掉跨服务 migration**: 移除修改 masterdata_db schema 的 migration

## Phase 3: 错误码修复

- [x] **3.1 修复 500400 为 5 位格式**: 找到并改为 50400 或 50040
- [x] **3.2 auth-service 错误码前缀修正**: `100002` 改为 50xxxx（Auth 前缀）

## Phase 4: 响应格式修复

- [x] **4.1 user-service API 改为统一响应格式**: `{code, msg, data}` wrapper
- [x] **4.2 ai-model-service API 响应格式对齐**: 统一 `{code, msg, data}`，修正字段名
- [x] **4.3 ai-model-service gRPC 添加 BaseResp**: 替换 gRPC status codes 为 BaseResp 模式
- [x] **4.4 master-data-service gRPC 添加 BaseResp**: 替换 `return nil, fmt.Errorf(...)` 为 BaseResp
- [x] **4.5 moderation-service gRPC 添加 BaseResp**: 对齐 proto + 代码

## Phase 5: 配置修复

- [x] **5.1 ai-model-service 改用 configx.MustLoad**: 2 个入口点，替换 `conf.MustLoad` → `configx.MustLoad`
- [x] **5.2 YAML 配置密码改为环境变量**: 10 个 MySQL 密码 + 5 个 Redis 密码，改为 `${MYSQL_PASSWORD}` / `${REDIS_PASSWORD}`
- [x] **5.3 RSA 私钥从 YAML 移出**: auth-service 的 RSA 私钥移到独立文件或环境变量
- [x] **5.4 JWT secrets 改为安全随机值**: 替换所有 changeme 占位符（auth-service、APISIX）

## Phase 6: 前端修复

- [x] **6.1 前端 ID 类型 string 化**: 所有 TypeScript ID 字段从 `number` 改为 `string`（30+ 处，types/ + views/）
- [x] **6.2 Vite 代理规则修复**: 移除 `/api/identity` → 8888，添加 `/api/masterdata` → 8889，添加 `/api/v1` → 8891

## Completed
- [x] CRITICAL issues identified from review
