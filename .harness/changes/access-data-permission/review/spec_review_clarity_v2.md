# Plan Review — access-data-permission（clarity 清晰可执行视角）

**审查维度**: 粒度 / 歧义 / 一致性（SHALL/MUST 唯一解释、proposal 影响范围 ↔ specs 职责、需再澄清/拆分的行为契约、3 个架构边界项标注）

**审查日期**: 2026-08-12（第 2 轮复审）
**审查者**: Plan Reviewer（clarity lens）
**审查输入**: request.md / proposal.md / specs/{scope-model, capability-layering, registered-user, join-auto-authorization, publish-validation}/spec.md（已修订）/ review/spec_review_clarity_v1.md

## 摘要
- 🔴 MUST FIX: **0** / 🟡 SHOULD FIX: 4 / 🔵 INFO: 3

上一轮 1 个 MUST FIX（capability-layering REQ-2.2 多授予角色聚合语义）+ 1 个补充性 MUST（scope-model 空 scope 物理表示）均已消除；3 个架构边界项已全部在 spec 正文内标注 `[NEEDS CLARIFICATION] 待阶段3架构定稿`。本轮无阻塞项，clarity 视角判定通过。剩余 4 个 SHOULD FIX 均为措辞一致性 / 设计一致性建议，不改变行为契约可解释性，移交阶段 3 时顺手处理即可。

## 上一轮 MUST FIX 复核（重点）

| # | v1 问题 | v2 现状 | 判定 |
|---|---------|---------|:---:|
| 1 | capability-layering REQ-2.2（L25）：`CheckPermission` 对「the granting rel_user_role」单数表述在多授予角色场景下无唯一解释（owner@A status=0 + owner@B status=2 同时授予 level-2 权限时，是全部/任一/主匹配？） | 已重写为显式聚合规则：「aggregate across all of them by taking the maximum satisfied level — the permission SHALL be granted if at least one granting role satisfies the required `min_verf_level`」。即：多授予角色取**最高满足层级**，**存在任一角色满足所需 level 即放行（∃）**——与本轮指定语义完全一致，且用「任一」替换了单数 granting role，消除了主匹配歧义。 | ✅ 消除 |
| 1b | 同一 REQ 仅 1 个多角色场景，缺反面场景 | 已补 3 个场景：L52-56 未认证+已认证并存→取最高放行；L57-60 全部未认证→level-2 拒绝；L62-65 registered_user（status=2、空 scope）参与命中不阻塞 owner 角色授权。正反两面齐备。 | ✅ 消除 |
| 2 | scope-model REQ-1.1/REQ-1.2：`empty` 在 `rel_user_role` 上的物理表示（scope_type 取值 / scope_id 空值约定）未定，且 spec 正文未显式标注「留待架构设计定稿」 | REQ-1.1（L13）已加注 blockquote：`[NEEDS CLARIFICATION] 待阶段3架构定稿`——`empty` 与 `global` 的具体存储表示（如 scope_type 取值 none vs global、scope_id=0）由架构设计确定，本 spec 仅约束行为且明示 empty 不得等价 global（指向 REQ-1.2）。 | ✅ 消除 |

## 3 个架构边界项标注复核

| 边界项 | 要求位置 | v2 实际标注 | 判定 |
|--------|---------|------------|:---:|
| 空 scope 表示（scope_type none vs global / scope_id=0） | scope-model REQ-1.1 | L13 blockquote `[NEEDS CLARIFICATION] 待阶段3架构定稿`（覆盖 empty 与 global 表示） | ✅ |
| 自有/租住权属 → 自动授权的 API 形状（JoinCommunity 是否携带权属、是否与房屋注册合并） | join-auto-authorization REQ-4.1 | L13 blockquote `[NEEDS CLARIFICATION] 待阶段3架构定稿` | ✅ |
| 错误码 80006 vs community-hub 现状 08xxxx 命名空间 | publish-validation REQ-5.1 | L13 blockquote `[NEEDS CLARIFICATION] 待阶段3架构定稿`（明确「最终错误码取值与归属由架构设计对齐，本 spec 只约束行为——拒绝且返回无数据权限类错误」） | ✅ |

三处均与 proposal「明确的边界说明」指向的 REQ 号一致（REQ-1.1 / REQ-4.1 / REQ-5.1），边界项进入阶段 3 处理路径正确。

## 发现

### 🔴 MUST FIX
无。

