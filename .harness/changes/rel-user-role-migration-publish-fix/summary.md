# 变更总结 — rel-user-role-migration-publish-fix

> 2026-08-15，L 级 spec-pipeline 阶段 0-6 全流程（复用既有变更，流水线完善后重跑）

## 做什么

修复两个 backlog 缺口：
- **P0**：rel_user_role 生命周期三列（status/verified_at/expires_at）缺建库 SQL，从零建库报 MySQL 1054
- **P1**：移动端寻失列表路径不匹配（/lost-found vs /lostfound）+ 三处静默吞错 + snake/camel 字段不匹配

## 执行结果

| 阶段 | 结果 |
|------|------|
| 0-1 需求分析 | 复用既有 D1-D15 决策，澄清只问 3 个执行载体未决点 |
| 2 需求评审 | 第 1 轮 2/4（validity+clarity 抓到 REQ-P0-2 矛盾）→ 反馈注入 → 第 2 轮 4/4 收敛 |
| 3 架构设计+设计评审 | 10 tasks，设计评审 2 视角（data-model REVISION→回架构→APPROVED） |
| 4 Proto | 0 变更，跳过 |
| 5 编码 | permission-service 18项全绿；web/mobile 字段对齐后全绿 |
| 6 归档 | 沙箱 fs 失效，summary/INDEX 由 Owner 手动补 |

## 评审结论

- 需求评审：4/4 APPROVED（2 轮收敛）
- 执行评审：permission-service 18 PASS 0 FAIL；web/mobile 2/3 PASS → design-biz 报 CRITICAL（snake/camel）→ 字段对齐修复后 1/1 PASS

## 核心决策（D1-D15 + 3 个澄清决策）

- 003 迁移 information_schema 幂等 guard，status DEFAULT=2（保留「有 grant 即活跃」）
- created_time→created_at 仅从零库（接受存量分叉），id 保留 AUTO_INCREMENT
- 只修 notice.vue 三处 catch，其余静默点记 backlog
- 临时构造三态库验证 003（非真实生产库）

## 暴露的流程 gap（「测试流水线」核心产出）

1. 评审报告/归档落盘在沙箱失效（spec-pipeline fs=null）
2. harness-pipeline 执行评审无 CRITICAL 一票否决（design-biz CRITICAL 被 2/3 放行，靠 Owner 补修）
3. spec 把「MySQL 三态库验证」定义成编码任务，Go 管线无法执行

## 交付物

- proposal.md + 3 capability spec + design.md + tasks.md
- services/permission-service/migration/003_add_role_lifecycle.sql
- docs/specs/migration.sql（rel_user_role 段修正）
- web/mobile 字段对齐（community.ts + notice.vue + notice-browse/detail.vue + notice.spec.ts）
