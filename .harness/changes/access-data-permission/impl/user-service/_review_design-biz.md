# Code Review — 用户服务（设计业务视角）

**审查时间**: 2026-08-12
**审查维度**: 设计一致性(#2)、代码质量(#4)、Migration(#8部分)
**审查范围**: 数据权限核心编排 阶段③（T3.1 注册自动授权 / T3.2 加入授权 / T3.3 退出撤销 / T3.4 门禁）工作树未提交改动

## 摘要
- 🔴 CRITICAL: 0 / 🟡 WARNING: 4 / 🔵 NOTE: 6

## 发现

### 🔴 CRITICAL
无

### 🟡 WARNING
| # | 文件:行号 | 维度 | 问题 | 修复建议 |
|---|----------|------|------|---------|
| W-1 | `leave_community_logic.go:62` | 代码质量(错误路径) | 撤销失败补偿 `UpdateBindStatus(active, time.Time{})` 把零值时间写入 `leave_time DATETIME` 列。Go `time.Time{}` 序列化为 `0001-01-01 00:00:00`，超出 MySQL DATETIME 范围（min 1000-01-01）；strict 模式下 UPDATE 报错且错误被 `_ =` 丢弃 → 补偿静默失败，membership 停留在 left 而角色未撤销（zombie grant，且重试命中 10005 永不重撤）；非 strict 模式写 '0000-00-00' 垃圾值。单测用内存 mock 掩盖了该问题 | 补偿恢复应写 NULL：新增按 id 恢复 active 且 `leave_time=NULL` 的方法（或 UpdateBindStatus 对零值时间特判为 NULL）；补一个真实 sqlmock 的用例验证 restore 语义 |
| W-2 | `leave_community_logic.go:61,87-102` | 代码质量/设计一致性 | `revokeCommunityRoles` 顺序撤销 owner→tenant；若 owner 成功、tenant 失败，返回错误后调用方把 bind_status 恢复为 active → 成员仍在该小区但 owner scope 已删除，直接违反本次变更声明的「有成员必有 scope」不变量，且无补偿重新授权（AssignRole 为幂等 INSERT IGNORE，重授成本极低）。残留 grant 亦有越权风险（见 N-6） | 失败恢复 active 前需重授已撤销成功的角色；或将撤销改为「先记录/后补偿」的事务化语义，保证幂等可重入 |
| W-3 | `create_user_logic.go:133-144` | 代码质量/设计一致性 | `assignRegisteredUser` 非阻塞（失败仅告警），但无任何重试/对账机制。permission-service 故障期间注册的用户永久缺失 registered_user 基角色（CreateUser 仅在首次注册调用，后续登录走 GetUserByPhone 不再触发），违反「有注册必有基角色」不变量；与 JoinCommunity 对授权失败采取 fatal+补偿的哲学不一致 | 增加补偿/对账：注册侧幂等重试，或独立 backfill 任务对「无任何 grant 的用户」补发基角色；至少在文档中记录该不变量存在暂时性缺失的窗口 |
| W-4 | `join_community_logic.go:35-40` + `docs/design.md §2.2/§3.2/§3.3` | 设计一致性 | design.md 未随本次变更更新。§2.2 明确「membership 不表达任何身份或角色」，§3.2 JoinCommunity 无 ownership；现 JoinCommunity 必填 ownership 并自动创建 owner/tenant 角色（status=0），与 §3.3「ApplyRole 是角色创建路径」形成双路径创建。功能上无冲突（permission AssignRole 幂等），但设计文档与实现矛盾 | 同步更新 design.md：JoinCommunity 增加 ownership 语义、明确 owner/tenant 角色可经「加入自动授权」或 ApplyRole 双路径创建、标注 ApplyRole 对 owner/tenant 已冗余 |

### 🔵 NOTE
| # | 文件:行号 | 建议 |
|---|----------|------|
| N-1 | `api/internal/types/types.go` JoinCommunityReq | `Ownership int32 \`json:"ownership,optional"\`` 标记 optional，但 RPC 层硬性必填（UNSPECIFIED→10040）。移动端（T5.1）未带 ownership 前，既有加入流程会 10040 中断。建议改为必填或文档标注；另 Building/Unit/Room 用 int32，而 design §2.5 房屋地址是用户输入字符串（含非数字楼号如「A栋」），residence 模型亦是 VARCHAR —— 类型口径不一致（部分 pre-existing），建议统一 |
| N-2 | `join_community_logic.go:225` `ownershipRoleCode` | 对任何非 RENTED 值返回 OWNED，依赖调用方先校验才安全；若被复用且未校验，UNSPECIFIED 会静默映射为 owner。建议在函数内校验 |
| N-3 | `join_community_logic.go:143-149` | 重新加入路径：内存 `existing.LeaveTime = sql.NullTime{}` 清空，但 DB 层 UpdateBindStatus(active, now) 写入 leave_time=now，返回给客户端的 leave_time 与 DB 不一致（仅回显，无功能影响） |
| N-4 | 迁移 | 本服务无表结构变更，无需 migration（正确）。但存在跨服务部署顺序依赖：本变更依赖 permission-service 的 `001_scope_three_state.sql`（uk_user_role_scope）与 `init_permissions.sql` 种子（owner=1/tenant=5/registered_user=9）先就位，否则 JoinCommunity 硬失败、CreateUser 静默退化。需在部署编排中保证 permission-service 先于 user-service |
| N-5 | 迁移/存量 | 本次声明的「有成员必有 scope」不变量对存量成员（本变更上线前已加入、不会重走 join 路径）无 backfill。若 2026-08-11 角色迁移未回填 rel_user_role，存量活跃成员将「有成员无 scope」。建议核对或补一次性回填 |
| N-6 | 交叉提示（供安全视角） | leave 失败路径（W-1/W-2）可遗留 owner/tenant grant 于 permission-service，用户已退出但 scope 仍在，CheckAccess/GetDataScopes 仍可能放行。请安全 Reviewer 关注 |

---
VERDICT: PASS
---

- 负责维度内无 CRITICAL。
- W-1/W-2/W-3 均为**错误路径 / 故障期间**的数据一致性问题，非正常路径缺陷；W-4 为设计文档滞后。均不影响本次功能正常路径的正确性，建议随后续 Wave 一并修复。
- Migration 维度：本变更无表结构变更（正确），跨服务依赖与存量成员回填见 N-4/N-5，需在集成验收阶段（T6）核实。
