# QA Report — master-data-service

**验证时间**: 2026-08-12 17:20 CST
**验证范围**: 当前工作树未提交改动 + 未跟踪文件（T2.2 整树缓存跨进程失效 + 本轮多视角评审修复：REST 路由 JWT 认证 + 审批/恢复拓扑缓存失效接线；Generator 直接改主树未提交，管线 QA 按工作树 diff 校验）
**验证人**: QA Engineer Agent (Go)

---

## 机械化检查结果 (harness-checks.sh — FRESH run)

> `bash .harness/skills/qa/scripts/harness-checks.sh --service master-data-service --json` → exit 0, **16/16 PASS, 0 FAIL, 0 WARN**

| # | 检查项 | 结果 | 详情 |
|---|--------|------|------|
| 1 | go build | ✅ | PASS — compilation succeeded (exit 0) |
| 2 | go vet | ✅ | PASS — no issues |
| 3 | go test | ✅ | PASS — 7 packages passed, ~38 test functions |
| 4 | Proto int64 jstype | ✅ | PASS — no proto changes in diff (skipped) |
| 5 | json:",string" | ✅ | PASS — all int64 ID fields have `json:",string"` (AST verified) |
| 6 | 跨服务DB导入 | ✅ | PASS — no Go changes in diff (skipped) |
| 7 | 错误码格式 | ✅ | PASS — no magic numbers found (all use named constants or 0) |
| 8 | 硬编码密钥 | ✅ | PASS — no secrets detected |
| 9 | Knowledge graph freshness | ✅ | PASS — graph up-to-date (synced 0h ago) |
| 10 | CLAUDE.md structural data | ✅ | PASS — no structural data duplication |
| 11 | Proto→TS alignment | ✅ | PASS — all proto fields match TS interfaces |
| 12 | API logic TODO stubs | ✅ | PASS — no TODO stubs found in API logic |
| 13 | Response single-wrap | ✅ | PASS — no double-wrap risk detected |
| 14 | Benchmark regression | ✅ | PASS — no benchmark functions (SKIP; 建议为热路径补 benchmark) |
| 15 | API smoke test | ✅ | PASS — no new routes detected in diff |
| 16 | Memory index freshness | ✅ | PASS — 索引最新 (2026-08-12T08:42:15Z) |

> 注：check #6/#15 在子模块内以"no Go changes in diff (skipped)"跳过（harness-checks 在父仓 diff 上只见子模块指针一行）；已在子模块内 FRESH 复跑 go build/vet/test 独立确认。

## 编译检查 (FRESH)
- [x] go build ./... → **exit 0**（空输出，无错误；在子模块内独立复跑确认）

## 静态分析 (FRESH)
- [x] go vet ./... → **exit 0**（clean output；在子模块内独立复跑确认）

## 单元测试 (FRESH, 禁用缓存)
- [x] go test ./... -count=1 → **exit 0，7 包 38 测试函数全过，0 FAIL**
  - ok `api/internal/handler` — 2（routes_test.go：JWT 写路由 401 ×2）
  - ok `api/internal/logic/approval` — 7（reviewItemLogic 失效/不失效/JWT reviewer_id）
  - ok `api/internal/logic/deleted_items` — 5（restore 失效/不失效/失败路径）
  - ok `api/internal/svc` — 2（InvalidateScopeAncestorCache ×2）
  - ok `api/internal/util` — 1（ExtractUserID 表驱动）
  - ok `rpc/internal/logic/scoperesolve` — 11（Resolve×8 + Invalidate×3）
  - ok `rpc/internal/svc` — 10（bumpVersion×2 / currentVersion×4 / loadSnapshot×3 / CrossProcess×1）

## 测试覆盖（aggregate -coverpkg，go tool cover -func FRESH）

