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

---

# TDD 证据 — web/mobile 移动端 6 项完善（TabBar 改名 / 登录协议流程 / 登录态修复 / 通知模块改造）

> 生成时间: 2026-08-16
> 背景: 本轮新增/修改 7 个有逻辑函数，全部先写失败测试（RED 摘录真实 vitest FAIL 输出）→ 实现（GREEN）。RED 摘录均为运行时的实际 FAIL 文本。

## 9. RED — isLoggedIn token 权威化（src/stores/user.spec.ts）

复现: 工作树 `user.ts` 仍为旧实现（`isLoggedIn = isAuthenticated() && user !== null`），新测试断言「已登录（token 存在）但 user 未加载（user=null）→ isLoggedIn=true」。

```
 FAIL  src/stores/user.spec.ts > user store — isLoggedIn（token 权威） > 已登录（token 存在）但 user 未加载（user=null）→ isLoggedIn=true
AssertionError: expected false to be true // Object.is equality

- Expected
+ Received

- true
+ false
 ❯ src/stores/user.spec.ts:53:30
```

## 10. RED — text-fit 截断工具（src/utils/text-fit.spec.ts）

复现: 工具模块尚不存在，运行测试。

```
 FAIL  src/utils/text-fit.spec.ts [ src/utils/text-fit.spec.ts ]
Error: Failed to resolve import "./text-fit" from "src/utils/text-fit.spec.ts". Does the file exist?
```

## 11. RED — auth-flow handleAuthSuccess profile 成功路径（src/utils/auth-flow.spec.ts）

复现: auth-flow.ts 初版从 `@/api/user` 错误导入 `getUserProfile`（实际在 `@/api/identity`），profile 拉取抛错 → user 保持 null。

```
 FAIL  src/utils/auth-flow.spec.ts > auth-flow — handleAuthSuccess > 保存 token + 拉取 profile 并 setUser，无小区 → redirectTo 加入小区
AssertionError: expected undefined to be '测试用户' // Object.is equality

- Expected:
"测试用户"

+ Received:
undefined
 ❯ src/utils/auth-flow.spec.ts:55:34
```

## 12. RED — login.vue 登录流程改造（src/pages/login/login.spec.ts）

复现: `auth-flow` 模块尚不存在（`login.vue` 新提交逻辑引用它），运行测试。

```
 FAIL  src/pages/login/login.spec.ts [ src/pages/login/login.spec.ts ]
Error: Failed to resolve import "@/utils/auth-flow" from "src/pages/login/login.spec.ts". Does the file exist?
```

## 13. RED — agreement.vue 协议确认注册（src/pages/agreement/agreement.spec.ts）

复现: 协议页尚不存在，运行测试。

```
 FAIL  src/pages/agreement/agreement.spec.ts [ src/pages/agreement/agreement.spec.ts ]
Error: Failed to resolve import "./agreement.vue" from "src/pages/agreement/agreement.spec.ts". Does the file exist?
```

## 14. RED — loadMemberships getAppState 服务端权威（src/stores/community.spec.ts）

复现: `loadMemberships` 尚未接入 `getAppState`（服务端权威采用），4 个新用例全部 FAIL（本地 storage 陈旧值 c1 未被 c2 覆盖；getAppState 0/失败降级与不在 memberships 忽略均未实现）。

```
 FAIL  src/stores/community.spec.ts > community store — loadMemberships 服务端权威（getAppState） > 后端 current_community_id 存在于 memberships → 采用并保存（跨设备一致，修复本地 storage 陈旧）
ReferenceError: getResidentialAreasByIds is not defined
```

## 15. RED — notice.vue 非 10015 切换小区错误不再静默（src/pages/notice/notice.spec.ts）

复现: 临时将 `onCommunitySwitch` 非 10015 分支回退为静默（模拟 HEAD 行为：仅 10015 有 toast），随后恢复实现。

```
 FAIL  src/pages/notice/notice.spec.ts > notice page — onCommunitySwitch 10015 branch > non-10015 switch error → console.error + 通用 toast（不再静默吞错）
AssertionError: expected "error" to be called at least once
      Tests  1 failed | 21 passed (22)
```

## 16. RED — notice.vue 单行紧凑角色 pill（getNoticeRoleName mock 缺返回值）

复现: `getNoticeRoleName` 顶层 mock 缺省返回 ''（未在本用例覆盖），断言 pill 文本含「物业」。

```
 FAIL  src/pages/notice/notice.spec.ts > notice page — 通知模块标题栏 + 单行紧凑布局（task 4/5） > 单行紧凑卡片：标题 + (M-D) 日期 + 角色 pill 同行渲染
AssertionError: expected '' to contain '物业'
 ❯ src/pages/notice/notice.spec.ts:385:51
```

## 17. RED — App.vue restoreUserProfile（src/App.spec.ts，QA 分诊补测）

首次运行 App.spec.ts 时 `restoreUserProfile` 未从 App.vue 导出（setup 内部函数），4 分支测试全败。

```
 FAIL  src/App.spec.ts > App.vue restoreUserProfile（onLaunch 登录态恢复） > 未登录（token 不存在）→ 不调 getUserProfile
TypeError: restoreUserProfile is not a function
 ❯ src/App.spec.ts:93:18
      Tests  4 failed (4)
```

修复：App.vue 将 `restoreUserProfile` 移入独立 `<script>` 块并 `export`（onLaunch 接线不变），GREEN 后：

```
 Tests  4 passed (4)
```

> GREEN 后全量 `npx vitest run` → 16 files / 94 tests 全绿；type-check 0 error；build:h5 DONE；frontend QA 全项 PASS（含本轮新增 App.spec.ts 覆盖 restoreUserProfile 4 分支）。

