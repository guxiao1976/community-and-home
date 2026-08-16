# Plan Review — notice-xss-sanitize-and-frontend-fixes（覆盖完整性视角）

**审查维度**: 需求覆盖 / 场景完整性 / 边界识别
**审查版本**: P1.4（磁盘最新；审查哈希 fallback:r1:rc1，独立重新审查，未沿用旧轮结论）
**审查时间**: 2026-08-16

## 摘要
- 🔴 MUST FIX: 0 / 🟡 SHOULD FIX: 3 / 🔵 INFO: 3

## 审查对象（磁盘）
- `request.md`（3 项用户需求 + 已拍板 Q1-Q6 决策 + 上轮 4 条 REVISION 反馈）
- `proposal.md`（D1-D14 决策日志）
- `specs/xss-sanitization/spec.md`、`specs/notice-double-load-fix/spec.md`、`specs/auth-toast-fix/spec.md`
- 交叉核对源码：createcontentpostlogic.go / updatecontentpostlogic.go / notice.vue / auth-flow.ts / community.ts（store）/ notice-detail.vue

## 覆盖完整性核验结论

### 需求决策点覆盖（request.md 全部决策点 → 已映射）
| 决策点 | spec 落点 | 结论 |
|------|----------|:---:|
| Q1=1 显式单次加载+守卫 | REQ-DBL-1 | ✅ |
| Q2=0 toast 合并 | REQ-TOAST-1 | ✅ |
| Q3=0 仅正文净化 | REQ-XSS-5 | ✅ |
| Q4=0 净化器本服务 | REQ-XSS-1 说明 | ✅ |
| Q5=0 不回填 | REQ-XSS-6 | ✅ |
| Q6=0 img 全剔除 | REQ-XSS-2 | ✅ |
| D7 净化/非空校验顺序 | REQ-XSS-1（080005 原始正文先行，净化后为空唯一化） | ✅ |
| D8 白名单穷举 | REQ-XSS-2（完整标签/属性集，a/div 去留） | ✅ |
| D9 submit 净化再发布 | REQ-XSS-1 入口③ + REQ-XSS-6 场景 | ✅ |
| D11 Update 未携带不重净化 | REQ-XSS-6 场景 | ✅ |
| D12 白名单交叉核对 | REQ-XSS-8 | ✅ |
| D14 toast icon:none 不承诺恢复 | REQ-TOAST-1/2 | ✅ |
| 上轮 4 条 MUST/SHOULD 反馈 | 全部对应修订项 | ✅ |

### 场景完整性（Requirement 粒度）
| Spec | REQ 数 | 正向 | 异常/边界 | 结论 |
|------|:---:|:---:|:---:|:---:|
| xss-sanitization | 8 | ≥1/REQ | ≥1/REQ（含空正文 080005、净化后为空、纯文本、并发、存量、submit、img 剔除等边界） | ✅ |
| notice-double-load-fix | 3 | ≥1/REQ | ≥1/REQ（服务端覆写、无小区、memberships 失败、切换失败 10015、加载失败留痕） | ✅ |
| auth-toast-fix | 2 | ≥1/REQ | ≥1/REQ（profile 成功/失败、有/无小区、memberships 异常） | ✅ |

### 与源码交叉核对（防止 spec 自洽但脱离代码）
- **写路径枚举完整**：`content_posts.text` 全仓仅 CreateContentPost（`InsertTx`，createcontentpostlogic.go:158）与 UpdateContentPost 编辑分支（`UpdateContentTx`，updatecontentpostlogic.go:215）两处写；REST 层为 RPC 代理（api/updatecontentpostlogic.go:42 透传），无绕过路径。✅
- **moderation-service 无回写**：consumer 已停 Redis 回调、跳过 notice 残留任务，不写 content_posts.text。✅
- **读路径唯一富文本渲染点** `notice-detail.vue:23 <rich-text :nodes="notice.content">`，已被写入路径净化覆盖；title 全文 `{{ }}` 插值（notice.vue:61 / notice-detail.vue:5），寻失/便民联络/PC 无 rich-text/v-html，D3「仅正文净化」安全。✅
- **loadMemberships 失败语义**：store `catch { communities.value = [] }`（community.ts:101-103），REQ-DBL-1 失败降级场景与代码一致；getAppState 覆写仅在 serverCurrentId ∈ memberships 时生效（community.ts:84）。✅

