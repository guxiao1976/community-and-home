# CHANGELOG — user-service

## 2026-08-17 — ApplyRole 检查 AssignRole 的 Base.Code（Review must-follow：业务错误静默当成功）

### 做了什么
- `apply_role_logic.go`：调用 permission-service AssignRole 后只查 `err`，未查 `resp.Base.GetCode()`——permission-service 以 Base.Code 返回业务错误（如 community_admin 每小区上限 60009）时，user-service 误当成功（用户看到"申请成功"但 grant 未建）。
- 修复：显式检查 `assignResp.Base.GetCode()!=0` → 记错误日志 + 透出 Base 错误码/消息。

### 门禁
- `go test ./...` 5 包全绿；harness-checks 18 PASS / 0 FAIL。

---

# CHANGELOG — user-service

## 2026-08-17 — 修复 standards-eng 复审 CRITICAL：REST 边界放行无房号 join + 幽灵记忆引用（字段契约/记忆修复）

### 背景
多视角复审（规范工程 2 CRITICAL）：
1. **REST 契约未随 RPC 语义同步**：join_community_logic.go 已把房号/权属改为「全有或全无」可选（无房号 join 合法且不自动授权），但 `JoinCommunityReq` 的 Building/Unit/Room 无 `,optional`（必填）——go-zero `httpx.Parse`（handler.go:16-20 → mapping/unmarshaler.go:959）对移动端无房号最小载荷 `{community_id}` 返回 `"building" is not set` → API 边界拦截，交付的「无房号加入小区」功能不可用。且 types.go:92 注释 `// SEE: [[api-required-field-marked-optional]] — 楼/单元/房号必填，移除 optional` 与新 RPC 语义直接矛盾（该记忆原则是 API 层与 RPC/Proto 语义一致，此处反向违背）。
2. **幽灵记忆引用（M2）**：leave_community_logic.go:79 `// SEE: [[membership-bound-scope-revoke-on-leave]]` 在 MEMORY.md/.memory-index.json/memory/ 全仓 grep 均无匹配——假引用误导（上一轮 apply_role 评审的 M1 仅覆盖 apply_role_logic.go 未覆盖此文件）。

### 做了什么
1. **types.go 房号/权属字段改为 `,optional`**：`Building/Unit/Room` 标记 `json:"building,optional"` 等（0=未提供，透传 RPC），`Ownership` 注释改为「与房号同为全有或全无——无房号不带 ownership，有房号必须同时带（部分提供 → RPC 10040）」；更新过期 `// SEE: [[api-required-field-marked-optional]]` 注释——该记忆原则（API 与 RPC/Proto 必填语义一致）在新语义下即「标记 optional」，与 08-13 必填化是同一原则的两次方向应用。
2. **创建记忆 `membership-bound-scope-revoke-on-leave`**（user-service/membership-bound-scope-revoke-on-leave.md，含 triggers + 验证方法）：真实规则（退出小区撤销全部社区作用域 grant，防服务角色 grant 残留）有价值且代码已正确实现，创建记忆使引用成立；重建 `.memory-index.json` 使 slug 可被检索。leave_community_logic.go 无需改代码（实现已正确）。

### 测试（TDD RED→GREEN）
| 测试文件 | 用例 | 类型 |
|---|---|---|
| `api/internal/handler/community/handler_test.go`（新增） | `TestJoinCommunityHandler_NoResidenceMinimalPayload`：REST 层最小载荷 `{"community_id":"123"}` → API 边界放行，透传 RPC（building/unit/room=0、ownership=UNSPECIFIED 断言） | 字段契约（接线/回归） |
| `api/internal/handler/community/handler_test.go`（新增） | `TestJoinCommunityHandler_WithResidence_PayloadPassesThrough`：带房号+权属载荷正常透传（字段契约不破） | 字段契约（接线） |

### RED 摘录（先写行为测试，修复未实现时）
```
handler_test.go:67: Not equal: expected: 0 / actual: 99500
  Messages: 无房号最小载荷应在 API 边界放行并透传 RPC（当前被必填字段拦截）
controller.go:269: missing call(s) to *mocks.MockUserServiceClient.JoinCommunity(...)  # parse 失败 → RPC 未被调用
```
（GREEN：Building/Unit/Room 标记 `,optional` 后 `go test ./...` 全绿，221 测试函数 PASS，harness-checks 18 PASS / 0 FAIL / 3 WARN 存量）

### 为什么
RPC 层已支持无房号 join（用户拍板模型），REST 层契约必须同步放行最小载荷，否则前端无房号流程（joinCommunity('c1') 只 POST {community_id}）在 API 边界即被 400 拦截，功能交付却不可用；api_smoke 只查路由不查字段契约故漏过，补 handler 级字段契约回归测试兜底。幽灵记忆引用消除误导：选择「创建记忆」而非「移除引用」，因该规则（退出撤销全部社区作用域 grant）真实且是安全不变量，值得沉淀供后续服务角色扩展复用。

