# 知识图谱上下文 — monitoring-service

> 自动生成于 2026-08-14 16:15:53 | 数据源: Neo4j 知识图谱 | 每次 `graph-sync.sh` 后刷新

## 服务标识

| 属性 | 值 |
|------|-----|
| 名称 | monitoring-service |
| 语言 | go |
| 端口 (gRPC) | None |
| 端口 (API)  | 8886 |

## 服务依赖

无外部依赖

## 被依赖方

无服务依赖本服务

## REST API 路由

| 方法 | 路径 |
|------|------|
| Get | /api/monitoring/health |

## gRPC 接口

无 gRPC 接口

## 数据库表

无数据库表

## 前端消费方

| 方法 | URL | 文件 |
|------|-----|------|
| GET | /api/monitoring/health | web/pc/src/api/monitoring.ts |

## 实体血缘（Proto → Go → DB）

无实体血缘数据

---
*此文件由 graph-sync.sh 自动生成，请勿手动编辑。*
