# QA Report — user-service

**验证时间**: 2026-08-16 15:33 (UTC) / 23:33 (本地)
**验证范围**: 工作树未提交改动 — GetProfileLogic 补传 ViewerId（profile 端点本人手机号脱敏修复，接线类型）+ 对应测试 + CHANGELOG + go.mod tidy

## 变更范围（工作树 diff）

| 文件 | 变更 |
|------|------|
| `api/internal/logic/user/user_logic.go` | `GetProfile()` 调 `UserRpc.GetUser` 补 `ViewerId: &viewerID`（=userId，原为 0） |
| `api/internal/logic/user/user_logic_test.go` | 新增 `TestGetProfileLogic_GetProfile`（4 用例）+ 自定义 `getUserReqMatcher` gomock.Matcher |
| `CHANGELOG.md` | 记录 2026-08-16 修复 |
| `go.mod` | `miniredis/v2` 从 indirect 提升为 direct（`go mod tidy` 卫生：`get_user_roles_logic_test.go` 直接 import） |
| `docs/graph-context.md` | 上下文同步 |

无未跟踪文件涉及 user-service。

## 机械化检查结果 (harness-checks.sh — FRESH run)

运行：`bash .harness/skills/qa/scripts/harness-checks.sh --service user-service --json`
结果：**18 PASS / 0 FAIL / 3 WARN，exit_code 0**

| # | 检查项 | 结果 | 详情 |
|---|--------|------|------|
| 1 | go build | ✅ | exit 0（手动复跑确认） |
| 2 | go vet | ✅ | exit 0（手动复跑确认） |
| 3 | go test | ✅ | 4 包通过（api/internal/logic/community、api/internal/logic/user、api/internal/types、rpc/internal/logic/user），143 测试函数，0 fail |
| 4 | Proto int64 jstype | ✅ | diff 无 proto 变更（skipped） |
| 5 | json:",string" | ✅ | 0 violations（AST verified） |
| 6 | 跨服务DB导入 | ✅ | 0 violations（diff 扫描 2 个 Go 文件） |
| 7 | 错误码格式 | ✅ | 无 magic number（均用命名常量或 0） |
| 8 | 硬编码密钥 | ✅ | 无 secrets |
| 9 | graph_freshness | ✅ | 图同步于 5h 前，最新 |
| 10 | claude_structural_data | ✅ | CLAUDE.md 无结构数据重复 |
| 11 | proto_ts_align | ⚠️ | WARN：TS 滞后 proto（LoginSmsRequest.phone / RefreshTokenRequest.device_type / SubmitReviewRequest.reviewer_id / RegisterRequest.phone / User.avatar_url）—— 均为历史已有、与本次 diff 无关 |
| 12 | api_stubs | ✅ | API logic 无 TODO stub |
| 13 | response_wrap | ✅ | 无双层包装风险 |
| 14 | bench_regression | ✅ | 无 benchmark（SKIP） |
| 15 | api_smoke | ✅ | diff 无新增路由 |
| 16 | memory_index | ✅ | 索引最新 |
| 16.5 | design_consistency | ⚠️ | WARN：user-service model 引用列未覆盖标准迁移源 `deleted_at` —— 历史遗留模型列，与本次 diff 无关 |
| 17 | git_hygiene | ⚠️ | WARN：gitlink 无 .gitmodules 条目（api-proto）—— 历史 git 治理漂移，非本次改动引入 |
| 18 | mutation_testing | ✅ | 变异分数 ?%（未解析到分数，不阻塞） |
| pipeline_evals | 管线自身 eval | ✅ | 全部通过 |

## 编译检查
- [x] go build ./...  — exit 0

## 静态分析
- [x] go vet ./...  — exit 0，clean

## 单元测试
- [x] go test ./... -count=1  — 4/4 包通过，143 测试函数，0 fail

## 测试覆盖（新增测试专项）
- [x] `go test ./api/internal/logic/user/ -run TestGetProfileLogic_GetProfile -count=1 -v` — 4/4 子用例 PASS（success 明文手机号 / Base 10001 透出 / 未登录 error / RPC 失败 error）

## TDD 证据检查（分诊：字段映射 vs 有逻辑函数）
| 新增/修改函数 | 类型 | 是否有测试 | RED 确认（仅「有逻辑」要求） | GREEN 确认 | 状态 |
|-------------|------|:---:|:---:|:---:|:---:|
| `GetProfileLogic.GetProfile` | 接线（补 RPC 请求字段 ViewerId） | ✅ `TestGetProfileLogic_GetProfile`（自定义 Matcher 断言 `viewer_id==userId`） | —（接线类，不要求） | ✅ 4/4 子用例 PASS | PASS |

- **分诊结论**：本次唯一业务变更 `GetProfile()` 为**纯接线/字段透出类**（在已有 RPC 请求上补传 `ViewerId` 字段，无分支/计算/转换逻辑），masking 语义未动，按规则**不要求 RED 摘录**。
- **测试有效性**：测试用自定义 `getUserReqMatcher` 断言 `GetUserRequest.id==1001 && viewer_id==1001`，能真实捕获「漏传 ViewerId」回归（`missing call`），非仅"调用不报错"。边界覆盖：未登录、Base 业务错误透出、RPC 失败、正常明文。
- CHANGELOG 补录了 RED 摘录（`controller.go:269: missing call(s)...`），与接线类分诊一致，作为开发过程证据记录。

## 测试质量评估
- 新增函数: 1（GetProfile 修改）/ 有测试: 1 / 缺失: 0
- 边界测试: ✅（未登录 / Base 错误 / RPC 失败 / 成功路径全覆盖）

## 发现
| 级别 | 问题 | 建议 |
|------|------|------|
| WARNING | proto_ts_align：TS 类型滞后 proto 5 处 | 前端 `identity.ts`/`moderation.d.ts` 待同步，独立任务跟进，不阻塞本次 |
| WARNING | design_consistency：model 引用列 `deleted_at` 未在迁移源 | 历史遗留，独立跟进 |
| WARNING | git_hygiene：api-proto gitlink 无 .gitmodules 条目 | 补 `.gitmodules`，历史漂移 |

> 注：3 项 WARN 均为历史已有/与本次 diff 无关，非本次变更引入。

---
VERDICT: PASS — 机械化检查 18 PASS / 0 FAIL + go build/vet/test 全部 exit 0 + 新增接线类变更配有真实行为断言测试（GREEN 4/4），TDD 证据充分。
---
