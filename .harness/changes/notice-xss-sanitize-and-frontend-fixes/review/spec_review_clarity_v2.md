# Plan Review — notice-xss-sanitize-and-frontend-fixes（清晰可执行视角）

**审查维度**: 粒度 / 歧义 / 一致性（SHALL/MUST 唯一解释、Scenario 可复现行为、术语一致）
**审查版本**: P1.3（fallback:r1:rc1，磁盘最新 spec 为 P1.4 修订轮）
**审查对象（磁盘）**: `.harness/changes/notice-xss-sanitize-and-frontend-fixes/{.change.yaml, request.md, proposal.md, specs/xss-sanitization/spec.md, specs/notice-double-load-fix/spec.md, specs/auth-toast-fix/spec.md}`

> 本视角独立按磁盘最新内容重新审查（P1.4 修订轮）。已对照实际代码核验 spec 事实性前提与机制可行性：
> - `createcontentpostlogic.go:50` `in.Title == "" || in.Text == "" → CodeInvalidParam(080005)`，与 REQ-XSS-1 顺序钉死前提一致；REST 为 RPC 代理，RPC 层单点净化可覆盖。
> - `updatecontentpostlogic.go:136` submit 分支 `UpdateStatusAndPublishTx` 现只改状态不写正文 → D9 需在事务内追加净化后正文写入，spec 已表述。
> - `notice-detail.vue` `<rich-text :nodes="notice.content">` 属实；title 走 `{{ }}` 插值，REQ-XSS-5 title 不净化安全前提成立。
> - **上轮 v1 评审 4 项 MUST/SHOULD 已逐条修订**：M1（白名单穷举）→ REQ-XSS-2 穷举 + REQ-XSS-8 交叉核对；M2（净化/非空校验顺序）→ REQ-XSS-1 钉死 D7；S1（toast 文案+icon）→ REQ-TOAST-1 收敛 icon:none + 不承诺自动恢复；INFO-1（request.md 缺失）→ 已补。均已解决。
> - 本轮在 P1.4 修订产物上发现 2 处新 MUST FIX + 2 处 SHOULD FIX（见下）。

## 摘要
- 🔴 MUST FIX: 2 / 🟡 SHOULD FIX: 2 / 🔵 INFO: 1

## 发现

### 🔴 MUST FIX

| # | 文件:章节 | 问题 | 修复建议 |
|---|----------|------|---------|
| M1 | specs/notice-double-load-fix/spec.md §REQ-DBL-1（守卫机制） | 守卫标志 `membershipsResolved` 的**置位时点未钉死**，两种实现产生两种行为，其中「字面正确」的实现反而复现双重加载。现状：`community.ts loadMemberships` 内 `currentCommunityId.value = serverCurrentId`（C2 覆写）后紧跟 `return`（无 await）；Vue `watch` 默认 `flush:'pre'`，其回调在调度器微任务队列中、**早于** onMounted 续体执行。若按 spec「finally 等价路径」把标志**放进 loadMemberships 内部**同步置 true → 覆写触发的 watch 回调运行时标志已为 true → 不被忽略 → `loadAll()` 触发一次；随后 onMounted 显式 `loadAll()` 再触发一次 → **双重加载，恰好是本变更要修的缺陷**。若把标志放在 onMounted 的 `await loadMemberships()` 之后（微任务序晚于 watch flush）才成立。spec 未指明 finally 是在 loadMemberships 内还是在调用方，实现者无法确定唯一行为，REQ-DBL-1「服务端权威覆写只加载一次」的 THEN 不被机制保证。 | 在 REQ-DBL-1 显式钉死置位时点：`membershipsResolved = true` 必须置于 onMounted 中 `await loadMemberships()` 返回**之后**（此时 loadMemberships 内 getAppState 覆写所触发的 watch 回调已在同一微任务批次先执行并被 `!membershipsResolved` 守卫忽略）；并明示「不得在 loadMemberships 内部（finally）同步置 true，否则守卫失效复现双载」。若嫌对微任务序依赖过强，可改为不依赖时序的设计（如首载仅由 onMounted 显式触发一次、watch 以 `firstLoadDone` 守卫，或 loadMemberships 返回后 `await nextTick()` 再置位），并在 spec 补一条该时序的断言（覆写场景接口只各请求一次）。 |
| M2 | specs/xss-sanitization/spec.md §REQ-XSS-2 规则「a：target/rel」vs Scenario「合法链接保留」 | 规则与场景矛盾，白名单属性策略行为不可判定。规则写「`target` 仅当同步强制 `rel="noopener noreferrer"` 时保留，否则 `target` 移除」；但「合法链接保留」场景输入为 `<a href="https://example.com/notice" target="_blank" rel="noopener">`（rel 单令牌 `noopener`，非 `noopener noreferrer`），THEN 断言「含 rel/target 完整保留」。字面执行规则 → `target` 移除（违背场景 THEN）；想满足场景 → 需净化器把 rel 改写为 `noopener noreferrer`（「完整保留」又与之矛盾）。两种实现两种产出，按场景写的测试与按规则写的实现互相冲突。 | 二选一钉死：(a) 场景输入改为 `rel="noopener noreferrer"`，与规则「已强制该 rel」自洽；(b) 或规则补充改写语义「净化器对带 `target` 的 `a` 强制重写 `rel="noopener noreferrer"`（覆盖输入 rel 值）」，场景 THEN 相应改为「rel 被改写为 `noopener noreferrer`、target 保留」。并确认与 bluemonday 具体 API（如 `AllowTargetBlankWithRel("noopener noreferrer")`）的映射在 spec 说明，避免实现者任选。 |

