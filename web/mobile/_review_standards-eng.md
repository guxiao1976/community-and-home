# Code Review — 移动端（规范工程视角）

**审查时间**: 2026-08-16 20:00
**审查维度**: 规范遵循(#3)、复用性(#6)、测试覆盖(#7)、可观测性(#9)、记忆遵守(#12)
**审查范围**: main 分支工作树未提交改动 + 未跟踪文件（登录协议流程安全加固 reg-pending 收敛 / 悬空记忆引用修复 + memory-refs 回归测试 / 首页通知两行布局 / 登录态修复 / 服务端权威当前小区）

## 摘要
- 🔴 CRITICAL: 0 / 🟡 WARNING: 1 / 🔵 NOTE: 3

## 发现

### 🔴 CRITICAL
| # | 文件:行号 | 维度 | 问题 | 修复建议 |
|---|----------|------|------|---------|
| — | — | — | 无 | — |

### 🟡 WARNING
| # | 文件:行号 | 维度 | 问题 | 建议 |
|---|----------|------|------|------|
| 1 | `web/mobile/src/utils/auth-flow.ts:8-11` | 复用性(#6) + 记忆遵守(#12, should-follow) | `AuthSuccessResult` 接口（accessToken/refreshToken/expiresAt）与 `@common/types/identity` 的 `LoginResponse`（同三字段 + userId）局部重复定义。api/identity.ts 已 import `LoginResponse`，`handleAuthSuccess` 入参实际就是 `loginWithSms`/`register` 返回的 `LoginResponse`，本地再定义子集接口违反 `[[web-common-type-reuse-no-redefine]]`（should-follow）「禁止重复定义 web/common 已有类型」 | 改用 `Pick<LoginResponse, 'accessToken' \| 'refreshToken' \| 'expiresAt'>`（`import type { LoginResponse } from '@common/types/identity'`），或直接以 `LoginResponse` 为入参类型，删除本地 `AuthSuccessResult` |

### 🔵 NOTE
| # | 文件:行号 | 建议 |
|---|----------|------|
| 1 | `web/mobile/src/utils/reg-pending.ts`（整体） | 复用性优秀：key/结构/TTL 逻辑收敛到单一契约源，login.vue / agreement.vue 均只 import，无 magic string。`[[sms-code-persist-localstorage]]`（must-follow）遵守到位：内存态主载体 + 仅 H5 sessionStorage 镜像 + TTL 5 分钟 + localStorage 零触碰 + try/catch 容错 |
| 2 | `web/mobile/package.json` / `_qa.md` | 测试覆盖(#7)：11 组有逻辑函数 100/100 测试全绿且断言真实行为；但 `@vitest/coverage-v8` 未安装，QA 无法量化覆盖率百分比，核心链路 ≥80% 门禁无法机械化验证。建议安装 coverage 工具并纳入 harness-checks-frontend 门禁 |
| 3 | `web/mobile/src/pages/notice/notice.vue:280-291` 等 | 可观测性(#9)：本轮 catch 块已全部改为 console.error/warn 留痕 + toast（无静默吞错），较前版显著改善；但整体仍依赖 console + toast，无统一前端错误上报/关键操作埋点（登录成功/注册完成/切小区等无 analytics）。工程规模下可作为后续技术债项（非阻塞） |

## 记忆遵守检查（M1-M3）

- [x] `// SEE: [[...]]` 引用已验证（共 8 个唯一 slug，均解析到项目或个人记忆文件，无悬空）
- [x] 引用准确性：`sms-code-persist-localstorage`（must-follow）/ `verify-api-before-calling`（must-follow）均被代码遵守；`frontend-business-rule-hardcode` / `frontend-cross-page-storage-contract` / `cross-page-sensitive-temp-data-storage` / `snake-camel-field-mismatch` / `tdd-red-evidence-requires-fail-excerpt` 均准确适用
- [x] M3 遗漏检查：变更关键词（smsCode/localStorage/sessionStorage/登录态/协议注册/storage 契约）未命中未覆盖的 must-follow 记忆；`vue-template-nested-interpolation`（web, must-follow）——agreement.vue 新增模板无嵌套 `{{ }}`，遵守
- [x] M4 记忆更新建议（交 Owner 落库）：`sms-code-persist-localstorage` / `frontend-cross-page-storage-contract` / `cross-page-sensitive-temp-data-storage` / `verify-api-before-calling` / `frontend-business-rule-hardcode` / `snake-camel-field-mismatch` — last_applied/apply_count 建议 +1

---
VERDICT: PASS
---

PASS — 本视角（规范遵循 #3 / 复用性 #6 / 测试覆盖 #7 / 可观测性 #9 / 记忆遵守 #12）无 CRITICAL。1 个 WARNING（`AuthSuccessResult` 与 web/common `LoginResponse` 局部重复定义，should-follow 记忆未完全遵守），修复成本低，不阻塞合入。机械化检查 type_check / unit_test / build / hardcoded_secrets / debug_artifacts 全绿（QA FRESH 复验），新增生产文件 0 处 `as any`、0 处 console.log/debugger、Snowflake ID 全 string、无 res.data 双解包、无嵌套 `{{ }}`。
