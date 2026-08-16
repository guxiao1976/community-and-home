# Plan Review — notice-xss-sanitize-and-frontend-fixes（业务有效性视角）

**审查维度**: 业务自洽 / 非功能（安全、性能、兼容性）/ 合规 / 架构冲突与依赖风险
**审查版本**: fallback:r0:rc1（P1.3）
**审查日期**: 2026-08-16

## 摘要
- 🔴 MUST FIX: 1 / 🟡 SHOULD FIX: 1 / 🔵 INFO: 3

## 事实核验（磁盘交叉验证）

| spec 断言 | 验证结果 |
|----------|---------|
| 移动端 `notice-detail.vue:23` 用 `<rich-text :nodes="notice.content">` 渲染公告正文 | ✅ 属实，且为 web/mobile 唯一富文本渲染点（`grep rich-text` 仅此一处；全仓无 v-html） |
| REST 为 RPC 代理，RPC 层净化可单点覆盖 RPC+REST 两入口 | ✅ 属实。`api/internal/logic/notice/createcontentpostlogic.go` / `updatecontentpostlogic.go` 为纯代理（`req.Text` → RPC `Text`）；路由仅 `/api/community/notices` 一组 |
| content 的 DB 写入点只有 Create/Update 两条 RPC 逻辑 | ✅ 属实。`ContentPostModel` 写路径仅 `createcontentpostlogic.go:158 InsertTx` 与 `updatecontentpostlogic.go:215 UpdateContentTx`；delete/pin/submit 均不重写正文；kafkapush producer/rescanner 仅写 push 状态列；moderation-service 无 content_posts 直写 |
| REST wire 键为 `content`、proto/DB 用 `text` | ✅ 属实（types.go `Text string json:"content"`，R2 兼容） |
| 首页双重加载根因（onMounted 显式 loadAll + watch(currentCommunityId) 并存；loadMemberships 内 getAppState 覆写 currentCommunityId） | ✅ 属实。notice.vue:344 watch / :350-352 onMounted / :356 下拉刷新；community.ts:79-86 getAppState 覆写 `currentCommunityId.value = serverCurrentId` |
| 登录 toast 覆盖缺陷（失败 toast 后紧跟成功 toast 覆盖） | ✅ 属实。auth-flow.ts:36 `showToast('获取用户资料失败', icon:none)` → :52 `showToast('登录成功', icon:success)` |
| 「稍后自动同步」口径有真实恢复路径支撑 | ✅ 属实。App.vue onLaunch 按 token 恢复 profile；my.vue/mine.vue 懒加载再拉取 |
| PC 无公告前台富文本渲染（out_of_scope 依据） | ✅ 属实。web/pc 无 v-html/rich-text 命中 |
| 未加入小区时不发请求（REQ-DBL-1 边界） | ✅ 属实。fetchNotices/fetchLostFound 已 `if (!cid) return` |
| bluemonday 为新增依赖（社区服务当前无 sanitize 类库） | ✅ 属实。go.mod 无 bluemonday；go 1.25 兼容；BSD-3 许可证合规 |

## 发现

### 🔴 MUST FIX

| # | 文件:行号/章节 | 问题 | 修复建议 |
|---|-------------|------|---------|
| 1 | `specs/xss-sanitization/spec.md` REQ-XSS-1/REQ-XSS-6 | **发布（submit）过渡路径绕过净化，安全目标存在可执行缺口**。现状核实：`updatecontentpostlogic.go` 的 submit 分支（`status==1`，仅 draft 可提交）直接 `UpdateStatusAndPublishTx(post.Id, StatusApproved)` 置为公开，**不重写正文、不经过净化器**。任何在净化上线**之前**已落库、未被净化过的正文（如净化前的存量 draft 草稿、或未来任何未净化写入路径入库的内容），可在上线后通过「提交发布」被置为公开且从未经过白名单净化 → 存储型 XSS 从 spec 声称的「every write path 覆盖」中漏出。REQ-XSS-6 仅覆盖「存量记录被**编辑**时按新语义净化」，对「存量草稿被**提交**发布」完全沉默。 | 二选一：(a) 在 `applyContentEdit` submit 分支（`UpdateStatusAndPublishTx` 前）对 `post.Text` 追加一次净化后再置公开；或 (b) 显式扩展 REQ-XSS-6 的残余风险声明，把「净化前存量草稿经 submit 发布」纳入 D5 接受范围并在 spec 明确记录决策，避免实现者误以为覆盖完备。建议 (a)，成本一行。 |

### 🟡 SHOULD FIX

| # | 文件:行号/章节 | 问题 | 建议 |
|---|-------------|------|------|
| 1 | `specs/xss-sanitization/spec.md` REQ-XSS-2 | 白名单标签/属性集合（含 img 全剔除、style 属性剔除）未与实际公告发布编辑器产出交叉核对——后台管理端编辑器本轮 out_of_scope，但它是合法富文本内容的实际产出源。若编辑器产出 `<h2>/<span style>/<table>/<img>` 等白名单外标签，存量合法公告一经编辑重存即被降级（样式/图片丢失），白名单「保留常用标签」的假设未获证据支撑。 | 在验收标准中增加一条：上线前对存量公告正文抽样（或核对后台编辑器实际产出 HTML），比对白名单标签/属性集合，确认无合法标签被误杀；比对结论记入 CHANGELOG/design 决策记录。 |

### 🔵 INFO

| # | 建议 |
|---|------|
| 1 | REQ-XSS-1「纯文本原样保存、不产生转义异常」的断言与 bluemonday 实际行为有细微出入：bluemonday 会对字面 `&`/`<`/`>` 做 HTML 实体转义（`A&B` → `A&amp;B`）。经 `<rich-text>` 渲染显示一致，但若未来任何路径以纯文本展示正文，会显示 `&amp;`。建议单测断言改为「渲染等价」而非「字节等价」。 |
| 2 | 新增依赖 bluemonday：明确锁版本（如 `go get github.com/microcosm-cc/bluemonday@v1.x`），go.sum 一并提交；BSD-3 许可证合规；确认与 go 1.25 编译矩阵无冲突。 |
| 3 | REQ-TOAST-1 合并文案「登录成功（资料加载失败，稍后自动同步）」约 19 字，超出 `uni.showToast` success 图标 toast 的标准展示宽度，长文案可能换行/截断。建议实现时确认 H5/小程序端渲染效果，必要时收敛文案长度或降级为 `icon:'none'`。 |

## 问题跟踪表（供下一轮校验）

| # | 问题 | 状态 |
|---|------|------|
| M1 | submit 过渡路径绕过净化（存量草稿发布缺口） | 待修复 |
| S1 | 白名单未与实际编辑器产出交叉核对 | 待处理 |

---
VERDICT: REVISION
---
