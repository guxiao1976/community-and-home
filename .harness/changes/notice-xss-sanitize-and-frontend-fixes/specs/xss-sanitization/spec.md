# XSS 写入路径净化（xss-sanitization）Specification

> **修订记录**：P1.4（REVISION 修复轮）。对照上轮评审 4 项 MUST/SHOULD 逐条修订：
> - clarity M1（白名单穷举 + a/div 去留）→ REQ-XSS-2 穷举 + REQ-XSS-8 交叉核对 — 已解决
> - clarity M2（净化与非空校验顺序）→ REQ-XSS-1 钉死顺序 — 已解决
> - validity M1（submit 过渡路径绕过净化）→ REQ-XSS-1 入口枚举 + REQ-XSS-6 submit 场景 — 已解决
> - validity S1（纯文本断言）/ structure S2（REQ-XSS-4 范围）/ coverage S1/S2/S3 / validity S1 等 SHOULD → 均已在本 spec 对齐 — 已解决

## Purpose

消除 community-hub-service 公告/内容帖正文的存储型 XSS：内容源（发公告者）可在正文注入 `<img onerror>` / `<iframe>` / `<script>`，而移动端用 `<rich-text :nodes>`（H5 等价 v-html）原样渲染，导致任意浏览用户端执行恶意脚本。本变更在**写入路径**用 HTML 白名单净化器净化正文后落库，使读路径天然安全，同时不破坏既有合法富文本公告的展示。

## Requirements

### Requirement: REQ-XSS-1 — 写入路径 HTML 白名单净化

The system SHALL sanitize the notice/content-post body content (database column `text`, REST wire key `content`) with the HTML whitelist sanitizer (REQ-XSS-2) on **every** write path that persists the body, before the value is written to the database. The write entry points covered by this change are:

1. **CreateContentPost**（RPC `Create`，及经 `api/` REST 代理的同一入口）
2. **UpdateContentPost 内容编辑分支**（RPC `Update` 且 `text` 被携带；REST 代理同一入口）
3. **UpdateContentPost submit 发布分支**（draft→public，对既有正文先净化再置公开，见 REQ-XSS-6）

Any future write path that persists body text SHALL also pass through the whitelist sanitizer before the DB write (structure 评审 S1 收敛：显式枚举入口，防未来直写 model 静默绕过)。

- 说明：
  - REST 层为 RPC 的代理（`api/` → `rpc/`），在 RPC 层上述入口落库前净化即覆盖 RPC 与 REST 两条入口（结构/覆盖评审已核实：`content_posts.text` 当前仅 Create/Update 两处写，submit 分支不写正文，本变更将其纳入净化）。
  - **净化与既有非空校验的执行顺序固定（D7）**：非空校验（`Title=="" || Text==""` → 080005「标题和内容不能为空」）以**原始正文**先行判定，语义与净化前完全一致；净化器在非空校验通过后、DB 落库前执行。
  - **正文经净化后为空的处理唯一化（D7）**：接受空串落库（与 REQ-XSS-4 读回一致），不再回判 080005。
  - 净化器位于 community-hub-service 本服务（D4），单例复用（REQ-XSS-3）。

#### Scenario: 合法富文本保留（正向主流程）
- **GIVEN** 发公告者提交正文 `<p>停水通知</p><p><strong>明早 8 点</strong>恢复供水，<br/>带来不便敬请谅解</p>`
- **WHEN** 该正文经 CreateContentPost 写入（或 UpdateContentPost 更新）
- **THEN** 落库内容保留合法标签与文本：`<p>停水通知</p><p><strong>明早 8 点</strong>恢复供水，<br/>带来不便敬请谅解</p>`，读路径返回的 content 与之一致，移动端 `<rich-text>` 正常渲染富文本样式

#### Scenario: 注入 payload 被剥离（XSS 攻击载荷）
- **GIVEN** 发公告者提交正文含 `<img src=x onerror=alert(1)><script>alert(document.cookie)</script><iframe src="//evil.example"></iframe>`
- **WHEN** 该正文经任一写入路径（Create/Update/submit，RPC/REST）提交
- **THEN** 落库内容中 `<script>`、`<iframe>`、`<img>`（含其事件属性 onerror）全部被剔除，不残留可执行 HTML；剩余纯文本/安全标签正常保存，数据库与读取响应中不出现 `onerror=`/`<script`/`<iframe` 字样

#### Scenario: 边界——原始正文为空（边界输入）
- **GIVEN** 提交正文为空字符串 `""`（title 非空）
- **WHEN** 经写入路径提交
- **THEN** 既有非空校验以**原始正文**判定，返回 080005「标题和内容不能为空」，不进入净化、不落库；净化不改变既有非空校验语义

