# Plan Review — notice-xss-sanitize-and-frontend-fixes（结构合理性视角）

**审查维度**: 职责边界、一致性
**审查版本**: P1.3 fallback:r0:rc1（磁盘最新内容，独立审查）
**审查时间**: 2026-08-16

## 摘要
- 🔴 MUST FIX: 0 / 🟡 SHOULD FIX: 2 / 🔵 INFO: 2

## 审查对象
- `.harness/changes/notice-xss-sanitize-and-frontend-fixes/proposal.md`
- `specs/xss-sanitization/spec.md`（REQ-XSS-1..7）
- `specs/notice-double-load-fix/spec.md`（REQ-DBL-1..3）
- `specs/auth-toast-fix/spec.md`（REQ-TOAST-1..2）
- 实代码核验（结构性前提是否成立）：community-hub-service create/update 写入路径、RPC 方法全集、渲染端、moderation 写回、web/pc 渲染

## 核验结论（结构前提成立，未发现架构违反）

1. **proposal 影响范围 ↔ specs 职责边界一致**：3 个 capability 各自清晰、无重叠 —— xss-sanitization→community-hub-service（写入路径）、notice-double-load-fix→web/mobile notice.vue、auth-toast-fix→web/mobile auth-flow.ts。文件零重叠。change.yaml revises 与 proposal/spec 一致。
2. **「REST 为 RPC 代理，RPC 层单点覆盖两入口」前提成立**：实代码 `api/internal/logic/notice/createcontentpostlogic.go:43`、`updatecontentpostlogic.go:68` 均调用 `ContentPostServiceRpc.Create/UpdateContentPost`，REST 不直写 DB。RPC 层单点净化可同时覆盖 REST 与 RPC。
3. **「every write path that persists the body」在当前代码成立**：content_posts 正文 text 仅两处落库 —— RPC Create `createcontentpostlogic.go:158` InsertTx、RPC Update `updatecontentpostlogic.go:215` UpdateContentTx（`model/content_post.go` 的 UpdateContent 唯一调用方）。RPC 方法全集只有 Create/Update 写正文；moderation-service 已移除 notice 写回（`task_handler.go` D4/D21 精确跳过，无 content_posts 写）；其余服务无 content_posts 写入；未发现 admin/管理端写路径。净化器挂在两个 RPC logic 文件即覆盖全部现有写入入口。
4. **读路径安全主张（REQ-XSS-4）成立**：移动端详情 `web/mobile/src/pages/notice-detail/notice-detail.vue:23` 用 `<rich-text :nodes="notice.content">` 原样渲染；`web/pc/src` grep `rich-text|v-html` 0 命中。写入路径净化后读路径天然安全，无需逐渲染端处理。

## 发现

### 🟡 SHOULD FIX
| # | 文件:行号/章节 | 问题 | 建议 |
|---|-------------|------|------|
| 1 | `specs/xss-sanitization/spec.md` REQ-XSS-1 | REQ-XSS-1 的 SHALL 语义是「**every** write path that persists the body… before the value is written to the database」绝对保证，但实施边界（proposal + change.yaml）把净化器放在 **RPC logic 层**（2 个文件），而非持久化/模型层（ContentPostModel Insert/UpdateContent）。当前等价（仅 2 条写入路径），但规范保证绑定在逻辑层而非 DB 边界层：未来新增写入路径（新 RPC 方法、管理/后台接口、数据导入/回填任务、moderation 自动改写）直写 model 即静默绕过净化，届时 REQ-XSS-1 已违反而实现者无感知，削弱「治本一处修复」的结构承诺。 | 二选一：(a) 将净化器下沉到持久化边界（ContentPostModel 的 insert/updateContent 内），使「every write path」在 DB 边界层天然成立；(b) 收紧 REQ-XSS-1 范围措辞，显式枚举入口（CreateContentPost/UpdateContentPost RPC + REST 代理）并注明「未来新增写入路径必须经净化器」。 |
| 2 | `specs/xss-sanitization/spec.md` REQ-XSS-4 vs REQ-XSS-6 | REQ-XSS-4 无前提地声明「all read paths and all renderers serve sanitized content without further per-render processing」，与 REQ-XSS-6（存量不回填，存量恶意 HTML 保持原样）语义冲突：对存量行，读路径返回的仍是非净化内容，REQ-XSS-4 的绝对保证不成立。会让实现者误判「所有行都已安全」而误引入回填/读路径净化等越界动作。 | 将 REQ-XSS-4 限定为「**newly written / updated** bodies」，存量行的残余风险明确指向 REQ-XSS-6，两需求范围对齐。 |

### 🔵 INFO
| # | 建议 |
|---|------|
| 1 | `sanitize.go` 位于 `rpc/internal/logic/notice/`，净化策略（白名单）与 notice capability 耦合。content_posts.text 是共享列，若未来对其它内容类型复用净化，需越过 notice 包。若净化器预期通用，可考虑更中性位置；当前范围（仅公告/内容帖正文）下可接受。 |
| 2 | REQ-DBL-1 行为定义正确，但「onMounted 显式一次 + watch 首载守卫」的守卫解除时机（须在显式初始 loadAll 完成后才放行 watch，避免守卫与显式加载同一 tick 竞态）未在 spec 固定为具体条件，建议 clarity 视角补一个可判定的守卫条件（如 membershipsReady 标志）供实现对齐。 |

---

VERDICT: APPROVED
---
