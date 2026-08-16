# QA Report — web/mobile

**验证时间**: 2026-08-17 00:04
**验证范围**: main 分支工作树未提交改动 + 未跟踪文件（本轮最新：【我的】页图标网格重构 + 手机号显示修复 + 退出登录接线）
**验证依据**: 当前工作树 diff（`git diff` + `git status` 未跟踪文件），不引用历史 commit；TDD 证据只看本次工作树新增/修改函数
**QA Agent**: 前端服务 QA Engineer

## 机械化检查结果 (harness-checks-frontend.sh — FRESH run)

```
bash .harness/skills/qa/scripts/harness-checks-frontend.sh --service mobile --json
timestamp: 2026-08-16T15:56:09Z | exit_code: 0 | summary: 6 PASS, 0 FAIL, 2 WARN
```

| # | 检查项 | 结果 | 详情 |
|---|--------|------|------|
| 1 | type_check (vue-tsc) | ✅ | exit 0 — type check passed |
| 2 | unit_test (vitest) | ✅ | exit 0 — 123 tests passed（21 files，0/0 假通过已排除） |
| 3 | build (vite) | ✅ | exit 0 — build succeeded |
| 4 | hardcoded_secrets | ✅ | 0 处密钥/令牌硬编码 |
| 5 | debug_artifacts | ✅ | 0 处 console.log/debugger 残留（console.error/warn 允许） |
| 6 | type_safety (`as any`) | ⚠️ WARN | 3 处 `as any`（aspirational ≤10）：`src/utils/request.ts:91` / `src/utils/crypto.ts:63` / `src/api/identity.ts:133` — **均为既有存量**（已用 `git show HEAD:` 核对 HEAD 各文件同样各 1 处，非本轮引入；新 logout 函数 0 处 as any） |
| 7 | api_field_align | ⚠️ WARN | 34 处 snake_case/camelCase 不匹配，**全部在 `web/pc/`**（.userType/.pageSize/.sortOrder/.realName），mobile 0 处 — 已用 `check-api-field-align.sh` 全量核对 34 条均为 web/pc/ 路径；本轮新增 logout body `{deviceId, kickAllDevices}` 与后端 `auth-service/api/internal/types/types.go` `LogoutReq` JSON tag（`deviceId`/`kickAllDevices`）逐字段一致，无新增 |
| 8 | unit_standard (rem only) | ✅ | rem only（无 rpx/px）— PASS |

## 编译检查（FRESH 复验）
- [x] `npm run build` → **exit 0**，日志末尾 `DONE  Build complete.`（仅有 Sass legacy-js-api deprecation 与 request.ts 动态导入提示，均非错误）

## 静态分析
- [x] `npm run type-check`（vue-tsc --noEmit -p tsconfig.app.json）→ **exit 0**，0 errors，clean output

## 单元测试（FRESH 复验）
- [x] `npm run test:unit` → **exit 0**，**21 test files / 123 tests passed**（0 fail）
- 本轮新增测试与 CHANGELOG 记载一致：`auth-flow.spec.ts`（+2 opts.phone）、`identity.spec.ts`（新建 +2 logout）、`my.spec.ts`（新建 +5 logout+owner-auth）
- 覆盖率：`@vitest/coverage-v8` 未安装，无法生成覆盖率百分比（可选，不阻塞）

## TDD 证据检查（分诊：字段映射 vs 有逻辑函数）

> 分诊口径：有逻辑函数（分支/转换/计算/条件/校验/异步时序）要求 RED→GREEN 实际 FAIL 摘录；字段映射类（纯接线/参数透出/纯机械替换/布局/文档）只要求「有对应测试」或「build/test 绿」、RED 记 —。
> 本轮为【我的】页图标网格重构 + 手机号显示修复 + 退出登录接线（CHANGELOG 顶部条目）。

| 新增/修改函数 | 类型 | 是否有测试 | RED 确认（仅「有逻辑」要求） | GREEN 确认 | 状态 |
|-------------|------|:---:|:---:|:---:|:---:|
| `handleAuthSuccess` 增 opts.phone（auth-flow.ts:32-34，`if (opts.phone)` 写 user_phone） | 有逻辑（条件分支） | ✅ auth-flow.spec.ts:133-161（+2 用例：phone 提供→写入 / 未提供→不写入） | ✅ 具体 FAIL: `expected "setStorageSync" to be called with arguments: [ 'user_phone', '13800001111' ] Number of calls: 0`（CHANGELOG 行 25 内联摘录；结构证明 HEAD 无 auth-flow.ts = 新文件） | ✅ 123/123 全绿 | PASS |
| `logout`（identity.ts:150-155，POST /api/auth/logout body 透出） | 字段映射/纯接线 | ✅ identity.spec.ts（新建 +2 用例：kickAllDevices=false / =true 断言路径+body） | —（不要求） | ✅ 123/123 全绿 | PASS |
| `onLogout`（my.vue:316-333，showModal 确认→logout→clearTokens→reLaunch + 失败分支） | 有逻辑（异步分支） | ✅ my.spec.ts:76-121（+3 用例：确认/取消/失败 toast） | ✅ 具体 FAIL: `wrapper.vm.onLogout is not a function`（CHANGELOG 行 25 内联摘录；结构证明 HEAD my.vue grep onLogout=0 命中） | ✅ my.spec.ts 5/5 全绿 | PASS |
| `onOwnerAuth`（my.vue:241-247，hasOwnerRole→toast 已是业主 / applyForRole('owner')） | 有逻辑（条件分支） | ✅ my.spec.ts:137-158（+2 用例：未认证→applyRole / 已认证→toast） | ✅ 具体 FAIL: `TypeError: wrapper.vm.onOwnerAuth is not a function`（_tdd_evidence.md §23 完整摘录含复现方式；结构证明 HEAD my.vue grep onOwnerAuth=0 命中） | ✅ my.spec.ts 5/5 全绿 | PASS |
| my.vue 布局重构（func-entries 网格） | 字段映射（布局） | ✅ build/test 绿（模板经 my.spec.ts mount 渲染） | —（不要求） | ✅ 123/123 全绿 | PASS |