#### Scenario: 边界——原始正文非空但净化后为空（写入语义唯一化）
- **GIVEN** 提交正文仅含 `<script>alert(1)</script><iframe src=x></iframe>`（原始正文非空 → 通过 080005；白名单外标签全剥离后为空串）
- **WHEN** 经写入路径净化并落库
- **THEN** 落库 `text` 为空串（D7 唯一化：接受空串落库，不回判 080005）；读路径返回空串 content，渲染端不渲染任何可执行 HTML（与 REQ-XSS-4 一致）

#### Scenario: 边界——纯文本正文（渲染等价）
- **GIVEN** 提交正文为纯文本 `小区明天开展义诊活动`（无任何 HTML 标签）
- **WHEN** 该正文经写入路径净化
- **THEN** 保存结果与原文**渲染等价**：不丢失任何可见字符；净化器对字面 `&`/`<`/`>` 的 HTML 实体转义（如 `A&B` → `A&amp;B`）属允许行为，经 `<rich-text>` 渲染显示一致（D13：断言以渲染等价为准，非字节等价）

### Requirement: REQ-XSS-2 — 白名单标签/属性策略（穷举）

The system SHALL apply an HTML whitelist sanitization policy with the following **exhaustive** allowlist (D8). Any tag not in this allowlist SHALL be removed (its child nodes preserved); any attribute not explicitly allowed SHALL be removed.

**允许标签（完整集合，穷举）**：
- 块级：`p`, `div`, `h1`, `h2`, `h3`, `h4`, `h5`, `h6`, `blockquote`, `ul`, `ol`, `li`, `pre`, `hr`
- 行内：`strong`, `em`, `b`, `i`, `u`, `s`, `span`, `br`, `a`

**允许属性（完整集合，穷举）**：
- `a`：`href`（scheme 白名单**仅** `http`/`https`/`mailto`；`javascript:`/`data:`/`vbscript:` → 该 href 属性移除、`a` 其余属性与文本保留）、`target`（仅当同步强制 `rel="noopener noreferrer"` 时保留，否则 `target` 移除）、`rel`、`title`
- 其余所有允许标签：**不允许任何属性**（`style`/`class`/`id`/`on*` 等一律移除）

**全局剔除（与标签无关）**：
- 所有 `on*` 事件属性（onclick/onerror/onload 等）
- 标签级 `script`、`iframe`、`object`、`embed`、`style`、`img`（D6：正文图片一律剔除，公告图片走附件机制 content_post_attachments）、`form`、`input`、`button`、`textarea`、`select`、`link`、`meta`、`video`、`audio`、`svg`、`canvas`、`frame`、`frameset`、`template` 及任何不在上述 allowlist 中的标签
- `style` 属性（CSS 注入面：`expression`/外部引用/`url()` 等）— 由「除 `a` 外不允许任何属性」覆盖

- 说明：`img` 全剔除是已确认策略（D6）；`a` 保留但仅允许安全 scheme，合法链接保留、危险链接降级为纯文本（D8）；`div`/`span` 保留为结构/行内容器但属性全剔除（含 style，D8）；`h1`-`h6`/`pre`/`hr`/`s` 纳入以覆盖常见富文本编辑器产出，避免合法公告降级。白名单冻结后需与既有公告内容实际标签集合交叉核对（REQ-XSS-8，validity 评审 M1）。

#### Scenario: 合法链接保留、仅危险 scheme/事件属性被剔除（正向）
- **GIVEN** 正文含 `<a href="https://example.com/notice" target="_blank" rel="noopener">官方通知</a>`
- **WHEN** 经写入路径净化
- **THEN** `<a>` 与合法 http(s) href 完整保留（含 rel/target），链接可点击、指向不变

#### Scenario: 事件属性与危险 scheme 剔除（攻击载荷）
- **GIVEN** 正文含 `<a href="javascript:alert(1)" onclick="steal()">点我</a>` 与 `<div style="background:url(javascript:evil())">`
- **WHEN** 经写入路径净化
- **THEN** `<a>` 保留但其 `onclick` 移除、`javascript:` href 移除（链接降级为无 href 的纯文本节点，文本「点我」保留）；`<div>` 保留（块级容器）但其 `style` 属性移除；安全文本保留

#### Scenario: 白名单外标签剥离后文本保留（防御纵深）
- **GIVEN** 正文含 `<marquee><b>滚动标题</b></marquee>`
- **WHEN** 净化执行
- **THEN** 白名单外标签 `<marquee>` 被剥离但其子安全标签/文本 `<b>滚动标题</b>` 保留，不把整个内容清空

#### Scenario: img 全剔除（D6 已确认）
- **GIVEN** 正文含 `<img src="/attachments/1.jpg" onerror="alert(1)">` 与 `<img src="data:image/png;base64,...">`
- **WHEN** 净化执行
- **THEN** `<img>`（含其 onerror、data: src）全部剔除，不残留任何 img 元素；正文图片通过附件机制（content_post_attachments）另行承载

