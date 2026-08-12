# TDD 证据 — web/mobile 阶段⑤ T5.1 加入小区流程携带 ownership

> 生成时间: 2026-08-12T19:01:00+08:00
> 复现方式: 测试文件（`src/api/user.spec.ts` / `src/pages/join-community/join-form.spec.ts` / `src/pages/join-community/join-community.spec.ts`）已在工作树（首次 Generator 交付时即写好，全部为 HEAD 不存在的未跟踪文件）。RED 摘录通过**临时回退实现到 HEAD 基线**后运行 vitest 捕获真实输出，随后恢复实现。
> 证明: `git show HEAD:web/mobile/src/api/user.ts` → 旧签名 `joinCommunity(communityId: string)` 仅传 `{community_id}`；`git show HEAD:web/mobile/src/pages/join-community/join-form.ts` → `fatal: path exists on disk, but not in 'HEAD'`（不存在）。RED 为真实失败，摘录为 vitest 实际输出（含 `AssertionError` / `Failed to resolve import`），非注释/口头描述。

## 1. RED — joinCommunity 载荷（src/api/user.spec.ts）

复现方式: 临时将 `src/api/user.ts` 的 `joinCommunity` 回退到 HEAD 1 参签名 `joinCommunity(communityId)`（载荷仅 `{community_id}`），测试保持 5 参断言不动，运行 `npx vitest run src/api/user.spec.ts`。

```
 FAIL  src/api/user.spec.ts > joinCommunity > POSTs community_id/building/unit/room/ownership to /api/users/communities/join
AssertionError: expected "vi.fn()" to be called with arguments: [ '/api/users/communities/join', …(1) ]

Received:

  1st vi.fn() call:

  [
    "/api/users/communities/join",
    {
-     "building": 1,
      "community_id": "c1",
-     "ownership": 1,
-     "room": 301,
-     "unit": 2,
    },
  ]


Number of calls: 1

 ❯ src/api/user.spec.ts:42:26
     40|     const membership = await joinCommunity('c1', 1, 2, 301, 1);
     41|
     42|     expect(request.post).toHaveBeenCalledWith('/api/users/communities/…
       |                          ^

 FAIL  src/api/user.spec.ts > joinCommunity > sends ownership=2 (RENTED) and keeps Snowflake community_id as string
AssertionError: expected "vi.fn()" to be called with arguments: [ '/api/users/communities/join', …(1) ]

Received:

  1st vi.fn() call:

  [
    "/api/users/communities/join",
    {
-     "building": 3,
      "community_id": "1234567890123456789",
-     "ownership": 2,
-     "room": 502,
-     "unit": 1,
    },
  ]

Number of calls: 1
 ❯ src/api/user.spec.ts:57:26

 Test Files  1 failed (1)
      Tests  2 failed (2)
```

## 2. RED — join-form 模块（src/pages/join-community/join-form.spec.ts）

复现方式: 临时移除新增文件 `src/pages/join-community/join-form.ts`（HEAD 中本就不存在），测试保持导入不动，运行 `npx vitest run src/pages/join-community/join-form.spec.ts`。覆盖 `validateJoinForm` / `joinFormToPayload` / `OWNERSHIP_OPTIONS` 三个导出——模块不存在则导入即失败。

```
 ❯ src/pages/join-community/join-form.spec.ts (0 test)

 FAIL  src/pages/join-community/join-form.spec.ts [ src/pages/join-community/join-form.spec.ts ]
Error: Failed to resolve import "./join-form" from "src/pages/join-community/join-form.spec.ts". Does the file exist?
  Plugin: vite:import-analysis
  File: /home/jiaoxh/my-project/community-and-home/web/mobile/src/pages/join-community/join-form.spec.ts:3:71
  1  |  // Unit tests for the join-community form validation + payload mapping.
  2  |  import { describe, it, expect } from "vitest";
  3  |  import { OWNERSHIP_OPTIONS, validateJoinForm, joinFormToPayload } from "./join-form";
     |                                                                          ^

 Test Files  1 failed (1)
      Tests  no tests
```

## 3. RED — 组件挂载失败（src/pages/join-community/join-community.spec.ts，join-form.ts 缺失）

复现方式: 同 2，`join-community.vue` 在 script setup 中 `import { OWNERSHIP_OPTIONS, validateJoinForm, joinFormToPayload } from './join-form'`，模块缺失导致组件无法编译挂载。覆盖 `confirmJoin` / `openJoinForm` / `closeJoinForm`。

```
 FAIL  src/pages/join-community/join-community.spec.ts [ src/pages/join-community/join-community.spec.ts ]
Error: Failed to resolve import "./join-form" from "src/pages/join-community/join-community.vue". Does the file exist?

 Test Files  1 failed (1)
      Tests  no tests
```

## 4. RED — 组件行为断言（src/pages/join-community/join-community.spec.ts，confirmJoin 回退 HEAD 行为）

复现方式: 恢复 `join-form.ts`，临时把 `join-community.vue` 的 `confirmJoin` 回退到 HEAD 旧流程（`joinSelectedCommunity` 直接 `await joinCommunity(target.id)`，不做权属/楼号/单元/房号校验与载荷映射），运行 `npx vitest run src/pages/join-community/join-community.spec.ts`。捕获行为型断言失败。

```
 FAIL  src/pages/join-community/join-community.spec.ts > join-community page — ownership join form > blocks confirm when ownership is not selected
AssertionError: expected "vi.fn()" to not be called at all, but actually been called 1 times

 FAIL  src/pages/join-community/join-community.spec.ts > join-community page — ownership join form > calls joinCommunity(communityId, building, unit, room, OWNED) for 自有 join
AssertionError: expected "vi.fn()" to be called with arguments: [ 'c1', 3, 1, 502, 1 ]

Received:

  1st vi.fn() call:

  [
    "c1",
-   3,
-   1,
-   502,
-   1,
  ]

Number of calls: 1
 ❯ src/pages/join-community/join-community.spec.ts:82:27
     80|     await flushPromises();
     81|
     82|     expect(joinCommunity).toHaveBeenCalledWith('c1', 3, 1, 502, 1);
       |                           ^

 FAIL  src/pages/join-community/join-community.spec.ts > join-community page — ownership join form > calls joinCommunity with ownership=2 (RENTED) for 租住 join
AssertionError: expected "vi.fn()" to be called with arguments: [ 'c1', 3, 1, 502, 2 ]

Received:

  1st vi.fn() call:

  [
    "c1",
-   3,
-   1,
-   502,
-   2,
  ]

Number of calls: 1
 ❯ src/pages/join-community/join-community.spec.ts:93:27
     91|     await flushPromises();
     92|
     93|     expect(joinCommunity).toHaveBeenCalledWith('c1', 3, 1, 502, 2);
       |                           ^

 Test Files  1 failed (1)
      Tests  3 failed | 2 passed (5)
```

## 5. GREEN — 恢复实现后全量验证

恢复 `joinCommunity` 5 参签名、`join-form.ts`、`confirmJoin` 完整流程后：

```
命令: npm run test:unit  →  Test Files 3 passed (3) / Tests 17 passed (17)
命令: npm run type-check →  vue-tsc --noEmit -p tsconfig.app.json → exit 0（clean）
命令: npm run build      →  DONE Build complete（exit 0）
```
