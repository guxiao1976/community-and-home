# CHANGELOG — web/mobile

## 2026-08-17 — joinCommunity 解包 {membership} 修复（Review：membershipId 回填静默失效）

### 做了什么
- `api/user.ts joinCommunity()` 原 `return res as unknown as CommunityMembership`——REST 响应解包 data 后仍是 `{membership:{...}}`，调用方取 `res.id` 恒 undefined → membershipId 回填失效（靠 getUserMemberships 兜底掩盖）。改为解出 `data.membership` 返回资源本身。
- 记忆 `frontend-api-return-wrapped-resource-unwrap` 落盘。

### 门禁
- `npx vitest run` → 24 files / 150 tests PASS；harness-checks-frontend 6 PASS / 0 FAIL。

---

# CHANGELOG — web/mobile

## 2026-08-17 — TDD 证据补录：joinCommunity 条件载荷真实 RED（QA FAIL 修复轮）

### 做了什么
- **QA FAIL 项**：CHANGELOG 分诊 B「joinCommunity 房号可选（有逻辑函数：条件载荷）… TDD RED→GREEN」，但 `_tdd_evidence.md` 无其真实 vitest FAIL 摘录（GREEN 阶段与新增 user.spec.ts 用例 3 一并实现，首跑即绿，RED 从未在旧实现下持久化）。
- **修复（证据补录，无生产代码改动）**：`git checkout HEAD -- src/api/user.ts` 回退旧实现（5 必填参数、恒发全部字段）→ `npx vitest run src/api/user.spec.ts` 复现真实 FAIL → `1 failed | 2 passed`，`AssertionError: expected { community_id: 'c1', …(4) } to not have property "building"` → 恢复工作树实现。
- **摘录落盘**：`_tdd_evidence.md` 新增 §RED 2.5（user.ts joinCommunity 条件载荷）；RED 汇总 3 files/8 failed → 4 files/9 failed；复现命令补充 user.ts 回退说明。
- **完整性注记**：`_tdd_evidence.md` §RED 1（join-choice 文案）Received「加入进行中」基线取自工作树中间态（非 HEAD「加入小区」），已注明避免误判。

### 测试
- `api/user.spec.ts:66-77` 用例 3（仅传 communityId → 载荷不含房号字段）即回归测试：旧实现下 FAIL（`not.toHaveProperty('building')`）→ 新实现 GREEN。无新增用例（用例已存在，本轮仅补录其 RED）。

### 门禁
- `npm run test:unit` → 24 files / 150 tests PASS；`npm run type-check` → 0 errors；`npm run build:h5` → DONE Build complete

---

# CHANGELOG — web/mobile

## 2026-08-17 — 加入流程重构：点「加入小区」立即建 membership，填写房号改为独立步骤（用户拍板）

### 分诊
- **A join-community onSelectArea 立即加入（有逻辑函数，异步时序）**：点「加入小区」→ `await joinCommunity(communityId)`（无房号）→ `communityStore.addCommunity` 立即进 store → 把 `membership.id` 回填 pending-join（新增 `membershipId` 字段）→ navigateTo join-choice。已加入 → toast「该小区已加入」不重复；maxReached → 上限警告；joinCommunity 失败 → toast 错误（如「每年最多加入 3 个新小区」）可重试。TDD RED→GREEN
- **B joinCommunity 房号可选（`src/api/user.ts`，有逻辑函数：条件载荷）**：building/unit/room/ownership 改为可选，仅传入时透传（并行管线已支持房号可选），纯 communityId 调用时载荷不含房号字段。TDD RED→GREEN（RED 摘录见 `_tdd_evidence.md` §RED 2.5：`git checkout HEAD -- src/api/user.ts` 回退旧实现复现 `AssertionError: expected { community_id: 'c1', …(4) } to not have property "building"`）
- **C join-residence confirmJoin 改 bindResidence + applyRole（有逻辑函数）**：加入已在上一步完成，本页独立步骤：读 `pending.membershipId`（无则 `getUserMemberships` 按 communityId 回退）→ `bindResidence({membership_id, building, unit, room, is_primary:1})` + `applyRole({community_id, role_code: ownership===1?'owner':'tenant'})` → 成功后 toast「房号登记成功」+ 清 pending-join + switchTab notice。bindResidence/applyRole 失败：提示 + 保留 pending-join 可重试。TDD RED→GREEN
- **D join-choice 页顶文案（字段/文案映射）**：「加入进行中」→「已加入 XX，请选择下一步」（保留社区名/地址），移除「请选择身份完成加入」引导提示（已加入，语义过时）。测试绿，无 RED
- **E pending-join membershipId 字段（字段映射类）**：`PendingJoin` 增加 `membershipId?: string`。测试绿，无 RED

### 做了什么
- `join-community.vue`：`onSelectArea` 改异步立即加入（joinCommunity 无房号 → addCommunity → savePendingJoin 带 membershipId → 导航 join-choice）；失败/已加入/上限三路提示
- `api/user.ts`：`joinCommunity` 四参可选 + 条件透传；`join-form.ts` 注释同步新模型（joinFormToPayload 供 bindResidence/applyRole 用）
- `join-residence.vue`：`confirmJoin` 移除 joinCommunity，改 bindResidence + applyRole；按钮「确认加入」→「确认登记」；loading「登记中...」
- `join-choice.vue`：页顶「已加入 XX，请选择下一步」+ 保留社区名/地址
- `pending-join.ts`：`PendingJoin` 增加 `membershipId?: string`，注释同步新模型

### 测试（+6，144→150）
- `join-community.spec.ts`：点击即 join（无房号）+ addCommunity + 存 pending 带 membershipId + 导航；已加入 toast；上限警告；join 失败 toast（共 5 用例）
- `join-residence.spec.ts`：bindResidence(is_primary:1) + applyRole(owner/tenant)；membershipId 回退 getUserMemberships；找不到成员 toast；bindResidence 失败保留 pending 可重试（共 7 用例）
- `join-choice.spec.ts`：页顶「已加入 XX，请选择下一步」
- `pending-join.spec.ts`：membershipId 往返 + 缺省不含该字段（+2）
- `api/user.spec.ts`：仅 communityId 调用载荷不含房号字段（+1）

### TDD 证据
- RED 摘录见 `_tdd_evidence.md` 新增章节（join-community joinCommunity 0 调用 / join-residence bindResidence 0 调用 / join-choice header 文案不符 / **user.ts joinCommunity 条件载荷（payload 不含房号字段，§RED 2.5 回退 HEAD 复现 `not to have property "building"`）** 等真实 vitest FAIL）

### 门禁
- `npx vitest run` → 24 files / 150 tests PASS（基线 144 → +6）；`npm run type-check` → 0 errors；`npm run build:h5` → DONE Build complete
- `harness-checks-frontend.sh --service mobile` → 6 PASS / 0 FAIL / 2 WARN（存量 type-safety 3 处 as any + api-field-align 34 处）

---

# CHANGELOG — web/mobile

## 2026-08-17 — join-choice 页 UX 修复：明确「加入进行中」（用户误以为到达即已加入）

### 做了什么
- **根因**：用户搜索小区点「加入」后到达 join-choice（2 个身份选项）页，页顶「加入小区 金辰富海」让用户误以为已加入，实际还需选身份 + 填房号才真正加入。
- **修复**：`join-choice.vue` 页顶改为「加入进行中」+ 社区名 + 引导提示「请选择身份完成加入——选择『填写房号成为业主』并登记房号后即完成加入；『其他身份认证』用于申请网格员/管理员等身份」。
- **流程验证**：前端 join-residence 填表提交 → joinCommunity → membership 落库成功（实测 361985 bind_status=1），加入流程本身无 bug。

### 门禁
- `npx vitest run` → 24 files / 144 tests PASS；`harness-checks-frontend.sh --service mobile` → 6 PASS / 0 FAIL / 2 WARN（存量）

---

# CHANGELOG — web/mobile

## 2026-08-17 — 加入小区：省市县自动级联 + 已加入反馈（去掉"下一步"）

### 做了什么
- **省市县自动级联**：`join-community.vue` 选完省立即进入城市列表、选完市立即进入县区列表，移除步骤 1/2 的「下一步」按钮（步骤 3 的「搜索小区」保留——那是搜索动作）
- **已加入反馈**：`onSelectArea` 对已加入小区由静默返回改为 toast「该小区已加入」（此前用户点了没反应误以为失败）
- **排查结论（未改后端）**：加入未完成的最可能根因是后端「每年最多加入 3 个新小区」限流（99500）——实测新用户 joinCommunity 成功、老用户命中限流；toast 已带错误文案

### 记忆沉淀
- 新建 `join-auto-grant-vs-frontend-reapply-role-mismatch` / `one-shot-pending-consume-on-success`（Review 建议落盘，memory-refs 转绿）

### 门禁
- `npx vitest run` → 24 files / 144 tests PASS；`harness-checks-frontend.sh --service mobile` → 6 PASS / 0 FAIL / 2 WARN（存量）

---

# CHANGELOG — web/mobile

## 2026-08-17 — 加入小区流程改造：选小区后分流（join-choice）+ 业主路径新页（join-residence）

### 分诊
- **A pending-join 契约模块（`src/utils/pending-join.ts`，有逻辑函数）**：跨页一次性「待加入小区」{communityId, communityName, address} 的唯一契约源。参考 reg-pending 模式：模块级内存态为主 + 仅 H5 镜像 sessionStorage（`{data, expiresAt}`，TTL 30 分钟，非敏感数据可放宽），绝不 localStorage。TDD RED→GREEN
- **B community store pendingCommunityId（`src/stores/community.ts`，字段映射类 + set/clear 函数）**：新增 `pendingCommunityId: ref('')` + `setPendingCommunityId` / `clearPendingCommunityId`。TDD RED→GREEN（含 set/clear 分支）
- **C join-community 选小区分流（`src/pages/join-community/join-community.vue`，有逻辑分支）**：去掉选小区后的「自有/租住 + 楼单元房号」modal（showJoinForm/joinTarget/confirmJoin/closeJoinForm 及样式、success-card 死代码一并清理）；选中小区改为 `savePendingJoin({id,name,address})` + `navigateTo('/pages/join-choice/join-choice')`。TDD RED→GREEN
- **D join-choice 新页（`src/pages/join-choice/join-choice.vue`，导航分支）**：顶部显示待加入小区名（读 pending-join）；【填写房号成为业主】→ navigateTo join-residence（pending 随行）；【其他身份认证】→ `setPendingCommunityId(id)` + `switchTab('/pages/my/my')`（我的页申请角色用）。TDD RED→GREEN
- **E join-residence 新页（`src/pages/join-residence/join-residence.vue`，有逻辑函数 confirmJoin）**：自有/租住 + 楼/单元/房号（复用 `validateJoinForm` / `joinFormToPayload` / `OWNERSHIP_OPTIONS`）；确认加入 → `joinCommunity(id, building, unit, room, ownership)` → `applyRole({community_id, role_code:'owner'})`（失败不阻塞加入成功，可稍后重新申请）→ `addCommunity` → 清 pending-join → toast 加入成功 + switchTab 首页。TDD RED→GREEN
- **F my.vue applyForRole 兼容 pending 小区（`src/pages/my/my.vue`，有逻辑分支）**：currentCommunityId 为空时回退 `pendingCommunityId`（若存在）并一次性清除；仍为空才提示「请先加入小区」。TDD RED→GREEN

