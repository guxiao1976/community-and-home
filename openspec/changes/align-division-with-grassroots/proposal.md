## Why

行政区划管理页面目前仅有基础的树形展示和简单 CRUD，缺少审批工作流。基层组织模块已实现完整的 提交→审批→撤回 流程，需要将相同的界面模式和业务逻辑同步到行政区划模块，保持两个模块的操作体验一致。

## What Changes

- **前端**：重写 `division/Index.vue`，对齐基层组织模块的三 Tab 结构（查询编辑、提交管理、提交记录）
- **前端**：补充操作按钮权限控制（基于 submission_status + submission_type），增加批量提交、撤回、发起删除、取消删除操作
- **前端**：增加提交记录查询 Tab
- **前端**：删除操作复用已有的 `requestDeleteDivision`、`cancelDeleteDivision`、`withdrawDivision`、`batchSubmitDivisions` API
- **后端**：无需改动，所有 API 端点和审批逻辑已就绪，前端直接调用即可

## Capabilities

### New Capabilities
- `division-submission-ui`: 行政区划提交管理 UI（三 Tab 结构、权限控制、批量操作、提交记录）

### Modified Capabilities

（无后端需求变更）

## Impact

- **前端文件**：`web/pc/src/views/division/Index.vue`（重写）
- **前端 API**：`web/pc/src/api/masterdata.ts`（复用已有函数，无改动）
- **前端类型**：`web/common/types/masterdata.d.ts`（复用已有类型，无改动）
- **后端**：无改动
