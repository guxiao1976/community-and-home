# 登录 toast 覆盖修复（auth-toast-fix）Specification

> **修订记录**：P1.4（REVISION 修复轮）。
> - clarity S1（「稍后自动同步」为空头承诺 + 合并 toast icon 语义未指定）→ REQ-TOAST-1 收敛文案（不承诺自动恢复）+ 固定 icon：失败时非 success（icon:none）— 已解决
> - validity S3（合并长文案在 success 图标下可能截断）→ 以 icon:none 缓解，并保留实现层渲染确认 — 已解决

## Purpose

消除 `web/mobile/src/utils/auth-flow.ts` `handleAuthSuccess` 中 profile 拉取失败提示被「登录成功」toast 覆盖的缺陷：失败时先 `showToast('获取用户资料失败')`，随后立即 `showToast('登录成功')`（`icon: 'success'`）把前者覆盖，用户无法得知资料未同步。本变更保证失败信息对用户可见（与成功提示合并展示），同时登录流程（存 token、判断小区、跳转）行为不回归。

## Requirements

### Requirement: REQ-TOAST-1 — 失败提示对用户可见且不被覆盖

When `getUserProfile()` fails during `handleAuthSuccess`, the user-facing feedback SHALL surface the profile-load failure and SHALL NOT be overwritten by a subsequent "登录成功" toast; the failure and success feedback SHALL be merged into a **single** visible toast（已确认 D2：合并提示口径）。The merged toast SHALL convey both login success and profile-load failure WITHOUT promising automatic recovery (D14)；它 SHALL use a non-success presentation (icon `none`, not the success checkmark) so it neither implies full success nor truncates on the target platform (H5)。The flow SHALL NOT emit a second standalone `showToast('登录成功')` after the merged toast within the same `handleAuthSuccess` run.

- 说明：示例文案「登录成功（资料加载失败）」，icon:none（文字型 toast，非 success 打勾）——success 图标既会误导「完全成功」又会因长文案换行/截断（validity 评审 S3）；不复用「稍后自动同步」措辞（profile 恢复仅发生于 App.vue onLaunch 启动时与 mine/my 页面懒加载，登录流程本身不承诺自动恢复，D14）。

#### Scenario: profile 拉取失败（正向主流程）
- **GIVEN** 登录成功、token 已写入，`getUserProfile()` 抛出异常
- **WHEN** `handleAuthSuccess` 进入 profile 拉取失败分支
- **THEN** 展示**单个**合并 toast（示例文案「登录成功（资料加载失败）」，icon:none），同时表达成功与失败、不承诺自动恢复；同一流程内不再弹出「登录成功」success toast 覆盖；`console.error` 保留错误留痕；随后仍按既有逻辑跳转（token 已存）

#### Scenario: profile 拉取成功（边界/正常）
- **GIVEN** 登录成功且 `getUserProfile()` 正常返回
- **WHEN** `handleAuthSuccess` 完成资料写入
- **THEN** 展示「登录成功」toast（icon:success），不附带失败信息；`userStore.setUser(user)` 正常设置

### Requirement: REQ-TOAST-2 — 登录流程行为不回归

The system SHALL preserve the existing login-success control flow: tokens saved, membership state resolved, and navigation performed (switchTab to `pages/notice/notice` when the user has communities, otherwise redirectTo `pages/join-community/join-community`) regardless of the profile-fetch outcome.

- 说明：profile 失败仅影响提示与 `user` 状态恢复（App.vue onLaunch / mine/my 页面懒加载会再恢复 profile），不得阻断跳转；toast 时序不得导致 `onCompleted` 回调丢失或跳转窗口延长；profile 失败不承诺自动恢复（与 REQ-TOAST-1 D14 一致）。

#### Scenario: 有小区用户 profile 失败仍进首页（异常但可控）
- **GIVEN** 登录成功，profile 拉取失败，但 `getUserMemberships()` 返回有小区
- **WHEN** `handleAuthSuccess` 完成（合并提示展示失败信息）
- **THEN** 800ms 后 `uni.switchTab` 到首页 `pages/notice/notice`，`onCompleted` 回调正常执行

#### Scenario: 无小区用户（边界）
- **GIVEN** 登录成功，profile 拉取失败或成功，`getUserMemberships()` 返回空
- **WHEN** `handleAuthSuccess` 完成
- **THEN** 跳转到 `pages/join-community/join-community`，不因 profile 失败而卡在登录流程

#### Scenario: 小区检查接口失败（依赖服务不可用）
- **GIVEN** `getUserMemberships()` 抛出异常
- **WHEN** `handleAuthSuccess` 捕获
- **THEN** 按既有兜底默认跳加入小区页（`console.warn` 留痕），toast 合并逻辑不受影响、不静默
