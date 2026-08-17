# QA Report — user-service

**验证时间**: 2026-08-17 06:10 (UTC)
**验证范围**: 当前工作树未提交改动 + 未跟踪文件 — joinCommunity 房号/权属改为可选（用户拍板）+ REST 边界同步放行最小载荷：新增 `joinResidenceProvided` + JoinCommunity 分支改造 + 6 个新逻辑测试 + 2 个新 REST 字段契约测试 + 3 个改用例 + types.go optional + CHANGELOG + go.mod

---

## 机械化检查结果 (harness-checks.sh — FRESH run)

> `bash .harness/skills/qa/scripts/harness-checks.sh --service user-service --json`
> 结果：**18 PASS / 0 FAIL / 3 WARN，exit_code=0**（timestamp 2026-08-17T05:58:15Z）

| # | 检查项 | 结果 | 详情 |
|---|--------|------|------|
| 1 | go build | ✅ | compilation succeeded (exit 0) |
| 2 | go vet | ✅ | no issues (exit 0) |
| 3 | go test | ✅ | 5 packages passed, ~157 test functions |
| 4 | Proto int64 jstype | ✅ | no proto changes in diff (skipped) |
| 5 | json:",string" | ✅ | all int64 ID fields have `json:",string"` (AST verified) |
| 6 | 跨服务DB导入 | ✅ | no violations (diff scan 5 Go files) |
| 7 | 错误码格式 | ✅ | no magic numbers found (all use named constants or 0) |
| 8 | 硬编码密钥 | ✅ | no secrets detected |
| 9 | 知识图谱新鲜度 | ✅ | graph up-to-date (synced 1h ago) |
| 10 | CLAUDE.md 结构化数据 | ✅ | no structural data duplication |
| 11 | Proto→TS 对齐 | ⚠️ WARNING | AUTO-MISMATCH: proto `LoginSmsRequest.phone`/`RefreshTokenRequest.device_type`/`SubmitReviewRequest.reviewer_id`/`RegisterRequest.phone`/`User.avatar_url` 未同步到 TS（identity.ts/moderation.d.ts）——**存量，与本 diff 无关**（前端类型未同步，待前端跟进） |
| 12 | API 逻辑 TODO stub | ✅ | no TODO stubs found in API logic |
| 13 | Response 单次包装 | ✅ | no double-wrap risk detected |
| 14 | Benchmark 回归 | ✅ | no benchmark functions — SKIP |
| 15 | API smoke 测试 | ✅ | no new routes detected in diff |
| 16 | Memory Index 新鲜度 | ✅ | 索引最新 (生成于 2026-08-17T05:54:35Z) |
| 16.5 | Design/代码一致性 | ⚠️ WARNING | user-service model 引用列未覆盖标准迁移源: `deleted_at;`——**存量 WARN，与本 diff 无关** |
| 17 | Git 卫生 | ⚠️ WARNING | gitlink 无 .gitmodules 条目: `api-proto;`——**存量 WARN，与本 diff 无关** |
| 18 | 变异测试 | ✅ | 变异分数 ?%（≥80% 或未解析到分数） |
| — | pipeline-evals | ✅ | 管线 eval 全部通过 |

**3 项 WARN 均为存量问题（TS 滞后 / deleted_at 列 / api-proto gitlink），非本次变更引入，不阻塞。**

---

## 编译检查
- [x] go build ./... — FRESH exit 0（`cd services/user-service && go build ./...`）

## 静态分析
- [x] go vet ./... — FRESH exit 0（clean output）

## 单元测试
- [x] go test ./... -count=1 — **5 包 / 157 测试函数（func Test 声明；含子测试共 221 条 RUN 线）/ 0 fail，exit 0**
  - api/internal/handler/community, api/internal/logic/community, api/internal/logic/user, api/internal/types, rpc/internal/logic/user

## 测试覆盖（不阻塞）
| 包 | 覆盖率 | 状态 |
|----|--------|------|
| rpc/internal/logic/user | 65.6% | 覆盖 joinResidenceProvided + JoinCommunity 分支 |
| api/internal/handler/community | 8.8% | 新增 join 字段契约用例覆盖 JoinCommunityHandler；其余 handler 未覆盖（存量） |
| api/internal/logic/user | 65.5% | — |

---

## TDD 证据检查（分诊：字段映射 vs 有逻辑函数）