### 做了什么
- 新增 `src/utils/pending-join.ts`（+5 用例 pending-join.spec.ts）
- 新增页面 `src/pages/join-choice/join-choice.vue`、`src/pages/join-residence/join-residence.vue`，并注册 `src/pages.json`（选择身份 / 登记房号）
- `join-community.vue`：移除 modal 及 dead success-card，选中小区存 pending-join 后导航 join-choice
- `src/stores/community.ts`：新增 pendingCommunityId 及 set/clear
- `my.vue applyForRole`：current 优先、pending 回退（一次性消费）
- `join-form.ts` 保留不动（join-residence 复用）
- TDD 证据见 `_tdd_evidence.md` 新增章节

### 门禁
- `npx vitest run` → 24 files / 144 tests PASS（基线 21/125 → +3 文件 / +19 用例）
- `npm run type-check` → 0 errors；`npm run build:h5` → DONE Build complete
- `harness-checks-frontend.sh --service mobile` → 6 PASS / 0 FAIL / 2 WARN（存量）

---

# CHANGELOG — web/mobile

## 2026-08-17 — 注册超时 UX 修复：账号可能已创建时不再卡死（10002/超时分流）

### 背景诊断
后端 register 实测 200ms 正常（日志：POST /api/auth/register → 200，创建 userId）；前端 30s 超时属一次性请求停滞（dev 环境 WSL2/代理抖动），且**账号实际已创建、一次性短信验证码已消费**——原前端 catch 只报「注册失败」导致用户卡死无法继续。

### 做了什么
`src/pages/agreement/agreement.vue` confirmRegister catch 按错误分流：
- **10002（手机号已注册，user-service CreateUser）** → 清临时数据 + toast「该手机号已注册，请直接登录」+ 1s 后返回登录页
- **timeout/网络中断（请求结果不确定，账号可能已创建）** → toast「注册超时，账号可能已创建；请重试或返回登录」+ 保留临时数据可重试（重试命中 10002 自动转登录）
- **其他业务错误** → 保持「注册失败，请重试」

### 新增测试（+2）
- 10002 → 清临时数据 + navigateBack 回登录 + 不调 handleAuthSuccess
- timeout → 提示账号可能已创建、保留数据、不调 handleAuthSuccess

### RED 摘录（QA 修复 2026-08-17 补：上一阶段判 FAIL — 有逻辑函数 RED 缺失）
- 复现：`git stash push -- src/pages/agreement/agreement.vue` 回退至 HEAD（catch 仅「注册失败，请重试」，无 10002/timeout 分支；spec 新增断言不受 stash 影响）→ `npx vitest run src/pages/agreement/agreement.spec.ts` 捕获真实 FAIL → `git stash pop` 恢复实现
- 真实 FAIL（vitest 实际输出）：
  - `手机号已注册(10002)` → `AssertionError: expected "vi.fn()" to be called with arguments: [ ObjectContaining{…} ]`，Received `{ icon: 'none', title: '注册失败，请重试' }`（期望 toast「该手机号已注册，请直接登录」），agreement.spec.ts:202
  - `注册超时（timeout）` → 同型 AssertionError，Received `{ icon: 'none', title: '注册失败，请重试' }`（期望 toast「注册超时，账号可能已创建；请重试或返回登录」），agreement.spec.ts:223
  - `Test Files 1 failed (1) / Tests 2 failed | 5 passed (7)`（完整摘录见 `_tdd_evidence.md` §24）

### 门禁
- `npx vitest run` → 24 files / 144 tests PASS（当前工作树全量；本修复 +2 用例）
- `harness-checks-frontend.sh --service mobile` → 6 PASS / 0 FAIL / 2 WARN（存量）

---

# CHANGELOG — web/mobile

## 2026-08-17 — 注册协议页改版：顶部提示 + 协议只读滚动框 + 确认注册固定可见

### 做了什么
- `src/pages/agreement/agreement.vue`：
  - 顶部新增提示语「您尚未注册账号，请阅读使用协议，勾选同意后注册。注册通过后，您可以登记为业主，也可以认证为网格员等其他身份，解锁 App 功能。」
  - 协议正文改为**只读可滚动文本框**（带边框/圆角/卡片底，`scroll-y` 内部滚动；`min-height:0` + `overflow:hidden` 允许 flex 收缩）
  - 页面 `height: calc(100vh - var(--window-top, 0px))` 扣除导航栏高度，**footer（勾选 + 确认注册）固定屏内可见**，不滚动即可点确认注册（Playwright 实测 footer/按钮在视口内、body 不再滚动）
- 配套：`unit_standard` 检查项 + `unit-system.spec.ts` 允许 `var(--window-top, 0px)` 兜底 0px（同 env() fallback 例外），不误报

### 门禁
- `npx vitest run` → 21 files / 123 tests PASS；`harness-checks-frontend.sh --service mobile` → 6 PASS / 0 FAIL / 2 WARN（存量）

---

# CHANGELOG — web/mobile

## 2026-08-17 — 运行时报错修复：退出登录 401 本地兜底 + user_app_state 迁移落库

### 做了什么
- **退出登录 401 卡死修复**：根因 = access token TTL 900s（15 分钟），用户会话超时后 logout 接口 401，且 refresh 未恢复 → 原 onLogout catch 只 toast 不登出。改为**后端注销尽力而为 + 本地登出兜底**：接口失败仍清本地 token/user + reLaunch 登录页（用户点退出就是要退出；token 已失效时本地清理即正确）。
- **user_app_state 表缺失（切换小区报错）**：migration `005_add_user_app_state.sql` 未执行 → 手动 `docker exec mysql` 落库（问题在运行环境，非代码）。切换小区恢复正常。

### 门禁
- `npx vitest run` → 21 files / 123 tests PASS（logout 失败分支断言改为「仍本地登出」）
- `harness-checks-frontend.sh --service mobile` → 6 PASS / 0 FAIL / 2 WARN（存量）

---

# CHANGELOG — web/mobile

## 2026-08-17 — Review 跟进：logout 清 user_phone 兜底缓存 + account-security 孤儿页清理

### 做了什么
- **logout 清登录期兜底缓存（should-follow）**：`src/stores/user.ts` `logout()` 增加 `uni.removeStorageSync('user_phone')`——登录/注册经 handleAuthSuccess 写入的 user_phone 兜底字段，退出时一并清除，防共享设备跨账号串号泄漏（`SEE: [[logout-clear-login-cache]]`，记忆文件已落盘）。my.spec 退出用例补断言：退出后 `getStorageSync('user_phone')` 为空。
- **account-security 孤儿页清理（should-follow）**：my.vue 已删账号安全入口（上轮），本轮从 `src/pages.json` 移除 `pages/account-security/account-security` 注册（无任何导航引用，且该页残留不调后端的老 logout 逻辑，不可达可深链触达行为不一致）。页面文件保留（未来复用可重新注册）。

### 门禁
- `npx vitest run` → 21 files / 123 tests PASS；`npm run type-check` → 0 errors
- `harness-checks-frontend.sh --service mobile` → 6 PASS / 0 FAIL / 2 WARN（存量）

---

# CHANGELOG — web/mobile

## 2026-08-16 —【我的】页图标网格重构 + 手机号显示修复 + 退出登录接线

### 分诊
- **A 手机号显示修复（`src/utils/auth-flow.ts`，有逻辑函数）**：`user_phone` 从未被写入（死兜底）+ 后端 profile 对本人也可能脱敏 → 登录/注册成功后把用户输入的手机号写入 storage。`handleAuthSuccess` 增加 `opts.phone` 可选参数，login.vue / agreement.vue 传入。TDD RED→GREEN
- **B 退出登录接线（`src/api/identity.ts` + `src/pages/my/my.vue`）**：`logout(deviceId, kickAllDevices?)` API 为纯接线（POST body 透出，免 RED）；my.vue `onLogout` 为异步分支流程（showModal 确认→调接口→清 token→reLaunch），TDD RED→GREEN
- **C 业主认证分支（`src/pages/my/my.vue`，有逻辑函数）**：新增 `onOwnerAuth`——已有已认证业主角色（verf_status=2）→ toast「已是业主」不重复申请；否则 `applyForRole('owner')`。TDD RED→GREEN
- **D【我的】页布局重构（模板/样式，字段映射类）**：展开式 menu-section → title + 4 列图标网格（参考首页 notice.vue `.func-entries`）；纯布局，无需 RED

### 做了什么
- **手机号兜底**：`auth-flow.ts handleAuthSuccess` 增加 `opts.phone`，在 profile 拉取**前**即 `uni.setStorageSync('user_phone', phone)`（profile 失败也不丢失手机号）；login.vue SMS 登录成功分支、agreement.vue 注册成功分支传入 `phone`。my.vue `displayPhone` 的 storage 兜底由此获得真实手机号（前端不依赖后端 profile 脱敏修复时序）
- **退出登录接线**：`api/identity.ts` 新增 `logout(deviceId, kickAllDevices = false)` → `POST /api/auth/logout {deviceId, kickAllDevices}`（后端已存在，需 JWT）。my.vue 账号管理 section 新增「退出登录」图标入口：`uni.showModal` 确认 → `await logout(getDeviceId())` → `userStore.logout()`（清 token/user）→ `uni.reLaunch('/pages/login/login')`（=退出页）；取消分支不动作；接口失败 toast「退出登录失败」（固定中文文案，不取 e.message 原文）并保持登录态。退出后 `isLoggedIn`（token 权威）自动 false → my.vue 回到未登录态
- **【我的】页重构**：4 个可展开 menu-section → 每个 section 一个 title + 4 列图标网格（icon 44rpx→1.375rem + label 24rpx→0.75rem，卡片底 #FAF8F5）
  - 小区管理：加入小区 / **查看退出**（原「退出小区」改名）
  - 业主/租户登记：业主登记 / 租户登记（保留 bind-residence 弹窗与「请先加入小区」禁用态）
  - **新增身份**（原「身份认证」改名）：业主认证（**新增**，`applyForRole('owner')`，与网格员等一致）+ 网格员 / 物业管理员 / 社区管理员 / 商家认证（沿用 applyForRole）；删除业委会入口
  - 账号管理：仅「退出登录」（**删除** 个人信息 / 账号安全 / 关于我们 入口，含 `goAccountSecurity` / `showDevToast` 相关代码）
