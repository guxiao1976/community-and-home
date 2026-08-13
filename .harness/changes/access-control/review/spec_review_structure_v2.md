# Plan Review — access-control（structure 结构合理性视角）

**审查维度**: 服务归属、职责边界、Proto 归属、依赖顺序、职责重叠/冲突

## 摘要

第 1 轮 MUST FIX（CLAR-4 的 user.proto 变更未归入 proto_changes）与 3 项 SHOULD FIX 均已核验修复到位：proto_changes 已补齐写入侧采集字段与 membership 三列；section-quota 职责边界越界括号已删除；current-community 消费侧已明确写入「范围外」；CLAR-2「取代/迁移」措辞已统一为「无存量数据、无需迁移、字段去留阶段3定」。

无新引入的 MUST FIX。剩余两项 SHOULD FIX 均为「proto_changes 确定性」与「proposal↔spec 措辞一致性」的收尾问题，不阻塞结构合理性。

- 🔴 MUST FIX: 0 / 🟡 SHOULD FIX: 2 / 🔵 INFO: 4

## 第 1 轮修复核验

| 第 1 轮发现 | 级别 | 修复状态 | 核验依据 |
|-----------|:---:|:---:|---------|
| MUST FIX #1：CLAR-4 的 user.proto 变更未列入 `.change.yaml proto_changes` | 🔴 | ✅ 已修复 | `.change.yaml` L46-47 新增 ①「JoinCommunityRequest 增 building/unit/room 必填字段（CLAR-4 选项 a）」②「CommunityMembership 增 building/unit/room 三列（CLAR-4 选项 c）」，与 design §3.5（api-proto/api/user/v1/user.proto，CommunityMembership/JoinCommunity）一致，写入侧采集字段与 membership 源数据列均已覆盖。 |
| SHOULD FIX #1：section-quota 职责边界混入成员约束 `sys_config` 键 | 🟡 | ✅ 已修复 | `specs/section-quota/spec.md` L89 仅声明 `sys_section_quota(section_type, max_count)`；成员约束 `sys_config` 键由 `specs/member-constraint/spec.md` L90 独占声明。两配置项边界清晰，无重复。 |
| SHOULD FIX #2：current-community 消费侧（community-hub 读当前小区）归属缺失 | 🟡 | ✅ 已修复 | 采用「暂不集成」方案：`specs/current-community/spec.md` 新增「范围外（本次不实现）」章节（L75-77）；`proposal.md` L72 与 `.change.yaml` L37 的 out_of_scope 均补列「community-hub 读当前小区作发布默认上下文」。proposal 影响范围 community-hub 行仅保留配额校验，无遗留无主任务。 |
| SHOULD FIX #3：CLAR-2「取代 preferences / 迁移存量」与「无存量迁移」措辞冲突 | 🟡 | ✅ 已修复 | `.change.yaml` L58、`specs/current-community/spec.md` L66、`proposal.md` L39/L77（D2）均已统一为「app_state 取代 preferences；开发阶段无存量数据，无需迁移；preferences 字段去留阶段3定」，措辞一致，无冲突。 |

## 发现

### 🔴 MUST FIX

（无）

### 🟡 SHOULD FIX

| # | 文件:行号/章节 | 问题 | 建议 |
|---|-------------|------|------|
| 1 | `.change.yaml` L44（proto_changes 第 5 项） | proto_changes 第 5 项「master-data 新增 sys_section_quota 配置读取 RPC（**或**并入既有 sys_config 读取）」含未定稿的「或」，阶段 4 仅按该清单执行 make ci 时无法确定要生成新 RPC 还是复用既有 sys_config 读取；且 design §7 明确 `sys_section_quota(section_type, max_count)` 为独立配置（非 `sys_config` 键），「并入 sys_config 读取」与 design 存在张力。 | 阶段 3 定稿该单项：若新增独立 RPC 则去掉「或」；若确认复用 sys_config 读取则改写该项，并说明 sys_section_quota 的载体归属。或将该未定项移入 `stage3_open`（与 STAGE3-1/2/3 并列），使 proto_changes 成为确定性清单。 |
| 2 | `proposal.md` L32（影响范围 master-data 行） | 「`sys_section_quota(...)` 板块配额；**成员约束/配额相关 `sys_config` 键**」措辞与已修复的 specs 不一致：成员约束用 `sys_config` 键，但板块配额用 `sys_section_quota`（独立配置表，非 sys_config 键）。第 1 轮 SHOULD FIX #1 只改了 section-quota spec，proposal 影响范围未同步，残留「配额相关 sys_config 键」的错误归属暗示。 | 改为「成员约束相关 `sys_config` 键」，或分别列两种配置机制（`sys_section_quota` 表 + `sys_config` 键），与 section-quota / member-constraint 两 spec 的职责边界保持一致。 |

### 🔵 INFO

| # | 建议 |
|---|------|
| 1 | `.change.yaml` `proto_changes` 第 6 项「community-hub 写接口配额校验错误码契约」仍是行为契约（错误码阶段3对齐），非 Proto 结构变更，建议标注「契约对齐项」与真正的 proto 结构变更区分，避免阶段 4 误当作需改 proto 文件的项（第 1 轮 INFO #1 持续有效）。 |
| 2 | `.change.yaml` 仍无显式 spec 间依赖/执行顺序字段，仅靠各 spec 职责边界 + `predecessor` 隐式表达。建议阶段 3 tasks.md 显式列序：permission-service（platforms 存储）→ auth-service（端判定）→ user-service（成员约束/当前小区/同屋互见）→ community-hub（配额）→ master-data（配置）→ 前端，并标注对 predecessor `access-data-permission` 交付物（`AssertPublishScope`/`GetDataScopes`/scope 三态）的依赖（第 1 轮 INFO #2 持续有效）。 |
| 3 | member-constraint 与 same-house-visibility 共享 membership `building/unit/room` 数据模型（均出自 CLAR-4），两 spec 声明一致、无冲突；阶段 3 需保证该列变更只实现一次、两 spec 复用，避免 user-service 内出现两份房屋字段采集/判定逻辑（第 1 轮 INFO #3 持续有效）。 |
| 4 | `specs/current-community/spec.md` L72「新接口权限码」职责边界条目未指定归属服务：权限码注册归 permission-service（design §8「权限码 → permission-service」），接口本身归 user-service。建议补明归属，避免阶段 3 出现无主任务。 |

---
VERDICT: APPROVED
---
