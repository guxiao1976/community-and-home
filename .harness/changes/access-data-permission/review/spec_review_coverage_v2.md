# Plan Review — access-data-permission（coverage 覆盖完整性视角 · 第 2 轮复审）

**审查维度**: 需求覆盖 / 场景完整性 / 边界识别 / `[NEEDS CLARIFICATION]` 遗漏

**审查输入**: `request.md` + `proposal.md` + `specs/{scope-model,capability-layering,registered-user,join-auto-authorization,publish-validation}/spec.md` + 上一轮 `spec_review_coverage_v1.md`（tasks.md 尚未生成，task 级覆盖待阶段 2 产出后补审）

**审查日期**: 2026-08-12

## 摘要
- 🔴 MUST FIX: 0 / 🟡 SHOULD FIX: 3 / 🔵 INFO: 6

## 上一轮 2 个 MUST FIX 复核（本轮重点）

| # | v1 MUST FIX | 复核结果 | 证据 |
|---|---|---|---|
| 1 | REQ-5.3 写接口枚举不完整（仅 create+notice update，漏 delete/状态更新） | ✅ 已消除 | REQ-5.3 现逐条枚举 8 类状态变更写操作：`CreateNotice` / `UpdateNotice` / `DeleteNotice` / `UpdateNoticeModerationStatus` / `CreateLostFound` / `ResolveLostFound` / `UpdateLostFoundModerationStatus` / `UpsertContacts`，与 `api-proto/api/community/v1/community.proto` 全量核对一致（community-hub 三服务仅此 8 个状态变更 RPC，无遗漏）。并补 5 个场景覆盖 create/update/delete/resolve/审核状态变更 各至少一个正向或异常断言（v1 点名的 delete 与状态更新各 1 异常场景均已补齐）。 |
| 2 | `AssertPublishScope` 的 `userId` 来源未约束为 JWT | ✅ 已消除 | 新增 REQ-5.5「AssertPublishScope 校验身份取自 JWT」：`userId` SHALL 仅派生自认证 JWT（与 `publisher_id` 同源）、客户端传值 SHALL 被忽略并替换、验证 MUST NOT 代表调用方选定的身份执行、适用于 REQ-5.3 覆盖的全部写操作。补 2 场景：伪造 global-scope 管理员 `userId` 借用数据权限被拒、校验身份与落库 `publisher_id` 一致。 |

## 上一轮 SHOULD/INFO 复核

| # | v1 条目 | 复核结果 |
|---|---|---|
| SHOULD-1 | status=3 无直接 CheckPermission 断言 | 🟡 基本消除：REQ-2.4「驳回后立即收回发布能力」（1→3）已断言 status=3 时 level-0 权限被拒；残余 status=3 对 level-2 未单独断言（见本轮 SHOULD-2） |
| SHOULD-2 | 多角色叠加 OR 语义未定义 | ✅ 已消除：REQ-2.2 文本显式「取最大满足层级（max satisfied level），至少一个授予角色满足即放行（OR）」+ 2 个叠加场景（未认证与已认证并存→放行；全部未认证→level-2 拒绝） |
| SHOULD-3 | scope REQ-1.8 验证状态变更触发无场景 | 🟡 未消除（见本轮 SHOULD-3） |
| SHOULD-4 | 3 处 `[NEEDS CLARIFICATION]` 标记缺失 | ✅ 已消除：scope REQ-1.1（空 scope 存储表示）、join REQ-4.1（JoinCommunity API 形状/权属）、publish REQ-5.1（80006 vs 08xxxx）三处均已标注「[NEEDS CLARIFICATION] 待阶段3架构定稿」，且 grep 确认仅此 3 处、无遗漏 |
| INFO-6 | 80006 vs 08xxxx 前缀 | ✅ 已关闭（REQ-5.1 已标注，行为契约「拒绝 + 无数据权限错误」明确） |

## 覆盖矩阵（proposal 承诺 → spec Requirement）

5 个 spec 的 Requirement 与 proposal「做什么」5 项、验收标准 13 条、request.md 7 项逐条比对，与 v1 一致且无回归：每个 Requirement ≥1 正向 + ≥1 异常/边界场景，无「裸 Requirement」。修订后 REQ 编号仍为 5.1~5.6（新 REQ-5.5 为 userId 取 JWT，原校验链路顺序顺延为 REQ-5.6），与 proposal 转换追溯 `REQ-5.1~5.6` 引用一致，无悬空编号。

## 发现

### 🔴 MUST FIX
无 — 上一轮 2 个 MUST FIX 均已消除，修订未引入新的 MUST 级缺口。

