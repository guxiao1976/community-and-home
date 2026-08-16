# QA Report — community-hub-service

**验证时间**: 2026-08-16 18:00 (CST)
**验证范围**: 工作树未提交改动 + 未跟踪文件（mobile-homepage-content-revamp Task 1.1-1.5：model 时间窗口 + RPC since_days 校验 + REST 透传 + migration 004/005 补表/索引）

## 机械化检查结果 (harness-checks.sh — FRESH run)

命令：`bash .harness/skills/qa/scripts/harness-checks.sh --service community-hub-service --json`（FRESH，timestamp 2026-08-16T09:42:28Z）

| # | 检查项 | 结果 | 详情 |
|---|--------|------|------|
| 1 | go build | ✅ | PASS — compilation succeeded |
| 2 | go vet | ✅ | PASS — no issues |
| 3 | go test | ✅ | PASS — 13 包通过, ~124 测试函数（工作树 go test ./... -count=1 实测 124 个 Test 函数，0 fail） |
| 4 | go_fmt | ✅ | 变更 Go 文件全部已格式化 |
| 5 | Proto int64 jstype | ✅ | diff 无 proto 变更 (skipped) |
| 6 | json:",string" | ✅ | AST 校验全过，所有 int64 ID 字段均带 `json:",string"` |
| 7 | 跨服务DB导入 | ✅ | no violations（8 Go 文件 diff 扫描） |
| 8 | 错误码格式 | ✅ | no magic numbers（全部命名常量或 0；RPC 080005 用 scope.CodeInvalidParam 命名常量） |
| 9 | 硬编码密钥 | ✅ | no secrets detected |
| 10 | graph_freshness | ✅ | graph up-to-date (synced 0h ago)——上轮 FAIL（api-proto 9c848cb 晚于图谱时间戳）已由 Owner 执行 graph-sync.sh 消除 |
| 11 | claude_structural_data | ✅ | 无结构数据重复 |
| 12 | proto_ts_align | ⚠️ | WARN — 跨服务 TS 滞后（identity.ts LoginSmsRequest.phone / RefreshTokenRequest.device_type / RegisterRequest.phone / User.avatar_url、moderation.d.ts SubmitReviewRequest.reviewer_id），均属其他服务/前端任务范围，本工作树未触及 |
| 13 | api_stubs | ✅ | 无 TODO 桩 |
| 14 | response_wrap | ✅ | 无双重包裹风险 |
| 15 | bench_regression | ✅ | 无 benchmark（SKIP） |
| 16 | api_smoke | ✅ | diff 无新增路由（透传基于既有 GET /api/community/notices） |
| 17 | memory_index / design_consistency | ✅/✅ | 索引最新 (生成于 2026-08-16T09:40:21Z)；model 列覆盖标准迁移源 |
| 18 | git_hygiene | ⚠️ | WARN — gitlink 无 .gitmodules 条目: api-proto（既有治理项，非本变更引入） |
| 19 | mutation_testing / pipeline_evals | ✅/✅ | 变异分数 ?%（≥80% 或未解析到分数）；管线 eval 全部通过 |

**汇总**: **19 PASS / 0 FAIL / 2 WARN，`exit_code=0`**

## 编译检查
- [x] go build ./... — **exit 0**（FRESH，BUILD_EXIT=0）

## 静态分析
- [x] go vet ./... — **exit 0**（FRESH，VET_EXIT=0，clean output）

## 单元测试
- [x] go test ./... -count=1 — **exit 0**（FRESH 禁缓存；13 包含测试，124 个测试函数，0 fail；rpc/api 主包 + handler/types/config 为薄层无测试）

## 测试覆盖（go test -cover，非阻塞记录）
| 包 | 覆盖率 | 状态 |
|----|--------|------|
| rpc/internal/logic/notice | 73.4% | 良 |
| api/internal/logic/notice | 63.8% | 良 |
| model | 60.9% | 中 |

## TDD 证据检查（分诊：字段映射 vs 有逻辑函数）

**分诊说明**：本次变更新增/修改函数共 8 个 + 2 个 migration（纯 DDL）。字段映射类（types.go SinceDays 字段、WithTimeWindow 构造器接线、ContentPostListOptionSince 自省访问器、migration DDL）只要求有测试/对齐，RED 不要求；「有逻辑函数」（buildWindowClause 条件分支、FindListByCommunity 变参+窗口谓词拼接、rpc ListContentPosts 校验分支、api ListContentPosts 错误上抛分支）要求 RED 摘录。

**RED 证据来源**：CHANGELOG 2026-08-16「移动端首页信息架构改造」TDD 段（L35-37）持久化了**具体 FAIL 输出文本**（含 `undefined:` / `too many arguments in call to` / `expected: 80005, actual: 0` / `unknown field ... in struct literal`），非仅口头描述；结构性佐证 `git show HEAD:` 确认所有新符号（WithTimeWindow / ContentPostListOptionSince / buildWindowClause / ContentPostListOption / SinceDays / GetSinceDays）在 HEAD 均 0 命中 → 测试文件在 HEAD 状态必然编译失败（RED 真实）。

