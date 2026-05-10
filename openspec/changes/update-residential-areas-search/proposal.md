## Why

住宅小区列表页搜索条件已有省份→城市→区县三级联动，但小区实际归属于街道/乡镇（level 4）和社区（level 5），仅按区县筛选粒度不够细。需要增加街道和社区两个下拉条件，实现完整的五级区划筛选。

## What Changes

- **前端筛选区域扩展**：在现有省份→城市→区县三级下拉后，增加街道/乡镇（level 4）和社区（level 5）两个 `el-select`，形成五级联动
- **联动逻辑**：选区县后加载街道，选街道后加载社区，切换上级时清空下级
- **后端无需修改**：`GET /api/masterdata/divisions` 已支持 `level`/`parent_id` 筛选，`GET /api/masterdata/residential-areas` 已支持 `community_div_id` 参数

## Capabilities

### New Capabilities

(无新增能力)

### Modified Capabilities

- `residential-area-list-filter`: 在现有省→市→区县三级联动基础上，扩展为省→市→区县→街道→社区五级联动

## Impact

- **前端**: `web/pc/src/views/residential-areas/List.vue` — 新增街道和社区两个 el-select 及联动逻辑
- **后端**: 无需修改