### 影响
- Proto: 无（可选字段语义不变，RPC 0=未提供已是合法输入）
- 调用方: 移动端无房号 joinCommunity 在 API 边界恢复可用（此前被 400 拦截）
- 数据库: 无
- 备注: 记忆应用见下节报告；`.harness/knowledge/memory/.memory-index.json` 已随新建记忆重建

---

## 2026-08-17 — joinCommunity 房号/权属改为可选（用户拍板：加入小区=建 membership，填写房号=独立步骤）

### 做了什么
- 新增纯函数 `joinResidenceProvided` 判定房号/权属的提供情况（全有或全无，部分提供 → 10040）：
  - **全部提供**（ownership∈{OWNED,RENTED} 且 building/unit/room>0）→ 现状：建带房号 membership + 自动授权 owner/tenant（`assignCommunityRole`，status=0）
  - **全不提供**（无房号+无权属）→ 仅建 membership（用户-小区关联，`building/unit/room` 落 0，`bind_status=active`），**不自动授权**（网格员/物业管理员等服务角色后续 `applyRole` 认证）
  - **部分提供**（只给了房号没给权属，或反之）→ 10040 参数错误（要么全有要么全无）
- 移除 ownership/building/unit/room 的强制必填校验；保留其他校验：用户存在 10001 / 上限 3 小区 10006 / 重复 10007 / 年度限流 10012 / 终身 10013 / 每户≤6 10014（**仅带房号时**执行，无房号无地址概念）
- 房号区间/格式校验（楼≤200、单元≤6、房号 3-4 位/楼层≤55/门牌 01-04）改为仅带房号时执行（权威校验，防前端绕过）
- 重新激活路径：无房号 join 将地址**清空为 0** 且不自动授权；有房号保持现状（UpdateAddress + assignCommunityRole + 失败补偿回 left）

### 测试（TDD RED→GREEN）
| 测试文件 | 用例 | 类型 |
|---|---|---|
| `join_community_optional_residence_test.go` | `TestJoinResidenceProvided`：9 组合表驱动（全有-自有/全有-租住/全无/仅权属/仅楼号/仅单元/仅房号/有房号无权属/缺房号） | 逻辑（纯函数） |
| `join_community_optional_residence_test.go` | `TestJoinCommunity_NoResidence_CreatesMembershipNoAutoGrant`：无房号 join → 建 membership（地址 0、active）+ **AssignRole 0 次调用** | 逻辑 |
| `join_community_optional_residence_test.go` | `TestJoinCommunity_NoResidence_NoAutoGrant_OnReactivate`：无房号重新激活 → 成功、地址清 0、不自动授权 | 逻辑 |
| `join_community_optional_residence_test.go` | `TestJoinCommunity_NoResidence_RespectsMaxCommunities`：无房号 join 仍受 10006 上限约束 | 逻辑 |
| `join_community_optional_residence_test.go` | `TestJoinCommunity_Partial_RoomOnly_NoOwnership_Returns10040` / `TestJoinCommunity_Partial_OwnershipOnly_Returns10040`：部分提供 → 10040 且不建 membership | 逻辑 |
| `join_community_member_constraints_test.go`（更新） | MissingBuilding/MissingUnit/MissingRoom 改为「部分提供 → 10040」语义（断言不变） | 逻辑（回归） |

### RED 摘录（先写行为测试，功能未实现时）
```
join_community_optional_residence_test.go:45:34: undefined: joinResidenceProvided        # 纯函数编译 RED
join_community_optional_residence_test.go:68: expected: 0 / actual: 10040               # 无房号 join 应成功（当前强制校验拒绝）
join_community_optional_residence_test.go:72: expected: 1 / actual: 0                   # 无房号 join 应建 membership
join_community_optional_residence_test.go:99: expected: 0 / actual: 10040               # 无房号重新激活应成功
join_community_optional_residence_test.go:122: expected: 10006 / actual: 10040          # 无房号 join 应受上限约束
```
（GREEN：`joinResidenceProvided` + JoinCommunity 分支改造后 `go build ./...` + `go test ./...` 全绿，~155 测试函数 PASS，harness-checks 18 PASS / 0 FAIL / 3 WARN 存量）

### 为什么
用户拍板：加入小区 = 建 membership（用户-小区关联），填写房号 = 独立步骤。网格员/物业管理员等服务角色加入小区无需房号/权属，后续通过 applyRole 认证；业主/租户带房号 join 自动授权 owner/tenant，或先无房号加入再 bindResidence + applyRole 补房号。

### 影响
- Proto: 无（可选字段语义不变，仅校验逻辑改）
- 调用方: 移动端 joinCommunity（支持无房号加入；有房号需同时携带 ownership，否则 10040）
- 数据库: 无表结构变更（membership.building/unit/room 存 0 表示无房号）
- 备注: `// SEE: [[auto-grant-unverified-grant-confers-scope-level0]]` 应用于无房号不自动授权；`// SEE: [[api-required-field-marked-optional]]` 保留于房号区间校验