- **未改动**：web/common/、api-proto/、pages.json、account-security 页面文件（仅删入口）

### 新增测试（RED → GREEN）
- `src/utils/auth-flow.spec.ts`（+2 用例）：`opts.phone 提供 → 登录成功后写入 uni storage user_phone`；`opts.phone 未提供 → 不写入`
- `src/api/identity.spec.ts`（新建 +2 用例，纯接线免 RED）：logout POST `/api/auth/logout {deviceId, kickAllDevices}`
- `src/pages/my/my.spec.ts`（新建 +5 用例）：确认退出 → `logout(getDeviceId())` + `clearTokens` + reLaunch 登录页；取消 → 不动作；logout 失败 → toast「退出登录失败」保持登录态；未认证业主 → `applyRole({community_id, role_code:'owner'})`；已有业主角色 → toast「已是业主」不重复申请
- RED 摘录：`expected "setStorageSync" to be called with arguments: [ 'user_phone', '13800001111' ] Number of calls: 0` / `logout is not a function` / `wrapper.vm.onLogout is not a function` / `wrapper.vm.onOwnerAuth is not a function`（完整摘录见 _tdd_evidence.md §23）

### 门禁
- `npm run test:unit` → 21 files / 123 tests PASS（+3 文件 / +9 用例）
- `npm run type-check` → 0 errors
- `npm run build:h5` → PASS（DONE Build complete）
- `harness-checks-frontend.sh --service mobile` → 6 PASS / 0 FAIL / 2 WARN（存量，未新增）

### QA 修复（2026-08-16 补）
- 上轮 QA 判 FAIL：`onOwnerAuth`（有逻辑函数）RED 摘录缺失（CHANGELOG RED 摘录 3 条均不含它；`_tdd_evidence.md` 无本轮章节，声称「完整摘录见 _tdd_evidence.md」引用悬空）。
- 修复：`git stash push` 回退 my.vue 至 HEAD 无 `onOwnerAuth` 态 → `npx vitest run src/pages/my/my.spec.ts` 捕获真实 FAIL（`TypeError: wrapper.vm.onOwnerAuth is not a function`，my.spec.ts:142/152）→ `git stash pop` 恢复 → 补录 `_tdd_evidence.md` §23 + 本行 RED 摘录（`wrapper.vm.onOwnerAuth is not a function`）。实现与测试无改动，5/5 全绿无回归。

---

# CHANGELOG — web/mobile

## 2026-08-16 — IDOR 修复 + unit_standard 检查项作用域（Review must-follow 跟进）

### 做了什么
- **viewer_id IDOR 修复（must-follow）**：`src/pages/user-detail/user-detail.vue` `load()` 原 `const viewerId = options?.viewer_id || userStore.userId` 从 URL query 取访问者身份（攻击者可手工构造 `?viewer_id=他人` 看他人真实手机号/房屋号）。改为 `const viewerId = userStore.userId || undefined`——访问者身份一律以已认证用户为准，数据范围/脱敏决策权威在后端。配 `SEE: [[api-accessor-identity-from-url]]`。
- **沉淀记忆**：新建 `.harness/knowledge/memory/web/api-accessor-identity-from-url.md`（must-follow / pitfall），memory-refs 回归测试转绿。
- **unit_standard 检查项作用域（决策 (b)）**：`harness-checks-frontend.sh` check 8 本轮仅约束 `web/mobile`；`--service pc` 跳过（pc 仍 px 体系，含 Element Plus，后续单独评估）。

### 门禁
- `npx vitest run` → 19 files / 114 tests PASS；`npm run type-check` → 0 errors
- `harness-checks-frontend.sh --service mobile` → 6 PASS / 0 FAIL / 2 WARN（存量）；`--service pc` → unit_standard 跳过 PASS

---

# CHANGELOG — web/mobile

## 2026-08-16 — QA 门禁修复：unit_standard 误报（spec 测试文件 + 块注释续行）

### 分诊
- **有逻辑函数（bash 状态化注释剥离）**：修复 `check_unit_standard`（harness-checks-frontend.sh）两处检查脚本缺陷，TDD RED→GREEN（回归测试 `src/unit-standard-gate.spec.ts` 先 FAIL 复现再修复）

### 修复内容（QA 检查脚本，非业务代码）
- **根因**：本轮工作树新增并首次启用的 `unit_standard` 门禁报告 5 处 rpx/px，逐条核实全部为脚本误报：
  1. 扫描含 `--include='*.ts'` 未排除 `*.spec.*`/`*.test.*` 测试文件 → unit-system.spec.ts:35/38/48/86（守卫测试用例名与正则）被误报；
  2. 注释排除仅匹配行内含 `/*` 的行，漏掉 `/* */` 块注释续行 → App.vue:44 被误报
- **修复**（`.harness/skills/qa/scripts/harness-checks-frontend.sh` `check_unit_standard`）：
  - 改为 `find` 逐文件 + 逐行扫描，追加 `*.spec.*`/`*.test.*` 跳过（对齐 check_type_safety 第 442-443 行写法）
  - 注释剥离状态化：跨行块注释续行一并剔除（对齐守卫测试 stripComments 语义）；`//` 行注释仅在行首/前导空白后剔除（不误伤 `url(http://...)`）；`/* */` 同行闭合与跨行开启均处理
  - 例外仍保留：html 根字号 `font-size: 16px`（含无空格变体）与 `env(safe-area-inset-bottom, 0px)` 兜底，先剥除再判残留
- 生产代码合规性未受影响（上轮单位换算已完成，唯一真实 px 为允许的根字号与 env fallback）

### 新增测试（RED → GREEN）
- `src/unit-standard-gate.spec.ts`（1 用例）：shell 调用 `harness-checks-frontend.sh --service mobile --json`，断言 `unit_standard` 状态为 PASS；含 `HARNESS_RECURSE=1` 递归守卫（检查脚本的 unit_test 步骤会再调 vitest，防无限递归），testTimeout 120s
- RED 摘录：`AssertionError: unit_standard detail: 5 rpx/px violations: web/mobile/src/App.vue:44...; web/mobile/src/unit-system.spec.ts:35...; :38...; :48...; :86...; : expected 'FAIL' to be 'PASS'`

### 门禁
- `npm run test:unit` → 19 files / 114 tests PASS（+1 文件 / +1 用例）
- `npm run build` → PASS（DONE Build complete）
- `bash .harness/skills/qa/scripts/harness-checks-frontend.sh --service mobile` → 6 PASS / 0 FAIL / 2 WARN（exit 0；WARN 为既有存量：3 处 `as any` 与 web/pc api-field-align）
- `bash .harness/skills/qa/scripts/harness-checks-frontend.sh --service pc` → unit_standard 仍报告 pc 真实 px 违规（pc 未做 rem 换算，属既有存量，不在本轮范围）

---

## 2026-08-16 — 单位体系改造：全部长度/字号 rpx、px → rem（根字号 16px 固定，锚定 375px）

### 分诊
- **纯机械替换（字段映射类）**：无逻辑函数、无组件/函数新增。样式单位全局换算，按任务规则「字段映射类只需测试绿、无需 RED 摘录」，新增守卫测试（`src/unit-system.spec.ts`）验证替换完整性与换算正确性，无需 RED→GREEN 循环
- 换算规则（用户已拍板，固定根字号 16px 非响应式）：`N rpx → N/32 rem`、`N px → N/16 rem`，锚定 375px 设计稿（1rpx=0.5px）

### 做了什么
- **全量替换（脚本批量 + 人工复核 diff）**：`src/` 下全部 `.vue` 与 `.scss` 共 604 处 `rpx` + 30 处 `px` → `rem`
  - 含 `<style>` 与模板内联 style（如 my.vue `style="margin-top: 24rpx;"` → `0.75rem`）
  - uni.scss 变量换算：`$uni-border-radius: 8px→0.5rem`、`$uni-border-radius-card: 12rpx→0.375rem`、`$uni-border-radius-btn: 48rpx→1.5rem`；`$uni-spacing-sm/md/lg/xl: 8/16/24/32px → 0.5/1/1.5/2rem`；`$uni-font-size-xs~xl: 10~20px → 0.625~1.25rem`；`$uni-shadow-sm/base` 的 px 偏移 → rem
- **设置根字号**：App.vue 全局 `<style>` 增加 `html { font-size: 16px; }`（固定，不加 JS 动态缩放），并附换算注释
- **保留项**：`100vh`/`100vw`、百分比、`env(safe-area-inset-bottom, ...)` 内 fallback、rgba() 内 0 值均未动；`calc()` 内长度正确换算（如 `calc(3.125rem + env(safe-area-inset-bottom))`）

### 新增测试（守卫，字段映射类免 RED）
- `src/unit-system.spec.ts`（4 用例）：① src 下 .vue/.scss 无 `rpx` 残留 ② 无长度 `px` 残留（仅允许 html 根字号 16px 与 env() fallback）③ uni.scss 变量换算值为 rem（逐变量断言）④ App.vue 含 `html { font-size: 16px }`
- `npm run test:unit` → 18 files / 113 tests PASS（+1 文件 / +4 用例）

### 门禁
- `npm run type-check` → PASS（0 errors）
- `npm run test:unit` → 18 files / 113 tests PASS
- `npm run build:h5` → PASS（DONE Build complete）

---

## 2026-08-16 — 首页首载去重（REQ-DBL）+ 登录 toast 合并（REQ-TOAST）

### 分诊
- **A 首页 watch 双重加载（src/pages/notice/notice.vue）**：初始进入首页时 onMounted 先 `await loadMemberships()` 再显式 `loadAll()`，而 loadMemberships 内 getAppState 服务端权威覆写 currentCommunityId 会触发 `watch(currentCommunityId)→loadAll()` 一次，同批接口（通知+寻失）被拉两遍。**有逻辑函数**（异步时序守卫），TDD RED→GREEN（`_tdd_evidence.md` §22.1）
- **B 登录 toast 覆盖（src/utils/auth-flow.ts）**：profile 拉取失败时先弹「获取用户资料失败」，随后立即被「登录成功」(icon:success) 覆盖，用户无法得知资料未同步。**有逻辑函数**（分支合并 + toast 时序），TDD RED→GREEN（§22.2）