### 🟡 SHOULD FIX

| # | 文件:章节 | 问题 | 建议 |
|---|----------|------|------|
| S1 | specs/notice-double-load-fix/spec.md §REQ-DBL-1 Scenario「loadMemberships 整体失败」 | 场景 THEN「页面按无小区空态引导…不卡 loading」，但所给机制（无小区/失败时**不发起 loadAll**）不产出该结果：`notice.vue` 中 `loading` 初值 `ref(true)`，仅 `loadAll()` 会置 false；跳过 loadAll → 模板 `v-if="loading"` 永远渲染骨架屏而非 `v-else-if="!hasCommunities"` 空态。场景断言的行为缺一个未说明的步骤（跳过加载时置 `loading=false`）。 | 在 REQ-DBL-1 补一句：无小区/失败跳过加载分支须将 `loading` 置 false（或以等价方式保证空态可见、不卡骨架屏），并在该场景 THEN 增补「loading=false、骨架屏消失」断言。 |
| S2 | specs/xss-sanitization/spec.md §REQ-XSS-8 Scenario「发现白名单外合法标签」 | 示例把 `<h2>` 列为「白名单外合法标签」，但 `h2` 已在 REQ-XSS-2 允许标签（h1–h6）内，示例自相矛盾，会误导实现者对白名单边界的判断。 | 示例改为真正白名单外的标签（如 `<table>`、`<span style="...">`），或改为「如 `<h2>` 若不在白名单（示例，实际已含）」的表述；保持与 REQ-XSS-2 穷举一致。 |

### 🔵 INFO

| # | 建议 |
|---|------|
| INFO-1 | REQ-DBL-1 引用 `hasCommunities` 未定义来源，建议标注「community store getter：`communities.length > 0`」，消除实现者猜测。 |

---

## 问题跟踪表（供下轮校验）

| 编号 | 问题 | 状态 |
|------|------|------|
| M1 | REQ-DBL-1 守卫置位时点未钉死（复现双载风险） | 待修复 |
| M2 | REQ-XSS-2 a/rel/target 规则与场景矛盾 | 待修复 |
| S1 | REQ-DBL-1 失败降级「不卡 loading」缺机制 | 待修复 |
| S2 | REQ-XSS-8 示例 h2 实为白名单内 | 待修复 |

---
VERDICT: REVISION
---