---

## 2026-08-17 — 退出小区同步撤销服务角色 grant（Review must-follow）

### 背景
`revokeCommunityRoles` 原只撤销 owner+tenant。服务角色（grid_worker/community_admin/property_admin）免 membership 自助申请后，用户 加入→认证服务角色→退出小区，grant 仍残留（含 resolvePublishScope 的 division 子树发布范围），「已加入小区所在 division」退化为「曾经加入过」。

### 做了什么
- `revokeCommunityRoles` 扩展到该小区全部社区作用域 grant：owner / tenant / grid_worker / property_admin / community_admin / committee（RevokeRole 幂等）。
- 配套：`mockJoinRoles` 补 5 个角色；leave/join 测试 RevokeRole 次数 2→6。

### 门禁
- `go test ./...` 4 包全绿；harness-checks 18 PASS / 0 FAIL / 3 WARN（存量）。

---

# CHANGELOG — user-service

## 2026-08-17 — 社区管理员回归 membership 绑定 + needMembership fail-closed（security-arch CRITICAL 修复 + 文档纠偏）

### 做了什么
- **community_admin 必须绑定目标小区有效 membership**：`needMembership` 白名单收敛为
  `grid_worker/property_admin/merchant → false`；`community_admin → true`（回归「数据范围绑定
  membership」）。修复 security-arch CRITICAL：任意注册用户可免 membership 自助申请 community_admin
  并对任意小区Id 得到 status=0 grant，grantActive 视其为活跃并驱动 `resolvePublishScope` 的
  division 子树展开，叠加已认证 level-2 发布角色（committee/property_admin 持
  `community:notice:create-api`=421）时可跨小区越权发布。
- **needMembership 改为 fail-closed**（修复 WARNING）：`switch default → true`（未知/未来角色默认
  需 membership）。任何未来在 permission-service 新增的特权角色若未同步白名单，将自动要求
  membership，杜绝「免 membership 自助申请特权角色」提权地雷。
- **安全模型如实披露**（修复 WARNING，与上一反转条目口径纠偏）：`grantActive` 使 status=0 未认证
  grant **立即**生效于 scope 聚合（`resolveUserScope`/`resolvePublishScope`）——「人工审核通过后
  才生效」仅由 `min_verf_level=2` 在**权限层级**上强制，division 子树展开**不**校验认证状态。
  修复后残余风险边界：已认证 level-2 发布角色叠加「已加入小区」的未认证 community_admin grant 时，
  发布范围仍按该小区所在 division 展开（不再任意小区）。该残余在用户拍板的自助申请模型内可接受，
  待 permission-service 支持「认证通过后再 grant scope」后彻底消除（记录在案，见 08-17 回滚条目）。
- **部署依赖声明**（修复 WARNING）：本反转的安全前提（6 个敏感码 `min_verf_level=2` 加固）依赖
  permission-service **migration/004_privileged_role_min_verf_level.sql 实际执行**（对存量库，
  `init_permissions.sql` 仅对从零建库生效）。部署清单必须显式包含该迁移并执行幂等验证
  （`SELECT` 应返回 hardened=6/6）；否则未认证 community_admin 仍可 level-0 删改公告/建活动。

### 测试（TDD RED→GREEN）
| 测试文件 | 用例 | 类型 |
|---|---|---|
| `apply_role_logic_test.go` | `TestApplyRole_CommunityAdmin_NoMembership_Rejected`：community_admin 无 membership → 10005 + **AssignRole 0 次调用**（安全断言：注册 ListRoles 但不注册 AssignRole，误放行则 gomock 报 unexpected call） | 逻辑（回归） |
| `apply_role_logic_test.go` | `TestNeedMembership`：community_admin→true；新增 super_admin/空串→true（fail-closed） | 逻辑 |
| `apply_role_logic_test.go` | `TestApplyRole_ServiceRoles_NoMembership_Allowed`：收敛为 grid_worker/property_admin（移除 community_admin） | 逻辑 |
| `apply_role_logic_test.go`（既有） | WithMembership / MembershipInactive / ResidentRoles_NoMembership 等全部保持绿（回归） | 逻辑 |

### RED 摘录（先写行为测试，修复未实现时）
```
apply_role_logic_test.go:474: expected: true / actual: false   # needMembership(community_admin)
apply_role_logic_test.go:474: expected: true / actual: false   # needMembership(super_admin) / needMembership("")
apply_role_logic.go:105: Unexpected call to *mocks.MockPermissionServiceClient.AssignRole(...
  user_id:1005 role_id:5 scope_type:"community" scope_id:2001 status:0 ... no expected calls
```
（GREEN：`needMembership` 白名单收敛 + default true 后 `go test ./...` 全绿）

