# TDD 证据 — web/mobile access-control（Task 7.1-7.3）TDD 缺口修复

> 生成时间: 2026-08-13
> 背景: 上一轮 QA 判 FAIL——`switchCommunity` / `onCommunitySwitch` / `request.ts` 响应拦截器 / `validateJoinForm` 被误标为「字段映射/纯接线」，漏了测试与 RED 证据。本轮补齐分诊与 RED 摘录。
> 复现方式: 新增 3 个测试文件（`src/stores/community.spec.ts` / `src/pages/notice/notice.spec.ts` / `src/utils/request.spec.ts`），并复用既有 `join-form.spec.ts`。RED 摘录通过**临时回退对应实现文件到 HEAD 基线**后运行 vitest 捕获真实输出（含 `AssertionError` / `TypeError` / `expected`/`Received`），随后恢复实现。摘录均为 vitest 实际输出，非注释/口头描述。

---

## 1. RED — switchCommunity（src/stores/community.spec.ts）

复现方式: 临时将 `src/stores/community.ts` 回退到 HEAD 基线（`switchCommunity` 为同步、不调后端 `setCurrentCommunity`、不抛错），测试保持 async 语义断言不动，运行 `npx vitest run src/stores/community.spec.ts`。

```
 FAIL  src/stores/community.spec.ts > community store — switchCommunity > persists to backend then updates local state + storage on success
AssertionError: expected "vi.fn()" to be called with arguments: [ 'c2' ]
Number of calls: 0
 ❯ src/stores/community.spec.ts:39:33
     37|     await store.switchCommunity('c2');
     38|
     39|     expect(setCurrentCommunity).toHaveBeenCalledWith('c2');
       |                                 ^

 FAIL  src/stores/community.spec.ts > community store — switchCommunity > keeps currentCommunityId unchanged when backend rejects (e.g. 10015 out of scope)
TypeError: You must provide a Promise to expect() when using .rejects, not 'undefined'.
 ❯ src/stores/community.spec.ts:54:46
     52|     );
     53|
     54|     await expect(store.switchCommunity('c2')).rejects.toMatchObject({ …
       |                                              ^

 Test Files  1 failed (1)
      Tests  2 failed | 1 passed (3)
```

说明: HEAD 的 `switchCommunity` 是同步纯本地操作（不调后端、不抛错），故 ①`setCurrentCommunity` 调用数为 0；②`.rejects` 收到 `undefined`（非 Promise）。

---

## 2. RED — onCommunitySwitch（src/pages/notice/notice.spec.ts）

复现方式: 临时将 `src/pages/notice/notice.vue` 回退到 HEAD 基线（`onCommunitySwitch` 为同步、无 try/catch 10015 分支），保持 `stores/community.ts`（工作树 async 版）不动，运行 `npx vitest run src/pages/notice/notice.spec.ts`。

```
 FAIL  src/pages/notice/notice.spec.ts > notice page — onCommunitySwitch 10015 branch > shows a specific toast when switch fails with code 10015
AssertionError: expected "vi.fn()" to be called with arguments: [ ObjectContaining{…} ]
Number of calls: 0
     65|     expect(uni.showToast).toHaveBeenCalledWith(
       |                                ^

 Test Files  1 failed (1)
      Tests  1 failed | 2 passed (3)
     Errors  2 errors
```

说明: HEAD 的 `onCommunitySwitch` 无 try/catch，`switchCommunity` 抛出的 `{code:10015}` 成为未处理拒绝（Unhandled Rejection），`uni.showToast` 调用数为 0 → 10015 专属提示缺失。

---

## 3. RED — request.ts 响应拦截器（src/utils/request.spec.ts）

复现方式: 临时将 `src/utils/request.ts` 回退到 HEAD 基线（业务错误 `Promise.reject(new Error(...))` 不附加 `code` 字段），测试保持 `.code` 断言不动，运行 `npx vitest run src/utils/request.spec.ts`。

```
 FAIL  src/utils/request.spec.ts > request response interceptor > rejects with an Error carrying the business code for business errors
AssertionError: expected undefined to be 10015 // Object.is equality
- Expected:
+ Received:
     60|     expect(err).toBeInstanceOf(Error);
     61|     expect(err.code).toBe(10015);
     62|     expect(err.message).toBe('目标小区不在数据范围');
       |                                  ^

 Test Files  1 failed (1)
      Tests  1 failed | 4 passed (5)
```

