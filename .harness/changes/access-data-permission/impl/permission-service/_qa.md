# QA Report — permission-service

**验证时间**: 2026-08-12 09:40 (UTC)
**Operator**: QA Engineer Agent (Go服务)
**验证范围**: 当前工作树未提交改动 + 未跟踪文件（access-data-permission 阶段① Wave1 T1.1-T1.8 + need_human FindByRoleId assign_time 修复）

## 机械化检查结果 (harness-checks.sh — FRESH run)

| # | 检查项 | 结果 | 详情 |
|---|--------|------|------|
| 1 | go build | ✅ | exit 0, compilation succeeded |
| 2 | go vet | ✅ | exit 0, no issues |
| 3 | go test | ✅ | 4 包有测试, 69 测试函数, 0 FAIL（无 new-package 缺口 WARN） |
| 4 | Proto int64 jstype | ✅ | 本次 diff 无 proto 变更（跳过） |
| 5 | json:",string" | ✅ | AST 校验：所有 int64 ID 字段带 `json:",string"`，0 违规 |
| 6 | 跨服务DB导入 | ✅ | 0 violations（diff 扫描 29 个 Go 文件） |
| 7 | 错误码格式 | ✅ | 无 magic number（060006 走 responsex 命名常量） |
| 8 | 硬编码密钥 | ✅ | no secrets detected |
| 9 | graph_freshness | ✅ | graph up-to-date（synced 1h ago） |
| 10 | claude_structural_data | ✅ | 无 structural data duplication |
| 11 | proto_ts_align | ✅ | 全部 proto 字段与 TS 对齐 |
| 12 | api_stubs | ✅ | API logic 无 TODO stub |
| 13 | response_wrap | ✅ | 无 double-wrap |
| 14 | bench_regression | ✅ | 无 benchmark（SKIP，tip: 热点路径建议补） |
| 15 | api_smoke | ✅ | 无新 REST 路由（diff 内） |
| 16 | memory_index | ✅ | 索引最新 (2026-08-12T09:16:10Z) |

**汇总**: 16 PASS / 0 FAIL / 0 WARN — exit_code 0

## 编译检查
- [x] go build ./... — BUILD_EXIT=0（FRESH，独立运行确认）

## 静态分析
- [x] go vet ./... — VET_EXIT=0，clean output（FRESH，独立运行确认）

## 单元测试
- [x] go test ./... -count=1 — TEST_EXIT=0，0 失败（FRESH，禁用缓存）
  - 有测试的包：`api/internal/logic/perm`（4 函数）、`api/internal/types`（2）、`model`（29）、`rpc/internal/logic/permission`（34）
  - 测试函数总数 = 69（4 包全 ok）

## 测试覆盖 (go test -cover, FRESH)
| 包 | 覆盖率 | 状态 |
|----|--------|------|
| model | 69.2% | ✅ |
| rpc/internal/logic/permission | 61.1% | ✅ |
| api/internal/types | 31.2% | ✅ |
| api/internal/logic/perm | 7.4% | ✅（从 0% 有测试，helpers 全覆盖） |

## 测试质量评估
- 新增函数: 核心新逻辑均有表驱动测试 + 真实 assert 断言（能力分层 expected/actual、缓存 HIT 不触 DB、MISS 写缓存带 TTL、幂等 AssertNumberOfCalls、sqlmock SQL 正则不含 assign_time）。
- 缺口: `grantSatisfiedLevel`/`permissionDefMinLevel`/`userMaxLevel`/`scopeState*` 无独立单测，但经 CapabilityLayering / GetDataScopes cache 用例间接行为覆盖，非阻塞。

## TDD 证据检查（本次工作树新增/修改函数 — RED 摘录见 CHANGELOG 2026-08-12 三节）

