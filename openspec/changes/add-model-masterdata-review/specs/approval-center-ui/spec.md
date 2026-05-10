## ADDED Requirements

### Requirement: 审核中心统计卡片
审核中心页面顶部 MUST 显示 4 个统计卡片，分别展示住宅小区、行政区划、系统配置、敏感词的待审数量。

#### Scenario: 页面加载获取待审数量
- **WHEN** 用户进入审核中心页面
- **THEN** 页面 SHALL 调用 pending-counts API，在 4 个卡片上显示各实体的待审数量

#### Scenario: 点击卡片过滤列表
- **WHEN** 用户点击某个统计卡片
- **THEN** 待审列表 SHALL 按对应 entity_type 过滤，高亮对应的过滤标签

### Requirement: 审核中心待审列表
审核中心 MUST 显示统一格式的待审列表表格，支持多维度过滤和分页。

#### Scenario: 列表默认显示全部待审
- **WHEN** 页面加载完成
- **THEN** 列表 SHALL 显示所有 submission_status=1 的待审记录，按 submit_time 降序排列

#### Scenario: 按实体类型过滤
- **WHEN** 用户选择"住宅小区"过滤标签
- **THEN** 列表 SHALL 仅显示 entity_type=residential_area 的待审记录

#### Scenario: 按操作类型过滤
- **WHEN** 用户选择"新增"过滤标签
- **THEN** 列表 SHALL 仅显示 submission_type=1（新增）的待审记录

#### Scenario: 列表列显示
- **THEN** 列表 SHALL 包含以下列：选择框、数据类型（Tag）、操作类型（Tag：新增/修改/待删除）、名称、变更摘要、提交人、提交时间、操作（详情按钮）

### Requirement: 审核中心批量审批
审核中心 MUST 支持批量通过和批量拒绝操作。

#### Scenario: 批量选择记录
- **WHEN** 用户勾选列表中的多条待审记录
- **THEN** 页面 SHALL 显示已选数量和"批量通过""批量拒绝"按钮

#### Scenario: 批量通过
- **WHEN** 用户点击"批量通过"
- **THEN** 页面 SHALL 弹出确认对话框，确认后调用 batch-review API（action=approve），成功后刷新列表

#### Scenario: 批量拒绝
- **WHEN** 用户点击"批量拒绝"
- **THEN** 页面 SHALL 弹出对话框要求填写拒绝原因（必填），确认后调用 batch-review API（action=reject, review_notes），成功后刷新列表

### Requirement: 审批详情抽屉
审核中心 MUST 提供详情抽屉展示待审记录的完整信息和变更对比。

#### Scenario: 查看修改类型详情
- **WHEN** 用户点击一条 submission_type=2（修改）记录的"详情"按钮
- **THEN** 页面 SHALL 打开 Drawer，展示变更对比表格（字段名、修改前值、修改后值），修改过的字段 SHALL 用颜色高亮

#### Scenario: 查看新增类型详情
- **WHEN** 用户点击一条 submission_type=1（新增）记录的"详情"按钮
- **THEN** 页面 SHALL 打开 Drawer，展示当前记录的所有业务字段值

#### Scenario: 查看删除类型详情
- **WHEN** 用户点击一条 submission_type=3（删除）记录的"详情"按钮
- **THEN** 页面 SHALL 打开 Drawer，展示将被删除的记录内容，并显示"将执行软删除"警告

#### Scenario: 在详情抽屉中审批
- **WHEN** 详情抽屉打开且记录状态为已提交
- **THEN** 抽屉底部 SHALL 显示"通过""拒绝"按钮和审核备注输入框，支持直接审批

### Requirement: 审核中心路由
系统 MUST 添加审核中心前端路由。

#### Scenario: 访问审核中心
- **WHEN** 用户访问 `/masterdata/approval-center`
- **THEN** 系统 SHALL 渲染审核中心页面
