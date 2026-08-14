# Request — 角色管理 bug 修复 + platforms 写链路补全

## 用户原始需求（原话）

> 编辑角色时，选中 移动端，保存时：Request failed with status code 500
> 另外，角色管理 主界面，ID 只有1-2位数字，占的太宽，角色名称、角色编码、描述都换行了，请调整。

## 分级信息

- 分级: **L**（涉及 api-proto + 多服务目录）
- 路由: L → spec-pipeline
- 涉及: api-proto / permission-service / web/pc
- 入口判定: dispatch Step 0（Owner 判定，2026-08-14）

## 需求要点（人工澄清后）

| # | 要点 | 用户拍板结论 |
|---|------|------|
| 1 | **Bug1** 编辑系统角色（如「业主」is_system=1）勾选移动端保存 500 | 根因：RPC 返 60004+Role=nil，API 层未 `ToError` 直接 `toRoleInfo(nil)` 空指针 panic → 500 |
| 2 | **Bug1b** platforms（允许登录端 pc/mobile）写链路整体缺失 | proto/API/RPC 全缺，DB 列+读链路+前端已就绪 |
| 3 | **Bug1c** 系统角色编辑策略 | D1 方案A：放行 name/description/platforms/sort_order；status 仍拦截(60004)；权限走独立流程；前端编辑按钮可用 |
| 4 | **Bug2** 列表列宽：ID 200 过宽致名称/编码/描述换行 | D7 方案A：ID→70、操作 380→260、文本列 min-width 自适应 |
| 5 | **范围补充** | 读链路补漏(RoleInfo.platforms)、同类 Base 审计整类消除、sort_order 潜伏 bug 顺带修 |

## 已确认设计决策（D1-D7）

见 `proposal.md` §已确认的设计决策 表（7 项全部用户拍板，唯一权威基线）。

## 本次人工修正记录（2026-08-14，评审 3 轮 REVISION 后 Owner 修正）

1. 补齐 request.md（评审覆盖视角 MUST FIX：缺对照原始需求的依据）。
2. base-check 审计范围从 4 文件扩为**全量 11 文件**（grep 核实）并改为**类级规则**：
   - 空指针 panic 类（deref RPC 响应）：createrolelogic / updaterolelogic / getrolelogic / getrolepermissionslogic / getuserpermissionslogic / getuserroleslogic / listpermissionslogic
   - 静默吞业务错误类（丢弃响应仅查 Go err）：deleterolelogic / assignrolepermissionslogic / assignuserrolelogic / revokeuserrolelogic
   - 已有正确模式：listroleslogic
3. REQ-UPDATE-4 钉死：系统角色含 status → 60004 **先于任何字段应用**（原子拒绝，无部分写入）；permission_ids 对系统角色**放行**（AssignRolePermissions 走 UpdateRole，当前被 60004 误拦，随本策略修复）。
4. REQ-PLAT-8 钉死机制：AssignRolePermissions 先读角色当前 platforms 再随 PermissionIds 一并透传，防 D3 无条件覆盖误清端限制。
5. REQ-PLAT-4 补：60008 用命名常量 `CodeInvalidPlatform`（防 QA 门禁漏检错误码魔数）。
6. .change.yaml capability 顺序修正：role-platforms-write-path 在前（role-update-fix 的 REQ-UPDATE-4 依赖其 platforms 写），并注明 helpers.go/types.go 共享文件由 write-path 先改、update-fix 复用。