### 为什么
security-arch 多视角复审判 CRITICAL/WARNING：上一反转（服务角色全免 membership）中 community_admin
无门槛自助申请 + status=0 grant 立即生效于 scope 聚合 → 跨小区越权发布放大；`needMembership` 对未知
角色 fail-open 是潜在提权地雷；CHANGELOG「审核通过后才生效」与代码不符且未披露 division 展开放大面。
用户拍板保留 grid_worker/property_admin/merchant 免 membership 自助申请模型，本修复将
community_admin 单独回归 membership 绑定（数据范围绑定成员关系，安全默认），并 fail-closed 兜底
未知角色，文档如实披露残余风险与加固边界。

### 影响
- Proto: 无（复用 ApplyRoleRequest/AssignRole）
- 调用方: 移动端 applyRole（community_admin 需先加入目标小区；grid_worker/property_admin/merchant 不变）
- 数据库: user-service 无变更；**部署依赖 permission-service migration/004 实际执行**（6 个敏感码 min_verf_level=2，幂等）
- 备注: `// SEE: [[auto-grant-unverified-grant-confers-scope-level0]]` 应用于 needMembership 与 applyRole 校验注释

---

## 2026-08-17 — 服务角色免 membership 自助申请（模型拍板反转 + 逻辑函数）

### 做了什么
- **恢复「服务角色免带房号 membership」**（有意反转当日早些时候的 security-arch 回滚）：新增 `needMembership(roleCode string) bool` —— `owner/tenant/committee → true`；`grid_worker/community_admin/property_admin/merchant → false`。
- **applyRole 分支反转**：`if in.RoleCode != RoleCodeMerchant { 查 membership }` → 作用域仍按「非 merchant 绑定小区（scope=community, scope_id=communityId）」，但 **membership 校验改为 `if needMembership(in.RoleCode)`**。服务角色跳过 membership 校验，直接 `AssignRole(scope=community, scope_id=communityId, status=0 未认证待审)`——数据权限来自角色 grant，无房号 membership；merchant 不变（scope=global）。
- **安全模型（用户拍板依据）**：服务角色自助申请是**预期模型**——每个申请需提交盖章文件、由数据审核人员人工审核、通过后才生效；敏感权限（写/管理类）由 permission-service 的 `min_verf_level=2` 加固（未认证 status=0 grant 不能行使破坏性操作）。居民角色（owner/tenant/committee）数据范围仍绑定有效小区成员关系（无则 10005），未变更。

### 测试（TDD RED→GREEN）
| 测试文件 | 用例 | 类型 |
|---|---|---|
| `apply_role_logic_test.go` | `TestApplyRole_ServiceRoles_NoMembership_Allowed`：grid_worker/community_admin/property_admin 无 membership → **允许**申请，AssignRole 被调 + scope=community + scope_id=2001 + status=0（table-driven；注释注明用户拍板模型） | 逻辑 |
| `apply_role_logic_test.go` | `TestApplyRole_ServiceRoles_WithMembership`：服务角色 + active membership → 成功（适配：免 membership，注释更新） | 逻辑 |
| `apply_role_logic_test.go` | `TestApplyRole_ServiceRoles_MembershipInactive`：服务角色 + 已退出 membership → **仍允许**（服务角色不查 membership，适配反转） | 逻辑 |
| `apply_role_logic_test.go`（既有） | `TestApplyRole_ResidentRoles_NoMembership`：owner/tenant/committee 无 membership → 仍 10005（保持不变） | 逻辑 |
| `apply_role_logic_test.go` | `TestNeedMembership`：7 角色表驱动（owner/tenant/committee→true；grid_worker/community_admin/property_admin/merchant→false） | 逻辑 |
| `apply_role_logic_test.go`（既有） | TestApplyRole_Owner/Merchant/NoMembership/GridWorker/RoleCodeNotFound 等全部保持绿（回归） | 逻辑 |

### RED 摘录（先写行为测试，功能未实现时）
```
apply_role_logic_test.go:300: Not equal: expected: 0 / actual: 10005
  Messages: 服务角色无 membership 应允许自助申请
controller.go:137: missing call(s) to *mocks.MockPermissionServiceClient.AssignRole(...)
  apply_role_logic_test.go:284
apply_role_logic_test.go:383: Not equal: expected: 0 / actual: 10005   # MembershipInactive
apply_role_logic_test.go:433:29: undefined: needMembership              # TestNeedMembership 编译 RED
```
（GREEN：新增 `needMembership` + 分支反转后 `go build ./...` + `go test ./...` 全绿，~148 测试函数 PASS，harness-checks 18 PASS / 0 FAIL）

### 为什么
用户明确：服务角色自助申请是预期业务模型（网格员/社区管理员/物业管理员可自助申请，无需本小区带房号 membership），安全由「盖章文件 + 人工审核 + 敏感权限 min_verf_level=2」保证（permission-service 并行加固）。上一轮按同样需求实现后被 security-arch 以「无门槛自助申请特权角色 = 提权漏洞」CRITICAL 拦下并回滚，现用户拍板**有意反转该回滚**——配合 min_verf_level=2 加固满足安全评审核心关切（未认证 grant 不能行使破坏性操作）。

