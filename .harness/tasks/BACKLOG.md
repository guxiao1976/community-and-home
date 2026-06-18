# 待办索引

> Loop 启动时读取本文件，按优先级调度任务。
> 格式：`- [标题](文件.md) — 服务, 优先级, 状态, 来源`
>
> 来源类型：`human`=人安排的战略任务 | `qa`=QA 传感器检测 | `review`=Review 发现 | `sensor`=自动传感器
>
> 最后更新：2026-06-18

## P0 — 立即处理（阻塞性问题）

（暂无）

## P1 — 本周

- [pipeline handler 响应格式统一为 responsex.Response](task-2026-06-16-001.md) — moderation-service, P1, open, review
- [gRPC 客户端超时配置（AiModelRpc/MasterDataRpc 三层超时对齐）](task-2026-06-16-005.md) — moderation-service, P1, open, review
- [pipeline 核心引擎单元测试（executor.go + config.go）](task-2026-06-16-006.md) — moderation-service, P1, open, review
- [REST API int64 ID 添加 json:\",string\" 标签](task-2026-06-16-007.md) — ai-model-service, P1, open, review
- [软删除后清除 ModelManager adapter 缓存](task-2026-06-16-008.md) — ai-model-service, P1, open, review


## P2 — 本月

- [HealthCheck/CallModel/CallModelBatch Stub 实现](task-2026-06-16-009.md) — ai-model-service, P2, open, human
- [MaxTokens 字段写入数据库（当前硬编码返回 0）](task-2026-06-16-010.md) — ai-model-service, P2, open, review
- [审计日志完善（user_id 从 JWT 提取）](task-2026-06-16-012.md) — master-data-service, P2, open, human
- [实现 Outbox 模式（outbox_messages 表 + MQ 集成）](task-2026-06-16-013.md) — master-data-service, P2, open, human


## P3 — 以后

（暂无）

## 进行中

- [PipelineId path 参数解析修复（json tag → path tag）](task-2026-06-16-002.md) — moderation-service, P0, in_progress, review
- [update_pipeline_logic 恒真条件修复（空字符串意外覆盖已有值）](task-2026-06-16-003.md) — moderation-service, P0, in_progress, review
- [activate_pipeline 竞态修复（事务 + 批量更新替代逐条循环）](task-2026-06-16-004.md) — moderation-service, P0, in_progress, review


## 已阻塞

（暂无）

---

## 统计

| 优先级 | 数量 |
|:------:|:----:|
| P0 | 0 |
| P1 | 4 |
| P2 | 4 |
| P3 | 0 |
| 进行中 | 3 |
| 已阻塞 | 0 |
| **合计** | **11** |

| 服务 | 数量 |
|------|:----:|
| moderation-service | 6 |
| ai-model-service | 4 |
| master-data-service | 2 |

