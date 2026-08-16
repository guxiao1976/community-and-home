# CHANGELOG — user-service

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
