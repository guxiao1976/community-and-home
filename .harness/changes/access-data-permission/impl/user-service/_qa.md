# QA Report — user-service

**验证时间**: 2026-08-12 10:25 UTC
**验证范围**: 工作树未提交改动 + 未跟踪文件（数据权限核心编排 阶段③：T3.1 注册自动授权 / T3.2 加入授权 / T3.3 退出撤销 / T3.4 门禁）
**执行方式**: 全部检查 fresh run（禁用缓存、独立跑），结果见下表

## 机械化检查结果 (harness-checks.sh — FRESH run)

命令: `bash .harness/skills/qa/scripts/harness-checks.sh --service user-service --json`
结果: **16 PASS / 0 FAIL / 0 WARN，exit_code=0**

| # | 检查项 | 结果 | 详情 |
|---|--------|------|------|
| 1 | go build | ✅ | exit 0 — compilation succeeded |
| 2 | go vet | ✅ | exit 0 — no issues |
| 3 | go test | ✅ | 4 packages passed, ~71 test functions |
| 4 | Proto int64 jstype | ✅ | no proto changes in diff (skipped) |
| 5 | json:",string" | ✅ | 0 violations — all int64 ID fields have `json:",string"`（AST verified，含 submit_certification_logic.go certMetadata MembershipId/CommunityId 修复） |
| 6 | 跨服务DB导入 | ✅ | 0 violations（diff scan 12 Go 文件） |
| 7 | 错误码格式 | ✅ | 0 magic numbers（全部命名常量） |
| 8 | 硬编码密钥 | ✅ | no secrets detected |
| 9 | graph_freshness | ✅ | graph up-to-date（synced 0h ago） |
| 10 | CLAUDE.md 结构数据 | ✅ | no structural data duplication |
| 11 | Proto→TS 对齐 | ✅ | all proto fields match TS interfaces |
| 12 | API logic TODO stubs | ✅ | no TODO stubs found |
| 13 | Response 单层包装 | ✅ | no double-wrap risk |
| 14 | Benchmark 回归 | ✅ | no benchmark functions — SKIP（提示：热路径可补 benchmark） |
| 15 | API smoke | ✅ | no new routes detected in diff |
| 16 | Memory Index 新鲜度 | ✅ | 索引最新 |

## TDD 证据检查

范围：本次工作树新增/修改的函数（含 `git show HEAD:` 结构性证明，确认新符号在 HEAD 不存在 → RED 为真实编译期/期望失败）。

| 新增/修改函数 | 是否有测试 | RED 确认（含 FAIL 输出摘录） | GREEN 确认 | 状态 |
|-------------|:---:|:---:|:---:|:---:|
| `CreateUserLogic.assignRegisteredUser` | ✅ (create_user_logic_test.go ×3) | 摘录见 CHANGELOG: `controller.go:137: missing call(s) to ...AssignRole` ×3；结构证明 `git show HEAD:` 无此符号 | `TestCreateUser_AssignsRegisteredUser/DuplicatePhone_NoReAssign/AssignRoleFailure_DoesNotBlockRegistration` PASS | PASS |
| `assignRoleToUser` (helper.go) | ✅ (经上述 + join/leave 测试) | 摘录 `missing call(s) to ...AssignRole` + `missing call(s) to ...RevokeRole` | 全部 PASS | PASS |
| `stringPtr` (helper.go) | ✅ (经 leave 测试) | 结构证明 HEAD 无此符号 | PASS | PASS |
| `JoinCommunityLogic.ownershipRoleCode` | ✅ (join_community_ownership_test.go ×2) | 摘录 `missing call(s) to ...AssignRole` | `TestJoinCommunity_Owned_AssignsOwnerGrant` / `_Rented_AssignsTenantGrant` PASS (100%) | PASS |
| `JoinCommunityLogic.assignCommunityRole` | ✅ (×5 + 更新 5 用例) | 摘录 `missing call(s) to ...AssignRole` + 10040 断言失败 | 全 PASS | PASS |
| `LeaveCommunityLogic.revokeCommunityRoles` | ✅ (leave_community_revoke_test.go ×4) | 摘录 `missing call(s) to ...RevokeRole` ×3 | `TestLeaveCommunity_RevokesOwnerAndTenant/OtherCommunityPreserved/RevokeFailure_RestoresAndFails/DuplicateLeave_Returns10005` PASS | PASS |

