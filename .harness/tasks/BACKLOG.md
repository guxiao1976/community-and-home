# 待办索引

> Loop 启动时读取本文件，按优先级调度任务。
> 格式：`- [标题](文件.md) — 服务, 优先级, 状态, 来源`
>
> 来源类型：`human`=人安排的战略任务 | `qa`=QA 传感器检测 | `review`=Review 发现 | `sensor`=自动传感器
>
> 最后更新：2026-08-15

## P0 — 立即处理（阻塞性问题）

- [P0: rel_user_role 生命周期三列缺建库 SQL（从零建库 1054）](task-2026-08-15-001.md) — permission-service, P0, open, human


## P1 — 本周

- [修复 harness-spec-pipeline 需求评审盲循环（REVISION 反馈未注入分析师）](task-2026-08-14-002.md) — harness, P1, review, review
- [P1: 移动端寻失列表路径不匹配（/lost-found vs /lostfound）导致 404 + 静默吞错](task-2026-08-15-002.md) — web/mobile, P1, open, human


## P2 — 本月

- [需求冲突检测确定性脚本（扫 .harness/changes/ 同服务/同接口重叠）](task-2026-08-15-003.md) — harness, P2, open, human


## P3 — 以后

（暂无）

## 进行中

- [前端 TS 类型同步 proto：11 个滞后字段（web/common/types）](task-2026-08-14-001.md) — web, P1, in_progress, qa


## 已阻塞

（暂无）

---

## 统计

| 优先级 | 数量 |
|:------:|:----:|
| P0 | 1 |
| P1 | 2 |
| P2 | 1 |
| P3 | 0 |
| 进行中 | 1 |
| 已阻塞 | 0 |
| **合计** | **5** |

| 服务 | 数量 |
|------|:----:|
| web | 1 |
| web/mobile | 1 |
| harness | 2 |
| permission-service | 1 |