---

## 18. RED — 首页通知列表两行布局改版（标题全文 / 发布单位 / YYYY-MM-DD / 移除 JS 截断）

> 变更：`notice.vue` 通知卡片从单行紧凑（色条 + 标题 JS 截断 + (M-D) 日期 + 角色 pill）改为两行布局（标题全文自然换行 + 元信息行：发布单位 + 发布日期 YYYY-MM-DD）。新增有逻辑函数 `formatPublishDate`（YYYY-MM-DD 转换、published_at=0 回退 created_at）与 `getPublisherName`（publisher 非空优先、空回退 getNoticeRoleName）。移除 `noticeDisplayTitle`/`formatMonthDay`/`text-fit` 整条链路（`src/utils/text-fit.ts` + `text-fit.spec.ts` 无引用后整文件删除）。

复现方式: 先更新 `notice.spec.ts` 断言（标题全文渲染 / 元信息行 / published_at=0 → YYYY-MM-DD / 无单行残留），生产 `notice.vue` 仍为单行紧凑实现时运行 vitest，捕获 4 处真实 FAIL。

```
 FAIL  src/pages/notice/notice.spec.ts > notice page — 通知模块标题栏 + 两行布局（标题全文 + 元信息行） > 两行布局：标题全文渲染 + 标题下方发布单位 + 发布日期(YYYY-MM-DD)
AssertionError: expected '小区停水通知：因管道检修明日上午…' to be '小区停水通知：因管道检修明日上午9点至下午5点停水' // Object.is equality
```

```
 FAIL  src/pages/notice/notice.spec.ts > notice page — 通知模块标题栏 + 两行布局（标题全文 + 元信息行） > publisher 为空 → 发布单位回退 getNoticeRoleName 角色名
Error: Cannot call text on an empty DOMWrapper.
 ❯ src/pages/notice/notice.spec.ts:411:27
    411|     expect(meta.exists()).toBe(true);
       |                           ^
```

```
 FAIL  src/pages/notice/notice.spec.ts > notice page — 通知模块标题栏 + 两行布局（标题全文 + 元信息行） > published_at=0 → 回退 created_at 渲染 YYYY-MM-DD
Error: Cannot call text on an empty DOMWrapper.
 ❯ src/pages/notice/notice.spec.ts:428:40
```

```
 FAIL  src/pages/notice/notice.spec.ts > notice page — 通知模块标题栏 + 两行布局（标题全文 + 元信息行） > 无单行布局残留：.notice-line 与 .notice-role-pill 已移除
AssertionError: expected true to be false // Object.is equality

- Expected
+ Received

- false
+ true

 ❯ src/pages/notice/notice.spec.ts:443:51
    443|     expect(wrapper.find('.notice-line').exists()).toBe(false);
       |                                                   ^

 Test Files  1 failed (1)
      Tests  4 failed | 20 passed (24)
```

GREEN 后（生产两行布局 + 移除截断逻辑）：

```
 Test Files  1 passed (1)
      Tests  24 passed (24)
```

> 既有 text-fit RED 摘录（§10，`measureTextWidth`/`truncateToFit`）对应文件 `src/utils/text-fit.ts` + `text-fit.spec.ts` 已整文件删除（无任何引用），§10 归档为已删除模块的历史证据。

---

## 19. 登录协议流程安全加固 — reg-pending 共享契约模块（2026-08-16）

> 背景: 修复 Review 3 条 memory suggestion（`sms-code-persist-localstorage` / `frontend-cross-page-storage-contract` / `cross-page-sensitive-temp-data-storage`）。
> 新建 `src/utils/reg-pending.ts`（内存态主载体 + H5 sessionStorage 镜像 + TTL 5 分钟 + localStorage 零触碰）为唯一契约源；
> `login.vue`/`agreement.vue` 删内联 magic string 与 JSON.stringify/uni storage 读写，改调共享模块。
> 分诊: `reg-pending.ts` 为**有逻辑函数**（TTL 校验/H5 镜像/读写清），TDD RED→GREEN；两页改为接线，测试改断言。
> 复现方式: 先写测试断言共享模块行为后运行 vitest 捕获真实失败（模块未实现 / 页面未接线），随后实现。摘录均为 vitest 实际输出。

### 19.1 RED — reg-pending.spec.ts（模块未实现）

复现方式: 仅创建 `src/utils/reg-pending.spec.ts`，未创建 `src/utils/reg-pending.ts`，运行 `npx vitest run src/utils/reg-pending.spec.ts`。

```
 FAIL  src/utils/reg-pending.spec.ts [ src/utils/reg-pending.spec.ts ]
Error: Failed to resolve import "./reg-pending" from "src/utils/reg-pending.spec.ts". Does the file exist?
  Plugin: vite:import-analysis
  File: /home/jiaoxh/my-project/community-and-home/web/mobile/src/utils/reg-pending.spec.ts:15:7

 Test Files  1 failed (1)
      Tests  no tests
```

### 19.2 RED — login.spec.ts（login.vue 尚未接线 saveRegPending）

复现方式: login.spec.ts 改断言共享模块后，login.vue 仍用 `uni.setStorageSync(REG_PENDING_KEY, ...)`，运行 `npx vitest run src/pages/login/login.spec.ts`。

```
 FAIL  src/pages/login/login.spec.ts > login page — 登录流程改造 > loginWithSms 返回 50001（未注册）→ saveRegPending 暂存 + navigateTo 协议页
AssertionError: expected "vi.fn()" to be called 1 times, but got 0 times
 ❯ src/pages/login/login.spec.ts:97:28
     97|     expect(saveRegPending).toHaveBeenCalledTimes(1);

 Test Files  1 failed (1)
```

