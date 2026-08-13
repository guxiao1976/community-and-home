# Plan Review — access-control（覆盖完整性视角·第 2 轮）

**审查维度**: 需求覆盖 / 场景完整性 / 边界识别
**审查时间**: 2026-08-13
**审查者**: 计划评审 Reviewer（coverage lens）
**审查对象**: request.md、proposal.md、.change.yaml、specs/ 下 5 个 spec（v2）、docs/specs/access-control-design.md、round 1 报告 spec_review_coverage_v1.md

## 摘要

- 🔴 MUST FIX: 1 / 🟡 SHOULD FIX: 1 / 🔵 INFO: 3

**总体结论**：round 1 的 MUST FIX（member-constraint「终身加入总数上限」缺正向 Scenario）已正确修复；5 个 SHOULD FIX 中 3 个已修复（CLAR-2 措辞、current-community 权限码注册、section-quota 下架/移除释放），**2 个未修复**：platform-restriction「未声明平台默认 fail-open」与 design §4.1 仍矛盾（SHOULD #1）、section-quota「个人 × 小区 × 板块」三维口径仍无异常 Scenario 且「个人」维度零覆盖（SHOULD #5）。其中后者与 round 1 已判为 MUST 的「缺正向 Scenario」属同一类覆盖规则违反（缺异常侧），本轮升级为 MUST FIX。

## Round 1 修复核验

| Round 1 发现 | 级别 | 核验结果 |
|------|:---:|------|
| MUST #1 — member-constraint「终身加入总数上限」缺正向 | 🔴 | ✅ **已修复**：新增「终身未达上限」Scenario（L45-48），happy path 可验证 |
| SHOULD #1 — platform-restriction「未声明平台=允许所有端」与 design §4.1 默认映射矛盾 | 🟡 | ❌ **未修复**：spec 仍写「未声明平台的角色 MUST 视为允许所有端（默认不拦截）」（L10、L17-20），design §4.1 仍保留「PC=sys_admin/...；移动=owner/.../registered_user」默认映射，未回写 design、未记为显式决策 |
| SHOULD #2 — CLAR-2「迁移存量数据」与「无存量数据」前提矛盾 | 🟡 | ✅ **已修复**：L66 改为「开发阶段无存量数据，无需迁移」 |
| SHOULD #3 — current-community 缺新接口权限码注册 | 🟡 | ✅ **已修复**：新增 Requirement「新接口权限注册」（L58-64）+ 职责边界条目（L72） |
| SHOULD #4 — section-quota「下架/移除释放」无 Scenario | 🟡 | ✅ **已修复**：新增「已解决/下架/移除释放配额」Scenario（L58-61） |
| SHOULD #5 — section-quota REQ-4 缺异常 Scenario + 未体现「个人」维度 | 🟡 | ❌ **未修复**：REQ-4 仍仅 2 条正向「放行」Scenario，无异常/拒绝，且「个人」（多用户互不计数）维度无任何 Scenario |

## 发现

### 🔴 MUST FIX

| # | 文件:章节 | 问题 | 修复建议 |
|---|-------------|------|---------|
| 1 | `specs/section-quota/spec.md` — Requirement「配额计数口径为个人 × 小区 × 板块」(L72-83) | 该 Requirement 标题明确为**三维口径「个人 × 小区 × 板块」**，但 2 个 Scenario（不同板块独立 L75-78、不同小区独立 L80-83）**均为正向「放行」**，无任何异常/拒绝 Scenario；且**「个人」维度零覆盖**——没有任何场景验证「不同用户之间配额互不计数」（用户 A 满额不影响同小区同板块的用户 B）。这是 round 1 SHOULD #5，本轮仍未修复；与 round 1 已判 MUST 的「缺正向 Scenario」属同一类「每 Requirement ≥1正向+1异常」规则违反（此处缺异常侧）。 | 补「个人维度隔离」Scenario：GIVEN 用户 A 在小区 X 板块 S 已满额、用户 B 同小区同板块未满额，WHEN B 发布，THEN B 放行（按 B 自身计数，不受 A 占用影响）。同时补 1 条异常侧（如用户 A 向已满额的小区 X 同板块再发 → 拒绝，跨小区/跨板块不受影响），与「放行」侧配对。 |

### 🟡 SHOULD FIX

| # | 文件:章节 | 问题 | 建议 |
|---|-------------|------|------|
| 1 | `specs/platform-restriction/spec.md` — Requirement「角色声明允许的端」(L9-20) | round 1 SHOULD #1 未修复：spec 的 fail-open 规则「未声明平台 → 允许所有端」与 design §4.1（L141 具体角色→端默认映射，`registered_user` 明确列移动端）对同一边界（未配置 `platforms` 的角色）得出**相反结果**——按 spec 放行 PC，按 design 拦截 PC。边界 Scenario「角色未声明平台」虽存在，但内容与设计矛盾，属覆盖口径不一致。 | 二选一对齐并落地为单一事实源：① 采用 design §4.1 默认映射为 seed，未声明角色改 fail-closed；② 若坚持 fail-open，则在 spec 加「✅ 已定」决策注记并回写 design §4.1，使阶段 3 实现不产生二义。 |

### 🔵 INFO

| # | 建议 |
|---|------|
| 1 | `specs/current-community/spec.md` — 新增 Requirement「新接口权限注册」(L58-64) 仅 1 个 Scenario，其 THEN 将「按权限码放行」与「未授权拒绝」两个分支合并。建议拆为 1 正向 + 1 异常两条 Scenario，与全库风格一致。 |
| 2 | `specs/member-constraint/spec.md` — Requirement「每年新加入数上限」(L27-38) 的 2 个 Scenario 均为负向（非认证达上限拒绝）或豁免（认证不受限），缺显式正向「非认证用户年内未达上限 → 放行」。当前由「认证用户不受限」间接覆盖，建议补 1 条显式正向。 |
| 3 | `specs/same-house-visibility/spec.md` — Requirement「退出房屋即撤销互见」(L32-43) 的 2 个 Scenario 均为负向（撤销语义，天然如此）。可补 1 条正向（active 期间互见持续有效）以严格满足「≥1正向+1异常」；属低风险。 |

---

## 覆盖矩阵核对（design → spec，第 2 轮）

| design 章节 | spec Requirement | 正向/异常 Scenario | 结论 |
|------|------|:---:|------|
| §4.1 / §4.2 端限制 | platform-restriction REQ-PLAT-1/2/3 | 4 正 + 3 异 + 1 映射 | ⚠️（REQ-1 fail-open 与 §4.1 矛盾未解，见 SHOULD #1） |
| §6 当前小区 | current-community REQ-CUR-1~4 | 6 正 + 4 异 | ✅（REQ-4 新增权限码已覆盖；单 Scenario 合并见 INFO #1） |
| §7 板块发布配额 | section-quota REQ-QUOTA-1~4 | 8 正 + 3 异 | 🔴（REQ-4 缺异常 + 缺「个人」维度，见 MUST #1） |
| §3.5 反滥用与房屋约束 | member-constraint REQ-MEM-1~4 | 6 正 + 4 异 | ✅（round 1 MUST #1 已修复） |
| §5.7 同屋互见 | same-house-visibility REQ-HOUSE-1~3 | 4 正 + 3 异 | ✅（REQ-2 双负向见 INFO #3） |
| §10 商户广告（模型兼容） | —（范围外） | — | ✅ 刻意舍弃 |
| §12 决策3 存量数据 | —（无迁移要求） | — | ✅（CLAR-2 措辞已修） |

---

VERDICT: REVISION
---
