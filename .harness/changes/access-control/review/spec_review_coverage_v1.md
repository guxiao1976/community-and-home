# Plan Review — access-control（覆盖完整性视角）

**审查维度**: 需求覆盖 / 场景完整性 / 边界识别
**审查时间**: 2026-08-13
**审查者**: 计划评审 Reviewer（coverage lens）
**审查对象**: request.md、proposal.md、.change.yaml、specs/ 下 5 个 spec、docs/specs/access-control-design.md

## 摘要

- 🔴 MUST FIX: 1 / 🟡 SHOULD FIX: 5 / 🔵 INFO: 7

**总体结论**：5 个 spec 对 design 对应章节（§3.5 / §4 / §5.7 / §6 / §7）的覆盖基本完整；request.md 7 项原始需求全部有去向（1/3 由前置变更 access-data-permission 交付、4 项本次 spec、1 项未来仅模型兼容），无偏离用户初衷。4 个 CLAR 决策均已以「✅ 已定」正确吸收进对应 spec，spec 内无残留 `[NEEDS CLARIFICATION]`。唯一硬性缺口是 member-constraint「终身加入总数上限」Requirement 缺正向 Scenario，违反「每 Requirement ≥1 正向 + ≥1 异常」的覆盖规则。

## 发现

### 🔴 MUST FIX

| # | 文件:章节 | 问题 | 修复建议 |
|---|-------------|------|---------|
| 1 | `specs/member-constraint/spec.md` — Requirement「终身加入总数上限」(L35) | 该 Requirement 仅 2 个异常 Scenario（「终身达上限」L38、「反复退出重加入刷身份」L43），**无任何正向 Scenario**。无法验证「终身累计未达上限时正常加入」这一 happy path，违反「每个 Requirement ≥1 正向 + ≥1 异常 Scenario」的覆盖规则。 | 补 1 条正向 Scenario，例如：GIVEN 用户终身累计已加入 5 个小区（上限 12）、WHEN 加入新小区、THEN 系统放行。 |

### 🟡 SHOULD FIX