### 影响
- Proto: 无（复用 ApplyRoleRequest/AssignRole）
- 调用方: 移动端 applyRole（服务角色无需先加入目标小区即可申请；数据审核人员人工审核通过后生效）
- 数据库: 无
- 备注: `// SEE: [[auto-grant-unverified-grant-confers-scope-level0]]` 应用于 applyRole 校验注释与文件头安全模型

---

## 2026-08-17 — 修复服务角色自助申请权限提升（逻辑函数 + 安全修复）

### 做了什么
- **回滚「服务角色免带房号 membership」**：applyRole 恢复为「所有小区作用域角色（含服务角色 grid_worker/community_admin/property_admin）必须先有目标小区有效成员关系」，否则 10005；删除 `needMembership` helper。
- **安全背景（SEE: [[auto-grant-unverified-grant-confers-scope-level0]]）**：permission-service 将 `status∈{0,1,2}` 的 grant 一律视为活跃（`grantActive`），`status=0` 未认证 grant 会**立即**产生 community 数据范围（`resolveUserScope`/`GetDataScopes`）+ level-0 能力；community_admin 还驱动 division 子树发布范围（`resolvePublishScope`）。上一变更使任意用户仅凭 JWT 即可为自己申请任意小区特权角色，立即获得数据权限/发布范围/CheckPermission API 权限，且 `expires_at=NULL` 永不自动过期——权限提升，security-arch/standards-eng 双视角判 CRITICAL。
- **修复（数据范围绑定 membership，安全默认）**：无目标小区成员关系 → 10005，且**不触达** permission-service AssignRole（无未认证特权 grant）；有 active 成员关系 → 正常 `AssignRole(scope=community, status=0)`。
- **错误日志补上下文**：`find membership error` 增加 `userId/communityId/roleCode`（并发下可定位）。

### min_verf_level 核验结论（记忆「怎么做」第 1 步）
核验 permission-service `init_permissions.sql` 服务角色敏感权限的 `min_verf_level`：
- ✅ `=2`（需已认证）：`user:read:list/detail-api`(111/112)、`moderation:read:list-api`(511)、`moderation:review:approve/reject-api`(521/522)、`community:notice:create-api`(421)
- ⚠️ 仍 `=0`（持角色+数据范围即可）：`role:read:list/detail-api`(211/212, community_admin)、`community:activity:create-api`(432, community_admin)、`community:notice:delete/update-api`(427/428, 服务角色+全移动端) —— 未认证 grant 即可获得这些 level-0 能力，故服务角色数据范围**必须**绑定 membership，绝不可免成员自我授予
- 建议（permission-service 边界外，交由安全复审）：考虑将 427/428、432 上调 `min_verf_level=2`

### 测试（TDD RED→GREEN）
| 测试文件 | 用例 | 类型 |
|---|---|---|
| `apply_role_logic_test.go` | `TestApplyRole_ServiceRoles_NoMembership_Rejected`：grid_worker/community_admin/property_admin 无 membership → 10005 + `AssignRole` 0 次调用（安全断言）| 逻辑 |
| `apply_role_logic_test.go` | `TestApplyRole_ServiceRoles_WithMembership`：服务角色 + active membership → 成功 + AssignRole scope=community/scope_id=2001/status=0（table-driven）| 逻辑 |
| `apply_role_logic_test.go` | `TestApplyRole_ServiceRoles_MembershipInactive`：服务角色 + 已退出 membership → 10005 | 逻辑 |
| `apply_role_logic_test.go`（既有） | TestApplyRole_Owner / Merchant / NoMembership / GridWorker / RoleCodeNotFound 等全部保持绿（回归）| 逻辑 |

### RED 摘录（先写行为测试，功能未实现时）
```
apply_role_logic_test.go:50: Unexpected call to ...MockPermissionServiceClient.ListRoles(...)
  role_mapper.go:50: ... there are no expected calls of the method "ListRoles"
user/apply_role_logic.go:115: ApplyRole success, userId=1005, roleCode=grid_worker, roleId=4, scope=community:2001
```
（GREEN：恢复「所有小区作用域角色先查 membership」后 `go test ./...` 全绿）

### 为什么
见「安全背景」。原「服务角色免 membership 直接授权」使任意注册用户可自我授予特权角色并立即获得小区数据读范围（level-0）+ community_admin 发布范围，需回滚。服务角色与居民解耦的正当诉求（非本小区居民也可服务）需 permission-service 支持「认证通过后再 grant scope」能力后另行实现——超出 user-service 边界，记录在案待全局评估。

