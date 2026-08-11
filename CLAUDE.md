# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 角色定位

这是 **全局架构协调层**。职责：

- **api-proto 管理**：所有服务接口的 Proto 定义、代码生成、破坏性变更检测
- **跨服务功能协调**：涉及多个微服务的大型功能开发、服务间接口对齐
- **全局规范定义**：错误码规范、响应格式、公共库版本策略
- **架构决策**：服务拆分/合并、技术选型、依赖方向

**本实例不负责单个服务的具体实现**。当需要开发具体功能时，切换到对应服务的子 Claude 实例。

**启动后首先阅读** [`.harness/agents/owner-agent.md`](.harness/agents/owner-agent.md) — 它告诉你收到需求后的完整调度流程、何时加载哪个 Skill、各阶段的门禁条件。

## 快速索引

> 完整索引见 [`.harness/knowledge/INDEX.md`](.harness/knowledge/INDEX.md)

| 我需要了解... | 去这里 |
|--------------|--------|
| 编码规范 / 硬性约束 | [`.harness/rules/项目编码规范.md`](.harness/rules/项目编码规范.md) |
| 跨服务设计规范 | [`docs/specs/INDEX.md`](docs/specs/INDEX.md) |
| RBAC 权限设计 | [`docs/specs/rbac-design.md`](docs/specs/rbac-design.md) |
| Proto 管理规则 | [`.harness/rules/Proto管理规范.md`](.harness/rules/Proto管理规范.md) |
| 踩过的坑 / 经验记忆 | [`.harness/knowledge/memory/MEMORY.md`](.harness/knowledge/memory/MEMORY.md) |
| 当前待办 | [`.harness/tasks/BACKLOG.md`](.harness/tasks/BACKLOG.md) | |

## 子服务

详见 [`.harness/knowledge/INDEX.md`](.harness/knowledge/INDEX.md) 和 [`.harness/registry/services.json`](.harness/registry/services.json)。

```bash
bash .harness/scripts/graph-query.sh <service-name>   # 知识图谱查询
```

## 7 条硬性约束

| # | 规则 | 详见 |
|---|------|------|
| 1 | Proto 定义统一在 `api-proto/`，服务间通信仅 gRPC | [Proto管理规范](.harness/rules/Proto管理规范.md) |
| 2 | Proto 变更仅全局 Claude 执行，子 Claude 禁止修改 api-proto/ | [Proto管理规范](.harness/rules/Proto管理规范.md) |
| 3 | Snowflake ID → Proto `[jstype=JS_STRING]` + Go `json:",string"` + TS `string` | [项目编码规范 §5](.harness/rules/项目编码规范.md) |
| 3.1 | 时间字段统一 `created_at`/`updated_at`/`deleted_at`，全链路一致 | [项目编码规范 §5.1](.harness/rules/项目编码规范.md) |
| 4 | 提交前必须 `bash .harness/skills/qa/scripts/harness-checks.sh --service <name>`，FAIL 不可提交 | [项目编码规范 §6](.harness/rules/项目编码规范.md) |
| 5 | 密钥在 `.env`，服务入口用 `configx.MustLoad` | [项目编码规范 §7](.harness/rules/项目编码规范.md) |
| 6 | 修改 `common/` 需全局评估影响 | [项目编码规范 §3](.harness/rules/项目编码规范.md) |
| 7 | 收到需求后首条响应必须输出路径选择（路径+理由+涉及服务），跨服务需求走 OpenSpec → `.harness/workflows/harness-pipeline.js`，禁止用外部技能替代 | [owner-agent.md §路径选择](.harness/agents/owner-agent.md) |

## 常用命令

```bash
# 启动全栈（按顺序）
docker compose up -d && bash scripts/init-databases.sh && bash scripts/start.sh && bash scripts/start-frontend.sh

# 停止 / 状态
bash scripts/stop.sh && bash scripts/status.sh

# Proto 管理
cd api-proto && make ci          # lint + breaking-check + generate

# Go 构建 / 测试 / 门禁
cd services/<name> && go build ./... && go test ./...
bash .harness/skills/qa/scripts/harness-checks.sh --service <name>

# 任务管理
bash .harness/scripts/harness-tasks.sh list
bash .harness/scripts/harness-tasks.sh scan --auto-create

# 前端
bash scripts/start-frontend.sh --stop
cd web/pc && npm run dev && npm run test:unit
```

---

## 硬性约束：工具调用失败处理

**连续 2 次相同的工具调用失败后，必须立即停止并诊断根因，换完全不同的方法。** 禁止连续 3 次以上相同的工具调用。详见 `.harness/docs/circuit-breaker.md`。
