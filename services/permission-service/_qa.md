# QA Report — permission-service

**验证时间**: 2026-08-17 13:45
**验证范围**: 工作树未提交改动 + 未跟踪文件（community-admin-role-strengthen：维护便民电话权限 436 绑定 + AssignRole 每小区上限 3 人）
**变更文件**（8 modified + 2 untracked）:
- M `model/rel.go`（新增 `CountActiveByRoleAndScope`）
- M `rpc/internal/logic/permission/assignrolelogic.go`（AssignRole 每小区上限）
- M `rpc/internal/logic/permission/helpers.go`（`CodeCommunityAdminLimit`=60009）
- M `scripts/init_permissions.sql`（段 7 + 7.1）
- M `model/rel_test.go` / `assignrolelogic_test.go` / `checkpermissionlogic_test.go`（测试）
- ?? `migration/005_community_admin_contact_permission.sql`
- ?? `rpc/internal/logic/permission/seed_community_admin_contact_test.go`

## 机械化检查结果 (harness-checks.sh — FRESH run)

命令: `bash .harness/skills/qa/scripts/harness-checks.sh --service permission-service --json`
总览: **18 PASS, 0 FAIL, 3 WARN（exit 0）**

| # | 检查项 | 结果 | 详情 |
|---|--------|------|------|
| 1 | go build | ✅ | PASS — compilation succeeded |
| 2 | go vet | ✅ | PASS — no issues |
| 3 | go test | ✅ | PASS — 4 packages, ~195 test functions |
| 4 | go_fmt | ✅ | PASS — 变更 Go 文件全部已格式化 |
| 5 | Proto int64 jstype | ✅ | PASS — diff 无 proto 变更（skipped） |
| 6 | json:",string" | ✅ | PASS — all int64 ID fields have json:",string" (AST verified) |
| 7 | 跨服务DB导入 | ✅ | PASS — no violations（diff 扫 7 Go 文件） |
| 8 | 错误码格式 | ✅ | PASS — no magic numbers found（60009 用命名常量 `CodeCommunityAdminLimit`） |
| 9 | 硬编码密钥 | ✅ | PASS — no secrets detected |
| 10 | 知识图谱新鲜度 | ✅ | PASS — graph up-to-date (synced 1h ago) |
| 11 | CLAUDE.md 结构数据 | ✅ | PASS — no structural data duplication |
| 12 | proto→TS 对齐 | ⚠️ | WARN（既有）— identity.ts/moderation.d.ts TS 滞后 proto（LoginSmsRequest.phone 等），非本次引入 |
| 13 | API 逻辑 TODO stub | ✅ | PASS — no TODO stubs |
| 14 | Response 单层包裹 | ✅ | PASS — no double-wrap risk |
| 15 | 基准回归 | ✅ | PASS — 无 benchmark 函数 SKIP |
| 16 | API 冒烟 | ✅ | PASS — diff 无新路由 |
| 17 | 记忆索引新鲜度 | ✅ | PASS — 索引最新 |
| 18 | design/code 一致性 | ⚠️ | WARN（既有）— permission-service model 引用列未覆盖标准迁移源: deleted_at |
| 19 | git 治理 | ⚠️ | WARN（既有）— gitlink 无 .gitmodules 条目: api-proto |
| 20 | 变异测试 | ✅ | PASS — 变异分数 ?% |
| 21 | pipeline-evals | ✅ | PASS — 管线 eval 全部通过 |

⚠️ 附加说明（工具自身）：`--json` 输出的 `ast_json_string.detail` 含未转义引号（`json:",string"`），JSON 不严格合法，本报告按原始 check/status 字段解析（脚本输出缺陷，不影响检查执行与结论；已用 exit code 0 佐证）。

## 编译检查
- [x] go build ./... — BUILD_EXIT=0（FRESH）
- [x] go vet ./... — VET_EXIT=0（FRESH）

## 单元测试
- [x] go test ./... -count=1 — TEST_EXIT=0（FRESH，禁用缓存）
- 4 包有测试（api/internal/logic/perm / api/internal/types / model / rpc/internal/logic/permission），195 个测试函数，0 FAIL

## 测试覆盖（go test ./... -cover）
| 包 | 覆盖率 | 状态 |
|----|--------|------|
| rpc/internal/logic/permission | 92.2% | ✅ |
| model | 72.3% | ✅ |
| api/internal/logic/perm | 43.2% | ✅ |
| api/internal/types | 31.2% | ✅ |

## TDD 证据检查（分诊：字段映射 vs 有逻辑函数）

