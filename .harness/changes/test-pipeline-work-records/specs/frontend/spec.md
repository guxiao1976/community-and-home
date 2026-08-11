# Frontend Specification — 测试流水线工作记录模块

## Purpose

为测试流水线验证提供工作记录管理功能的前端界面。在 PC 管理后台新增一级菜单"AI-Coding自我提升"，包含工作事项维护和工作内容记录两个页面，支持 CRUD 操作、日期范围查询和关联选择。前端通过 REST API 与 master-data-service 交互。

## Requirements

---

### Requirement: 路由配置

The system SHALL add a new first-level menu "AI-Coding自我提升" with two sub-pages in the PC admin portal.

#### Scenario: 菜单结构
- **GIVEN** 前端路由配置文件 `@/config/route.config.ts`
- **WHEN** 加载应用
- **THEN** 侧边栏显示一级菜单"AI-Coding自我提升"，包含子菜单：
  - "工作事项维护" (路由 `/pipeline-test/work-items`)
  - "工作内容记录" (路由 `/pipeline-test/work-records`)

#### Scenario: 路由权限
- **GIVEN** 用户已登录
- **WHEN** 访问 `/pipeline-test/work-items` 或 `/pipeline-test/work-records`
- **THEN** 无需特殊权限检查（测试模块，管理员可访问）
- **AND** 页面正常加载

#### Scenario: 菜单图标
- **GIVEN** 菜单配置
- **WHEN** 渲染侧边栏
- **THEN** 使用图标 `el-icon-document` 或类似（工作记录相关图标）

---

### Requirement: TypeScript 类型定义

The system SHALL define TypeScript interfaces for all data models with proper ID string types.

#### Scenario: 工作事项类型
- **GIVEN** TypeScript 类型文件 `@/api/pipeline-test.ts`
- **WHEN** 定义 `WorkItem` 接口
- **THEN** 包含字段：
  ```typescript
  interface WorkItem {
    id?: string;            // Snowflake ID as string
    name: string;
    sort_order: number;
    status: number;         // 0=禁用 1=启用
    created_by?: string;
    created_time?: string;
    updated_time?: string;
    delete_time?: string;
  }
  ```

#### Scenario: 工作记录类型
- **GIVEN** TypeScript 类型文件 `@/api/pipeline-test.ts`
- **WHEN** 定义 `WorkRecord` 接口
- **THEN** 包含字段：
  ```typescript
  interface WorkRecord {
    id?: string;
    work_item_id: string;
    work_item_name?: string;  // 用于列表展示
    work_date: string;        // YYYY-MM-DD
    description: string;
    duration_hours: number;
    created_by?: string;
    created_time?: string;
    updated_time?: string;
    delete_time?: string;
  }
  ```

#### Scenario: 查询参数类型
- **GIVEN** TypeScript 类型文件
- **WHEN** 定义查询参数接口
- **THEN** 包含：
  ```typescript
  interface WorkItemQuery {
    page?: number;
    page_size?: number;
    status?: number;
    include_disabled?: boolean;
  }

  interface WorkRecordQuery {
    page?: number;
    page_size?: number;
    start_date?: string;
    end_date?: string;
    work_item_id?: string;
  }
  ```

---

### Requirement: API 函数实现

The system SHALL implement API functions using axios with proper type safety.

#### Scenario: 工作事项 API 函数
- **GIVEN** API 文件 `@/api/pipeline-test.ts`
- **WHEN** 定义 API 函数
- **THEN** 包含函数：
  ```typescript
  export const getWorkItems = (params: WorkItemQuery) => 
    request.get<{ list: WorkItem[], total: number }>('/api/masterdata/pipeline-test/work-items', { params });
  
  export const getWorkItem = (id: string) => 
    request.get<WorkItem>(`/api/masterdata/pipeline-test/work-items/${id}`);
  
  export const createWorkItem = (data: WorkItem) => 
    request.post<WorkItem>('/api/masterdata/pipeline-test/work-items', data);
  
  export const updateWorkItem = (id: string, data: WorkItem) => 
    request.put<WorkItem>(`/api/masterdata/pipeline-test/work-items/${id}`, data);
  
  export const deleteWorkItem = (id: string) => 
    request.delete(`/api/masterdata/pipeline-test/work-items/${id}`);
  ```