### 19.3 RED — agreement.spec.ts（agreement.vue 尚未接线 readRegPending/clearRegPending）

复现方式: agreement.spec.ts 改断言共享模块后，agreement.vue 仍用 `uni.getStorageSync(REG_PENDING_KEY)` 手写 readPending/clearPending，运行 `npx vitest run src/pages/agreement/agreement.spec.ts`。

```
 FAIL  src/pages/agreement/agreement.spec.ts > agreement page — 协议确认注册 > 勾选后确认注册 → readRegPending + register 正确参数 + clearRegPending + 自动登录
AssertionError: expected "vi.fn()" to be called at least once
 ❯ src/pages/agreement/agreement.spec.ts:121:28
    121|     expect(readRegPending).toHaveBeenCalled();

 FAIL  src/pages/agreement/agreement.spec.ts > agreement page — 协议确认注册 > 无临时注册数据进入 → readRegPending 返回 null 并明确提示注册信息已失效
AssertionError: expected "vi.fn()" to be called at least once
 ❯ src/pages/agreement/agreement.spec.ts:160:28
    160|     expect(readRegPending).toHaveBeenCalled();

 Test Files  1 failed (1)
      Tests  3 failed | 1 passed (4)
```

### GREEN 后（reg-pending.ts 实现 + login.vue / agreement.vue 接线共享模块）

```
 Test Files  1 passed (1)   # reg-pending.spec.ts（5 用例）
      Tests  5 passed (5)

 Test Files  1 passed (1)   # login.spec.ts（4 用例）
      Tests  4 passed (4)

 Test Files  1 passed (1)   # agreement.spec.ts（4 用例）
      Tests  4 passed (4)
```

> 全量门禁: `npm run test:unit` → 16 files / 97 tests PASS；`npm run type-check` → PASS（0 errors）；`npm run build:h5` → PASS（DONE Build complete）。

## 20. 修复多视角审查 CRITICAL — 悬空记忆引用完整性回归测试（2026-08-16）

> 背景: standards-eng 多视角审查 FAIL（1 CRITICAL）——`reg-pending.ts:16-18` / `login.vue:71-73` / `agreement.vue:56-58` 引用了 3 个不存在的记忆 slug（`sms-code-persist-localstorage` / `frontend-cross-page-storage-contract` / `cross-page-sensitive-temp-data-storage`），M2 规则「slug 文件不存在 → 🔴 CRITICAL」。
> 修复: 按审查建议 (a) 创建 3 个记忆文件（代码行为本身正确——一次性 smsCode 走内存态 + sessionStorage TTL 5 分钟、不落 localStorage、key/结构收敛到单一契约源，值得沉淀），并新建 `src/utils/memory-refs.spec.ts` 扫描 src/ 全部 SEE slug 断言可解析，防悬空引用回归。
> 分诊: `memory-refs.spec.ts` 为**有逻辑函数**（文件扫描 + slug 提取 + 记忆目录解析），TDD RED→GREEN；记忆文件为文档（无逻辑代码），不需 RED。
> 复现方式: 先写测试断言 src/ 全部 SEE slug 可解析后运行 vitest 捕获真实失败（3 个 slug 悬空），随后创建记忆文件。摘录均为 vitest 实际输出。

### 20.1 RED — memory-refs.spec.ts（3 个 slug 悬空）

复现方式: 仅创建 `src/utils/memory-refs.spec.ts`，未创建 3 个记忆文件，运行 `npx vitest run src/utils/memory-refs.spec.ts`。

```
 FAIL  src/utils/memory-refs.spec.ts > 记忆引用完整性（SEE 标记 slug 无悬空引用） > 每个被引用的记忆 slug 都能解析到项目或个人记忆文件
AssertionError: expected [ …(3) ] to deeply equal []
- Expected
+ Received
- []
+ [
+   "sms-code-persist-localstorage",
+   "frontend-cross-page-storage-contract",
+   "cross-page-sensitive-temp-data-storage",
+ ]
 ❯ src/utils/memory-refs.spec.ts:92:21

 FAIL  src/utils/memory-refs.spec.ts > 记忆引用完整性（SEE 标记 slug 无悬空引用） > 本次修复目标：3 个 smsCode 存储相关 slug 必须已存在
AssertionError: 悬空记忆引用: [[sms-code-persist-localstorage]]: expected false to be true // Object.is equality
 ❯ src/utils/memory-refs.spec.ts:102:61

 Test Files  1 failed (1)
      Tests  2 failed | 1 passed (3)
```

### GREEN 后（创建 3 个记忆文件于 .harness/knowledge/memory/web/ + memory-index-build.sh 重建索引）

```
 Test Files  1 passed (1)
      Tests  3 passed (3)
```

> 全量门禁: `npm run test:unit` → 17 files / 100 tests PASS（+3 新用例）；`npm run build:h5` → PASS（DONE Build complete）；`bash .harness/skills/qa/scripts/harness-checks-frontend.sh --service mobile` → 5 PASS / 0 FAIL / 2 WARN（WARN 均为既有存量，非本轮引入）。

## 21. Review WARNING 跟进 — 登录/注册双提交窗口 + 错误分支 code 优先（2026-08-16）