### 影响
- Proto: 无（复用 ApplyRoleRequest/AssignRole）
- 调用方: 移动端 applyRole（服务角色申请需先加入目标小区，与 08-11 之前行为一致；当前前端仅 owner 走 applyRole）
- 数据库: 无
- 备注: 与 JoinCommunity 自动授权（owner/tenant status=0）保持一致——数据范围均绑定 membership；`// SEE: [[auto-grant-unverified-grant-confers-scope-level0]]` 应用于 applyRole 校验注释

---

## 2026-08-16 — 修复 profile 端点本人手机号被脱敏（接线类型）

### 做了什么
- **GetProfileLogic 补传 ViewerId**：`api/internal/logic/user/user_logic.go` 的 `GetProfile()` 调 `UserRpc.GetUser` 时补 `ViewerId: userId`（原为 0）。
- **效果**：`GET /api/users/profile`（本人查自身）→ user-service `GetUser` 命中 `viewerId == in.Id` 分支，返回**明文手机号 + 自身房屋号**；移动端【我的】页 `user.phone` 有值，不再显示「未绑定手机号」。
- **不变量保持**：未改动 `get_user_logic.go` 的 masking 语义——他人查看仍按 `viewer_id==0`（脱敏）或同屋判定脱敏；仅本人查看路径解封。

### 测试（TDD）
| 测试文件 | 用例 | 类型 |
|---|---|---|
| `api/internal/logic/user/user_logic_test.go` | GetProfileLogic：RPC 请求断言 `viewer_id==userId`（自定义 gomock.Matcher）/ RPC Base 10001 透出 error / 未登录 error / RPC 调用失败 error | 接线（断言请求参数） |

### RED 摘录
```
controller.go:269: missing call(s) to *mocks.MockUserServiceClient.GetUser(is anything, GetUserRequest{id=1001, viewer_id=1001})
  controller.go:269: ... because: expected call at user_logic_test.go:227 doesn't match the argument at index 1.
  Got: id:1001 (*userv1.GetUserRequest)
```
（GREEN：补 `ViewerId: userId` 后 `go test ./...` 全绿，~143 测试函数 PASS，harness-checks 18 PASS / 0 FAIL）

### 为什么
`GetProfileLogic` 原未传 `ViewerId`（=0）→ RPC 层按「无查看上下文」走 `maskPhone` 脱敏（默认安全），导致本人查自己 profile 也拿不到明文手机号。

### 影响
- Proto: 无（复用 `GetUserRequest.viewer_id`，Owner 已生成）
- 调用方: 移动端【我的】页（手机号展示恢复）
- 数据库: 无
- 备注: 本仓库 golang/mock v1.6.0 未导出 `gomock.MatchedBy`，测试用自定义 Matcher 断言 RPC 请求参数

---

## 2026-08-13 — 访问控制与数据权限改造（user-service 部分，Task 3.1-3.7）

### 做了什么
- **Task 3.1 — `user_app_state` 表**：`migration/005_add_user_app_state.sql`，账号级当前小区（user_id PK + current_community_id），跨设备一致
- **Task 3.2 — `UserAppState` model**：`FindOne(userId)`（无记录返回 `ErrNotFound`）+ `Upsert(userId, communityId)`（`ON DUPLICATE KEY UPDATE`）
- **Task 3.3 — `GetAppState`/`SetCurrentCommunity`**：GetAppState 读 model，无记录 `current_community_id=0`；SetCurrentCommunity 调 `PermissionClient.GetDataScopes(user_id,"community")`，`GLOBAL`→放行、`EMPTY`→`10015`、`LIMITED` 命中 scope_ids 才放行否则 `10015`，放行后 `Upsert`；抽出 `inScope(state, scopeIds, communityID)`；RPC 层注册 + `ServiceContext` 增 `UserAppStateModel`
- **Task 3.4 — 房屋必填 + 每户 ≤6**：JoinCommunity 顶部增必填校验（building/unit/room 缺一→`10040`）；model 增 `CountActiveByAddress`（`bind_status=active AND user_id<>exclude`）替换 `FindByAddress` 唯一性校验，`>= user.max_house_members`（默认 6）→`10014`，移除 `10011` 路径
- **Task 3.5 — 终身限制对齐 + per-community 认证**：终身 `10013` 校验移出 `!isVerifiedOwnerOrTenant` 块（对全部用户生效）；`isVerifiedOwnerOrTenant` 增 `targetCommunityId`，仅校验目标小区 `community_id` 的 owner/tenant 认证状态（STAGE3-1）
- **Task 3.6 — GetUser 同屋互见**：`maskPhone`（`138****1234`，非 11 位兜底原样）+ `isSameHouse`（同小区同楼/单元/房号 active membership，地址非零才判定）；`viewer_id==0`→脱敏、`==target`→明文+自身房屋号、否则同屋判定
- **Task 3.7 — API 层**：`JoinCommunityReq` 楼/单元/房号移除 `,optional`；新增 `GetAppStateReq/Resp`、`SetCurrentCommunityReq/Resp`；注册 `GET /api/users/me/app-state`、`PUT /api/users/me/current-community`（JWT）；逻辑层用 `responsex.ToError` 透出 `10015`

