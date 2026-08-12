# Plan Review — access-data-permission（coverage 覆盖完整性视角）

**审查维度**: 需求覆盖 / 场景完整性 / 边界识别 / `[NEEDS CLARIFICATION]` 遗漏

**审查输入**: `request.md` + `proposal.md` + `specs/{scope-model,capability-layering,registered-user,join-auto-authorization,publish-validation}/spec.md`（tasks.md 尚未生成，属阶段 2 评审中，task 级覆盖不在本次审查范围）

**审查日期**: 2026-08-12

## 摘要
- 🔴 MUST FIX: 2 / 🟡 SHOULD FIX: 4 / 🔵 INFO: 6

### 覆盖矩阵（proposal 承诺 → spec Requirement）

| proposal 承诺 | spec Requirement | 覆盖 |
|---|---|:---:|
| scope 三态（global/限定/空） | scope-model REQ-1.1 | ✅ |
| 严禁「空」当 global | scope-model REQ-1.2 | ✅ |
| 祖先链统一判据 A(t)∩S≠∅（≤6） | scope-model REQ-1.3 | ✅ |
| master-data 祖先链解析（整树缓存） | scope-model REQ-1.4 | ✅ |
| 授权来源可插拔（商户未来兼容） | scope-model REQ-1.5 | ✅ |
| 读操作 GetDataScopes 过滤 / 注册用户读空列表 | scope-model REQ-1.6 | ✅ |
| global / sys_admin 例外 | scope-model REQ-1.7 | ✅ |
| scope 缓存失效（join/leave/验证变更） | scope-model REQ-1.8 / join REQ-4.4 / cap REQ-2.4 | ✅（验证变更触发场景仅在 cap，见 SHOULD-3） |
| min_verf_level 0/2 属性 | capability REQ-2.1 | ✅ |
| CheckPermission 能力分层判定 | capability REQ-2.2 | ⚠️（status3 无直接场景 / 多角色叠加未定义，见 SHOULD-1/2） |
| 能力矩阵（注册/未认证/认证） | capability REQ-2.3 | ✅ |
| 认证状态变更即时生效 | capability REQ-2.4 | ✅ |
| registered_user 正式角色 + browse 配置 | registered REQ-3.1 | ✅ |
| 注册自动分配（status=2、空 scope） | registered REQ-3.2 | ✅ |
| 空数据范围约束 + 权限叠加 | registered REQ-3.3 | ✅ |
| 分配幂等 | registered REQ-3.4 | ✅ |
| 加入自动授权（owner/tenant、community scope、status=0） | join REQ-4.1 | ✅ |
| 自动授权幂等（含并发） | join REQ-4.2 | ✅ |
| 退出撤销（仅撤销目标小区） | join REQ-4.3 | ✅ |
| 授权变更即时失效缓存 | join REQ-4.4 | ✅ |
| AssertPublishScope 判定（global/逐目标/80006） | publish REQ-5.1 | ✅ |
| 多目标全部通过才放行 | publish REQ-5.2 | ✅ |
| 所有写接口挂载 | publish REQ-5.3 | ⚠️（枚举不完整，见 MUST-1） |
| publisher_id 取 JWT | publish REQ-5.4 | ⚠️（userId 来源未约束，见 MUST-2） |
| 校验链路顺序 | publish REQ-5.5 | ✅ |

proposal「验收标准」全部 13 条均有对应 Requirement + 场景，无整体性遗漏。请求「工作量分级 / 涉及服务 / 范围外」与 .change.yaml 一致。

## 发现

### 🔴 MUST FIX
| # | 文件:行号/章节 | 问题 | 修复建议 |
|---|-------------|------|---------|
| 1 | `specs/publish-validation/spec.md` REQ-5.3「挂载所有社区写接口」 | proposal 承诺"所有写接口（notices/lostfound/contacts）落库前校验"，但 REQ-5.3 只枚举 `notice create/update, lostfound create, contact upsert`。实际 community-hub 内这三类内容还含写接口：`DeleteNotice`（软删）、`ResolveLostFound`（状态更新）、notice/lostfound 的 `UpdateModerationStatus`（审核回写）——均属"写接口"，漏挂 AssertPublishScope 即留下跨小区/越权写通道。场景仅覆盖 create + notice update，无 delete/状态更新类。 | 将 REQ-5.3 改为「该三类内容的所有写接口（含 create/update/delete/状态变更）在落库/软删前校验」，并补 delete 与状态更新各一个正向/异常场景；或在架构设计阶段产出完整写接口清单逐一核对（该清单当前缺失）。 |
| 2 | `specs/publish-validation/spec.md` REQ-5.1 / REQ-5.5 | 设计 §5.5 承诺"API→RPC 全程透传认证身份"，但 spec 仅对 `publisher_id` 强制 JWT（REQ-5.4），未约束 `AssertPublishScope(userId, ...)` 的 `userId` 参数也必须取自 JWT 认证身份。若实现允许客户端传 `userId`，攻击者可传 sys_admin/global 用户 id 使数据权限判定通过、从而在未授权小区发布——与 proposal 显式防护的"publisher_id 伪造"（REQ-5.4）是同等级漏洞的另一种实现入口。 | 在 REQ-5.1 或新增 Requirement 显式声明 `AssertPublishScope` 的 `userId` MUST 来自 JWT（与 REQ-5.4 的 publisher_id 同源），补 Scenario「篡改 AssertPublishScope userId 为他人 id → 仍按 JWT 身份判定 → 拒绝」。 |

