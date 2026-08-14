# QA Report — permission-service

**验证时间**: 2026-08-14 12:28
**验证范围**: 工作树未提交改动 + 未跟踪文件（role-list-sort：FindList 签名扩展 + orderByClause + validateSort + API 透传/ToError 修复）
**分支**: master/main（工作树 diff）

## 机械化检查结果 (harness-checks.sh — FRESH run)

`bash .harness/skills/qa/scripts/harness-checks.sh --service permission-service --json` — 2026-08-14T04:27:15Z，exit_code=0

| # | 检查项 | 结果 | 详情 |
|---|--------|------|------|
| 1 | go build | ✅ | exit 0 — compilation succeeded |
| 2 | go vet | ✅ | exit 0 — no issues |
| 3 | go test | ✅ | 4 包通过（api/internal/logic/perm、api/internal/types、model、rpc/internal/logic/permission），~151 测试函数，0 FAIL |
| 4 | Proto int64 jstype | ✅ | 本次 diff 无 proto 变更（skipped） |
| 5 | json:",string" | ✅ | AST 校验：所有 int64 ID 字段均有 json:",string" |
| 6 | 跨服务DB导入 | ✅ | no violations（diff scan 11 Go 文件） |
| 7 | 错误码格式 | ✅ | 无裸数字，validateSort/ListRoles 用 errx.CodeInvalidParam 常量 |
| 8 | 硬编码密钥 | ✅ | no secrets detected |
| 9 | graph_freshness | ✅ | graph up-to-date（synced 0h ago） |
| 10 | claude_structural_data | ✅ | 无 CLAUDE.md 结构数据重复 |
| 11 | proto_ts_align | ✅ | 全部 proto 字段匹配 TS 接口 |
| 12 | api_stubs | ✅ | API logic 无 TODO stub |
| 13 | response_wrap | ✅ | 无双重包装风险 |
| 14 | bench_regression | ✅ | 无 benchmark 函数（SKIP，建议热路径补） |
| 15 | api_smoke | ✅ | 无新路由（GET /api/perm/roles 为既有路由） |
| 16 | memory_index | ✅ | 索引最新 |
| 17 | git_hygiene | ⚠️ | **WARN** — gitlink api-proto 无 .gitmodules 条目（预存仓库级状态，非本次变更引入；建议补 .gitmodules 登记） |
| 18 | mutation_testing | ✅ | 变异分数未解析（默认放行） |

汇总：**17 PASS, 0 FAIL, 1 WARN**。

## 编译检查
- [x] go build ./...（FRESH，BUILD_EXIT=0）

## 静态分析
- [x] go vet ./...（FRESH，VET_EXIT=0，clean output）

## 单元测试
- [x] go test ./... -count=1（FRESH，TEST_EXIT=0，4/4 包通过，~151 测试函数，0 FAIL）

定向 FRESH 复验（本次新增函数全部 GREEN）：
- `go test ./model/ -run 'TestOrderByClause|TestSysRoleModel_FindList' -v` → PASS（TestSysRoleModel_FindList_WithSortField 断言 `order by role_name desc, id asc`、TestOrderByClause 8 subtests 全 PASS）
- `go test ./rpc/internal/logic/permission/ -run 'TestValidateSort|TestListRoles' -v` → PASS（TestValidateSort 11 subtests、TestListRoles_SortPassthrough/InvalidSortField/InvalidSortOrder/SortEmptyFieldNoError 全 PASS）
- `go test ./api/internal/logic/perm/ -run TestListRoles -v` → PASS（SortPassThrough/SortOnlyNoOrder/BaseErrorToGoError 全 PASS）
- `go test ./api/internal/types/ -run TestListRolesReq -v` → PASS（SortFormTags）

## TDD 证据检查（分诊：字段映射 vs 有逻辑函数）

