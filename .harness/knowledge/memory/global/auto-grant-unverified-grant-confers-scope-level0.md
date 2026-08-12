---
triggers: ["自动授权", "加入小区", "JoinCommunity", "AssignRole", "status=0", "未认证 grant", "数据范围", "scope", "level-0", "min_verf_level", "认证绕过", "权限种子"]
service: all
severity: must-follow
type: pitfall
status: active
created: 2026-08-12
updated: 2026-08-12
---

# status=0 未认证 grant 立即生效：数据范围 + level-0 能力

## 为什么会有这条经验

permission-service 的 `FindActiveRolesByUserId` / `resolveUserScope` / `grantSatisfiedLevel` 将 `status∈{0,1,2}` 的 grant 一律视为活跃：未认证（status=0）grant 会立即产生 LIMITED 数据范围（scope_type + scope_id，scope_id≠0），并按 grantSatisfiedLevel 计算为 level-0 能力。因此任何「注册/加入即自动授权」的调用方（如 user-service JoinCommunity 自报 ownership 即 grant owner/tenant status=0）都会让用户未认证就获得：1) 对应 scope 的数据读权限；2) 该角色下所有 min_verf_level=0 的能力。认证门禁（status=2 → level-2）只拦截 min_verf_level=2 的权限。

## 怎么做

1. 自动授权 status=0 角色前，必须核验 permission-service 的 sys_permission 种子里所有敏感写/管理权限 `min_verf_level=2`
2. 安全默认：数据范围应绑定 membership 而非角色 grant，或在认证通过后再 grant 角色
3. 涉及自动授权的调用方（user/community-hub 等）需在 review 时检查此语义

## 怎么验证

- 用户未认证时通过自动授权获得 scope，检查其能调用的能力是否仅为 min_verf_level=0 的权限
- 敏感写/管理权限种子 `min_verf_level` 是否为 2

## 关联经验

[[is-system-no-permission-shortcut]] [[permission-seed-api-path-must-match-routes]]
