# QA Report — community-hub-service

**验证时间**: 2026-08-12 11:20
**验证范围**: 工作树未提交改动 + 未跟踪文件（FIX 多视角评审修复：Get-by-ID 数据范围 + RPC 身份伪造/回环绑定 + T4.0-T4.8 数据权限消费方）

## 机械化检查结果 (harness-checks.sh — FRESH run)

| # | 检查项 | 结果 | 详情 |
|---|--------|------|------|
| 1 | go build | ✅ | exit 0，compilation succeeded（本报告独立复跑确认） |
| 2 | go vet | ✅ | exit 0，no issues（独立复跑确认） |
| 3 | go test | ✅ | 10 包 33 测试函数，0 fail（独立复跑 -count=1 确认） |
| 4 | Proto int64 jstype | ✅ | 0 violations（diff 无 proto 变更，skipped） |
| 5 | json:",string" | ✅ | 0 violations（AST verified） |
| 6 | 跨服务DB导入 | ✅ | no violations（diff 扫描 51 Go 文件） |
| 7 | 错误码格式 | ✅ | no magic numbers（全部命名常量/0） |
| 8 | 硬编码密钥 | ✅ | no secrets detected |

其余检查（9-16）全部 PASS：graph_freshness / claude_structural_data / proto_ts_align / api_stubs / response_wrap / bench_regression(SKIP) / api_smoke / memory_index。

**汇总：16 PASS / 0 FAIL / 0 WARN**

## 编译检查
- [x] go build ./... — exit 0

## 静态分析
- [x] go vet ./... — exit 0，clean output

## 单元测试
- [x] go test ./... -count=1 — 10/10 包通过，33 个测试函数，0 fail

测试包清单（FRESH -count=1）：
- api/internal/logic/{contact,lostfound,notice}、api/internal/svc、api/internal/util
- rpc/internal/config、rpc/internal/logic/{contact,lostfound,notice,scope}
- 无 `0/0` 假通过（所有包均有实际测试函数）

## 测试覆盖

| 包 | 覆盖率 | 状态 |
|----|--------|------|
| api/internal/util | 86.4% | 良好 |
| rpc/internal/logic/notice | 57.6% | 良好 |
| rpc/internal/logic/lostfound | 51.2% | 良好 |
| api/internal/logic/contact | 45.8% | 合理 |
| rpc/internal/logic/scope | 41.9% | 合理 |
| api/internal/svc | 40.0% | 合理 |
| api/internal/logic/lostfound | 38.6% | 合理 |
| api/internal/logic/notice | 33.3% | 合理 |
| rpc/internal/logic/contact | 25.5% | 偏低（写接口主体覆盖） |

## TDD 证据检查