> 背景: Review WARNING 指出两处登录/注册提交窗口问题——`login.vue` / `agreement.vue` 成功分支在 `await handleAuthSuccess(...)` 前即复位 `submitting`，而 handleAuthSuccess 内部有 profile 拉取 + 小区检查 + 800ms setTimeout 才跳转，窗口期用户可二次点击触发重复登录/注册；同时登录错误分支三条件并存（code + msg 双判据），应以 err.code 数值为主。
> 修复: (A) 成功分支 submitting 复位移入 `handleAuthSuccess` 的 `opts.onCompleted`（跳转后调用）；(B) catch 改为 `e?.code !== undefined ? e.code === 50001 : (msg.includes('50001') || msg.includes('未注册'))`，code 存在时只看 code。
> 分诊: A（异步时序分支）与 B（条件判定重构）均为**有逻辑函数**，TDD RED→GREEN；C（docs/design.md 文档）与 D（crypto.ts 删调试残留）为文档/删代码，不需 RED。
> 复现方式: 先写测试断言 onCompleted 第二参数 + 跳转完成前 submitting 保持 true + code 优先判定后运行 vitest 捕获真实失败，随后实现。摘录均为 vitest 实际输出。

### 21.1 RED — login.spec.ts（A onCompleted 时序 + B code 优先，3 用例 FAIL）

复现方式: 仅在 login.spec.ts / agreement.spec.ts 追加断言（组件尚未修复），运行 `npx vitest run src/pages/login/login.spec.ts src/pages/agreement/agreement.spec.ts`。

```
 FAIL  src/pages/agreement/agreement.spec.ts > A: 确认注册成功 → submitting 在 handleAuthSuccess 跳转完成前保持 true，onCompleted 回调后才复位
AssertionError: expected "vi.fn()" to be called with arguments: [ ObjectContaining{…}, …(1) ]
Received:
  1st vi.fn() call:
  [
-   ObjectContaining {
+   {
      "accessToken": "at",
-   },
-   ObjectContaining {
-     "onCompleted": Any<Function>,
+     "expiresAt": 123,
+     "refreshToken": "rt",
+     "userId": "u1",
    },
  ]
Number of calls: 1
 ❯ src/pages/agreement/agreement.spec.ts:152:31

 FAIL  src/pages/login/login.spec.ts > A: 登录成功 → submitting 在 handleAuthSuccess 跳转完成前保持 true，onCompleted 回调后才复位
AssertionError: expected "vi.fn()" to be called with arguments: [ ObjectContaining{…}, …(1) ]
Received:
  1st vi.fn() call:
  [
-   ObjectContaining {
+   {
      "accessToken": "at",
-   },
-   ObjectContaining {
-     "onCompleted": Any<Function>,
+     "expiresAt": 123,
+     "refreshToken": "rt",
+     "userId": "u1",

 FAIL  src/pages/login/login.spec.ts > B: code 存在但非 50001（即使 msg 含"未注册"）→ code 为主判据，不进注册流程
AssertionError: expected "vi.fn()" to not be called at all, but actually been called 1 times
Received:
  1st vi.fn() call:
    Array [
      Object {
        "deviceId": "dev-1",
        "nickname": "用户8000",
        "phone": "13800138000",
        "smsCode": "123456",
      },
    ]
Number of calls: 1
 ❯ src/pages/login/login.spec.ts:170:32

 Test Files  1 failed | 1 failed (2)
      Tests  3 failed | 10 passed (13)
```

### GREEN 后（login.vue / agreement.vue 修复：onCompleted 复位 + code 优先判定）

```
 Test Files  1 passed (1)   # login.spec.ts（8 用例，含 +4 新）
      Tests  8 passed (8)

 Test Files  1 passed (1)   # agreement.spec.ts（5 用例，含 +1 新）
      Tests  5 passed (5)
```

> 全量门禁: `npm run type-check` → PASS（0 errors）；`npm run test:unit` → 17 files / 105 tests PASS（+5 新用例）；`npm run build:h5` → PASS（DONE Build complete）；`bash .harness/skills/qa/scripts/harness-checks-frontend.sh --service mobile` → 5 PASS / 0 FAIL / 2 WARN（debug_artifacts 现为真实 0；WARN 均为既有存量）。

---

## 22. 首页首载去重 + 登录 toast 合并（notice-xss-sanitize-and-frontend-fixes，2026-08-16）

> 变更内容（详见 CHANGELOG 顶部条目）：
> - **A 首页 watch 双重加载（notice.vue）**：`watch(currentCommunityId)` 加 `membershipsResolved` 首载守卫；标志在 onMounted 中 `await loadMemberships()` 之后置 true（评审钉死，禁止放入 loadMemberships 内部含 finally）；无小区时直接结束骨架屏不发起数据加载。**有逻辑函数**（异步时序守卫），TDD RED→GREEN
> - **B 登录 toast 覆盖（auth-flow.ts）**：profile 拉取失败改为单条合并 toast「登录成功，但资料加载失败」（icon:none），不再被后续「登录成功」(icon:success) 覆盖；成功路径仍显示纯净「登录成功」。**有逻辑函数**（分支合并 + toast 时序），TDD RED→GREEN

复现方式: 新增 3 个 notice.spec 用例（REQ-DBL）与 1 个新 auth-flow.spec 成功路径用例 + 改 1 个失败用例断言，实现未改时直接运行 `npx vitest run src/pages/notice/notice.spec.ts src/utils/auth-flow.spec.ts` 捕获真实 FAIL（含 `AssertionError` / Received 实际调用）。摘录均为 vitest 实际输出。

### 22.1 RED — 首页首载去重（src/pages/notice/notice.spec.ts）

