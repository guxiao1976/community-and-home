# QA Report — user-service

**验证时间**: 2026-08-17 04:0x (UTC)
**验证范围**: 工作树未提交改动 + 未跟踪文件 — ApplyRole 服务角色免 membership 自助申请（模型拍板反转 + 逻辑函数）：新增 `needMembership()` + applyRole 分支反转 + 5 个测试函数 + CHANGELOG + go.mod tidy

## 变更范围（工作树 diff / FRESH 确认）

| 文件 | 变更 |
|------|------|
| `rpc/internal/logic/user/apply_role_logic.go` | 新增 `needMembership(roleCode string) bool`（居民角色 owner/tenant/committee→true；服务角色 grid_worker/community_admin/property_admin 与 merchant→false）；applyRole 分支反转：membership 校验改为 `if needMembership(in.RoleCode)`，服务角色免带房号 membership 直接 `AssignRole(scope=community, scope_id=communityId, status=0)`；`find membership error` 日志补 `userId/communityId/roleCode` 上下文；文件头补安全模型 doc（SEE: auto-grant-unverified-grant-confers-scope-level0） |
| `rpc/internal/logic/user/apply_role_logic_test.go` | 新增 5 测试函数：`TestApplyRole_ServiceRoles_NoMembership_Allowed`（3 角色 × DoAndReturn 断言 scope=community/2001/roleId/status=0）/ `TestApplyRole_ServiceRoles_WithMembership`（3 角色）/ `TestApplyRole_ServiceRoles_MembershipInactive` / `TestApplyRole_ResidentRoles_NoMembership`（3 角色→10005 回归）/ `TestNeedMembership`（7 角色表驱动） |
| `CHANGELOG.md` | 记录 2026-08-17 模型拍板反转（含 RED 摘录 + 安全模型说明） |
| `go.mod` | `miniredis/v2` indirect → direct（`go mod tidy` 卫生，既有测试 import） |

未跟踪文件：`_review_design-biz.md` / `_review_standards-eng.md`（评审记录，非本次源码）。

## 机械化检查结果 (harness-checks.sh — FRESH run)

运行：`bash .harness/skills/qa/scripts/harness-checks.sh --service user-service --json`
结果：**18 PASS / 0 FAIL / 3 WARN，exit_code 0**（timestamp 2026-08-17T03:59:26Z）

| # | 检查项 | 结果 | 详情 |
|---|--------|------|------|
| 1 | go build | ✅ | exit 0（"compilation succeeded"，手动复跑确认） |
| 2 | go vet | ✅ | exit 0（"no issues"，手动复跑确认） |
| 3 | go test | ✅ | 4 包通过（api/internal/logic/community、api/internal/logic/user、api/internal/types、rpc/internal/logic/user），**148 测试函数**，0 fail |
| 4 | Proto int64 jstype | ✅ | diff 无 proto 变更（skipped） |
| 5 | json:",string" | ✅ | 0 violations（AST verified，all int64 ID fields have json:",string"） |
| 6 | 跨服务DB导入 | ✅ | 0 violations（diff 扫描 2 个 Go 文件） |
| 7 | 错误码格式 | ✅ | 无 magic number（均用命名常量或 0） |
| 8 | 硬编码密钥 | ✅ | 无 secrets |
| 9 | graph_freshness | ✅ | 图同步于 3h 前，最新 |
| 10 | claude_structural_data | ✅ | CLAUDE.md 无结构数据重复 |
| 11 | proto_ts_align | ⚠️ | WARN：TS 滞后 proto（LoginSmsRequest.phone / RefreshTokenRequest.device_type / SubmitReviewRequest.reviewer_id / RegisterRequest.phone / User.avatar_url）——历史已有、本次无 proto 变更 |
| 12 | api_stubs | ✅ | API logic 无 TODO stub |
| 13 | response_wrap | ✅ | 无双层包装风险 |
| 14 | bench_regression | ✅ | 无 benchmark（SKIP） |
| 15 | api_smoke | ✅ | diff 无新增路由 |
| 16 | memory_index | ✅ | 索引最新 |
| 16.5 | design_consistency | ⚠️ | WARN：user-service model 引用列未覆盖标准迁移源 `deleted_at`——历史遗留 |
| 17 | git_hygiene | ⚠️ | WARN：gitlink 无 .gitmodules 条目（api-proto）——历史 git 治理漂移 |
| 18 | mutation_testing | ✅ | 变异分数 97.4%（≥80%） |
| pipeline_evals | 管线自身 eval | ✅ | 全部通过 |

