## Context

住宅小区列表页已实现省份→城市→区县三级联动筛选（update-residential-areas-1 完成）。后端 `GET /api/masterdata/divisions` 支持 `level` 和 `parent_id` 参数，`GET /api/masterdata/residential-areas` 支持 `community_div_id` 参数。无需后端改动。

## Goals / Non-Goals

**Goals:**
- 在区县下拉后增加街道/乡镇（level 4）和社区（level 5）两个 `el-select`
- 选区县后加载街道，选街道后加载社区，逐级联动
- 搜索时优先使用 `community_div_id` 筛选，其次 `county_id`

**Non-Goals:**
- 不修改后端 API
- 不修改新建/编辑表单

## Decisions

与 update-residential-areas-1 设计完全一致：三个独立 el-select 按需加载，上级切换清空下级。

## Risks / Trade-offs

- 五级下拉占水平空间较多 → 使用紧凑宽度（130px）保持一行排列