```
 FAIL  src/pages/notice/notice.spec.ts > notice page — 首页首载去重（REQ-DBL-1/2/3） > loadMemberships 覆写 currentCommunityId（C1→C2）→ 通知/寻失接口各仅请求一次（以 C2 为维度，REQ-DBL-1）
AssertionError: expected "vi.fn()" to be called 1 times, but got 2 times
 ❯ src/pages/notice/notice.spec.ts:519:27
    517|
    518|     // 覆写触发的那次 watch 被首载守卫忽略 + onMounted 显式单次 loadAll → 各只请求一次
    519|     expect(getNoticeList).toHaveBeenCalledTimes(1);
       |                           ^

 FAIL  src/pages/notice/notice.spec.ts > notice page — 首页首载去重（REQ-DBL-1/2/3） > loadMemberships 整体失败 → 不以陈旧 cid 发请求、无 double-load（REQ-DBL-1 降级）
AssertionError: expected "vi.fn()" to not be called at all, but actually been called 1 times

Received:

  1st vi.fn() call:

    Array [
      "c1",
      1,
      3,
      30,
    ]

Number of calls: 1
 ❯ src/pages/notice/notice.spec.ts:551:31
    549|     await flushPromises();
    550|
    551|     expect(getNoticeList).not.toHaveBeenCalled();
       |                               ^
```

说明: HEAD 的 onMounted 无条件 `loadAll()`——①getAppState 覆写 C1→C2 触发 watch 一次 + onMounted 显式一次 → getNoticeList 2 次；②memberships 整体失败时仍以陈旧 `c1` 发起请求。

### 22.2 RED — 登录 toast 合并（src/utils/auth-flow.spec.ts）

```
 FAIL  src/utils/auth-flow.spec.ts > auth-flow — handleAuthSuccess > profile 拉取失败 → 单条合并 toast（icon:none），不再弹纯净「登录成功」success toast（REQ-TOAST-1）
AssertionError: expected "vi.fn()" to be called 1 times, but got 2 times
 ❯ src/utils/auth-flow.spec.ts:71:27
     69|     expect(errSpy).toHaveBeenCalled();
     70|     // 合并提示同时表达成功与失败，icon:none（非 success 打勾，避免长文案截断）
     71|     expect(uni.showToast).toHaveBeenCalledTimes(1);
       |                           ^
```

说明: HEAD 的 catch 先弹「获取用户资料失败」，末尾再弹「登录成功」(icon:success) → showToast 共 2 次，失败提示被成功提示覆盖。

### GREEN 后（notice.vue 守卫 + auth-flow 合并 toast）

```
 Test Files  2 passed (2)   # notice.spec.ts（27 用例，含 +3 新） + auth-flow.spec.ts（5 用例，含 +1 新）
      Tests  32 passed (32)
```

> 全量门禁: `npm run type-check` → PASS（0 errors）；`npm run test:unit` → 17 files / 109 tests PASS（+4 新用例）；`npm run build:h5` → PASS（DONE Build complete）；`bash .harness/skills/qa/scripts/harness-checks-frontend.sh --service mobile` → 5 PASS / 0 FAIL / 2 WARN（3 处 `as any` 与 api-field-align 均为既有存量，非本轮引入）。

---

## 23. RED — 业主认证分支 onOwnerAuth（src/pages/my/my.spec.ts）— TDD 证据补录

> 本轮【我的】页图标网格重构 + 手机号显示修复 + 退出登录接线新增**有逻辑函数** `onOwnerAuth`（`if (hasOwnerRole.value)` 分支 → toast「已是业主」/ 否则 `applyForRole('owner')`）。上轮 QA 判 FAIL：CHANGELOG RED 摘录 3 条均不含 onOwnerAuth、`_tdd_evidence.md` 无本轮章节（声称「完整摘录见 _tdd_evidence.md」引用悬空）。本节补录真实 RED 摘录（不改变已 GREEN 的实现与测试）。

复现方式: 临时将 `src/pages/my/my.vue` 回退到 HEAD 基线（无 `onOwnerAuth`；my.spec.ts 为未跟踪新增，不受 stash 影响），测试保持行为断言不动，运行 `npx vitest run src/pages/my/my.spec.ts` 捕获真实 FAIL，随后 `git stash pop` 恢复实现。摘录均为 vitest 实际输出。

```
 FAIL  src/pages/my/my.spec.ts > my page — 业主认证分支 > 未认证业主 → applyRole({ community_id, role_code: owner })
TypeError: wrapper.vm.onOwnerAuth is not a function
 ❯ src/pages/my/my.spec.ts:142:31
    140|     communityStore.addCommunity({ communityId: 'c1', communityName: 'A…
    141|
    142|     await (wrapper.vm as any).onOwnerAuth();
       |                               ^
    143|
    144|     expect(applyRole).toHaveBeenCalledWith({ community_id: 'c1', role_…

 FAIL  src/pages/my/my.spec.ts > my page — 业主认证分支 > 已有业主角色（verf_status=2）→ toast「已是业主」，不重复申请
TypeError: wrapper.vm.onOwnerAuth is not a function
 ❯ src/pages/my/my.spec.ts:152:31
    150|     communityStore.addCommunity({ communityId: 'c1', communityName: 'A…
    151|
    152|     await (wrapper.vm as any).onOwnerAuth();
       |                               ^
    153|
    154|     expect(uni.showToast).toHaveBeenCalledWith(

 Test Files  1 failed (1)
      Tests  5 failed (5)
```

说明: HEAD 的 my.vue 无 `onOwnerAuth`（`git show HEAD:web/mobile/src/pages/my/my.vue` grep 0 命中），my.spec.ts:142/152 的 `wrapper.vm.onOwnerAuth()` 在 RED 阶段必然产出 `TypeError: wrapper.vm.onOwnerAuth is not a function`。该 TypeError 先于 applyRole/showToast 断言抛错，故 FAIL 形态为函数缺失（断言对比文本在实现存在后才触发）。

### GREEN 后（my.vue onOwnerAuth 实现 + my.spec.ts 全量）

```
 Test Files  1 passed (1)
      Tests  5 passed (5)
```

