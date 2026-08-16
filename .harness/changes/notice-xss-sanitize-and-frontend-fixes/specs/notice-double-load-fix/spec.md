# 首页数据加载去重（notice-double-load-fix）Specification

> **修订记录**：P1.4（REVISION 修复轮）。
> - structure INFO-2（守卫解除时机不可判定）→ REQ-DBL-1 以 `membershipsResolved` 布尔标志固化守卫条件 — 已解决
> - coverage S3（loadMemberships 整体失败降级边界未定义）→ REQ-DBL-1 新增降级场景 — 已解决

## Purpose

消除移动端首页 `notice.vue` 初始进入时的重复数据加载。现状：`onMounted` 先 `await communityStore.loadMemberships()` 再显式 `loadAll()`，同时 `watch(currentCommunityId) → loadAll()`；`loadMemberships` 内 `getAppState` 服务端权威覆写 `currentCommunityId` 时会触发 watch 一次，加上 onMounted 显式一次，同一批接口（通知列表 + 寻失列表）被拉两遍。本变更保证初始进入首页时同批接口只拉一遍，同时保留用户切换小区、下拉刷新、失败留痕等既有行为不回归。

## Requirements

### Requirement: REQ-DBL-1 — 初始进入单次加载

The system SHALL load the homepage data (notice list + lost-and-found list, the same `loadAll()` batch) **exactly once** on initial entry to `pages/notice/notice`, even when `loadMemberships()` overwrites `currentCommunityId` with the server-authoritative value from `getAppState`.

- 采用「显式单次加载 + watch 首载守卫」（D1）。守卫条件可判定（结构评审 INFO-2 收敛）：模块级布尔标志 `membershipsResolved`，默认 `false`；`watch(currentCommunityId)` 处理器在 `!membershipsResolved` 时直接 return（忽略变更，含 `loadMemberships` 内 getAppState 覆写触发的那次）；该标志在 `loadMemberships()` 结束时（无论成功/失败，finally 等价路径）置为 `true`；onMounted 在 `loadMemberships()` 结束后且存在当前小区（`hasCommunities == true`）时才显式触发一次 `loadAll()`。

#### Scenario: 本地存储值即服务端权威值（正向主流程）
- **GIVEN** 用户已登录且已加入小区，本地存储 `current_community_id = C1`，服务端 `getAppState` 返回同一 `C1`
- **WHEN** 首次进入首页，onMounted 执行 `loadMemberships()` 后触发初始加载
- **THEN** 通知列表与寻失列表接口各只请求一次（共 1 次 `loadAll` 批次），`currentCommunityId` 保持 `C1`，watch 不产生额外触发

#### Scenario: 服务端权威覆写导致 currentCommunityId 变更（异常/关键场景）
- **GIVEN** 本地存储 `current_community_id = C1`（陈旧），服务端 `getAppState` 权威返回 `C2`，`C2` 在 memberships 内
- **WHEN** onMounted 内 `loadMemberships()` 执行，将 `currentCommunityId` 覆写为 `C2`
- **THEN** 首页数据仍只加载一次（以最终 `C2` 为维度的 1 次 `loadAll` 批次），不因 `C1→C2` 的覆写再触发第二次加载——通知列表与寻失列表接口合计仍只各请求一次（覆写触发的 watch 被 `membershipsResolved == false` 守卫忽略）

#### Scenario: 用户未加入任何小区（边界输入）
- **GIVEN** `loadMemberships()` 成功后 `communities` 为空（`hasCommunities == false`）
- **WHEN** 首页挂载完成
- **THEN** 不发起通知/寻失数据加载，页面展示「请先加入小区」空态引导（既有行为不变），且不因守卫缺失而抛错

#### Scenario: loadMemberships 整体失败（降级，网络异常）
- **GIVEN** 首页进入时 `loadMemberships()` 网络失败，`communities` 被置为 `[]`、未抛错，`currentCommunityId` 可能保留陈旧本地值
- **WHEN** `loadMemberships()` 结束（`membershipsResolved` 置 `true`）
- **THEN** 不发起通知/寻失数据加载（**不以陈旧 cid 发请求**），页面按无小区空态引导（既有降级行为）；守卫已就绪，后续用户切换小区仍正常触发单次加载；不 double-load、不卡 loading

### Requirement: REQ-DBL-2 — 用户切换小区单次加载

The system SHALL trigger **exactly one** data reload per user-initiated community switch (`onCommunitySwitch` → `switchCommunity` → `currentCommunityId` 变更), loading the newly selected community's data once.

#### Scenario: 切换小区（正向）
- **GIVEN** 首页已加载小区 `C1` 数据，用户点击小区切换器选择 `C2`
- **WHEN** `switchCommunity(C2)` 成功，`currentCommunityId` 变为 `C2`
- **THEN** watch 触发恰好一次 `loadAll()`，通知与寻失列表加载 `C2` 的数据；不重复加载、不加载 `C1` 旧数据

#### Scenario: 切换失败（权限异常）
- **GIVEN** 用户尝试切换到不在数据范围的小区（后端返回 10015）
- **WHEN** `switchCommunity` 抛出 10015 错误
- **THEN** 页面提示「目标小区不在你的数据范围」，`currentCommunityId` 保持原小区不变，不触发任何数据加载

### Requirement: REQ-DBL-3 — 拉取失败留痕与下拉刷新不回归

The fix SHALL NOT regress existing error-handling and refresh behavior: on a data-load failure, the page SHALL retain `console.error` + block-level failure toast（「通知加载失败」/「寻失加载失败」）；pull-to-refresh SHALL still perform a fresh `loadAll()`.

#### Scenario: 数据加载失败留痕（异常）
- **GIVEN** 首次进入或切换小区后通知列表接口返回失败
- **WHEN** 加载失败
- **THEN** 页面 `console.error` 记录错误并弹出「通知加载失败」提示，不静默吞错；失败不影响另一区块（寻失）的正常加载与展示

#### Scenario: 下拉刷新（交互）
- **GIVEN** 首页已展示某小区数据
- **WHEN** 用户下拉触发 `onPullDownRefresh`
- **THEN** 执行一次全新 `loadAll()` 并结束刷新动画（`uni.stopPullDownRefresh`），守卫不得拦截下拉触发的加载