| 函数 | 覆盖 | 状态 |
|------|------|------|
| api handler.RegisterHandlers（本轮 JWT 接线） | 100.0% | ✅ |
| api logic approval.getReviewerId（本轮 JWT user_id 归属） | 100.0% | ✅ |
| api logic approval.reviewResidentialArea（本轮失效接线） | 77.8% | ✅ |
| api logic approval.reviewDivision（本轮失效接线） | 59.4% | ✅ |
| api logic deleted_items.RestoreDeletedItem（本轮失效接线） | 92.9% | ✅ |
| api svc.InvalidateScopeAncestorCache | 100.0% | ✅ |
| api util.ExtractUserID | 100.0% | ✅ |
| ScopeAncestorCache.Invalidate | 100.0% | ✅ |
| ScopeAncestorCache.bumpVersion | 100.0% | ✅ |
| ScopeAncestorCache.currentVersion | 92.9% | ✅ |
| ScopeAncestorCache.SetVersionReaderForTest | 100.0% | ✅ |
| ScopeAncestorCache.SetSnapshot / snapshot | 100.0% / 100.0% | ✅ |
| ScopeAncestorCache.loadSnapshot | 93.9% | ✅ |
| ScopeAncestorCache.ResolveAncestors | 100.0% | ✅ |

## TDD 证据检查

> 范围仅含**本次工作树新增/修改函数**（不含 git log 历史 commit，避免审到旧提交）。
> RED 证据来源：`_tdd_evidence.md`（T2.2 编译失败摘录）+ CHANGELOG 本轮（JWT/审批/恢复）断言失败摘录。

| 新增/修改函数 | 是否有测试 | RED 确认（含 FAIL 输出摘录） | GREEN 确认 | 状态 |
|-------------|:---:|:---:|:---:|:---:|
| api ServiceContext.InvalidateScopeAncestorCache（新增） | ✅ | FAIL: `serviceContext_test.go:21:7: ctx.InvalidateScopeAncestorCache undefined (type *ServiceContext has no field or method InvalidateScopeAncestorCache)` ✅ | 2 用例 PASS ✅ | PASS |
| ScopeAncestorCache.Invalidate(ctx)（签名变更+INCR） | ✅ | FAIL: `invalidate_test.go:33:39: too many arguments in call to svcCtx.ScopeAncestorCache.Invalidate / have (context.Context) / want ()` ✅ | 3 用例 PASS ✅ | PASS |
| ScopeAncestorCache.bumpVersion / currentVersion / SetVersionReaderForTest（新增） | ✅ | FAIL: `scopeancestorcache_test.go:113:4: c.bumpVersion undefined` / `:139:13: c.currentVersion undefined` / `invalidate_test.go:72:28: SetVersionReaderForTest undefined` ✅ | 全用例 PASS ✅ | PASS |
| ScopeAncestorCache.snapshot（generation 漂移）/ loadSnapshot（nil-model 守卫） | ✅ | FAIL: `undefined: model.ScopeAncestorVersionKey`（T2.1 基线无此常量，测试无法编译）✅ | 全用例 PASS ✅ | PASS |
| model.ScopeAncestorVersionKey（新增常量） | ✅(间接) | FAIL: `undefined: model.ScopeAncestorVersionKey` ✅ | 经 bump/currentVersion 用例引用 PASS ✅ | PASS |
| api handler.RegisterHandlers 全路由挂 rest.WithJwt（本轮） | ✅ routes_test.go 2 用例 | FAIL: `expected: 401, actual: 200/400/EOF`（无 JWT 时请求直达 handler，不返回 401）✅ | 7 写路由+无效 token 全 401 ✅ | PASS |
| api logic approval.getReviewerId → util.ExtractUserID（本轮） | ✅ TestReviewItem_ReviewerIdFromJWTUserIDClaim | FAIL: 有描述（"reviewer_id 由 7 变 0"）但 reviewer_id 专属断言摘录未单独持久化 ⚠️ | 987654 断言 PASS ✅ | PASS(⚠️) |
| api logic approval.reviewResidentialArea / reviewDivision 失效接线（本轮） | ✅ reviewItemLogic_test.go 7 用例 | FAIL: `expected: "1", actual: ""`（落库后 generation key 应为 1，修复前为 ""）✅ | approve 删小区/reject 硬删/approve 编辑/reject 回滚 全失效 ✅ | PASS |
| api logic deleted_items.RestoreDeletedItem 失效接线（本轮） | ✅ restoreDeletedItemLogic_test.go 5 用例 | ❌ **仅写 "RED → GREEN"（CHANGELOG:29），无实际 FAIL 输出摘录**（行为型断言失败，非编译错误，未捕获文本） | 5 用例 PASS ✅ | **FAIL** |
| division/residential_area API+RPC 创建/更新/删除 接线调用点（本轮，各 +1 行） | ⚠️ | 非新增函数，为一行为委托调用；失效语义由被调方法级测试覆盖 | 方法级 GREEN ✅ | PASS（委托调用，语义已测） |