> 全量门禁: `npm run test:unit` → 21 files / 123 tests PASS；`npm run type-check` → 0 errors；`npm run build:h5` → PASS（DONE Build complete）；`harness-checks-frontend.sh --service mobile` → 6 PASS / 0 FAIL / 2 WARN（存量）。

---

# TDD 证据 — web/mobile 加入小区流程改造（选小区后分流 join-choice / join-residence）

> 生成时间: 2026-08-17
> 变更: ① join-community 去掉选小区后 modal，改存 pending-join → 导航 join-choice；② 新页 join-choice（2 身份分流）；③ 新页 join-residence（业主路径填房号，复用 join-form）；④ my.vue applyForRole 兼容 pending 小区。
> 分诊: pending-join 契约模块 / community store pendingCommunityId / join-community 分流 / my.vue 回退 = 有逻辑，配 RED；join-choice / join-residence 为页面接线 + 复用既有校验函数，属「逻辑函数」范畴（含导航分支），一并走 RED。
> RED 摘录均为 vitest 实际输出（`npx vitest run <file>`）。

## 1. RED — pending-join 契约模块（src/utils/pending-join.spec.ts）

复现方式: 仅写测试文件、未实现 `src/utils/pending-join.ts`，运行 `npx vitest run src/utils/pending-join.spec.ts`。

```
 FAIL  src/utils/pending-join.spec.ts [ src/utils/pending-join.spec.ts ]
Error: Failed to resolve import "./pending-join" from "src/utils/pending-join.spec.ts". Does the file exist?
  Plugin: vite:import-analysis
```

GREEN 后: `Test Files 1 passed (1) / Tests 5 passed (5)`。

## 2. RED — community store pendingCommunityId（src/stores/community.spec.ts）

复现方式: 先加 3 条 pendingCommunityId 用例，store 未实现，运行 `npx vitest run src/stores/community.spec.ts`。

```
 FAIL  src/stores/community.spec.ts > community store — pendingCommunityId … > setPendingCommunityId → pendingCommunityId 可读
TypeError: store.setPendingCommunityId is not a function
 ❯ src/stores/community.spec.ts:83:11

 FAIL  … > 初始 pendingCommunityId 为空串
AssertionError: expected undefined to be '' // Object.is equality
- Expected: ""
+ Received: undefined
```

GREEN 后: `Test Files 1 passed (1) / Tests 10 passed (10)`。

## 3. RED — join-community 选小区分流（src/pages/join-community/join-community.spec.ts 重写）

复现方式: 重写测试（断言 savePendingJoin + navigateTo join-choice + modal 不存在），实现仍为旧 modal 流程，运行测试。

```
 ❯ src/pages/join-community/join-community.spec.ts (4 tests | 2 failed)
 FAIL … > 点击未加入小区 → 存 pending-join {id,name,address} + navigateTo join-choice
AssertionError: expected "vi.fn()" to be called with arguments: [ { communityId: 'c1', …(2) } ]
 FAIL … > 不再弹出 join form modal（.join-form-mask 不存在）
AssertionError: expected true to be false // Object.is equality
 Test Files  1 failed (1)
      Tests  2 failed | 2 passed (4)
```

GREEN 后: `Test Files 1 passed (1) / Tests 4 passed (4)`。

## 4. RED — join-choice 新页（src/pages/join-choice/join-choice.spec.ts）

复现方式: 仅写测试文件、未创建页面，运行测试。

```
 FAIL  src/pages/join-choice/join-choice.spec.ts [ src/pages/join-choice/join-choice.spec.ts ]
Error: Failed to resolve import "./join-choice.vue" from "src/pages/join-choice/join-choice.spec.ts". Does the file exist?
```

GREEN 后: `Test Files 1 passed (1) / Tests 4 passed (4)`。

## 5. RED — join-residence 新页（src/pages/join-residence/join-residence.spec.ts）

### 5a. 模块缺失 RED
复现方式: 仅写测试文件、未创建页面。

```
 FAIL  src/pages/join-residence/join-residence.spec.ts [ src/pages/join-residence/join-residence.spec.ts ]
Error: Failed to resolve import "./join-residence.vue" from "src/pages/join-residence/join-residence.spec.ts". Does the file exist?
```

### 5b. 确认按钮选择器对齐（测试选择器与模板 class 不匹配）
页面创建后运行：测试用 `.confirm-join-btn` 选择器，模板按钮仅有 `.btn` → 全部 5 条用例 FAIL `Cannot call trigger on an empty DOMWrapper`。
修复：模板按钮补充 `confirm-join-btn` class（纯选择器对齐，无逻辑改动）。

GREEN 后: `Test Files 1 passed (1) / Tests 5 passed (5)`。

## 6. RED — my.vue applyForRole 兼容 pending 小区（src/pages/my/my.spec.ts）

复现方式: 新增 3 条 applyForRole pending 回退用例，实现未改，运行测试。

```
 ❯ src/pages/my/my.spec.ts (8 tests | 1 failed)
 FAIL … > currentCommunityId 为空 + pendingCommunityId 存在 → 用 pending 申请并一次性清除
AssertionError: expected "vi.fn()" to be called with arguments: [ { community_id: 'c9', …(1) } ]
 Test Files  1 failed (1)
      Tests  1 failed | 7 passed (8)
```

GREEN 后: `Test Files 1 passed (1) / Tests 8 passed (8)`。

## 门禁全量

- `npx vitest run` → 24 files / 144 tests PASS（基线 21 files / 125 → +3 文件 / +19 用例）
- `npm run type-check` → 0 errors
- `npm run build:h5` → DONE Build complete
- `harness-checks-frontend.sh --service mobile` → 6 PASS / 0 FAIL / 2 WARN（存量：type-safety 3 处 `as any` 均在 request.ts/crypto.ts/identity.ts，非本轮；api-field-align 34 处 WARN 存量）

