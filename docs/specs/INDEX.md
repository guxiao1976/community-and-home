# docs/specs/ 索引

> **定位**：跨服务规范 + 全局设计方案的存放处。
> 单服务设计文档在 `services/<name>/docs/design.md`；服务实现细节随代码演化。
> 过时/早期愿景文档归档在 `docs/archived/specs/`。

## 活跃文档

| 文件 | 内容 | 读者 |
|------|------|------|
| [`rbac-design.md`](rbac-design.md) | RBAC 权限体系完整方案：数据模型、权限校验链路、缓存策略、前端对接、验收清单 | 所有服务 + 前端 |
| [`common-technical-spec.md`](common-technical-spec.md) | Community Common 库技术规范：错误码、HTTP/gRPC 响应格式、Proto 管理、测试规范 | 所有 Go 服务 |
| [`ai-dev-team-design.md`](ai-dev-team-design.md) | AI 软件开发团队三层架构、角色化 Agent 设计、知识持久化 | 全局 Claude |
| [`harness-construction-guide.md`](harness-construction-guide.md) | Harness Engineering 方法论：Rules/Skills/Knowledge/Changes 四支柱 | 全局 Claude |
| [`tool-selection-standard.md`](tool-selection-standard.md) | 每次需求到达时的工具选择决策树（直接改 / Dev Agent / OpenSpec / Workflow / Ralph） | 全局 Claude |

## 归档文档（docs/archived/specs/）

以下文档已归档，因其描述的服务未实现（愿景）、被 services/ 下 design.md 取代、或为一次性过程记录：

- **未实现服务愿景**：`approve.md`（审批中心）、`audit.md`（审计中心）、`risk.md`（风控中心）、`notify.md`（通知中心）
- **服务已建成/改名**：`community-dynamics-design.md`（实际为 community-hub-service）、`monitoring-module-design.md`（已建成有 design.md）
- **早期草稿（被 design.md 取代）**：`permission.md`、`user.md`、`auth.md`、`auth-design.md`、`user-design.md`、`user-auth-test-plan.md`
- **历史审查/过程性**：`architecture-audit.md`（2026-06 快照）、`execution-log.md`、`task-1.md`

---

## 维护约定

- **新规范**：跨服务/全局的 → 加到本目录；单服务的 → 加到 `services/<name>/docs/design.md`
- **归档标准**：服务未实现、文档被取代、内容过时 → 移到 `docs/archived/specs/`
- **命名规范**：小写连字符，如 `rbac-design.md`、`common-technical-spec.md`
