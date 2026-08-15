# 待办索引

> Loop 启动时读取本文件，按优先级调度任务。
> 格式：`- [标题](文件.md) — 服务, 优先级, 状态, 来源`
>
> 来源类型：`human`=人安排的战略任务 | `qa`=QA 传感器检测 | `review`=Review 发现 | `sensor`=自动传感器
>
> 最后更新：2026-08-15

## P0 — 立即处理（阻塞性问题）

（暂无）

## P1 — 本周

- [修复 harness-spec-pipeline 需求评审盲循环（REVISION 反馈未注入分析师）](task-2026-08-14-002.md) — harness, P1, review, review


## P2 — 本月

- [需求冲突检测确定性脚本（扫 .harness/changes/ 同服务/同接口重叠）](task-2026-08-15-003.md) — harness, P2, open, human
- [web/mobile 其余静默吞错点清理](task-2026-08-15-006.md) — web/mobile, P2, open, human


## P3 — 以后

- [permission-service 兄弟表 AUTO_INCREMENT/created_time 结构不一致收敛](task-2026-08-15-005.md) — permission-service, P3, open, human


## 进行中

- [前端 TS 类型同步 proto：11 个滞后字段（web/common/types）](task-2026-08-14-001.md) — web, P1, in_progress, qa


## 已阻塞

（暂无）

---

## 统计

| 优先级 | 数量 |
|:------:|:----:|
| P0 | 0 |
| P1 | 1 |
| P2 | 2 |
| P3 | 1 |
| 进行中 | 1 |
| 已阻塞 | 0 |
| **合计** | **5** |

| 服务 | 数量 |
|------|:----:|
| web | 1 |
| web/mobile | 1 |
| harness | 2 |
| permission-service | 1 |

