# QA Report — web/pc

**验证时间**: 2026-08-14 14:29 UTC
**验证范围**: 当前工作树未提交改动 + 未跟踪文件（`web/pc/src/views/roles/List.vue` 角色列表列宽重排 + `web/pc/stats.html` 构建产物）
**变更来源**: `.harness/changes/role-platforms-save/specs/role-list-column-layout/spec.md`（REQ-LAYOUT-1/2）

## 机械化检查结果 (harness-checks-frontend.sh — FRESH run)

| # | 检查项 | 结果 | 详情 |
|---|--------|------|------|
| 1 | type_check | ✅ | `pc: type check passed` |
| 2 | unit_test | ✅ | 52 tests passed（5 个测试文件，无 0/0 假通过） |
| 3 | build | ✅ | `pc: build succeeded` |
| 4 | hardcoded_secrets | ✅ | no secrets detected in source |
| 5 | debug_artifacts | ✅ | no debug artifacts in production code |
| 6 | type_safety | ⚠️ | 71 'as any' usages（aspirational target ≤10）— 全为历史存量，不在本次变更文件中 |
| 7 | api_field_align | ⚠️ | 34 处 snake_case/camelCase 不匹配（WARN级别，关注 created_at/user_type 等）— 历史存量 |

**Machine Check Summary**: 5 PASS, 2 WARN, 0 FAIL（exit code 0）

---

## 编译检查 (FRESH)

- [x] `npm run build` — **PASSED (exit 0)**

**Evidence**:
```bash
$ cd /home/jiaoxh/my-project/community-and-home/web/pc && npm run build
> vue-tsc -b && vite build
✓ 1952 modules transformed.
✓ built in 3.82s
BUILD_EXIT_CODE=0
```

注意：构建仅有 Sass `@import`/`darken()` deprecation 警告与 chunk >1000kB 警告（`INVALID_ANNOTATION` 来自 node_modules/@vueuse），均为既有噪音，非本次变更引入，不阻塞。

---

## 静态分析 (FRESH)

- [x] `npm run type-check` — **PASSED (exit 0)，输出 clean**

**Evidence**:
```bash
$ cd /home/jiaoxh/my-project/community-and-home/web/pc && npm run type-check
> vue-tsc -b
TYPECHECK_EXIT_CODE=0
```

---

## 单元测试 (FRESH)

- [x] `npx vitest run --reporter=verbose` — **PASSED (52/52 tests, 5 test files)**

**Evidence**:
```bash
$ cd /home/jiaoxh/my-project/community-and-home/web/pc && npx vitest run --reporter=verbose
 Test Files  5 passed (5)
      Tests  52 passed (52)
   Start at  22:29:13
   Duration  1.40s
TEST_EXIT_CODE=0
```

**Test Files**:
1. `tests/unit/directives/permission.spec.ts`
2. `tests/unit/stores/auth.spec.ts`
3. `tests/unit/stores/permission.spec.ts`
4. `tests/unit/utils/request.spec.ts`
5. `tests/unit/views/aimodel/ModelForm.spec.ts`

**Status**: 全部通过，无 0/0 假通过。

---

## TDD 证据检查（分诊：字段映射 vs 有逻辑函数）

| 新增/修改函数 | 类型 | 是否有测试 | RED 确认（仅「有逻辑」要求） | GREEN 确认 | 状态 |
|-------------|------|:---:|:---:|:---:|:---:|
| —（无新增/修改函数，纯模板属性变更） | 字段映射/纯布局 | — | —（不要求） | ✅ build+test 绿 | PASS |

**分诊结论**：本次变更 `web/pc/src/views/roles/List.vue` 的 diff **全部为 `<el-table-column>` 模板属性调整**（ID 列宽 200→70、name/code/description 增加 `min-width` + `show-overflow-tooltip`、操作列宽 380→260）。`<script setup>` 无任何改动（经 grep 校验：无 `function/const/if/for/return` 逻辑变更）。**属于「字段映射/纯接线」类别（布局/展示层），无独立业务逻辑，RED 列记 —（不要求）**。GREEN 证据为 FRESH build exit 0 + 52 单元测试全绿，满足「有对应测试 或 build/test 绿」门槛。

---

## 测试覆盖

未单独执行 `--coverage`（不阻塞，仅记录）。本次变更为纯模板属性调整，无新增逻辑函数，测试覆盖风险极低。

## 测试质量评估

- 新增函数: 0 / 有测试: — / 缺失: —
- 无逻辑函数新增，`role:update/permission/delete` 等按钮、`loadRoles`、分页交互均未改动（REQ-LAYOUT-2 无回退：功能按钮与分页在 diff 中零改动，build/type-check/test 全绿佐证）

## 发现

| 级别 | 问题 | 建议 |
|------|------|------|
| ⚠️ WARNING | 71 处 `as any`（目标 ≤10）：division.ts:136、permission.ts:44、request.ts:99/114/120/186/244、crypto.ts:75 等 | 历史存量，非本次变更引入；建议后续渐进重构 |
| ⚠️ WARNING | api_field_align 34 处 snake_case/camelCase 不匹配（关注 created_at/user_type 等） | 历史存量；`List.vue` 中 `prop="created_at"` 属 Element Plus 列绑定既有写法，未新增 |
| ℹ️ INFO | `web/pc/stats.html`（构建产物）被构建流程重新生成 | 非源码，无需评审 |

---

## VERDICT: PASS

**依据（全为 FRESH 运行证据）**：
- `harness-checks-frontend.sh --service pc --json` → exit 0，5 PASS / 2 WARN / 0 FAIL
- `npm run build` → exit 0（vue-tsc -b + vite build 成功）
- `npm run type-check` → exit 0，输出 clean
- `npx vitest run` → exit 0，5 测试文件 / 52 测试全绿
- 变更范围确认为纯模板列宽布局调整，无逻辑函数，无 TDD RED 要求；REQ-LAYOUT-1/2 功能不变量（按钮/分页）零改动
- 2 项 WARN 均为历史存量，非本次变更引入，不阻塞

---
