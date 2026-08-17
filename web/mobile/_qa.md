# QA Report — web/mobile

**验证时间**: 2026-08-17 14:01（UTC 2026-08-17T06:01Z）
**验证范围**: main 分支工作树未提交改动 + 未跟踪文件（本轮 = 加入流程重构「点加入立即建 membership」+ QA FAIL 修复轮「joinCommunity 条件载荷真实 RED 补录」，14 个 web/mobile 文件）
**验证依据**: 当前工作树 diff（`git diff HEAD` + `git status`），不引用历史 commit；TDD 证据只看本次工作树新增/修改函数
**QA Agent**: 前端服务 QA Engineer

## 机械化检查结果 (harness-checks-frontend.sh --service mobile --json — FRESH run)

```
bash .harness/skills/qa/scripts/harness-checks-frontend.sh --service mobile --json
timestamp: 2026-08-17T05:59:37Z | exit_code: 0 | summary: 6 PASS, 0 FAIL, 2 WARN
```

| # | 检查项 | 结果 | 详情 |
|---|--------|------|------|
| 1 | type_check (vue-tsc) | ✅ | exit 0 — type check passed |
| 2 | unit_test (vitest) | ✅ | exit 0 — **150 tests passed**（24 files，0/0 假通过已排除） |
| 3 | build (vite) | ✅ | exit 0 — build succeeded |
| 4 | hardcoded_secrets | ✅ | 0 处密钥/令牌硬编码 |
| 5 | debug_artifacts | ✅ | 0 处 console.log/debugger 残留（console.error/warn 留痕允许） |
| 6 | type_safety (`as any`) | ⚠️ WARN | 3 处 `as any`（aspirational ≤10）：`src/utils/request.ts:91` / `src/utils/crypto.ts:63` / `src/api/identity.ts:133` — **均为既有存量**（本轮新增生产代码 `grep as any` 0 命中，仅 .spec.ts 测试文件内使用） |
| 7 | api_field_align | ⚠️ WARN | 34 处 snake_case/camelCase 不匹配（WARN 级别，关注 created_at/user_type 等）— 存量，非本轮引入 |
| 8 | unit_standard (rem only) | ✅ | rem only（无 rpx/px）— PASS |

## 编译检查（FRESH 复验）
- [x] `npm run build` → **exit 0**，日志末尾 `DONE  Build complete.`（仅有 Sass legacy-js-api deprecation 与 request.ts 动态导入提示，均非错误）

## 静态分析
- [x] `npm run type-check`（vue-tsc --noEmit -p tsconfig.app.json）→ **exit 0**，0 errors，clean output

## 单元测试（FRESH 复验）
- [x] `npm run test:unit` → **exit 0**，**24 test files / 150 tests passed**（0 fail）
- 5 个本轮改动 spec 文件单独 FRESH 跑 → **26/26 PASS**（user.spec 3 / join-choice.spec 4 / join-community.spec 5 / join-residence.spec 7 / pending-join.spec 7）
- CHANGELOG 门禁基线 144 → 150（+6，join-community +1、join-residence +2、pending-join +2、user +1，join-choice 改 1），与当前工作树一致
- 覆盖率：`@vitest/coverage-v8` 未安装，无法生成百分比（可选，不阻塞）

## TDD 证据检查（分诊：字段映射 vs 有逻辑函数）

> 分诊口径：有逻辑函数（分支/转换/计算/条件/校验/异步时序）要求 RED→GREEN 实际 FAIL 摘录；字段映射类（纯接线/参数透出/状态 ref/布局/文案）只要求「有对应测试」或「build/test 绿」、RED 记 —。
> 本轮为加入流程重构（点「加入」立即建 membership）+ QA FAIL 修复轮补录 `api/user.ts` joinCommunity 条件载荷 RED（`_tdd_evidence.md` §RED 2.5）。RED 摘录真实性核对：`git show HEAD:` 结构证明 HEAD 为旧实现，与摘录 FAIL 一致（见下）。