- RED 列均含**具体 FAIL 摘录**（gomock `missing call(s) to ...` 实际错误文本 + 断言目标 10040），非仅"看到失败"文字描述。
- 结构性佐证（只读）：3 个新测试文件均不在 HEAD（未跟踪）；`assignRegisteredUser / assignRoleToUser / stringPtr / ownershipRoleCode / assignCommunityRole / revokeCommunityRoles` 均不在 HEAD → 测试先于实现存在时编译/期望必然失败，RED 真实。
- 说明：`controller.go:137` 为 Generator 记录的 gomock 期望失败定位（mockgen controller 惯例）；`10040 断言失败` 为断言目标摘录，未含完整 testify Error Trace（minor，见发现表）。

## 编译检查
- [x] `go build ./...` — exit 0（fresh）

## 静态分析
- [x] `go vet ./...` — clean output, exit 0（fresh）

## 单元测试
- [x] `go test ./... -count=1` — **4/4 包通过**（禁用缓存），共 **71** 个测试函数，0 fail
  - rpc/internal/logic/user: ok (60 个测试函数)
  - api/internal/logic/community: ok
  - api/internal/logic/user: ok
  - api/internal/types: ok
  - 新增 12 个测试函数：create_user(3) + join_community_ownership(5) + leave_community_revoke(4)

## 测试覆盖
| 包 | 覆盖率 | 状态 |
|----|--------|------|
| rpc/internal/logic/user | 51.8% | 核心逻辑包 |
| api/internal/logic/community | 19.7% | 薄 RPC 代理层 |
| api/internal/logic/user | 36.5% | — |

新增函数逐项覆盖率（`go tool cover`）：
- `assignRegisteredUser` 75.0% / `assignRoleToUser` 75.0% / `ownershipRoleCode` 100.0% / `assignCommunityRole` 75.0% / `revokeCommunityRoles` 66.7%

## 测试质量评估
- 新增函数: 6 / 有测试: 6 / 缺失: 0
- 断言真实：全部使用 testify `assert`/`require` + gomock `EXPECT().Times(n)`（非"调用不报错"），逐参数校验 grant 内容（role_id / scope_type / scope_id / status）
- 边界/错误路径覆盖：
  - CreateUser：AssignRole 失败不阻塞注册（非阻塞告警路径）、重复注册幂等（10002 不重复分配）
  - JoinCommunity：UNSPECIFIED→10040、授权失败补偿 membership→left、重复加入→10007 不重复授权、rejoin 重新授权（更新测试）
  - LeaveCommunity：双调撤销 owner+tenant、其他小区保留、撤销失败恢复 active、重复 leave→10005
- 覆盖充分，无缺边界场景阻塞项

## 发现
| 级别 | 问题 | 建议 |
|------|------|------|
| WARNING | CHANGELOG 中 ownership 10040 的 RED 摘录为"10040 断言失败"，未含 testify 完整 Error Trace（`Error Trace:` / `Error:` / `Test:` 行）；按 memory `tdd-red-evidence-requires-fail-excerpt.md` 严格标准可更规范 | 后续 RED 阶段直接持久化完整断言失败文本 |
| INFO | `revokeCommunityRoles` 覆盖 66.7%（未测 PermissionClient==nil 分支） | 可补 1 例 nil-client 短路用例 |
| INFO | LeaveCommunity 补偿恢复用 `time.Time{}`（零值）更新 bind_status，遗留 leave_time 零值语义 | 与既有模式一致，非本次引入，观察即可 |
| INFO | bench_regression 为 SKIP（无 benchmark） | 热路径（CreateUser/JoinCommunity）建议后续补 benchmark |

---
VERDICT: **PASS**

- 机械化检查 16/16 PASS，0 FAIL 0 WARN（exit 0）
- `go build` / `go vet` / `go test -count=1` 全部 fresh exit 0
- TDD 证据：6/6 新函数有测试；RED 列含具体 gomock FAIL 摘录 + `git show HEAD:` 结构性证明（新符号在 HEAD 不存在）；GREEN 全绿
- 新函数覆盖率 66.7%–100%，边界/错误路径覆盖充分
---