### 测试（TDD，RED→GREEN）
| 测试文件 | 用例 | 类型 |
|---|---|---|
| `current_community_logic_test.go` | GetAppState 无记录→0 / 有记录→id+updated_at；SetCurrentCommunity GLOBAL 放行 / EMPTY→10015 / LIMITED 命中放行 / 未命中→10015 / GetDataScopes 失败透传 | 逻辑 |
| `join_community_member_constraints_test.go` | 缺楼/单元/房号→10040、房屋 5 人+新用户放行、6 人→10014、退出者不计、重新激活排除自身、认证用户终身 12→10013、A 小区认证加入 B 受每年限制 | 逻辑 |
| `same_house_test.go` | maskPhone 脱敏/兜底；isSameHouse 同屋/不同房/不同小区/零地址；GetUser 无 viewer 脱敏 / self 明文 / 同屋明文 / 非同屋脱敏 / 解密失败兜底 | 逻辑 |
| `api/internal/logic/user/app_state_logic_test.go` | GetAppState 转发 / SetCurrentCommunity 转发+透出 10015 | 逻辑 |

### RED 摘录（回溯补录：移除新增实现后 `go test` 编译失败）
```
rpc/internal/logic/user/get_user_logic.go:49:21: undefined: maskPhone
rpc/internal/logic/user/get_user_logic.go:52:14: undefined: ownHouseInfo
rpc/internal/logic/user/get_user_logic.go:56:26: undefined: isSameHouse
rpc/internal/logic/user/current_community_logic_test.go:23:11: undefined: NewGetAppStateLogic
rpc/internal/logic/user/current_community_logic_test.go:50:11: undefined: NewSetCurrentCommunityLogic
api/internal/logic/user/app_state_logic_test.go:44:7: undefined: NewGetAppStateLogic
api/internal/handler/routes.go:91:19: undefined: user.GetAppStateHandler
rpc/internal/svc/servicecontext.go:30:37: undefined: model.UserAppStateModel
```
（GREEN：恢复新增实现后 `go test ./...` 全绿，98 测试函数 PASS）

### 为什么
当前小区切换（账号级跨设备）+ 成员约束补齐（每户≤6、终身限制对齐）由 user-service 权威执行；同屋互见按「同小区同楼/单元/房号」判定，非同屋默认脱敏（安全）。

### 影响
- Proto: 消费 `user/v1` 新增 RPC `GetAppState`/`SetCurrentCommunity` + `GetUserRequest.viewer_id`/`GetUserResponse.same_house`（Owner 已生成）
- 依赖: `permission-service`（`GetDataScopes` 已交付）
- 数据库: 新增 `user_app_state` 表（`migration/005_add_user_app_state.sql`，需在库执行验证）
- 备注: 重新生成 stale 的 `api-proto/gen/go/user/v1/mocks/user_grpc_mock.go`（补 `GetAppState`/`SetCurrentCommunity`，解除 API 层测试编译阻塞；未改动任何 proto 契约）

---

## 2026-08-12 — 数据权限核心编排（阶段③：注册自动授权 / 加入授权 / 退出撤销）

### 做了什么
- **Task 3.1 — CreateUser 自动分配 registered_user**：DB 落库成功后（moderation 回调前）同步 `AssignRole(userId, role_id(registered_user=9), scope_type='', scope_id=0, status=2)`；role_id 经既有 `roleMapper` 解析；**失败仅告警不阻塞注册**；重复注册（手机号已注册）不重复分配
- **Task 3.2 — JoinCommunity ownership + 自动授权**：校验 `ownership ∈ {OWNED, RENTED}`，UNSPECIFIED → 10040；membership 落库后同步 `AssignRole(user_id, roleIDByCode(owner|tenant), 'community', community_id, status=0)`；**授权失败 → 补偿恢复 membership（置 left）并返回失败**（不留「有成员无 scope」）
- **Task 3.3 — LeaveCommunity 撤销授权**：membership 置 left 后双调 `RevokeRole(owner_role_id + tenant_role_id, 'community', community_id)`（幂等）；**失败 → 恢复 bind_status=active 并返回失败**
- **Task 3.4 — 门禁**：`go build ./...` + `go test ./...` + `harness-checks.sh` **16 PASS / 0 FAIL**
- REST API 层（api/internal）`JoinCommunity` 透传 ownership + building/unit/room（供移动端加入流程）
- 修复 gate 阻塞：`submit_certification_logic.go` certMetadata.MembershipId/CommunityId 补 `json:",string"`（Snowflake 硬约束 #3，pre-existing 违规）
- 新增 model 常量：`RoleCodeRegisteredUser`、`ScopeTypeGlobal/Empty/Community`；helper 新增 `assignRoleToUser`/`stringPtr`