### 做了什么
- **A**：`watch(currentCommunityId)` 加 `membershipsResolved` 首载守卫——标志默认 `false`，watch 在 `!membershipsResolved` 时直接 return（忽略 loadMemberships 内 getAppState 覆写触发的那次变更）；标志在 notice.vue 的 onMounted 中 `await loadMemberships()` **之后**置 `true`（评审钉死：禁止放入 loadMemberships 内部含 finally，否则覆写触发时标志已 true → 双重加载依旧）；随后 `hasCommunities == true` 才显式单次 `loadAll()`；无小区（含 loadMemberships 整体失败降级）不发请求、直接结束骨架屏展示「请先加入小区」空态（不以陈旧 cid 发请求，既有空态行为不回归）。用户手动切换小区时标志已 true → watch 正常单次触发；pull-refresh 逻辑不变；不丢首次渲染数据
- **B**：profile 拉取失败 → `profileFailed` 标记 + `console.error` 留痕，末尾改为**单条合并 toast**「登录成功，但资料加载失败」（icon:none，文字型非 success 打勾，避免长文案截断）；不再各自弹出两条 toast 导致失败提示被成功提示覆盖；成功路径仍显示「登录成功」(icon:success)。文案不承诺自动恢复（profile 恢复仅发生在 App.vue onLaunch / mine 页面懒加载）。跳转流程（switchTab 首页 / redirectTo 加入小区）与 onCompleted 时序不回归

### 新增测试（RED → GREEN）
- `src/pages/notice/notice.spec.ts`（+3 用例，REQ-DBL-1/2/3）：loadMemberships 覆写 C1→C2 时通知/寻失接口各仅请求一次（以 C2 为维度）；用户手动切换小区 → watch 正常触发单次 loadAll；loadMemberships 整体失败 → 不以陈旧 cid 发请求、无 double-load
- `src/utils/auth-flow.spec.ts`（+1 新 + 改 1）：profile 失败 → showToast 恰 1 次且为合并文案（icon:none）、绝无纯净「登录成功」(icon:success)；新增成功路径用例断言纯净 success toast

### TDD 证据（RED 摘录）
- A：`AssertionError: expected "vi.fn()" to be called 1 times, but got 2 times`（覆写场景 getNoticeList 2 次）+ `expected "vi.fn()" to not be called at all, but actually been called 1 times`（失败降级仍以陈旧 c1 请求）
- B：`AssertionError: expected "vi.fn()" to be called 1 times, but got 2 times`（showToast 2 次，失败提示被覆盖）
- 全量摘录已持久化至 `_tdd_evidence.md` §22

### 门禁
- `npm run type-check` → PASS（0 errors）
- `npm run test:unit` → 17 files / 109 tests PASS（+4 新用例）
- `npm run build:h5` → PASS（DONE Build complete）
- `bash .harness/skills/qa/scripts/harness-checks-frontend.sh --service mobile` → 5 PASS / 0 FAIL / 2 WARN（WARN 均为既有存量：3 处 `as any` 与 web/pc api-field-align，非本轮引入）

---

## 2026-08-16 — Review WARNING 跟进：登录/注册双提交窗口 + 错误分支 code 优先 + 文档漂移 + crypto 调试残留

### 分诊
- **A 双提交漏洞（login.vue / agreement.vue）**：成功分支 `submitting` 复位移入 `handleAuthSuccess` 的 `onCompleted` 回调（跳转完成后才复位），封堵跳转窗口期二次点击触发重复登录/注册。**有逻辑函数**（异步时序分支），TDD RED→GREEN（`_tdd_evidence.md` §21）
- **B 登录错误分支 code 优先（login.vue handleSubmit catch）**：`err.code` 数值为主判据，msg 字符串匹配仅作 code 缺失时的旧后端兜底。**有逻辑函数**（条件判定重构），TDD RED→GREEN（§21）
- **C 设计文档漂移（docs/design.md）**：登录/注册段落与页面表更新为现行实现（协议注册页 + reg-pending 契约模块）。纯文档，不需测试
- **D crypto.ts 调试残留（src/utils/crypto.ts）**：删除 3 处 `console.log`（58/64/66 行，QA debug_artifacts 证据漂移为 0）。删代码，不需测试，修后 `grep -rnE "console\.(log|debug)" src` 0 命中

### 做了什么
- **A**：`login.vue` 成功分支删除前置 `submitting.value = false`，改为 `await handleAuthSuccess(loginRes, { onCompleted })`（onCompleted 内复位）；`agreement.vue` confirmRegister 同样处理；所有错误路径（50001 分支 / saveRegPending 失败 / catch 异常）的 submitting 复位保持不变
- **B**：`login.vue` catch 中三条件并存改为：`code !== undefined ? code === 50001 : (msg.includes('50001') || msg.includes('未注册'))`——code 存在时只看 code（例如 code=10040 即使 msg 含"未注册"也不进注册流程），code 缺失才回退 msg。保持原意图：仅 50001 未注册进入协议注册流程
- **C**：design.md 页面表 `pages/register/register` → 现行 `pages/agreement/agreement`（协议注册）等；新增「登录 / 注册流程（现行实现）」小节：登录页无协议勾选 → loginWithSms 失败 50001 → navigateTo 协议页 → register API → handleAuthSuccess 自动登录；说明 reg-pending 机制（内存态主载体 + H5 sessionStorage 镜像 TTL 5 分钟，绝不 localStorage）与共享契约模块 `src/utils/reg-pending.ts`
- **D**：删除 `crypto.ts` 中 `[Crypto] Using cached public key` / `[Crypto] Fetching public key...` / `[Crypto] Raw response...` 三处 console.log，保留 console.error（错误留痕，QA 允许）

### 新增测试（RED → GREEN）
- `src/pages/login/login.spec.ts`（+4 用例）：A onCompleted 时序（跳转完成前 submitting=true，onCompleted 后复位）+ onCompleted 以第二参数传入；B code=10040 且 msg 含"未注册"→ code 为主不进注册流程；B code 缺失 msg 含"未注册"→ msg 兜底进注册流程；B code 缺失 msg 无特征 → 复位返回
- `src/pages/agreement/agreement.spec.ts`（+1 用例）：A 确认注册成功 → submitting 跳转完成前保持 true，onCompleted 后复位

### TDD 证据（RED 摘录）
- A（login + agreement）：`AssertionError: expected "vi.fn()" to be called with arguments: [ ObjectContaining{…}, …(1) ]`（handleAuthSuccess 未传 onCompleted 第二参数）
- B：`AssertionError: expected "vi.fn()" to not be called at all, but actually been called 1 times`（saveRegPending 被误调——code=10040 仍进注册流程）
- 全量摘录已持久化至 `_tdd_evidence.md` §21

### 门禁
- `npm run type-check` → PASS（0 errors）
- `npm run test:unit` → 17 files / 105 tests PASS（+5 新用例）
- `npm run build:h5` → PASS（DONE Build complete）
- `bash .harness/skills/qa/scripts/harness-checks-frontend.sh --service mobile` → 5 PASS / 0 FAIL / 2 WARN（WARN 均为既有存量：3 处 `as any` 与 web/pc api-field-align，非本轮引入；debug_artifacts 现为真实 0）

---

## 2026-08-16 — 修复多视角审查 CRITICAL：补齐 3 个悬空记忆引用（创建记忆文件 + 引用完整性回归测试）

### 分诊
- 新建 `src/utils/memory-refs.spec.ts`：**有逻辑函数**（扫描 src/ 全部 `// SEE: [[slug]]` 引用，断言每个 slug 能在项目/个人记忆目录解析到 .md 文件，防悬空引用回归），TDD RED→GREEN
- 新建 3 个记忆文件于 `.harness/knowledge/memory/web/`：`sms-code-persist-localstorage.md` / `frontend-cross-page-storage-contract.md` / `cross-page-sensitive-temp-data-storage.md`：文档（无逻辑代码），记忆遵守闭环
- `.harness/knowledge/memory/MEMORY.md` + `.memory-index.json` 同步登记新记忆
- 测试基建（供 memory-refs.spec.ts 使用 node 内置模块）：`package.json` devDependencies 新增 `@types/node`（与 web/pc 对齐）；`tsconfig.app.json` `types` 数组追加 `"node"`

### 做了什么（standards-eng 视角 M2 CRITICAL 修复）
- 上一阶段 multi-review 规范工程视角 FAIL（1 CRITICAL）：`reg-pending.ts:16-18` / `login.vue:71-73` / `agreement.vue:56-58` 引用了 3 个不存在的记忆 slug（`[[sms-code-persist-localstorage]]` / `[[frontend-cross-page-storage-contract]]` / `[[cross-page-sensitive-temp-data-storage]]`），M2 规则「slug 文件不存在 → 🔴 CRITICAL」
- 按审查建议 (a) 创建记忆文件（代码行为本身正确——一次性 smsCode 走内存态 + sessionStorage TTL 5 分钟、不落 localStorage、key/结构收敛到单一契约源，值得沉淀）：三记忆均为前端存储安全/契约指南
  - `[[sms-code-persist-localstorage]]`（must-follow / pitfall）：一次性验证码禁止落 localStorage 持久化残留，仅 H5 镜像 sessionStorage + TTL
  - `[[frontend-cross-page-storage-contract]]`（should-follow / guideline）：跨页共享 key/结构收敛到单一共享模块，禁止两端各写 magic string
  - `[[cross-page-sensitive-temp-data-storage]]`（should-follow / guideline）：跨页一次性敏感数据优先内存态载体，非必要不持久化
- `memory-index-build.sh` 重建倒排索引（60 条记忆）+ MEMORY.md 索引登记

### 新增测试（RED → GREEN）
- `src/utils/memory-refs.spec.ts`（新建 3 用例）：扫描 src/ 提取全部 SEE slug 断言可解析 / 修复目标 3 slug 必须存在 / 测试自身能扫到引用。**RED 证据**：`expected [] to deeply equal [ "sms-code-persist-localstorage", "frontend-cross-page-storage-contract", "cross-page-sensitive-temp-data-storage" ]` + `悬空记忆引用: [[sms-code-persist-localstorage]]: expected false to be true`（2 FAIL，GREEN 后 3 PASS）

### 门禁
- `npm run test:unit` → 17 files / 100 tests PASS（+3 新用例）
- `npm run build:h5` → PASS（DONE Build complete）
- `bash .harness/skills/qa/scripts/harness-checks-frontend.sh --service mobile` → 5 PASS / 0 FAIL / 2 WARN（WARN 均为既有存量，非本轮引入）

### 追加：lockfile 加固（Review WARNING 跟进）
- 本轮 `npm install`（补 `@types/node`）重生成 `package-lock.json` 时，npmmirror 镜像在顶层 hoisted 了一个**非官方版本 `lodash@4.18.1`**（官方 lodash 止于 4.17.21；whatwg-url 声明 `^4.7.0` 被镜像解析成 4.18.1）。已在 `package.json` 加 `"overrides": { "lodash": "4.17.21" }` 并重生成 lock：`lodash@4.18.1` 全仓消失，改为 whatwg-url 下嵌套的规范 `4.17.21`。无生产代码直接依赖 lodash，无功能影响

---

## 2026-08-16 — 登录协议流程安全加固：reg_pending 一次性注册数据收敛共享契约模块（消灭 localStorage 持久化残留）