## 发现

### 🟡 SHOULD FIX

| # | 文件:行号/章节 | 问题 | 修复建议 |
|---|-------------|------|---------|
| 1 | notice-double-load-fix/spec.md REQ-DBL-1「未加入任何小区」+「loadMemberships 整体失败」场景 | **loading 标志释放机制未指定，与 guard 规则内部矛盾**。notice.vue `loading=ref(true)` 且仅 `loadAll()` 置 false；guard 规则「仅 hasCommunities==true 时显式 loadAll」字面执行 → 无小区/失败路径永不调用 loadAll → loading 恒 true → 空态引导（模板 `v-if="!hasCommunities && !loading"`）永不渲染、页面卡骨架屏。场景已断言「空态引导 + 不卡 loading」，但未说明如何达成（机制留白）。 | spec 明确：onMounted 在 membershipsResolved 置 true 后，无小区/失败路径显式置 `loading=false`（或空态渲染不依赖 loading），并补一条「无小区页面不卡骨架、空态正常渲染」的显式断言。 |
| 2 | notice-double-load-fix/spec.md REQ-DBL-1「服务端权威覆写」场景 | **watch 首载守卫的时序未钉死**。守卫「membershipsResolved 置 true 后 watch 放行」依赖 Vue watch flush 次序：loadMemberships 内覆写 C1→C2 触发的 watcher 若在成员就绪标志已置 true 之后才执行（如 flush:'post' 渲染后 flush），会与 onMounted 显式 loadAll 构成双载——正是本修复要消除的回归。默认 flush:'pre' 下按微任务次序可规避，但 spec 未钉死该次序，实现存在静默复发窗口。 | 在 REQ-DBL-1 钉死 watch 采用默认 pre flush（或同步判定），并补明「membershipsResolved 置 true 前，loadMemberships 期间所有 currentCommunityId 变更的 watcher 必须已消费/忽略」的次序保证。 |
| 3 | xss-sanitization/spec.md REQ-XSS-7 | **测试覆盖仅限净化器直接单测（注入/合法/幂等），未要求对三个写入口各自「调用净化器 + 落库净化后内容」的写路径集成断言**。若实现只把净化器接入 CreateContentPost，submit（D9 新写路径）与 Update 编辑分支的缺口可能静默复发（上轮 validity M1 恰是 submit 绕过问题）。 | REQ-XSS-7 增加写路径集成断言要求：至少 Create 与 submit 两条入口各配一条「含恶意 payload 写入 → DB/读回为净化后内容」的断言用例（submit 断言 draft 净化后置公开）。 |

### 🔵 INFO

| # | 建议 |
|---|------|
| 4 | REQ-XSS-1/REQ-XSS-6：Kafka push 携带净化后 text 未显式钉死。Create entry_status=1 与 submit 均在事务后 `Producer.Push(post)`（producer.go:145 `Text: post.Text`）；若净化在 post 构造/改写前完成则天然带净化后文本，建议补一句「推送内容为净化后文本」消除实现歧义。 |
| 5 | REQ-DBL-2：仅覆盖 10015 切换失败；非 10015（网络异常）分支（onCommunitySwitch catch → 通用 toast，currentCommunityId 不变）无场景断言「不触发加载、不 double-load」。可补一条边界。 |
| 6 | REQ-TOAST-1：合并 toast 的发出位置（catch 内 vs 替换 step4 success toast）相对 800ms 跳转窗口未钉死；行为契约（失败信息可见、不被覆盖）已断言，实现留白可接受，但建议明确「合并 toast 在 profile 失败分支即时展示」避免跳转时已消失。 |

## 与上轮（r1）对照
上轮 4 条 MUST/SHOULD（白名单穷举、净化/校验顺序、submit 缺口、白名单交叉核对）均已在本轮 spec 逐条闭环；本轮独立审查未发现新的 MUST FIX。

---
VERDICT: APPROVED
---
