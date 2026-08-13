# 知识图谱上下文 — auth-service

> 自动生成于 2026-08-13 07:57:06 | 数据源: Neo4j 知识图谱 | 每次 `graph-sync.sh` 后刷新

## 服务标识

| 属性 | 值 |
|------|-----|
| 名称 | auth-service |
| 语言 | go |
| 端口 (gRPC) | 8083 |
| 端口 (API)  | 8881 |

## 服务依赖

| 依赖服务 | 依赖类型 |
|---------|---------|
| master-data-service | gRPC |
| user-service | gRPC |

## 被依赖方

无服务依赖本服务

## REST API 路由

| 方法 | 路径 |
|------|------|
| Post | /api/auth/login |
| Post | /api/auth/login/sms |
| Post | /api/auth/logout |
| Get | /api/auth/public-key |
| Post | /api/auth/register |
| Post | /api/auth/sms/send |
| Post | /api/auth/token/refresh |

## gRPC 接口

| RPC 方法 | 输入消息 | 输出消息 |
|---------|---------|---------|
| Login | LoginRequest | LoginResponse |
| LoginSms | LoginSmsRequest | LoginResponse |
| Logout | LogoutRequest | LogoutResponse |
| RefreshToken | RefreshTokenRequest | RefreshTokenResponse |
| Register | RegisterRequest | RegisterResponse |
| ValidateToken | ValidateTokenRequest | ValidateTokenResponse |

## 数据库表

| 表名 | 列 |
|------|-----|
| auth_credential | updated_at (datetime), created_at (datetime), updated_time (datetime), created_time (datetime), credential (varchar), identifier (varchar), identity_type (varchar), user_id (bigint), id (bigint) |

## 前端消费方

| 方法 | URL | 文件 |
|------|-----|------|
| POST | /api/auth/logout | web/pc/src/api/identity.ts |
| POST | /api/auth/token/refresh | web/mobile/src/api/identity.ts |
| POST | /api/auth/token/refresh | web/pc/src/api/identity.ts |
| POST | /api/auth/sms/send | web/mobile/src/api/identity.ts |
| POST | /api/auth/sms/send | web/pc/src/api/identity.ts |
| POST | /api/auth/register | web/mobile/src/api/identity.ts |
| POST | /api/auth/register | web/pc/src/api/identity.ts |
| POST | /api/auth/login/sms | web/mobile/src/api/identity.ts |
| POST | /api/auth/login/sms | web/pc/src/api/identity.ts |
| POST | /api/auth/login | web/mobile/src/api/identity.ts |
| POST | /api/auth/login | web/pc/src/api/identity.ts |

## 实体血缘（Proto → Go → DB）

无实体血缘数据

---
*此文件由 graph-sync.sh 自动生成，请勿手动编辑。*
