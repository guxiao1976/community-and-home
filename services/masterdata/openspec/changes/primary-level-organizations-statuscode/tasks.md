## 1. 后端 Types & API 定义

- [x] 1.1 `types.go` Division 结构体增加 `SubmissionType *int64` 字段（json tag: `submission_type`）
- [x] 1.2 `masterdata.api` 新增 3 条路由：POST `/divisions/:id/request-delete`、POST `/divisions/:id/cancel-delete`、POST `/divisions/:id/withdraw`
- [x] 1.3 运行 goctl 生成 types、handler、logic 骨架代码
- [x] 1.4 手动补全生成的 types：`RequestDeleteResp`、`CancelDeleteResp`、`WithdrawResp`（均为 `{ Success bool }`）

## 2. 后端 Model 层

- [x] 2.1 model 新增 `UpdateStatusAndType(ctx, id, submissionStatus, submissionType int64) error` 方法
- [x] 2.2 model 新增 `FindById(ctx, id int64)` 方法（如不存在），供 logic 层读取当前状态做校验
- [x] 2.3 model 新增 `UpdateStatus(ctx, id, submissionStatus int64) error` 方法（仅改 status，不改 type）

## 3. 后端新增 3 个接口 Logic

- [x] 3.1 `requestDeleteLogic`：校验当前 `submission_status=2`，调用 `UpdateStatusAndType(id, 0, 3)`
- [x] 3.2 `cancelDeleteLogic`：校验当前 `submission_status=0, submission_type=3`，调用 `UpdateStatus(id, 2)`
- [x] 3.3 `withdrawLogic`：校验当前 `submission_status=1`，调用 `UpdateStatus(id, 0)`

## 4. 后端修改现有 Logic

- [x] 4.1 `createDivisionLogic`：新增数据时设置 `submission_type=1`（检查是否已设置，未设置则补上）
- [x] 4.2 `updateDivisionLogic`：保存时检查当前状态，若 `submission_status=2` 则改为 `submission_status=0, submission_type=2`；若 `submission_status=0` 则仅更新内容
- [x] 4.3 `deleteDivisionLogic`：增加前置校验，仅允许 `submission_status=0 AND submission_type=1`，不满足时返回 403 错误
- [x] 4.4 `getDivisionsLogic`：返回数据时填充 `SubmissionType` 字段（从 model 的 `sql.NullInt64` 转为 `*int64`）

## 5. 后端审批中心适配

- [x] 5.1 `reviewItemLogic`：审批通过时，若 entity_type=administrative_division 且 submission_type=3，执行 SoftDelete
- [x] 5.2 `reviewItemLogic`：审批拒绝时，若 entity_type=administrative_division 且 submission_type=3，直接设 submission_status=2（恢复已批准）

## 6. 数据迁移

- [x] 6.1 编写 SQL 迁移脚本：`UPDATE md_administrative_division SET submission_status=0, submission_type=3 WHERE submission_status=4`
- [x] 6.2 更新 `masterdata_schema.sql`，同步 submission_type 等字段的 DDL 定义

## 7. 后端编译 & 部署

- [x] 7.1 编译 masterdata 服务确认无错误
- [x] 7.2 执行 SQL 迁移脚本
- [x] 7.3 重启 masterdata-api 服务

## 8. 前端类型 & API 层

- [x] 8.1 `web/common/types/masterdata.d.ts` AdministrativeDivision 接口增加 `submission_type: number | null`
- [x] 8.2 `web/pc/src/api/masterdata.ts` 新增 `requestDeleteDivision`、`cancelDeleteDivision`、`withdrawDivision` 三个 API 函数

## 9. 前端查询编辑 Tab 改造

- [x] 9.1 重写 helper 函数组：`canEdit`、`canDelete`、`canSubmit`、`canWithdraw`、`canCancelDelete`、`canAddChild`，按权限矩阵实现
- [x] 9.2 状态筛选栏移除「待删除」按钮，保留：全部/待提交/已提交/已批准/已拒绝
- [x] 9.3 表格操作列按 helper 函数动态渲染按钮：编辑、删除（物理删除/发起删除）、取消删除、添加下级
- [x] 9.4 新增「操作类型」列，仅在 submission_status=0 时显示 Tag（新增蓝/修改橙/删除红）
- [x] 9.5 `handleDelete` 函数区分两种场景：新建待提交弹"物理删除确认"并调用 DELETE，已批准弹"发起删除确认"并调用 requestDelete
- [x] 9.6 新增 `handleCancelDelete` 函数，调用 cancelDelete API
- [x] 9.7 新增 `handleSubmit` 后刷新逻辑兼容新状态体系
- [x] 9.8 `submissionStatusMap` 移除 key=4，更新标签映射

## 10. 前端提交管理 Tab 改造

- [x] 10.1 状态筛选栏移除「待删除」按钮，与查询编辑 Tab 保持一致
- [x] 10.2 新增「操作类型」列，Tag 样式同查询编辑 Tab
- [x] 10.3 操作列增加「撤回」按钮（仅 submission_status=1 时显示），调用 withdraw API
- [x] 10.4 待提交行的删除按钮增加条件：仅 submission_type=1（新建待提交）时显示
- [x] 10.5 批量提交逻辑不变，确认三种 type 均可批量提交