#### Scenario: 工作记录 API 函数
- **GIVEN** API 文件 `@/api/pipeline-test.ts`
- **WHEN** 定义 API 函数
- **THEN** 包含函数：
  ```typescript
  export const getWorkRecords = (params: WorkRecordQuery) => 
    request.get<{ list: WorkRecord[], total: number }>('/api/masterdata/pipeline-test/work-records', { params });
  
  export const getWorkRecord = (id: string) => 
    request.get<WorkRecord>(`/api/masterdata/pipeline-test/work-records/${id}`);
  
  export const createWorkRecord = (data: WorkRecord) => 
    request.post<WorkRecord>('/api/masterdata/pipeline-test/work-records', data);
  
  export const updateWorkRecord = (id: string, data: WorkRecord) => 
    request.put<WorkRecord>(`/api/masterdata/pipeline-test/work-records/${id}`, data);
  
  export const deleteWorkRecord = (id: string) => 
    request.delete(`/api/masterdata/pipeline-test/work-records/${id}`);
  ```

#### Scenario: axios 拦截器处理
- **GIVEN** axios 响应拦截器已配置
- **WHEN** API 返回 `{code: 0, msg: "success", data: {...}}`
- **THEN** 拦截器提取 `data` 字段返回给调用方
- **AND** 调用方直接获得业务数据，无需 `.data.data`

---

### Requirement: 工作事项维护页面

The system SHALL provide a work items management page at `/pipeline-test/work-items`.

#### Scenario: 页面布局
- **GIVEN** 用户访问工作事项维护页面
- **WHEN** 页面加载
- **THEN** 显示：
  - 页面标题 "工作事项维护"
  - 操作按钮："新增"（主按钮）
  - 表格列：序号、工作事项名称、排序号、状态（启用/禁用标签）、操作（编辑/删除）
  - 分页器（如总数 > 10）

#### Scenario: 新增工作事项
- **GIVEN** 用户点击"新增"按钮
- **WHEN** 弹出对话框
- **THEN** 显示表单字段：
  - 工作事项名称（必填，文本输入框，最大 100 字符）
  - 排序号（必填，数字输入框，默认 0，范围 0-9999）
  - 状态（单选，默认"启用"）
- **AND** 表单包含"确定"和"取消"按钮

#### Scenario: 新增工作事项（提交成功）
- **GIVEN** 用户填写表单：name="需求分析", sort_order=10, status=1
- **WHEN** 点击"确定"
- **THEN** 调用 `createWorkItem()`
- **AND** 成功后关闭对话框，显示成功提示"新增成功"
- **AND** 刷新列表，新记录出现在第一页

#### Scenario: 新增工作事项（表单校验失败）
- **GIVEN** 用户填写表单但工作事项名称为空
- **WHEN** 点击"确定"
- **THEN** 前端表单校验失败，显示错误提示"工作事项名称不能为空"
- **AND** 对话框不关闭

#### Scenario: 新增工作事项（后端返回错误）
- **GIVEN** 用户提交重复的工作事项名称
- **WHEN** 后端返回 400 错误
- **THEN** 显示错误提示"工作事项名称已存在"
- **AND** 对话框不关闭

#### Scenario: 编辑工作事项
- **GIVEN** 列表中存在工作事项 ID 123
- **WHEN** 用户点击"编辑"按钮
- **THEN** 弹出对话框，表单字段回显该工作事项的当前值
- **AND** 用户修改后点击"确定"，调用 `updateWorkItem(123, data)`
- **AND** 成功后刷新列表

#### Scenario: 编辑回显数据完整性检查
- **GIVEN** 工作事项包含字段：name, sort_order, status
- **WHEN** 打开编辑对话框
- **THEN** 所有字段正确回显：
  - name 输入框显示当前名称
  - sort_order 输入框显示当前排序号
  - status 单选按钮选中当前状态
- **AND** 不出现字段为空或默认值的情况（参考 [[edit-form-data-integrity]]）

#### Scenario: 删除工作事项（确认）
- **GIVEN** 列表中存在工作事项 ID 123
- **WHEN** 用户点击"删除"按钮
- **THEN** 弹出确认对话框"确定要删除该工作事项吗？"
- **AND** 用户点击"确定"，调用 `deleteWorkItem(123)`
- **AND** 成功后显示"删除成功"，刷新列表

