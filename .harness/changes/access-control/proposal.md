# Proposal: 访问控制与数据权限改造（access-control）

> 设计依据（权威来源）：`docs/specs/access-control-design.md`（已提交 a2fcc3b，覆盖 7 项需求），并修订 `docs/specs/rbac-design.md` §2.5。
> 阶段：L 变更（OpenSpec → N×Pipeline）。开发阶段，无生产存量数据，**不考虑存量数据迁移**（设计 §12 决策3）。

## 为什么做

系统面向**全国小区**，业主/租户是最大群体，管理操作不能依赖人工审批。本次变更把「你是谁 / 你在哪 / 你能做什么 / 能发多少」全部收口为**后端权威**：角色分层（谁）、端限制（在哪登录）、数据权限（能发到哪）、当前小区（默认上下文）、发布配额（能发多少）、房屋约束与同屋互见（防冒充/防刷）。

其中，**角色分层（需求 1）与数据权限统一模型（需求 3）已由前置变更 `access-data-permission`（Wave 1+2）完整交付**——scope 三态、能力分层 `min_verf_level`、`registered_user` 基角色、加入自动授权/退出撤销、`AssertPublishScope` 发布校验均已落地并通过端到端验收。**本变更聚焦其余 4 项新需求（端限制、当前小区、板块发布配额、房屋约束+同屋互见）**，商户广告（需求 7）仅要求模型兼容、本次不实现。

## 做什么（7 项需求 → 交付状态）

| # | 需求 | 设计章节 | 本次交付 | 状态 |
|---|------|---------|---------|:---:|
| 1 | 角色分层（注册用户→未认证业主/租户→认证业主，退出撤销） | §1.4 / §3.1-3.4 | — | ✅ 已由 access-data-permission 交付 |
| 2 | 端限制（角色 `platforms` 配置驱动，移动端角色禁 PC 登录，UX 引导非安全边界） | §4 | 新 spec `platform-restriction` | 本次 |
| 3 | 数据权限统一模型（祖先链命中 + scope 三态） | §5 | — | ✅ 已由 access-data-permission 交付 |
| 4 | 当前小区服务端持久化（`user_app_state`，跨设备一致） | §6 | 新 spec `current-community` | 本次 |
| 5 | 板块发布配额（占位状态定义） | §7 | 新 spec `section-quota` | 本次 |
| 6 | 房屋注册上限 ≤6 + 同屋互见手机号（防冒充） | §3.5 / §5.7 | 新 spec `member-constraint` + `same-house-visibility` | 本次 |
| 7 | 商户广告投放（范围+期限订单） | §10 | 仅模型兼容 | 未来（本次不实现） |

## 影响范围

| 服务 | 变更类型 | 说明 |
|------|:---:|------|
| auth-service | 新校验逻辑 | 登录/刷新时按角色 `platforms` 判定端准入；`device_type` 归类（web/admin→PC；ios/android/miniapp→移动端） |
| permission-service | 数据模型 + 透出 | `sys_role` 新增 `platforms` 属性；`sys_admin` 维护；角色查询/登录链路透出 platforms |
| user-service | 新数据模型 + 新 API + 新校验 | `user_app_state`（当前小区）；`app-state` / `current-community` 接口；JoinCommunity 房屋约束（≤6/加入次数）；用户详情同屋互见判定 |
| community-hub-service | 新校验 | 写接口在 `AssertPublishScope` 之后执行板块发布配额计数；超限拒绝 |
| master-data-service | 新配置 | `sys_section_quota(section_type, max_count)` 板块配额；成员约束/配额相关 `sys_config` 键 |
| web/pc + web/mobile | 前端 | 移动端登录引导页、当前小区切换 UI、房屋（楼/单元/房号）表单、同屋手机号展示 |
| moderation-service | 无 | 认证审核 AI 自动化（房产证 OCR）属未来（§3.3/§10），本次不涉及 |

## 风险评估

- **端限制被绕过（当安全边界）**：可能性 中 / 影响 数据越权 / 缓解：明确定位为**UX 引导非安全边界**，真正安全由后端 RBAC + 数据权限兜底（§4 标题即标注）。
- **「当前小区」与 `preferences.default_community_id` 双源不一致**：可能性 中 / 影响 首页/发布默认上下文错乱 / 缓解：已定 `user_app_state` 取代 preferences（开发阶段无存量数据，无需迁移）；`preferences` 字段去留阶段3定。
- **配额被「发→删→重发」刷爆审核队列**：可能性 高 / 影响 审核队列无限堆积 / 缓解：待审（`moderation_status=0`）与展示中（`status=active`）均占配额（§7 占位状态定义）。
- **房屋上限/同屋互见依赖「楼/单元/房号」在加入时采集，但现有 `user_residence` 仅认证通过后创建**：可能性 高 / 影响 防冒充失效 / 缓解：已定 JoinCommunity 即采集楼/单元/房号 + membership 增 building/unit/room 三列（CLAR-4 a+c），并引 [[api-required-field-marked-optional]] 确保字段必填。
- **错误码命名空间不一致**：设计文档用 `050006`/`80007`/`10014`，与各服务现状 6 位码（auth 500xx、community-hub 080xxx、user 100xx）不对齐 / 影响 低 / 缓解：spec 以行为契约为准（拒绝并给出语义唯一错误），最终错误码在阶段3架构设计对齐，引 [[error-code-collision-and-namespace-alignment]]。
- **手机号同屋互见泄露**：可能性 中 / 影响 PII 泄露 / 缓解：仅限「同一小区+楼+单元+房号 的 active membership」互见，手机号读取必须解密（[[phone-encryption]]），非同屋一律脱敏。

