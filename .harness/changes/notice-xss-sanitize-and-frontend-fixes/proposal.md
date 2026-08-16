# Proposal: 通知公告存储型 XSS 净化 + 首页双重加载修复 + 登录 toast 覆盖修复

> **优先级**: P1 · **改动规模**: 中 · **影响风险**: 中
> **核心风险点**: ① 存储型 XSS 是安全缺陷（写入路径当前无 HTML 净化，发公告者可注入 `<img onerror>/<iframe>`），须治本（后端白名单净化）且不破坏既有合法富文本公告；② 首载守卫改动 notice.vue 数据加载时序，需保证「初始进入只拉一遍」且不回归拉取失败留痕/下拉刷新。
> **变更类型**: modify（3 项对既有行为/既有功能的修复）

## 为什么做

**安全（治本）**：`web/mobile/src/pages/notice-detail/notice-detail.vue:23` 用 `<rich-text :nodes="notice.content">` 渲染后端公告 content（H5 端等价 v-html），而 community-hub-service 的公告/内容帖写入路径（Create/Update）无 HTML 净化（无 bluemonday/sanitize 类依赖）。内容源（发公告者）注入 `<img onerror>` / `<iframe>` / `<script>` 即构成存储型 XSS：一旦通过审核公开，所有浏览该公告的用户端（含 H5）执行恶意脚本。修复点在**写入路径**做白名单净化，DB 存净化后内容，读路径天然安全——一处修复覆盖全部渲染端（移动端/PC 端未来同源）。

**体验（两处前端缺陷）**：
1. 首页 `notice.vue` 双重加载——`onMounted await store.loadMemberships(); loadAll()` 与 `watch(currentCommunityId) → loadAll()` 并存；`loadMemberships` 内 `getAppState` 服务端权威覆写 `currentCommunityId` 时触发 watch 一次 + onMounted 显式一次，同一批接口（通知列表 + 寻失列表）拉两遍。
2. 登录成功 toast 覆盖失败提示——`auth-flow.ts handleAuthSuccess` profile 拉取失败先 `showToast('获取用户资料失败')`，随后立即 `showToast('登录成功')`（icon success）覆盖，失败提示不可见，用户无从得知资料未同步。

## 做什么

修复 3 项：

1. **存储型 XSS 修复（community-hub-service，治本）**：引入 `github.com/microcosm-cc/bluemonday` 白名单 HTML 净化器，对公告/内容帖正文 `content`（DB 列 `text`，REST wire 键 `content`）在**写入路径**做白名单净化后落库；保留穷举白名单标签（p/div/h1-h6/blockquote/ul/ol/li/pre/hr/strong/em/b/i/u/s/span/br/a），剔除 script/iframe/**img 全剔除**/事件属性（on*）等，`a` 仅允许 http/https/mailto href 并剔 javascript:/data:（D8 穷举，含 a/div 去留）；净化器单例化（D4，本服务，不引 community-common）。净化覆盖三个写入口：**CreateContentPost / UpdateContentPost 内容编辑分支 / UpdateContentPost submit 发布分支**（submit 对既有 draft 正文先净化再置公开，D9 关闭存量草稿发布缺口）。净化与非空校验顺序钉死（D7）：非空校验（080005）以原始正文先行，净化在通过后、落库前；正文净化后为空接受空串落库。写测试覆盖注入剥离、合法富文本保留、幂等；上线前抽样比对既有公告/编辑器产出与白名单（D12）。
2. **首页双重加载修复（web/mobile notice.vue）**：采用「显式单次加载 + watch 首载守卫」（D1）——初始加载由 onMounted 在 `loadMemberships` 完成后显式触发一次；`watch(currentCommunityId)` 通过 `membershipsResolved` 布尔守卫在首次 memberships 就绪前忽略变更（避免 getAppState 覆写触发重复），用户后续切换小区仍每次触发一次加载；`loadMemberships` 整体失败时按无小区空态处理、不以陈旧 cid 发请求。
3. **登录 toast 覆盖修复（web/mobile auth-flow.ts）**：profile 拉取失败时失败信息与「登录成功」合并到同一提示展示（D2），合并 toast 以 icon:none 展示、不承诺自动恢复（D14）；token 已存，登录跳转逻辑（switchTab notice / redirectTo join-community）不变。

## 决策日志（澄清结论 → 追溯依据）

| 决策 ID | 决策内容 | 依据 | 修订来源 |
|--------|---------|------|:---:|
| D1 | 首页双重加载采用「显式单次加载 + watch 首载守卫（membershipsResolved）」 | Q1=1 用户拍板 | — |
| D2 | 登录 toast 失败与成功合并到同一提示展示 | Q2=0 用户拍板 | — |
| D3 | 净化字段范围仅公告/内容帖正文 content/text | Q3=0 用户拍板 | — |
| D4 | 净化器放 community-hub-service 本服务（不引 community-common） | Q4=0 用户拍板 | — |
| D5 | 不做存量数据回填净化 | Q5=0 用户拍板 | — |
| D6 | 正文 img 全剔除 | Q6=0 用户拍板 | — |
| D7 | 净化与非空校验顺序钉死：080005 以原始正文先行，净化在通过后落库前；净化后为空接受空串落库 | REVISION clarity M2 / coverage S1 | 已解决 |
| D8 | 白名单穷举允许标签/属性集合（含 a/div 去留；a href 仅 http/https/mailto 剔 javascript:/data:） | REVISION clarity M1 | 已解决 |
| D9 | submit 发布路径对既有正文先净化再置公开（关闭存量草稿发布缺口） | REVISION validity M1 | 已解决 |
| D10 | REQ-XSS-4 限定「新写入/更新正文」；存量行残余风险指向 REQ-XSS-6 | REVISION structure S2 | 已解决 |
| D11 | Update 正文未携带（proto3 presence）不重写不重净化，归入 D5 残余风险 | REVISION coverage S2 | 已解决 |
| D12 | 上线前抽样既有公告/编辑器产出与白名单比对，结论入 CHANGELOG/决策记录 | REVISION validity M1（白名单交叉核对） | 已解决 |
| D13 | 纯文本保存断言以「渲染等价」为准（净化器 HTML 实体转义属允许行为） | REVISION validity S1 | 已解决 |
| D14 | 合并 toast 文案不承诺自动恢复 + 失败时 icon:none 非 success | REVISION clarity S1 / validity S3 | 已解决 |

