## 1. 数据库迁移

- [x] 1.1 编写并执行 SQL 迁移脚本：4 张表各加 `submission_type TINYINT NULL` 和 `change_snapshot TEXT NULL`
- [x] 1.2 md_administrative_division 补齐 submitter_id、submit_time、reviewer_id、review_time、review_notes 字段
- [x] 1.3 md_configuration 补齐 submitter_id、submit_time、reviewer_id、review_time、review_notes 字段
- [x] 1.4 md_sensitive_word 补齐 submitter_id、submit_time、reviewer_id、review_time、review_notes、delete_time 字段
- [x] 1.5 回填历史数据：已通过记录设 submission_type=1

## 2. 后端 Model 层

- [x] 2.1 修改 mdResidentialAreaModel_gen.go：struct 加 SubmissionType、ChangeSnapshot 字段，修改 Insert/Update SQL
- [x] 2.2 修改 mdAdministrativeDivisionModel_gen.go：struct 加 SubmissionType、ChangeSnapshot、SubmitterId、SubmitTime、ReviewerId、ReviewTime、ReviewNotes 字段，修改 Insert/Update SQL
- [x] 2.3 修改 mdConfigurationModel_gen.go：struct 加 SubmissionType、ChangeSnapshot、SubmitterId、SubmitTime、ReviewerId、ReviewTime、ReviewNotes 字段，修改 Insert/Update SQL
- [x] 2.4 修改 mdSensitiveWordModel_gen.go：struct 加 SubmissionType、ChangeSnapshot、SubmitterId、SubmitTime、ReviewerId、ReviewTime、ReviewNotes、DeleteTime 字段，修改 Insert/Update SQL
- [x] 2.5 4 个 custom model 各加 CountBySubmissionStatus 和 FindPendingBySubmissionStatus 方法

## 3. 后端 API 定义与代码生成

- [x] 3.1 修改 masterdata.api：新增 approval 路由组（pending-counts、pending-items、detail、review、batch-review）
- [x] 3.2 运行 goctl api go 生成 handler/logic/types/routes 骨架
- [x] 3.3 修复 goctl 生成后 handler 文件的 responsex.Response 包装

## 4. 统一审核 API Logic 实现

- [x] 4.1 实现 getPendingCountsLogic：查询 4 张表 submission_status=1 的记录数
- [x] 4.2 实现 getPendingItemsLogic：按 entity_type 过滤查对应表，不过滤时查 4 张表合并排序分页
- [x] 4.3 实现 getApprovalDetailLogic：按 entity_type+id 查记录，序列化 current_data 和 snapshot_data
- [x] 4.4 实现 reviewItemLogic：按 entity_type 分发，处理通过（含删除类型软删除）、拒绝（含修改类型快照恢复）
- [x] 4.5 实现 batchReviewItemsLogic：遍历 ids 调用 reviewItem 逻辑，统计成功/失败数

## 5. 修改现有 CRUD 逻辑 — submission_tracking

- [x] 5.1 修改 4 个 create logic：设置 SubmissionType=1（新增）
- [x] 5.2 修改 4 个 update logic：修改前序列化业务字段到 ChangeSnapshot，设置 SubmissionType=2（修改）
- [x] 5.3 修改 3 个 submit logic（division、configuration、sensitiveword）：补齐 submitter_id、submit_time

## 6. 修改现有 Delete 逻辑 — delete_approval

- [x] 6.1 修改 division/deleteDivisionLogic：改为设置 submission_status=4、submission_type=3
- [x] 6.2 修改 configuration/deleteConfigurationLogic：改为设置 submission_status=4、submission_type=3
- [x] 6.3 修改 sensitiveword/deleteSensitiveWordLogic：改为设置 submission_status=4、submission_type=3
- [x] 6.4 3 张表的默认列表查询排除 submission_status=4 的记录

## 7. 构建部署后端

- [x] 7.1 编译 masterdata-api 并解决编译错误
- [x] 7.2 编译 masterdata-rpc 并解决编译错误
- [ ] 7.3 重启服务

## 8. 前端类型与 API

- [x] 8.1 修改 web/common/types/masterdata.d.ts：新增 SubmissionType 枚举、ApprovalPendingItem、PendingCounts、ApprovalDetail 类型，更新 SubmissionStatus 加 PendingDelete=4
- [x] 8.2 修改 web/pc/src/api/masterdata.ts：新增 getPendingCounts、getPendingItems、getApprovalDetail、reviewItem、batchReviewItems 函数

## 9. 前端审核中心页面

- [x] 9.1 新增路由 `/masterdata/approval-center` 指向 approval-center/Index.vue
- [x] 9.2 实现 approval-center/Index.vue：统计卡片 + entity_type 过滤标签 + 操作类型过滤 + 待审列表表格 + 批量审批按钮
- [x] 9.3 实现 approval-center/DetailDrawer.vue：按 submission_type 展示详情（新增/修改/删除三种视图）
- [x] 9.4 实现 approval-center/ChangeComparison.vue：字段名、修改前值、修改后值对比表格，变更字段高亮
- [x] 9.5 DetailDrawer 内集成审批操作（通过/拒绝按钮 + 备注输入）

## 10. 集成验证

- [ ] 10.1 验证新建→提交→审核通过流程
- [ ] 10.2 验证修改→保存快照→提交→审核查看变更对比→通过/拒绝恢复
- [ ] 10.3 验证删除→标记待删除→提交→审核通过软删除 / 拒绝恢复
- [ ] 10.4 验证前端审核中心统计卡片、列表过滤、批量审批功能
- [ ] 10.5 验证 4 个实体的列表默认排除 submission_status=4 记录，主动选择待删除可查看