#### Scenario: 删除工作事项（存在关联记录）
- **GIVEN** 工作事项 ID 123 有关联工作记录
- **WHEN** 用户点击删除并确认
- **THEN** 后端返回 400 错误
- **AND** 显示错误提示"该工作事项存在关联记录，无法删除"

#### Scenario: 状态筛选
- **GIVEN** 页面加载
- **WHEN** 用户点击状态筛选下拉框，选择"仅启用"
- **THEN** 查询参数 `status=1`，列表仅显示启用的工作事项

#### Scenario: 状态切换（快捷操作）
- **GIVEN** 列表显示工作事项，状态列为标签（绿色"启用"/灰色"禁用"）
- **WHEN** 用户点击状态标签
- **THEN** 调用 `updateWorkItem(id, {status: 1-current_status})`
- **AND** 成功后标签切换颜色，列表不刷新（本地更新）

---

### Requirement: 工作内容记录页面

The system SHALL provide a work records management page at `/pipeline-test/work-records`.

#### Scenario: 页面布局
- **GIVEN** 用户访问工作内容记录页面
- **WHEN** 页面加载
- **THEN** 显示：
  - 页面标题 "工作内容记录"
  - 查询表单：日期范围选择器（带快捷选项）、工作事项下拉框、查询/重置按钮
  - 操作按钮："新增"（主按钮）
  - 表格列：序号、工作事项、工作日期、工作内容、工作时长（小时）、操作（编辑/删除）
  - 统计行：总计工作时长（当前查询结果的总和）
  - 分页器

#### Scenario: 日期范围查询（快捷选项）
- **GIVEN** 用户打开日期选择器
- **WHEN** 显示快捷选项
- **THEN** 包含选项：今天、昨天、本周、上周、本月、上月、最近 7 天、最近 30 天、自定义
- **AND** 点击"本周"，自动填充本周一至今天的日期范围

#### Scenario: 日期范围查询（执行查询）
- **GIVEN** 用户选择日期范围 2026-06-20 ~ 2026-06-22
- **WHEN** 点击"查询"按钮
- **THEN** 调用 `getWorkRecords({start_date: '2026-06-20', end_date: '2026-06-22'})`
- **AND** 表格显示该日期范围内的工作记录

#### Scenario: 默认查询范围
- **GIVEN** 页面首次加载
- **WHEN** 未设置日期筛选
- **THEN** 默认查询最近 30 天的工作记录

#### Scenario: 工作事项下拉列表加载
- **GIVEN** 打开新增/编辑对话框
- **WHEN** 渲染工作事项下拉框
- **THEN** 调用 `getWorkItems({status: 1})` 获取仅启用的工作事项
- **AND** 下拉列表显示工作事项名称，值为 ID（字符串）

#### Scenario: 新增工作记录
- **GIVEN** 用户点击"新增"按钮
- **WHEN** 弹出对话框
- **THEN** 显示表单字段：
  - 工作事项（必填，下拉选择）
  - 工作日期（必填，日期选择器，默认今天）
  - 工作内容描述（必填，多行文本框，最大 2000 字符）
  - 工作时长（必填，数字输入框，单位：小时，范围 0.1~24，步进 0.5）

#### Scenario: 新增工作记录（提交成功）
- **GIVEN** 用户填写表单：work_item_id="123", work_date="2026-06-22", description="完成需求分析", duration_hours=2.5
- **WHEN** 点击"确定"
- **THEN** 调用 `createWorkRecord(data)`
- **AND** 成功后关闭对话框，显示成功提示"新增成功"
- **AND** 刷新列表，更新统计行总时长

#### Scenario: 新增工作记录（表单校验）
- **GIVEN** 用户填写表单但工作内容描述为空
- **WHEN** 点击"确定"
- **THEN** 前端表单校验失败，显示"工作内容描述不能为空"

#### Scenario: 新增工作记录（工作时长校验）
- **GIVEN** 用户输入工作时长 25
- **WHEN** 点击"确定"
- **THEN** 前端校验失败，显示"工作时长不能超过 24 小时"

#### Scenario: 新增工作记录（工作时长校验 - 负数）
- **GIVEN** 用户输入工作时长 -1
- **WHEN** 点击"确定"
- **THEN** 前端校验失败，显示"工作时长必须大于 0"

