# 待办索引

> Loop 启动时读取本文件，按优先级调度任务。
> 格式：`- [标题](文件.md) — 服务, 优先级, 状态, 来源`
>
> 来源类型：`human`=人安排的战略任务 | `qa`=QA 传感器检测 | `review`=Review 发现 | `sensor`=自动传感器
>
> 最后更新：2026-08-16

## P0 — 立即处理（阻塞性问题）

（暂无）

## P1 — 本周

- [修复 harness-spec-pipeline 需求评审盲循环（REVISION 反馈未注入分析师）](task-2026-08-14-002.md) — harness, P1, review, review
- ~~[通用图文发布组件重构（content_posts 通用化 + 内容级审核 + Kafka）](task-2026-08-16-001.md) — community-hub-service, P1, done, human~~（2026-08-16 完成；数据依赖未验证项见任务执行记录）


## P2 — 本月

- [web/mobile 其余静默吞错点清理](task-2026-08-15-006.md) — web/mobile, P2, open, human
- [设计一致性: auth-service model 列未覆盖标准迁移源](task-2026-08-15-011.md) — auth-service, P2, open, human
- [设计一致性: master-data-service model 列未覆盖标准迁移源](task-2026-08-15-012.md) — master-data-service, P2, open, human
- [设计一致性: moderation-service model 列未覆盖标准迁移源](task-2026-08-15-013.md) — moderation-service, P2, open, human
- [设计一致性: permission-service model 列未覆盖标准迁移源](task-2026-08-15-014.md) — permission-service, P2, open, human
- [设计一致性: user-service model 列未覆盖标准迁移源](task-2026-08-15-015.md) — user-service, P2, open, human
- [小区/村行政区划层级统筹（城市四级 vs 农村三级）](task-2026-08-15-016.md) — master-data-service, P2, open, human


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
| P1 | 2 |
| P2 | 7 |
| P3 | 1 |
| 进行中 | 1 |
| 已阻塞 | 0 |
| **合计** | **11** |

| 服务 | 数量 |
|------|:----:|
| moderation-service | 1 |
| community-hub-service | 1 |
| auth-service | 1 |
| master-data-service | 2 |
| web | 1 |
| web/mobile | 1 |
| harness | 1 |
| permission-service | 2 |
| user-service | 1 |