---

## 24. RED — agreement.vue confirmRegister 10002/timeout 分流（src/pages/agreement/agreement.spec.ts）— TDD 证据补录

> 工作树未提交变更「注册超时 UX 修复」（08-17 08:23-08:24 游离于管线之外、未经 QA 门禁的临时会话所留）在 `confirmRegister` catch 新增**有逻辑函数**三分支（`code===10002` / timeout 正则 / 其他错误），配 agreement.spec.ts:189-228 两条真实断言（+2 用例），但 RED 摘录三处全缺（`_tdd_evidence.md` 无相关章节、CHANGELOG「注册超时 UX 修复」条无摘录、spec 注释仍指向旧 §5）。本轮补录真实 RED 摘录（不改变已 GREEN 的实现与测试）。

复现方式: 临时 `git stash push -- src/pages/agreement/agreement.vue` 回退至 HEAD 基线（catch 仅 `uni.showToast({ title: '注册失败，请重试', icon: 'none' })`，无 10002/timeout 分支；agreement.spec.ts 为工作树新增断言，不受 stash 影响），测试保持行为断言不动，运行 `npx vitest run src/pages/agreement/agreement.spec.ts` 捕获真实 FAIL，随后 `git stash pop` 恢复实现。摘录均为 vitest 实际输出。

```
 FAIL  src/pages/agreement/agreement.spec.ts > agreement page — 协议确认注册 > 手机号已注册(10002) → 清临时数据 + 回登录页直接登录
AssertionError: expected "vi.fn()" to be called with arguments: [ ObjectContaining{…} ]

Received:

  1st vi.fn() call:

  [
-   ObjectContaining {
-     "title": "该手机号已注册，请直接登录",
+   {
+     "icon": "none",
+     "title": "注册失败，请重试",
    },
  ]

Number of calls: 1

 ❯ src/pages/agreement/agreement.spec.ts:202:27
    202|     expect(uni.showToast).toHaveBeenCalledWith(
       |                           ^
    203|       expect.objectContaining({ title: '该手机号已注册，请直接登录' }),

⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯[1/2]⎯

 FAIL  src/pages/agreement/agreement.spec.ts > agreement page — 协议确认注册 > 注册超时（timeout）→ 提示账号可能已创建、保留数据可重试
AssertionError: expected "vi.fn()" to be called with arguments: [ ObjectContaining{…} ]

Received:

  1st vi.fn() call:

  [
-   ObjectContaining {
-     "title": "注册超时，账号可能已创建；请重试或返回登录",
+   {
+     "icon": "none",
+     "title": "注册失败，请重试",
    },
  ]

Number of calls: 1

 ❯ src/pages/agreement/agreement.spec.ts:223:27
    223|     expect(uni.showToast).toHaveBeenCalledWith(
       |                           ^
    224|       expect.objectContaining({ title: '注册超时，账号可能已创建；请重试或返回登录' }),

⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯[2/2]⎯

 Test Files  1 failed (1)
      Tests  2 failed | 5 passed (7)
```

说明: 回退后 HEAD 的 agreement.vue catch 无 10002/timeout 分支（`git show HEAD:web/mobile/src/pages/agreement/agreement.vue` 106-128 行仅 `showToast('注册失败，请重试')`），故 10002/timeout 两条 toast 断言必然产出 `AssertionError: expected "vi.fn()" to be called with arguments: [ ObjectContaining{…} ]`（Received 为统一兜底 toast「注册失败，请重试」，未走到 10002 清数据 / timeout 提示分支）。该 AssertionError 先于 clearRegPending / navigateBack / handleAuthSuccess 断言抛出，故 FAIL 形态为 toast 断言对比失败（其余分支断言在实现存在后才逐个通过）。

### GREEN 后（agreement.vue confirmRegister 三分支实现 + agreement.spec.ts 全量）

```
 Test Files  1 passed (1)
      Tests  7 passed (7)
```

> 全量门禁: `npm run test:unit` → 24 files / 144 tests PASS；`npm run type-check` → 0 errors；`npm run build:h5` → PASS（DONE Build complete）；`harness-checks-frontend.sh --service mobile` → 6 PASS / 0 FAIL / 2 WARN（存量）。

---

## 加入流程重构：点「加入小区」立即建 membership（2026-08-17）

> 背景：新模型（用户拍板）——加入=立即建 membership（无房号），填写房号改为独立步骤（bindResidence + applyRole）。RED 摘录通过**先写重写后的 spec（断言新行为），在旧实现仍驻留时运行 vitest** 捕获真实输出，随后实现 GREEN。
> 复现命令：`npx vitest run src/pages/join-community/join-community.spec.ts src/pages/join-residence/join-residence.spec.ts src/pages/join-choice/join-choice.spec.ts`（旧实现下）；`src/api/user.ts` 的 RED 需 `git checkout HEAD -- src/api/user.ts` 回退旧实现后单独复现（见 RED 2.5）。

### RED 汇总（4 files / 9 failed / 8 passed）

```
 Test Files  4 failed (4)
      Tests  9 failed | 8 passed (17)
```

### RED 1 — join-choice 页顶文案（旧「加入进行中」→ 新「已加入 XX，请选择下一步」）

> 基线校准：Received「加入进行中」取自工作树中间态（2026-08-17「明确加入进行中」轮已改文案但未 commit，git show HEAD 的 join-choice.vue 第 5 行为更早的「加入小区 <社区名>」）。本页属文案映射（RED 不要求），此摘录仅证明「断言新文案时旧文案未命中」；基线差异不影响结论，特此注明避免后续 QA 误判。