| 新增/修改函数 | 类型 | 是否有测试 | RED 确认（仅「有逻辑」要求） | GREEN 确认 | 状态 |
|-------------|------|:---:|:---:|:---:|:---:|
| `joinResidenceProvided`（新纯函数：全有/全无/部分判定） | 有逻辑 | ✅ `TestJoinResidenceProvided`（9 组合表驱动） | ✅ 看到 FAIL: `join_community_optional_residence_test.go:45:34: undefined: joinResidenceProvided`（col 34 实测 = `joinResidenceProvided` 引用位置；符号不在 HEAD 已验证） | ✅ `go test -run TestJoinResidenceProvided` PASS | PASS |
| `JoinCommunity` 分支改造（hasResidence 条件化授权/地址/每户校验） | 有逻辑 | ✅ 6 新用例 + 3 改用例 | ✅ 看到 FAIL: `:68 expected: 0 / actual: 10040`（无房号应成功）、`:72 expected: 1 / actual: 0`（应建 membership）、`:99 expected: 0 / actual: 10040`（重新激活应成功）、`:122 expected: 10006 / actual: 10040`（应受上限约束）—— 逐行核对断言消息一致 | ✅ `go test -run 'TestJoinCommunity_(NoResidence|Partial)'` PASS | PASS |
| `types.go JoinCommunityReq` Building/Unit/Room `,optional` | 字段映射（接线） | ✅ `handler_test.go` 2 用例（最小载荷透传 + 带房号透传） | —（字段映射不要求 RED） | ✅ `go test ./api/internal/handler/community/` PASS | PASS |
| `go.mod` miniredis v2.38.0 indirect→direct | 字段映射（依赖接线） | —（构建级，测试文件 `get_user_roles_logic_test.go` 直接 import） | — | ✅ `go build ./...` exit 0 | PASS |

**RED 证据核验说明**：
- `joinResidenceProvided` 为**新增符号**，`git show HEAD:` 中不存在 → `undefined:` 编译 RED 真实（结构性证明 RED baseline）。
- CHANGELOG 摘录的 RED 文本（行号+断言消息）已与当前测试文件**逐行核对一致**：`join_community_optional_residence_test.go:68/72/99/122`、`handler_test.go:67` 的断言行与摘录的 expected/actual/消息完全吻合；`:45:34` 实测为 `joinResidenceProvided` 引用位置。RED 摘录真实可信。
- 修改的 `join_community_member_constraints_test.go`（MissingBuilding/Unit/Room → 部分提供语义 10040）保持绿。

---

## 测试质量评估
- 新增/修改函数: 4（`joinResidenceProvided` 逻辑 + `JoinCommunity` 分支 + types.go 字段契约 + go.mod 依赖）
- 有测试: 4 / 缺失: 0
- 边界测试: ✅
  - `joinResidenceProvided`：9 组合覆盖全有(自有/租住)/全无/5 种部分提供（仅权属/仅楼号/仅单元/仅房号/有房号无权属/缺房号）
  - 行为测试：无房号不自动授权（AssignRole 0 次调用）、重新激活清 0、仍受 10006 上限、部分提供 10040 不建 membership
  - REST 契约：最小载荷 `{"community_id":"123"}` 边界放行透传（building/unit/room=0 + ownership=UNSPECIFIED）+ 带房号载荷不破
- 边界缺口: 无（既有 ownership/上限/每户≤6/地址校验用例回归保持绿）

## 发现
| 级别 | 问题 | 建议 |
|------|------|------|
| WARNING | proto_ts_align：5 个 TS 字段滞后 proto（identity.ts/moderation.d.ts） | 存量，与本次变更无关；待前端同步 |
| WARNING | design_consistency：model 引用列 `deleted_at` 未覆盖标准迁移源 | 存量 |
| WARNING | git_hygiene：api-proto 子模块无 .gitmodules 条目 | 存量，补 .gitmodules 条目 |
| INFO | handler/community 包覆盖率 8.8% | 仅新增 handler 有测试，存量 handler 未覆盖（不阻塞本变更） |

---
VERDICT: **PASS**

证据（every fresh run）：
- `go build ./...` → EXIT_CODE=0
- `go vet ./...` → EXIT_CODE=0
- `go test ./... -count=1` → 5 包 ok / 0 fail，EXIT_CODE=0
- harness-checks.sh --json → 18 PASS / 0 FAIL / 3 WARN（存量），exit_code=0
- TDD：`joinResidenceProvided`/`JoinCommunity` 分支 RED 摘录（undefined + 4 组 expected/actual）经行号核对真实；字段映射（types.go optional + go.mod）不要求 RED
---
