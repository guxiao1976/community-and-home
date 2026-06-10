# 项目知识索引

> 静态知识（已沉淀的架构、业务、数据）— 和 `memory/`（过程性经验）互补。
> Agent 需要理解"系统是什么样的"时从这里按图索骥。

## 按服务

| 服务 | 设计文档 | 知识图谱 | 关键信息 |
|------|---------|---------|------|
| user-service | [docs/design.md](../../services/user-service/docs/design.md) | [graph-context.md](../../services/user-service/docs/graph-context.md) | 用户 CRUD、手机号加密、小区归属 |
| auth-service | [docs/design.md](../../services/auth-service/docs/design.md) | [graph-context.md](../../services/auth-service/docs/graph-context.md) | AT/RT 双 Token、SMS 验证码、RSA |
| permission-service | [docs/design.md](../../services/permission-service/docs/design.md) | [graph-context.md](../../services/permission-service/docs/graph-context.md) | RBAC 权限模型 |
| file-service | [docs/design.md](../../services/file-service/docs/design.md) | [graph-context.md](../../services/file-service/docs/graph-context.md) | MinIO 上传/下载 |
| master-data-service | [docs/design.md](../../services/master-data-service/docs/design.md) | [graph-context.md](../../services/master-data-service/docs/graph-context.md) | 主数据管理 |
| moderation-service | [docs/design.md](../../services/moderation-service/docs/design.md) | [graph-context.md](../../services/moderation-service/docs/graph-context.md) | AI 内容审核 |
| ai-model-service | [docs/design.md](../../services/ai-model-service/docs/design.md) | [graph-context.md](../../services/ai-model-service/docs/graph-context.md) | Go+Python AI 模型 |
| community-hub-service | [docs/design.md](../../services/community-hub-service/docs/design.md) | [graph-context.md](../../services/community-hub-service/docs/graph-context.md) | 通知/联络/寻失 |
| monitoring-service | [docs/design.md](../../services/monitoring-service/docs/design.md) | [graph-context.md](../../services/monitoring-service/docs/graph-context.md) | TCP/Docker/AI 三层监控 |

## 按知识类型

### 架构知识

| 文档 | 说明 |
|------|------|
| `.harness/rules/工程结构.md` | Go Workspace、服务分层、Proto 组织、中间件 IP |
| `docs/specs/ai-dev-team-design.md` | AI 五层架构、完整流水线设计 |
| `docs/specs/architecture-audit.md` | 90+ 发现的全量架构审查报告 |

### 业务知识

| 文档 | 领域 |
|------|------|
| `docs/specs/user-design.md` | 用户体系 |
| `docs/specs/auth-design.md` | 认证体系 |
| `docs/specs/permission.md` | 权限体系 |
| `docs/specs/community-dynamics-design.md` | 社区动态 |
| `docs/specs/monitoring-module-design.md` | 监控模块 |
| `docs/specs/notify.md` | 通知系统 |

### 数据模型

各服务 `docs/design.md` 中包含完整的 DDL 和 ER 关系。`services/*/docs/graph-context.md`（Neo4j 自动生成）包含数据表结构和实体血缘。

### 经验记忆

见本目录下 `memory/MEMORY.md` — 踩过的坑和编码约束。

## 使用方式

- **编码前**: 读目标服务的 `design.md` + `graph-context.md`
- **架构决策前**: 读 `rules/工程结构.md` + `docs/specs/architecture-audit.md`
- **理解业务前**: 读对应领域的 `docs/specs/` 设计文档
- **避免踩坑**: 读 `memory/MEMORY.md`，按触发词匹配
