# Plan Review — access-control（清晰可执行视角）

**审查维度**: 粒度、歧义、一致性（SHALL/MUST 唯一解释、Scenario 可测试性、proposal 影响范围 ↔ specs 职责、术语统一、过大 Requirement）

## 摘要
- 🔴 MUST FIX: 1 / 🟡 SHOULD FIX: 5 / 🔵 INFO: 5

## 发现

### 🔴 MUST FIX
| # | 文件:行号/章节 | 问题 | 修复建议 |
|---|-------------|------|---------|
| 1 | `specs/current-community/spec.md:58`（CLAR-2 结论）↔ `proposal.md:4`（范围声明）↔ `docs/specs/access-control-design.md` §6/§12-决策3 | **存量数据迁移自相矛盾**。spec 将「`user_app_state` 取代 `preferences.default_community_id`（**迁移存量数据**）」写为已定结论（CLAR-2 选项 a），但 proposal 首行明确「开发阶段，无生产存量数据，**不考虑存量数据迁移**」；设计 §6 又完全未提迁移，§12 决策3 仅针对「存量 scope 数据」而非 preferences。实施者无法唯一确定：是否要写迁移逻辑（把现有 preferences 值拷入 user_app_state）？三处口径互相冲突。 | 统一口径二选一：(a) 若确认开发阶段无存量数据，删除 spec 与 .change.yaml 中的「迁移存量数据」字样，仅保留「app_state 为唯一权威、preferences 废弃」；(b) 若确有开发库 preferences 存量需迁移，则在 proposal 范围声明与设计 §6 明确迁移内容，并同步 .change.yaml 的 `data_migration_required`（现为 false，与「迁移」矛盾）。 |

### 🟡 SHOULD FIX
| # | 文件:行号/章节 | 问题 | 建议 |
|---|-------------|------|------|
| 2 | `specs/member-constraint/spec.md:10,49`、`specs/same-house-visibility/spec.md:10,37` | **「active membership」术语不统一、未落到字段**。member-constraint 用「有效小区成员关系 / 活跃注册人数 / 活跃成员」，same-house 用「活跃成员关系（active membership）」，且仅 same-house Scenario 出现一次「`bind_status` 非活跃」。实际代码字段为 `CommunityMembership.bind_status`（`0=Left / 1=Active`，见 `services/user-service/model/vars.go:36-37`），但 spec 全文未统一命名该谓词。此外 `bind_status`（成员生命周期）与「认证状态 status 0/2」（能力分层）是两个正交轴，spec 混用「活跃/有效/退出」与「认证/未认证」，易使实施者混淆。 | 在两个 spec 的 Purpose 或 Requirement 中统一定义谓词：`active membership ⟺ bind_status = Active(1)`；「退出/失效 ⟺ bind_status = Left(0)」。明确区分「成员生命周期（bind_status）」与「认证状态（能力分层）」两个维度，避免把「活跃成员」与「认证业主」混谈。 |
| 3 | `specs/section-quota/spec.md:41`、`docs/specs/access-control-design.md` §7 | **「下架/移除」无具体状态字段且无对应 Scenario**。REQ-QUOTA-3 列「下架/移除均释放配额」，但未给出其对应的状态值（`status=?`），「驳回/解决/删除释放配额」Scenario 亦未覆盖「下架」。设计 §7 表格同样只有文字「下架/移除 ❌ 不可见即释放」。实施者无法写出可测试的计数条件。 | 明确「下架」对应哪个具体状态值（如 `status='removed'` 或 `is_offline=true`），并为「下架释放配额」补一条 GIVEN/WHEN/THEN Scenario；若现网无下架状态，则删除该口径或标注「未来状态，本次不枚举」。 |
| 4 | `specs/section-quota/spec.md:41,43-56` ↔ `docs/specs/access-control-design.md:246` | **占配额计数谓词与设计公式矛盾**。spec 表述「待审（`moderation_status=0`）与展示中（`status=active`）占用配额」；但设计 §7 给出的计数条件 `deleted_at IS NULL AND status='active' AND moderation_status IN (0,1)` 会把「业务 status 非 active 的待审内容」排除掉——待审内容（moderation_status=0）的业务 `status` 到底是不是 `active`？spec 与设计均未定义。若实现按该公式落地，待审内容将不再占配额，与「防发→删→重发刷审核队列」目标直接冲突。 | 明确待审内容的业务 `status` 取值，并给出唯一计数谓词，例如 `deleted_at IS NULL AND (moderation_status=0 OR (status='active' AND moderation_status=1))`；同步修正设计 §7 公式，使 spec 与设计一致。 |
| 5 | `specs/member-constraint/spec.md:23,30-33`、`docs/specs/access-control-design.md` §3.5 | **「仅非认证用户受限」的「认证」粒度模糊**。认证状态是 per-community（每小区角色 status 0/2，见设计 §3.2），用户可能在 A 小区已认证、B 小区未认证。「认证用户不受每年次数限制」究竟指「任一小区已认证」还是「目标小区已认证」？若按「任一已认证」放行，可通过在 A 认证后对 B 反复退出/重加入刷未认证发布，绕过该约束。 | 明确「认证用户」判定口径（建议：全局维度「是否存在任一 status=2 的认证成员关系」，或按目标小区维度），并补一条边界 Scenario 覆盖「在 A 已认证、在 B 未认证」的用户加入新小区的判定结果。 |
| 6 | `specs/platform-restriction/spec.md:49-51` | **刷新 Scenario 的 WHEN 模糊，与已定 CLAR-1 结论不一致**。GIVEN「持有一个 PC 端会话的 RT」、WHEN「刷新请求**能识别为** PC 端」——「能识别」未说明机制；而底部已定结论（CLAR-1 选项 a）明确 `RefreshTokenRequest` 增 `device_type`。Scenario 文本未体现该决定，且 RT 本身并不天然携带「PC 会话」标记，可测性不足。 | 将 WHEN 改写为具体可测表述：「WHEN 用户以 `device_type=web` 的 RefreshTokenRequest 请求刷新」，与 CLAR-1 已定结论对齐；GIVEN 改为「用户仅持有 owner（platforms=[mobile]），持有由 PC 端签发的 RT」。 |

