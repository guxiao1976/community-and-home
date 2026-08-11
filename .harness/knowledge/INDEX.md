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
| `docs/archived/specs/architecture-audit.md` | 90+ 发现的全量架构审查报告（2026-06 快照，已归档） |

### Harness 方法论

| 文档 | 说明 |
|------|------|
| `docs/specs/harness-construction-guide.md` | Harness 构建说明书——可复用到其他项目的通用方法论 |

### 业务知识

> 服务级业务知识见各服务 `docs/design.md`。跨服务方案见 `docs/specs/`。

| 文档 | 领域 |
|------|------|
| `docs/specs/rbac-design.md` | RBAC 权限体系（角色/权限/认证/生命周期） |
| 各服务 `docs/design.md` | 用户/认证/权限/文件/主数据/审核等单服务设计 |
| `docs/archived/specs/` | 早期愿景/已归档设计（user/auth/permission/community-dynamics/notify 等） |

### 业务流程

见 [`.harness/knowledge/business-flows.md`](business-flows.md) — 7 条端到端业务流程（用户入驻、Token 生命周期、内容审核、审批工作流、RBAC、文件上下传、社区枢纽），含状态机和已知空白。

### 数据模型

各服务 `docs/design.md` 中包含完整的 DDL 和 ER 关系。`services/*/docs/graph-context.md`（Neo4j 自动生成）包含数据表结构和实体血缘。

### 经验记忆

见本目录下 `memory/MEMORY.md` — 踩过的坑和编码约束。

## 知识图谱

各服务的 `docs/graph-context.md` 由 **Neo4j 知识图谱自动生成**，包含服务依赖、REST 路由、gRPC 接口、数据库表、前端消费方、实体血缘（Proto→Go→DB）。每次 `graph-sync.sh` 后自动刷新，请勿手动编辑。

| 脚本 | 作用 | 位置 |
|------|------|------|
| `graph-sync.sh` | 同步项目源码到 Neo4j，生成所有 graph-context.md | `.harness/scripts/` |
| `graph-query.sh` | 按需查询单个服务的知识图谱 | `.harness/scripts/` |
| `graph-populator/` | Go 源码解析器（Proto/Go/TypeScript） | `.harness/scripts/graph-populator/` |

QA 机械化检查第 9 项会验证图谱新鲜度（`graph_freshness`），确保 graph-context.md 与代码同步。

## 使用方式

- **编码前**: 读目标服务的 `design.md` + `graph-context.md`（图谱自动生成）
- **架构决策前**: 读 `rules/工程结构.md` + `docs/specs/INDEX.md`
- **理解业务前**: 读对应服务的 `docs/design.md` 或 `docs/specs/`
- **避免踩坑**: 运行 `bash .harness/scripts/knowledge-load.sh --service <服务> --task "<任务描述>"` 加载相关记忆
- **图谱过期**: 运行 `bash .harness/scripts/graph-sync.sh`
