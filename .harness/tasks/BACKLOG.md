# 待办索引

> Loop 启动时读取本文件，按优先级调度任务。
> 格式：`- [标题](文件.md) — 服务, 优先级, 状态, 来源`
>
> 来源类型：`human`=人安排的战略任务 | `qa`=QA 传感器检测 | `review`=Review 发现 | `sensor`=自动传感器
>
> 最后更新：2026-06-19

## P0 — 立即处理（阻塞性问题）

（无）


## P1 — 本周

- [QA FAIL: response_wrap](task-2026-06-16-011.md) — all, P1, completed, qa
- [QA FAIL: graph_freshness — graph is stale (last sync: 33h ago, latest commit is newer) — run: bash .harness/scripts/graph-sync.sh](task-2026-06-17-002.md) — all, P1, completed, qa
- [QA FAIL: graph_freshness — graph is stale (last sync: 36h ago, latest commit is newer) — run: bash .harness/scripts/graph-sync.sh](task-2026-06-17-003.md) — all, P1, completed, qa
- [QA FAIL: graph_freshness — graph is stale (last sync: 53h ago, latest commit is newer) — run: bash .harness/scripts/graph-sync.sh](task-2026-06-18-001.md) — all, P1, completed, qa


## P2 — 本月

- [同步知识图谱（29h 未更新）](task-2026-06-17-001.md) — global, P2, completed, sensor


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
| P0 | 3 |
| P1 | 9 |
| P2 | 5 |
| P3 | 0 |
| 进行中 | 0 |
| 已阻塞 | 0 |
| **合计** | **17** |

| 服务 | 数量 |
|------|:----:|
| moderation-service | 6 |
| all | 4 |
| master-data-service | 2 |
| global | 1 |
| ai-model-service | 4 |