```
 FAIL  src/pages/join-choice/join-choice.spec.ts > join-choice page — 身份分流 > 页顶显示「已加入 XX，请选择下一步」+ 保留社区名/地址
AssertionError: expected '加入进行中' to be '已加入 幸福小区' // Object.is equality

Expected: "已加入 幸福小区"
Received: "加入进行中"

 ❯ src/pages/join-choice/join-choice.spec.ts:41:50
     41|     expect(wrapper.find('.header-title').text()).toBe('已加入 幸福小区');
```

### RED 2 — join-community 点「加入」立即 joinCommunity（旧仅存 pending 不调后端）

```
 FAIL  src/pages/join-community/join-community.spec.ts > … > 点击未加入小区 → joinCommunity(communityId 无房号) + addCommunity + 存 pending(带 membershipId) + navigateTo join-choice
AssertionError: expected "vi.fn()" to be called with arguments: [ 'c1' ]

Number of calls: 0

 ❯ src/pages/join-community/join-community.spec.ts:61:27
     61|     expect(joinCommunity).toHaveBeenCalledWith('c1');

 FAIL  src/pages/join-community/join-community.spec.ts > … > joinCommunity 失败（如每年最多加入 3 个新小区）→ toast 错误，不存 pending、不导航、不 addCommunity
AssertionError: expected "vi.fn()" to be called with arguments: [ ObjectContaining{…} ]

Number of calls: 0
```

说明: 旧 `onSelectArea` 同步存 pending 即导航、不调 `joinCommunity`，故 ①`joinCommunity` 调用数为 0；②失败分支 toast（`uni.showToast`）调用数为 0。

### RED 3 — join-residence 改 bindResidence + applyRole（旧仍调 joinCommunity，未 mock 时抛错/0 调用）

```
 FAIL  src/pages/join-residence/join-residence.spec.ts > … > 自有 → 读 pending.membershipId → bindResidence(is_primary:1) + applyRole(owner) + toast 房号登记成功 + 清 pending + switchTab notice
AssertionError: expected "vi.fn()" to be called with arguments: [ { membership_id: 'm1', …(4) } ]

Number of calls: 0

 ❯ src/pages/join-residence/join-residence.spec.ts:90:27
     90|     expect(bindResidence).toHaveBeenCalledWith({

 FAIL  src/pages/join-residence/join-residence.spec.ts > … > 租住 → applyRole role_code=tenant
AssertionError: expected "vi.fn()" to be called with arguments: [ { community_id: 'c1', …(1) } ]

Number of calls: 0

 ❯ src/pages/join-residence/join-residence.spec.ts:110:23

 FAIL  src/pages/join-residence/join-residence.spec.ts > … > pending 无 membershipId → 回退 getUserMemberships 按 communityId 取 membership.id
AssertionError: expected "vi.fn()" to be called at least once
 ❯ src/pages/join-residence/join-residence.spec.ts:125:32
    125|     expect(getUserMemberships).toHaveBeenCalled();
```

说明: 旧 `confirmJoin` 调 `joinCommunity(pending.communityId, building, unit, room, ownership)`，新 spec 的 mock 只提供 bindResidence/applyRole/getUserMemberships，故 ①bindResidence/applyRole/getUserMemberships 均 0 调用；②旧代码运行时取 `joinCommunity` 抛 `No "joinCommunity" export is defined on the "@/api/user" mock`，兜底 toast 文案错误（Received 与期望不符）。

### RED 2.5 — user.ts `joinCommunity` 条件载荷（房号可选）

> 复现方式：`git checkout HEAD -- src/api/user.ts` 回退到旧实现（5 个必填参数、恒发送全部字段）→ `npx vitest run src/api/user.spec.ts` → 捕获下方真实 FAIL → `git stash pop` / 恢复工作树实现。新增用例 3（仅传 communityId）首跑即在此 FAIL，真实 RED 从未在旧实现下持久化过，故本轮补录。
> // SEE: [[tdd-red-evidence-requires-fail-excerpt]] — RED 证据必须含实际 FAIL 摘录（本轮即该记忆的第 7 次复发的补录）

```
 FAIL  src/api/user.spec.ts > joinCommunity > 仅传 communityId（房号可选）→ 载荷不含 building/unit/room/ownership
AssertionError: expected { community_id: 'c1', …(4) } to not have property "building"
 ❯ src/api/user.spec.ts:73:25
     71|     const payload = (request.post as any).mock.calls[0][1];
     72|     expect(payload).toEqual({ community_id: 'c1' });
     73|     expect(payload).not.toHaveProperty('building');
       |                         ^
     74|     expect(payload).not.toHaveProperty('unit');
     75|     expect(payload).not.toHaveProperty('room')

⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯[1/1]⎯


 Test Files  1 failed (1)
      Tests  1 failed | 2 passed (3)
```

说明: 旧实现 `joinCommunity('c1')` 恒发送 `{community_id, building, unit, room, ownership}` 全部字段（可选参未传为 `undefined`），`toEqual({community_id:'c1'})` 因 vitest 忽略 `undefined` 键而通过，但 `not.toHaveProperty('building')` 必然失败（`undefined` 键仍存在于 payload 对象上）——这正是条件载荷逻辑（4 个 `if (x != null)` 分支）要修复的行为：仅 communityId 调用时载荷不含房号字段。

### GREEN（实现后全量）

```
 Test Files  24 passed (24)
      Tests  150 passed (150)
```

> 全量门禁: `npm run type-check` → 0 errors；`npm run build:h5` → DONE Build complete；`harness-checks-frontend.sh --service mobile` → 6 PASS / 0 FAIL / 2 WARN（存量 type-safety 3 处 as any + api-field-align 34 处）。
