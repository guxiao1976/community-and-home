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

| 我需要了解... | 去这里 |
|--------------|--------|
| 工程结构 / 服务分层 / 中间件 | [`.harness/rules/工程结构.md`](.harness/rules/工程结构.md) |
| Proto 管理规则 / 变更流程 | [`.harness/rules/Proto管理规范.md`](.harness/rules/Proto管理规范.md) |
| 编码规范 / 硬性约束 / 提交前检查 | [`.harness/rules/项目编码规范.md`](.harness/rules/项目编码规范.md) |
| 项目知识（架构/业务/数据） | [`.harness/knowledge/INDEX.md`](.harness/knowledge/INDEX.md) |
| 踩过的坑 / 经验记忆 | [`.harness/knowledge/memory/MEMORY.md`](.harness/knowledge/memory/MEMORY.md) |
| 当前待办 / 任务队列 | [`.harness/tasks/BACKLOG.md`](.harness/tasks/BACKLOG.md) |
| 历史变更追溯 | [`.harness/changes/INDEX.md`](.harness/changes/INDEX.md) |
| AI 团队工作流程 / 工具选择 | [`docs/specs/ai-dev-team-design.md`](docs/specs/ai-dev-team-design.md) |
| 执行日志 / 流程遵守 | [`docs/specs/execution-log.md`](docs/specs/execution-log.md) |

## 子 Claude 实例索引

| 中文名 | 目录 | 职责范围 |
|--------|------|---------|
| 用户服务 | `services/user-service/` | 用户服务（API+RPC 双层） |
| 认证服务 | `services/auth-service/` | 认证服务（API+RPC 双层，AT+RT 双 Token） |
| 权限服务 | `services/permission-service/` | 权限服务（API+RPC 双层，RBAC） |
| 主数据服务 | `services/master-data-service/` | 主数据服务（API+RPC+Cron） |
| 审核服务 | `services/moderation-service/` | 内容审核服务（API+RPC 双层） |
| AI模型服务 | `services/ai-model-service/` | AI 模型服务（Go+Python 混合） |
| 文件服务 | `services/file-service/` | 文件服务（MinIO 上传/下载） |
| 社区枢纽服务 | `services/community-hub-service/` | 社区枢纽服务（通知/联络/寻失） |
| 监控服务 | `services/monitoring-service/` | 运行监控服务（API-only） |
| `api-proto/` | `api-proto/` | API 契约定义、Proto 代码生成 |
| `common/` | `common/` | Go 共享库（v2），10 个工具包 |
| 前端 | `web/` | 前端层入口（管理后台 + 移动端） |

## 知识图谱上下文

每个子 Claude 实例的 `docs/graph-context.md` 由 Neo4j 知识图谱自动生成，包含：服务依赖、REST 路由、gRPC 接口、数据库表、前端消费方、实体血缘（Proto→Go→DB）。该文件由 `graph-sync.sh` 自动刷新，请勿手动编辑。

```bash
bash .harness/scripts/graph-query.sh <service-name>   # 查询图谱
```

## 7 条硬性约束

| # | 规则 | 详见 |
|---|------|------|
| 1 | Proto 定义统一在 `api-proto/`，服务间通信仅 gRPC | [Proto管理规范](.harness/rules/Proto管理规范.md) |
| 2 | Proto 变更仅全局 Claude 执行，子 Claude 禁止修改 api-proto/ | [Proto管理规范](.harness/rules/Proto管理规范.md) |
| 3 | Snowflake ID → Proto `[jstype=JS_STRING]` + Go `json:",string"` + TS `string` | [项目编码规范 §5](.harness/rules/项目编码规范.md) |
| 4 | 提交前必须 `bash .harness/skills/qa/scripts/harness-checks.sh --service <name>`，FAIL 不可提交 | [项目编码规范 §6](.harness/rules/项目编码规范.md) |
| 5 | 密钥在 `.env`，服务入口用 `configx.MustLoad` | [项目编码规范 §7](.harness/rules/项目编码规范.md) |
| 6 | 修改 `common/` 需全局评估影响 | [项目编码规范 §3](.harness/rules/项目编码规范.md) |
| 7 | 收到需求后首条响应必须输出路径选择（路径+理由+涉及服务），跨服务需求走 OpenSpec → `harness-pipeline.js`，禁止用外部技能替代 | [owner-agent.md §路径选择](.harness/agents/owner-agent.md) |

## 常用命令

### 基础设施
```bash
docker compose up -d    # 启动中间件
docker compose down     # 停止中间件
bash scripts/start.sh   # 启动所有服务
bash scripts/stop.sh    # 停止所有服务
bash scripts/status.sh  # 查看服务状态
```

### Proto 管理
```bash
cd api-proto && make generate    # 生成代码
cd api-proto && make lint        # 规范检查
cd api-proto && make ci          # lint + breaking-check + generate
```

### Go 构建/测试
```bash
cd services/<name> && go build ./...
cd services/<name> && go test ./...
bash .harness/skills/qa/scripts/harness-checks.sh --service <name>   # 15 项机械化检查
```

### 任务管理
```bash
bash .harness/scripts/harness-tasks.sh list                  # 列出所有待办
bash .harness/scripts/harness-tasks.sh list --priority P0    # 只看 P0
bash .harness/scripts/harness-tasks.sh scan --auto-create    # 传感器扫描+自动建任务
bash .harness/scripts/harness-tasks.sh stats                 # 统计概览
```

### 前端
```bash
cd web/pc && npm run dev         # Vite dev server
cd web/pc && npm run build       # Type-check + build
cd web/pc && npm run test:unit   # Vitest
```

---

## 硬性约束：工具调用失败处理

### 规则

**连续 2 次相同的工具调用失败后，必须立即停止并诊断。**

**触发条件**：
- 工具名称相同
- 错误类型相同（如 `InputValidationError`）
- 参数完全相同

### 强制流程

```
第 1 次失败 → 记录
第 2 次失败 → 记录
第 3 次尝试前 → 必须执行以下步骤：
  1. 停止当前方法
  2. 输出诊断：
     - 我调用了什么工具
     - 传递了什么参数
     - 收到什么错误
     - 工具定义是什么（从系统提示引用）
     - 根本原因分析
  3. 尝试完全不同的方法
```

### 禁止行为

❌ 连续 3 次以上相同的工具调用
❌ "也许这次会成功"的重复尝试
❌ 忽略错误信息中的提示

### 案例

#### ❌ 错误（违反规则）

```
TaskList(command="bash ...") → InputValidationError
TaskList(command="bash ...") → InputValidationError
TaskList(command="bash ...") → 违反规则！
```

#### ✅ 正确（遵守规则）

```
TaskList(command="bash ...") → InputValidationError
TaskList(command="bash ...") → InputValidationError

[诊断]
- 工具：TaskList
- 我的调用：TaskList(command="bash ...", description="...")
- 错误：unexpected parameter 'command'
- 工具定义：TaskList() — 不接受任何参数
- 根因：我传递了工具不支持的参数
- 新方法：分两步
  1. TaskList() — 查看当前任务
  2. Bash("bash script.sh") — 单独执行命令

[执行新方法]
TaskList() → 成功 ✅
```

### 参考

- 熔断器设计：`.harness/docs/circuit-breaker.md`
- 问题案例：2026-06-21 会话（30+ 次 TaskList 死循环）
