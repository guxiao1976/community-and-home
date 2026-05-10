## Why

主数据模块的审核流程中，数据维护员提交的新增/修改/删除请求在审核完成后缺乏留痕。特别是新增数据被拒绝后执行物理删除，提交者无法查看审核结果和拒绝原因。需要建立独立的提交记录机制，覆盖所有主数据实体的审核流程。

## What Changes

- 新建 `md_submission_record` 表，独立记录每次提交审核的完整生命周期
- 在后端提交、审核（通过/拒绝）、撤回逻辑中写入记录
- 前端基层组织页面新增"提交记录"Tab，展示当前用户的提交历史
- 审核中心新增"审核记录"Tab，展示审核人的历史操作

## Capabilities

### New Capabilities
- `submission-record-table`: 提交记录表的 DDL 定义、Model 层（CRUD）、后端写入逻辑
- `submission-record-api`: 查询提交记录的 API（按用户/审核人过滤、分页）
- `submission-record-ui`: 前端提交记录展示（基层组织和审核中心页面）

### Modified Capabilities
<!-- 无需修改现有 spec -->

## Impact

- **数据库**: 新增 `md_submission_record` 表
- **后端 Model**: 新增 submission_record 的 model 文件
- **后端 Logic**: 修改 `reviewItemLogic`、`submitDivisionLogic`、`withdrawDivisionLogic` 等逻辑，增加记录写入
- **后端 API**: 新增查询提交记录的路由和 handler/logic
- **前端 API**: 新增提交记录查询函数
- **前端页面**: 基层组织 Index.vue 新增 Tab；审核中心 Index.vue 新增 Tab