### Requirement: REQ-XSS-3 — 净化器单例化与幂等

The system SHALL reuse a single sanitizer instance (singleton) for all write-path sanitization within the service, and sanitization SHALL be idempotent: applying the sanitizer to already-sanitized content MUST NOT further alter it.

- 说明：单例化避免每次请求重建策略（性能 + 策略一致，D4）；幂等保证 Update（重新净化既有值）与 submit（净化存量 draft 正文）不产生渐进式漂移。

#### Scenario: 幂等（重复净化稳定）
- **GIVEN** 内容 `a` 已经过一次净化，结果记为 `s(a)`
- **WHEN** 再次对 `s(a)` 执行净化
- **THEN** 输出与 `s(a)` 完全一致（`s(s(a)) == s(a)`）

#### Scenario: 并发写入（并发冲突）
- **GIVEN** 两个并发 CreateContentPost 请求携带不同正文
- **WHEN** 各自执行净化后落库
- **THEN** 净化互不干扰，两条记录分别存各自净化后内容；单例实例不产生共享可变状态导致的串扰（净化为纯函数）

### Requirement: REQ-XSS-4 — 持久化净化后内容（限定新写入/更新正文）

The system SHALL persist the sanitized body to the database for **newly written and updated** bodies, so that read paths (detail / list / marquee) and all renderers (mobile `rich-text`, future PC) serve sanitized content for those rows without further per-render processing. This requirement applies to rows written or updated **after this change goes live**; rows stored before the change are not covered here and their residual risk is governed by REQ-XSS-6 (D5/D10, structure 评审 S2 收敛范围对齐).

#### Scenario: 读路径返回净化后内容（正向验证）
- **GIVEN** 已有一条**本变更上线后**经净化写入的公告（正文曾含恶意标签，落库时已剥离）
- **WHEN** 用户浏览公告详情（GetContentPost / 详情 REST）
- **THEN** 返回的 `content` 为净化后内容，不含 script/iframe/img/on* 等可执行 HTML

#### Scenario: 正文全部被剥离后仍可读（边界输入）
- **GIVEN** 提交正文 `<script>alert(1)</script><iframe src=x></iframe>`（原始非空 → 净化后空串，落库为空串）
- **WHEN** 该内容落库后被读路径（详情/list/marquee 任一）返回
- **THEN** 返回的 `content` 为空串或纯文本，读路径与渲染端不抛错、不渲染任何可执行 HTML

### Requirement: REQ-XSS-5 — 净化范围与零接口变更

The system SHALL apply sanitization only to the notice/content-post body `content`（`text`）field and SHALL NOT alter any other field, SHALL NOT add any public API, and SHALL NOT change the api-proto wire contract.

- 说明：title 等纯文本字段（`{{ }}` 插值渲染）不在净化范围（D3）；净化为服务内部处理，wire 契约零变更。

#### Scenario: 非正文字段不受影响（范围边界）
- **GIVEN** 提交正文含恶意标签、同时携带 title/community_ids/attachment_ids
- **WHEN** 写入净化执行
- **THEN** 仅正文 content/text 被净化，title 等其余字段按既有语义原样存储，请求/响应结构（wire 契约）与净化前完全一致

#### Scenario: 合法正文写入 wire 契约不变（正向）
- **GIVEN** 提交合法富文本正文与常规 title/scope/attachments
- **WHEN** 写入净化执行
- **THEN** 响应与请求结构（字段名/类型/错误码）与净化前完全一致，无新增/删除字段，api-proto 未变更

#### Scenario: title 等纯文本字段不受净化（范围边界）
- **GIVEN** 提交 title 含 `<b>公告</b>` 等 HTML 字符串（渲染端以 `{{ }}` 插值纯文本展示）
- **WHEN** 写入净化执行
- **THEN** title 原样存储、不执行 HTML 净化（D3 范围仅正文），渲染端以纯文本插值安全展示

### Requirement: REQ-XSS-6 — 存量数据不回填 + submit 发布路径净化（残余风险声明）

The system SHALL NOT retroactively sanitize already-stored published content rows during this change（已确认 D5：不做存量数据回填清洗）。**例外**：存量 draft 草稿在经 `submit`（status==1）发布为公开时，系统 SHALL 在置为公开前对其既有正文执行一次白名单净化（D9，validity 评审 M1 关闭「净化前存量草稿经 submit 发布」缺口）。Update 时正文未携带（proto3 optional presence）不重写正文、不重净化，归入 D5 已接受残余风险（D11，coverage 评审 S2 收敛）。

#### Scenario: 存量已发布记录保持原样不迁移（边界/存量）
- **GIVEN** 库中已存在一条净化前写入、正文含 `<img onerror=...>` 的**已发布**公告（未回填）
- **WHEN** 该公告被读取展示（不经任何写入）
- **THEN** 存量记录内容保持不变（本变更不做迁移清洗，残余风险按 out_of_scope 声明接受）