## 编译检查
- [x] go build ./... — exit 0（FRESH，2026-08-17）

## 静态分析
- [x] go vet ./... — exit 0，clean（FRESH，2026-08-17）

## 单元测试
- [x] go test ./... -count=1 — 4/4 包通过，**148 测试函数**，0 fail（FRESH，禁用缓存，exit 0）

## 测试覆盖（新增测试专项）
- [x] `go test ./rpc/internal/logic/user/ -run 'TestNeedMembership|TestApplyRole_ServiceRoles|TestApplyRole_ResidentRoles_NoMembership' -count=1 -v` — 全部 PASS（FRESH，exit 0）
  - `TestApplyRole_ServiceRoles_NoMembership_Allowed` 3 子用例（grid_worker/community_admin/property_admin 无 membership → 允许，AssignRole 断言 ScopeType=community + ScopeId=2001 + RoleId + status=0）
  - `TestApplyRole_ServiceRoles_WithMembership` 3 子用例（服务角色 + active membership → 成功）
  - `TestApplyRole_ServiceRoles_MembershipInactive`（服务角色 + 已退出 membership → 仍允许）
  - `TestApplyRole_ResidentRoles_NoMembership` 3 子用例（owner/tenant/committee 无 membership → 10005 回归）
  - `TestNeedMembership` 7 子用例（owner/tenant/committee→true；grid_worker/community_admin/property_admin/merchant→false）
- 包覆盖率（非阻塞）：rpc/internal/logic/user 64.8%

## TDD 证据检查（分诊：字段映射 vs 有逻辑函数）
| 新增/修改函数 | 类型 | 是否有测试 | RED 确认（仅「有逻辑」要求） | GREEN 确认 | 状态 |
|-------------|------|:---:|:---:|:---:|:---:|
| `needMembership(roleCode string) bool`（新增） | 有逻辑（switch/分支） | ✅ `TestNeedMembership` 7 角色表驱动（含断言消息） | ✅ CHANGELOG 摘录 `apply_role_logic_test.go:433:29: undefined: needMembership`；**已 FRESH 核验**：HEAD 54e1a60 该文件 `needMembership` 出现 0 次（符号确为本次新增），当前测试文件 433 行 29 列恰为 `needMembership(tc.roleCode)` 引用位置——逐字匹配，编译 RED 真实 | ✅ 主树 `go test ./... -count=1` FRESH 148 测试函数 0 fail，`-run TestNeedMembership -v` 7 子用例全 PASS | PASS |
| `ApplyRole`（membership 校验分支反转 + 作用域重构） | 有逻辑（分支/条件/校验） | ✅ 新增 3 个服务角色测试 + 1 个居民角色回归测试 + 既有 11 用例回归 | ✅ CHANGELOG 摘录：`apply_role_logic_test.go:300: Not equal: expected: 0 / actual: 10005 / Messages: 服务角色无 membership 应允许自助申请`、`controller.go:137: missing call(s) to *mocks.MockPermissionServiceClient.AssignRole(...)`、`apply_role_logic_test.go:383: Not equal: expected: 0 / actual: 10005`（MembershipInactive）；**已 FRESH 核验路径**：旧实现（全部非 merchant 查 membership）下服务角色无 membership → 10005（断言期望 0 失败）+ 未触达 AssignRole（gomock missing call）——摘录与行为逐条吻合 | ✅ 主树全量测试 FRESH 0 fail；`-run 'TestApplyRole_ServiceRoles...' -v` 全部 PASS | PASS |

