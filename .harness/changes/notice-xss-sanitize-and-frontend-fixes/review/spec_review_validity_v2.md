# Plan Review — notice-xss-sanitize-and-frontend-fixes（业务有效性视角）

**审查维度**: 业务自洽 / 非功能（安全、性能、兼容性）/ 合规 / 架构冲突与依赖风险
**审查版本**: fallback:r1:rc1（P1.4 修复轮，哈希较 r0:rc1 已变更，独立重新审查磁盘最新内容）
**审查日期**: 2026-08-16

## 摘要
- 🔴 MUST FIX: 0 / 🟡 SHOULD FIX: 1 / 🔵 INFO: 2

## 上轮 validity 发现核对（r0:rc1 → r1:rc1）

| 上轮问题 | 状态 | P1.4 落点 |
|---------|:---:|---------|
| M1 submit 过渡路径绕过净化（存量草稿发布缺口） | ✅ 已解决 | REQ-XSS-1 入口枚举第 3 项 + REQ-XSS-6「同一事务净化正文 + 置公开」+ 正向场景「存量 draft 经 submit 发布时净化」。代码核验：`updatecontentpostlogic.go:131-154` submit 分支现仅 `UpdateStatusAndPublishTx`（status-only、不写正文），spec 要求新增「置公开前净化既有正文、同事务写入」与实现路径一致且可行 |
| S1 白名单未与实际编辑器产出交叉核对 | ✅ 已解决 | REQ-XSS-8 + D12：上线前抽样比对、发现白名单外合法标签须发布前修订，结论入 CHANGELOG/决策记录 |
| INFO 1 纯文本断言字节 vs 渲染等价 | ✅ 已解决 | D13：断言以渲染等价为准，明确净化器实体转义属允许行为 |
| INFO 2 bluemonday 锁版本 | ✅ 已解决 | REQ-XSS-7：SHALL 锁定明确版本、BSD-3、go 1.25 兼容、go.mod/go.sum 一并提交 |
| INFO 3 合并 toast 长文案截断 | ✅ 已解决 | REQ-TOAST-1：合并 toast 固定 icon:none（非 success），缓解截断并避免误导 |

## 本轮事实核验（磁盘交叉验证）

| spec 断言 | 验证结果 |
|----------|---------|
| submit 分支现不重写正文、置公开即可发布存量未净化正文 | ✅ 属实。`updatecontentpostlogic.go:137` `UpdateStatusAndPublishTx`（status-only）；正文写仅在 Create `InsertTx` / Update `UpdateContentTx`（r0 已核） |
| 首页双重加载根因（watch + onMounted 并存；getAppState 覆写触发 watch） | ✅ 属实。`notice.vue:344-352` watch→loadAll / onMounted→loadMemberships+loadAll；`stores/community.ts:86` `currentCommunityId.value = serverCurrentId` |
| loadMemberships 整体失败时 communities 置空、未抛错（REQ-DBL-1 降级前提） | ✅ 属实。`stores/community.ts:101-102` `catch { communities.value = [] }` |
| 登录 toast 覆盖缺陷 | ✅ 属实。`utils/auth-flow.ts:36` 失败 toast(icon:none) → `:52` 成功 toast(icon:success,1500ms) 覆盖 |
| REST 为 RPC 代理，RPC 层单点净化覆盖 RPC+REST 两入口 | ✅ 属实（r0 已核：api 层纯代理，路由仅 `/api/community/notices`） |
| 存储型 XSS 渲染面（`notice-detail.vue:23` `<rich-text>` 为移动端唯一富文本渲染点） | ✅ 属实（r0 已核，全仓无 v-html） |

## 发现

### 🔴 MUST FIX
无。

### 🟡 SHOULD FIX

| # | 文件:行号/章节 | 问题 | 建议 |
|---|-------------|------|------|
| 1 | `specs/xss-sanitization/spec.md` REQ-XSS-2 / REQ-XSS-8 | **`a` 的 href scheme 白名单缺 `tel:`**：社区公告高频含「联系电话」类内容，后台编辑器若产出 `<a href="tel:xxx">`，该 href 会被剔除降级为纯文本，存量合法公告一经编辑重存即失去「点击拨号」。属常见合法业务内容，非 XSS 面（tel: 不执行脚本）。 | 在 REQ-XSS-8 上线前抽样比对时**显式评估 `tel:`（及编辑器可能产出的其他 scheme，如 `weixin:` 等）**：若既有公告/编辑器产出命中，建议在 D8 白名单纳入 `tel:`（经安全评估，无脚本执行面）或记录明确决策；比对结论入 CHANGELOG/决策记录。 |

### 🔵 INFO

| # | 建议 |
|---|------|
| 1 | REQ-XSS-2 对危险 scheme 的 `a` 表述有轻微不一致：正文规则「该 href 属性移除、`a` 其余属性与文本保留」vs 场景「链接降级为无 href 的纯文本节点」。建议实现以场景为准（剔除 href，保留安全文本；target/rel 一并移除以免悬空），供 clarity 视角对齐措辞。 |
| 2 | 存量已发布未净化公告仍会渲染执行（D5 不做回填，Q5=0 用户拍板）：已明确 out_of_scope + 文档记录残余风险，本视角认可该决策；建议持续监控内容源受信度，必要时后续单独发起存量清洗变更（proposal 已声明）。 |

## 问题跟踪表（供下一轮校验）

| # | 问题 | 状态 |
|---|------|------|
| M1 | submit 过渡路径绕过净化 | 已解决（REQ-XSS-1/6 + D9，代码核验一致） |
| S1 | 白名单未与实际编辑器产出交叉核对 | 已解决（REQ-XSS-8 + D12） |
| S1-new | `tel:` scheme 未纳入白名单评估 | 待处理（建议并入 REQ-XSS-8 交叉核对） |

---
VERDICT: APPROVED
---
