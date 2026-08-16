# QA Report — web/mobile

**验证时间**: 2026-08-16 17:56
**验证范围**: main 分支工作树未提交改动 + 未跟踪文件（mobile-homepage-content-revamp，首页信息架构改造 Task 2.1-2.6）
**验证依据**: 当前工作树 diff（`git diff` + `git status` 未跟踪文件），不引用历史 commit

## 机械化检查结果 (harness-checks-frontend.sh — FRESH run)

```
bash .harness/skills/qa/scripts/harness-checks-frontend.sh --service mobile --json
exit_code: 0  |  summary: 5 PASS, 0 FAIL, 2 WARN
```

| # | 检查项 | 结果 | 详情 |
|---|--------|------|------|
| 1 | type_check (vue-tsc) | ✅ | exit 0 — type check passed |
| 2 | unit_test (vitest) | ✅ | exit 0 — 62 tests passed（10 files，0/0 假通过已排除） |
| 3 | build (vite) | ✅ | exit 0 — build succeeded |
| 4 | hardcoded_secrets | ✅ | 0 处密钥/令牌硬编码 |
| 5 | debug_artifacts | ✅ | 0 处 console.log/debugger（console.error 允许，本次变更无） |
| 6 | type_safety (`as any`) | ⚠️ WARN | 3 处 `as any`（aspirational ≤10）：`src/utils/request.ts:91`、`src/utils/crypto.ts:65`、`src/api/identity.ts:133`（均为既有代码，非本次变更引入） |
| 7 | api_field_align | ⚠️ WARN | 34 处 snake_case/camelCase 不匹配，**全部位于 web/pc/**（`pageSize`/`userType`/`sortOrder` 等）；本次移动端变更 0 违规，新增 `file_id`/`file_type`/`since_days` 已 snake_case 对齐 wire |

**单项 FRESH 门禁复验**：
- `npm run type-check`（vue-tsc --noEmit -p tsconfig.app.json）→ **exit 0**，0 errors，clean output
- `npm run build` → **exit 0**，日志末尾 `DONE  Build complete.`（仅有 Sass legacy-js-api deprecation warning，非错误）
- `npm run test:unit`（vitest）→ **exit 0**，**10 test files / 62 tests passed**（0 fail，Duration 1.48s）

## TDD 证据检查（分诊：字段映射 vs 有逻辑函数）

> 分诊口径：有逻辑函数（分支/转换/计算/条件/校验）要求 RED→GREEN 实际 FAIL 摘录；字段映射类（struct 加字段/参数透出/纯接线/seed）只要求测试+绿色、RED 记 —。
> RED 摘录标准：必须为实际 vitest FAIL 输出文本（`AssertionError:` / `expected...Received...` / `Number of calls: 0` 等）。已逐一核对 `_tdd_evidence.md` §1-§8 摘录中引用的文件:行号与当前工作树源码一致（community.spec.ts:10、contact-list.spec.ts:75、notice-detail.spec.ts:62/136、contact-list.vue:75）。

| 新增/修改函数 | 类型 | 是否有测试 | RED 确认（仅「有逻辑」要求） | GREEN 确认 | 状态 |
|-------------|------|:---:|:---:|:---:|:---:|
| `isImageAttachment(fileType?)`（community.ts） | 有逻辑（谓词/分支） | ✅ `community.spec.ts` 3 用例 | ✅ `_tdd_evidence.md` §1 `TypeError: isImageAttachment is not a function` | ✅ 62/62 全绿 | PASS |
| `onFuncEntry(entry)`（notice.vue） | 有逻辑（分支：target→navigateTo / 无→toast） | ✅ `notice.spec.ts` 4 用例 | ✅ §2 `expected +0 to be 4` / `TypeError: Cannot read properties of undefined (reading 'trigger')`（HEAD 无 func-entries） | ✅ 全绿 | PASS |
| `onAttachmentClick(att)`（notice-detail.vue） | 有逻辑（图片/文档/空 url/无法识别分支） | ✅ `notice-detail.spec.ts` 8 用例 | ✅ §3 `expected "vi.fn()" to be called with arguments: [ObjectContaining{…}] Number of calls: 0` | ✅ 全绿 | PASS |
| `fetchDetail(id)`（notice-detail.vue，loadError 失败态分支） | 有逻辑（错误分支） | ✅ `notice-detail.spec.ts` 失败态/不存在用例 | ✅ §4 `expected '通知不存在' to contain '加载失败'` | ✅ 全绿 | PASS |
| notice-browse `onMounted` fetch（since_days=30 + loadError） | 有逻辑（窗口传参 + 错误分支） | ✅ `notice-browse.spec.ts` 5 用例 | ✅ §5 `expected "vi.fn()" to be called with arguments: ['c1', 1, 50, 30]`（HEAD 无窗口传参） | ✅ 全绿 | PASS |
| notice-browse `formatTime`（dayjs.unix + !ts 空守卫） | 有逻辑（转换/条件） | ✅ `notice-browse.spec.ts` L80-81 时间文本断言（本轮已修复零断言 FAIL） | ✅ §6 `expected +0 to be 2` | ✅ 全绿 | PASS |
| contact-list `fetchContacts(communityId)`（新增页） | 有逻辑（fetch + 错误分支） | ✅ `contact-list.spec.ts` 渲染/失败用例 | ✅ §7 `ReferenceError: fetchContacts is not defined` | ✅ 全绿 | PASS |
| contact-list `onCall(contact)`（phone 空守卫 + makePhoneCall） | 有逻辑（条件/校验） | ✅ `contact-list.spec.ts` 拨号用例 | ✅ §8 `Number of calls: 0`（makePhoneCall 实现缺失） | ✅ 全绿 | PASS |
| `getNoticeList` 增 `sinceDays?` 参数（仅 >0 注入 params.since_days） | 字段映射/纯接线（参数透出） | ✅ 经 notice.spec L238 `['c1',1,3,30]` / notice-browse.spec `['c1',1,50,30]` 断言 | —（不要求；§2/§5 同参数 RED 已覆盖） | ✅ 全绿 | PASS |
| `NoticeAttachment` 扩展 `file_id?`/`file_type?` | 字段映射（接口加可选字段） | ✅ build/test 绿，snake_case 对齐 wire | —（不要求） | ✅ 全绿 | PASS |
| pages.json 注册 `pages/contact-list/contact-list` | 纯接线（路由声明） | ✅ 页面经组件测试可达 | —（不要求） | ✅ 全绿 | PASS |
| vitest.setup.ts 增补 previewImage/downloadFile/openDocument/makePhoneCall stub | 测试基础设施 | ✅ 被附件/拨号用例消费 | —（不要求） | ✅ 全绿 | PASS |

## 结构性 RED 佐证（本轮独立复核）
- `git show HEAD:web/mobile/src/api/community.ts` → 无 `isImageAttachment`/`IMAGE_FILE_TYPES`/`since_days`
- `git show HEAD:web/mobile/src/pages/notice/notice.vue` → 无 `FUNCTION_ENTRIES`/`onFuncEntry`/`func-entries`；HEAD `loadAll` 仍含 `fetchContacts`（3 并发）
- `git show HEAD:.../contact-list.vue` → 不存在（新页面）
- `git show HEAD:.../notice-browse.vue` → 含 `threeMonthsAgo` 客户端 3 个月过滤（本轮移除）
- `git show HEAD:.../notice-detail.vue` → 含 H5 `window.open` 分支（本轮移除，改白名单分发）

## 实现一致性核验（本轮独立复核）
- **Task 2.1** community.ts：`NoticeAttachment.file_id?/file_type?`（optional-safe）、`getNoticeList(cid, page, pageSize, sinceDays?)` → `params.since_days`（仅 >0）、`IMAGE_FILE_TYPES` + `isImageAttachment` —— 与 REQ-NDP-2/4 一致 ✅
- **Task 2.2** notice.vue：`getNoticeList(cid, 1, 3, 30)`；4 功能入口固定顺序（便民联络 navigateTo contact-list / 其余 toast「功能开发中」）；未加入小区不渲染入口区；已移除内嵌联络网格 —— REQ-NTW-1 / REQ-FE-1/2/3 一致 ✅
- **Task 2.3** 区块全序：通知 → 4入口 → 邻里互助占位 → 寻失互助 → 底部广告集中区（3 个垂直堆叠）—— REQ-HL-1/2/3/4 一致 ✅
- **Task 2.4** notice-browse：`getNoticeList(cid, 1, 50, 30)` 单请求卡片列表；移除 threeMonthsAgo；点卡片 navigateTo；空态/失败态 —— REQ-NTW-4/5 一致 ✅
- **Task 2.5** notice-detail：附件按 `isImageAttachment(file_type)` 分发（图片 previewImage / 文档 downloadFile+openDocument / 空 url 明确 toast）；不直连 file-service REST；失败/不存在明确态 —— REQ-NDP-1/2/3 一致 ✅
- **Task 2.6** contact-list：`getContacts(cid)` 拨号网格 + makePhoneCall + 空态/失败态；pages.json 注册 —— REQ-CLP-1 一致 ✅

## 测试质量评估
- 新增/修改「有逻辑函数」8 个全部有测试（62/62 全绿），断言为真实行为断言（调用参数、渲染文本、uni 调用、toast）
- 测试覆盖：附件分支、拨号、失败态、空态、窗口参数透出、区块全序均有精确断言
- 过程证据：8 个有逻辑函数 RED 摘录全部为真实 vitest FAIL 文本（`_tdd_evidence.md` §1-§8），行号与当前源码一致；`git stash` 回退复现为真实机制

## 发现
| 级别 | 问题 | 状态 |
|------|------|------|
| WARN | 3 处 `as any`（request.ts:91 / crypto.ts:65 / identity.ts:133） | 既有代码，非本轮引入；aspirational 目标 ≤10，可接受 |
| WARN | api_field_align 34 处不匹配全部位于 web/pc/ | PC 端存量问题，不在本服务范围 |
| INFO | 测试用例与生产代码同工作树、未提交（Generator 直改主树设计） | 按既有约定，新增测试文件须与生产代码一起提交 |

## 前一轮 FAIL 闭环确认
上一轮唯一 FAIL（notice-browse formatTime 零断言）已闭环：`notice-browse.spec.ts` L80-81 已新增时间文本断言 `expect(cards[i].text()).toContain(dayjs.unix(...).format('YYYY-MM-DD HH:mm'))`，RED §6 有真实 `expected +0 to be 2` 摘录。本轮 FRESH 全量复验 62/62 全绿。

---
VERDICT: PASS
---