#### Scenario: 存量 draft 经 submit 发布时净化（正向/关闭缺口）
- **GIVEN** 一条净化前写入的 draft（正文含 `<img onerror=...><script>...`），作者对其实施 submit（status==1）
- **WHEN** 发布流程执行
- **THEN** 在置为公开前，该正文经白名单净化（恶意标签剥离），同一事务将净化后的正文写入并将状态置为公开（REQ-XSS-3 幂等保证既有已净化正文不被二次改写）；发布后读路径返回净化后内容

#### Scenario: 存量已发布记录被编辑时按新语义净化（正向）
- **GIVEN** 同一存量记录（正文含 `<img onerror=...>`）经 UpdateContentPost 重新提交正文（`text` 携带新富文本）
- **WHEN** 更新写入路径净化执行
- **THEN** 该记录正文更新为净化后的新内容，恶意标签随之剥离

#### Scenario: Update 正文未携带时保持现值（边界/D5 残余）
- **GIVEN** 存量记录经 UpdateContentPost 仅修改 title/scope/attachments，请求未携带 `text`（proto3 presence 语义）
- **WHEN** 更新写入路径执行
- **THEN** 正文列不被重写、不执行重净化，现值保持不变（归入 D5 已接受残余风险，与「不回填」一致）

### Requirement: REQ-XSS-7 — 净化器测试覆盖

The system SHALL provide unit tests for the sanitizer covering at minimum: (a) injection payloads（`<img onerror>` / `<script>` / `<iframe>` / `on*` 事件属性 / `javascript:`/`data:` href）被剥离；(b) legitimate rich text（p/strong/em/br/a/div/h2 等，含合法 http(s) 链接保留）保留；(c) 幂等（`s(s(a)) == s(a)`）。测试以 TDD RED→GREEN 摘录形式记录（`services/community-hub-service/CHANGELOG.md`）。新增依赖 `github.com/microcosm-cc/bluemonday` SHALL 锁定明确版本（BSD-3 许可证，与 go 1.25 编译兼容），go.mod/go.sum 一并提交（validity 评审 S2）。

#### Scenario: 注入载荷用例通过（验收）
- **GIVEN** 单测用例以含 `<img onerror>`/`<script>`/`<iframe>`/`javascript:` href 的 payload 调用净化器（RED 阶段用例失败暴露缺陷）
- **WHEN** `go test ./...` 运行该用例（GREEN 阶段）
- **THEN** 用例通过，断言输出中不含 `onerror`/`<script`/`<iframe`/`javascript:`，CHANGELOG 记录 RED→GREEN 证据

#### Scenario: 合法富文本用例通过（验收）
- **GIVEN** 单测用例以 `<p><strong>...</strong></p><br/>`、`<a href="https://...">`、`<h2>`、`<div>` 等合法富文本调用净化器
- **WHEN** `go test ./...` 运行该用例
- **THEN** 用例通过，断言安全标签与文本完整保留、合法链接 href 不变，不产生误删

#### Scenario: 幂等用例通过（验收）
- **GIVEN** 单测用例对净化结果二次调用净化器
- **WHEN** `go test ./...` 运行该用例
- **THEN** 二次净化输出与一次净化完全一致（`s(s(a)) == s(a)`）

### Requirement: REQ-XSS-8 — 白名单与既有内容交叉核对（上线前验收）

Before this change is released, the system SHALL verify the frozen whitelist (REQ-XSS-2) against the actual set of tags/attributes present in existing announcement bodies (sample stored rows, or the admin editor's real HTML output where available), confirm no legitimate tag/attribute used by existing content is wrongly stripped, and record the comparison conclusion in the CHANGELOG / design decision record（D12，validity 评审 M1）。If the comparison reveals legitimate tags/attributes outside the allowlist, the allowlist SHALL be revised before release (with the change recorded).

#### Scenario: 抽样比对无合法标签误杀（验收）
- **GIVEN** 上线前对存量公告正文抽样（或后台编辑器实际产出 HTML）
- **WHEN** 比对抽样标签/属性集合与 REQ-XSS-2 白名单
- **THEN** 比对结论确认无合法标签/属性被误杀（或已修订白名单后确认），结论记录于 CHANGELOG/design 决策记录

#### Scenario: 发现白名单外合法标签（边界）
- **GIVEN** 抽样发现既有公告使用白名单外合法标签（如 `<h2>`/`<table>`/`<span style>`）
- **WHEN** 上线前比对
- **THEN** 在发布前修订白名单纳入该标签（经安全评估）或按既有公告实际使用情况记录明确决策，杜绝「存量合法公告一经编辑重存即降级」
