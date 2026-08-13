# 待办索引

> Loop 启动时读取本文件，按优先级调度任务。
> 格式：`- [标题](文件.md) — 服务, 优先级, 状态, 来源`
>
> 来源类型：`human`=人安排的战略任务 | `qa`=QA 传感器检测 | `review`=Review 发现 | `sensor`=自动传感器
>
> 最后更新：2026-08-13

## P0 — 立即处理（阻塞性问题）

- [PipelineId path 参数解析修复（json tag → path tag）](task-2026-06-16-002.md) — moderation-service, P0, closed, review
- [update_pipeline_logic 恒真条件修复（空字符串意外覆盖已有值）](task-2026-06-16-003.md) — moderation-service, P0, closed, review
- [activate_pipeline 竞态修复（事务 + 批量更新替代逐条循环）](task-2026-06-16-004.md) — moderation-service, P0, closed, review
- [access-data-permission 后续 Wave：user/community-hub/web-mobile/集成验收](task-2026-08-12-nextwave.md) — multi, P0, completed, human


## P1 — 本周

- [pipeline handler 响应格式统一为 responsex.Response](task-2026-06-16-001.md) — moderation-service, P1, closed, review
- [gRPC 客户端超时配置（AiModelRpc/MasterDataRpc 三层超时对齐）](task-2026-06-16-005.md) — moderation-service, P1, closed, review
- [pipeline 核心引擎单元测试（executor.go + config.go）](task-2026-06-16-006.md) — moderation-service, P1, closed, review
- [REST API int64 ID 添加 json:\",string\" 标签](task-2026-06-16-007.md) — ai-model-service, P1, closed, review
- [软删除后清除 ModelManager adapter 缓存](task-2026-06-16-008.md) — ai-model-service, P1, closed, review
- [QA FAIL: response_wrap](task-2026-06-16-011.md) — all, P1, closed, qa
- [QA FAIL: graph_freshness — graph is stale (last sync: 69h ago, latest commit is newer)](task-2026-07-11-001.md) — all, P1, closed, qa
- [QA FAIL: graph_freshness — graph is stale (last sync: 97h ago, latest commit is newer)](task-2026-07-12-001.md) — all, P1, closed, qa
- [QA FAIL: graph_freshness — graph is stale (last sync: 645h ago, latest commit is newer)](task-2026-08-04-001.md) — all, P1, closed, qa
- [QA FAIL: graph_freshness — graph is stale (last sync: 645h ago, latest commit is newer)](task-2026-08-04-002.md) — all, P1, closed, qa
- [QA FAIL: graph_freshness — graph is stale (last sync: 645h ago, latest commit is newer)](task-2026-08-04-003.md) — all, P1, closed, qa
- [QA FAIL: graph_freshness — graph is stale (last sync: 645h ago, latest commit is newer)](task-2026-08-04-004.md) — all, P1, closed, qa
- [QA FAIL: graph_freshness — graph is stale (last sync: 646h ago, latest commit is newer)](task-2026-08-04-005.md) — all, P1, closed, qa
- [QA FAIL: graph_freshness — graph is stale (last sync: 646h ago, latest commit is newer)](task-2026-08-04-006.md) — all, P1, closed, qa
- [QA FAIL: graph_freshness — graph is stale (last sync: 646h ago, latest commit is newer)](task-2026-08-04-007.md) — all, P1, closed, qa
- [QA FAIL: graph_freshness — graph is stale (last sync: 648h ago, latest commit is newer)](task-2026-08-04-008.md) — all, P1, closed, qa
- [harness-pipeline 与 superpowers 流程衔接规则待收敛](task-2026-08-13-001.md) — global, P1, open, human


## P2 — 本月

- [HealthCheck/CallModel/CallModelBatch Stub 实现](task-2026-06-16-009.md) — ai-model-service, P2, closed, human
- [MaxTokens 字段写入数据库（当前硬编码返回 0）](task-2026-06-16-010.md) — ai-model-service, P2, closed, review
- [审计日志完善（user_id 从 JWT 提取）](task-2026-06-16-012.md) — master-data-service, P2, closed, human
- [实现 Outbox 模式（outbox_messages 表 + MQ 集成）](task-2026-06-16-013.md) — master-data-service, P2, closed, human
- [同步知识图谱（29h 未更新）](task-2026-06-17-001.md) — global, P2, closed, sensor


## P3 — 以后

（暂无）

## 进行中

（暂无）

## 已阻塞

（暂无）

---

## 统计

| 优先级 | 数量 |
|:------:|:----:|
| P0 | 4 |
| P1 | 17 |
| P2 | 5 |
| P3 | 0 |
| 进行中 | 0 |
| 已阻塞 | 0 |
| **合计** | **26** |

| 服务 | 数量 |
|------|:----:|
| moderation-service | 6 |
| all | 11 |
| all | 3 |
| master-data-service | 2 |
| global | 2 |
| ai-model-service | 4 |
| multi | 1 |