| # | 文件:章节 | 问题 | 建议 |
|---|-------------|------|------|
| 1 | `specs/platform-restriction/spec.md` — Requirement「角色声明允许的端」(L9-L20) | spec 新增「未声明平台的角色 MUST 视为允许所有端（默认不拦截）」规则，**与 design §4.1（L138-L141）不一致**：design 给出的是「默认：PC=sys_admin/community_admin/property_admin/审核员类；移动=owner/tenant/grid_worker/committee/merchant/registered_user」的具体角色→端默认映射，而非「缺失即放行」。spec 的 fail-open 默认会让未配 `platforms` 的角色（如 design 明确列为移动端的 `registered_user`）获得 PC 准入，削弱本特性的限制意图。 | 二选一对齐：① 采用 design §4.1 的默认映射作为 seed + 未声明角色改 fail-closed；② 若坚持「缺失即放行」，将该行为记为显式决策并回写 design §4.1，避免阶段 3 实现时按 design 默认表与按 spec 规则得出相反结果。 |
| 2 | `specs/current-community/spec.md` — CLAR-2 注记 (L58) | spec 吸收的决策文本为「app_state 取代 preferences.default_community_id（**迁移存量数据**）」，与本变更的前提「开发阶段无生产存量数据、不考虑存量数据迁移」(.change.yaml `data_migration_required: false`、design §12 决策3、proposal L4) 相矛盾。 | 将「迁移存量数据」改为「无存量数据、无需迁移」，或明确说明仅做 schema 层面 replace，不涉及数据搬迁，消除实现阶段的歧义。 |
| 3 | `specs/current-community/spec.md` — 全 Requirement | design §9（L284）要求「新接口（app-state、current-community）注册权限码并纳入自动发现」，spec 职责边界仅写 user-service 提供存储/读取/切换，**未覆盖新接口权限码注册**这一功能项。 | 在 current-community spec 补 1 条 Requirement 或职责边界条目：新接口须注册权限码并纳入权限自动发现（避免新接口无权限码被拦截或越权）。 |
| 4 | `specs/section-quota/spec.md` — Requirement「占配额状态口径」(L40-L56) | Requirement 正文列出 6 类状态（待审/展示占，驳回/解决/删除/**下架/移除**释放），但 Scenario 仅覆盖驳回/解决/删除释放（L48），**「下架/移除释放配额」无对应 Scenario**，该口径分支不可验证。 | 补 1 条 Scenario 覆盖「下架/移除释放配额」，或将「下架/移除」并入现有「驳回/解决/删除释放」Scenario 的 GIVEN/THEN。 |
| 5 | `specs/section-quota/spec.md` — Requirement「配额计数口径为个人 × 小区 × 板块」(L58-L69) | 该 Requirement 的 2 个 Scenario（不同板块独立计数 L61、不同小区独立计数 L66）**均为正向「放行」**，无异常 Scenario；且其内容与 REQ-2 的「按目标小区计数」Scenario（L35-L38）高度重复，未独立体现「个人」维度（多用户间配额互不影响）。 | 补 1 条异常/边界 Scenario（如：用户在小区 A 该板块已满额，同板块向未满额的小区 B 放行、但向满额的小区 A 拒绝），或合并进 REQ-2 并明确「个人」维度（不同用户互不计数）。 |

### 🔵 INFO

| # | 建议 |
|---|------|
| 1 | design §7（L237-L246）自身存在内部矛盾：占位状态表定义「待审（moderation_status=0）占配额」，但其下「计数条件 = `deleted_at IS NULL AND status='active' AND moderation_status IN (0,1)`」会排除待审（待审内容 `status` 尚未 `active`）。spec REQ-3 正确采用了表的语义（待审+展示均占），建议回写修正 design §7 的计数公式，避免阶段 3 照抄公式实现出错。 |
| 2 | current-community spec 未提及 design §6 的缓存 `Redis user:appstate:{userId}`（L227）；属实现细节，可忽略，但阶段 3 实现需保留跨设备一致的缓存失效要求。 |
| 3 | section-quota spec 未提及 design §7 的索引 `user_id+community_id+section_type+status`（L248）；实现细节，供阶段 3 参考。 |
| 4 | member-constraint spec 职责边界仅示例 1 个 `sys_config` 键（`user.max_community_join_count` 等），design §3.5（L123-L127）枚举了 4 个键（同时/每年/终身/每户）；建议 spec 补全 4 个键名，保证配置驱动契约唯一。 |
| 5 | section-quota CLAR-3 注记「管理员/官方发布默认不配置」与 design §7「适用所有发布者」（L247）存在潜在歧义：机制是「按板块不配置→不限」，即同一板块一旦配置，管理员与普通用户**同样受限**，「管理员默认不配置」是配置实践而非角色豁免。建议补 1 条 Scenario 明确「已配置板块的管理员发布是否计入」，避免实现偏差。 |
| 6 | platform-restriction spec 未覆盖「用户零角色」「miniapp→移动端」等显式边界（miniapp 已在端归类映射 Scenario 覆盖，零角色未覆盖）；鉴于定位为 UX 引导，属低风险，可阶段 3 补充。 |
| 7 | 配额校验「先计数后落库」存在并发竞态（同用户并发发帖可能同时通过计数）；design 与 spec 均未提事务/唯一约束/行级锁。若配额强度要求严格，阶段 3 需明确并发语义，否则记为实现假设即可。 |

---

## 覆盖矩阵核对（design → spec）

| design 章节 | spec Requirement | 正向/异常 Scenario | 结论 |
|------|------|:---:|------|
| §4.1 / §4.2 端限制 | platform-restriction REQ-PLAT-1/2/3 | 4 正 + 3 异 + 1 映射 | ✅（REQ-1 默认规则偏离 §4.1，见 SHOULD FIX #1） |
| §6 当前小区 | current-community REQ-CUR-1~3 | 5 正 + 3 异 | ✅（缺 §9 权限码，见 SHOULD FIX #3） |
| §7 板块发布配额 | section-quota REQ-QUOTA-1~4 | 6 正 + 3 异 | ⚠️（REQ-4 缺异常；下架/移除无 Scenario） |
| §3.5 反滥用与房屋约束 | member-constraint REQ-MEM-1~4 | 5 正 + 4 异 | 🔴（终身 Requirement 缺正向） |
| §5.7 同屋互见 | same-house-visibility REQ-HOUSE-1~3 | 4 正 + 3 异 | ✅ |
| §10 商户广告（模型兼容） | —（范围外） | — | ✅ 刻意舍弃，predecessor 已保证模型可插拔 |
| §12 决策3 存量数据 | —（无迁移要求） | — | ✅（CLAR-2 措辞需修订，见 SHOULD FIX #2） |

---

VERDICT: REVISION
---
