# QA Report — permission-service

**验证时间**: 2026-08-17 12:05
**验证范围**: 工作树未提交改动 + 未跟踪文件（敏感权限 min_verf_level=2 加固：init_permissions.sql 段 6.8 + migration/004 + 4 个新测试）
**分支**: main（工作树 diff）
**检查方式**: every-fresh-run（harness-checks.sh --json / go build / go vet / go test -count=1 / RED 复现均本次现场执行）

**工作树变更范围（本次 QA 判定对象）**:
- `M scripts/init_permissions.sql`（段 6.8 加固 UPDATE + 幂等验证 SELECT）
- `?? migration/004_privileged_role_min_verf_level.sql`（既有库幂等加固）
- `?? rpc/internal/logic/permission/seed_min_verf_level_test.go`（迷你 SQL 解析/模拟 + 种子断言 + 审计回归）
- `?? rpc/internal/logic/permission/privileged_roles_min_verf_level_test.go`（12 用例行为判定）
- `M CHANGELOG.md`（2026-08-17 条目 + TDD 摘录）
- 无生产 Go 逻辑变更（CheckPermission 的 min_verf_level 执行路径为既有，本次仅数据加固）

## 机械化检查结果 (harness-checks.sh — FRESH run)

| # | 检查项 | 结果 | 详情 |
|---|--------|------|------|
| 1 | go build | ✅ | exit 0（独立复跑 `go build ./...` → BUILD_EXIT=0） |
| 2 | go vet | ✅ | exit 0（独立复跑 `go vet ./...` → VET_EXIT=0，clean output） |
| 3 | go test | ✅ | 4 包通过，186 测试函数（api 29 + model 32 + rpc 125），0 FAIL（`go test ./... -count=1` → TEST_EXIT=0） |
| 4 | Proto int64 jstype | ✅ | diff 无 proto 变更（skipped），0 violations |
| 5 | json:",string" | ✅ | AST 检查全部 int64 ID 字段合规，0 violations |
| 6 | 跨服务DB导入 | ✅ | 2 Go 文件 diff 扫描，0 violations |
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
- [x] go test ./... -count=1（4/4 包通过，0 FAIL，TEST_EXIT=0；全包 186 测试函数，permission 包 125）
- 新增 4 个测试 FRESH -v 复跑全 PASS：
  - `TestSeedSensitivePermissions_HardenedToLevel2`（6 敏感码从零建库后有效层级=2）✅
  - `TestMigrationPrivilegedRoleMinVerfLevel_HardenedToLevel2`（迁移 004 覆盖同 6 码）✅
  - `TestSeedPrivilegedRoles_DestructiveWritePerms_HardenedToLevel2`（角色 2/3/4 破坏性写权限必须 level-2 审计）✅
  - `TestCheckPermission_PrivilegedRoleDestructiveOps_NeedVerf`（12 用例：6 码 × 未认证 status=0 拒绝 / 已认证 status=2 放行）✅

## 测试覆盖
| 包 | 覆盖率 | 状态 |
|----|--------|------|
| rpc/internal/logic/permission | 92.1% | 既有执行逻辑高覆盖；本次变更主体为种子/迁移 SQL + 测试代码 |
| model | 70.7% | 既有 |
| api/internal/logic/perm | 43.2% | 既有 |
| api/internal/types | 31.2% | 既有 |

> 本次无新增生产 Go 逻辑（数据驱动加固），覆盖率非核心判据；行为判定经 TestCheckPermission 12 用例直接验证既有 CheckPermission 执行路径。

## 测试质量评估
- 新增函数: 有逻辑的均位于测试代码内（`simulateEffectiveMinVerfLevel` 迷你 SQL 模拟/解析器、`parseRolePermBindings`/`parsePermissionDefs`/`isWriteMethod` 审计解析）；无新增生产逻辑函数
- 断言真实（非「调用不报错」型）：
  - 种子测试：`assert.Equal(t, 2, eff[code], ...)` + `assert.Contains`（防加固幻影）✅
  - 行为测试：`assert.Equal(t, want, resp.Allowed, ...)`，12 用例含未认证拒绝/已认证放行双向断言 ✅
  - 审计回归：断言到具体权限码 + path + 角色绑定（防未来新增破坏性写权限漏设门槛）✅
- 边界/防御：从零建库执行序（INSERT IGNORE 默认 0 → UPDATE 覆盖）、幂等可重跑、注释剥离（`--` 整行/行内）、VALUES 元组分隔 `), (`、INSERT 前缀大小写归一——均覆盖 ✅
- 字段映射类（init_permissions.sql 段 6.8 / migration/004）：由种子结构测试 + 幂等验证 SELECT 兜底 ✅

## TDD 证据检查（分诊：字段映射 vs 有逻辑函数）