| 新增/修改函数 | 是否有测试 | RED 确认（含 FAIL 输出摘录） | GREEN 确认 | 状态 |
|-------------|:---:|:---:|:---:|:---:|
| RPC GetNoticeLogic.GetNotice（scope 过滤） | ✅ TestGetNotice_FilterByScope×5 | ✅ CHANGELOG: `expected: 80006, actual: 0` + `Expected nil, but got: &communityv1.Notice{...CommunityId:200...}` | ✅ 全绿 | PASS |
| RPC GetLostFoundLogic.GetLostFound（scope 过滤） | ✅ TestGetLostFound_FilterByScope×5 | ✅ CHANGELOG 同 RED 模式 | ✅ 全绿 | PASS |
| API GetNoticeLogic.GetNotice（CallCtx+ToError） | ✅ TestGetNotice_InjectsIdentity/SurfacesScopeDenied | ✅ CHANGELOG: `Should be true / 出站 metadata 必须存在`、`An error is expected but got nil` | ✅ 全绿 | PASS |
| API GetLostFoundLogic.GetLostFound（CallCtx+ToError） | ✅ TestGetLostFound_InjectsIdentity/SurfacesScopeDenied | ✅ CHANGELOG 同 RED 模式 | ✅ 全绿 | PASS |
| rpc config ListenOn 回环绑定 | ✅ TestRpcConfig_BindsLoopback | ✅ CHANGELOG: `Should be true / RPC 必须绑定回环（127.0.0.1/localhost），当前 host="0.0.0.0"` | ✅ 全绿 | PASS |
| scope.FilterAllowed | ✅ TestFilterAllowed×6 | ✅ CHANGELOG RED 模式 | ✅ 全绿 | PASS |
| scope.AssertCommunityScope / CheckPublishScope / CheckSystemPublishScope | ✅ TestAssertCommunityScope + publishscope_test | ✅ CHANGELOG RED 模式 | ✅ 全绿 | PASS |
| ServiceContext.CallCtx | ✅ TestCallCtx×3 | ✅ 行为型断言 | ✅ 全绿 | PASS |
| util.JWTUserID / WithUserID | ✅ TestJWTUserID×8 / TestWithUserID×2 | ✅ CHANGELOG: `undefined: JWTUserID` | ✅ 全绿 | PASS |
| RPC ListNoticesLogic.ListNotices（scope 过滤） | ✅ TestListNotices_FilterByScope×4 | ✅ 行为型断言（不查库=空列表） | ✅ 全绿 | PASS |
| RPC ListLostFoundLogic.ListLostFound（scope 过滤） | ❌ 无测试 | ❌ 无 | ✅ 编译通过 | **FAIL** |
| RPC ListContactsLogic.ListContacts（scope 过滤） | ❌ 无测试 | ❌ 无 | ✅ 编译通过 | **FAIL** |
| API ListNoticesLogic.ListNotices（CallCtx） | ❌ 无测试 | ❌ 无 | ✅ 编译通过 | **FAIL** |
| API ListLostFoundLogic.ListLostFound（CallCtx） | ❌ 无测试 | ❌ 无 | ✅ 编译通过 | **FAIL** |
| API ListContactsLogic.ListContacts（CallCtx） | ❌ 无测试 | ❌ 无 | ✅ 编译通过 | **FAIL** |
| API UpdateNoticeLogic.UpdateNotice（CallCtx+ToError） | ❌ 无测试 | ❌ 无 | ✅ 编译通过 | **FAIL** |
| API DeleteNoticeLogic.DeleteNotice（CallCtx+ToError） | ❌ 无测试 | ❌ 无 | ✅ 编译通过 | **FAIL** |
| API ResolveLostFoundLogic.ResolveLostFound（CallCtx+ToError） | ❌ 无测试 | ❌ 无 | ✅ 编译通过 | **FAIL** |

**TDD 判定：8 个工作树新增/修改函数无对应测试 → FAIL（TDD 证据不足）**

核心安全函数（FIX 轮次：Get-by-ID 数据范围 + 回环绑定）全部有测试且 RED 摘录留档完整；缺口集中在 ListLostFound/ListContacts/List/Update/Delete/Resolve 的 API/RPC 包装逻辑——这些函数在工作树中被挂了 scope 过滤 / CallCtx / ToError，但无测试覆盖。ListNotices 有同构测试（listnotices_filter_test.go），说明 ListLostFound/ListContacts 完全可按同模式补齐。

## 测试质量评估
- 新增/修改函数: 26（FIX 轮 10 个 + T4.x 数据权限 16 个）
- 有测试: 18 / 缺失: 8（均为 API/RPC 列表/更新/删除/解决包装逻辑）
- 边界测试: ✅（scope 三态、fail-closed userID=0、系统身份回调、gRPC 传输错误、metadata 缺失 均覆盖）

## 发现
| 级别 | 问题 | 建议 |
|------|------|------|
| FAIL | 8 个函数修改后无对应测试：RPC ListLostFound/ListContacts 的 scope 过滤；API ListNotices/ListLostFound/ListContacts 的 CallCtx 注入；API UpdateNotice/DeleteNotice/ResolveLostFound 的 CallCtx+ToError | 按 listnotices_filter_test.go 同模式补 ListLostFound/ListContacts scope 过滤测试；按 getnoticelogic_test.go 同模式补 API 列表/更新/删除/解决测试（InjectsIdentity + SurfacesScopeDenied） |

---
VERDICT: FAIL — 机械化检查 16/16 全绿、build/vet/test 全通过，但 8 个工作树修改函数无 TDD 测试（证据不足）
---
