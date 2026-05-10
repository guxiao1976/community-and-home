## Context

基层组织模块（`web/pc/src/views/grassroots/Index.vue`）管理 level 4（街道/乡镇）和 level 5（社区/村）两级行政区划。当前使用单一 `submission_status` 字段（0-4）标识数据状态，其中 4 代表"待删除"。这种设计无法区分操作意图，导致操作权限模糊。

当前后端状态：
- `Division` types 结构体缺少 `SubmissionType` 字段，但 model 层 `MdAdministrativeDivision` 已有 `SubmissionType sql.NullInt64`
- API 路由已有 7 个 division 端点，需新增 3 个
- 审批中心 logic 层已能处理 `submission_type`（2=修改，3=删除），但前端未完整对接
- 数据库 schema 中 `submission_type` 列实际已存在（model 已映射），schema.sql 文件未同步

## Goals / Non-Goals

**Goals:**
- 建立双字段状态体系（`submission_status` + `submission_type`），明确操作权限矩阵
- 新增发起删除、取消删除、撤回提交 3 个接口
- 前端查询编辑 Tab 和提交管理 Tab 按新状态体系展示操作按钮和筛选
- 审批中心正确处理删除申请审批通过后的软删除

**Non-Goals:**
- 不修改行政区划管理（level 1-3）模块，仅改基层组织
- 不修改住宅小区模块的审批流程
- 不做审批中心 UI 改动，仅后端逻辑适配
- 不处理并发编辑冲突（同一数据两人同时编辑的场景）

## Decisions

### D1: 复用已有 submission_type 字段，不新增 DDL

数据库 model 层已有 `SubmissionType sql.NullInt64` 字段，直接使用。新增数据时设 `submission_type=1`（新增），历史数据 `submission_type` 为 NULL 的视为已批准数据（兼容旧数据）。

**替代方案**：新建一张审批流水表。 rejected — 过度设计，当前只需要单字段即可。

### D2: submission_status=4 废弃，迁移为 submission_status=0 + submission_type=3

原 `submission_status=4`（待删除）语义等同于新的 `删除待提交(0,3)`。部署时执行 SQL 迁移：
```sql
UPDATE md_administrative_division SET submission_status = 0, submission_type = 3 WHERE submission_status = 4;
```
前端移除对状态 4 的所有引用。

### D3: 删除待提交审批拒绝时直接恢复已批准

边界情况讨论中有两种选择：A）拒绝后 status=3（已拒绝），用户需手动取消删除；B）拒绝后直接恢复 status=2（已批准）。选择 B，因为拒绝删除申请的本质是"保留数据"，恢复已批准是更直觉的行为。

### D4: 物理删除接口复用现有 DELETE，仅加前置校验

不新建物理删除接口，在现有 `DELETE /divisions/:id` 的 logic 中增加校验：仅允许 `submission_status=0 AND submission_type=1` 时执行物理删除。其他情况返回 403 错误和明确的错误信息。

### D5: 编辑已批准数据时后端自动设置 submission_type=2

前端编辑对话框无需感知 `submission_type`。后端 `Update` logic 在执行更新时检查：若当前 `submission_status=2`，则将 `submission_status` 改为 0、`submission_type` 改为 2。若当前 `submission_status=0`（编辑待提交中的数据），则仅更新内容，不改变状态。

### D6: 审批通过删除申请时在审批 logic 中执行软删除

审批中心 `reviewItemLogic` 已有按 entity_type 处理的逻辑。在审批通过（action=approve）时，若 `submission_type=3`，对 `administrative_division` 类型执行 `SoftDelete`（设 `delete_time`）。无需额外调用删除接口。

### D7: 前端操作按钮通过 unified helper 函数控制

在 `grassroots/Index.vue` 中新增一组 helper 函数（`canEdit`、`canDelete`、`canSubmit`、`canWithdraw`、`canCancelDelete`、`canAddChild`），接收 `{ submission_status, submission_type, level }` 参数返回布尔值，替代现有的简单比较。这使操作权限矩阵集中在一处，便于维护。

## Risks / Trade-offs

- **[历史数据兼容]** 现有 `submission_type=NULL` 的已批准数据 → 查询和展示时 NULL 等同于 type=1（新增），审批中心查询不受影响
- **[前端改动范围大]** 查询编辑和提交管理两个 Tab 都需调整操作列逻辑 → 两个 Tab 的状态筛选栏和操作按钮改动应同步进行，避免不一致
- **[撤回操作的时效性]** 已提交但审批人尚未处理时可撤回 → 如果审批人正好在审批，存在竞态。不处理此竞态，实际业务中审批量小，概率极低
