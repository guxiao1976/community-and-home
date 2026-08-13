# Plan Review — access-control（清晰可执行视角）

**审查维度**: 粒度、歧义、一致性（SHALL/MUST 唯一解释、Scenario 可测试性、proposal 影响范围 ↔ specs 职责、术语统一、过大 Requirement）
**轮次**: 第 2 轮（核验 round 1 修复 + 检查新引入歧义/不一致）

## 摘要
- 🔴 MUST FIX: 0 / 🟡 SHOULD FIX: 2 / 🔵 INFO: 7

## round 1 修复核验

| round 1 项 | 类型 | 核验结果 |
|-----------|------|---------|
| #1 CLAR-2「迁移存量数据」口径矛盾 | MUST FIX | ✅ 已修复。`specs/current-community/spec.md:66` 改为「开发阶段无存量数据，无需迁移；preferences 字段去留由阶段3定」；`proposal.md:4/39` 均为「不考虑存量数据迁移/无需迁移」；`.change.yaml` `data_migration_required: false` 且 CLAR-2 决策文本一致。三处口径已统一，无「迁移」字样残留。 |
| #2 active membership 术语不统一 | SHOULD FIX | ✅ 已修复。两个 spec 均新增「术语」节：`active membership ⟺ bind_status = 1（Active/有效）`，并明确「成员生命周期 bind_status 与认证状态 verf_status 正交」。经核对代码 `services/user-service/model/vars.go:36-37`（bind_status 0=Left/1=Active）与 `api-proto/api/user/v1/user.proto:268,333`（bind_status / verf_status），术语现已落到真实字段，且比 v1 更精确（v1 误写「认证状态 status 0/2」，实际字段为 `verf_status`）。 |
| #3 下架状态无字段无 Scenario | SHOULD FIX | ✅ 已修复。`specs/section-quota/spec.md:58-59` 增补「已解决/下架/移除释放配额」Scenario（复用「status 非 active」语义），`:70` [STAGE3] 注明「现网 lost_found 仅 active/resolved 两态，无独立下架状态」。 |
| #4 计数谓词与设计公式矛盾 | SHOULD FIX | ⚠️ 已标 [STAGE3]（spec:68、proposal STAGE3-2、.change.yaml stage3_open 一致），spec 明确「以本节唯一计数谓词为行为契约」。核验为「已标 STAGE3」的接受形态；但注文措辞有新的失真，见 SHOULD FIX #2。 |
| #5 认证粒度模糊 | SHOULD FIX | ✅ 已标 [STAGE3]（`member-constraint/spec.md:40`、proposal STAGE3-1、.change.yaml 一致），明确「全局 vs per-community」两个选项待阶段3定。 |
| #6 刷新 Scenario WHEN 模糊 | SHOULD FIX | ✅ 已修复。`platform-restriction/spec.md:50` WHEN 已改为「用户以 `device_type=web` 的 RefreshTokenRequest 刷新」，与 CLAR-1 已定结论对齐。GIVEN 措辞仍有一处轻微残留（见 INFO #3）。 |

## 发现

### 🔴 MUST FIX
无。