### 分诊
- **新建 `src/utils/reg-pending.ts`**（唯一契约源 `REG_PENDING_KEY` + `RegPending` + `saveRegPending`/`readRegPending`/`clearRegPending`）：**有逻辑函数**（内存态主载体 + H5 sessionStorage 镜像 + TTL 5 分钟过期校验 + localStorage 零触碰），TDD RED→GREEN（`_tdd_evidence.md` §19）
- `login.vue` 删内联 `REG_PENDING_KEY` + JSON.stringify → 调 `saveRegPending(regPending)`：接线（无新逻辑分支），测试改断言
- `agreement.vue` 删内联 `REG_PENDING_KEY` + `readPending`/`clearPending` 手写实现 → 用 `readRegPending()`/`clearRegPending()`：接线（无新逻辑分支），测试改断言
- `login.spec.ts` / `agreement.spec.ts`：断言共享模块行为，不再依赖 `uni.setStorageSync('reg_pending')` magic string

### 做了什么（Review 3 条 memory suggestion 修复）
- **1. 消灭 smsCode 持久化残留**：原 `uni.setStorageSync` 在 H5 即映射 localStorage，一次性验证码 + 手机号会残留可被共享设备复用。现改模块级内存变量为主载体，仅 H5 镜像到 `window.sessionStorage`（`{data, expiresAt}`，TTL 5 分钟，过期即清并返回 null，save 先清旧再写，sessionStorage 访问 try/catch 容错），**绝不 localStorage**。非 H5 环境统一走内存态
- **2. 收敛 magic string**：`reg_pending` key 与 `RegPending` 结构此前在 login.vue 与 agreement.vue 各手写一份，一端改名另一端静默失效。现收敛到 `src/utils/reg-pending.ts` 单一契约源，两端 import 共享
- **3. 跨页一次性敏感数据改内存态载体**：页面栈内导航（login → agreement）由模块级内存变量直接传递，不落任何持久化
- `login.vue` 存储失败处理保留（toast + 不复位 submitting）；`agreement.vue` onLoad 判空回退逻辑保留

### 新增测试（RED → GREEN）
- `src/utils/reg-pending.spec.ts`（新建 5 用例）：save→read 往返 / H5 镜像写 sessionStorage 带 `{data,expiresAt}` 且断言 localStorage 零调用 / TTL 5 分钟过期返回 null 并清除镜像 / clear 后内存+镜像一并清除返回 null / 空数据返回 null
- `src/pages/login/login.spec.ts`（改断言）：50001 分支断言 `saveRegPending` 以 `{phone,smsCode,deviceId,nickname}` 调用一次 + navigateTo；非 50001 / 成功分支断言 `saveRegPending` 不被调用
- `src/pages/agreement/agreement.spec.ts`（改断言）：确认注册断言 `readRegPending` 被调 + `register` 正确参数 + `clearRegPending` 调用一次；注册失败断言 `clearRegPending` 不被调（可重试）；无临时数据断言 `readRegPending` 返回 null 提示失效

### TDD 证据（RED 摘录）
- `reg-pending.spec.ts`：`Error: Failed to resolve import "./reg-pending" ... Does the file exist?`（模块未实现）
- `login.spec.ts`：`AssertionError: expected "vi.fn()" to be called 1 times, but got 0 times`（saveRegPending 未调用）
- `agreement.spec.ts`：`AssertionError: expected "vi.fn()" to be called at least once`（readRegPending 未调用，3 用例 FAIL）
- 全量摘录已持久化至 `_tdd_evidence.md` §19

### 记忆应用
- `[[sms-code-persist-localstorage]]` — 一次性验证码禁止落 localStorage 持久化残留（H5 镜像仅 sessionStorage + TTL）
- `[[frontend-cross-page-storage-contract]]` — 跨页共享 key/结构收敛到单一共享模块，禁止两端各写 magic string
- `[[cross-page-sensitive-temp-data-storage]]` — 跨页一次性敏感数据优先内存态载体
- 三记忆在 `src/utils/reg-pending.ts` / `login.vue` / `agreement.vue` 均以 `// SEE:` 标注落地

### 门禁
- `npm run test:unit` → 16 files / 97 tests PASS（+5 新用例）
- `npm run type-check` → PASS（0 errors）
- `npm run build:h5` → PASS（DONE Build complete）

---

## 2026-08-16 — 修复多视角审查 CRITICAL：两个空 catch 补日志留痕（禁止静默吞错）

### 分诊
- `src/stores/community.ts` `loadMemberships` getAppState catch 补 `console.error`：日志留痕（接线，无新逻辑分支），配回归测试断言
- `src/utils/auth-flow.ts` `handleAuthSuccess` getUserMemberships catch 补 `console.warn`：日志留痕（接线，无新逻辑分支），配回归测试断言

### 做了什么
- **1. getAppState 失败留痕**：`community.ts:90` 空 catch → `console.error('[community] getAppState 获取失败，降级本地', e)`，回退逻辑保留——修复 app-state 接口故障时零 trace、『服务端权威当前小区』静默失效、运维/QA 无从察觉的问题
- **2. getUserMemberships 失败留痕**：`auth-flow.ts:46` 空 catch → `console.warn('[auth-flow] 小区检查失败，默认加入小区', e)`，默认跳 join-community 逻辑保留——修复 membership 接口故障时已有小区用户被静默误导且零日志信号的问题
- 两处均 `// SEE: [[verify-api-before-calling]]`（该记忆『怎么做』要求至少打日志，禁止 `try{await someApi()}catch{}` 坏模式）

### 新增测试（RED → GREEN）
- `src/stores/community.spec.ts`（改 1 用例）：getAppState 请求失败 → `console.error` 断言留痕（`AssertionError: expected "error" to be called … Number of calls: 0`）+ 降级本地回退不受影响
- `src/utils/auth-flow.spec.ts`（+1 用例）：getUserMemberships 失败 → `console.warn` 断言留痕 + 默认 redirectTo 加入小区（原无该失败分支覆盖）

### TDD 证据（RED 摘录）
- 两个新断言真实 vitest FAIL 摘录：`expected "error" to be called with arguments: [ StringContaining{…}, Any<Error> ]` / `expected "warn" to be called … Number of calls: 0`（均因修复前不打日志而失败）

### 记忆应用
- `[[verify-api-before-calling]]` — 禁止空 catch 静默吞错，catch 至少打日志（本次两处修复直接执行该记忆『怎么做』）

### 门禁
- `npm run test:unit` → 15 files / 92 tests PASS（+1 新用例）
- `npm run type-check` → PASS（0 errors）
- `npm run build:h5` → PASS（DONE Build complete）

---

## 2026-08-16 — 首页通知列表两行布局改版（标题全文显示 + 发布单位 + 日期 / 移除 JS 字符宽度截断）

### 分诊
- `notice.vue` `formatPublishDate`（YYYY-MM-DD 转换，published_at=0 回退 created_at）：**有逻辑函数**，TDD RED→GREEN
- `notice.vue` `getPublisherName`（publisher 非空优先、空回退 getNoticeRoleName）：**有逻辑函数**，TDD RED→GREEN
- 移除 `noticeDisplayTitle`/`formatMonthDay`/`text-fit`（JS 字符宽度截断整条链路删除）：字段映射/纯删除，测试改断言，无独立逻辑

### 做了什么
- **1. 标题全文显示**：`notice.vue` 通知卡片标题改直接渲染 `item.title`（去掉 `noticeDisplayTitle` JS 截断与 `white-space: nowrap`），`white-space: normal` + `word-break: break-all` 自然换行，不截断、无省略号
- **2. 标题下方元信息行**：新增 `.notice-meta` 行渲染发布单位 + 发布日期——发布单位：`item.publisher` 非空则用它、否则回退 `getNoticeRoleName(item.role)`（`getPublisherName`）；发布日期：`formatPublishDate(item.published_at || item.created_at)` → `YYYY-MM-DD`（published_at=0 时回退 created_at）
- **3. 简洁样式**：`.notice-card` 去掉卡片底色 `$uni-bg-color-card`/圆角/阴影，透明底 + 细分隔线（`border-bottom`）区分行；行首保留左侧角色色条 `getNoticeRoleColor`；移除角色 pill（`.notice-role-pill`）
- **4. 移除 text-fit**：`notice.vue` 删除 text-fit import 与截断逻辑；`src/utils/text-fit.ts` + `text-fit.spec.ts` 无任何引用 → **整文件删除**，并清理 CHANGELOG 相关记录（见下条「通知单行紧凑」历史项修订）

### 新增测试（RED → GREEN）
- `src/pages/notice/notice.spec.ts`（改版断言）：两行布局标题全文渲染（20 字标题 + 发布单位「物业管理处」+ YYYY-MM-DD 日期）/ publisher 空回退角色名 / published_at=0 回退 created_at 渲染 YYYY-MM-DD / 无 `.notice-line` 与 `.notice-role-pill` 残留；「超长标题 JS 截断日期恒显示」断言删除，「published_at=0 回退 (M-D)」断言改为 YYYY-MM-DD

### 历史项修订
- 下条「移动端 6 项完善」中：分诊「`formatMonthDay`/`noticeDisplayTitle`（字符宽度截断）」已删除（本改版移除）；「新增测试」移除 `text-fit.spec.ts` 条目；TDD 证据 §10 text-fit RED 摘录归档为已删除模块

### TDD 证据（RED 摘录）
- 4 个新/改断言真实 vitest FAIL 摘录已持久化至 `_tdd_evidence.md` §18（标题截断 `expected '小区停水通知…' to be '小区停水通知：因管道检修明日上午9点至下午5点停水'` / `.notice-meta` 空 wrapper / `.notice-line` 残留 true to be false）

### 记忆应用
- `[[tdd-red-evidence-requires-fail-excerpt]]` — 新逻辑函数配真实 RED→GREEN 摘录
- `[[verify-before-deliver]]` — 改后全量门禁验证（type-check / test:unit / build:h5 / frontend QA）

### 门禁
- `npm run test:unit` → 全绿（含 notice.spec.ts 24 用例）
- `npm run type-check` → PASS（0 errors）
- `npm run build:h5` → PASS（DONE Build complete）
- `harness-checks-frontend.sh --service mobile` → 全项 PASS / 0 FAIL（既有 WARN 存量非本轮引入）

---

## 2026-08-16 — 移动端 6 项完善（TabBar 改名 / 登录协议流程 / 登录态修复 / 通知模块改造）

