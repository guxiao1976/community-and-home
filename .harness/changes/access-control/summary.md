# 变更摘要 — access-control

**创建时间**: 2026-08-13
**完成时间**: 未完成（6/7 交付，web/pc 待债务清理）
**路径**: OpenSpec（L 级跨服务）

---

## 阶段 0: 路径选择

- **路径**: OpenSpec → N×Workflow
- **理由**: 跨 7 服务 + Proto 变更 + 数据模型/架构决策（端限制、当前小区、配额、成员约束、同屋互见）
- **涉及服务**: permission / auth / user / community-hub / master-data + web/pc + web/mobile（moderation 不涉及；AI 认证属未来）
- **前置**: 需求 1（角色分层）+ 需求 3（数据权限统一模型）已由 `access-data-permission` Wave 1+2 交付，本变更不再重复

## 阶段 1: 需求分析

- **执行方式**: 子 Agent（requirement-analyst）
- **产出**: proposal.md + 5 specs（platform-restriction / current-community / section-quota / member-constraint / same-house-visibility）+ .change.yaml
- **澄清决策**: 4 个 CLAR 全定（见「关键决策」）
- **状态**: ✅ 通过

## 阶段 2: 需求评审

- **执行方式**: 3 子 Agent 并行（coverage/structure/clarity），2 轮
- **投票结果**:
  - round 1: coverage REVISION（1 MUST）/ structure REVISION（1 MUST）/ clarity REVISION（1 MUST）→ 0/3 回退
  - round 2: structure APPROVED / clarity APPROVED / coverage REVISION（个人维度缺 Scenario，Owner 补修）→ **2/3 通过**
- **状态**: ✅ 通过（少数 REVISION 记入本 summary）

## 阶段 3: 架构设计

- **执行方式**: 子 Agent（architecture-designer）
- **产出**: design.md + tasks.md（31 任务：全局/Proto 6 + permission 4 + auth 3 + user 7 + community-hub 4 + master-data 2 + web/pc 2 + web/mobile 3）
- **开放项定稿**: 6 个 STAGE3 项全定（含 **认证粒度 per-community**——用户纠正架构师原「全局」判定）
- **关键发现**: member-constraint 的「同时≤3/每年≤3/终身≤12」现状代码已实现，真正新增仅「每户≤6」；CLAR-4 的 building/unit/room 已存在，无需改 proto
- **状态**: ✅ 通过

## 阶段 4: Proto 变更

- **执行方式**: Owner 内联
- **变更清单**: auth.proto（RefreshTokenRequest.device_type）、permission.proto（Role.platforms）、user.proto（viewer_id/same_house + GetAppState/SetCurrentCommunity）、masterdata.proto（GetSectionQuota）——全部兼容新增
- **状态**: ✅ 通过（api-proto `c86dc84`，make ci lint+breaking+generate 全绿）

## 阶段 5: 编码+测试

- **执行方式**: N×Workflow 并行（harness-pipeline.js）
- **结果**:

| 服务 | 迭代 | QA | 置信度 | 状态 | 提交 |
|------|:---:|:---:|:---:|------|------|
| permission | 3 轮 | 18/18 + TDD | 0.1→回溯补录 | ✅ | `29198af` |
| community-hub | 1 轮 | 18/18 + TDD 齐全 | 0.8 | ✅ | `e9d60bd` |
| auth | 1 轮 | 18/18 + TDD 齐全 | 0.8 | ✅ | `9c25025` |
| master-data | 1 轮 | 18/18 + TDD 齐全 | 0.8 | ✅ | `71cc08f` |
| user-service | 2 轮 | 18/18 + RED 补录 | 0.1→补录 | ✅ | `b5f63f2` |
| web/mobile | 3 轮 | 前端全绿 + TDD 齐全 | 0.8 | ✅ | 最新提交 |
| web/pc | 2 轮 | 122 历史 TS 错误 | 0.1 | ⏭️ 单独立项 | — |

- **Review**: chore 任务均跳过（taskType=chore）
- **状态**: ⚠️ 后端 5/7 完成，前端 2 个待收

## 阶段 6: 集成归档

- **全链路编译**: 后端 5 服务全绿；前端被 web/pc 历史债挡住
- **运行时冒烟**: ⏭️ 未跑
- **状态**: ⏳ 待 web/mobile + web/pc 收尾

## 关键决策

| # | 决策点 | 决策 | 原因 |
|---|--------|------|------|
| CLAR-1 | 刷新端识别 | RefreshTokenRequest 增 device_type（a） | 与登录一致，改动最小 |
| CLAR-2 | 两个「当前小区」 | app_state 取代 preferences（a） | 开发阶段无存量数据，无需迁移 |
| CLAR-3 | 配额板块清单 | 配置为权威，配置了限、未配置不限（c） | 配置驱动，不硬编码 |
| CLAR-4 | 房屋号采集时点 | JoinCommunity 即采集 + membership 加列（a+c） | 每户≤6 与同屋互见都依赖 active membership 携带房屋号 |
| STAGE3-1 | 认证粒度 | **per-community**（目标小区认证） | 用户纠正：A 小区认证不代表 B 小区认证 |
| STAGE3-3 | 错误码 | 5 位 XXYYY：50007/80007/10014/10015 | 规范化为 5 位，消除 6 位误写 |

## 例外 & 未解决问题

| # | 问题 | 影响 | 后续计划 |
|---|------|------|---------|
| 1 | web/pc 122 条历史 TS 错误 + type-check 假通过 | 阻塞 web/pc 交付 | `web-pc-debt-cleanup` 变更 |
| 2 | web/mobile 分诊滥用（逻辑函数误标纯接线） | TDD 缺口 | 补测中（重新派发） |
| 3 | RED 证据无法机械强制（第 4 次复发） | QA 反复 FAIL | `harness-pipeline-fix`（分诊 + 变异测试 stub） |
| 4 | 嵌套仓库检测盲区（cross_service_import/proto_jstype 跳过） | master-data 检测不全 | harness-checks 按工作树 diff 判定嵌套仓库 |
| 5 | RED 证据标准不一致（结构性 vs 具体摘录） | auth 宽松、web/mobile 严格 | 统一标准 |

## 流水线检验收获（本轮真实发现，6+ 项）

1. RED 摘录无法机械强制（靠 Generator 自觉，4 次复发）→ 变异测试才是正解
2. 分诊滥用（Generator 钻「纯接线」标签空子）
3. RED 标准不一致
4. 嵌套仓库检测盲区
5. type-check 假通过（前端）
6. 前端历史 build 债（web/pc 122 条）

已沉淀：`harness-pipeline-fix/design-tdd-evidence.md` + 分诊落地（qa.js/generator.js）+ 变异测试 stub（check #18）+ `web-pc-debt-cleanup` 立项。

## 交付清单

- [x] 后端 5 服务代码提交（permission/community-hub/auth/master-data/user）
- [x] Proto 变更提交（api-proto）
- [x] CHANGELOG 已更新（各服务 + api-proto）
- [x] 设计稿修正（access-control-design.md §4/§7）
- [ ] web/mobile 补测收尾
- [ ] web/pc 历史债务清理后收尾
- [ ] 集成归档（移动 QA/Review 到 impl/ + 更新 INDEX）
- [ ] Memory Suggestions 处理