| 新增/修改函数 | 类型 | 是否有测试 | RED 确认（仅「有逻辑」要求） | GREEN 确认 | 状态 |
|-------------|------|:---:|:---:|:---:|:---:|
| `api/user.ts` `joinCommunity`（building/unit/room/ownership 可选 + 4 个 `if (x != null)` 条件载荷） | 有逻辑（条件载荷） | ✅ user.spec.ts:66-77 用例 3（仅传 communityId → 载荷不含 building/unit/room/ownership；含 `not.toHaveProperty('building')` 等 4 断言） | ✅ 具体 FAIL: `AssertionError: expected { community_id: 'c1', …(4) } to not have property "building"`（§RED 2.5，`1 failed | 2 passed`）。结构证明：HEAD joinCommunity 为 5 必填参、恒发 `{community_id,building,unit,room,ownership}` 全键（未传为 undefined），`not.toHaveProperty` 必然失败 → 摘录真实 | ✅ 3 passed | PASS |
| `join-community.vue` `onSelectArea`（await joinCommunity → addCommunity → savePendingJoin 带 membershipId → navigateTo；maxReached / isJoined 分支 + join 失败 catch toast） | 有逻辑（异步时序 + 分支） | ✅ join-community.spec.ts 5 用例（点击即 join/addCommunity/存 pending/导航 / 已加入 toast 不重复 / 上限警告不调 join / join 失败 toast 不存不导航不 addCommunity） | ✅ 具体 FAIL: `AssertionError: expected "vi.fn()" to be called with arguments: [ 'c1' ]` + `Number of calls: 0`（joinCommunity 0 调用）+ `AssertionError: …to be called with arguments: [ ObjectContaining{…} ]` `Number of calls: 0`（toast 0 调用）（§RED 2）。结构证明：HEAD onSelectArea 仅 savePendingJoin+navigateTo、无 try/catch | ✅ 5 passed | PASS |
| `join-residence.vue` `confirmJoin`（读 pending.membershipId → 缺省 getUserMemberships 回退 → 找不到 toast → bindResidence({membership_id,building,unit,room,is_primary:1}) + applyRole(owner/tenant) → 清 pending → toast+switchTab；失败保留 pending 可重试） | 有逻辑（异步时序 + 条件 + 错误处理） | ✅ join-residence.spec.ts 7 用例（空表单不调 API / 深链 toast / 自有 bindResidence(is_primary:1)+applyRole(owner)+清 pending / 租住 applyRole(tenant) / membershipId 回退 getUserMemberships / 找不到成员 toast 不清 pending / bindResidence 失败不清 pending 不 switchTab） | ✅ 具体 FAIL: `AssertionError: expected "vi.fn()" to be called with arguments: [ { membership_id: 'm1', …(4) } ]` + `Number of calls: 0`（bindResidence 0 调用）+ `…[ { community_id: 'c1', …(1) } ]` `Number of calls: 0`（applyRole 0 调用）+ `expected "vi.fn()" to be called at least once`（getUserMemberships 0 调用）（§RED 3）。结构证明：HEAD confirmJoin 调 joinCommunity（新 spec mock 不提供 → 抛错） | ✅ 7 passed | PASS |
| `join-community.vue` `selectProvince`/`selectCity`（加载后 `step.value = 2/3` 自动级联，移除「下一步」按钮） | 状态推进（字段映射级） | ✅ build/test 绿（spec 触发列表点击覆盖到步骤渲染） | —（不要求，无分支/计算/校验逻辑） | ✅ 全绿 | PASS |
| `join-choice.vue` 页顶文案（「加入小区」→「已加入 XX，请选择下一步」，保留社区名/地址） | 文案映射 | ✅ join-choice.spec.ts 改断言（`.header-title`=已加入 幸福小区 / `.header-sub`=请选择下一步 / `.header-addr`=幸福路1号） | —（不要求）§RED 1 附摘录仅证明「断言新文案时旧文案未命中」（Received「加入进行中」为工作树中间态，已注明基线） | ✅ 4 passed | PASS |
| `pending-join.ts` `PendingJoin.membershipId` 可选字段（join-community 回填、join-residence 优先读取） | 字段映射 | ✅ pending-join.spec.ts +2（membershipId 随 save→read 往返 / 缺省 read 返回对象不含该字段） | —（不要求） | ✅ 7 passed | PASS |
| `join-form.ts` 注释同步（joinFormToPayload 归一化说明，无逻辑改动） | 注释 | — 无逻辑改动 | — | ✅ build/test 绿 | PASS |

