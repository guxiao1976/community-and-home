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
