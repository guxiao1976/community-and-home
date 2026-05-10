## Why

目前主数据管理系统的 4 张表（住宅小区、行政区划、系统配置、敏感词）各自独立管理审批流程，缺少统一的审核入口。只有住宅小区有完整的审核逻辑，其余 3 个实体仅有提交功能。同时，系统无法区分"新增""修改""删除"操作类型，修改数据时也无法展示变更前后的对比内容，审核人难以做出准确判断。

## What Changes

- 4 张主数据表新增 `submission_type` 字段（1=新增, 2=修改, 3=删除），区分操作类型
- 4 张主数据表新增 `change_snapshot` 字段（TEXT/JSON），存储修改前的数据快照
- 3 张表（行政区划、系统配置、敏感词）补齐审核人字段（reviewer_id, review_time, review_notes）和提交人字段（submitter_id, submit_time）
- 敏感词表补齐 `delete_time` 软删除字段
- 新增统一审核 API 路由组（`/api/masterdata/approval/*`），提供待审列表、数量统计、详情查看、单条/批量审批
- 修改现有 Create/Update/Delete/Submit 逻辑：Create 设置 submission_type=1，Update 保存快照并设置 submission_type=2，Delete 标记待删除（submission_status=4, submission_type=3）而非直接软删除
- 行政区划、系统配置、敏感词的 Delete 逻辑改为审批制（与住宅小区一致）
- 新增前端统一审核中心页面，含统计卡片、待审列表、变更对比抽屉、批量审批功能

## Capabilities

### New Capabilities
- `approval-center-api`: 统一审核后端 API（待审聚合查询、数量统计、详情查看、审批操作）
- `approval-center-ui`: 统一审核前端页面（统计卡片、待审列表表格、变更对比抽屉、批量审批）
- `submission-tracking`: 操作类型区分与变更快照（submission_type 字段、change_snapshot 快照存储、修改前数据捕获、拒绝时快照恢复）
- `delete-approval`: 删除走审批流程（行政区划、系统配置、敏感词的删除改为标记待删除，审批通过后软删除）

### Modified Capabilities
<!-- No existing specs to modify -->

## Impact

**数据库**: 4 张表结构变更（ALTER TABLE），需执行迁移脚本
**后端 Model**: 4 个 `_gen.go` 文件修改 struct 和 SQL，4 个 custom model 文件新增查询方法
**后端 API**: `masterdata.api` 新增 approval 路由组，goctl 重新生成 handler/logic/types/routes
**后端 Logic**: 4 个实体的 create/update/delete/submit logic（共约 16 个文件需修改）
**前端**: 新增 3 个 Vue 组件（approval-center），修改 types、API、router 文件
**兼容性**: 现有住宅小区审核页面 `Review.vue` 可保留为快捷入口或废弃，统一入口为新的审核中心