### 🟡 SHOULD FIX
| # | 文件:行号/章节 | 问题 | 建议 |
|---|-------------|------|------|
| 1 | `specs/publish-validation/spec.md` REQ-5.3 / REQ-5.5 | 审核状态回调（`UpdateNoticeModerationStatus` / `UpdateLostFoundModerationStatus`）的调用方身份未定义。proto 注释表明其由 **moderation-service 审核完成后回调（服务间调用）**，且 `UpdateModerationStatusRequest` 仅含 `id`+`moderation_status`、无 `community_id`（目标小区需服务端按 id 反查）。若回调持系统/服务身份（无用户 JWT），REQ-5.5「userId 取自用户 JWT」对这两个接口不成立；若持审核员用户 JWT，则 REQ-5.3 场景「审核状态回调目标小区未覆盖则拒绝」可达。当前 spec 对「谁来认证回调、服务身份如何满足数据权限（是否系统级 global）」无交代，该场景可能不可达/不可测。 | 在 REQ-5.3 或 REQ-5.5 补一条 `[NEEDS CLARIFICATION]`：审核状态回调的认证身份（服务身份 vs 审核员用户身份）与数据权限满足方式（系统级 global 放行 vs 用户 scope 校验）由架构设计定稿，并声明行为契约「服务身份回调一律放行（moderation 系统账号具 global 权限）、用户身份回调按用户 scope 校验」，避免实现阶段两难。 |
| 2 | `specs/capability-layering/spec.md` REQ-2.2 | status=3（已驳回）对 level-2 权限未单独断言：REQ-2.4「驳回后立即收回发布能力」（1→3）已断言 level-0 拒绝，但「status=3 持 level-2 权限经 CheckPermission 亦拒绝」无直接场景，仅由 REQ-2.2 文本「status 3 satisfies neither level」隐含。 | 补一条直接断言「GIVEN status=3 持有 level-2 权限 / WHEN CheckPermission → THEN 拒绝（level-0 与 level-2 均不授予）」，或复用 REQ-2.4 上下文扩展。 |
| 3 | `specs/scope-model/spec.md` REQ-1.8 | 需求文本列 3 个失效触发（join / leave / 验证状态变更），场景仅覆盖 join 与 leave 两项；验证状态变更触发的 scope 缓存失效在本 spec 无直接场景，依赖 capability REQ-2.4 跨 spec 间接覆盖，本 spec 内自洽性无保证。 | 补一个「认证状态变更后 GetDataScopes 立即反映新范围 / 缓存已失效」正向场景，使 3 个失效触发各有一个本 spec 内的断言。 |

### 🔵 INFO
| # | 建议 |
|---|------|
| 1 | REQ-5.3 中 contact upsert（`UpsertContacts`）仅出现在枚举文本，无专属场景（其余 7 类至少被 create/update/delete/resolve/审核 场景之一覆盖）。建议补「联络信息目标小区超出数据范围被拒」场景或声明与 create 同型。 |
| 2 | capability REQ-2.4 状态迁移仅覆盖 0→2 与 1→3；0→3（直接驳回）、2→4（认证后过期）、2→3 未覆盖，建议补矩阵或在设计中声明懒校验兜底路径。 |
| 3 | scope REQ-1.4 祖先链 ≤6 节点边界未定义（覆盖树深度 >6 时截断行为、被截掉的节点恰为授权节点时如何处理）；master-data 对未知/失效 scope 节点 id 的处理未定义（建议明确「返回空链 → 安全拒绝」）。 |
| 4 | scope REQ-1.4 拓扑失效场景仅覆盖「重挂父节点」；division 新增/删除节点、`residential_area` 出树（小区删除）的缓存失效未覆盖。 |
| 5 | design §1.4「退出后回到注册用户层级」无显式场景（可由 join REQ-4.3/4.4 + registered REQ-3.3 推断），建议在 registered REQ-3.3 或 join REQ-4.4 补一条回落断言（退出后仅 browse + 空 scope）。 |
| 6 | proposal 引用 must-follow 记忆 `[[permission-seed-api-path-must-match-routes]]`，但 specs 未落 Requirement/验证场景（新增 `registered_user`/发布权限的种子 path 须与实际 REST 路由一致），建议在 registered-user / capability 验收场景加入断言。 |

## 待补审
- `tasks.md` 尚未生成（阶段 2 评审中），task 级覆盖（每个 Requirement 是否有对应 task）待阶段 2 产出后补审，同 v1。

---
VERDICT: APPROVED
---

**裁决依据**: 上一轮 2 个 MUST FIX 均已消除——① 写接口已全量覆盖 8 类（与 community.proto 逐条核对一致）并补齐 delete/状态更新场景；② `AssertPublishScope` 的 `userId` 已显式约束为仅取 JWT（新 REQ-5.5，含伪造 admin userId 场景）。修订未引入新的 MUST 级缺口，5 个 spec 对 proposal/验收标准/request 均保持完整覆盖。剩余 3 个 SHOULD 为边界/可测性细化（审核回调身份、status=3 level-2 断言、scope 缓存验证触发场景），按 review.md 规则「无 MUST FIX → APPROVED」，不阻塞进入阶段 3。
