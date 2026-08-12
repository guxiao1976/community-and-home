# Proposal: 数据权限核心（access-data-permission）

> 设计依据（权威来源）：`docs/specs/access-control-design.md` §1.4 角色分层状态机 / §3 角色与认证 / §5 数据权限统一模型，并修订 `docs/specs/rbac-design.md` §2.5 鉴权规则。提交 commit：a2fcc3b。
> 阶段：L 变更①（OpenSpec → N×Pipeline）。本次为开发阶段，无生产存量数据，**不考虑存量数据迁移**。

## 为什么做

系统面向**全国小区**，业主/租户是最大群体。管理操作不能依赖人工审批（否则管理员忙不过来），因此「你是谁 / 你在哪 / 你能做什么」必须**后端权威**。现状是：数据权限（能发到哪）与成员资格脱节，`rel_user_role` 的 scope 语义未落地（scope 三态、祖先链命中均未实现），发布接口没有数据权限校验（`publisher_id` 甚至由客户端传入，可被伪造）。

本次落地「数据权限统一模型」：以**授权节点集合 + 祖先链命中**为唯一判据，让"加入小区 = 自动授权、退出 = 撤销"、注册用户天然无数据范围、发布严格限定在自己范围内——用模型保证安全，而不是靠人工审批。

## 做什么

1. **scope 三态与祖先链统一规则**（permission-service 权威 + master-data 解析 + community-hub/user 消费）：
   - scope 三态：`global`（全放行）/ `限定`（具体节点）/ `空`（无数据范围），严禁把「空」当 `global`。
   - 统一覆盖判据：目标 t 被授权集合 S 覆盖 ⟺ A(t) ∩ S ≠ ∅（A(t)=t 及其祖先，≤6 节点，覆盖树=行政区划树+小区叶子）。
   - master-data 新增「任意 scope 节点 → 祖先链」解析能力（行政区划整树缓存）。
   - 读操作按 `GetDataScopes` 过滤；`global` 例外放行跨小区查看。
   - 授权来源可插拔：本次仅成员资格；商户广告为未来来源，模型兼容但不实现。
2. **能力分层**（`sys_permission.min_verf_level`，修订 rbac-design §2.5）：`0`=持角色+数据范围即可；`2`=需已认证。未认证业主可发布、不可选举；认证期间（待审）保持未认证能力。
3. **注册用户基角色 `registered_user`**（正式角色）：注册自动分配（`status=2` 永久有效），空数据范围，browse 类权限。
4. **加入=自动授权 / 退出=撤销**：JoinCommunity 成功 → 同步 `AssignRole`(owner/tenant + community scope + status=0)；LeaveCommunity 成功 → 同步 `RevokeRole`。幂等，缓存即时失效。
5. **发布校验 `AssertPublishScope`**：community-hub 所有写接口（notices/lostfound/contacts）落库前校验目标小区 ∈ 数据范围；`publisher_id` 一律取自 JWT，忽略客户端传值。

## 影响范围