### 🟡 SHOULD FIX
| # | 文件:行号/章节 | 问题 | 建议 |
|---|-------------|------|------|
| 1 | `specs/capability-layering/spec.md` REQ-2.2 | 需求文本声明 status=3（已驳回）"grant neither level"，但 5 个场景只覆盖 status 0/1/2/4；status=3 仅经 REQ-2.4 缓存失效场景（1→3 后拒绝）间接出现，无"已驳回业主直连 CheckPermission"的断言场景。 | 补一个直接场景：「GIVEN rel_user_role status=3 持有 publish 权限 / WHEN CheckPermission → THEN 拒绝（level-0 与 level-2 均不授予）」。 |
| 2 | `specs/capability-layering/spec.md` REQ-2.2 | 多角色叠加同一权限但认证状态不同的边界未定义：用户同时持 status=0 角色（含 publish）与 status=2 角色（含同一权限）时，`min_verf_level=2` 判定取"任一授予角色"还是"某一角色"？spec 用单数 "the granting rel_user_role record" 有歧义。 | 明确为「任一授予该权限的角色满足层级要求即放行（OR 语义）」并补一个叠加场景；否则实现可能误判已认证用户因一个未认证角色而被拒绝。 |
| 3 | `specs/scope-model/spec.md` REQ-1.8 | 需求文本列 3 个失效触发（join/leave/验证状态变更），场景只覆盖 join/leave 两项；验证状态变更触发的 scope 缓存失效在本 spec 无场景，依赖 capability REQ-2.4 间接覆盖，跨 spec 自洽性无保证。 | 补一个「认证通过后 GetDataScopes 立即反映新范围 / 认证变更后 scope 缓存已失效」场景，使 scope 缓存失效三触发各有一个正向断言。 |
| 4 | proposal「明确的边界说明」+ 风险区 | proposal 明确留白 3 处：① JoinCommunity 是否携带权属/与房屋注册合并的 API 形状；② 空 scope 在 rel_user_role 的表示（区别于 global）；③ 错误码 80006 vs community-hub 现状 08xxxx 前缀。3 处均未在对应 spec（join REQ-4.1、scope REQ-1.1、publish REQ-5.1）标注 `[NEEDS CLARIFICATION]`，阶段 3 架构设计时极易被当作已定稿而忽略。全仓 grep 无任何 `[NEEDS CLARIFICATION]` 标记。 | 在 3 处对应 REQ 后显式标注 `[NEEDS CLARIFICATION]` 并指向 proposal 的留白说明；错误码在 publish REQ-5.1 注明"最终错误码架构阶段对齐"。 |

### 🔵 INFO
| # | 建议 |
|---|------|
| 1 | capability REQ-2.4 认证状态变更场景只覆盖 0→2 与 1→3；0→3（直接驳回）、2→4（认证后过期）、2→3 未覆盖，建议补矩阵或在设计中声明懒校验兜底路径。 |
| 2 | scope REQ-1.4 祖先链 ≤6 节点边界未定义：若覆盖树深度 >6，截断行为如何（被截掉的可能恰为授权节点）；master-data 解析器对未知/失效 scope 节点 id 的处理（建议明确"返回空链 → 安全拒绝"）。 |
| 3 | scope REQ-1.4 拓扑失效场景只覆盖"重挂父节点"；division 新增/删除节点、`residential_area` 出树（小区删除）的缓存失效场景未覆盖。 |
| 4 | design §1.4 "退出后回到注册用户层级"无显式场景：退出 A 后应立即回落空 scope + 仅 browse。可由 join REQ-4.3/4.4 推断，建议在 registered REQ-3.3 或 join REQ-4.4 补一条回落断言。 |
| 5 | proposal 引用 must-follow 记忆 `[[permission-seed-api-path-must-match-routes]]`，但 specs 未落 Requirement 或验证场景（新增 registered_user/发布权限的种子 path 须与实际 REST 路由一致）。建议在 registered-user / capability 验收场景加入断言。 |
| 6 | publish REQ-5.1 场景固定 80006 为参考码，与 community-hub 现状 08xxxx（080001~080005）前缀不一致——proposal 风险区已声明"架构阶段对齐"，行为契约（拒绝 + "无数据权限"错误）已明确，可接受；建议在 REQ 中写明"最终错误码架构阶段对齐"以防实现硬编码冲突。 |

---
VERDICT: REVISION
---

**裁决依据**: 存在 2 个 MUST FIX（写接口枚举不完整、AssertPublishScope userId 来源未约束为 JWT），均属数据权限安全核心的行为契约遗漏，修复后可复审。

**其余审查结论**:
- 5 个 spec 覆盖了 proposal「做什么」全部 5 项与验收标准 13 条，覆盖矩阵整体良好。
- 每个 Requirement 均 ≥1 正向 + ≥1 异常/边界 Scenario，无"裸 Requirement"。
- 边界条件覆盖较全面：scope 三态、未认证/待审/已认证/驳回/过期状态（0/1/2/4 直接、3 间接）、幂等（含并发重复加入）、缓存失效即时性、多目标全通过、仅撤销目标小区、global 例外。
- 无整体性需求遗漏；tasks.md 未生成，task 级覆盖待阶段 2 产出后补审。