说明: HEAD 拒绝的 `Error` 未附加业务 `code`（`err.code` 为 `undefined`），调用方（`onCommunitySwitch`）无法按 `10015` 分支。

---

## 4. RED — validateJoinForm（src/pages/join-community/join-form.spec.ts）

复现方式: 临时将 `src/pages/join-community/join-form.ts` 回退到 HEAD 基线（硬编码区间 `room 100-999`），测试保持「去硬编码区间」断言（`room:'30'` 应放行）不动，运行 `npx vitest run src/pages/join-community/join-form.spec.ts`。

```
 FAIL  src/pages/join-community/join-form.spec.ts > validateJoinForm > rejects room <= 0 (positive required, no hardcoded 3-digit range)
AssertionError: expected false to be true // Object.is equality
- Expected
+ Received

 Test Files  1 failed (1)
      Tests  1 failed | 9 passed (10)
```

说明: HEAD 硬编码 `room >= 100`，`room:'30'` 被误判非法（`valid=false`），而「去硬编码区间」后应放行（`valid=true`）。

---

## 5. GREEN — 恢复实现后全量验证

恢复 4 个实现文件到工作树版本后：

```
命令: npx vitest run           →  Test Files 6 passed (6) / Tests 28 passed (28)
命令: npm run type-check       →  vue-tsc --noEmit -p tsconfig.app.json → exit 0（clean）
命令: npm run build            →  DONE Build complete（exit 0）
```

---

## mobile-homepage-content-revamp（Task 2.1-2.6）TDD RED 证据

> 生成时间: 2026-08-16
> 复现方式: `git stash push` 回退对应生产文件到 HEAD 基线（contact-list 为新增文件用函数改名法），运行对应 spec 捕获真实 vitest FAIL 输出，随后恢复。所有摘录均为 vitest 实际输出，非注释/口头描述。
> 基线验证: 恢复后 `npm run build` / `npm run type-check` / `npm run test:unit`（62 tests）全部 exit 0。

### 1. RED — isImageAttachment（src/api/community.ts，新增谓词）

复现: `git stash push -- src/api/community.ts` 后运行 `npx vitest run src/api/community.spec.ts`。

```
 FAIL  src/api/community.spec.ts > isImageAttachment — file_type 图片白名单分发（REQ-NDP-2） > 白名单小写扩展名 png/jpg/jpeg/gif → true
TypeError: isImageAttachment is not a function
 ❯ src/api/community.spec.ts:10:12
     10|     expect(isImageAttachment('png')).toBe(true);
       |            ^

 FAIL  ... > 文档扩展名 pdf/doc/docx → false（走文档分支 REQ-NDP-3）
TypeError: isImageAttachment is not a function
 ❯ src/api/community.spec.ts:17:12

 FAIL  ... > 缺失/无法识别 file_type（undefined/空串/未知）→ false 不崩溃
TypeError: isImageAttachment is not a function
 ❯ src/api/community.spec.ts:23:12
```

### 2. RED — onFuncEntry / onCommunitySwitch / 通知区静默 catch（src/pages/notice/notice.vue）

复现: `git stash push -- src/pages/notice/notice.vue` 后运行 `npx vitest run src/pages/notice/notice.spec.ts`。

```
 FAIL  src/pages/notice/notice.spec.ts > notice page — onCommunitySwitch 10015 branch > does not toast for a non-10015 switch error
AssertionError: expected "vi.fn()" to not be called at all, but actually been called 2 times
Number of calls: 2
 ❯ src/pages/notice/notice.spec.ts:95:31

 FAIL  ... > fetch 静默 catch 消除（REQ-P1-ERR） > 双请求并发全部失败 → toast ≥1 次 + console.error 恰好 2 次
AssertionError: expected "error" to be called 2 times, but got 3 times
 ❯ src/pages/notice/notice.spec.ts:161:20

 FAIL  ... > 4 功能图标入口（REQ-FE-1/2/3） > 4 个入口按固定顺序渲染：便民联络/物业报修/二手闲置/租房卖房
AssertionError: expected +0 to be 4 // Object.is equality
+ Received
 ❯ src/pages/notice/notice.spec.ts:266:28

 FAIL  ... > 4 功能图标入口（REQ-FE-1/2/3） > 点击便民联络 → uni.navigateTo 到 contact-list（做实跳页）
TypeError: Cannot read properties of undefined (reading 'trigger')
 ❯ src/pages/notice/notice.spec.ts:276:17

 FAIL  ... > 首页通知 30 天窗口参数（REQ-NTW-1/2） > 首页通知以 since_days=30 & page_size=3 调用 getNoticeList
AssertionError: expected "vi.fn()" to be called with arguments: [ 'c1', 1, 3, 30 ]
Number of calls: 1
 ❯ src/pages/notice/notice.spec.ts:242:27
```

