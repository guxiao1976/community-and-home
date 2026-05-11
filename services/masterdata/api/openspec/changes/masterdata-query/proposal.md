## Why

需要一个只读的主数据查询模块，方便用户按行政区划和关键词快速检索小区/村数据，并直观看到完整的行政区划归属链路（城市、区县、街道、社区名称+ID）。现有住宅小区管理页面功能复杂（增删改查、提交审核），不适合作为纯查询入口。

## What Changes

- 新增后端 API `GET /api/masterdata/query/residential-areas`，复用现有筛选条件，返回结果中包含各级行政区划名称和 ID
- 新增前端"数据查询"页面，包含多个 tab 页，第 1 个 tab 为"小区&村"
- 第 1 个 tab 查询条件与现有住宅小区列表一致（省/市/区县/街道/社区级联、关键词、小区类型）
- 结果展示：ID、小区名称、小区编码、城市名称（ID）、县区名称（ID）、街道/乡镇名称（ID）、社区名称（ID）、地址、小区类型、操作（仅详情）
- 页面纯只读，无增删改操作

## Capabilities

### New Capabilities
- `masterdata-query`: 主数据查询模块 — 只读查询页面，多 tab 结构，第 1 个 tab 为小区&村查询

### Modified Capabilities

## Impact

- 后端：新增 API 路由、handler、logic，复用现有 model 方法
- 前端：新增页面和路由
- 无数据库 schema 变更
