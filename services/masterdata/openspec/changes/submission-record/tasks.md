## 1. 数据库

- [x] 1.1 编写 `md_submission_record` 建表 DDL 并添加到 `scripts/sql/masterdata_schema.sql`
- [x] 1.2 在数据库中执行建表语句

## 2. 后端 Model 层

- [x] 2.1 在 `model/` 下手动编写 `mdSubmissionRecordModel.go`（Insert、UpdateResult、FindBySubmitter、FindByReviewer 方法）
- [x] 2.2 在 `svc/servicecontext.go` 中注册 `MdSubmissionRecordModel`

## 3. 后端记录写入逻辑

- [x] 3.1 `submitDivisionLogic`：提交成功后插入一条 submission_record
- [x] 3.2 `batchSubmitDivisionsLogic`：批量提交成功后逐条插入 submission_record
- [x] 3.3 `reviewItemLogic.reviewDivision`：审核通过/拒绝时更新 submission_record 的 review_result；type=1 拒绝物理删除前先更新记录
- [x] 3.4 `withdrawDivisionLogic`：撤回时更新 submission_record 的 review_result=3
- [x] 3.5 `submitResidentialAreaLogic`：提交成功后插入 submission_record
- [x] 3.6 `reviewItemLogic.reviewResidentialArea`：审核通过/拒绝时更新 submission_record
- [x] 3.7 `submitConfigurationLogic`：提交成功后插入 submission_record
- [x] 3.8 `reviewItemLogic.reviewConfiguration`：审核通过/拒绝时更新 submission_record
- [x] 3.9 `submitSensitiveWordLogic`：提交成功后插入 submission_record
- [x] 3.10 `reviewItemLogic.reviewSensitiveWord`：审核通过/拒绝时更新 submission_record

## 4. 后端查询 API

- [x] 4.1 `masterdata.api` 新增 2 条路由：GET `/submission-records/my`、GET `/submission-records/reviewed`
- [x] 4.2 `types.go` 新增 `SubmissionRecord` 类型和相关请求/响应结构体
- [x] 4.3 运行 goctl 生成 handler 骨架
- [x] 4.4 实现 `getMySubmissionRecordsLogic`（按 submitter_id 查询，支持分页和过滤）
- [x] 4.5 实现 `getReviewedSubmissionRecordsLogic`（按 reviewer_id 查询，支持分页和过滤）

## 5. 后端编译部署

- [x] 5.1 编译 masterdata 服务确认无错误
- [x] 5.2 重启 masterdata-api 服务

## 6. 前端类型 & API

- [x] 6.1 `web/common/types/masterdata.d.ts` 新增 `SubmissionRecord` 接口
- [x] 6.2 `web/pc/src/api/masterdata.ts` 新增 `getMySubmissionRecords`、`getReviewedSubmissionRecords` 函数

## 7. 前端基层组织页面

- [x] 7.1 基层组织 Index.vue 新增"提交记录"Tab，表格展示：实体名称、代码、操作类型 Tag、提交时间、审核结果 Tag、审核备注
- [x] 7.2 支持按审核结果筛选和分页

## 8. 前端审核中心页面

- [x] 8.1 审核中心 Index.vue 新增"审核记录"Tab，表格展示：实体名称、代码、实体类型、操作类型、提交人、提交时间、审核结果、审核时间、审核备注
- [x] 8.2 支持按审核结果筛选和分页