### 分诊
- `handleAuthSuccess`（`src/utils/auth-flow.ts`，profile 失败分支 + 小区跳转）：**有逻辑函数**，TDD RED→GREEN
- login.vue `handleSubmit`（50001 未注册分支 → 暂存 + 跳协议页）：**有逻辑函数**，TDD RED→GREEN
- agreement.vue `confirmRegister`（checkbox 校验 + 读 storage + register + 清数据 + 自动登录）：**有逻辑函数**，TDD RED→GREEN
- stores/user.ts `isLoggedIn`（token 权威化）：**有逻辑函数**，TDD RED→GREEN
- stores/community.ts `loadMemberships`（getAppState 服务端权威采用/降级）：**有逻辑函数**，TDD RED→GREEN
- notice.vue `formatMonthDay`/`noticeDisplayTitle`（字符宽度截断 + M-D 日期转换）：**有逻辑函数**，TDD RED→GREEN ~~（已由「首页通知列表两行布局改版」删除，见上方条目 §18）~~
- notice.vue `onCommunitySwitch`（非 10015 → console.error + 通用 toast）：**有逻辑函数**，TDD RED→GREEN
- App.vue `restoreUserProfile`（onLaunch 登录态恢复：isAuthenticated 守卫 + user 缓存守卫 + getUserProfile try/catch）：**有逻辑函数**，TDD RED→GREEN（QA 分诊补测，见 `_tdd_evidence.md` §17）
- 字段映射/纯接线：pages.json TabBar 改名（公告信息→我的小区）+ navigationBarTitleText 同步 + 注册 `pages/agreement/agreement`

### 做了什么
- **1. TabBar 改名**：`src/pages.json` tabBar.list[0].text「公告信息」→「我的小区」；`pages/notice/notice` 的 navigationBarTitleText 同步改为「我的小区」
- **2. 登录流程改造**：`login.vue` 移除「已阅并同意《使用协议》」勾选区（agreed/toggleAgreement/showAgreement 删除，canSubmit 仅依赖手机号+验证码）；提交先 `loginWithSms`，仅当 code 50001（未注册）→ 暂存 `{phone,smsCode,deviceId,nickname:'用户'+phone后4位}` 到 uni storage `reg_pending` → `uni.navigateTo('/pages/agreement/agreement')`；其他错误保持拦截器 toast。新建 `src/pages/agreement/agreement.vue`（已在 pages.json 注册）：展示《社区家园使用协议》正文 + checkbox + 确认注册；未勾选点确认 → toast「请先阅读并同意使用协议」；确认注册 → 读临时 storage → `register` → 成功清临时数据 + 走 `handleAuthSuccess`（注册完成自动登录一次）；失败 toast + 保留临时数据可重试
- **3. 登录态修复**：`stores/user.ts` `isLoggedIn = computed(() => isAuthenticated())`（token 权威，user 是 profile 缓存）；`App.vue onLaunch` 若已登录但 user 未加载 → `getUserProfile()` → `setUser()` 全局恢复（移除 console.log）；`handleAuthSuccess` profile 拉取失败 → `console.error` + toast「获取用户资料失败」，仍继续跳转（token 已存，页面懒加载再恢复）
- **4. 通知模块**：`notice.vue` 删除跑马灯（.marquee-bar/marqueeText/marquee CSS/动画），原位替换为标题栏（📢 通知公告 + 更多）；「更多」= onMoreNotice，移除 `notices.length===0` 拦截（空态也进浏览页）；列表 `v-if="notices.length>0"`，为空不渲染列表与空态块，仅保留标题栏下方内容自然上移
- **5. 通知单行紧凑**：`[色条] 标题 (M-D) [角色pill]` 单行布局；标题按字符宽度 JS 截断（新增 `src/utils/text-fit.ts`，CJK≈28rpx/半角≈14rpx，末尾加 …），容器可用宽度≈750-64(页面 padding)-10(色条)-48(body padding)=628rpx，日期恒显示；新增 `formatMonthDay(ts)` → `(M-D)`（published_at=0 回退 created_at）；列表 gap→10rpx、卡片 padding 缩小；角色 pill 保留行尾 ~~（`src/utils/text-fit.ts` 与 `formatMonthDay`/`noticeDisplayTitle` 已由「首页通知列表两行布局改版」删除）~~
- **6. 切换小区出错**：`notice.vue onCommunitySwitch` 非 10015 → `console.error` + toast「切换小区失败」（不再静默）；`stores/community.ts loadMemberships` 加载后调 `getAppState()`，后端 current_community_id 存在于 memberships → 采用并保存（服务端权威、跨设备一致，修复本地 storage 陈旧导致切换/显示不一致）；getAppState 失败/缺失降级忽略

### 新增测试（RED → GREEN）
- `src/stores/user.spec.ts`（3 用例）：isLoggedIn 以 token 为权威（user=null 仍 true / 无 token user 缓存仍 false / token+user true）
- `src/utils/auth-flow.spec.ts`（3 用例）：profile 成功 setUser + 无小区 redirectTo / profile 失败 console.error+toast 仍 switchTab / 有小区 switchTab
- `src/pages/login/login.spec.ts`（4 用例）：协议勾选区已移除 canSubmit 仅手机号+验证码 / 50001 暂存+跳协议页 / 非 50001 不暂存不跳 / 登录成功走 handleAuthSuccess
- `src/pages/agreement/agreement.spec.ts`（4 用例）：未勾选 toast 不调 register / 勾选 register 正确参数+清数据+自动登录 / 注册失败 toast+保留数据 / 无临时数据提示失效
- `src/stores/community.spec.ts`（+4 用例）：loadMemberships 服务端权威采用并保存 / getAppState 0 降级本地回退 / getAppState 失败容错忽略 / 服务端值不在 memberships 忽略
- `src/pages/notice/notice.spec.ts`（更新 +5 用例）：跑马灯移除标题栏「通知公告+更多」/ 空态不渲染列表与空态块 / 更多空态进浏览页 / 单行紧凑渲染 / 超长标题 JS 截断日期恒显示；更新「非 10015 → toast」「双请求失败无空态块」「published_at=0 回退 created_at → (M-D)」断言（「单行紧凑渲染 / 超长标题 JS 截断日期恒显示」与 `(M-D)` 断言已由「首页通知列表两行布局改版」改为两行布局断言，见上方条目）
- `src/App.spec.ts`（4 用例，QA 分诊补测）：restoreUserProfile 未登录不调 / user 未加载则 getUserProfile+setUser / user 已加载跳过 / getUserProfile 失败 console.error 不抛错

### TDD 证据（RED 摘录）
- 8 个有逻辑函数真实 vitest FAIL 摘录已持久化至 `_tdd_evidence.md` §9-§17（`expected false to be true` / `Failed to resolve import` / `expected '' to contain '物业'` / `TypeError: restoreUserProfile is not a function` 等）

### 记忆应用
- `[[frontend-business-rule-hardcode]]` — 当前小区权威在后端（app-state），前端以服务端为准；前端只传参/消费
- `[[verify-before-deliver]]` — 改后全量门禁验证（test:unit 90 / type-check / build:h5 / frontend QA）
- `[[snake-camel-field-mismatch]]` — 无新 snake_case 契约字段（登录协议流程不涉及 API 契约变更）

### 门禁
- `npm run test:unit` → 16 files / 94 tests PASS
- `npm run type-check` → PASS（0 errors）
- `npm run build:h5` → PASS（DONE Build complete）
- `harness-checks-frontend.sh --service mobile` → 全项 PASS / 0 FAIL / 2 WARN（既有 `as any` ×3 + PC 端 api_field_align 存量，非本轮引入）

---

## 2026-08-16 — 首页信息架构改造（mobile-homepage-content-revamp Task 2.1-2.6）

### 分诊
- **Task 2.1** `isImageAttachment`（file_type 图片白名单谓词，分支逻辑）：**有逻辑函数**，TDD RED→GREEN
- **Task 2.2** 通知区 `getNoticeList(cid,1,3,30)` 传参 / 4 功能入口点击分发（跳页 vs toast）：**有逻辑函数**，TDD RED→GREEN
- **Task 2.3** 区块垂直全序 + 邻里互助占位 + 广告集中：**布局调整**，组件测试补全序断言
- **Task 2.4** notice-browse 30 天卡片列表（fetch since_days + 失败态分支）：**有逻辑函数**，TDD RED→GREEN
- **Task 2.5** notice-detail 附件点击 file_type 分发（图片/文档/空 url）：**有逻辑函数**，TDD RED→GREEN
- **Task 2.6** contact-list 新建（fetch + 失败态 + 拨号点击）：**有逻辑函数**，TDD RED→GREEN
- 字段映射/纯接线：`NoticeAttachment` 扩展 `file_id`/`file_type`（REQ-NDP-4）、`getNoticeList` 增 `since_days` 参数、pages.json 注册 contact-list

### 做了什么
- **2.1** `src/api/community.ts`：`NoticeAttachment` 扩展 `file_id?`/`file_type?`（snake_case 对齐 wire，optional-safe 缺省不崩溃 REQ-NDP-4 场景 2）；`getNoticeList` 新增 `sinceDays?: number` 参数 → `params.since_days`（仅 >0 传参，缺省不传保持 PC 行为）；新增 `IMAGE_FILE_TYPES=['png','jpg','jpeg','gif']` + `isImageAttachment(fileType?)`（与 `file-service` `guard/magic.go` 白名单对齐，`// SEE: [[frontend-business-rule-hardcode]]`）
- **2.2** `src/pages/notice/notice.vue`：通知区改调 `getNoticeList(cid, 1, 3, 30)`（since_days=30 + page_size=3，30 天窗口后端强制）；**移除首页内嵌联络拨号网格**（fetchContacts/contacts/contactGroups/onCallPhone/样式）；通知下方新增 4 个功能图标入口（便民联络 → `uni.navigateTo('/pages/contact-list/contact-list')` 做实跳页；物业报修/二手闲置/租房卖房 → `uni.showToast('功能开发中')` 不跳转，REQ-FE-1/2/3）；未加入小区不渲染入口区
- **2.3** `notice.vue` 区块垂直全序重排（REQ-HL-4）：通知（跑马灯+卡片）→ 4 功能入口 → **邻里互助占位**（「互助功能开发中」，点击不导航、不伪造数据，REQ-HL-1）→ 寻失互助（样式数据不动，REQ-HL-2）→ **底部广告集中区**（3 个广告垂直堆叠，内容保留硬编码，点击预留不跳转，REQ-HL-3）；移除联络骨架屏
- **2.4** `src/pages/notice-browse/notice-browse.vue`：从单条翻页浏览改为 30 天卡片列表（`getNoticeList(cid, 1, 50, 30)` 单请求，REQ-NTW-4）；渲染与首页一致的卡片（role 色条/role 标签/标题/时间，REQ-NTW-5 视觉契约）；**移除客户端 3 个月过滤**（窗口由后端强制）；点卡片 → `uni.navigateTo` 详情；空态「暂无通知公告」+ 加载失败明确提示（不静默）；移除 currentIndex/翻页逻辑
- **2.5** `src/pages/notice-detail/notice-detail.vue`：附件点击按 `isImageAttachment(att.file_type)` 白名单分发（REQ-NDP-2/3）：图片 → `uni.previewImage` 全屏预览（失败 toast「预览失败」，不降级文档打开器）；文档（pdf/doc/docx 或缺失/无法识别）→ `uni.downloadFile` 成功后 `uni.openDocument`（失败 toast「附件打开失败」）；**消费详情响应重生 file_url，不直连 file-service REST**（所有权限制，REVISION #4）；移除 H5 `window.open` 平台分支；详情加载失败态「加载失败」/「通知不存在」区分；无附件不渲染附件区
- **2.6** 新建 `src/pages/contact-list/contact-list.vue`：`onLoad` 读当前小区 → `getContacts(cid)` 渲染拨号网格（类别图标/类别名/电话，样式沿用首页原联络网格），点击 → `uni.makePhoneCall`（REQ-CLP-1）；空态「暂无联络信息」+ 加载失败明确提示；`pages.json` 注册 `pages/contact-list/contact-list`（navigationBarTitleText 便民联络）

