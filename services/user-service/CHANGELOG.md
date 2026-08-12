# CHANGELOG — user-service

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