**RED 证据真实性核对（结构证明辅助 + 具体 error 文本为准）**：
- §RED 2.5（joinCommunity 条件载荷，本 QA FAIL 修复轮补录）：`AssertionError: expected { community_id: 'c1', …(4) } to not have property "building"`，`Test Files 1 failed (1) / Tests 1 failed | 2 passed (3)`，附 user.spec.ts:73 断言行与复现方式（`git checkout HEAD -- src/api/user.ts` 回退旧 5 必填恒发全字段实现）。结构证明：HEAD 实现对象恒含 `building: undefined` 等 4 键 → `not.toHaveProperty` 必然失败，摘录真实。
- §RED 2（onSelectArea）/ §RED 3（confirmJoin）：均为 vitest 实际输出（`AssertionError` + `Number of calls: 0` + 断言文件行号），结构证明 HEAD 为旧实现。
- 上一轮 QA FAIL 项（joinCommunity RED 缺失）已修复：`_tdd_evidence.md` §RED 2.5 已落盘真实 FAIL 摘录，`_qa.md` 与 CHANGELOG 引用一致、无悬空。

**测试边界覆盖评估**：
- joinCommunity：全字段透传 / Snowflake ID string / 仅 communityId 条件载荷（4 断言）✅
- onSelectArea：成功全链路 / 已加入 toast / 上限警告 / join 失败 toast（失败不污染 store）✅
- confirmJoin：空表单校验 / 深链 / 自有+租住两权属 / membershipId 回退 / 找不到成员 / bindResidence 失败可重试 ✅
- pending-join：往返 / 缺省不含 membershipId / sessionStorage 镜像 / TTL / localStorage 零触碰 ✅

## 测试质量评估
- 新增/修改有逻辑函数: 3（joinCommunity 条件载荷 / onSelectArea / confirmJoin）→ 均有对应测试 + 具体 RED error 文本（_tdd_evidence.md §RED 2.5、§RED 2、§RED 3）✅
- 字段/文案映射: membershipId / join-choice 页顶文案 / selectProvince-selectCity 状态推进 / join-form 注释 → 有测试或 build/test 绿 ✅
- 边界测试: ✅（成功/失败/空值/权属两分支/回退路径均覆盖）

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
1. `harness-checks-frontend.sh --service mobile --json` → exit 0，6 PASS / 0 FAIL / 2 WARN（timestamp 2026-08-17T05:59:37Z，150 tests）
2. `npm run build` → exit 0（DONE Build complete）
3. `npm run type-check` → exit 0（0 errors，clean output）
4. `npm run test:unit` → exit 0（24 files / 150 tests，0 fail）；5 个改动 spec 文件单独跑 26/26 PASS
5. TDD 证据：3 个有逻辑函数（joinCommunity 条件载荷 / onSelectArea / confirmJoin）均有测试 + 具体 RED error 文本（_tdd_evidence.md §RED 2.5 / §RED 2 / §RED 3），结构证明真实；字段映射类有测试或 build/test 绿；上一轮 QA FAIL 项已修复
6. 无机械化 FAIL、无编译/类型/测试失败、测试覆盖合理
