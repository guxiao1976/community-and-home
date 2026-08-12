# Plan Review — access-data-permission（clarity 清晰可执行视角）

**审查维度**: 粒度 / 歧义 / 一致性（SHALL/MUST 唯一解释、proposal 影响范围 ↔ specs 职责、需再澄清/拆分的行为契约、3 个架构边界项标注）

**审查日期**: 2026-08-12
**审查者**: Plan Reviewer（clarity lens）
**审查输入**: request.md / proposal.md / specs/{scope-model, capability-layering, registered-user, join-auto-authorization, publish-validation}/spec.md / docs/specs/access-control-design.md §3.2–§5 / docs/specs/rbac-design.md §2.5 / services/permission-service/scripts/init_permissions.sql（现有 sys_role.status 约定）

## 摘要
- 🔴 MUST FIX: 1 / 🟡 SHOULD FIX: 5 / 🔵 INFO: 3

request 的 7 个要点 → proposal 的 5 个能力 → 5 份 spec 全部追溯命中（含「商户广告模型兼容不实现」REQ-1.5、「master-data 祖先链解析」REQ-1.4）。proposal「影响范围」的服务归属与各 spec 的 Requirement 逐一对应，无职责漂移。整体契约清晰度高，主要问题集中在：① 能力分层 SHALL 在「多授予角色」场景下存在非唯一解释；② 3 个已知架构边界项在 proposal 中标注清晰、但在 spec 正文内未显式标记为「待架构阶段定稿」。

## 发现

### 🔴 MUST FIX
| # | 文件:行号/章节 | 问题 | 修复建议 |
|---|-------------|------|---------|
| 1 | capability-layering/spec.md REQ-2.2（L25） | `CheckPermission` SHALL「evaluate ... against the certification status of **the granting** rel_user_role record」——单数「the granting」在**多授予角色**场景下无唯一解释。registered-user REQ-3.3 定义有效权限 = registered_user ∪ 社区角色（并集），且 join REQ-4.3 明确允许同一用户同时持有 A、B 两小区的 owner。则当用户 owner@A（status=0）+ owner@B（status=2）同时授予 `committee:election:vote` 时，「granting role」是哪一个？全部须为 status=2、任一为 status=2、还是主匹配角色？三者结论互斥，且安全后果不同（过度授权 = 数据越权 / 欠授权 = 功能失灵）。 | 明确聚合语义：建议「同一权限被多个授予角色命中时，**任一**授予角色满足 min_verf_level 即放行（∃），否则拒绝」；或若意图为逐社区隔离，需显式说明该权限按哪个 rel_user_role 判定并给出与「跨社区认证状态」的对应规则。二者选一并补一条多角色场景。 |
| 2 | scope-model/spec.md REQ-1.1 / REQ-1.2（L11/L30） | （补充性 MUST）「empty 空数据范围」的物理表示完全未定（scope_type='empty'？scope_type='community' + NULL scope_id？），spec 正文也未标注「表示留待架构设计定稿」。行为契约（empty→拒绝、≠global）虽唯一，但 `rel_user_role` 落库字段语义不唯一——这是「空=global 灾难」风险的核心承载体，仅靠 proposal 的边界说明兜底不足。 | 在 REQ-1.1/REQ-1.2 显式加一行：「empty scope 的 rel_user_role 物理表示（scope_type 取值 / scope_id 空值约定）待架构设计阶段定稿，本 spec 仅约束行为」。与 registered-user REQ-3.2、join REQ-4.1 关联标注。 |

> 说明：为保持 1 个 MUST FIX 语义聚焦，将「空 scope 表示」与「多授予角色」合并判为 REVISION 级；二者任一并修都改变行为契约可解释性，建议同轮处理。

