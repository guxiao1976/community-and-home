# Plan Review — notice-xss-sanitize-and-frontend-fixes（覆盖完整性视角）

**审查维度**: 需求覆盖 / 场景完整性 / 边界识别
**审查版本**: fallback:r0:rc1（首次评审轮，按磁盘最新内容独立审查，未沿用旧轮结论）
**审查对象**: proposal.md + specs/xss-sanitization + specs/notice-double-load-fix + specs/auth-toast-fix

## 摘要
- 🔴 MUST FIX: 0 / 🟡 SHOULD FIX: 3 / 🔵 INFO: 3

## 覆盖实证（非空泛，逐项对代码核对）

| 覆盖点 | 核对结果 | 证据 |
|--------|---------|------|
| 写路径覆盖 | ✅ 单点净化成立 | `content_posts.text` 仅两个写入口：`rpc/internal/logic/notice/createcontentpostlogic.go`（InsertTx）、`updatecontentpostlogic.go`（UpdateContentTx）。API 层 `api/internal/logic/notice/createcontentpostlogic.go` 为 RPC 纯代理（`ContentPostServiceRpc.CreateContentPost`），update 同构；API svcCtx 虽注入 ContentPostModel，但仅 `api/.../getcontentpostlogic.go` 用于 `ResolveReadableCommunityForCompat` 读取兼容，全仓无 API 层直写（`grep .UpdateContent(` 0 调用、无 API 层 InsertTx）。model 非事务 `UpdateContent` 无任何调用方。 |
| 渲染面覆盖 | ✅ D3 范围成立 | 全仓唯一富文本渲染 `web/mobile/src/pages/notice-detail/notice-detail.vue:23 <rich-text :nodes="notice.content">`；title（notice.vue/notice-detail.vue）均 `{{ }}` 插值自动转义；寻失 description 无 v-html/rich-text 渲染点 → title / 寻失 description 排除不构成遗留存储型 XSS 向量。 |
| 生命周期覆盖 | ✅ 无绕净化二次写入 | Create 即净化，draft/submitted 状态内容均已净化；Update submit（status=1）仅改 status/published_at 不触 text；moderation 消费者对 notice 任务精确跳过、不回调不改文本（task_handler.go D4/D21 注释）。 |
| 决策点覆盖 | ✅ | D1（watch 首载守卫）→ REQ-DBL-1；D2（toast 合并口径）→ REQ-TOAST-1；D3（仅正文范围）→ REQ-XSS-5 + out_of_scope；D4（本服务单例）→ REQ-XSS-3；D5（不回填）→ REQ-XSS-6；D6（img 全剔）→ REQ-XSS-2。 |
| 双载前提 | ✅ | `stores/community.ts` loadMemberships 内 getAppState 服务端权威覆写 currentCommunityId（L78-89）——与 proposal 所述双重加载成因一致。 |
| 每 REQ 场景 | ✅ 基本满足 | 各 Requirement 均含 ≥1 正向 + ≥1 异常/边界场景（明细见 SHOW FIX 3 遗漏项）。 |

## 发现

### 🔴 MUST FIX
无。

### 🟡 SHOULD FIX
| # | 文件:章节 | 问题 | 建议 |
|---|---------|------|------|
| 1 | specs/xss-sanitization/spec.md — REQ-XSS-1 边界场景 / REQ-XSS-4 | 「输入非空但净化后为空」写路径语义未定义：正文仅含 `<img>`/`<iframe>`（如 `<img src=x onerror=alert(1)>`，img 无内文、净化后剩空）时，Create 的预净化非空校验（createcontentpostlogic.go L50 `in.Text==""→080005`）与 Update（updatecontentpostlogic.go L169-174）均通过，但落库后 content 为空串。REQ-XSS-4 仅覆盖「空后仍可读」读路径，未定义写路径应「存空串」还是「净化后校验非空→080005 拒绝」，与「内容不能为空」不变量语义存在未收敛歧义（可能产生有标题无正文的已发布公告）。 | 新增写路径边界场景：输入非空但净化后为空 → 明确「落库存空串/纯文本（对齐 REQ-XSS-4）」或「净化后再做非空校验 → 080005 拒绝」，二选一并在 spec 固定，实现者不会分叉。 |
| 2 | specs/xss-sanitization/spec.md — REQ-XSS-6 场景 2 | Update 的 Text 为 proto3 optional（presence 语义），Text 未携带时正文保持现值不重净化；REQ-XSS-6 场景 2 仅覆盖「重新提交正文（携带新富文本）」的情形。存量恶意记录若被编辑但前端未重发 text（仅改 title/scope/attachments），恶意正文不会随编辑被清理——与「存量记录被编辑时按新语义净化」的表述存在条件缺口（虽属已接受的 D5 残余风险子集）。 | 显式补充 Text 未携带的 Update 场景：要么声明「Text 未携带时重净化既有 text（幂等，净化器单例支持）」，要么声明「存量正文仅在 text 被重新提交时净化，其余归入 D5 已接受残余风险」，确保实现与残余风险声明一致。 |
| 3 | specs/notice-double-load-fix/spec.md — REQ-DBL-1 / REQ-DBL-3 | 未覆盖 `loadMemberships()` 整体失败/降级边界：`stores/community.ts` L101-103 catch 置 `communities=[]` 后不抛错，此时 currentCommunityId 可能保留陈旧本地值；watch 首载守卫 + onMounted 显式 loadAll 在此情形下是否仍执行、是否会以陈旧 cid 误拉数据、是否会卡 loading，spec 均未定义。REQ-DBL-1 场景 3 仅覆盖「memberships 成功但为空」。 | 新增边界场景：memberships 拉取失败（网络错误 → communities=[]）→ 明确「守卫/显式加载在降级路径仍按无小区空态处理、不 double-load、不以陈旧 cid 发请求」。 |

### 🔵 INFO
| # | 建议 |
|---|------|
| 1 | REQ-XSS-4 需求文本列出 detail/list/marquee 三读路径，但仅 detail（GetContentPost）有场景测试；list/marquee 共享同一 DB 值，建议补一条 list/marquee 返回净化内容的验收说明以闭环（低优先）。 |
| 2 | change 目录缺 `request.md`（评审输入清单预期文件），当前以 proposal.md 承担原始需求角色；建议补 request.md 或在 .change.yaml 标注来源，便于后续评审对照初衷。 |
| 3 | REQ-TOAST-1 合并 toast「登录成功（资料加载失败…）」在 `icon:'success'` 下长文案于非 H5 端（微信小程序）可能被截断；本变更目标平台 H5 可接受，建议在实现层确认 uni.showToast title 长度与 icon 组合（低优先）。 |

---
VERDICT: APPROVED
---