## 验收标准（高层）

- owner 在 PC 登录 → ❌ 050006 引导移动端；owner+community_admin 在 PC → ✅（任一角色允许）；RT 刷新 PC 会话 → ❌ 同规则拦截。
- 切到未加入小区 → ❌；切到已加入小区 → ✅ 更新且跨设备生效。
- 板块达 5 条再发 → ❌ 80007；驳回/解决/删除/下架后释放；待审仍占配额。
- 房屋第 7 人注册 → ❌ 10014；第 4 个同时加入 → ❌ 10006；反复退出重加入 → ❌ 每年/终身次数限制。
- 同屋用户互见手机号+楼栋房号 → ✅；非同屋查看 → 手机号脱敏。

## 转换追溯（Step 0：设计章节 → Spec 覆盖）

| access-control-design 决策 | proposal 章节 | spec Requirement | 覆盖 |
|----------------------|-------------|-----------------|:---:|
| §4.1 `sys_role.platforms`（`[pc]/[mobile]/[pc,mobile]`） | 做什么 2 | platform-restriction REQ-PLAT-1 | ✅ |
| §4.2 判定（任一角色允许 / 端归类映射 / RefreshToken 同规则） | 做什么 2 | platform-restriction REQ-PLAT-2/3 | ✅（已定 CLAR-1 a：RefreshTokenRequest 增 device_type） |
| §6 `user_app_state` + `app-state`/`current-community` API + 切换校验 ∈ 数据范围 | 做什么 4 | current-community REQ-CUR-1~3 | ✅ |
| §7 `sys_section_quota` + 占位状态定义 + 个人×小区×板块口径 | 做什么 5 | section-quota REQ-QUOTA-1~4 | ✅ |
| §3.5 反滥用约束（同时≤3/每年≤3/终身≤12/每户≤6） | 做什么 6 | member-constraint REQ-MEM-1~4 | ✅ |
| §5.7 同屋互见（手机号+楼栋房号，active membership） | 做什么 6 | same-house-visibility REQ-HOUSE-1~3 | ✅ |
| §10 商户广告（订单范围+期限，模型兼容） | 范围外 | — | ⚠️ 刻意舍弃：未来迭代，模型已兼容 |
| §11.2/11.3/11.4/11.5 验收矩阵 | 验收标准 | 各 spec 正向/异常场景 | ✅ |
| §12 决策3（存量数据不考虑，开发阶段） | 全文 | —（无迁移要求） | ✅ |

## 范围外 / 后续变更

- **商户广告**（§10，订单=范围+期限）→ 未来，仅模型兼容。
- **AI 认证自动化**（§3.3/§10，房产证 OCR + 比对，接 ai-model/moderation）→ 未来。
- **街道/区县/市级层级授权 + 小区广告上限 10**（§1.3 未来列）→ 未来。
- **community-hub 读当前小区作发布默认上下文**（§6 消费侧）→ 本次不实现，留待后续独立跟进。

## 已定决策（本次需求分析阶段定稿）

- **D1**（platform-restriction）：端限制刷新判定采用「RefreshTokenRequest 增 `device_type`」（原 CLAR-1 选项 a）。
- **D2**（current-community）：`user_app_state.current_community_id` 取代 `user_base.preferences.default_community_id`（原 CLAR-2 选项 a）；开发阶段无存量数据，无需迁移；`preferences` 字段去留阶段3定。
- **D3**（section-quota）：配额以 `sys_section_quota` 配置为权威（原 CLAR-3 选项 c）；配置了限制则计、未配置不限；管理员/官方发布默认不配置。
- **D4**（member-constraint / same-house-visibility）：JoinCommunity 即采集楼/单元/房号 + membership 增 building/unit/room 三列（原 CLAR-4 选项 a+c）。

## 阶段3 开放项（架构设计定稿，spec 以 [STAGE3] 标注，只约束行为契约）

- **STAGE3-1**（member-constraint）：「认证用户不受每年次数限制」的认证粒度：全局（任一小区已认证）vs per-community（目标小区认证）。
- **STAGE3-2**（section-quota）：占配额唯一计数谓词与 design §7 公式矛盾（待审内容 status 非 active 会被公式排除）——tasks.md 需记录「阶段3 修正 design §7 计数公式」。
- **STAGE3-3**（全部）：错误码命名空间对齐（050006/80007/10014 与各服务现状 6 位码；10006 语义冲突）。

## 相关经验记忆

- [[auto-grant-unverified-grant-confers-scope-level0]] — 加入自动授权已即时赋予 scope+level-0 能力，房屋上限/加入次数限制必须防「反复退出重加入」绕过。
- [[phone-encryption]] — 同屋互见返回手机号前必须解密；脱敏与加密存储互斥。
- [[error-code-collision-and-namespace-alignment]] — 050006/80007/10014 等新错误码必须语义唯一、避让现状命名空间。
- [[api-required-field-marked-optional]] — JoinCommunity 房屋字段（楼/单元/房号）为必填，禁止误标 optional。
- [[frontend-business-rule-hardcode]] — 房屋上限 ≤6 / 配额上限不得前端硬编码，须从后端配置读取。
- [[is-system-no-permission-shortcut]] — `platforms` 为配置驱动属性，不得用 `is_system` 短路端判定。