### 🟡 SHOULD FIX
| # | 文件:行号/章节 | 问题 | 建议 |
|---|-------------|------|------|
| 1 | publish-validation/spec.md REQ-1.1（L11） | 错误码仅写「reference code `80006`」——「reference」是弱标注，未明确「最终错误码在架构设计阶段与 community-hub 现状命名空间 08xxxx（080001~080005）对齐」。单读 spec 的实现者可能直接落 80006，与存量错误码体系冲突。 | 改为显式：「本 spec 仅以 80006 作为行为契约的引用码，最终错误码在架构设计阶段对齐（现状 community-hub 为 08xxxx 命名空间，见 proposal 风险评估）」。 |
| 2 | scope-model/spec.md REQ-1.6（L96） | 读过滤 SHALL「filter results to the caller's visible communities using the scope」——但 REQ-1.4 允许任意「scope 节点」（含 md_administrative_division）作为 scope，division 级 limited scope 如何映射到「可见小区」（覆盖其全部后代小区？）未定义。本变更实际只产生 community 级 scope（join）+ global（auditor），故当前不触发，但按 SHALL 措辞读法不唯一。 | 两种选一：a) 显式声明「本变更 limited 读 scope 恒为 community 级，division 级读 scope 留待后续变更」；b) 补 division 级 scope 的读覆盖规则（目标小区 ∈ scope 节点后代 ⇔ 祖先链 ∩ scope ≠ ∅）。建议 a，最小化歧义面。 |
| 3 | capability-layering/spec.md REQ-2.3（L54） vs publish-validation/spec.md REQ-5.1（L72） | REQ-2.3 能力矩阵写「publishing (**within quota**) is available to ...」，而 REQ-5.1 明确「Quota ... out of scope」、proposal 范围外亦列「板块发布配额 → 独立变更」。两处措辞互相矛盾，读者可能误判配额需在本变更实现。 | 统一措辞：REQ-2.3 的「within quota」改为「发布能力以 min_verf_level=0 + 数据范围为准，配额限制属独立变更（本次不实现）」，与 REQ-5.1 对齐。 |
| 4 | registered-user/spec.md REQ-3.1（L11）/ REQ-3.2（L24） | 同一 spec 内出现两个语义不同的 status：REQ-3.1 `sys_role.status=1`（角色目录状态，现有种子全部为 1=启用）与 REQ-3.2 rel_user_role `status=2`（个体角色生命周期，0-4）。数值域重叠（rel 的 1=待审），读者易把 registered_user 目录状态 1 误读为「待审」。 | 在 REQ-3.1/3.2 各自括号内标注字段全限定名与语义（如 `sys_role.status`（目录，1=启用）vs `rel_user_role.status`（生命周期 0=未认证…4=已过期）），消除同值歧义。 |
| 5 | scope-model/spec.md REQ-1.7（L110） | 「hold a global scope **or the sys_admin role**」放行数据权限——sys_admin 的全局访问被表述为「角色名判定」式 OR 条件，与 proposal「以授权节点集合为唯一判据」以及记忆 [[is-system-no-permission-shortcut]]（不得用 is_system/字段短路）存在实现张力：是给 sys_admin 的 rel_user_role 赋 global scope，还是代码里对角色名特判？spec 未定。 | 建议统一为「sys_admin 的全局数据访问同样以 global scope 授权表达（建模为 rel_user_role scope_type=global）」，在 REQ-1.7 补一句说明，避免实现层走角色名短路。 |

### 🔵 INFO
| # | 建议 |
|---|------|
| 1 | capability-layering REQ-2.1 取值域为 `0`/`2`、默认 0，与设计 §3.4 一致、无歧义。可加一句「1 为未来层级保留」以免实现者追问为何跳过 1。 |
| 2 | join-auto-authorization REQ-4.1/4.3 回滚语义「rolled back or otherwise kept consistent」为宽松析取，但结果契约清晰（不残留 member-without-scope / scope-without-membership），实现细节留给架构合理，可接受。建议在 task 拆解时把「失败补偿策略（回滚 vs 重试）」显式列为实现项。 |
| 3 | publish-validation REQ-1.1 `AssertPublishScope(userId, targets []scopeRef)` 未定义 scopeRef 可接受取值域（community_id / division_id / building？）。本变更目标恒为小区（REQ-3.1 用 content.community_id），建议在 spec 补一句「本变更 targets 仅小区级，scopeRef 具体定义随 Proto 由 Owner 定稿」，防止实现扩展出非预期目标类型。 |

---

VERDICT: REVISION
---

**裁决依据**: 存在 ≥1 MUST FIX（capability-layering REQ-2.2 多授予角色下 `min_verf_level` 判定聚合语义非唯一解释，security-relevant）。该 MUST FIX 与「空 scope 表示」边界项建议同轮修复后复审；复审时仅复核该 2 处变更，不重审已通过部分。