### 新增测试（RED → GREEN）
- `src/api/community.spec.ts`（3 用例）：isImageAttachment 白名单 true / 文档 false / 缺失 undefined 不崩溃
- `src/pages/notice/notice.spec.ts`（新增 9 用例，共 17）：since_days=30&page_size=3 调用 / 4 入口按固定顺序渲染 / 便民联络 navigateTo / 占位入口 toast 不跳转 / 未加入小区不渲染 / 区块全序（func-entries→邻里互助→寻失→底部广告）/ 双请求失败 console.error 恰好 2 次（联络网格移除后由 3 变 2）
- `src/pages/notice-browse/notice-browse.spec.ts`（5 用例）：since_days=30&page_size=50 / 卡片契约渲染 / 点卡片进详情 / 加载失败态 / 空态
- `src/pages/notice-detail/notice-detail.spec.ts`（8 用例）：图片 previewImage / 文档 downloadFile+openDocument / 图片空 url 预览失败 / 文档空 url 附件打开失败 / 无法识别按文档 / 无附件不渲染 / 加载失败态 / 通知不存在
- `src/pages/contact-list/contact-list.spec.ts`（4 用例）：getContacts 渲染网格 / 点击 makePhoneCall / 空态 / 加载失败态

### TDD 证据（RED 摘录）
- 8 个有逻辑函数（isImageAttachment / onFuncEntry / onAttachmentClick / fetchDetail / notice-browse onMounted / formatTime / fetchContacts / onCall）真实 vitest FAIL 摘录已持久化至 `_tdd_evidence.md` §1-§8（`git stash` 回退生产文件复现，含 `TypeError: isImageAttachment is not a function` / `expected +0 to be 4` / `Number of calls: 0` / `ReferenceError: fetchContacts is not defined` 等）

### 基础设施
- `vitest.setup.ts` uni stub 增补 `previewImage`/`downloadFile`/`openDocument`/`makePhoneCall`（附件预览 + 拨号测试）

### 记忆应用
- `[[frontend-business-rule-hardcode]]` — file_type 白名单与 file-service guard/magic.go 对齐；30 天窗口后端强制前端只传参
- `[[snake-camel-field-mismatch]]` — NoticeAttachment file_id/file_type snake_case 对齐 wire
- `[[verify-api-before-calling]]` — 各 API 路由 graph-context 已确认；失败明确提示不静默吞错
- `[[change-verification-checklist]]` — 改后全量门禁验证（test:unit/type-check/build）

### 门禁
- `npm run test:unit` → 10 files / 62 tests PASS
- `npm run type-check` → PASS（0 errors）
- `npm run build` → PASS（DONE Build complete）

---

## 2026-08-15 — 字段名对齐后端 snake_case JSON tag（Task 3.4）

### 分诊
- `community.ts` 三套 interface 字段名（camelCase → snake_case）：**字段映射/纯接线**（对齐后端 community-hub-service `types.go` JSON tag），无 RED 摘录
- `notice.vue` / `notice-detail.vue` / `notice-browse.vue` 消费点同步改名：**字段映射/纯接线**，无 RED 摘录

### 做了什么
- **3.4** `src/api/community.ts` 四套 interface 字段名对齐后端 `types.go` JSON tag（snake_case）：
  - `Notice`：`communityId→community_id`、`publisherId→publisher_id`、`isPinned→is_pinned`、`publishedAt→published_at`、`createdAt→created_at`、`updatedAt→updated_at`
  - `NoticeAttachment`：`fileName→file_name`、`fileUrl→file_url`、`fileSize→file_size`
  - `Contact`：`communityId→community_id`、`sortOrder→sort_order`
  - `LostFoundItem`：`communityId→community_id`、`imageUrls→image_urls`、`contactPhone→contact_phone`、`publisherId→publisher_id`、`createdAt→created_at`
- 同步消费点：`notice.vue`（`item.published_at`/`created_at`/`image_urls`）、`notice-detail.vue`（`notice.published_at`/`created_at`、`att.file_name`/`file_url`/`file_size`）、`notice-browse.vue`（`currentNotice.published_at`/`created_at`、`n.published_at`/`created_at`）
- 后端 `types.go` 零改动（契约方）
- **3.3 门禁**：`npm run test:unit` 35 PASS、`npm run type-check` 0 error、`npm run build` PASS

### 新增测试
- `src/pages/notice/notice.spec.ts` 新增 1 用例：`snake_case 字段渲染：通知 created_at 回退 + 寻失 image_urls[0]`（mock 镜像后端 JSON tag，验证改名后渲染不坏）

### 记忆应用
- `[[snake-camel-field-mismatch]]` — Go snake_case 与 TS camelCase 字段名不匹配

---

## 2026-08-15 — 寻失列表路径对齐 + 静默 catch 消除（Task 3.1-3.3）

### 分诊
- `community.ts` `getLostFoundList` 请求路径 `/api/community/lost-found` → `/api/community/lostfound`：**字段映射/纯接线**（单路径字符串对齐后端已注册路由），无 RED 摘录
- `notice.vue` 三处 `catch { /* silent */ }` → `console.error`（错误对象 + 区块标识）+ `uni.showToast`（icon:'none'）：**有逻辑函数**（错误处理分支），TDD RED→GREEN

### 做了什么
- **3.1** `src/api/community.ts:156` 请求路径对齐后端 `GET /api/community/lostfound`（无连字符）；后端 community-hub-service 零改动。全仓 web 源码无 `/api/community/lost-found` 调用残留（仅 `services/*/docs/graph-context.md` Neo4j 自动生成文档含旧路径，非调用点，随下一次 graph 同步刷新，不在本服务范围）
- **3.2** `src/pages/notice/notice.vue` 三处静默 catch 消除：失败时 `console.error('[notice] <区块>加载失败', e)` + `uni.showToast({ title: '<区块>加载失败', icon: 'none' })`，不 rethrow；成功路径不弹错误 toast
- **3.3** 门禁：`npm run test:unit` 34 PASS（含 notice.spec.ts 新增 6 用例）、`npm run type-check` 0 error、`npm run build` PASS

### 新增测试（RED → GREEN）
- `src/pages/notice/notice.spec.ts` 新增 6 用例（`notice page — fetch 静默 catch 消除`）：
  - `getLostFoundList`/`getNoticeList`/`getContacts` 各自失败 → 对应区块 toast + `console.error`
  - 三请求并发全部失败 → toast ≥1 次 + `console.error` 恰好 3 次 + loading 复位页面不崩
  - 局部失败（通知失败 + 其余成功）→ 成功区块数据仍渲染
  - 成功场景 → 无错误 toast

### TDD RED 证据
- RED：4 个失败路径用例在实现前运行 FAIL（`uni.showToast` Number of calls: 0 / `toHaveBeenCalled` at least once 失败），GREEN 后 9/9 通过。摘录见本节（`expected "vi.fn()" to be called ... Number of calls: 0`）。

### 记忆应用
- `[[change-verification-checklist]]` — 改后全量门禁验证（test:unit/type-check/build）
- `[[verify-api-before-calling]]` — 调用 API 前验证路由存在（lostfound 无连字符对齐后端注册路由）

---

## 2026-08-13 — access-control 前端 TDD 缺口修复（Task 7.1-7.3 补测）

### 分诊纠正
上一轮将 4 个「有逻辑函数」误标为「字段映射/纯接线」而漏测、漏 RED。本轮逐一分诊纠正：
- `switchCommunity`（`stores/community.ts`）：async + 后端持久化失败抛错 + 本地 `currentCommunityId` 不变 → 有逻辑函数
- `onCommunitySwitch`（`pages/notice/notice.vue`）：`e?.code === 10015` 条件分支 toast 提示 → 有逻辑函数
- `request.ts` 响应拦截器：业务错误 `code` 附加 + 成功解包 / 无 code 透传 → 有逻辑函数
- `validateJoinForm`（`join-form.ts`）：去硬编码区间的校验逻辑 → 有逻辑函数

### 新增测试（RED → GREEN）
- `src/stores/community.spec.ts`（3 用例）：成功后持久化+更新本地+落 storage / 失败 10015 本地不变 / 非成员 no-op
- `src/pages/notice/notice.spec.ts`（3 用例）：10015 专属 toast / 非 10015 不 toast / 成功不 toast 且更新
- `src/utils/request.spec.ts`（5 用例）：code===0 解包 / 无 code 透传 / 业务错误 code 附加+toast / 非数字 code 置 undefined
- `join-form.spec.ts` 已有 10 用例，本轮补 RED 摘录

### TDD RED 证据
4 个函数均按 must-follow 记忆 `tdd-red-evidence-requires-fail-excerpt` 真实重放 RED（临时回退实现到 HEAD 基线 → vitest 捕获真实 `AssertionError` / `TypeError` / `expected` / `Received` → 恢复实现），持久化于 `_tdd_evidence.md` §1-§4。

### 记忆应用
- `[[tdd-red-evidence-requires-fail-excerpt]]` — RED 摘录为 vitest 实际输出
- `[[testing-discipline]]` — 有逻辑函数按正常/空值/边界/错误路径覆盖
- `[[frontend-business-rule-hardcode]]` — 切换校验/区间权威在后端，前端仅 UX 提示

### 门禁
- `npx vitest run` → 6 files / 28 tests PASS
- `npm run type-check` → PASS（0 errors）
- `npm run build` → PASS

> 关于「50007 跳过 toast」：该行为归属 web/pc 登录/注册页（design §7「端限制登录引导」），移动端 `request.ts` 不消费 50007，不在本服务 7.1-7.3 范围。移动端响应拦截器「有逻辑」部分为业务错误 `code` 附加 + 成功解包 / 无 code 透传，均已测试。

---

## 2026-08-13 — access-control 前端（Task 7.1-7.3）

### 做了什么
- **7.1 加入小区楼/单元/房号必填**（`src/pages/join-community/join-form.ts`）
  - `validateJoinForm` 去除硬编码区间（楼号 1-150 / 单元 1-5 / 房号 100-999），改为仅「必填 + 正整数」UX 提示，权威校验对齐后端（JoinCommunity 拒绝 `<=0` → 10040）
  - `join-form.spec.ts` 同步更新（区间用例改为 `<=0` 拒绝 / 正数放行），17/17 GREEN
