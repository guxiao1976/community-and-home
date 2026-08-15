# P1 移动端静默错误处理 Specification

## Purpose

消除移动端首页三个列表请求（通知 fetchNotices / 联络 fetchContacts / 寻失 fetchLostFound）的静默 catch 吞错，使任何加载失败对用户可见（toast）且可排查（控制台日志），避免「列表空白无报错」的排障黑洞；同时保持并发加载互不阻断。

## Requirements

### Requirement: REQ-P1-ERR-1 失败不再静默（唯一解释）

The system SHALL NOT silently swallow failures of `fetchNotices`, `fetchContacts`, or `fetchLostFound`; at the moment of failure each of the three SHALL invoke its section-specific failure toast (`uni.showToast`，如「通知加载失败」/「联络加载失败」/「寻失加载失败」，`icon: 'none'`) and write a `console.error` with the error object. **唯一解释（并发收敛）**：when multiple fetches fail concurrently, uni-app `showToast` single-instance replace semantics collapse the visible toast to the last-invoked one — this is acceptable and satisfies this requirement as "at least one visible toast remains, and every failed fetch emits its own `console.error`"; a per-failure toast being overwritten by a later one does NOT count as a violation. Success SHALL NOT show an error toast.

#### Scenario: 三请求全部成功（正向）

- **GIVEN** 用户已加入小区，网络正常，后端三个接口均成功
- **WHEN** `loadAll` 并发执行 `fetchNotices`/`fetchContacts`/`fetchLostFound`
- **THEN** 通知/联络/寻失三个区块正常渲染数据，不弹出错误 toast，无 `console.error`

#### Scenario: 寻失请求失败（网络异常）

- **GIVEN** 网络中断或后端不可达，仅 `fetchLostFound` 失败（其余两个请求成功）
- **WHEN** `loadAll` 执行
- **THEN** 寻失失败时刻触发 `uni.showToast`（如「寻失加载失败」）+ `console.error`（含错误对象与区块标识）；通知/联络区正常渲染且不弹错误 toast

#### Scenario: 通知请求失败（业务/HTTP 异常）

- **GIVEN** 后端对 `GET /api/community/notices` 返回业务错误或 5xx，仅 `fetchNotices` 失败
- **WHEN** `loadAll` 执行
- **THEN** 通知失败时刻触发 `uni.showToast`（如「通知加载失败」）+ `console.error`，通知区不静默空白，其余区块正常

#### Scenario: 联络请求失败（HTTP 异常）

- **GIVEN** 后端对 `GET /api/community/contacts` 返回非 2xx，仅 `fetchContacts` 失败
- **WHEN** `loadAll` 执行
- **THEN** 联络失败时刻触发 `uni.showToast`（如「联络加载失败」）+ `console.error`，联络区有可见反馈，其余区块正常

#### Scenario: 三请求并发全部失败（toast 收敛的唯一解释）

- **GIVEN** 后端不可达或网络中断（本变更要修的原始 bug 场景：`fetchNotices`/`fetchContacts`/`fetchLostFound` 同时失败）
- **WHEN** `loadAll` 并发执行三个请求
- **THEN** 每个失败时刻各触发一次 `uni.showToast`（共三次调用）与各一次 `console.error`（共三次）；可见 toast 收敛为最后一次调用（单实例替换），页面仍有至少一个可见失败提示；此收敛为唯一解释，不算违反「失败不再静默」

### Requirement: REQ-P1-ERR-2 失败不阻断并发加载

The error handling SHALL NOT rethrow after surfacing the failure, so that the concurrent `Promise.all` in `loadAll` continues for unaffected fetches and successfully loaded sections still render. When all three fetches fail concurrently, the page SHALL NOT crash, `loadAll` SHALL complete (loading state resets), previously rendered section data SHALL NOT be cleared, and every failed fetch SHALL still emit its own `console.error`; the visible toast collapses to the last-invoked one via uni-app `showToast` single-instance replace semantics (repeated failures do not stack) — at least one visible toast remains, no additional accumulation guard is required.

#### Scenario: 局部失败仍渲染成功区块（正向）

- **GIVEN** `fetchNotices` 失败，但 `fetchContacts`/`fetchLostFound` 成功
- **WHEN** `loadAll` 执行 `Promise.all([fetchNotices(), fetchContacts(), fetchLostFound()])`
- **THEN** 通知区提示失败、联络与寻失区正常渲染（catch 内部 toast + console 后不向上抛错，`Promise.all` 不被单个失败中断导致整页空白）

#### Scenario: 三请求并发全失败（核心触发场景）

- **GIVEN** 后端不可达/网络中断，`fetchNotices`/`fetchContacts`/`fetchLostFound` 三个请求同时失败（即本变更要修的原始 bug 场景）
- **WHEN** `loadAll` 并发执行三个请求
- **THEN** 至少一次可见 `uni.showToast`（收敛为最后一次调用）+ 三次 `console.error`（每失败一次）；页面不崩溃、`loadAll` 正常结束（loading 复位）、此前已渲染的区块数据不被清空；若之后部分请求恢复成功，成功区块正常渲染

#### Scenario: 重复触发不累积误报（边界）

- **GIVEN** 用户下拉刷新触发第二次 `loadAll`
- **WHEN** 某请求再次失败
- **THEN** 再次 toast + console 提示（幂等反馈）；toast 由单实例替换语义保证不出现重复堆叠的错误提示块
