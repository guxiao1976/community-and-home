---
triggers: ["退出小区", "LeaveCommunity", "revoke", "撤销授权", "grant", "残留", "服务角色", "grid_worker", "property_admin", "community_admin", "committee", "RevokeRole", "scope", "作用域", "membership", "免membership", "权限漂移"]
service: user-service
type: pitfall
severity: must-follow
status: active
created: 2026-08-17
updated: 2026-08-17
---

# 退出小区必须撤销该小区全部社区作用域 grant，防服务角色 grant 残留

## 为什么会有这条经验

`revokeCommunityRoles` 原只撤销 owner + tenant。服务角色（grid_worker/property_admin/community_admin 等）支持**免带房号 membership 自助申请**后，出现了权限残留链：

用户 **加入小区（无房号 membership）→ applyRole 认证为服务角色（scope=community, scope_id=该小区）→ 退出小区** 时，若只撤销 owner/tenant，服务角色的 grant 仍残留在 permission-service 的 rel_user_role 表（`expires_at=NULL` 永不自动过期）。而 permission-service 的 `resolvePublishScope` 会按该 grant 的 division 子树展开发布范围——「已加入小区所在 division」退化为「**曾经加入过**」，「退出即失去数据范围」不变量被破坏，且发布权限跨小区越权面被放大。

## 怎么做

1. `LeaveCommunity` 的 `revokeCommunityRoles` 必须遍历**该小区下全部社区作用域角色**：owner / tenant / grid_worker / property_admin / community_admin / committee，逐一 `RevokeRole(scope_type=community, scope_id=communityId)`。
2. `RevokeRole` 幂等（只撤销存在的），无需区分「该角色是否真实持有」。
3. role_id 经 `roleIDByCode`（roleMapper 缓存）解析；任一角色解析失败或撤销失败 → 返回错误，调用方补偿恢复 `bind_status=active`（不留「有成员无 scope」的半提交状态）。
4. 新增服务角色时，检查是否应纳入 `revokeCommunityRoles` 的角色清单（社区作用域 grant 都应在退出时撤销）。

## 怎么验证

- 加入小区 → applyRole 认证服务角色（scope=community:该小区）→ 退出小区 → 断言该用户对该小区**所有**社区作用域 grant 已撤销（permission-service `FindActiveRolesByUserId` / `GetUserRoles` 无残留），而非仅 owner/tenant 被撤销。
- `RevokeRole` 调用次数应与角色清单数量一致（既有测试断言 leave 时 6 次调用）。
- 撤销失败路径：mock RevokeRole 返回错误 → membership 应被补偿恢复为 active 且 LeaveCommunity 返回失败。

## 关联经验

[[auto-grant-unverified-grant-confers-scope-level0]]
