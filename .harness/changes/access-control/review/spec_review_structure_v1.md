# Plan Review — access-control（structure 结构合理性视角）

**审查维度**: 服务归属、职责边界、Proto 归属、依赖顺序、职责重叠/冲突

## 摘要

整体结构清晰：5 个 spec 均含「职责边界」章节，服务归属与 `docs/specs/access-control-design.md` §8 一致，依赖方向正确（auth/permission/user 基础 → community-hub/master-data → 前端），Proto 变更集中在 `.change.yaml`（全局执行，符合硬性约束 #2），无 spec 越权认领 api-proto。

主要问题是 **CLAR-4 决定的房屋数据模型 proto 变更未归入全局 `proto_changes` 清单**，会直接影响阶段 4（Owner 执行 make ci）的完整性。

- 🔴 MUST FIX: 1 / 🟡 SHOULD FIX: 3 / 🔵 INFO: 3

## 发现

### 🔴 MUST FIX

| # | 文件:行号/章节 | 问题 | 修复建议 |
|---|-------------|------|---------|
| 1 | `.change.yaml` L38-44（proto_changes）+ `specs/member-constraint/spec.md` L71（CLAR-4 已定）+ `specs/same-house-visibility/spec.md` L53（CLAR-4 已定） | CLAR-4 已定为「JoinCommunity 即采集楼/单元/房号，membership 增 building/unit/room 三列」，这是 `api-proto/api/user/v1/user.proto` 的字段/列变更，且被 member-constraint（房屋人数计数）与 same-house-visibility（同屋判定）两个 spec 共同依赖。但 `.change.yaml` 的 `proto_changes` 清单**未列出**「JoinCommunityRequest 增 building/unit/room 必填字段」与「CommunityMembership 增 building/unit/room 三列」两项（仅 item4 覆盖了「用户详情响应」的读取侧字段，未覆盖写入侧采集字段与 membership 源数据列）。阶段 4 仅按该清单执行 make ci，会遗漏这两处 proto 变更，导致两个 spec 的数据模型无承载、无法落地。 | 在 `.change.yaml` `proto_changes` 增补：① `user.proto JoinCommunityRequest 增 building/unit/room 必填字段（CLAR-4 选项 a）`；② `user.proto CommunityMembership 增 building/unit/room 三列（CLAR-4 选项 c）`。并在阶段 3 tasks.md 中对应到 api-proto 变更任务，确保由全局执行而非子服务 spec 认领。 |

### 🟡 SHOULD FIX

| # | 文件:行号/章节 | 问题 | 建议 |
|---|-------------|------|------|
| 1 | `specs/section-quota/spec.md` L75（职责边界·master-data-service） | 职责边界越界：`master-data-service` 条目括号「（以及成员约束相关 `sys_config` 键）」把 member-constraint 的配置职责混入 section-quota，而 member-constraint spec L78 已自行声明该职责，形成重复/混淆。section-quota 只应声明 `sys_section_quota`。 | 删除 section-quota 职责边界中该括号内容，让 member-constraint 独占声明 `sys_config` 键，保持「一个配置项一个 spec 声明」的边界清晰。 |
| 2 | `specs/current-community/spec.md` L60-64（职责边界）+ `proposal.md` L26-33（影响范围 community-hub 行） | 消费侧归属缺失：设计 §6「用途：首页/发布默认上下文」及 §8 依赖图「community-hub ──读当前小区──▶ user-service」，但 current-community spec 职责边界仅列 user-service / permission-service / 前端，未列 community-hub 为消费者；proposal 影响范围 community-hub 行也未提「读当前小区作发布默认上下文」。该集成点无人认领。 | 明确归属：若本次实现「发布默认上下文」，则在 current-community spec 职责边界补充 community-hub 消费侧；若暂不集成，则明确写入范围外/后续变更，避免阶段 3 出现无主任务。 |
| 3 | `.change.yaml` L59（CLAR-2 decision）+ `specs/current-community/spec.md` L58 + `proposal.md` L4（无存量迁移） | CLAR-2 决定「app_state 取代 preferences（迁移存量数据）」，与 change 级「开发阶段无生产存量数据、不考虑存量数据迁移（§12 决策3）」措辞冲突；且若「取代」意味着移除/弃用 `user_base.preferences.default_community_id` 字段，则该 proto 变更也未出现在 `proto_changes` 中。 | 阶段 3 明确 `preferences.default_community_id` 字段去留：保留则说明「并存但 app_state 为权威」；移除则在 `proto_changes` 补齐弃用/删除项，并统一「迁移」与「无存量迁移」的措辞。 |

### 🔵 INFO

| # | 建议 |
|---|------|
| 1 | `.change.yaml` `proto_changes` 第 6 项「community-hub 写接口配额校验错误码契约」本质是行为契约（错误码阶段 3 对齐）而非 Proto 结构变更，建议标注为「契约对齐项」，与真正的 proto 结构变更区分，避免阶段 4 误当作需要改 proto 文件的项。 |
| 2 | `.change.yaml` 无显式 spec 间依赖/执行顺序字段，依赖顺序目前仅靠各 spec 职责边界与 `predecessor` 隐式表达。建议阶段 3 tasks.md 显式列出顺序：permission-service（platforms 存储）→ auth-service（端判定）→ user-service（成员约束/当前小区/同屋互见）→ community-hub（配额）→ master-data（配置）→ 前端，并标注对 predecessor `access-data-permission` 交付物（`AssertPublishScope`/`GetDataScopes`/scope 三态）的依赖。 |
| 3 | member-constraint 与 same-house-visibility 共享 membership `building/unit/room` 数据模型（均出自 CLAR-4），两 spec 声明一致、无冲突；但需在阶段 3 保证该列变更只实现一次、两个 spec 复用，避免 user-service 内出现两份重复的房屋字段采集/判定逻辑。 |
| 4 | platform-restriction spec L64 职责边界将前端统一写作「web/mobile」，但端引导提示的实际展示端是 PC（移动端角色在 PC 登录时被引导），移动端对该角色的提示场景未区分；属 coverage/clarity 范畴，供并行视角参考。 |

---
VERDICT: REVISION
---