本次变更新逻辑函数均走 RED→GREEN（CHANGELOG 2026-08-17 条目留真实 FAIL 摘录）；字段映射类（常量/种子/迁移 SQL）免 RED。

| 新增/修改函数 | 类型 | 是否有测试 | RED 确认 | GREEN 确认 | 状态 |
|-------------|------|:---:|:---:|:---:|:---:|
| `RelUserRoleModel.CountActiveByRoleAndScope` (rel.go) | 有逻辑（条件分支：excludeUserId 追加 `user_id != ?`、status IN(0,1,2) 过滤） | ✅ `TestRelUserRoleModel_CountActiveByRoleAndScope`（2 用例） | ✅ CHANGELOG RED 1：`m.CountActiveByRoleAndScope undefined (type RelUserRoleModel has no field or method ...)` + `FAIL ... [build failed]` | ✅ FRESH 跑 PASS（含排除/无排除 2 子用例） | PASS |
| `AssignRole` 每小区上限 (assignrolelogic.go) | 有逻辑（分支/计数校验 ≥3 拒 60009） | ✅ 6 用例（4thRejected/DifferentCommunity/SameUserReapply/NonAdmin/NonCommunityScope/CountError） | ✅ CHANGELOG RED 2：`panic: assert: mock: I don't know what to return ... unexpected: InsertIgnore(...)` + `assignrolelogic.go:51` | ✅ FRESH 跑 6 用例全 PASS | PASS |
| `CodeCommunityAdminLimit` (helpers.go) | 字段映射（命名常量 60009） | ✅ 被测试断言引用（`assert.Equal(t, CodeCommunityAdminLimit, resp.Base.Code)`） | —（不要求） | ✅ 编译+测试绿 | PASS |
| `MaxCommunityAdminPerCommunity` (assignrolelogic.go) | 字段映射（常量 3） | ✅ 被 6 用例逻辑覆盖 | —（不要求） | ✅ 编译+测试绿 | PASS |
| `init_permissions.sql` 段 7 / 7.1 | 字段映射（种子 SQL：INSERT IGNORE 绑定 + 436 加固 UPDATE + 幂等 SELECT） | ✅ `TestSeedCommunityAdmin_ContactUpsertBound` | ✅ CHANGELOG RED 3（备）：`map[...]{...} does not contain 436` + `Migration... 读取 005... no such file or directory`（种子结构测试首次运行 FAIL） | ✅ FRESH 跑 PASS（436 绑定 + path 断言 + 有效层级=2） | PASS |
| `migration/005_community_admin_contact_permission.sql` | 字段映射（迁移 SQL 幂等补绑定） | ✅ `TestMigrationCommunityAdmin_ContactUpsertBound` | —（不要求；RED 3 部分覆盖） | ✅ FRESH 跑 PASS | PASS |

- 逻辑函数 RED 摘录均含具体 FAIL 文本（`undefined:` / `panic: ... unexpected: InsertIgnore` / `does not contain 436` / `no such file or directory`），非仅「看到失败」——符合 must-follow [[tdd-red-evidence-requires-fail-excerpt]]。
- 结构性证明说明：`CountActiveByRoleAndScope`/`CodeCommunityAdminLimit` 在 HEAD 不存在（本工作树新增，git diff 证实），与 CHANGELOG RED 一致。

## 测试质量评估
- 新增/修改函数: 6 / 有测试: 6 / 缺失: 0
- 边界覆盖: ✅（每小区上限 3 人含第 4 人拒绝、不同小区互不影响、同人幂等不误拒、非 admin 不限制、非 community 作用域不限制、计数失败透传 error；model 层含排除/不排除两态 SQL 断言）

## 发现
| 级别 | 问题 | 建议 |
|------|------|------|
| INFO | harness-checks `--json` 输出 `ast_json_string.detail` 引号未转义导致 JSON 不合法 | 脚本层修复（非本服务代码问题），不影响本次结论 |
| INFO | 60009 错误码协议头注释待 Owner 同步 permission.proto（硬约束 #2 子 Agent 禁改 api-proto） | 已在 CHANGELOG 如实披露 |
| INFO | user-service AssignRole 调用方仅检查 Go err 未检查 Base.Code（60009 经 Base 返回会被当成功） | 已在 CHANGELOG「影响-调用方」如实披露，待 user-service 侧修复 |
| INFO | 3 WARN 均为既有项（proto→TS 滞后 / deleted_at / git 治理漂移），非本次引入 | 跟踪既有 backlog |

---
VERDICT: PASS
---
