# Plan Review — notice-xss-sanitize-and-frontend-fixes（清晰可执行视角）

**审查维度**: 粒度 / 歧义 / 一致性（SHALL/MUST 唯一解释、Scenario 可复现行为、术语一致）
**审查版本**: P1.3（fallback:r0:rc1）
**审查对象（磁盘）**: `.harness/changes/notice-xss-sanitize-and-frontend-fixes/{.change.yaml, proposal.md, specs/xss-sanitization/spec.md, specs/notice-double-load-fix/spec.md, specs/auth-toast-fix/spec.md}`

> 本视角独立按磁盘最新内容重新审查。已对照实际代码核验 spec 的事实性前提：
> - 写路径仅 CreateContentPost（InsertTx）/ UpdateContentPost（UpdateContentTx）两处，REST 为 RPC 代理（api-proto 含 gateway），RPC 层单点净化可覆盖两入口 — spec 陈述准确。
> - 字段命名三重映射（DB 列 `text` / proto-RPC 字段 `text` / REST wire 键 `content` / 前端 `notice.content`）在 spec 中已说明，REST wire 键 `content` 属实（proto 注释「原 content 改名，wire 由 REST 层映射 content」），术语一致。
> - 双加载缺陷描述与 notice.vue 现码一致（onMounted await loadMemberships()+loadAll() 与 watch 并存）；auth-flow.ts 现码确实先「获取用户资料失败」再被「登录成功」覆盖。
> - 注意：`.harness/changes/<name>/request.md` 缺失，只能凭 .change.yaml 标题核验初衷（见 INFO-1）。

## 摘要
- 🔴 MUST FIX: 2 / 🟡 SHOULD FIX: 1 / 🔵 INFO: 2

## 发现

### 🔴 MUST FIX

| # | 文件:章节 | 问题 | 修复建议 |
|---|----------|------|---------|
| M1 | specs/xss-sanitization/spec.md §REQ-XSS-2 及 Scenario「事件属性与危险 scheme 剔除」 | 白名单标签集未穷举，且与自身 Scenario 矛盾，导致合法内容行为不可判定。REQ-XSS-2 仅列「p/br/strong/em（及 b/i/u/ul/ol/li/blockquote 等）」，以「等」留白；但 Scenario 显式以 `<a href="javascript:..." onclick=...>点我</a>` 为输入，断言 onclick/javascript: href 被剔除（隐含 `<a>` 被保留仅净化属性）；同时以 `<div style=...>` 为输入，断言「style 属性被剔除」（隐含 div 保留）。`a`、`div` 均不在已列举集合内。实现者若按字面枚举：`<a>` 整体被剥除（不在白名单）→ 合法链接 `<a href="https://example.com">官网</a>` 的存库/展示行为未定义；`<div>` 整体被剥除 vs 保留仅去 style，两种实现产出不同。Scenario 的「或移除」只覆盖 javascript: 情形，未覆盖合法 href 的命运 → 同一 spec 两种实现两种行为。 | 在 REQ-XSS-2 中穷举完整允许标签集（显式含 `a`、`div` 的去留）：若保留 `<a>`，则同时枚举其允许属性（href 及其允许 scheme 如 http/https/mailto，以及 rel/target 等）、`javascript:`/`data:` 处理方式；若 `<div>` 保留，明确其允许属性（如剔除全部 style）。并让所有 Scenario 与该枚举一致（如为 `<a href="https://example.com">官网</a>` 增加一个「合法链接保留、仅危险 scheme/事件属性被剔除」的正向场景）。 |
| M2 | specs/xss-sanitization/spec.md §REQ-XSS-1 Scenario 3 与 §REQ-XSS-4 Scenario 2 | 净化器与既有非空校验的执行顺序未指定，两个 Scenario 对同一输入 `<script>...</script><iframe ...></iframe>`（净化后为空）隐含互相排斥的行为：REQ-XSS-1 Scenario 3 称「空正文仍按既有校验拒绝（080005）」，REQ-XSS-4 Scenario 2 则预设该内容「落库后被读路径返回空串」。若实现者在非空校验**之前**净化 → script-only 正文变空 → 080005 拒绝（违反 REQ-XSS-4-2 的可读前提）；若**之后**净化 → 非空原始值通过 → 净化后为空串并落库（满足 REQ-XSS-4-2）。spec 未给唯一解释。 | 显式钉死顺序：非空校验（080005）以**原始正文**先行判定（语义不变），净化器在非空校验通过后、DB 落库前执行；并补充声明「正文净化后为空的处理」为二选一且唯一（建议：接受空串落库，与 REQ-XSS-4-2 一致；若改判 080005 拒绝则同步修改 REQ-XSS-4-2）。 |

### 🟡 SHOULD FIX

| # | 文件:章节 | 问题 | 建议 |
|---|----------|------|------|
| S1 | specs/auth-toast-fix/spec.md §REQ-TOAST-1 示例文案与 §REQ-TOAST-2 说明 | 示例合并文案「登录成功（资料加载失败，稍后自动同步）」中的「稍后自动同步」与实际行为契约不符：profile 恢复仅发生在 App.vue `onLaunch`（启动时，登录前已触发过，登录后不会重跑）与 mine/my 页面懒加载时；常规「启动后登录」流程下 profile 并不会自动恢复，「稍后自动同步」是空头承诺，且合并 toast 的 icon 语义未指定（保留 success 打勾还是 icon:none）。 | 收敛口径：要么把「自动恢复」落实为行为契约（如跳转后页面/下一路由触发 restoreUserProfile），要么示例文案改为不承诺自动同步（如「登录成功（资料加载失败，稍后重试）」），并在 spec 明确合并 toast 的 icon 取值。 |

### 🔵 INFO

| # | 建议 |
|---|------|
| INFO-1 | 变更目录缺 `request.md`，无法从标准输入核验 spec 是否偏离用户原始需求（仅 .change.yaml 标题可用）。建议补 request.md 或由 coverage 视角记录，不阻塞本视角。 |
| INFO-2 | proposal 称「评审核对白名单与公告编辑器实际产出标签集合」，但后台编辑器改造在 out_of_scope，本轮无法比对。建议 spec 直接冻结允许清单（同 M1），避免与编辑器产出漂移成为后续技术债。 |

---

VERDICT: REVISION
---