| 新增/修改函数 | 类型 | 是否有测试 | RED 确认（仅「有逻辑」要求） | GREEN 确认 | 状态 |
|-------------|------|:---:|:---:|:---:|:---:|
| SysRoleModel.FindList（签名 4→6 参扩展 + orderByClause 接线） | 有逻辑（排序/分支） | ✅ | ✅ RED1：`model/permission_test.go:133:86: too many arguments in call to m.FindList` / `have (context.Context, nil, int64, int64, string, string)` / `want (context.Context, *int64, int64, int64)` | ✅ 定向复验 PASS | PASS |
| orderByClause（model 新纯函数） | 有逻辑（白名单/方向分支/回落） | ✅ TestOrderByClause 8 cases | ✅ 经 Task 1.1 RED1 编译期捕获（`too many arguments in call to m.FindList`，实现与签名变更同批引入；`git show HEAD:` 确认 orderByClause/roleSortFieldWhitelist 在 HEAD 不存在） | ✅ 定向复验 8 subtests PASS | PASS |
| validateSort（sort.go 新文件） | 有逻辑（校验/分支） | ✅ TestValidateSort 11 cases | ✅ RED2：`rpc/internal/logic/permission/sort_test.go:36:25: undefined: validateSort`（sort.go 为未跟踪新文件，HEAD 不存在） | ✅ 定向复验 11 subtests PASS | PASS |
| ListRoles（rpc 排序集成 + Base 业务错误） | 有逻辑（校验/分支） | ✅ 4 tests | ✅ RED3：`panic: mock: Unexpected Method Call` / `FindList(...,string,string) Diff: 4: FAIL: (string=) != (string=role_name)` | ✅ 定向复验 4 tests PASS | PASS |
| ListRoles（api Sort 透传 + ToError 修复） | 有逻辑（分支/转换） | ✅ 3 tests | ✅ RED4：`Expected value not to be nil` / `Messages: sort 参数应透传到 gRPC 请求` | ✅ 定向复验 3 tests PASS | PASS |
| ListRolesReq.SortBy/SortOrder（types.go） | 字段映射（struct 加字段） | ✅ TestListRolesReq_SortFormTags | —（不要求） | ✅ 定向复验 PASS | PASS |
| MockRoleModel.FindList（assignrolelogic_test.go 签名同步） | 测试接线 | ✅（签名同步） | —（不要求） | ✅ build/test 绿 | PASS |

- **字段映射类**（ListRolesReq.SortBy/SortOrder）：无独立逻辑，有对应测试（TestListRolesReq_SortFormTags），RED 列记 —（不要求）。
- **有逻辑函数**（FindList/orderByClause/validateSort/rpc ListRoles/api ListRoles）：RED 列均有具体 FAIL 输出摘录（`too many arguments` / `undefined: validateSort` / mock Diff / `Expected value not to be nil`），非仅文字描述；GREEN 已由 QA 本次定向 FRESH 复验全 PASS。
- **结构性证明补充**：`git show HEAD:` 确认 orderByClause、roleSortFieldWhitelist、validateSort、sort.go 在 HEAD 均不存在；FindList 在 HEAD 为 4 参签名 — 印证 RED 真实，且与 CHANGELOG 摘录互为佐证。
- 无「有逻辑函数 RED 缺失」，不判 TDD 证据 FAIL。

## 测试覆盖
| 包 | 覆盖情况 | 状态 |
|----|--------|------|
| model | FindList 排序/分页/状态过滤 + orderByClause 全分支 | ✅ |
| rpc/internal/logic/permission | validateSort 11 cases + ListRoles 排序集成（非法字段/方向/空字段） | ✅ |
| api/internal/logic/perm | ListRoles 透传 + Base→Go error | ✅ |
| api/internal/types | ListRolesReq form tag 契约 | ✅ |

（覆盖率百分比非本次门禁项，未阻塞）

## 测试质量评估
- 新增函数: 5（FindList 扩展/orderByClause/validateSort/rpc ListRoles 集成/api ListRoles ToError）/ 有测试: 5 / 缺失: 0
- 边界测试: ✅（orderByClause：注入载荷被拒、空字段、非法方向、大写归一；validateSort：空字段+非空方向不报错 REQ-4、非法方向 99400、注入载荷被拒）
- 安全：ORDER BY 仅拼接白名单字面量，注入载荷（`role_name; drop table sys_role`）被 RPC 白名单 + model 二次防御两层拦截

## 发现
| 级别 | 问题 | 建议 |
|------|------|------|
| WARN | git_hygiene：api-proto gitlink 无 .gitmodules 条目 | 预存仓库级状态，非本次变更引入；建议全局补 .gitmodules 登记 |
| INFO | bench_regression 无 benchmark 函数 | 热路径（FindList/validateSort）可补基准测试（非阻塞） |
| INFO | orderByClause 无独立 `undefined: orderByClause` 摘录，RED 经 Task 1.1 FindList 编译错误体现 | 可接受：orderByClause 与 FindList 签名变更同批引入，编译期 RED 为同一时刻；如需更严格可回溯补独立摘录（不阻塞 PASS） |

---
VERDICT: **PASS**
---

**证据（every-fresh-run）**：
- `harness-checks.sh --service permission-service --json` → 17 PASS / 0 FAIL / 1 WARN（git_hygiene），exit 0
- `go build ./...` → BUILD_EXIT=0
- `go vet ./...` → VET_EXIT=0（clean）
- `go test ./... -count=1` → TEST_EXIT=0（4 包，~151 测试函数，0 FAIL）
- 4 组定向 `go test -run` 复验新增函数 → 全 PASS