- **7.2 当前小区切换 UI**（`src/api/user.ts` + `src/stores/community.ts` + `src/pages/notice/notice.vue` + `src/utils/request.ts`）
  - 新增 `getAppState()` / `setCurrentCommunity(communityId)` API 封装（GET `/api/users/me/app-state` / PUT `/api/users/me/current-community`）
  - `communityStore.switchCommunity` 改为 async：先调后端 `setCurrentCommunity` 持久化（跨设备一致），成功才更新本地；失败抛错、本地不变
  - `notice.vue onCommunitySwitch` 捕获 `code===10015` → toast「目标小区不在你的数据范围」；切换后首页上下文经既有 watch 刷新
  - `request.ts` 业务错误对象附加 `code` 字段（向后兼容），供调用方分支
- **7.3 同屋互见用户详情**（`src/api/user.ts` + `src/pages/user-detail/user-detail.vue`）
  - 新增 `SameHouseInfo` / `GetUserDetailResult` 类型 + `getUser(id, viewerId)` API（GET `/api/users/:id?viewer_id=xxx`）
  - 新增用户详情页：消费 `same_house`，`same_house=true` 展示楼栋房屋号；手机号由后端按 viewer 上下文决定明文/脱敏，前端仅展示、不做二次脱敏（权威在后端）
  - `pages.json` 注册 `pages/user-detail/user-detail`

### 类型标注
- 原标注「均为字段映射/纯接线、无 RED 摘录」**有误**——`switchCommunity` / `onCommunitySwitch` / `request.ts` 拦截器 / `validateJoinForm` 为有逻辑函数，已在上方「TDD 缺口修复」条目纠正并补齐测试与 RED 证据。
- 记忆应用：`[[frontend-business-rule-hardcode]]`（去硬编码区间 / 切换校验权威在后端）、`[[web-common-type-reuse-no-redefine]]`（用户详情类型复用，不重复定义共享层）

### 门禁
- `npm run type-check` → PASS（0 errors）
- `npm run test:unit` → PASS（3 files / 17 tests）
- `npm run build` → PASS

---

## 2026-08-12 — 加入小区流程携带 ownership（access-data-permission 阶段⑤ T5.1）

### 做了什么
- `src/api/user.ts` — `joinCommunity` 签名改为 `(communityId, building, unit, room, ownership)`，载荷传 `{community_id, building, unit, room, ownership}`
  - ownership 与 user.proto `CommunityOwnership` 对齐：`OWNED=1` / `RENTED=2`（前端仅传 1/2，后端 10040 拒绝 UNSPECIFIED）
  - building/unit/room 与 proto `JoinCommunityRequest` 注释对齐（楼号 1-150 / 单元 1-5 / 房号 3位数字）
- `src/pages/join-community/join-community.vue` — 点击「加入」弹出加入表单，收集「自有/租住」选择（必填）+ 楼/单元/房号输入
  - 校验通过后调用更新后的 `joinCommunity(target.id, building, unit, room, ownership)`；校验失败阻止提交并展示错误提示
  - 取消/遮罩点击关闭表单；加入成功关闭表单并展示成功卡片
- 新增 `src/pages/join-community/join-form.ts` — 纯函数：`OWNERSHIP_OPTIONS` / `validateJoinForm`（权属必填 + 楼/单元/房号区间校验，兼容 string|number 输入）/ `joinFormToPayload`
- 新增 TDD 测试（17 cases）：`src/api/user.spec.ts`（joinCommunity 载荷）、`src/pages/join-community/join-form.spec.ts`（校验全路径）、`src/pages/join-community/join-community.spec.ts`（表单交互流程）
- 新增测试基础设施：`vitest.config.ts`（happy-dom + vue 插件 + @/@common 别名）、`vitest.setup.ts`（uni 全局 stub）、`.gitignore` 忽略 `*.tsbuildinfo`

### 基础设施修复（为通过既有门禁，未改动业务行为）
- `package.json` — `type-check` 改为 `vue-tsc --noEmit -p tsconfig.app.json`（vue-tsc 1.0.24 不支持 solution-style tsconfig 的编译器选项透传）
- `tsconfig.app.json` — `moduleResolution: "node"`（"bundler" 是 TS5.0 特性，与 TS 4.9/vue-tsc 1.x 不兼容）；新增 `skipLibCheck`、`esModuleInterop`；include 补充移动端实际依赖的 3 个 common 文件（identity.ts/auth.ts/config.ts）
- `src/api/identity.ts` — `getUserProfile` 响应类型断言 `as unknown as any`（axios 类型为 AxiosResponse 而拦截器已解包 data，属既有潜在类型错误）
- 依赖环境：`npm install --no-save --no-package-lock` 安装 vitest@4.1.10 / @vue/test-utils / happy-dom，并将 vue 家族统一到 3.4.21 + pinia@2.1.7（与 uni 工具链 compiler-sfc 3.4.21 对齐、TS 4.9 可编译；pinia 2.3.x 要求 vue^3.5 会 ERESOLVE）

### 为什么
后端阶段3 user-service T3.2：`JoinCommunityRequest.ownership` 必填，决定 owner/tenant 角色授权。前端必须携带权属加入，否则后端返回 10040。

### 影响
- Proto: 无变更（契约已在 api-proto 阶段0 提交，子 Agent 无权限）
- 调用方: `join-community.vue` 是 `joinCommunity` 唯一调用方，已同步更新
- 数据库: 无
- 关联: 依赖 user-service JoinCommunity ownership 校验 + AssignRole(owner|tenant)（阶段3 T3.2 已完成）

### 门禁
- `npm run type-check` → PASS（0 errors）
- `npm run test:unit` → PASS（3 files / 17 tests）
- `npm run build` → PASS

### TDD RED 证据（修复轮 2026-08-12）
- 上一轮 QA 判定 TDD RED 证据不足（RED 列无实际 FAIL 输出摘录，仅注释口头描述并错误指向 CHANGELOG）。本轮按 must-follow 记忆 `tdd-red-evidence-requires-fail-excerpt.md` 真实重放 RED 并持久化：
  - `joinCommunity`：临时回退 `src/api/user.ts` 至 HEAD 1 参签名 → vitest 捕获真实 `AssertionError: expected "vi.fn()" to be called with arguments: [ '/api/users/communities/join', …(1) ]`（Received 仅 `{community_id}`，缺 building/unit/room/ownership），2 tests FAIL
  - `validateJoinForm`/`joinFormToPayload`/`OWNERSHIP_OPTIONS`：临时移除 `join-form.ts` → vitest 捕获真实 `Error: Failed to resolve import "./join-form" ... Does the file exist?`，0 tests
  - `confirmJoin`/`openJoinForm`/`closeJoinForm`：`join-form.ts` 缺失时组件挂载失败 `Failed to resolve import "./join-form" from "join-community.vue"`；恢复 `join-form.ts` 后回退 `confirmJoin` 至 HEAD 行为 → vitest 捕获真实 `AssertionError: expected "vi.fn()" to be called with arguments: [ 'c1', 3, 1, 502, 1/2 ]`（Received 仅 `["c1"]`），3 tests FAIL
  - 全部摘录为 vitest 实际输出（`AssertionError` / `Failed to resolve import`），持久化于 `_tdd_evidence.md`（本服务首个 TDD 证据文件）；随后恢复实现，确认 17/17 GREEN + type-check + build 全 PASS

---

## 2026-06-06 — 验证码登录/注册页面 + RSA 加密

### 做了什么
- 实现 `src/pages/login/login.vue`：验证码登录/注册一体化页面
  - 手机号 + 短信验证码 + 协议复选框
  - 先尝试 LoginSms → 失败则自动注册（昵称="用户"+手机尾号4位）
  - SMS 按钮最小 30s 冷却（成功后延长至 60s）
- 实现 `src/utils/crypto.ts`：RSA 加密工具（Web Crypto API，RSA-OAEP + SHA-256，与后端 `common/pkg/crypto/rsa.go` 对齐）
- 实现 `src/utils/device.ts`：设备 ID 生成器（localStorage 持久化）
- 更新 `src/api/identity.ts`：集成 RSA 加密，修复 API 路径（`sms/send`、`token/refresh`）
- 更新 `src/pages/mine/mine.vue`：未登录状态点击跳转登录页
- 更新 `src/pages.json`：注册登录页路由

### 为什么
移动端需要验证码登录/注册入口，RSA 加密是后端 API 的硬性要求（手机号必须加密传输）。

### 影响
- Proto: 无变更
- 调用方: 无
- 数据库: 无
- 关联: 依赖 auth-service 的 `/api/auth/sms/send`、`/api/auth/login/sms`、`/api/auth/register`、`/api/auth/public-key` 接口

## 2026-06-06 — 初始化 Uni-app Vue 3 移动端框架

### 做了什么
- 使用 Uni-app (Vue 3) + Vite + TypeScript 初始化项目脚手架
- 配置 TypeScript 路径别名（`@/*`、`@common/*`，引用 `web/common/`）
- 安装核心依赖：Pinia 2.1.7（状态管理）、Axios（HTTP 客户端）、lossless-json（Snowflake ID 解析）、dayjs（日期格式化）、uni-ui（官方组件库）
- 配置 Vite：uni 插件、API 代理（6 个后端服务）、开发端口 3004
- 配置 `pages.json`：三 Tab 路由（首页/发现/我的）+ TabBar + globalStyle
- 配置 `manifest.json`：H5 模式（hash 路由、端口 3004）
- 编写 `uni.scss`：完整的主题变量体系（品牌色、文字色、背景色、间距、字号、阴影）
- 实现 `src/utils/request.ts`：Axios 实例 + lossless-json Snowflake ID 解析 + Token 自动刷新（并发刷新保护）
- 实现 `src/stores/user.ts`：Pinia 用户状态管理（token、profile、login/logout）
- 实现 `src/api/identity.ts`：认证 API 封装（login、loginWithSms、register、refreshToken、getUserProfile、sendSmsCode）
- 实现三个 Tab 占位页面：首页（含用户状态引用）、发现页、我的（登录/未登录状态切换）
- 创建 `CLAUDE.md`（子 Claude 配置）、`docs/design.md`（设计文档）、`CHANGELOG.md`
- 更新父级 `web/CLAUDE.md` 添加 mobile 子实例索引

### 为什么
社区家园需要移动端访问入口，Uni-app (Vue 3) 提供跨端能力（H5+小程序），复用 `web/common/` 共享类型和工具函数保证与 PC 端一致。

### 影响
- Proto: 无变更（使用现有 API）
- 调用方: 无
- 数据库: 无
- 关联: 框架初始化，为后续移动端功能开发提供基础设施