### 测试（TDD，RED→GREEN）
| 测试文件 | 用例 | RED 摘录 | GREEN |
|---|---|---|---|
| `create_user_logic_test.go` | 注册成功→registered_user grant / 重复注册幂等 / AssignRole 失败不阻塞 | `controller.go:137: missing call(s) to ...AssignRole` ×3 | PASS |
| `join_community_ownership_test.go` | OWNED→owner / RENTED→tenant / 授权失败补偿 / 缺 ownership→10040 / 重复加入幂等 | `missing call(s) to ...AssignRole` + 10040 断言失败 | PASS |
| `leave_community_revoke_test.go` | 双调撤销 owner+tenant / 其他小区保留 / 撤销失败恢复 / 重复 leave→10005 | `missing call(s) to ...RevokeRole` ×3 | PASS |
| `join_community_logic_test.go`（更新） | 既有 5 用例补 ownership + permission mock | — | PASS |

### 为什么
permission-service 成为角色唯一权威；加入小区=自动授权（owner/tenant + community scope），退出=撤销，保证「有成员必有 scope」不变量（REQ-4.1/4.2/4.3，design.md §5.3/5.4）。

### 影响
- Proto: 无（复用 AssignRole/RevokeRole/ListRoles）
- 调用方: auth-service（CreateUser 幂等语义不变）、移动端（JoinCommunity 需携带 ownership，Task 5.1 跟进）
- 数据库: 无表结构变更（membership 不落权属）
- 备注: 为解除测试编译阻塞，用 mockgen 重新生成了 stale 的 permission/masterdata gRPC mock（gen/go 为未跟踪生成物，含新 RPC AssertPublishScope/ResolveScopeAncestors），未改动任何 proto 契约

---

## 2026-08-11 — RBAC 角色体系合并 + 认证 REST API

### 做了什么
- **废弃 `user_membership_role`**：角色授予迁移到 permission-service 的 `rel_user_role`
- `ApplyRole` 改调 permission-service AssignRole（写入 rel_user_role，status=0）
- `SubmitCertification` 改走 permission-service（提交时 UpdateUserRoleStatus status=1）
- `ReviewCertification` 改走 permission-service（通过 status=2+expires，驳回 status=3）
- `GetUserRoles`/`CheckAccess` 改为代理 permission-service
- 新增 `role_mapper.go`：role_code↔role_id 映射（调 permission-service ListRoles 缓存）
- 新增认证 REST API：
  - `POST /api/users/certifications`（提交认证材料）
  - `GET /api/users/certifications`（我的认证记录）
  - `GET /api/verifications`（管理员列表认证申请）
  - `POST /api/verifications/:id/review`（管理员审核）
- 移动端 `my.vue` hasOwnerRole 改为真实查询

### 为什么
permission-service 成为角色唯一权威，认证流程从 user-service 自管角色改为调用 permission-service。

### 影响
- Proto: 无（复用现有 RPC）
- 调用方: auth-service（JWT roles 经代理获取）、移动端（applyRole/getUserRoles）
- 数据库: 废弃 `user_membership_role` 表，rel_user_role 承载角色
- 关联: 提交待定

## 2026-06-04 — 错误码 6 位 → 5 位统一

### 做了什么
- 所有错误码从 6 位 `10X00Y` 改为 5 位 `10X0Y`（去掉中间多余的 0）
- 更新文件：`rpc/internal/logic/user/` 下 12 个 .go 文件，`docs/design.md` 错误码表

### 映射
| 旧码 | 新码 | 含义 |
|------|------|------|
| 100001 | 10001 | 用户不存在 |
| 100002 | 10002 | 手机号已注册 |
| 100003 | 10003 | 重复提交认证 |
| 100004 | 10004 | 信用分不足 |
| 100005 | 10005 | 小区成员不存在/退出 |
| 100006 | 10006 | 最多加入5个小区 |
| 100007 | 10007 | 认证申请不存在 |
| 100008 | 10008 | 角色已存在 |
| 100009 | 10009 | 角色已过期 |
| 100010 | 10010 | 权限不足 |
| 100400 | 10040 | 参数校验失败 |

### 影响
- Proto: `api-proto/api/user/v1/user.proto` 注释中的错误码需同步更新（告知全局 Claude）
- 调用方: auth-service 需关注错误码变化
- 数据库: 无

---

## 2026-06-04 — 全局公约与设计文档迁移

### 做了什么
- `CLAUDE.md` 新增 `## 全局公约` 章节，引用根 CLAUDE.md
- 设计文档迁移：`docs/specs/user-design.md` → `services/user-service/docs/design.md`
- 添加 `docs/CHANGELOG.md`（本文件）

### 为什么
项目规范化——统一文件布局，子 Claude 启动时能感知全局架构规则。

### 影响
- Proto: 无
- 调用方: 无
- 数据库: 无