| 服务 | 变更类型 | 说明 |
|------|:---:|------|
| permission-service | 数据模型 + 新 RPC + 鉴权逻辑 | `sys_permission.min_verf_level`；`sys_role` 增 `registered_user`；`rel_user_role` scope 支持三态；`CheckPermission` 能力分层（修订 rbac §2.5）；`GetDataScopes` 三态语义；新增 `AssertPublishScope` RPC；scope/权限缓存失效（加入/退出/认证状态变更） |
| user-service | 编排（非权威） | CreateUser 成功后自动分配 `registered_user`；JoinCommunity 成功同步 `AssignRole`（owner/tenant，自有/租住决定）；LeaveCommunity 成功同步 `RevokeRole`；不做权限判定 |
| community-hub-service | 消费方 + 新校验 | 所有写接口挂 `AssertPublishScope` 后再落库；`publisher_id` 从 JWT 取；读列表按 `GetDataScopes` 过滤 |
| master-data-service | 新 RPC | 提供「scope 节点 → 祖先链」解析（行政区划整树缓存；小区经 `community_div_id` 入树）。消费方向钉死为 `community-hub → permission-service → master-data`：community-hub 不直接调 master-data 解析 scope |
| api-proto | Proto 变更 | ① `sys_permission.min_verf_level`（列表/详情响应透出）；② `GetDataScopes` 响应支持三态可区分（空 / global / 限定）；③ 新增 `ScopeRef` message（scope_type + scope_id）；④ 新增 `AssertPublishScope(userId, repeated ScopeRef targets)` RPC（req/resp）；⑤ master-data 新增祖先链解析 RPC（如 `ResolveScopeAncestors` req/resp：入参 scope 节点 id，出参祖先链 id 列表）。由全局 Owner 执行，遵守 Proto 管理规范 |
| common | 不变更 | ScopeRef/AssertPublishScope/ResolveScopeAncestors 客户端桩均在 api-proto（permission/v1、masterdata/v1）经 gen/go 生成，消费方各用自身 gen 包，本变更不触碰 common/（架构设计已定稿） |

## 风险评估

