# QA Report — web/mobile

**验证时间**: 2026-08-17 11:20
**验证范围**: main 分支工作树未提交改动 + 未跟踪文件（本轮最新：加入小区流程改造 — pending-join 契约模块 + join-choice/join-residence 新页 + join-community 选小区分流 + my.vue applyForRole pending 回退；含工作树游离的 agreement.vue 注册超时 10002/timeout 分流）
**验证依据**: 当前工作树 diff（`git diff` + `git status` 未跟踪文件），不引用历史 commit；TDD 证据只看本次工作树新增/修改函数
**QA Agent**: 前端服务 QA Engineer

## 机械化检查结果 (harness-checks-frontend.sh --service mobile --json — FRESH run)

```
bash .harness/skills/qa/scripts/harness-checks-frontend.sh --service mobile --json
timestamp: 2026-08-17T03:17:55Z | exit_code: 0 | summary: 6 PASS, 0 FAIL, 2 WARN
```

| # | 检查项 | 结果 | 详情 |
|---|--------|------|------|
| 1 | type_check (vue-tsc) | ✅ | exit 0 — type check passed |
| 2 | unit_test (vitest) | ✅ | exit 0 — 144 tests passed（24 files，0/0 假通过已排除） |
| 3 | build (vite) | ✅ | exit 0 — build succeeded |
| 4 | hardcoded_secrets | ✅ | 0 处密钥/令牌硬编码 |
| 5 | debug_artifacts | ✅ | 0 处 console.log/debugger 残留（console.error/warn 留痕允许） |
| 6 | type_safety (`as any`) | ⚠️ WARN | 3 处 `as any`（aspirational ≤10）：`src/utils/request.ts:91` / `src/utils/crypto.ts:63` / `src/api/identity.ts:133` — **均为既有存量**（本轮新增生产代码 grep `as any` 0 命中；新文件 pending-join.ts / join-choice.vue / join-residence.vue 的 `as any` 仅在 .spec.ts 测试文件内，非生产代码） |
| 7 | api_field_align | ⚠️ WARN | 34 处 snake_case/camelCase 不匹配（WARN 级别，关注 created_at/user_type 等）— 存量，非本轮引入 |
| 8 | unit_standard (rem only) | ✅ | rem only（无 rpx/px）— PASS（unit-standard-gate.spec.ts 单独 FRESH 跑 PASS） |

## 编译检查（FRESH 复验）
- [x] `npm run build` → **exit 0**，日志末尾 `DONE  Build complete.`（仅有 Sass legacy-js-api deprecation 与 request.ts 动态导入提示，均非错误）

## 静态分析
- [x] `npm run type-check`（vue-tsc --noEmit -p tsconfig.app.json）→ **exit 0**，0 errors，clean output

## 单元测试（FRESH 复验）
- [x] `npm run test:unit` → **exit 0**，**24 test files / 144 tests passed**（0 fail）
- 7 个本轮改动 spec 文件单独跑 → **43/43 PASS**（pending-join 5 / join-choice 4 / join-residence 5 / community 10 / my 8 / join-community 4 / agreement 7）
- 基线 21 files / 125 tests → +3 文件 / +19 用例，与 CHANGELOG 门禁一致
- 覆盖率：`@vitest/coverage-v8` 未安装，无法生成覆盖率百分比（可选，不阻塞）

## TDD 证据检查（分诊：字段映射 vs 有逻辑函数）

> 分诊口径：有逻辑函数（分支/转换/计算/条件/校验/异步时序）要求 RED→GREEN 实际 FAIL 摘录；字段映射类（纯接线/参数透出/状态 ref/布局）只要求「有对应测试」或「build/test 绿」、RED 记 —。
> 本轮为加入小区流程改造 + agreement.vue 10002/timeout 分流（CHANGELOG 顶部两条目）。

