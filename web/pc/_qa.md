# QA Report — web/pc

**验证时间**: 2026-06-23 08:46
**验证范围**: master branch (commit dde46b5)

## 机械化检查结果 (harness-checks-frontend.sh — FRESH run)

| # | 检查项 | 结果 | 详情 |
|---|--------|------|------|
| 1 | TypeScript type check | ✅ | type check passed |
| 2 | Unit tests | ✅ | 52 tests passed (5 test files) |
| 3 | Production build | ❌ | **164 TypeScript errors** — TS1294: erasableSyntaxOnly conflicts with export enum |
| 4 | Hardcoded secrets | ✅ | no secrets detected in source |
| 5 | Debug artifacts | ✅ | no debug artifacts in production code |
| 6 | TypeScript type safety | ⚠️ | 67 'as any' usages (aspirational target ≤10) |

**Machine Check Summary**: 4 PASS, 1 FAIL, 1 WARN

---

## 编译检查 (FRESH)

- [ ] `npm run build` — **FAILED (exit 1)**

**Evidence**:
```bash
$ cd /home/jiaoxh/my-project/community-and-home/web/pc && npm run build
> vue-tsc -b && vite build

../common/types/identity.ts(18,13): error TS1294: This syntax is not allowed when 'erasableSyntaxOnly' is enabled.
../common/types/identity.ts(23,13): error TS1294: This syntax is not allowed when 'erasableSyntaxOnly' is enabled.
...
[164 total errors]
```

**Root Cause**: `tsconfig.app.json` has `"erasableSyntaxOnly": true`, which disallows `export enum` declarations. The project has 7+ enums in `web/common/types/identity.ts` (UserType, UserStatus, VerificationStatus, RoleStatus, PermissionType, PermissionStatus, HomeownerVerificationStatus).

**Critical Detail**: 
- ✅ `npm run type-check` (vue-tsc --noEmit) passes — no emit phase, no erasableSyntaxOnly check
- ❌ `npm run build` (vue-tsc -b) fails — incremental build triggers erasableSyntaxOnly enforcement

This is a **known memory pattern**: see `.harness/knowledge/memory/typescript-erasable-syntax-enum-conflict.md`

---

## 静态分析 (FRESH)

- [x] `npm run type-check` — **PASSED (exit 0)**

**Evidence**:
```bash
$ cd /home/jiaoxh/my-project/community-and-home/web/pc && npm run type-check
> vue-tsc --noEmit
[clean output, exit 0]
```

---

## 单元测试 (FRESH)

- [x] `npm run test:unit` — **PASSED (52/52 tests, 5 test files)**

**Evidence**:
```bash
$ cd /home/jiaoxh/my-project/community-and-home/web/pc && npm run test:unit
> vitest

 Test Files  5 passed (5)
      Tests  52 passed (52)
   Start at  08:46:41
   Duration  1.56s (transform 837ms, setup 456ms, import 3.63s, tests 215ms, environment 1.64s)
```

**Test Files**:
1. `tests/unit/directives/permission.spec.ts`
2. `tests/unit/stores/auth.spec.ts`
3. `tests/unit/stores/permission.spec.ts`
4. `tests/unit/utils/request.spec.ts`
5. `tests/unit/views/aimodel/ModelForm.spec.ts`

**Status**: All tests pass, no 0/0 fake passes detected.

---

## 测试覆盖

Not executed (blocked by build failure).

---

## 测试质量评估

Cannot evaluate new functions (no recent code changes in web/pc/ working tree — all changes are in `.harness/` and service submodules).

**Recent commits affecting web/pc**:
- 4d43466 feat(web-pc): RBAC 管理界面前端实现 (committed 2026-06-19)

---

## 其他质量指标

### Type Safety (WARN)

67 `as any` usages detected (aspirational target ≤10):
- `web/pc/src/stores/division.ts:135`
- `web/pc/src/utils/permission.ts:44`
- `web/pc/src/utils/request.ts:93, 108, 114, 211`
- `web/pc/src/utils/crypto.ts:74`
- `web/pc/src/views/division/Index.vue:156, 164, 312, 355, 368, 382`
- (57 more locations)

**Recommendation**: Gradual refactoring to replace `as any` with proper types.

---

## 发现

| 级别 | 问题 | 建议 |
|------|------|------|
| ❌ **BLOCKING** | Production build fails with 164 TS1294 errors due to `erasableSyntaxOnly: true` + `export enum` conflict | **Solution 1** (1 min): Remove `"erasableSyntaxOnly": true` from `tsconfig.app.json` and `tsconfig.node.json`<br>**Solution 2** (2h): Convert all `export enum` to `export const enum`<br>**Solution 3** (4h): Refactor to `as const` objects<br>**Recommended**: Solution 1 (see `.harness/knowledge/memory/typescript-erasable-syntax-enum-conflict.md`) |
| ⚠️ WARNING | 67 `as any` usages exceed aspirational target (≤10) | Gradual refactoring (non-blocking) |

---

## Root Cause Analysis

### Why This Happened

1. **Upstream change**: `tsconfig.app.json` enabled `"erasableSyntaxOnly": true` (likely from Vue 3 official template or a recent TS upgrade)
2. **Existing codebase**: Project has 7+ `export enum` declarations in shared types (`web/common/types/identity.ts`)
3. **Silent failure mode**: `npm run type-check` (used in dev) passes because `--noEmit` skips the erasable check
4. **Late detection**: Build phase (`vue-tsc -b`) triggers the check, but this only runs in CI or manual build

### Why type-check Passed But Build Failed

| Command | Tool | Behavior | erasableSyntaxOnly Check |
|---------|------|----------|-------------------------|
| `npm run type-check` | `vue-tsc --noEmit` | Type check only, no code generation | ❌ Not triggered |
| `npm run build` | `vue-tsc -b` | Incremental build, emits declaration files | ✅ Triggered |

**Key insight**: The check only fires during the **emit phase** (when generating .js/.d.ts files), which `--noEmit` skips by design.

---

## Memory Update Required

**Existing memory**: `.harness/knowledge/memory/typescript-erasable-syntax-enum-conflict.md` already documents this exact failure mode.

**Action**: ✅ No new memory file needed (this is a known pattern).

**Verification**: The memory file correctly predicted:
1. ✅ `npm run type-check` passes
2. ✅ `npm run build` fails
3. ✅ Error code TS1294
4. ✅ Root cause is `erasableSyntaxOnly` + `export enum`

---

## VERDICT: FAIL

**Reason**: Production build failure (exit 1) — blocks deployment.

**Blocking Issue**: 164 TypeScript TS1294 errors prevent generating deployable static assets.

**Next Steps**:
1. Remove `"erasableSyntaxOnly": true` from `web/pc/tsconfig.app.json` and `web/pc/tsconfig.node.json`
2. Re-run `npm run build` to verify fix
3. Update `.harness/skills/qa/scripts/harness-checks-frontend.sh` to ensure `build` check runs **before** deployment gate

**Non-blocking Issues** (address in future refactoring):
- 67 `as any` usages (should gradually reduce to ≤10)