## 影响范围

| 服务 | 变更类型 | 说明 |
|------|:---:|------|
| community-hub-service | 修改（安全净化） | 新增 bluemonday 依赖 + 净化器单例；CreateContentPost / UpdateContentPost（内容编辑分支 + submit 发布分支）写入前净化正文 |
| web/mobile | 修改（两处修复） | `notice.vue` 数据加载时序去重；`auth-flow.ts` toast 合并 |
| api-proto | 无 | 不新增公开 API，不改 proto（净化在服务内部，wire 不变） |

## 风险评估

- **合法富文本被误伤**：白名单策略穷举保留常用公告标签（含 div/span/h1-h6/pre 等），剔除 img（D6）与全部非白名单属性（含 style）。风险：中 → 缓解：单例统一策略 + 单元测试锁定保留/剔除清单 + 上线前抽样比对既有公告/编辑器产出（REQ-XSS-8，D12）确认无误杀。
- **存量已发布含恶意 HTML 的公告仍会渲染执行**：D5 用户拍板不做存量回填，仅新写入 + submit 发布路径净化。风险：残留（当前环境内容源受信度低时可接受）→ 缓解：明确纳入 out_of_scope + 文档记录残余风险；后续可单独发起存量数据清洗变更。
- **submit 发布存量 draft 引入净化时点**：D9 在置公开前净化既有正文，属新增写路径行为。风险：低 → 缓解：净化器幂等（REQ-XSS-3），既有已净化正文不产生二次改写；submit 单测覆盖。
- **首页加载时序回归**：守卫改动可能造成「首载不加载」或「切换小区不刷新」。风险：中 → 缓解：REQ-DBL 场景覆盖（首载单次/服务端覆写/切换单次/无小区/memberships 失败降级/下拉刷新），TDD RED→GREEN 摘录。
- **toast 合并语义理解偏差**：合并文案需同时表达「成功」与「资料未同步」且不承诺自动恢复。风险：低 → 缓解：spec 以行为契约定义（失败信息必须可见、icon:none），文案由实现按「登录成功（资料加载失败）」口径落地，H5 渲染确认。

## 不做清单（MoSCoW 的 Won't have — 本轮明确不实现）

- **存量已发布公告数据回填净化**（D5）：仅新写入 + submit 发布路径净化，存量已发布恶意 HTML 不迁移清洗（存量 draft 经 submit 发布时净化属写路径例外，D9）
- **前端富文本渲染改为转义/消毒**（如 rich-text 改为过滤渲染）：依赖后端写入路径净化（治本），前端 `rich-text :nodes` 渲染不变
- **title 等纯文本字段净化**（D3）：范围仅公告/内容帖正文 content/text；title 以 `{{ }}` 插值纯文本渲染，无富文本 XSS 面
- **寻失 description / 便民联络等非公告字段**（D3）：净化范围仅公告/内容帖正文
- **新增公开 API / api-proto 变更**：净化在服务内部，wire 契约零变更
- **PC 端（web/pc）公告前台改造**：本变更前端文件均位于 web/mobile；PC 无公告前台富文本渲染（`grep v-html/rich-text web/pc/src` 0 命中）
- **后台管理端公告发布编辑器改造**：不涉及

## 验收标准

- community-hub-service：`go build ./...` + `go test ./...` 无 FAIL；`bash .harness/skills/qa/scripts/harness-checks.sh --service community-hub-service` 无 FAIL；新增净化器单测覆盖注入剥离 + 合法富文本保留 + 幂等（RED→GREEN 摘录入 CHANGELOG）
- web/mobile：`npm run type-check` + `npm run test:unit` + `npm run build:h5` 全绿；notice.vue / auth-flow.ts 新逻辑函数配 TDD RED→GREEN 摘录入 CHANGELOG
- 行为验收：写入含 `<img onerror>/<script>/<iframe>` 的公告 → DB 与读取内容已净化；合法富文本（p/strong/div/a 链接等）保留渲染、合法 http(s) 链接可点击；存量 draft 经 submit 发布后读回为净化内容；首页初始进入通知+寻失接口各只拉一遍（含服务端覆写 cid、memberships 失败降级场景）；登录 profile 失败时失败信息可见（合并 toast，icon:none）
- **白名单交叉核对（D12）**：上线前抽样既有公告正文/编辑器产出与白名单比对，确认无合法标签误杀，结论记入 CHANGELOG/决策记录
- CHANGELOG 同步（services/community-hub-service/CHANGELOG.md、web/mobile/CHANGELOG.md）