#### Scenario: 编辑工作记录
- **GIVEN** 列表中存在工作记录 ID 789
- **WHEN** 用户点击"编辑"按钮
- **THEN** 弹出对话框，表单字段回显当前值：
  - 工作事项下拉框选中当前 work_item_id
  - 工作日期显示当前 work_date
  - 工作内容描述显示当前 description
  - 工作时长显示当前 duration_hours
- **AND** 用户修改后点击"确定"，调用 `updateWorkRecord(789, data)`

#### Scenario: 编辑回显数据完整性检查
- **GIVEN** 工作记录包含字段：work_item_id, work_date, description, duration_hours
- **WHEN** 打开编辑对话框
- **THEN** 所有字段正确回显，无字段丢失或显示默认值（参考 [[edit-form-data-integrity]]）

#### Scenario: 删除工作记录
- **GIVEN** 列表中存在工作记录 ID 789
- **WHEN** 用户点击"删除"按钮并确认
- **THEN** 调用 `deleteWorkRecord(789)`
- **AND** 成功后显示"删除成功"，刷新列表，更新统计行总时长

#### Scenario: 工作时长统计
- **GIVEN** 查询结果包含 5 条工作记录，时长分别为 2.0, 1.5, 3.0, 0.5, 2.5
- **WHEN** 列表渲染完成
- **THEN** 统计行显示"总计工作时长：9.5 小时"

#### Scenario: 表格排序
- **GIVEN** 列表加载完成
- **WHEN** 显示顺序
- **THEN** 按工作日期降序（最新在前），同一天按创建时间降序

#### Scenario: 工作内容描述截断
- **GIVEN** 工作记录描述超过 100 字符
- **WHEN** 在列表中显示
- **THEN** 截断为 100 字符 + "..."，鼠标悬停显示完整内容（tooltip）

---

### Requirement: 表单验证规则

The system SHALL implement client-side form validation before API calls.

#### Scenario: 工作事项名称验证
- **GIVEN** 工作事项表单
- **WHEN** 用户输入名称
- **THEN** 验证规则：
  - 必填
  - 最大长度 100 字符
  - 不能包含特殊字符（仅字母、数字、中文、空格、常用标点）

#### Scenario: 排序号验证
- **GIVEN** 工作事项表单
- **WHEN** 用户输入排序号
- **THEN** 验证规则：
  - 必填
  - 整数
  - 范围 0~9999

#### Scenario: 工作时长验证
- **GIVEN** 工作记录表单
- **WHEN** 用户输入工作时长
- **THEN** 验证规则：
  - 必填
  - 数字
  - 范围 0.1~24
  - 最多 2 位小数

#### Scenario: 工作内容描述验证
- **GIVEN** 工作记录表单
- **WHEN** 用户输入描述
- **THEN** 验证规则：
  - 必填
  - 最大长度 2000 字符

---

### Requirement: 响应式交互

The system SHALL provide smooth user interactions with proper loading and error states.

#### Scenario: 加载状态
- **GIVEN** 任何 API 调用
- **WHEN** 请求发送
- **THEN** 显示加载指示器（按钮 loading 状态或全局 loading）
- **AND** 禁用提交按钮防止重复提交

#### Scenario: 成功提示
- **GIVEN** API 调用成功
- **WHEN** 响应返回
- **THEN** 显示成功消息提示（Element Plus Message 组件，绿色，3 秒自动关闭）

#### Scenario: 错误提示
- **GIVEN** API 调用失败
- **WHEN** 响应返回错误
- **THEN** 显示错误消息提示（Element Plus Message 组件，红色，显示后端返回的 msg 字段）

#### Scenario: 确认对话框
- **GIVEN** 用户点击删除按钮
- **WHEN** 触发删除操作前
- **THEN** 显示确认对话框（Element Plus MessageBox），包含：
  - 标题："确认删除"
  - 内容："确定要删除该xxx吗？删除后无法恢复。"
  - 按钮："取消"（默认）、"确定"（危险按钮）

---

### Requirement: 数据完整性验证

The system SHALL ensure edit form data integrity as defined in [[edit-form-data-integrity]].