- 结论：**restoreDeletedItemLogic_test.go 的 RED 列缺具体 FAIL 输出摘录（仅文字"RED → GREEN"）→ 视为 ❌ → TDD 证据不足 → 判定 QA FAIL**
- 佐证（只读复现检查法）：`git show HEAD:api/internal/logic/deleted_items/restoreDeletedItemLogic.go` 显示 HEAD 基线 RestoreDeletedItem 无 `InvalidateScopeAncestorCache` 调用 → 恢复后 generation key 不 INCR → `require.True(t, ok)` 在修复前必失败（RED 真实存在），但失败文本从未持久化。

## 发现

| 级别 | 问题 | 建议 |
|------|------|------|
| **FAIL** | `restoreDeletedItemLogic_test.go` RED 证据缺实际 FAIL 摘录（CHANGELOG:29 仅写 "RED → GREEN"，`_tdd_evidence.md` 未覆盖本轮 3 个新测试文件） | 按 `_tdd_evidence.md` 同模式持久化 restore 的 RED 断言失败文本（如 `restoreDeletedItemLogic_test.go:87: Error: Should be true / Test: TestRestoreDeletedItem_ResidentialArea_InvalidatesScopeAncestorCache`）；顺带补 getReviewerId 专属摘录 |
| WARNING | `getReviewerId` RED 专属断言摘录（`expected: 987654, actual: 0`）未单独持久化（CHANGELOG 只引用 generation 摘录） | 与 restore 一并补录 |
| WARNING | `rpc/internal/svc/scopeancestorcache_test.go:177` 尾随注释对齐不符 gofmt（上轮已记录，纯空白） | 随手 `gofmt -w`（非阻塞） |
| WARNING | 既有文件 import 顺序不洁（serviceContext.go、createdivisionlogic.go 等），经 `git show HEAD:` 核对均为变更行之外，非本轮引入 | 后续清理（非阻塞） |
| WARNING | division/residential_area create/update/delete 及 rpc division 接线调用点无 handler 级直接测试（相关 logic 包无测试文件），行为仅经方法级测试间接覆盖 | 可选：补最小接线测试；当前委托语义已被测方法覆盖，风险低 |
| WARNING | bench_regression 无 benchmark 函数 | 为热路径（ResolveScopeAncestors / 缓存快照）补 benchmark（非阻塞） |
| INFO | 多数包无测试文件（预存现状，非本次变更） | 不在本次范围 |

## 记忆应用

- 命中 **[[tdd-red-evidence-requires-fail-excerpt]]**（severity: must-follow, apply_count: 0）：本轮为首次实际应用——Generator 已对 routes/reviewItem 补摘录，但 restore 仍未补 → 证明该经验尚未根治。QA 已按 FAIL 流程更新该记忆（追加 restore 复现场景：行为型断言 RED 也需持久化失败文本；`_tdd_evidence.md` 须覆盖全部新增测试文件，不能只覆盖编译错误类）。

---

VERDICT: FAIL — 机械化检查 16/16 PASS、`go build`/`go vet`/`go test`（7 包 38 函数，FRESH）全绿、覆盖达标，**但 TDD 证据硬性门禁 FAIL**：restoreDeletedItemLogic_test.go 的 RED 列缺实际 FAIL 输出摘录（仅文字 "RED → GREEN"），按规则「RED 列缺少具体 FAIL 输出摘录 → ❌ → 判定 QA FAIL」。请 Generator 按 `_tdd_evidence.md` 模式补录 restore（及 getReviewerId）的真实失败文本后重跑。
---