| 新增/修改函数 | 是否有测试 | RED 确认（含 FAIL 输出摘录） | GREEN 确认 | 状态 |
|-------------|:---:|:---:|:---:|:---:|
| `resolveUserScope` (scope.go) | ✅ TestResolveUserScope (7 表驱动) | ✅ `undefined: resolveUserScope`（getdatascopeslogic.go:48:16） | ✅ go test PASS | **PASS** |
| `AssertPublishScope` (assertpublishscopelogic.go) | ✅ TestAssertPublishScope (6 表驱动) | ✅ `no field or method AssertPublishScope` | ✅ PASS | **PASS** |
| `invalidateUserCaches` (invalidate_caches.go) | ✅ TestInvalidateUserCaches_ScanDelete | ✅ `undefined: invalidateUserCaches`（assignrolelogic.go:66:2） | ✅ PASS | **PASS** |
| `FindActiveRolesByUserId` (model/rel.go) | ✅ 2 测试（+ExpiredExcluded） | ✅ `no field or method FindActiveRolesByUserId`（rel_test.go:271:19） | ✅ PASS | **PASS** |
| `FindScopesByUserId` 三态过滤 (rel.go) | ✅ 2 测试 | ✅ `could not match actual sql ... "ur.scope_id != 0"` | ✅ PASS | **PASS** |
| `CheckPermission` 能力分层 (checkpermissionlogic.go) | ✅ TestCapabilityLayering 等 4 | ✅ `expected: true, actual: false`（未认证业主发布） | ✅ PASS | **PASS** |
| `GetDataScopes` 三态 (getdatascopeslogic.go) | ✅ 5 测试 | ✅ `expected: 2, actual: 1` | ✅ PASS | **PASS** |
| `AssignRole` 幂等 (assignrolelogic.go) | ✅ TestAssignRole_Idempotent | ✅ `mock: unexpected Insert ...`（幂等依赖 InsertIgnore） | ✅ PASS | **PASS** |
| `RevokeRole` 缓存失效 (revokerolelogic.go) | ✅ 2 测试 | ✅ `Should be empty, but was cached_permissions` | ✅ PASS | **PASS** |
| `grantSatisfiedLevel` (helpers.go) | ✅ 经 CapabilityLayering 间接 | ✅ `undefined: grantSatisfiedLevel` | ✅ PASS | **PASS** |
| `permissionDefMinLevel`/`userMaxLevel` | ✅ 经 CheckPermission 间接 | ✅ `no field or method permissionDefMinLevel / userMaxLevel` | ✅ PASS | **PASS** |
| `scopeCacheData`/`scopeStateString/FromString` | ✅ 经 GetDataScopes cache 间接 | ✅ `undefined: scopeCacheData / scopeStateString / scopeStateFromString` | ✅ PASS | **PASS** |
| `FindByPath` (model/permission.go) | ✅ 3 测试 (Hit/Miss/StatusFiltered) | ✅ `could not match actual sql ... "where path = ? and status = 1"` | ✅ PASS | **PASS** |
| `InsertIgnore` (model/rel.go) | ✅ TestInsertIgnore_Idempotent | ✅ `could not match actual sql: "insert ignore into ..." with expected regexp "insert into ..."` | ✅ PASS | **PASS** |
| `UpdateUserRoleStatus` (updateuserrolestatuslogic.go) | ✅ 2 测试（缓存失效+err 传播） | ✅ `Received unexpected error: redis: nil` | ✅ PASS | **PASS** |
| `GetUserPermissions` (getuserpermissionslogic.go) | ✅ 3 测试 | ✅ `should not contain "community:lostfound:create-api"` | ✅ PASS | **PASS** |
| `toPermissionInfo`/`toRoleInfo`/`toPermissionInfoList` (api/perm/helpers.go) | ✅ helpers_test.go (4 测试) | ✅ `expected: 0, actual: 2`（MinVerfLevel 透传） | ✅ PASS | **PASS** |
| `FindByRoleId` assign_time 修复 (model/rel.go) | ✅ TestRelUserRoleModel_FindByRoleId | ✅ rel_test.go:520 `could not match actual sql ... "assign_time FROM"` | ✅ PASS | **PASS** |

- 全部 18 项均有测试，且 RED 列含**真实 FAIL 输出摘录**（编译期 `undefined`/`no field or method` + 行为期 `expected/actual`、sqlmock SQL 正则、miniredis `redis: nil`），非文字描述 → 按 must-follow [[tdd-red-evidence-requires-fail-excerpt]] 合规。
- 无 FAIL 行 → TDD 证据检查 **PASS**。

## 发现

| 级别 | 问题 | 建议 |
|------|------|------|
| INFO | `grantSatisfiedLevel`/`scopeState*`/`permissionDefMinLevel`/`userMaxLevel` 无独立单测，仅间接行为覆盖 | 后续可补直接单测（非阻塞） |
| INFO | `api/internal/logic/perm` 覆盖率仅 7.4%（仅 helpers 被测） | 非阻塞 |
| INFO | bench_regression SKIP（无 benchmark） | 热点路径（CheckPermission / GetDataScopes 缓存）建议补 benchmark |

---
VERDICT: PASS
---

## every-fresh-run 证据

```
# 1. harness-checks.sh --service permission-service --json (FRESH, 2026-08-12T09:34:46Z)
{"check":"go_build","status":"PASS","detail":"compilation succeeded"}
{"check":"go_vet","status":"PASS","detail":"no issues"}
{"check":"go_test","status":"PASS","detail":"4 packages passed, ~69 test functions"}
{"check":"proto_jstype","status":"PASS","detail":"no proto changes in diff (skipped)"}
{"check":"ast_json_string","status":"PASS","detail":"all int64 ID fields have json:",string" (AST verified)"}
{"check":"cross_service_import","status":"PASS","detail":"no violations"}
{"check":"error_codes","status":"PASS","detail":"no magic numbers found (all use named constants or 0)"}
{"check":"hardcoded_secrets","status":"PASS","detail":"no secrets detected"}
{"check":"graph_freshness","status":"PASS","detail":"graph up-to-date (synced 1h ago)"}
{"check":"claude_structural_data","status":"PASS","detail":"no structural data duplication in CLAUDE.md"}
{"check":"proto_ts_align","status":"PASS","detail":"all proto fields match TS interfaces"}
{"check":"api_stubs","status":"PASS","detail":"no TODO stubs found in API logic"}
{"check":"response_wrap","status":"PASS","detail":"no double-wrap risk detected"}
{"check":"bench_regression","status":"PASS","detail":"no benchmark functions — SKIP"}
{"check":"api_smoke","status":"PASS","detail":"no new routes detected in diff"}
{"check":"memory_index","status":"PASS","detail":"索引最新 (生成于 2026-08-12T09:16:10Z)"}
summary: {"pass": 16, "fail": 0, "warn": 0}, exit_code: 0

# 2. go build ./... (FRESH)
BUILD_EXIT=0

# 3. go vet ./... (FRESH)
VET_EXIT=0

# 4. go test ./... -count=1 (FRESH)
ok  api/internal/logic/perm       0.019s
ok  api/internal/types            0.003s
ok  model                         0.013s
ok  rpc/internal/logic/permission 0.076s
TEST_EXIT=0   （4 包, 69 测试函数, 0 FAIL）

# 5. go test ./... -cover (FRESH)
model                       69.2%
rpc/internal/logic/permission 61.1%
api/internal/types          31.2%
api/internal/logic/perm      7.4%
```
