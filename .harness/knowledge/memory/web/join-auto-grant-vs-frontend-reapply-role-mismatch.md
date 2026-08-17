---
triggers: ["joinCommunity", "applyRole", "加入小区", "角色申请", "ownership", "权属", "自动授权", "重复授权", "owner", "tenant"]
service: web/mobile
severity: should-follow
type: pitfall
status: active
created: 2026-08-17
updated: 2026-08-17
last_applied: null
apply_count: 0
---

# 加入流程后端已按权属自动授权 owner/tenant，前端不应再无条件自申请 owner

## 为什么会有这条经验

后端 `JoinCommunity` 已按 ownership 经 `assignCommunityRole` 自动授权 owner（自有）或 tenant（租住）角色（status=0，失败补偿回滚 membership）。前端加入流程若再无条件调 `applyRole({role_code:'owner'})`：
- OWNED 用户：重复产生未认证 owner 授权（冗余）；
- RENTED 用户：租户在自动 tenant 之外自申请到未认证 owner 授权——后端 ApplyRole 仅校验「有效成员关系」不校验权属声明，会放行（越权）。

`auto-grant-unverified-grant-confers-scope-level0`：status=0 未认证 grant 立即产生 community 数据范围 + level-0 能力，权属不符的授权即越权。

## 怎么做

1. 加入流程仅调 `joinCommunity`，不重复 `applyRole('owner')`——后端按 ownership 自动授权 owner/tenant
2. 若确需按 ownership 分支 role_code（OWNED→owner / RENTED→tenant），也不要重复授权
3. 修改加入/申请角色流程时检查是否存在「后端自动授权 + 前端重复申请」的双重授权

案例：`web/mobile/src/pages/join-residence/join-residence.vue` `confirmJoin` 已移除多余的 `applyRole('owner')`。

## 怎么验证

- 加入流程代码仅含 `joinCommunity`，无 `applyRole('owner')`
- grep 变更中「joinCommunity 后紧跟 applyRole」的组合