| 新增/修改函数 | 类型 | 是否有测试 | RED 确认（仅「有逻辑」要求） | GREEN 确认 | 状态 |
|-------------|------|:---:|:---:|:---:|:---:|
| `pending-join.ts` `savePendingJoin`/`readPendingJoin`/`clearPendingJoin`（内存态主载体 + H5 sessionStorage 镜像 TTL 30min + validFrom 过期判定） | 有逻辑 | ✅ pending-join.spec.ts +5（save→read 往返 / sessionStorage 镜像带 expiresAt 且 localStorage 零触碰 / TTL 过期返 null+清镜像 / clear 双清 / 空数据 null） | ✅ 具体 FAIL: `Error: Failed to resolve import "./pending-join" ... Does the file exist?`（_tdd_evidence.md 本轮 §1；结构证明 HEAD 无此文件） | ✅ 5 passed | PASS |
| `community.ts` `pendingCommunityId` ref | 字段映射 | ✅ community.spec.ts 初始空串断言（`expected undefined to be ''`） | ✅ §2 RED 含 `AssertionError: expected undefined to be ''`（实现前 undefined） | ✅ 10 passed | PASS |
| `community.ts` `setPendingCommunityId` | 字段映射/接线 | ✅ community.spec.ts +1 | ✅ 具体 FAIL: `TypeError: store.setPendingCommunityId is not a function`（§2，community.spec.ts:83） | ✅ 10 passed | PASS |
| `community.ts` `clearPendingCommunityId` | 字段映射/接线 | ✅ community.spec.ts +1 | ✅ 同 §2（实现前 not a function） | ✅ 10 passed | PASS |
| `join-community.vue` `onSelectArea`（maxReached / isJoined 分支 + savePendingJoin + navigateTo join-choice，删除 modal） | 有逻辑分支 | ✅ join-community.spec.ts 重写 4 用例（存 pending+导航 / modal 不存在 / 已加入不存不导航 / 上限警告不存不导航） | ✅ 具体 FAIL: `AssertionError: expected "vi.fn()" to be called with arguments: [ { communityId: 'c1', …(2) } ]` + `expected true to be false`（.join-form-mask 残留）（§3） | ✅ 4 passed | PASS |
| `join-choice.vue` `goResidence`/`goOtherAuth`/`goBack`（导航分支 + setPendingCommunityId + switchTab） | 导航分支（有逻辑） | ✅ join-choice.spec.ts 4 用例（显示小区名 / 业主→navigateTo join-residence / 其他→pendingCommunityId=c1+switchTab my / 深链空态不渲染卡片） | ✅ 具体 FAIL: `Error: Failed to resolve import "./join-choice.vue" ... Does the file exist?`（§4；结构证明 HEAD 无此页） | ✅ 4 passed | PASS |
| `join-residence.vue` `confirmJoin`（validateJoinForm 校验→joinCommunity→applyRole 失败不阻塞→addCommunity→clearPendingJoin→toast+switchTab） | 有逻辑 | ✅ join-residence.spec.ts 5 用例（空表单校验错误不调 API / 深链 toast 不调 API / OWNED=1 全链路断言 / RENTED=2 ownership / joinCommunity 失败不 addCommunity 不 applyRole） | ✅ 具体 FAIL: `Error: Failed to resolve import "./join-residence.vue"`（§5a）+ `Cannot call trigger on an empty DOMWrapper`（§5b 选择器对齐，模板补 confirm-join-btn class，纯选择器无逻辑改动） | ✅ 5 passed | PASS |
| `my.vue` `applyForRole`（current 优先 + pending 回退一次性消费 + 双空 toast） | 有逻辑分支 | ✅ my.spec.ts +3（pending 回退并一次性清除 / current 优先不消费 pending / 双空 toast「请先加入小区」不调 API） | ✅ 具体 FAIL: `AssertionError: expected "vi.fn()" to be called with arguments: [ { community_id: 'c9', …(1) } ]`（§6） | ✅ 8 passed | PASS |
| `agreement.vue` `confirmRegister` catch（10002 清数据回登录 / timeout 提示可重试保留数据 / 其他兜底 三分支） | 有逻辑 | ✅ agreement.spec.ts +2（10002→toast 该手机号已注册+clearRegPending+navigateBack+不调 handleAuthSuccess / timeout→toast 可能已创建+保留数据+不调 handleAuthSuccess） | ✅ 具体 FAIL: `AssertionError: expected "vi.fn()" to be called with arguments: [ ObjectContaining{…} ]`，Received `{ icon:'none', title:'注册失败，请重试' }`，agreement.spec.ts:202/223（§24，git stash 回退 HEAD 复现真实 FAIL，结构证明 HEAD catch 无 10002/timeout 分支） | ✅ 7 passed | PASS |
| pages.json 注册 join-choice / join-residence | 字段映射（路由声明） | — build/test 绿 | —（不要求） | ✅ 144/144 全绿 | PASS |

