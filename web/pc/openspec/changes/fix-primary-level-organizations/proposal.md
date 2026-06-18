## Why

基层组织页面将"查询编辑"和"提交"两个工作流混在同一个视图中。查询编辑以地理位置（省-市-区县）为核心维度，提交以审批状态为核心维度，两者意图不同。当前状态筛选和区县级联选择器并列时，提交时用户需手动清空区县条件才能看到全部待提交数据，操作步骤反直觉，体验割裂。

## What Changes

- 将基层组织页面重构为双 Tab 布局：**查询编辑** 和 **提交管理**
- **Tab 1 查询编辑**：保留现有省市县级联选择器 + 树形表格，移除提交按钮，状态筛选作为可选辅助过滤
- **Tab 2 提交管理**：独立视图，以状态筛选为主维度，平铺表格展示所有匹配记录（含完整路径），支持单条和批量提交
- **BREAKING**：提交操作从查询编辑视图中移除，统一到提交管理 Tab

## Capabilities

### New Capabilities
- `grassroots-submit-management`: 基层组织提交管理 Tab，包含状态筛选、平铺表格（含完整路径）、单条提交和批量提交

### Modified Capabilities
- `grassroots-management`: 基层组织查询编辑 Tab，移除提交按钮，状态筛选改为可选辅助过滤

## Impact

- **前端**：`views/grassroots/Index.vue` 重构为双 Tab 结构
- **后端**：提交管理 Tab 可能需要新 API（批量查询按状态过滤的街道/社区列表），复用现有 `getAdministrativeDivisions` 增加参数或新增专用接口
- **API 层**：可能需要在 divisions 列表接口中支持仅查询 level 4-5 的数据
- **前端 API**：`src/api/masterdata.ts` 可能需要新增按状态查询基层组织的函数