| 新增/修改函数 | 类型 | 是否有测试 | RED 确认（仅「有逻辑」要求） | GREEN 确认 | 状态 |
|-------------|------|:---:|:---:|:---:|:---:|
| init_permissions.sql 段 6.8（UPDATE + 幂等 SELECT） | 字段映射/seed | ✅（TestSeedSensitivePermissions） | —（不要求） | ✅ PASS（模拟有效层级=2） | PASS |
| migration/004_privileged_role_min_verf_level.sql | 字段映射/seed | ✅（TestMigrationPrivilegedRoleMinVerfLevel） | —（不要求） | ✅ PASS | PASS |
| simulateEffectiveMinVerfLevel 及解析器（stripSQLComments/splitStatements/applyUpdateMinLevel/applyInsertDefault/extractValueField 等） | 有逻辑 | ✅（种子 3 测试驱动） | ✅ 现场复现 FAIL | ✅ PASS | PASS |
| parseRolePermBindings / parsePermissionDefs / isWriteMethod（审计解析） | 有逻辑 | ✅（TestSeedPrivilegedRoles_DestructiveWritePerms） | ✅ 现场复现 FAIL | ✅ PASS | PASS |
| TestCheckPermission_PrivilegedRoleDestructiveOps_NeedVerf（12 用例行为判定） | 有逻辑（分支/条件，回归验证既有 CheckPermission 执行路径；本次无新增生产逻辑） | ✅（自身即测试） | ✅ 现场复现 FAIL | ✅ 12/12 PASS | PASS |

**RED 摘录真实性核验（本次现场独立复现，非仅引用 CHANGELOG）**：

1. 种子 RED — 将 HEAD 版 init_permissions.sql（无段 6.8）与新增测试同置临时树运行（api-proto/common 子模块内容复制入临时树），6 个敏感码全部 `expected: 2 actual: 0`：
```
Error: Not equal: expected: 2  actual: 0
Messages: 权限码 community:notice:delete-api 从零建库后有效 min_verf_level 应为 2（需已认证）；当前=0。检查加固 UPDATE 是否放置于 427/428 的 INSERT IGNORE 之后（4.3.2 早于段 6.4 会被默认 0 覆盖）
...（community:notice:update-api / community:activity:create-api / role:read / role:read:list-api / role:read:detail-api 共 6 码均 expected: 2 actual: 0）
--- FAIL: TestSeedSensitivePermissions_HardenedToLevel2 (0.00s)
```
与 CHANGELOG L12-17 摘录一致；GREEN 现场 4 测试全 PASS，证据链闭合。

2. 审计 RED — 同一 HEAD 树，`TestSeedPrivilegedRoles_DestructiveWritePerms_HardenedToLevel2` 复现失败：
```
Error: Not equal: expected: 2  actual: 0
Messages: 角色 3 绑定的破坏性写权限 community:activity:create-api(POST:/api/community/activities) 必须 min_verf_level=2（未认证 grant 不得行使）
```

3. 行为 RED — 临时树将 6 个权限 MinVerfLevel 置 0（模拟加固前数据态）运行 TestCheckPermission，6 个「未认证status0_拒绝」子用例全部失败（即未认证 grant 被放行，正是本变更修复的安全洞）：
```
Error: Not equal: expected: false  actual: true
Messages: [网格员删公告 427] status=0 min_verf_level=0 → allowed 应=false
...（428/432/210/211/212 共 6 码同构）
--- FAIL: TestCheckPermission_PrivilegedRoleDestructiveOps_NeedVerf
```

## 发现

| 级别 | 问题 | 建议 |
|------|------|------|
| WARN | 3 个既有机械检查 WARN（proto_ts_align / design_consistency deleted_at / git_hygiene） | 均非本次引入，由对应服务/前端/治理流程跟进 |
| WARN | 行为测试 TestCheckPermission 依赖 mock 硬编码 MinVerfLevel=2，不读种子数据 | 种子真实值一致性由 TestSeed* 系列保证（同一 6 码集合 + 审计），可加跨测试码集一致性断言（不阻塞） |
| INFO | init_permissions.sql 段 6.8 / migration/004 为 DB 变更，需执行到真库验证 | 由 Owner 执行验证（CHANGELOG 已声明「执行由 Owner 验证」；migration-must-execute） |

---
VERDICT: PASS
---
机械检查 18 PASS 0 FAIL 3 WARN（WARN 均既有）；go build/vet/test 全绿（FRESH exit 0，186 测试函数 0 FAIL）；字段映射类（种子段 6.8 + 迁移 004）由种子结构测试 + 幂等验证 SELECT 兜底；有逻辑测试函数（迷你 SQL 模拟器 + 审计解析 + 行为判定 12 用例）RED 均现场独立复现（expected:2 actual:0 / expected:false actual:true 具体 error 摘录）、GREEN 全 PASS，证据链闭合。
