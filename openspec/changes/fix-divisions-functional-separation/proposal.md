## Why

当前行政区划页面将省-市-区县-街道/乡镇-社区/村五个层级放在同一个树形表格中管理，但实际上省市区县的变更非常少，而街道/乡镇和社区/村的变更频繁。将两者混在一起既不方便高频操作，也增加了误操作风险。需要按业务使用频率将页面拆分为两个独立模块。

## What Changes

- **新增"基层组织"菜单和页面**：专门管理街道/乡镇（level 4）和社区/村委会（level 5）
  - 页面顶部放置"省-市-区县"三级级联选择器，用于定位到目标区县
  - 【搜索】按钮：根据选中的区县查询其下属的街道/乡镇数据
  - 【新增】按钮：在当前选中的区县下新增街道/乡镇
  - 数据展示方式与当前行政区划页面相同（树形表格，街道→社区）
  - 每行提供编辑、提交、添加下级（社区级）、删除按钮
- **修改"行政区划"模块**：只管理到区县级（level 1-3）
  - 移除街道级和社区级的显示和操作
  - 区县行不再显示"添加下级"按钮（或限制为不可用）
  - 现有的新增区划功能保留，但级别选项只到区县级
- **新增路由**：`/masterdata/grassroots` 指向基层组织页面
- **修改侧边栏菜单**：在"主数据管理"下新增"基层组织"菜单项

## Capabilities

### New Capabilities
- `grassroots-management`: 基层组织管理模块，包含街道/乡镇和社区/村委会的 CRUD、提交审核、区县级联定位

### Modified Capabilities
- `division-management`: 行政区划模块限制为只管理省-市-区县（level 1-3），移除街道级和社区级的创建、展示和操作

## Impact

- **前端页面**：新增 `views/grassroots/Index.vue`，修改 `views/division/Index.vue`
- **路由**：`src/router/index.ts` 新增路由
- **侧边栏**：`src/components/layout/AppSidebar.vue` 新增菜单项
- **API 层**：无变更，复用现有 `getAdministrativeDivisions` 等 API（通过 `parent_id` 和 `level` 参数过滤）
- **Store**：复用现有 `stores/division.ts`，无需修改
- **类型定义**：无变更，复用现有 `DivisionLevel` 枚举
- **审核中心**：无影响，审核流程不变