> 注：HEAD 无 `.func-entry` 元素/`onFuncEntry`/`FUNCTION_ENTRIES`（grep 0 命中），且 HEAD `loadAll` 仍含 `fetchContacts`（3 并发）——故「4 入口渲染」期望 4 实得 0、「点便民联络」`navigateTo` 元素不存在、`getNoticeList` 收到 HEAD 旧签名 `(cid,1,3)` 缺 `30` 窗口参数。onFuncEntry 分支（target→navigateTo / 无→toast）在 HEAD 下均未实现 → 三类断言全部 FAIL。

### 3. RED — onAttachmentClick（src/pages/notice-detail/notice-detail.vue）

复现: `git stash push -- src/pages/notice-detail/notice-detail.vue` 后运行 `npx vitest run src/pages/notice-detail/notice-detail.spec.ts`。

```
 FAIL  ... > 附件点击图片 → previewImage（file_type ∈ 白名单）
AssertionError: expected "vi.fn()" to be called with arguments: [ ObjectContaining{…} ]
Number of calls: 0
 ❯ src/pages/notice-detail/notice-detail.spec.ts:62:30
```

### 4. RED — fetchDetail loadError 失败态分支（notice-detail.vue）

```
 FAIL  ... > 详情加载失败 → 明确失败态 + console.error
AssertionError: expected '通知不存在' to contain '加载失败'
Expected: "加载失败"
Received: "通知不存在"
 ❯ src/pages/notice-detail/notice-detail.spec.ts:136:28
```

### 5. RED — notice-browse onMounted（since_days=30 窗口传参，src/pages/notice-browse/notice-browse.vue）

复现: `git stash push -- src/pages/notice-browse/notice-browse.vue` 后运行 `npx vitest run src/pages/notice-browse/notice-browse.spec.ts`。

```
 FAIL  ... > 通知卡片列表渲染
AssertionError: expected "vi.fn()" to be called with arguments: [ 'c1', 1, 50, 30 ]
Received: 1st vi.fn() call:  <老实现无 since_days 传参>
```

### 6. RED — notice-browse formatTime（dayjs 转换 + !ts 空守卫）

```
 FAIL  ... > 时间格式化为 YYYY-MM-DD HH:mm
AssertionError: expected +0 to be 2 // Object.is equality
- Expected
+ Received
```

### 7. RED — contact-list fetchContacts（src/pages/contact-list/contact-list.vue，新增页）

复现: 临时将 `fetchContacts` 改名（新文件无法 stash）后运行 `npx vitest run src/pages/contact-list/contact-list.spec.ts`，随后恢复。

```
ReferenceError: fetchContacts is not defined
 ❯ src/pages/contact-list/contact-list.vue:75:5
     73|   const cid = communityStore.currentCommunityId;
     74|   if (cid) {
     75|     fetchContacts(cid);
       |     ^
```

### 8. RED — contact-list onCall（phone 空守卫 + makePhoneCall）

复现: 临时注释 `onCall` 内 `uni.makePhoneCall` 调用（模拟实现缺失）后运行 `npx vitest run src/pages/contact-list/contact-list.spec.ts`，随后恢复。

```
 FAIL  src/pages/contact-list/contact-list.spec.ts > contact-list — 联络拨号网格（REQ-CLP-1） > 点击联络卡片 → uni.makePhoneCall 拨号
AssertionError: expected "vi.fn()" to be called with arguments: [ ObjectContaining{…} ]
Number of calls: 0
 ❯ src/pages/contact-list/contact-list.spec.ts:75:31

 Test Files  1 failed (1)
      Tests  1 failed | 3 passed (4)
```

> 注：`onCall` 的 phone 空守卫（`if (contact.phone)`）为条件分支；实现缺失（无 `makePhoneCall` 调用）时拨号断言收到 `Number of calls: 0`（GREEN 后恢复 → 62/62 全绿）。
