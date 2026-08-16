# Plan Review — notice-xss-sanitize-and-frontend-fixes（结构合理性视角）

**审查维度**: 职责边界、一致性
**审查版本**: P1.4（磁盘最新，独立审查；本任务给定 fallback r1:rc1 与 v1 轮 r0:rc1 不同，spec 已更新至 P1.4，故按磁盘内容重新审查，不沿用旧轮结论）
**审查时间**: 2026-08-16

## 摘要
- 🔴 MUST FIX: 0 / 🟡 SHOULD FIX: 0 / 🔵 INFO: 3

## 审查对象
- `.harness/changes/notice-xss-sanitize-and-frontend-fixes/proposal.md`
- `specs/xss-sanitization/spec.md`（REQ-XSS-1..8）
- `specs/notice-double-load-fix/spec.md`（REQ-DBL-1..3）
- `specs/auth-toast-fix/spec.md`（REQ-TOAST-1..2）
- `.change.yaml`（revises / specs / out_of_scope / proto_change_required）
- 实代码核验（结构前提是否成立）

## 核验结论（结构前提成立，未发现架构违反）

1. **proposal 影响范围 ↔ specs 职责边界一致、capability 无重叠**：3 个 capability 各归一方——xss-sanitization→community-hub-service（写入路径净化）、notice-double-load-fix→web/mobile notice.vue、auth-toast-fix→web/mobile auth-flow.ts；文件零重叠，change.yaml revises/specs/out_of_scope 与 proposal/spec 一致。
2. **「REST 为 RPC 代理，RPC 层单点覆盖」前提成立（重验）**：`api/internal/logic/notice/createcontentpostlogic.go:43`、`updatecontentpostlogic.go:68` 均调 `ContentPostServiceRpc.*`，REST 不直写 DB。
3. **content_posts.text 写路径枚举完整（重验，REQ-XSS-1 入口 1/2/3 与代码一一对应）**：
   - CreateContentPost → `InsertTx`（createcontentpostlogic.go:158）；
   - UpdateContentPost 内容编辑分支（text 携带）→ `UpdateContentTx`（updatecontentpostlogic.go:215）；
   - UpdateContentPost submit 分支 → `UpdateStatusAndPublishTx`（updatecontentpostlogic.go:137，当前仅置状态/时间，D9 需在同一事务追加净化正文写入——spec REQ-XSS-1 入口 3 + REQ-XSS-6 场景已明确捕获，`post.Text` 已在函数顶部 FindOne 载入，结构上可实现）。
   - 全仓 `ContentPostModel` 写方法调用核验：无 admin/后台直写、无 job/mq 直写（community-hub-service 无 job/mq 目录）；moderation-service 已移除 notice 写回（task_handler.go D4/D21 精确跳过），无 content_posts 写；`UpdateIsPinned`/`UpdateAttachmentCountTx`/`Withdraw`/`UpdateKafkaPushStatus` 均不写 text 列。枚举完整。
4. **前端结构前提成立（重验）**：`notice.vue` `watch(currentCommunityId)→loadAll()` + `onMounted{ await loadMemberships(); loadAll() }` 并存；`stores/community.ts loadMemberships` 内 `getAppState` 服务端权威覆写 `currentCommunityId`（:86），且外 try/catch 整体失败置 `communities=[]` 不抛错（:100-102，支撑 REQ-DBL-1 降级场景）；`auth-flow.ts` handleAuthSuccess 中「获取用户资料失败」toast（:36）随后被「登录成功」success toast（:52）覆盖——REQ-TOAST-1 前提成立。
5. **移动端富文本渲染面唯一**：web/mobile 仅 `notice-detail/notice-detail.vue:23` 一处 `<rich-text :nodes="notice.content">`，无其他 v-html/rich-text；寻失 description 无富文本渲染。D3 范围（仅公告/内容帖正文）与真实 XSS 面一致，out_of_scope（寻失 description 等）不构成隐藏 XSS 面。

## 上轮结构问题修复验证（v1 → P1.4）

| v1 问题 | P1.4 修复状态 | 验证 |
|---------|:---:|------|
| v1 SHOULD FIX 1（REQ-XSS-1「every write path」绝对保证绑定逻辑层而非 DB 边界层） | 已解决（选方案 b） | REQ-XSS-1 显式穷举 3 入口 + 新增前向 SHALL「Any future write path…SHALL also pass through」；入口与实代码逐一核对完整。逻辑层放置为已声明边界决策（D4/D5），剩余风险文本自明 |
| v1 SHOULD FIX 2（REQ-XSS-4 绝对保证与 REQ-XSS-6 存量不回填冲突） | 已解决 | REQ-XSS-4 限定「newly written / updated bodies…after this change goes live」，存量残余风险明确指向 REQ-XSS-6（D5/D10） |
| v1 INFO-2（守卫解除时机不可判定） | 已解决 | REQ-DBL-1 以模块级布尔 `membershipsResolved` 固化守卫条件（loadMemberships 结束 finally 等价路径置 true） |
| v1 INFO-1（sanitize.go 放 notice 包） | 保留（可接受） | 当前范围仅公告/内容帖正文；若未来跨内容类型复用需移出 notice 包，见 INFO-1 |

## 发现

### 🔴 MUST FIX
无。

### 🟡 SHOULD FIX
无。

### 🔵 INFO
| # | 建议 |
|---|------|
| 1 | `sanitize.go` 位于 `rpc/internal/logic/notice/`，白名单策略与 notice capability 耦合；REQ-XSS-1 的前向 SHALL 依赖未来开发者自觉（逻辑层净化，直写 model 的路径不强制拦截）。已选方案 b 并核对入口完整，当前范围可接受；若未来出现多内容类型共用 content_posts.text，建议将净化器下沉到持久化边界或抽公共包，届时重新评估。 |
| 2 | REQ-XSS-2 规则「a.target 仅当同步强制 rel="noopener noreferrer" 时保留」与其正向 Scenario「合法链接保留（含 rel/target 完整保留）」在输入 rel="noopener"（非 "noopener noreferrer"）时存在字节级不一致：按规则需强制改写 rel 后才保留 target，输出 rel 值会变为 "noopener noreferrer"。属 clarity/validity 视角范畴，请该视角评审确认实现层可达成（bluemonday 声明式策略对「依另一属性条件保留属性」支持有限，可能需要自定义链接重写）。 |
| 3 | D9（submit 对既有 draft 正文净化）需在 `UpdateStatusAndPublishTx` 所在事务内追加 text 列写入，当前 model 无「text+status 原子更新」方法，需新增或改造 model 方法并在 `sanitize.go`/`updatecontentpostlogic.go` 覆盖单测（spec 已要求 submit 单测覆盖，实现时勿遗漏）。 |

---

## 问题跟踪表（上轮 REVISION 结构项 → 本轮）

| 状态 | 问题 | 备注 |
|------|------|------|
| 已修复 | v1 SHOULD FIX 1（净化范围绑定逻辑层） | P1.4 方案 b + 入口穷举，代码核对完整 |
| 已修复 | v1 SHOULD FIX 2（REQ-XSS-4 vs REQ-XSS-6 语义冲突） | D10 收敛 |
| 已修复 | v1 INFO-2（守卫解除时机） | `membershipsResolved` 固化 |
| 已修复 | 上轮 request.md 缺失（REVISION 根因） | request.md 本轮补齐 |

---
VERDICT: APPROVED
---