### 🟡 SHOULD FIX
| # | 文件:行号/章节 | 问题 | 建议 |
|---|-------------|------|------|
| 1 | capability-layering/spec.md REQ-2.3（L69） vs publish-validation/spec.md REQ-5.6（L103） | v1 SHOULD #3 未修：REQ-2.3 能力矩阵仍写「publishing (**within quota**) is available to ...」，而 REQ-5.6 明确「Quota ... out of scope for this change」、proposal 范围外也列「板块发布配额 → 独立变更」。两处措辞互相矛盾，读者可能误判配额需在本变更实现。 | 将 REQ-2.3 的「within quota」改为「发布能力以 min_verf_level=0 + 数据范围为准，配额限制属独立变更（本次不实现）」，与 REQ-5.6 对齐。 |
| 2 | registered-user/spec.md REQ-3.1（L11）/ REQ-3.2（L25） + capability-layering REQ-2.2（L25） | v1 SHOULD #4 部分修复、且产生新残留：REQ-3.2 已补「status `2`（permanently valid）」，但 `registered_user` 的 rel_user_role status=2 与 capability-layering 聚合规则中「status `2` = 已认证（satisfies level-2）」**数值同义相撞**。当前权限集（browse-only, level-0）下无实际越权，聚合规则亦不会对 registered_user 求值 level-2（cap-layering L62-65 场景已覆盖）；但若未来给 registered_user 追加任一 level-2 权限，聚合规则会因其 status=2 误判为已认证而放行。 | 在 REQ-3.2 补一句「`registered_user` 的 status=2 表示『永久有效』，不构成『已认证』能力；其有效权限恒为 browse-only（level-0）」，或为永久有效基角色引入独立 status 值，杜绝未来给基角色配 level-2 权限时被误放行。 |
| 3 | scope-model/spec.md REQ-1.7（L114） | v1 SHOULD #5 未修：REQ-1.7 仍以「hold a `global` scope **or the `sys_admin` role**」角色名判定式 OR 放行数据权限，与 proposal「以授权节点集合为唯一判据」及记忆 [[is-system-no-permission-shortcut]]（不得引入 is_system/字段短路）存在实现张力——sys_admin 的全局访问是建模为 global scope 授权、还是代码按角色名特判，spec 未定。 | 建议统一为「sys_admin 的全局数据访问以 global scope 授权表达（建模为 rel_user_role scope_type=global）」，在 REQ-1.7 补一句，避免实现层走角色名短路；L121-124 场景同步改为「持有 global scope（建模为 sys_admin 授权）」。 |
| 4 | scope-model/spec.md REQ-1.6（L100） | v1 SHOULD #2 部分修复：读过滤已改为「using the scope returned by `GetDataScopes(user_id, "community")`」，把读过滤钉在 community 型 scope，较 v1 明确；但 division 级 limited scope 如何映射到「可见小区」（覆盖其全部后代小区？）仍无定义。本变更只产生 community 级 scope（join）+ global（auditor），故当前不触发，属理论残留。 | 最小化收口：补一句「本变更 limited 读 scope 恒为 community 级，division 级读 scope 留待后续变更」，或将 `GetDataScopes(user_id,"community")` 的语义明确为「仅返回 community 型 scope，不展开 division」。 |

### 🔵 INFO
| # | 建议 |
|---|------|
| 1 | v1 INFO #3 已顺带消除：publish-validation REQ-5.3（L46）已把校验目标钉死为「operation's content `community_id`」，本变更 targets 恒为小区级，ScopeRef 取小区 id，未扩展出非预期目标类型。 |
| 2 | v1 INFO #1 仍可选：capability-layering REQ-2.1 取值域 0/2、默认 0，可加一句「1 为未来层级保留」以免实现者追问。非阻塞。 |
| 3 | 本轮无 tasks.md（阶段 2 计划评审，任务拆解在阶段 3 产出）。clarity 的任务粒度维度（每个 task 1-4h / 1-5 文件）待 tasks.md 生成后补评；本轮其余维度均已覆盖。 |

---

VERDICT: APPROVED
---

**裁决依据**: 无 MUST FIX。上一轮 REVISION 的两个 MUST FIX（capability-layering REQ-2.2 多授予角色聚合语义、scope-model 空 scope 物理表示）均已在修订 spec 中消除，且 3 个架构边界项全部在 spec 正文内标注 `[NEEDS CLARIFICATION] 待阶段3架构定稿`。剩余 4 个 SHOULD FIX 为措辞/设计一致性建议，不改变行为契约可解释性、不阻塞进入阶段 3，建议在阶段 3 架构定稿时一并处理（重点：#2 registered_user status=2 与已认证语义相撞，属未来误配安全点）。
