## Why

住宅小区列表页的区县筛选目前直接从区县（level 3）开始选择，缺少省份（level 1）和城市（level 2）两个层级，用户无法直观定位目标区域。需要增加省份、城市下拉列表，实现省→市→区县三级联动，并要求必须选择到城市才能触发搜索，提升数据筛选的准确性和用户体验。

## What Changes

- **前端筛选区域改造**：在小区列表页查询条件区域，将现有的单级区县下拉改为"省份 → 城市 → 区县"三级联动下拉列表
- **搜索限制**：必须选择到城市级别才允许执行搜索；未选择城市时，搜索按钮禁用或点击后提示
- **后端无需修改**：`GET /api/masterdata/divisions` 已支持 `level` 和 `parent_id` 筛选参数，可直接按层级加载；`GET /api/masterdata/residential-areas` 已支持 `county_id` 参数筛选

## Capabilities

### New Capabilities

(无新增能力)

### Modified Capabilities

- `residential-area-list-filter`: 住宅小区列表页的查询筛选区域，从单级区县选择改为省→市→区县三级联动，并增加城市必选校验

## Impact

- **前端**: `web/pc/src/views/residential-areas/List.vue` — 筛选区域 UI 和交互逻辑
- **前端 API**: `web/pc/src/api/masterdata.ts` — `getAdministrativeDivisions` 已满足需求，无需修改
- **后端**: 无需修改，现有 divisions API 和 residential-areas API 已支持所需功能
