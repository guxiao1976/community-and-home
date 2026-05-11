## Context

现有住宅小区管理页面（`List.vue`）提供完整的 CRUD + 审核流程，筛选条件包括省/市/区县/街道/社区五级级联、关键词、小区类型、提交状态。后端 `GetResidentialAreasLogic` 支持这些筛选并返回 `ResidentialArea` 列表，但响应中只有 city_id、county_id 等 ID，没有行政区划名称。`MdAdministrativeDivisionModel.FindOne` 可按 ID 查行政区划名称。

## Goals / Non-Goals

**Goals:**
- 新增独立的只读查询 API，复用现有筛选逻辑，额外返回各级行政区划名称
- 前端新建查询页面，多 tab 结构，第 1 个 tab 为"小区&村"
- 名称和 ID 合并展示，如"东营市（283989）"

**Non-Goals:**
- 不修改现有住宅小区管理页面的行为
- 不做分页之外的高级查询（如导出）
- 后续其他 tab 页（如行政区划查询）本期不实现

## Decisions

### 1. 新增独立 API 而非复用现有接口

新增 `GET /api/masterdata/query/residential-areas`，请求参数与 `GetResidentialAreasReq` 基本一致（去掉 submission_status），响应新增 city_name、county_name、street_name、community_name 字段。

不修改现有 `GetResidentialAreas` 接口，因为返回结构不同，混在一起增加复杂度。

### 2. 行政区划名称解析策略

在 logic 层查询小区列表后，遍历结果：
- 用 `city_id` → `FindOne` 获取城市名称
- 用 `county_id` → `FindOne` 获取区县名称
- 用 `street_id` → `FindOne` 获取街道名称（可能为空）
- 用 `community_div_id` → `FindOne` 获取社区名称（可能为空）

为避免 N+1 查询，对同一 city_id/county_id 做内存缓存（map[int64]string），一次查询页面 20 条最多额外查 4×20=80 次，可接受。后续可优化为批量查询。

### 3. 前端 tab 结构

使用 `el-tabs` 组件，当前只有"小区&村"一个 tab，后续扩展其他 tab 直接添加 `el-tab-pane`。

查询条件区域复用现有五级级联选择器逻辑，搜索按钮触发查询。

### 4. 默认只查已批准数据

查询模块面向普通用户查看，默认只返回 submission_status=2（已批准）的数据，不暴露待提交/已拒绝等状态。

## Risks / Trade-offs

- [N+1 查询] → 单页 20 条最多 80 次额外 FindOne，有 Redis 缓存，性能可接受
- [street_id/community_div_id 为空的高德数据] → 对应列显示"-"