| 新增/修改函数 | 类型 | 是否有测试 | RED 确认 | GREEN 确认 | 状态 |
|-------------|------|:---:|:---:|:---:|:---:|
| model.ContentPostListOption / contentPostListParams | 字段映射(类型) | ✅ 间接（ContentPostListOptionSince 测试） | —（不要求） | ✅ | PASS |
| model.WithTimeWindow | 字段映射(构造器接线，无分支) | ✅ TestContentPostListOption_WithTimeWindow | —（不要求；CHANGELOG 另有 `undefined: WithTimeWindow` 编译期摘录） | ✅ PASS | PASS |
| model.ContentPostListOptionSince | 字段映射(自省访问器) | ✅ TestContentPostListOption_WithTimeWindow | —（不要求；CHANGELOG 另有 `undefined: ContentPostListOptionSince` 摘录） | ✅ PASS | PASS |
| model.buildWindowClause | 有逻辑(条件分支 since==nil) | ✅ TestContentPostModel_FindListByCommunity_WithWindow（sqlmock 锁定 SQL 谓词） | ✅ `too many arguments in call to m.FindListByCommunity`（变参签名编译期 RED，CHANGELOG L35） | ✅ PASS（2 子用例） | PASS |
| model.FindListByCommunity（变参签名 + 窗口谓词） | 有逻辑(SQL 拼接分支) | ✅ TestContentPostModel_FindListByCommunity_WithWindow | ✅ 同上 `too many arguments in call to m.FindListByCommunity`（CHANGELOG L35） | ✅ PASS | PASS |
| rpc ListContentPosts（since_days 校验 + 窗口传参） | 有逻辑(校验分支 <0\|\|>365→080005) | ✅ TestListContentPosts_SinceDays×5 | ✅ `Should NOT be empty, but was []`（30 未传窗口）+ `expected: 80005, actual: 0`（-1/366 未拦截）（CHANGELOG L36） | ✅ 5 用例 PASS | PASS |
| api ListContentPosts（SinceDays 透传 + responsex.ToError 上抛） | 有逻辑(错误上抛分支) | ✅ TestListContentPosts_SinceDaysAndBaseError×3 | ✅ `unknown field SinceDays in struct literal`（编译期 RED，CHANGELOG L37） | ✅ 3 用例 PASS | PASS |
| api types.ListContentPostsReq.SinceDays | 字段映射 | ✅ 同上透传断言 | —（不要求） | ✅ | PASS |
| migration 004/005 | 字段映射(DDL) | —（纯 DDL，无 TDD 要求） | — | —（幂等守卫 + 001/design 对齐） | PASS |

### GREEN 复现（工作树，-count=1 禁缓存，FRESH）

```
$ go test ./model/ -run 'TestContentPostListOption_WithTimeWindow|TestContentPostModel_FindListByCommunity_WithWindow' -count=1 -v
  --- PASS: TestContentPostListOption_WithTimeWindow
  --- PASS: TestContentPostModel_FindListByCommunity_WithWindow (含 2 子用例)
  ok   model  0.010s  (exit 0)
$ go test ./rpc/internal/logic/notice/ -run TestListContentPosts_SinceDays -count=1 -v
  --- PASS: TestListContentPosts_SinceDays (5 子用例: 30/0/-1/366/365)
  ok   ...notice  0.028s  (exit 0)
$ go test ./api/internal/logic/notice/ -run TestListContentPosts_SinceDaysAndBaseError -count=1 -v
  --- PASS: TestListContentPosts_SinceDaysAndBaseError (3 子用例)
  ok   ...notice  0.027s  (exit 0)
$ go test ./... -count=1   # exit 0，13P/0F
```

## 测试质量评估
- 新增/修改函数: 8，有测试: 8，缺失: 0
- 边界测试: ✅（since_days 0/30/365/366/-1 边界矩阵；无窗口 additive 缺省；窗口谓词 count/list 双 SQL sqlmock 锁定；NULL/未来行排除由 SQL 谓词保证；REST 080005 上抛 + wire 键保持）

## 发现

| 级别 | 问题 | 建议 |
|------|------|------|
| ⚠️ WARN | proto_ts_align：跨服务 TS 滞后（identity.ts 5 字段 / moderation.d.ts reviewer_id），本工作树未触及这些文件 | 各归属服务/前端任务同步 |
| ⚠️ WARN | git_hygiene：api-proto gitlink 无 .gitmodules 条目 | 按 Git治理规范补 .gitmodules |
| ℹ️ 提示 | bench_regression：无 benchmark 函数 | 热路径（FindListByCommunity 窗口谓词）可考虑补 benchmark |

## TDD 纪律达成
- 本次 4 个「有逻辑函数」RED 全部有**具体 FAIL 输出摘录**（编译期 `undefined:` / `too many arguments in call to` / `unknown field SinceDays in struct literal` + 行为型 `Should NOT be empty, but was []` / `expected: 80005, actual: 0`），且 `git show HEAD:` 结构性佐证新符号在 HEAD 均不存在（RED 真实）。相较 `tdd-red-evidence-requires-fail-excerpt` 记忆前 5 次复发（Generator 仅口头描述），本轮 CHANGELOG 已持久化真实失败文本，未复发。
- 字段映射类（types.SinceDays、WithTimeWindow、ContentPostListOptionSince、migration DDL）RED 不要求，均有测试/对齐。
- GREEN 全绿：3 组定向测试 + 全量 `go test ./... -count=1`（124 测试函数 0 fail）。

---
VERDICT: PASS — 机械化检查 **19 PASS / 0 FAIL / 2 WARN（exit_code=0）**；go build/vet/test 全绿（exit 0，124 测试函数 0 fail）；TDD 证据完整（4 个有逻辑函数均有具体 RED FAIL 摘录 + GREEN 定向测试通过）。2 项 WARN（proto_ts_align / git_hygiene）均为跨服务/既有治理项，非本变更引入。
---
