# QA Report — web/mobile (frontend)

**验证时间**: 2026-08-12 19:04 (+08) / 11:04 UTC
**验证范围**: 工作树未提交改动 + 未跟踪文件（加入小区流程携带 ownership — 阶段⑤ T5.1，含 TDD RED 证据修复轮）
**QA Agent**: QA Engineer (frontend)

## 机械化检查结果 (harness-checks-frontend.sh — FRESH run)

命令: `bash .harness/skills/qa/scripts/harness-checks-frontend.sh --service mobile --json`
exit_code = 0, summary: pass 5 / fail 0 / warn 2

| # | 检查项 | 结果 | 详情 |
|---|--------|------|------|
| 1 | type_check (vue-tsc) | ✅ | type check passed |
| 2 | unit_test (vitest) | ✅ | 17 tests passed (3 files)，非假通过（非 0/0） |
| 3 | build (uni build) | ✅ | build succeeded |
| 4 | hardcoded_secrets | ✅ | no secrets detected in source |
| 5 | debug_artifacts | ✅ | no debug artifacts in production code |
| 6 | type_safety | ⚠️ | WARN — 3 'as any'（目标 ≤10）：`web/mobile/src/utils/request.ts:91`（既有，拦截器解包 data）; `web/mobile/src/utils/crypto.ts:65`（既有，RSA 密钥调试）; `web/mobile/src/api/identity.ts:133`（本次新增，`as unknown as any` 防御性断言，axios 类型 AxiosResponse 与已解包拦截器不符，CHANGELOG 已说明） |
| 7 | api_field_align | ⚠️ | WARN — 34 处 snake_case/camelCase 不匹配（WARN 级别，既有全仓问题，非本变更引入） |

> 注：脚本为**前端服务**检查项（7 项：type-check/unit-test/build/hardcoded-secrets/debug-artifacts/type-safety/api-field-align），非 Go 服务的 8 项，故上表用前端实际检查项呈现。

## 编译检查（FRESH）
- [x] `npm run build` → exit 0（DONE Build complete；仅 Dart Sass legacy-js-api DEPRECATION WARNING，非阻断）

## 静态分析（FRESH）
- [x] `npm run type-check`（vue-tsc --noEmit -p tsconfig.app.json）→ exit 0，clean output 无任何报错（vue-tsc 1.0.24，moduleResolution=node 兼容）

## 单元测试（FRESH）
- [x] `npm run test:unit`（vitest v4.1.10）→ exit 0
  - Test Files 3 passed (3)
  - Tests 17 passed (17)
  - 无 skipped / failed
- 非假通过（非 0/0）：实际 3 文件 / 17 测试

## 测试覆盖（针对本次变更的新增/修改函数）

| 函数/模块 | 测试文件 | 测试数 | 覆盖 |
|-----------|---------|:---:|------|
| `joinCommunity`（签名 1→5 参 + 载荷） | `src/api/user.spec.ts` | 2 | OWNED/RENTED + Snowflake string ID 载荷断言（toHaveBeenCalledWith） |
| `validateJoinForm`（新增） | `src/pages/join-community/join-form.spec.ts` | 8 | 正常/权属空/权属非法/楼号边界/单元边界/房号边界/全空多错误路径 |
| `joinFormToPayload`（新增） | `src/pages/join-community/join-form.spec.ts` | 1 | string→number 转换 |
| `OWNERSHIP_OPTIONS`（新增） | `src/pages/join-community/join-form.spec.ts` | 1 | OWNED=1 / RENTED=2 |
| `confirmJoin`/`openJoinForm`/`closeJoinForm`（新增+修改） | `src/pages/join-community/join-community.spec.ts` | 5 | 打开表单/未选权属阻止提交/自有传1/租住传2/成功卡片 |
| `getUserProfile`（仅类型断言修改） | 无 | — | 纯类型修复（`as unknown as any`），无行为变更，可接受豁免 |

## 测试质量评估
- 新增/修改函数: 8 / 有测试: 7（joinCommunity、validateJoinForm、joinFormToPayload、OWNERSHIP_OPTIONS、confirmJoin、openJoinForm、closeJoinForm）/ 缺失: 1（getUserProfile 为纯类型断言修改，无行为变更，属可接受豁免）
- 测试有真实断言（expect.toHaveBeenCalledWith / expect(r.valid).toBe / expect(errors.x).toBeTruthy），非"仅调用不报错" ✅
- 边界覆盖: ✅ 权属 null/0/1/2、楼号 0/151/12、单元 0/6/4、房号 30/100/999、全空表单多错误同时上报