- **「空」被误当 `global`**：可能性 低 / 影响 灾难（注册用户放行全国数据）/ 缓解：spec REQ-1.2 显式禁止 + 验收场景覆盖 + 空 scope 返回空集合而非全放行。
- **加入/授权不一致**（membership 有、scope 无，或反之）：可能性 中 / 影响 数据越权或功能失灵 / 缓解：同步调用 + 失败即回滚/重试 + 幂等（REQ-4.1/4.2）。
- **`publisher_id` 伪造**：可能性 中（现状存在）/ 影响 身份冒用发布 / 缓解：JWT 为唯一来源，忽略请求体字段（REQ-5.4）。
- **scope 缓存过期不生效**：可能性 中 / 影响 退出后仍能发布 / 缓解：加入/退出/认证状态变更时 DEL 缓存（REQ-4.4），验收矩阵 §11.4「退出 B 后立刻在 B 发布」。
- **错误码命名空间不一致**：设计文档用 `80006`，community-hub 现状错误码为 `08xxxx`（080001~080005）。影响 低 / 缓解：spec 以行为契约为准（拒绝且给出"无数据权限"错误），最终错误码在架构设计阶段对齐。
- **common/ 变更影响面**：已定稿为**不触碰 common/**（ScopeRef 与两个 RPC 客户端桩经 api-proto gen/go 生成，无需 common 包装）；`.change.yaml common_change_required=false` 与之一致。

## 验收标准

- 注册用户（无小区）发布 → ❌ 拒绝（无数据范围）；未认证业主发布（配额内）→ ✅；未认证业主选举 → ❌（`min_verf_level=2`）；认证业主选举 → ✅；待审期间发布 → ✅。
- 加入小区 → `rel_user_role` 自动出现 owner/tenant + community scope；退出 → 该角色/scope 撤销；重复加入幂等。
- owner@A 在 B（∉ scope）发布 → ❌；抓包改 `publisher_id` → ❌（取自 JWT）；审核员（global）跨小区 → ✅。
- 退出 B 后立刻在 B 发布 → ❌（scope 缓存 DEL 生效）。
- 读列表按数据范围过滤；注册用户读不到小区内部内容。

## 转换追溯（Step 0：设计依据 → Spec 覆盖）

| access-control-design 决策 | proposal 章节 | spec Requirement | 覆盖 |
|----------------------|-------------|-----------------|:---:|
| §1.4 角色分层状态机（注册→未认证→认证→退出撤销） | 做什么 2/3/4 | cap-layering REQ-2.2 / registered REQ-3.3 / join REQ-4.1/4.3 | ✅ |
| §3.1 registered_user（正式角色、空范围、注册自动分配、权限∪） | 做什么 3 | registered-user REQ-3.1~3.4 | ✅ |
| §3.2 加入=自动授权 / 退出=撤销（owner/tenant、status=0、幂等） | 做什么 4 | join-auto-auth REQ-4.1~4.4 | ✅ |
| §3.3 认证状态（待审保持未认证能力） | 做什么 2 | capability-layering REQ-2.2 | ✅ |
| §3.4 能力分层 `min_verf_level`（修订 rbac §2.5） | 做什么 2 | capability-layering REQ-2.1~2.4 | ✅ |
| §5.1 scope 三态 + 禁止「空=global」 | 做什么 1 | scope-model REQ-1.1/1.2 | ✅ |
| §5.2 统一规则：祖先链命中 | 做什么 1 | scope-model REQ-1.3 | ✅ |
| §5.3 覆盖树与祖先链解析（master-data 整树缓存） | 做什么 1 | scope-model REQ-1.4 | ✅ |
| §5.4 授权来源可插拔（商户广告未来） | 做什么 1 | scope-model REQ-1.5 | ✅ |
| §5.5 AssertPublishScope + publisher_id 取 JWT | 做什么 5 | publish-validation REQ-5.1~5.6 | ✅ |
| §5.6 读操作过滤 | 做什么 1 | scope-model REQ-1.6 | ✅ |
| §5.8 global 例外 | 做什么 1 | scope-model REQ-1.7 | ✅ |
| §5.9 缓存与失效 | 做什么 1/4 | scope-model REQ-1.8 / join REQ-4.4 | ✅ |
| §11.1~§11.5 验收矩阵 | 验收标准 | 各 spec 正向/异常场景 | ✅ |
| §12 决策3（存量数据不考虑，开发阶段） | 全文 | —（无迁移要求） | ✅ |
| §5.7 同屋互见 | **范围外（变更②）** | — | ⚠️ 刻意舍弃：归属变更② |

## 范围外 / 后续变更

- **同屋互见**（§5.7，户级数据可见性）→ 变更②。
- **端限制（登录准入）**（§4，`sys_role.platforms`）→ 独立变更（auth-service）。
- **用户应用状态/当前小区**（§6，`user_app_state`）→ 独立变更（user-service）。
- **板块发布配额**（§7，`sys_section_quota`）→ 独立变更（community-hub + master-data）。
- **成员约束反滥用**（§3.5：加入次数上限 / 每户人数 ≤6）→ 独立变更（user-service 执行），本变更不实现其强制检查，但自动授权/撤销与其天然一致。
- **认证审核 AI 自动化**（§3.3 / §10）→ 未来（ai-model + moderation）。
- **商户广告**（§5.4 未来来源）→ 本次仅模型兼容（REQ-1.5），不实现订单/范围/期限。

## 相关经验记忆

- [[is-system-no-permission-shortcut]] — 能力分层不得引入 `is_system`/字段短路；`CheckPermission` 必须走 `rel_role_permission → sys_permission` 全路径。
- [[permission-seed-api-path-must-match-routes]] — 新增 `registered_user`/发布相关权限的种子 path 必须与实际 REST 路由一致。

## 明确的边界说明（供架构设计对齐）

> 以下 3 项已分别在对应 spec 正文内标注 `[NEEDS CLARIFICATION] 待阶段3架构定稿`，spec 只约束行为契约。

- **空 scope 表示**：`empty` 在 `rel_user_role` 上的存储表示（scope_type 取值 `none` vs `global`、scope_id=0）——标注于 `specs/scope-model` REQ-1.1。
- **自有/租住权属 API 形状**：owner/tenant 由「自有/租住」决定（§3.2），加入触发自动授权的确切 API 形状（JoinCommunity 是否携带权属、是否与房屋注册合并）——标注于 `specs/join-auto-authorization` REQ-4.1。
- **错误码命名空间**：`80006` vs community-hub 现状 `08xxxx`——标注于 `specs/publish-validation` REQ-5.1。
