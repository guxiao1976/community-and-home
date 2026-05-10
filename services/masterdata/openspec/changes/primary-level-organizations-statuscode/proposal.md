## Why

当前基层组织模块的提交状态仅用单个 `submission_status` 字段（0-4）标识，无法区分"新建待提交"、"修改待提交"、"删除待提交"三种不同操作意图。导致操作权限模糊（如新建数据和已批准数据共用同一删除逻辑），也无法在提交管理页面按操作类型分类展示。需要引入双字段状态体系，明确每种状态下的操作权限，使审批流程更严谨。

## What Changes

- 采用双字段方案：`submission_status`（流程状态）+ `submission_type`（操作类型），替换原有的单字段状态码
- `submission_status`：0=待提交，1=已提交，2=已批准，3=已拒绝（移除原有的 4=待删除）
- `submission_type`：1=新增，2=修改，3=删除
- 新增 3 个后端接口：发起删除、取消删除、撤回提交
- 修改现有接口的校验逻辑：物理删除仅限新建待提交；编辑已批准数据自动转为修改待提交；审批通过删除申请时执行软删除
- 查询编辑 Tab：按状态+类型动态显示操作按钮，移除「待删除」筛选，待提交条件下显示操作类型列
- 提交管理 Tab：新增操作类型列和撤回按钮，状态筛选与查询编辑 Tab 保持一致

## Capabilities

### New Capabilities
- `submission-status-system`: 双字段状态码体系定义、操作权限矩阵、状态流转规则
- `delete-workflow`: 发起删除、取消删除、撤回提交三个新接口及对应的前端交互
- `submit-management-enhancement`: 提交管理 Tab 增强——操作类型列、撤回按钮、筛选栏调整

### Modified Capabilities
（无需修改现有 spec，本次改动不涉及已有能力的行为变更）

## Impact

- **后端**：`services/masterdata` — types.go（Division 结构体增加 submission_type）、model 层（新增/修改查询和更新方法）、logic 层（create/update/delete/submit/review 逻辑调整）、新增 3 个 handler+logic
- **前端**：`web/pc/src/views/grassroots/Index.vue` — 操作列按钮逻辑重写、筛选栏调整、表格列增加；`web/common/types/masterdata.d.ts` — 类型定义补充
- **API 层**：`web/pc/src/api/masterdata.ts` — 新增 3 个 API 调用函数
- **数据库**：`md_administrative_division` 表的 `submission_type` 字段已存在，无需 DDL 变更；`submission_status=4` 的历史数据需迁移（改为 submission_status=0, submission_type=3）
- **审批中心**：审批通过删除申请时需执行软删除，审批拒绝删除申请时直接恢复已批准状态
