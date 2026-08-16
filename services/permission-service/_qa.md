# QA Report — permission-service

**验证时间**: 2026-08-16 11:35
**验证范围**: 工作树未提交改动 + 未跟踪文件（content-post-generalization Task 3.1+3.2：AssertPublishScope 社区管理员角色感知展开 + 权限种子矩阵）
**分支**: main（工作树 diff）
**检查方式**: every-fresh-run（harness-checks.sh --json / go build / go vet / go test -count=1 均本次现场执行）

## 机械化检查结果 (harness-checks.sh — FRESH run)

| # | 检查项 | 结果 | 详情 |
|---|--------|------|------|
| 1 | go build | ✅ | exit 0（独立复跑 `go build ./...` → BUILD_EXIT=0） |
| 2 | go vet | ✅ | exit 0（独立复跑 `go vet ./...` → VET_EXIT=0） |
| 3 | go test | ✅ | 4 包通过，~182 测试函数，0 FAIL（`go test ./... -count=1` → TEST_EXIT=0） |
| 4 | Proto int64 jstype | ✅ | diff 无 proto 变更（skipped），0 violations |
| 5 | json:",string" | ✅ | AST 检查全部 int64 ID 字段合规，0 violations |
| 6 | 跨服务DB导入 | ✅ | 3 Go 文件 diff 扫描，0 violations |
| 7 | 错误码格式 | ✅ | 无裸魔数（全部命名常量或 0） |
| 8 | 硬编码密钥 | ✅ | 无密钥泄露 |

> harness-checks.sh 汇总：**18 PASS, 0 FAIL, 3 WARN**，exit_code=0。3 个 WARN 均为既有问题、非本次引入（与 CHANGELOG 声明一致）：
> - `proto_ts_align`（WARN）— TS 滞后 proto 字段（identity.ts LoginSmsRequest.phone / RefreshTokenRequest.device_type / RegisterRequest.phone / User.avatar_url；moderation.d.ts SubmitReviewRequest.reviewer_id），待前端同步
> - `design_consistency`（WARN）— permission-service model 引用列未覆盖标准迁移源: `deleted_at`
> - `git_hygiene`（WARN）— gitlink 无 .gitmodules 条目: api-proto

## 编译检查
- [x] go build ./...（BUILD_EXIT=0，FRESH）

## 静态分析
- [x] go vet ./...（VET_EXIT=0，FRESH，clean output）

## 单元测试
- [x] go test ./... -count=1（4/4 包通过，0 FAIL，TEST_EXIT=0；全包 ~182 测试函数）
- 新增测试 `TestAssertPublishScope_CommunityAdminDivisionExpansion`（assertpublishscope_division_test.go）：14 个 table 子用例全部 PASS（FRESH -v 复跑确认）

## 测试覆盖
| 包 | 覆盖率 | 状态 |
|----|--------|------|
| rpc/internal/logic/permission | 92.1% | 覆盖新增 has-logic 函数全部路径（展开/非admin不展开/过期驳回不展开/多division并集/fail-closed） |

## 测试质量评估
- 新增 has-logic 函数: 3（resolvePublishScope / holdsCommunityAdminRole / expandCommunityAdminDivision）；全部有对应测试（经 AssertPublishScope RPC 入口间接覆盖）：✅
- 边界/防御覆盖：空 grants→EMPTY、global 支配短路、scope_id=0 占位排除、dup 去重、GetResidentialArea 传输错误 fail-closed、未知节点 found=false 拒绝、过期(4)/驳回(3) grant 不驱动展开、非 admin 不展开回归、多 grant 多 division 并集：✅
- 断言真实：`assert.Equal(t, tt.wantAllowed, resp.Allowed, ...)` + 拒绝错误码 60007 断言，非「调用不报错」型测试：✅
- 新增字段映射类（AssertPublishScope 接线切换 resolveUserScope→resolvePublishScope；init_permissions.sql 段 6 种子）：无独立逻辑，由 division 测试覆盖接线 + SQL 幂等验证查询兜底：✅

## TDD 证据检查（分诊：字段映射 vs 有逻辑函数）

| 新增/修改函数 | 类型 | 是否有测试 | RED 确认 | GREEN 确认 | 状态 |
|-------------|------|:---:|:---:|:---:|:---:|
| resolvePublishScope（scope.go:90） | 有逻辑 | ✅（经 AssertPublishScope 14 用例间接覆盖） | ✅ CHANGELOG L12-17 摘录 | ✅ 14/14 PASS | PASS |
| holdsCommunityAdminRole（scope.go:136） | 有逻辑 | ✅（非admin/过期/驳回分支覆盖） | ✅ 同上摘录 | ✅ | PASS |
| expandCommunityAdminDivision（scope.go:155） | 有逻辑 | ✅（子树展开/并集/fail-closed 覆盖） | ✅ 同上摘录 | ✅ | PASS |
| AssertPublishScope（assertpublishscopelogic.go:38 接线切换） | 字段映射/纯接线 | ✅ | —（不要求） | ✅ | PASS |
| init_permissions.sql 段 6（种子矩阵） | 字段映射/seed | ✅（段 6.7 幂等验证查询断言精确到码） | —（不要求） | ✅（执行由 Owner 验证） | PASS |

**RED 摘录真实性核验**（CHANGELOG.md L12-17，含具体 error 文本，非仅「看到失败」）：
```
--- FAIL: TestAssertPublishScope_CommunityAdminDivisionExpansion (0.02s)
    --- FAIL: TestAssertPublishScope_CommunityAdminDivisionExpansion/community_admin@100_发同division_101_✅（子树展开生效） (0.02s)
        Not equal: expected: true  actual: false   (Allowed 应为 true)
```
根因描述与 RED 一致（现状 resolveUserScope 不展开 division）。GREEN 现场复跑 14/14 PASS，证据链闭合。

## 发现

| 级别 | 问题 | 建议 |
|------|------|------|
| WARN | CHANGELOG L20 记「15 用例全 PASS」，实际测试文件为 **14** 个 table 用例（L92-202） | 修正 CHANGELOG 用例数描述（不阻塞） |
| WARN | 3 个既有机械检查 WARN（proto_ts_align / design_consistency deleted_at / git_hygiene） | 均非本次引入，由对应服务/前端/治理流程跟进 |
| INFO | init_permissions.sql 段 6 为种子变更，需执行到真库验证（脚本已含段 6.7 幂等验证查询） | 由 Owner 执行验证（CHANGELOG 已声明） |

---
VERDICT: PASS
---

机械检查 18 PASS 0 FAIL 3 WARN（WARN 均既有）；go build/vet/test 全绿（FRESH exit 0）；有逻辑函数 resolvePublishScope/holdsCommunityAdminRole/expandCommunityAdminDivision 均有测试、RED 摘录含具体 error 文本（Not equal: expected: true actual: false）、GREEN 14/14 PASS；字段映射类接线与种子 SQL 由测试 + 幂等验证查询兜底。