## TDD 证据检查

| 新增/修改函数 | 是否有测试 | RED 确认（含 FAIL 输出摘录） | GREEN 确认 | 状态 |
|-------------|:---:|:---:|:---:|:---:|
| joinCommunity | ✅ | ✅ `_tdd_evidence.md` §1：真实 `AssertionError: expected "vi.fn()" to be called with arguments: [ '/api/users/communities/join', …(1) ]`，Received diff 缺 building/unit/room/ownership（user.spec.ts:42:26 / :57:26） | ✅ 17/17 PASS | PASS |
| validateJoinForm | ✅ | ✅ `_tdd_evidence.md` §2：真实 `Error: Failed to resolve import "./join-form" from "src/pages/join-community/join-form.spec.ts". Does the file exist?`（join-form.spec.ts:3:71） | ✅ 17/17 PASS | PASS |
| joinFormToPayload | ✅ | ✅ 同上 §2（同一模块导入失败覆盖） | ✅ 17/17 PASS | PASS |
| OWNERSHIP_OPTIONS | ✅ | ✅ 同上 §2 | ✅ 17/17 PASS | PASS |
| confirmJoin / openJoinForm / closeJoinForm | ✅ | ✅ `_tdd_evidence.md` §3/§4：真实 `Error: Failed to resolve import "./join-form" from "join-community.vue". Does the file exist?` + `AssertionError: expected "vi.fn()" to be called with arguments: [ 'c1', 3, 1, 502, 1 ]`（join-community.spec.ts:82:27） | ✅ 17/17 PASS | PASS |

- **RED 摘录核验（QA 侧只读复核，2026-08-12 FRESH）**：
  - 摘录为 vitest 实际输出格式（`AssertionError` 含 `Received:`/`-`/`+` diff；`Failed to resolve import`），非注释/口头描述 ✅
  - 行号对应当前测试文件真实断言行：`user.spec.ts:42:26` / `:57:26` 均为 `expect(request.post).toHaveBeenCalledWith(...)`；`join-community.spec.ts:82:27` 为 `expect(joinCommunity).toHaveBeenCalledWith('c1', 3, 1, 502, 1)` ✅
  - 结构性证明一致：`git show HEAD:web/mobile/src/api/user.ts` → 旧签名 `joinCommunity(communityId: string)` 仅传 `{community_id}`；`git show HEAD:.../join-form.ts`、`git show HEAD:.../user.spec.ts` → `fatal: path exists on disk, but not in 'HEAD'`（新文件/新签名在 HEAD 中不存在）✅
  - 无残留旧引用：`grep joinSelectedCommunity src/` → none（旧函数已完全改名 openJoinForm）✅
- **判定**：RED 列已有具体 FAIL 输出摘录（含 `AssertionError` / `Failed to resolve import`），满足 must-follow memory `tdd-red-evidence-requires-fail-excerpt.md` → ✅ → **TDD 证据 PASS**

## 发现

| 级别 | 问题 | 建议 |
|------|------|------|
| WARN | type_safety：3 处 `as any`（identity.ts:133 为本变更新增） | 优先消除本变更新增的 identity.ts:133；剩余两处为既有，可排期治理 |
| WARN | api_field_align：34 处 snake_case/camelCase 不匹配 | 既有全仓问题，非本变更引入，建议单独治理 |
| INFO | Sass legacy-js-api DEPRECATION WARNING（build 时多条） | 非阻断，工具链升级时可一并处理 |
| INFO | vitest 提示 CJS build of Vite Node API deprecated + configLoader native 警告 | 非阻断；后续可将 vitest.config.ts 改 .mjs 或 package.json 加 `"type": "module"` |
| INFO | 新增 `uni.hideLoading()` 于 `confirmJoin` 开头 | 防御性清理残留 loading，行为合理，无问题 |

---
VERDICT: **PASS** — 机械化检查无 FAIL（5 PASS / 2 WARN，exit 0）+ build / type-check / test:unit 全通过（均 FRESH exit 0，17/17 测试非假通过）+ 新增/修改函数测试覆盖合理（8 函数 7 有测试，1 为纯类型断言豁免）+ TDD RED 证据已补真实 FAIL 输出摘录（`_tdd_evidence.md`，摘录为 vitest 实际输出且与当前测试文件行号一致，结构性 HEAD 证明成立）。
---
