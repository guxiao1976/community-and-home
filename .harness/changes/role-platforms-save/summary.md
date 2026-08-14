# Summary — role-platforms-save

> 变更 Single Source of Truth（2026-08-14 归档）

## 变更目标

修复角色管理两个线上 bug + 补全 platforms（允许登录端）写链路（用户确认完整修复范围）：

1. **Bug1**：编辑系统角色（如「业主」is_system=1）勾选移动端保存返回 500 —— API 层未检查 RPC Base 空指针 panic
2. **Bug1b**：platforms 写链路整体缺失（proto/API/RPC 全缺），自定义角色保存移动端静默失效
3. **Bug2**：角色列表 ID 列 width=200 过宽，挤压名称/编码/描述导致换行

## 交付内容

| 服务 | 变更 |
|------|------|
| **api-proto** | `CreateRoleRequest.platforms`(6) / `UpdateRoleRequest.platforms`(7)、错误码 `060008 非法登录端` 头注释登记、UpdateRole @description、make ci（lint+breaking+generate 全过）、CHANGELOG |
| **permission-service** | ① 500 修复：API `UpdateRole` 补 `responsex.ToError(grpcResp.Base)` 前置 + `toRoleInfo` nil 防御 + **base-check 类级审计 11 文件**整类消除；② platforms 写链路：`CodeInvalidPlatform`(60008 命名常量) + `validatePlatforms` 校验/去重 + CreateRole/UpdateRole 写 platforms（joinPlatforms）+ 系统角色字段级策略（60004 status-only 原子拒绝，name/description/platforms/sort_order/permission_ids 放行）+ 缓存失效 + HTTP 读链路闭合（RoleInfo.platforms）；③ AssignRolePermissions 先 GetRole 保留 platforms（防 D3 无条件覆盖误清端限制）；④ sort_order 落库修复 |
| **web/pc** | `List.vue` 列宽整体重排：ID 200→70、名称/编码/描述 min-width 自适应 + show-overflow-tooltip、操作 380→260（Vite HMR 已热更新） |

## 设计决策（用户拍板 D1-D7）

D1 系统角色字段级策略（status 仍拦截）/ D2 HTTP 读链路闭合 / D3 platforms 空=显式清空（fail-open，API 恒透传）/ D4 值域校验 60008 / D5 base-check 整类消除 / D6 sort_order 修复 / D7 列宽整体重排。详见 `proposal.md`。

## 质量门禁

- **QA**: permission-service `harness-checks` 16 PASS/0 FAIL（go build/vet/test 全绿，181 测试函数）；web/pc 前端 QA PASS（52 测试全绿）
- **Review**: permission-service 3/3 PASS（security-arch/standards-eng/design-biz），10 WARNING 均非阻塞（已记 memory 建议）
- **spec-pipeline**: 需求评审经 3 轮人工修正 cycle 后 2/3 APPROVED；集成归档完成
- **冒烟**: 重启后 API 401 正常鉴权（无 500 panic）

## 已知/后续（WARNING 级，非阻塞）

- D3「空=清空」语义依赖前端恒传全量 platforms 强契约（其他 API 调用方部分更新会清空，已在 spec/注释文档化）
- design.md/CHANGELOG/proto 三方文档已同步（060008 + platforms 列 + 字段级策略）
- 系统角色字段级可编辑的安全性（提权风险）已记 memory：`privileged-system-entity-field-edit-needs-tier-authz`
- 管线缺陷（spec-pipeline 需求评审盲循环）已记 BACKLOG task-2026-08-14-002 + memory：`spec-pipeline-review-blind-loop`

## 执行记录

| 阶段 | 结果 |
|------|------|
| 0 路径选择 | L（dispatch 入口判定） |
| 1 需求分析 | 3 specs / 14 REQ（经 3 轮人工修正收敛） |
| 2 需求评审 | 2/3 APPROVED |
| 3 架构设计 | 14 tasks / 3 服务组 |
| 4 Proto 变更 | Owner 执行 make ci 全过 |
| 5 编码测试 | permission-service PASS(1轮, conf 1.0) + web/pc PASS(1轮, conf 0.8) |
| 6 集成归档 | ✅ 全流程完成，重启生效 |