**RED 证据真实性核对（结构性证明辅助，但以具体 error 文本为准）**：
- `git show HEAD:web/mobile/src/pages/my/my.vue` → grep `onOwnerAuth|onLogout` = **0 命中**（HEAD 无这两个函数）
- `git show HEAD:web/mobile/src/api/identity.ts` → grep `export async function logout` = **0 命中**（logout 为新增）
- `git show HEAD:web/mobile/src/utils/auth-flow.ts` → **fatal: path exists on disk, but not in 'HEAD'**（auth-flow.ts 为未跟踪新文件）
- onOwnerAuth 的 RED 为 _tdd_evidence.md §23 真实 vitest 输出（含 git stash 复现方式与 my.spec.ts:142/152 定位），符合「真实摘录」标准
- handleAuthSuccess-phone 与 onLogout 的 RED 摘录内联于 CHANGELOG 行 25（具体 error 文本），结构证明与摘录一致

**测试边界覆盖评估**：
- handleAuthSuccess opts.phone：正反两分支（提供→写 / 未提供→不写，避免误写空值）✅
- onLogout：成功→logout+clearTokens+reLaunch / 取消→不动作 / 失败→toast「退出登录失败」保持登录态 ✅
- onOwnerAuth：未认证→applyRole({community_id, role_code:'owner'}) / 已有 owner 角色(verf_status=2)→toast「已是业主」不重复申请 ✅
- logout：kickAllDevices 默认 false / 显式 true 两分支 ✅

## 测试质量评估
- 新增/修改有逻辑函数: 3（handleAuthSuccess-phone / onLogout / onOwnerAuth）→ 均有测试 + RED 摘录 ✅
- 新增纯接线/字段映射: 2（logout / 布局重构）→ logout 有测试，布局经 build/test 绿 ✅
- 边界测试: ✅（各函数正反/成功失败/取消路径均覆盖）

## 发现
| 级别 | 问题 | 建议 |
|------|------|------|
| ⚠️ WARNING | type_safety 3 处 `as any`（request.ts:91 / crypto.ts:63 / identity.ts:133），aspirational ≤10 未达标，但为既有存量非本轮引入 | 后续轮次清理 |
| ⚠️ WARNING | api_field_align 34 处不匹配，全部在 web/pc/（.userType/.pageSize/.sortOrder/.realName），mobile 0 处 | pc 端存量，不在 mobile 范围；mobile 侧本轮新增 logout body 与后端 LogoutReq JSON tag 完全一致 |
| ⚠️ WARNING | CHANGELOG 行 25 声称「完整摘录见 _tdd_evidence.md §23」，但 §23 仅持久化 onOwnerAuth 摘录；handleAuthSuccess-phone 与 onLogout 的 RED 摘录仅内联于 CHANGELOG 行 25，未在 _tdd_evidence.md 单列章节 | 建议后续将 phone/onLogout 的 RED 摘录补录至 _tdd_evidence.md（如新增 §24），避免「完整摘录见 §23」引用不完全落位（上轮 onOwnerAuth 即因同类引用悬空被判 FAIL） |

---
VERDICT: PASS
---

**判定依据（every-fresh-run 证据）**：
1. `harness-checks-frontend.sh --service mobile --json` → exit 0，6 PASS / 0 FAIL / 2 WARN（timestamp 2026-08-16T15:56:09Z）
2. `npm run build` → exit 0（DONE Build complete）
3. `npm run type-check` → exit 0（0 errors）
4. `npm run test:unit` → exit 0（21 files / 123 tests，0 fail）
5. TDD 证据：3 个有逻辑函数均有测试 + 具体 RED error 文本（onOwnerAuth 为 §23 真实 vitest 输出；phone/onLogout 为 CHANGELOG 内联具体 error 文本 + HEAD 结构证明）；2 个字段映射类有测试或 build/test 绿；无 RED 缺失
6. 无机械化 FAIL、无编译/类型/测试失败、测试覆盖合理
