## ADDED Requirements

### Requirement: 三 Tab 页面结构
行政区划管理页面 SHALL 包含三个 Tab：查询编辑、提交管理、提交记录。Tab 结构与基层组织模块一致。

#### Scenario: 切换 Tab 加载数据
- **WHEN** 用户切换到"提交管理"或"提交记录"Tab
- **THEN** 系统自动加载对应数据

### Requirement: 查询编辑 Tab 操作权限
查询编辑 Tab 中的操作按钮 SHALL 基于 submission_status 和 submission_type 控制可见性，规则与基层组织一致：
- 待提交(0) 且非删除类型：显示"编辑"
- 待提交(0) 且新增类型(1)：显示"删除"（物理删除）
- 已批准(2)：显示"编辑"、"添加下级"、"删除"（发起删除）
- 已拒绝(3)：显示"编辑"
- 待提交(0) 且删除类型(3)：显示"取消删除"

#### Scenario: 已批准区划可发起删除
- **WHEN** 用户对 submission_status=2 的区划点击"删除"
- **THEN** 系统调用 requestDeleteDivision API，状态变为待提交+删除类型

#### Scenario: 新增数据未提交可物理删除
- **WHEN** 用户对 submission_status=0 且 submission_type=1 的区划点击"删除"
- **THEN** 系统直接物理删除该区划

### Requirement: 提交管理 Tab 功能
提交管理 Tab SHALL 展示 level 1-3 中待提交和已提交的区划数据，支持单条提交、批量提交和撤回操作。

#### Scenario: 批量提交
- **WHEN** 用户勾选多条待提交数据并点击"批量提交"
- **THEN** 系统调用 batchSubmitDivisions API

#### Scenario: 撤回已提交数据
- **WHEN** 用户对 submission_status=1 的区划点击"撤回"
- **THEN** 系统调用 withdrawDivision API

### Requirement: 提交记录 Tab
提交记录 Tab SHALL 调用 getMySubmissionRecords API，按审核结果筛选，展示 entity_type=administrative_division 的记录。

#### Scenario: 查看提交记录
- **WHEN** 用户切换到"提交记录"Tab
- **THEN** 显示 entity_type=administrative_division 的提交记录列表

### Requirement: 数据刷新与状态保持
操作（新增、编辑、删除、发起删除、取消删除）完成后 SHALL 刷新当前列表，并保持树形展开状态。仅在用户已搜索过数据时才刷新。

#### Scenario: 操作后刷新
- **WHEN** 用户已搜索且列表有数据，执行编辑操作
- **THEN** 列表刷新并保持展开状态

#### Scenario: 未搜索时操作不刷新
- **WHEN** 用户未搜索（列表为空），新增一条区划
- **THEN** 列表保持为空，不自动填充
