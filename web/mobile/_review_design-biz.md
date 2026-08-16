# Code Review — 移动端（设计业务视角）

**审查时间**: 2026-08-16 23:10
**审查维度**: 设计一致性(#2)、代码质量(#4)、Migration(#8部分)

## 摘要
- 🔴 CRITICAL: 0 / 🟡 WARNING: 2 / 🔵 NOTE: 7

## 发现

### 🔴 CRITICAL
无

### 🟡 WARNING

| # | 文件:行号 | 维度 | 问题 | 建议 |
|---|----------|------|------|------|
| 1 | `src/pages/my/my.vue:330,348` / `src/pages/join-community/join-community.vue:357-359` | 代码质量(错误处理) | 双重 toast + 原始 `e.message`：`request.ts` 响应拦截器已对业务错误(code≠0)与网络错误各 toast 一次，页面 catch 又自行 toast 第二次——网络失败时第二次显示英文 `Network Error`，覆盖了拦截器的友好中文提示，违反项目记忆 `[[axios-network-error-raw-message-toast]]`（toast 一律用固定中文文案，不取 e.message 原文）。**既有存量**（本轮 my.vue / join-community.vue 仅单位换算改动，错误处理逻辑未动），但与本轮代码同处审查范围内 | 页面 catch 不再自行 toast（拦截器已兜底）；若确需页面级提示，用固定中文文案且避免与拦截器重复 |
| 2 | `docs/design.md:23-28` | 设计一致性 | 「Tab 页面（底部导航栏）」段仍写 3 个 Tab（`pages/index/index` 首页 / `pages/discover/discover` 发现 / `pages/mine/mine` 我的），实际 `pages.json` 为 **4 个 Tab**（`notice`「我的小区」/ `interact`「邻里互动」/ `mine`「我的家庭」/ `my`「我的」），且 `pages/index`、`pages/discover` 已不存在于 pages.json。本轮更新 design.md 补齐了子页面表与登录流程段落，但 Tab 结构段未同步 → 文档与实际继续漂移 | 同步 design.md Tab 段为现行 4 Tab 结构，并注明首页已由 notice 页承担 |

### 🔵 NOTE

| # | 文件:行号 | 建议 |
|---|----------|------|
| 1 | `src/pages/notice-browse/notice-browse.vue:85-87` | onMounted 直接读 `communityStore.currentCommunityId` 而未先 `await loadMemberships()`。冷启动/直达该页时 currentCommunityId 可能为空或陈旧，会显示误导性「暂无通知公告」空态。建议 onMounted 先 `loadMemberships()` 兜底（当前仅能从 notice 页进入，风险低） |
| 2 | `src/pages/notice-browse/notice-browse.vue:92` | 单请求 `page_size=50` 拉取且丢弃响应 `total`，30 天窗口内通知 >50 条时旧通知不可达（无分页/加载更多）。典型小区量级可接受，属完整性缺口 |
| 3 | `src/pages/login/login.vue:120-143` | 发送验证码失败仍保持 30s 冷却，重试需等 30s。反连点意图可理解，但失败场景可考虑缩短冷却更友好 |
| 4 | `src/utils/reg-pending.ts:80` | `readRegPending` 从 session 镜像恢复时以 `Date.now() + TTL_MS` 重新续期，与「5 分钟 TTL」的字面意图（自保存起算）不符；因数据一次性消费、注册成功后清除，实际风险可忽略 |
| 5 | `src/components/community-switcher.vue:79-83` | `select()` emit `update:modelValue`，但所有父组件均只绑定 `:model-value` + `@switch`，未使用 v-model —— 属死 emit，可移除 |
| 6 | `src/pages/notice/notice.vue:281-283` | `imageErrors` Set 在 `onPullDownRefresh`/`loadAll` 后不清空，曾经加载失败的图片在后续刷新成功后仍渲染占位符，须离开页面重建组件才恢复 |
| 7 | `src/pages/notice/notice.vue:293-307` + `community-switcher.vue` | 切换小区失败（10015）时 dropdown 已立即关闭，用户看不到「当前小区保持不变」的视觉反馈（仅 toast），建议失败时保持面板打开或高亮当前项 |

### 正向确认（本维度通过项）
- 登录/注册流程与 design.md「现行实现」段落完全一致：reg-pending 契约源收敛、TTL 5 分钟、绝不 localStorage、agreement 页 onLoad 无数据提示失效并返回
- API 字段与 api-proto 对齐：`auth.proto`（encrypted_phone / sms_code / device_id / nickname / LoginSmsRequest）、`user.proto`（CommunityMembership 全 snake_case、`verf_status`、CommunityOwnership=1/2、building≤150/unit≤5/room 3位）；my.vue `r.verf_status` 与 proto 字段一致
- uni.scss 主题变量与 design.md 色彩方案一致（#B8956A 咖色系、间距/字号 rem 化）
- loading/empty/error 三态在 notice / notice-browse / notice-detail 均完整覆盖；notice.vue 首载去重守卫（REQ-DBL）与 10015 切换错误处理完善
- Migration #8：agreement 页已注册 pages.json；新增页面无漏注册
- Snowflake ID 全链路 `string`，无 `Number()` 转换；lossless-json 正常
- 测试覆盖充分（reg-pending 5 用例 / notice 首载去重 3 用例 / agreement 5 用例 / App restore 4 分支），QA 全绿（19 files / 114 tests）

---
VERDICT: PASS
---