### 🟡 SHOULD FIX
| # | 文件:行号/章节 | 问题 | 建议 |
|---|-------------|------|------|
| 1 | `.change.yaml:46-47`（proto_changes）、`proposal.md:78`（D4）、`specs/member-constraint/spec.md:83` | **CLAR-4 决策「增字段/增列」与实际不符**。三处均称「JoinCommunityRequest 增 building/unit/room 必填字段」「CommunityMembership 增 building/unit/room 三列」，但该字段/列**已存在**：proto 于 2026-06-07 commit `1841fd2` 落地（`api-proto/api/user/v1/user.proto:275-277,290-292`），model 已含 `Building/Unit/Room`（`services/user-service/model/user_community_membership.go:22-24`），migration `003_add_address_fields.sql` 已 `ADD COLUMN building/unit/room`。实际剩余工作是 **JoinCommunity 逻辑层「采集并落库 + 必填校验」**，而非 proto/schema 变更。实施者若按「增三列」执行会写出重复 migration（duplicate column）或误以为字段不存在。 | 将 proto_changes 这两条改为「已存在，本次仅需逻辑层采集/落库 + 必填校验（引 [[api-required-field-marked-optional]]）」，或删除并明确剩余工作为逻辑层；同步修正 proposal D4 与 member-constraint:83 的「增三列」表述。 |
| 2 | `specs/section-quota/spec.md:68`（[STAGE3] 注） | **[STAGE3] 注对「张力」来源表述失真**。注文称「唯一计数谓词 `deleted_at IS NULL AND moderation_status IN (0,1) AND status='active'` 与 design §7 原文公式存在张力」，但随后引用的 design §7 公式 `deleted_at IS NULL AND status='active' AND moderation_status IN (0,1)` 与该谓词**完全相同**（仅 AND 顺序不同），二者不存在张力。真实张力是「待审内容（moderation_status=0）的业务 `status` 未定义」——spec 谓词要求 `status='active'`，而「待审内容占配额」Scenario（:43-46）未声明待审内容的 `status` 取值。注文前半句会误导实施者去「修正一个与谓词相同的公式」。 | 删除「与 design §7 原文公式存在张力」的误导表述，直接陈述真实缺口：「须保证待审内容创建时 `status` 即为 `active`，或明确『待审』在计数谓词中的落入方式」；并保留「阶段3 统一内容状态机 + 修正 design §7 计数公式」的任务记录。 |

### 🔵 INFO
| # | 建议 |
|---|------|
| 3 | 术语漂移：「楼栋房号」（`proposal.md:45`、`specs/same-house-visibility/spec.md:54`）与「楼栋房屋号」（same-house spec Purpose/Req/其余 Scenario、design §5.7）未统一。建议全篇统一为「楼栋房屋号」。 |
| 4 | `specs/member-constraint/spec.md:78-79`「计数排除当前用户」Scenario 触发时机「更新自身信息」仍模糊（v1 INFO#11 未处理）：6 人校验在 JoinCommunity 触发，「更新自身信息」是否也触发该校验未说明。建议明确该 Requirement 的触发入口（仅 JoinCommunity，还是含资料更新/重新激活）。 |
| 5 | `specs/platform-restriction/spec.md:49` GIVEN「持有一个 PC 端会话的 RT」仍暗示 RT 绑定会话，与已定机制「device_type 随请求携带」（CLAR-1）略有张力（v1 #6 的 GIVEN 建议未完全采纳）。建议改为「持有由 PC 端签发的 RT」或删去「PC 端会话」限定，仅保留「用户仅持有 owner 角色」。 |
| 6 | 术语漂移 `request.md:10`「占位状态」vs spec/design「占配额状态」（v1 INFO#7 未处理）。建议统一为「占配额状态」。 |
| 7 | `specs/platform-restriction/spec.md:10`「未声明平台 MUST 视为允许所有端（fail-open）」与 design §4.1「默认：PC=…/移动=…」的「默认」语义（种子配置值 vs 运行时回退）未澄清（v1 INFO#8 未处理）。建议明确二者关系，避免「默认允许所有端」与「默认种子平台」被误读为冲突。 |
| 8 | `specs/section-quota/spec.md:76`「second_hand」示例板块，proposal CLAR-3 已注明 community-hub 现状无 second_hand（v1 INFO#9 未处理）。建议标注「（未来板块，现仅 lost_found）」。 |
| 9 | `specs/platform-restriction/spec.md:40-43`「端归类映射」Scenario 同义反复（GIVEN 即规则本身，THEN 仅复述规则）（v1 INFO#10 未处理）。建议并入 REQ-PLAT-2 说明文字。 |

## proposal 影响范围 ↔ specs 职责一致性（抽查结论）

5 个 spec 的职责边界与 proposal「影响范围」的服务归属一一对应（auth=端准入、permission=platforms 存储透出、user=app_state/成员约束/同屋互见、community-hub=配额校验、master-data=配额与 sys_config 配置、前端=引导/切换/表单/互见展示），未发现职责漂移或越界。Requirement 粒度适中（每 spec 1-4 个，Scenario 2-8 条），SHALL/MUST 均有明确主语与判定条件，未见 MAY 式模糊表述；无需再拆的过大 Requirement。唯一残留的不一致为 SHOULD FIX #1（CLAR-4 表征「schema 变更」而实际已落地，属计划表征与现状不符，非 spec 内部行为歧义）。

---
VERDICT: APPROVED
---