#### Scenario: 8 层全链路检查
- **GIVEN** 新增或修改任何表单字段
- **WHEN** 开发完成
- **THEN** 必须验证：
  1. TypeScript 类型定义包含该字段
  2. API 函数参数类型包含该字段
  3. 创建 API 调用传递该字段
  4. 更新 API 调用传递该字段
  5. 列表/详情 API 响应包含该字段
  6. 表单 data 对象包含该字段
  7. 表单输入控件绑定该字段（v-model）
  8. 编辑回显时 Object.assign 包含该字段

#### Scenario: 编辑回显测试
- **GIVEN** 任何编辑功能
- **WHEN** 执行端到端测试
- **THEN** 测试步骤：
  1. 创建记录，填充所有字段
  2. 列表中查看，验证所有字段正确显示
  3. 点击编辑，验证表单所有字段正确回显
  4. 不修改直接保存，验证数据不丢失

---

### Requirement: 无障碍性

The system SHALL follow basic accessibility guidelines.

#### Scenario: 表单标签
- **GIVEN** 任何表单输入控件
- **WHEN** 渲染表单
- **THEN** 每个输入框有清晰的 label
- **AND** 必填字段标注红色星号

#### Scenario: 键盘导航
- **GIVEN** 对话框打开
- **WHEN** 用户按 Tab 键
- **THEN** 焦点在表单控件间正确移动
- **AND** 按 Enter 提交表单（焦点在输入框时）
- **AND** 按 ESC 关闭对话框

#### Scenario: 屏幕阅读器友好
- **GIVEN** 表单控件
- **WHEN** 使用屏幕阅读器
- **THEN** 控件有正确的 aria-label 或关联 label
- **AND** 错误提示与输入框关联（aria-describedby）

---

### Requirement: 前端构建验证

The system SHALL pass TypeScript type checking and build process.

#### Scenario: TypeScript 编译
- **GIVEN** 所有前端代码
- **WHEN** 执行 `npm run build`
- **THEN** 无 TypeScript 错误
- **AND** 无 ESLint 错误（如配置了 linting）
- **AND** 生成可部署的 dist 目录

#### Scenario: 类型安全检查
- **GIVEN** API 调用代码
- **WHEN** TypeScript 类型检查
- **THEN** 所有参数类型正确
- **AND** 响应类型正确
- **AND** 无 any 类型（除非必要）

---

## Component Structure

### 建议组件文件结构

```
web/pc/src/
├── api/
│   └── pipeline-test.ts          // API 函数 + TypeScript 类型
├── views/
│   └── pipeline-test/
│       ├── WorkItems.vue         // 工作事项维护页面
│       ├── WorkRecords.vue       // 工作内容记录页面
│       ├── components/
│       │   ├── WorkItemForm.vue  // 工作事项表单对话框
│       │   └── WorkRecordForm.vue // 工作记录表单对话框
├── config/
│   └── route.config.ts           // 路由配置（新增模块）
```

### 路由配置示例

```typescript
// @/config/route.config.ts 新增模块
{
  path: '/pipeline-test',
  name: 'PipelineTest',
  meta: {
    title: 'AI-Coding自我提升',
    icon: 'el-icon-document',
    requiresAuth: true
  },
  children: [
    {
      path: 'work-items',
      name: 'WorkItems',
      component: () => import('@/views/pipeline-test/WorkItems.vue'),
      meta: { title: '工作事项维护' }
    },
    {
      path: 'work-records',
      name: 'WorkRecords',
      component: () => import('@/views/pipeline-test/WorkRecords.vue'),
      meta: { title: '工作内容记录' }
    }
  ]
}
```

## UI Component Library

使用 Element Plus 组件：
- `el-table` — 列表展示
- `el-pagination` — 分页
- `el-dialog` — 对话框
- `el-form` / `el-form-item` — 表单
- `el-input` / `el-input-number` / `el-date-picker` / `el-select` — 表单控件
- `el-button` — 按钮
- `el-tag` — 状态标签
- `el-message` — 消息提示
- `el-message-box` — 确认对话框

## Non-Functional Requirements

### Performance
- The system SHALL render lists with < 100 records within 500ms
- The system SHALL debounce search input with 300ms delay

### Usability
- The system SHALL provide clear error messages for validation failures
- The system SHALL auto-focus first input field when dialog opens
- The system SHALL remember last query parameters (optional enhancement)

### Browser Compatibility
- The system SHALL support Chrome 90+, Firefox 88+, Safari 14+, Edge 90+
