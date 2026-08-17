# Code Review — 用户服务（规范工程视角）

**审查时间**: 2026-08-17 12:40 (UTC)
**审查范围**: 工作树 diff — ApplyRole 服务角色免带房号 membership（`needMembership` 新增 + applyRole 作用域/校验拆分）
**审查维度**: 规范遵循(#3)、复用性(#6)、测试覆盖(#7)、可观测性(#9)、记忆遵守(#12)

## 摘要
- 🔴 CRITICAL: 1 / 🟡 WARNING: 3 / 🔵 NOTE: 5

## 发现

### 🔴 CRITICAL

| # | 文件:行号 | 维度 | 问题 | 修复建议 |
|---|----------|------|------|---------|
| 1 | `rpc/internal/logic/user/apply_role_logic.go:62-66` | #12 记忆遵守 | **遗漏 must-follow 记忆 `auto-grant-unverified-grant-confers-scope-level0`**。变更把服务角色（grid_worker/community_admin/property_admin）从「必须 membership」改为「无需 membership，申请即 grant community 数据范围（status=0, scope_type=community）」。该 must-follow 记忆（service: all，触发词 AssignRole/status=0/未认证 grant/scope/认证绕过）明确：status=0 未认证 grant 会**立即**产生数据范围 + level-0 能力；「安全默认：数据范围应绑定 membership 而非角色 grant，或在认证通过后再 grant」；且「涉及自动授权的调用方需在 review 时检查此语义」。变更既未核验 permission-service sys_permission 种子中服务角色敏感写/管理权限的 min_verf_level=2，也未绑定 membership，更未引用该记忆（knowledge-load 打分召回 + 手工场景匹配均命中）。这意味着任意注册用户可自我申请服务角色并获得小区数据读范围（level-0），构成记忆所描述的认证绕过场景的扩大。 | 按记忆「怎么做」执行：1) 核验 permission-service sys_permission 种子里 grid_worker/community_admin/property_admin 的敏感写/管理权限 `min_verf_level=2`（交由安全视角复验）；2) 安全默认改回「数据范围绑定 membership」或「认证通过后再 grant 服务角色 scope」；3) 若确需保留免 membership 授 scope，至少在本逻辑补 `// SEE: [[auto-grant-unverified-grant-confers-scope-level0]]` 并在 CHANGELOG 记录核验结论。 |

### 🟡 WARNING

| # | 文件:行号 | 维度 | 问题 | 建议 |
|---|----------|------|------|------|
| 1 | `rpc/internal/logic/user/apply_role_logic.go:37-44` | #6 复用性 | `needMembership` 用 switch 定义「角色是否需要 membership」分类，与 `model/vars.go` 既有 `RolesRequiringResidence`（owner/tenant）、`RolesWithExpiry`（tenant/grid_worker/community_admin/property_admin/committee）并存为第三种角色分类来源，同类语义（角色→属性）分散三处，将来任一新增角色都需同步三个地方，易漂移。committee 的差异（需要 membership 但不需要 residence）未在任何一处说明。 | 收敛到 `model/vars.go` 定义为 `RolesRequiringMembership map[string]bool`（与既有 map 风格一致），并注释与 `RolesRequiringResidence` 的边界（committee 差异）；apply_role_logic.go 直接查 map，便于其他调用方（join_community / auth-service 数据范围解析）复用。 |
| 2 | `rpc/internal/logic/user/apply_role_logic.go:74,82` | #3 规范遵循 | 错误码使用裸数字字面量 `NewBaseRespWithError(10005, ...)`（10001/10005/10008/50000 全为裸数字，service 内无命名常量），违反 should-follow 记忆 `error-code-literal-bypasses-qa-gate` / `error-code-collision-and-namespace-alignment`；且本变更 QA 报告第 7 项声称「错误码格式 ✅ 无 magic number（均用命名常量或 0）」与代码事实不符——正是该记忆描述的 QA 门禁盲区（responsex 裸数字可绕过）。本 diff 未新增裸数字（10005 为原逻辑搬迁），但重构后仍停留在违反状态，QA 结论失真。 | 在 model 或独立 constants 定义命名错误码常量（如 `ErrMembershipNotFound = 10005`）并在本文件/全 service 收敛；至少修复 QA 检查使其能识别 responsex 裸数字（对齐记忆 `error-code-literal-bypasses-qa-gate`）。 |
| 3 | `rpc/internal/logic/user/apply_role_logic.go:77` | #9 可观测性 | `l.Errorf("find membership error: %v", err)` 缺少 userId/communityId/roleCode 上下文，多用户并发下无法定位是哪个用户/小区的 membership 查询失败；对比同函数 AssignRole 失败日志（:111）含 userId/roleId，不一致。该行处于本次重构的 `needMembership` 块内（重构即补上下文的时机）。 | 改为 `l.Errorf("find membership error: userId=%d communityId=%d roleCode=%s err=%v", in.UserId, in.CommunityId, in.RoleCode, err)`。 |

### 🔵 NOTE

| # | 文件:行号 | 建议 |
|---|----------|------|
| 1 | `apply_role_logic.go:74,81-82` | 业务拒绝路径（membership 不存在/非 active → 10005）无任何日志，无法观测服务角色/居民角色申请被拒频次（潜在滥用/探测）。建议 debug/info 级日志带 userId+roleCode。 |
| 2 | `apply_role_logic.go:115` | ApplyRole 成功有 Infof 但无指标（counter）埋点；本次变更后「服务角色免 membership 自我授予」是异常增长点，建议加 per-role grant 计数指标 + 告警阈值，否则只能事后查日志。 |
| 3 | `apply_role_logic.go:36` | M2 验证：`// SEE: [[change-verification-checklist]]` slug 存在（global/change-verification-checklist.md），变更流程也遵守了该记忆（build/vet/test/QA 全过，RED 双证）。但把 process 类记忆标注在业务函数注释上位置异常——验证清单是变更流程约束而非函数实现契约，建议移除或仅保留在 CHANGELOG。 |
| 4 | `apply_role_logic_test.go:109` | `TestApplyRole_GridWorker` fixture 仍创建 membership（5005），与「服务角色无需 membership」新不变量不一致（虽不失败，fixture 语义过时），建议删除 membership 创建以对齐新测试。 |
| 5 | `docs/design.md §2.3` | design.md 原设计 grid_worker/community_admin 的 membership_id 有值，本次变更放宽为无需 membership；设计文档未同步。供设计一致性(#2)视角 reviewer 复核。 |

## 记忆遵守检查（M1-M4）

- **M1 收集引用**：变更文件中 `// SEE: [[...]]` 共 2 处（`apply_role_logic.go:36`、`CHANGELOG.md:37`），均指向 `change-verification-checklist`。
- **M2 验证准确性**：slug `change-verification-checklist` 存在（global/），指导（变更后 build+test 验证）已遵守 → 非虚假匹配；但标注位置异常（见 NOTE-3）。
- **M3 检查遗漏**：命中 must-follow 记忆 `auto-grant-unverified-grant-confers-scope-level0`（trigger: AssignRole/status=0/scope/认证绕过，type=pitfall，service=all）→ 遗漏且未遵守 → **🔴 CRITICAL**（见 CRITICAL-1）。should-follow `error-code-literal-bypasses-qa-gate` → 遗漏（见 WARNING-2）。`error-code-collision-and-namespace-alignment` 场景核对：本次复用 10005 语义一致（"小区成员关系不存在或已退出"），未引入同码异义，仅提示。
- **M4 记忆更新建议**：`change-verification-checklist` last_applied/apply_count 可 +1（已遵守）；`auto-grant-unverified-grant-confers-scope-level0` 在修复后应记录为已应用。Reviewer 保持只读，由 Owner 落库。

## 记忆建议（Memory Suggestions）

- `role-category-convergence` — 角色分类语义（是否需要 membership/房屋/有效期）分散多源易漂移，应收敛到 model 层单一 map 来源。category: guideline / should-follow。

---
VERDICT: FAIL — 存在 1 个 CRITICAL（must-follow 记忆遗漏 + 未按记忆核验/绑定 membership 的安全默认），须修复后重新审查。
---