**RED 证据真实性核对（结构性证明辅助，但以具体 error 文本为准）**：
- 全部有逻辑函数 RED 均为 vitest 实际输出（`AssertionError` / `TypeError` / `Failed to resolve import`），持久化于 `_tdd_evidence.md` §24（agreement 10002/timeout）与本轮新章节 §1-§6，每节含复现方式。
- 结构性证明：`git show HEAD:` — pending-join.ts / join-choice.vue / join-residence.vue 均不在 HEAD（未跟踪新文件）；HEAD 的 join-community.vue 无 savePendingJoin（旧实现含 .join-form-mask modal）；HEAD community.ts 无 setPendingCommunityId；HEAD my.vue applyForRole 无 pending 回退；HEAD agreement.vue catch 仅统一「注册失败，请重试」。
- 上一轮 QA 记录的「CHANGELOG 声称 §23 引用悬空」类问题本轮已修复：§24 已将 agreement 10002/timeout 的完整 RED 摘录落盘到 `_tdd_evidence.md`，引用真实存在。

**测试边界覆盖评估**：
- pending-join：往返 / TTL 过期 / 空数据 / localStorage 零触碰（三 spy 断言）✅
- join-residence confirmJoin：空表单校验 / 深链 / 成功（自有+租住两权属）/ 失败不污染 store ✅
- join-community onSelectArea：正常分流 / 已加入 / 上限警告 / modal 死代码不存在 ✅
- my applyForRole：current 优先 / pending 回退消费 / 双空 ✅
- agreement 分支：10002 / timeout / 其他兜底（既有）✅

## 测试质量评估
- 新增/修改有逻辑函数: 8（pending-join 契约 / community set+clear / onSelectArea / join-choice 导航 / confirmJoin / applyForRole / confirmRegister 分支）→ 均有测试 + 具体 RED 摘录 ✅
- 新增字段映射: pendingCommunityId ref + pages.json 路由 → 有测试或 build/test 绿 ✅
- 边界测试: ✅（各函数正反/成功失败/空值/权属两分支均覆盖）

## 发现
| 级别 | 问题 | 建议 |
|------|------|------|
| ⚠️ WARNING | type_safety 3 处 `as any`（request.ts:91 / crypto.ts:63 / identity.ts:133），aspirational ≤10 未达标，但为既有存量非本轮引入 | 后续轮次清理 |
| ⚠️ WARNING | api_field_align 34 处 snake_case/camelCase 不匹配，存量非本轮引入 | 持续关注 created_at/user_type |
| INFO | join-residence.vue catch 参数 `(e: any)` 为错误对象类型标注（非断言），不构成 type_safety 违规 | 无需处理 |

---
VERDICT: PASS
---

**判定依据（every-fresh-run 证据）**：
1. `harness-checks-frontend.sh --service mobile --json` → exit 0，6 PASS / 0 FAIL / 2 WARN（timestamp 2026-08-17T03:17:55Z）
2. `npm run build` → exit 0（DONE Build complete）
3. `npm run type-check` → exit 0（0 errors）
4. `npm run test:unit` → exit 0（24 files / 144 tests，0 fail）；7 个改动 spec 文件 43/43 PASS；unit-standard-gate.spec.ts PASS
5. TDD 证据：8 个有逻辑函数均有测试 + 具体 RED error 文本（_tdd_evidence.md §1-§6、§24）；字段映射类有测试或 build/test 绿；无 RED 缺失
6. 无机械化 FAIL、无编译/类型/测试失败、测试覆盖合理
