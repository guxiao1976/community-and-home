# Proposal: 角色管理 bug 修复 + platforms 写链路补全

## 为什么做

角色管理页面存在两个线上缺陷与一个写链路缺失：

1. **编辑系统角色必现 500（Bug 1）**：编辑系统角色（如「业主」is_system=1）时，RPC 层 `UpdateRole` 返回 Base 业务错误 60004「系统角色不可修改」且 `Role=nil`；API 层 `updaterolelogic.go` 未检查 `grpcResp.Base`（未调 `responsex.ToError`），直接 `toRoleInfo(grpcResp.Role)` 触发空指针 panic → HTTP 500。同类隐患还潜伏在 create/getrole/getuserroles 三个 API logic 中（[[:rpc-callback-must-check-response-base]]）。
2. **platforms（允许登录端）写链路整体缺失（Bug 1b）**：读链路（proto Role.platforms → RPC toRolePb → 前端列表/表单展示）已就绪，但写链路从 proto 请求 → API 类型 → API logic → RPC logic 全部缺失，`sys_role.platforms` 只能靠 seed SQL 初始化，无法通过管理后台配置。选「移动端」保存时数据被丢弃（配合 Bug 1 表现为 500）。
3. **列表列宽挤压（Bug 2）**：ID 列 `width=200` 过宽（实际 ID 仅 1-2 位数字），挤压角色名称/编码/描述导致换行，操作列 380px 过宽。

用户价值：管理员可正常编辑系统角色配置字段、可配置角色允许登录端、列表呈现清爽无换行。

## 做什么

一次门禁周期内交付（范围已由用户拍板，见「已确认的设计决策」）：

- **Bug 1 + base-check-audit**：`responsex.ToError(grpcResp.Base)` 类级规则推广到 perm 包全部 API logic（grep 核实 11 文件：create/update/getrole/getrolepermissions/getuserpermissions/getuserroles/listpermissions 7 个「空指针 panic」类 + delete/assign-role-permissions/assign-user-role/revoke-user-role 4 个「静默吞业务错误」类；listroles 已有正确模式），`toRoleInfo` 加 nil 防御，整类消除「业务错误静默吞掉 + 空指针 panic」隐患。
- **sys-role-edit-policy**：系统角色放行 name/description/platforms/sort_order 字段编辑，`status` 仍拦截（防误禁用锁死后台，返回 60004），权限继续走独立分配流程（`AssignRolePermissions`），前端编辑按钮保持可用。
- **Bug 1b platforms 写链路**：`CreateRoleRequest`/`UpdateRoleRequest` 增 `repeated string platforms`（值域 pc/mobile）→ `make generate` → API 类型 + create/update API logic 透传 → create/update RPC logic 经 `joinPlatforms` 写库 → 值域校验（非法端 60008 拒绝）→ platforms 变更触发 `invalidateRoleCache`。
- **sortorder-latent-bug**：`sys_role` Update SQL 补 `sort_order = ?`（当前写库 SQL 遗漏该列，属潜伏缺陷，成本近零顺带修复）。
- **http-read-path-gap**：`types.RoleInfo` 增 `Platforms []string` + `toRoleInfo` 透传，HTTP 读链路与写链路一并闭合。
- **Bug 2 列宽重排**：List.vue ID 列收窄至约 70、操作列 380→约 260（四按钮并排）、名称/编码/描述 min-width 自适应、状态/系统角色/允许登录端/创建时间保持固定宽度。

## 影响范围

| 服务 | 变更类型 | 说明 |
|------|:---:|------|
| api-proto | proto 变更 + 生成 | `permission/v1/permission.proto` 的 `CreateRoleRequest`/`UpdateRoleRequest` 增 `platforms` 字段；`make generate` 更新 gen/go；同步更新 CHANGELOG.md |
| permission-service | API+RPC+模型 | API types（RoleInfo/CreateRoleReq/UpdateRoleReq）+ 11 个 API logic（ToError 类级审计）+ 2 个 RPC logic（platforms 写 + 值域校验 + 系统角色字段级策略）+ `model.SysRole.Update` 补 sort_order；补测试 |
| web/pc | 前端 | `List.vue` 列宽重排（Bug 2）；编辑对话框无需改（platforms 表单已存在，编辑按钮保持可用） |

## 已确认的设计决策

用户已拍板（本 proposal 与 specs 以此为唯一权威基线）：

| # | 决策 | 方案 |
|---|------|------|
| D1 | sys-role-edit-policy | **方案A**：系统角色允许编辑 name/description/platforms/sort_order/permission_ids；`status` 仍拦截（60004，防误禁用锁死后台，先于其他校验）；权限仍走独立分配流程（`AssignRolePermissions`→UpdateRole 对系统角色放行 permission_ids，前端权限页交互路径不变）；前端编辑按钮保持可用 |
| D2 | http-read-path-gap | **方案A**：`types.RoleInfo` 增 `Platforms`(json:'platforms') + `toRoleInfo` 透传（`UserRoleInfo.Role` 复用），与写链路一并交付 |
| D3 | update-platforms-empty-semantics | **方案A**：空=显式清空——API 层始终透传，RPC 无条件覆盖 `existing.Platforms`（`joinPlatforms` 后为空串即允许所有端，fail-open）；前端总是发全量，行为一致 |
| D4 | platforms-validation | **方案A**：RPC 校验并拒绝——新增业务错误码 60008「非法登录端」+ proto 注释，非法值直接报错，杜绝误锁 |
| D5 | base-check-audit | **方案A**：整类消除——`listroleslogic` 的 ToError 模式推广为类级规则，grep 核实全量 11 个受影响 API logic + `toRoleInfo` 加 nil 防御 |
| D6 | sortorder-latent-bug | **方案A**：顺带修复——同一 Update SQL 补 `sort_order=?`，成本近零 |
| D7 | column-width-plan | **方案A**：整体重排——ID 列收窄至约 70，操作列 380→约 260（四按钮并排），名称/编码/描述 min-width 自适应，状态/系统角色/允许登录端/创建时间保持固定宽度 |

