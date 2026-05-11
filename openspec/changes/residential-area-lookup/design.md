## Context

当前系统已有正向行政区划筛选（省→市→区县→小区）和小区列表管理功能。`md_residential_area` 表通过 `county_id`、`city_id`、`street_id`、`community_div_id` 关联 `md_administrative_division` 表。行政区划表使用 materialized path（`/省id/市id/区县id/`）存储层级关系，level 字段区分层级（1=省, 2=市, 3=区县, 4=街道, 5=社区）。

现有的 `SearchByName` 方法支持按名称模糊查询，但需要传 countyId/streetId 等筛选条件。需要新增一个不限制区域的搜索方法。

## Goals / Non-Goals

**Goals:**
- 输入小区/村名或代码，返回匹配结果及完整行政区划归属链路
- 前端展示清晰，每个小区一行，显示省→市→区县→街道→社区
- 支持分页，限制单次查询结果量

**Non-Goals:**
- 不做地图展示
- 不做行政区划编辑
- 不改变现有小区管理页面的行为

## Decisions

### 1. 后端 API 设计：新增独立查询接口

新增 `GET /api/masterdata/residential-areas/lookup?keyword=xxx&page=1&page_size=20`

- `keyword` 同时匹配 `name`（LIKE）和 `code`（= 或 LIKE 前缀）
- 复用现有 `MdAdministrativeDivisionModel.FindOne` 通过 county_id/street_id/community_div_id 查询各级区划名称
- 不新增 model 方法，在 logic 层组装查询逻辑

**备选方案**：扩展现有 `GET /api/masterdata/residential-areas` 接口。不采用，因为返回结构不同（需要带行政区划链路），混在一起会增加现有接口复杂度。

### 2. 行政区划链路查询：用 area 的外键逐级查 division 表

`md_residential_area` 已有 `city_id`、`county_id`、`street_id`、`community_div_id` 四个外键字段。

查询策略：
- `city_id` → `FindOne` 得到城市名
- `county_id` → `FindOne` 得到区县名
- `street_id` → `FindOne` 得到街道名（可能为空）
- `community_div_id` → `FindOne` 得到社区名（可能为空）
- 省份：从区县的 `parent_id` 查到城市，再从城市的 `parent_id` 查到省份（或直接用 `FindOne` 查区县记录的 `path` 解析）

用 `county_id` 对应的 division 记录的 `path` 字段解析最可靠（`/省id/市id/区县id/`），一次查询即可获取省市。

### 3. 前端页面：新增独立路由页面

路径：`masterdata/residential-areas/lookup`，菜单名为"小区查询"

- 顶部搜索框 + 查询按钮
- 结果列表用 el-table，每行显示：小区名、代码、省、市、区县、街道、社区
- 分页组件

### 4. Model 层：新增 `SearchByKeyword` 方法

在 `mdResidentialAreaModel.go` 新增方法，支持不限制区域的名称+代码联合搜索，因为现有 `SearchByName` 强制要求传 countyId 相关参数。

```go
SearchByKeyword(ctx, keyword string, submissionStatus *int32, page, pageSize int64, excludeStatuses ...int32) ([]*MdResidentialArea, error)
CountByKeyword(ctx, keyword string, submissionStatus *int32, excludeStatuses ...int32) (int64, error)
```

SQL: `WHERE (name LIKE ? OR code LIKE ?) AND delete_time IS NULL`

## Risks / Trade-offs

- [数据量大时 LIKE 查询慢] → 限制 `keyword` 最少2个字符，分页 size 上限50，考虑后续加索引
- [部分小区的 street_id/community_div_id 为空] → 前端显示"-"，不影响其他层级展示
- [高德同步的小区没有 street_id/community_div_id] → 只展示到区县级别，后续可通过行政区划代码匹配补充