- **分诊结论**：`needMembership` 与 `ApplyRole` 分支反转均为**有逻辑函数**（分支/条件/校验），按规则要求 RED 摘录。
- **RED 真实性核验（FRESH）**：
  - `undefined: needMembership` — HEAD 该符号 0 次出现（结构证明），且当前测试文件 433 行 29 列 `needMembership(tc.roleCode)` 与 CHANGELOG 摘录行列**逐字一致**，非编造。
  - `expected: 0 / actual: 10005`（×2）+ `missing call(s) to AssignRole` — 与旧实现行为精确对应：旧实现所有非 merchant 角色必查 membership → 无 membership 服务角色返回 10005 且不触达 AssignRole。
- **GREEN（FRESH）**：`go build ./...` exit 0 + `go vet ./...` clean + `go test ./... -count=1` 148 函数 0 fail + 新增用例 `-v` 全 PASS。
- **测试有效性**：`_NoMembership_Allowed` 用 `AssignRole` `DoAndReturn` 断言 `ScopeType=community + ScopeId=2001 + RoleId + status=0`（验证数据范围来自角色 grant），`_ResidentRoles_NoMembership` 锁定「居民角色数据范围仍绑定 membership（10005）」不变量，能真实捕获「服务角色免 membership」行为回归。
- **安全说明（非 QA 判定依据）**：本变更为用户拍板**有意反转**上一轮 security-arch 回滚（CHANGELOG 记载），安全模型 = 盖章文件 + 人工审核 + permission-service `min_verf_level=2` 加固（未认证 status=0 grant 不能行使破坏性操作）。此为用户/安全评审决策边界，QA 只验证代码正确性与 TDD 证据，不代行安全裁决。

## 测试质量评估
- 新增/修改函数: 2（`needMembership` + `ApplyRole` 分支反转）/ 有测试: 2 / 缺失: 0
- 分支覆盖: 服务角色无 membership→允许+作用域/status=0 断言（3 角色）、有 active membership→成功（3 角色）、inactive membership→仍允许、居民角色无 membership→10005（3 角色）、needMembership 7 角色全码表、既有回归（Owner/Merchant/NoMembership/UserNotFound/FindOneError/FindMembershipError/MembershipNotActive/AssignRoleError/RoleCodeNotFound/PermissionClientNil）全绿
- 边界测试: ✅（无 membership / 非 active / 三种服务角色 / 三种居民角色 / merchant / 全错误路径回归）

## 发现
| 级别 | 问题 | 建议 |
|------|------|------|
| WARNING | proto_ts_align：TS 类型滞后 proto 5 处 | 前端 `identity.ts`/`moderation.d.ts` 待同步，独立任务跟进，不阻塞本次 |
| WARNING | design_consistency：model 引用列 `deleted_at` 未在迁移源 | 历史遗留，独立跟进 |
| WARNING | git_hygiene：api-proto gitlink 无 .gitmodules 条目 | 补 `.gitmodules`，历史漂移 |

> 注：3 项 WARN 均为历史已有/与本次 diff 无关，非本次变更引入。无 FAIL。

---
VERDICT: PASS — 机械化检查 18 PASS / 0 FAIL / 3 WARN + go build/vet/test 全部 exit 0（148 测试函数 0 fail）+ 有逻辑变更点（`needMembership` 新增 + `ApplyRole` 分支反转）配真实行为断言测试（RED 摘录具体 error 文本且经 FRESH 核验：`undefined: needMembership` 行列逐字匹配 HEAD 缺失证据、`expected 0/actual 10005` 与旧行为吻合、GREEN 全绿），TDD 证据充分。
---