## 风险评估

- **权限分配路径误清 platforms（中-高）**：`AssignRolePermissions` API 复用 `UpdateRole` RPC（仅传 `PermissionIds`）。在 D3「RPC 无条件覆盖 platforms」语义下，若该路径不显式透传 platforms，权限配置会静默把角色 platforms 清空为 fail-open，**解除既有端限制（如 owner 从 mobile-only 变 all-ends），属安全相关回归**。
  - 缓解：spec 增 REQ-PLAT-8 —— 权限独立更新路径 SHALL 保留角色现有 platforms（实现：AssignRolePermissions 先读现有 platforms 再随 `PermissionIds` 一并透传）。
  - 已写入 spec，作为 D3 语义的一致性不变量，非范围扩展。
- **系统角色权限配置现状（中）**：现有 RPC `UpdateRole` 在权限替换之前就对 `is_system=1` 整体返回 60004，导致 `AssignRolePermissions`（→ UpdateRole + permission_ids）对系统角色**当前即失败**。D1 字段级策略放行 permission_ids 后随本变更修复（系统角色权限配置可用），非回归。
- **系统角色混合请求（低-中）**：D1 字段级策略下，系统角色请求若同时携带 `status` 与可编辑字段，须**原子拒绝**——先校验 `status`，命中即返 60004，任何字段（含 name/platforms）都不落库，避免部分写入。已由 REQ-UPDATE-4 场景钉死。
- **60004 语义收紧（低）**：系统角色「整体不可修改」收紧为「仅 status 拦截」；已有消费方若依赖 60004 整体拦截，仅前端编辑（受本变更修复）受影响，无其他已知消费方。
- **错误码魔数（低）**：permission-service 现有 60001/60004/60006 均为裸字面量（[[:error-code-literal-bypasses-qa-gate]]）。新增 60008 须用命名常量（rpc 侧建常量如 `CodeInvalidPlatform`），本轮仅新增处合规，存量魔数不属本变更范围（记入 backlog）。**60004 语义本轮收紧（UpdateRole 路径）**：由「系统角色不可修改」（整单拦截）收窄为「系统角色状态不可修改」（仅 status 拦截），message 相应更新；`DeleteRole` 路径的 60004（系统角色不可删除）与「角色已被分配，无法删除」语义保持不改。同码异义问题本身记入 backlog，本变更只改 UpdateRole 路径的 60004 语义。
- **D7 偏离编码规范 §5.2（低）**：§5.2 规定 ID 列宽统一 200px，但其前提是 Snowflake 19 位 ID；`sys_role.id` 为自增小整数（seed 1-2 位数字），故收窄至 ~70px 成立。风险：若未来角色迁移到 Snowflake ID，~70px 会过窄。缓解：仅角色列表偏离，其余 ID 列表仍守 200px；迁移时回退 §5.2 标准。
- **proto 破坏性（无）**：新增 `repeated string platforms` 为向后兼容追加字段，`buf breaking` 应通过。

## 不做清单（Won't have — 本轮明确不实现）

- 前端不新增 `sort_order` 编辑控件（现有编辑表单无该字段，API 已支持，UI 保持现状）。
- 不新增 status 编辑控件 / 不改状态开关（编辑对话框无状态开关，保持现状；系统角色 status 由后端拦截）。
- 不做端准入权威判定（auth-service 负责，本变更只补全 permission-service 的 platforms 配置链路）。
- 不改动 seed 数据 `init_permissions.sql`（现有 platforms 初值已正确）。
- 不重构存量错误码魔数（60001/60004/60006 命名常量化为独立 backlog）。
- 不改 `Role.is_system` 既有误导注释（「系统角色不可删除，且自动获得所有权限」——与 [[is-system-no-permission-shortcut]] 矛盾，本轮不动，记入 backlog）。
- 不做角色批量操作/其他列表功能增强。
- 不调整 `Role` 消息既有字段 / 不引入破坏性 proto 变更。

## 验收标准

- [ ] `cd api-proto && make ci` 通过（lint + breaking-check + generate 一致），gen/go 含 `Platforms` 字段。
- [ ] 前端编辑系统角色（改名称/描述/登录端）保存成功，不再 500；改状态（若直连 API）返回 60004。
- [ ] 创建/编辑自定义角色设置「移动端」，列表与详情回显正确；传空列表=允许所有端。
- [ ] 传非法端（如 `web`）返回 60008，角色不落库。
- [ ] 权限配置（AssignRolePermissions）后角色 platforms 不被误清。
- [ ] 更新角色（含 platforms 变更）后持有用户的 perm 缓存失效。
- [ ] 列表页 ID 列窄、名称/编码/描述不换行、操作四按钮并排。
- [ ] `harness-checks.sh --service permission-service` 通过（含新增/更新测试，`TestUpdateRole_SystemRoleCannotModify` 由 skip 改为断言字段级策略）。
