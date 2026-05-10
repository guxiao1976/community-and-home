## 1. 重写 division/Index.vue 基础结构

- [x] 1.1 以 grassroots/Index.vue 为蓝本重写 division/Index.vue，建立三 Tab 结构（查询编辑、提交管理、提交记录）
- [x] 1.2 查询编辑 Tab：树形表格展示全部行政区划（level 1-3），懒加载子节点，支持编辑、删除操作
- [x] 1.3 权限控制函数：canEdit、canAddChild、canDelete、canCancelDelete（基于 submission_status + submission_type）

## 2. 查询编辑 Tab 完整功能

- [x] 2.1 新增/编辑对话框：区划代码（新增时可编辑、编辑时禁用）、区划名称、上级区划（新增下级时显示）
- [x] 2.2 删除逻辑：待提交+新增类型→物理删除，已批准→发起删除（requestDeleteDivision）
- [x] 2.3 取消删除按钮（submission_status=0 且 submission_type=3 时显示）
- [x] 2.4 操作完成后刷新列表并保持展开状态

## 3. 提交管理 Tab

- [x] 3.1 状态筛选：待提交、已提交两个 Radio 切换
- [x] 3.2 数据表格：展示 level 1-3 的区划数据，支持勾选
- [x] 3.3 单条提交、批量提交、撤回、物理删除操作按钮
- [x] 3.4 分页

## 4. 提交记录 Tab

- [x] 4.1 调用 getMySubmissionRecords API（entity_type=administrative_division）
- [x] 4.2 按审核结果筛选（全部、待审核、已通过、已拒绝、已撤回）
- [x] 4.3 分页

## 5. 清理

- [x] 5.1 保留 division store 的 createDivision、updateDivision、deleteDivision（仍被对话框和物理删除使用）
- [x] 5.2 后端无需改动，所有 API 端点已复用
