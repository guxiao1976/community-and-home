# Request — notice-xss-sanitize-and-frontend-fixes

> 本文件为本变更的标准需求输入（评审曾指出 request.md 缺失，本轮补齐）。用户原始需求 + 已拍板设计决策 + 上轮评审反馈一并收录。

## 用户需求

修复 3 项（XSS 走后端净化 + 2 处前端）。涉及 community-hub-service（后端净化）与 web/mobile（前端）。

### 1. 存储型 XSS 修复（community-hub-service，治本）

现状：`web/mobile/src/pages/notice-detail/notice-detail.vue:23` 用 `<rich-text :nodes="notice.content">` 渲染后端公告 content（H5 端等价 v-html），community-hub 写入路径无 HTML 净化（无 bluemonday/sanitize），内容源（发公告者）注入 `<img onerror>/<iframe>` 即存储型 XSS。

修复（后端写入路径净化）：
- 引入 bluemonday（github.com/microcosm-cc/bluemonday）白名单净化器，对公告/内容类字段的 content 在写入路径（Create/Update notice 等，含 RPC 与 REST）做 HTML 白名单净化（保留安全标签如 p/br/strong/em 等，剔除 script/iframe/img onerror 等），DB 存净化后内容
- 净化器单例化；写测试覆盖：含 `<img onerror>/<script>` 的注入 payload 被剥离、合法富文本标签保留
- 若净化逻辑放 community-common（跨服务共享）需评估影响；如仅 community-hub 用则放本服务

### 2. 首页 watch 双重加载（web/mobile notice.vue）

onMounted await store.loadMemberships(); loadAll() 同时 watch currentCommunityId → loadAll；loadMemberships 内 getAppState 服务端权威覆写 currentCommunityId 时触发 watch 一次 + onMounted 显式一次，同一批接口拉两遍。

### 3. 登录成功 toast 覆盖失败提示（web/mobile auth-flow.ts）

handleAuthSuccess profile 拉取失败先 showToast('获取用户资料失败')，随后 800ms 内又 showToast('登录成功') 覆盖，失败提示不可见。

## 约束

- 不新增公开 API；不改 api-proto（净化在服务内部）
- community-hub 改动需 go build ./... + go test ./... + bash .harness/skills/qa/scripts/harness-checks.sh --service community-hub-service 无 FAIL
- web/mobile 改动需 npm run type-check + test:unit + build:h5 全绿
- 新逻辑函数配 TDD RED→GREEN 摘录
- CHANGELOG 同步

## 已确认的设计决策（用户已拍板）

| 决策 | 结论 | 说明 |
|------|------|------|
| Q1 | 1 | 首页双重加载采用「显式单次加载 + watch 首载守卫」 |
| Q2 | 0 | 登录 toast 失败提示与成功合并到同一提示展示 |
| Q3 | 0 | 净化字段范围仅公告/内容帖正文 content/text（title 等不净化） |
| Q4 | 0 | 净化器放 community-hub-service 本服务（不引 community-common） |
| Q5 | 0 | 不做存量数据回填净化 |
| Q6 | 0 | 正文 img 全剔除（图片走附件机制） |

## 上轮评审反馈（REVISION 原因，本轮必须逐条对照修订）

1. **[clarity] 白名单标签集未穷举且与自身 Scenario 矛盾**：REQ-XSS-2 以「等」留白，但 Scenario 以 `<a href="javascript:">`/`<div style=...>` 断言属性剔除（隐含 a/div 保留）。→ 穷举完整允许标签/属性集（显式含 a、div 去留）+ 补「合法链接保留」正向场景。
2. **[clarity] 净化器与既有非空校验顺序未指定**：REQ-XSS-1 Scenario 3（空正文 080005）与 REQ-XSS-4 Scenario 2（净化后空串落库）对同一输入隐含互相排斥行为。→ 钉死顺序：非空校验以原始正文先行，净化在通过后落库前；净化后为空唯一化（建议接受空串落库）。
3. **[validity] submit 过渡路径绕过净化**：updatecontentpostlogic.go submit 分支直接 UpdateStatusAndPublishTx 置公开，不重写正文、不经过净化器，存量 draft 可经 submit 发布为未净化公开内容。→ (a) 在 submit 置公开前对 post.Text 追加净化（推荐）；或 (b) 扩展残余风险声明。
4. **[validity] 白名单未与实际编辑器产出交叉核对**：后台编辑器 out_of_scope 但它是合法富文本产出源，若产出白名单外标签，存量合法公告一经编辑重存即降级。→ 验收增加抽样比对既有公告正文/编辑器产出与白名单，结论入 CHANGELOG/决策记录。
