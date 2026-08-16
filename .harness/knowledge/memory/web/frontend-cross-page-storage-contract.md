---
triggers: ["跨页", "cross-page", "storage", "存储", "key", "magic string", "契约源", "共享模块", "setStorageSync", "getStorageSync", "localStorage", "sessionStorage", "两端", "改名", "静默失效"]
service: all
severity: should-follow
type: guideline
status: active
created: 2026-08-16
updated: 2026-08-16
last_applied: null
apply_count: 0
---

# 跨页共享存储的 key/结构必须收敛到单一共享模块，禁止两端各写 magic string

## 为什么会有这条经验

多页面（如 login.vue 与 agreement.vue）共享同一份跨页临时数据时，若两端各自手写相同的 storage key 字符串（如 `'reg_pending'`）和重复的对象结构，会形成 magic string 双份：一端改名另一端静默失效、一端加字段另一端解析 undefined，且 QA 很难从两端代码对比发现。此前 reg_pending 的 key 与结构就是两端各写一份。

## 怎么做

1. 新建共享模块（如 `src/utils/reg-pending.ts`）作为**唯一契约源**：导出 key 常量 + 数据结构接口 + save/read/clear 函数
2. 消费方（login / agreement）只 `import` 该模块，禁止内联 magic string 或手写 JSON.stringify/parse
3. 读写逻辑（TTL 校验、过期清除、容错降级）集中在该模块一处实现，两端行为一致
4. 变更 key 或结构时只改契约源，其他页面自动同步

案例：`web/mobile/src/utils/reg-pending.ts` 收敛 `REG_PENDING_KEY` + `RegPending` + `saveRegPending`/`readRegPending`/`clearRegPending`，login.vue 与 agreement.vue 均只 import，不再内联 `'reg_pending'`。

## 怎么验证

- grep 页面代码中的 storage key 字符串，确认没有脱离共享模块的重复定义
- 两端读取逻辑一致（同一 TTL、同一过期行为），非各写一份
- 修改 key 后两端同时生效，无一端静默失效

## 关联经验

[[sms-code-persist-localstorage]] [[cross-page-sensitive-temp-data-storage]]