### 🔵 INFO
| # | 建议 |
|---|------|
| 7 | 术语漂移：`request.md:10` 用「占位状态」，`specs/section-quota` 与设计 §7 用「占配额状态」。二者含义接近但字面不同，建议全篇统一为「占配额状态」。 |
| 8 | `specs/platform-restriction/spec.md:10` 的「未声明平台的角色 MUST 视为允许所有端（默认不拦截）」是 fail-open；与设计 §4.1「默认：PC=…/移动=…」的「默认」语义（指角色种子配置值，还是运行时回退）未澄清。建议明确二者关系，避免「默认允许所有端」与「默认种子平台」被误读为冲突。 |
| 9 | `specs/section-quota/spec.md:62` 的「不同板块独立计数」Scenario 使用 `second_hand` 板块，但 proposal CLAR-3 已注明 community-hub 现状无 `second_hand`。作为示意可接受，建议标注「（未来板块，现仅 lost_found）」，避免实施者误以为需立即支持 second_hand。 |
| 10 | `specs/platform-restriction/spec.md:40-43`「端归类映射」Scenario 为同义反复（GIVEN 即规则本身，THEN 仅复述规则），无新增可测信息。建议并入 REQ-PLAT-2 的说明文字，而非独立 Scenario。 |
| 11 | `specs/member-constraint/spec.md:66-69`「计数排除当前用户」Scenario 的 GIVEN「用户已是某房屋成员并**更新自身信息**」触发时机模糊——6 人校验在 JoinCommunity 触发，「更新自身信息」是否也触发校验未说明。建议明确该 Requirement 的触发入口（仅 JoinCommunity，还是含资料更新）。 |

---

## proposal 影响范围 ↔ specs 职责一致性（抽查结论）

核对 proposal「影响范围」与 5 个 spec「职责边界」：auth-service / permission-service / user-service / community-hub-service / master-data-service / 前端 的服务归属与变更类型一一对应，未发现职责漂移或越界（唯一例外为 MUST FIX #1 的迁移口径矛盾，属范围声明内部冲突而非职责错位）。5 个 spec 的 Requirement 粒度适中（每个 1-4 个、Scenario 2-4 条），未发现需再拆的过大 Requirement；SHALL/MUST 均有明确主语与判定条件，未见 MAY 式模糊表述。

---
VERDICT: REVISION
---
