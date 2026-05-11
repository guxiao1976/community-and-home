## ADDED Requirements

### Requirement: 实时统计 API 端点
系统 SHALL 提供 `GET /api/masterdata/statistics/division-counts/realtime` 端点，返回实时行政区划小区/村统计数据。

#### Scenario: 无 parent_id 查询省级实时统计
- **WHEN** 客户端请求 `GET /api/masterdata/statistics/division-counts/realtime`（不传 parent_id）
- **THEN** 系统返回所有省级（level=1）已审批行政区划的实时统计数据，每条包含 id、name、level、community_count、village_count、total_count，按 total_count 降序排列

#### Scenario: 传入 parent_id 查询子级实时统计
- **WHEN** 客户端请求 `GET /api/masterdata/statistics/division-counts/realtime?parent_id=X`
- **THEN** 系统返回 parent_id=X 下所有子级已审批行政区划的实时统计数据，格式同上

### Requirement: 实时查询仅统计已审批且未删除的记录
实时统计 SHALL 只计算 `submission_status=2` 且 `delete_time IS NULL` 的行政区划和住宅小区/村数据。

#### Scenario: 未审批小区不计入实时统计
- **WHEN** 存在 submission_status != 2 的住宅小区
- **THEN** 实时统计结果中 SHALL NOT 包含这些小区的计数

#### Scenario: 已软删除小区不计入实时统计
- **WHEN** 存在 delete_time 非 NULL 的住宅小区
- **THEN** 实时统计结果中 SHALL NOT 包含这些小区的计数

### Requirement: 实时统计按层级聚合
实时统计 SHALL 按行政区划层级正确聚合：区县按 county_id 聚合，街道按 street_id 聚合，市和省通过子查询逐级聚合。

#### Scenario: 区县级实时统计
- **WHEN** 传入 parent_id 为市级区划 ID
- **THEN** 返回该市下辖所有区县的实时统计，community_count 和 village_count 通过 md_residential_area.county_id 直接聚合计算

#### Scenario: 街道级实时统计
- **WHEN** 传入 parent_id 为区县级区划 ID
- **THEN** 返回该区县下辖所有街道的实时统计，community_count 和 village_count 通过 md_residential_area.street_id 直接聚合计算

#### Scenario: 市级实时统计
- **WHEN** 传入 parent_id 为省级区划 ID
- **THEN** 返回该省下辖所有市的实时统计，community_count/village_count/total_count 通过区县级统计按 parent_id 聚合计算

### Requirement: 实时统计响应格式与预计算接口一致
实时统计端点 SHALL 复用 `DivisionCountsReq` 和 `DivisionCountsResp` 类型，返回 `DivisionCountItem` 数组。

#### Scenario: 响应格式验证
- **WHEN** 任意实时统计请求成功
- **THEN** 响应 body 格式为 `{ code: 0, message: "success", data: { list: [{ id, name, level, community_count, village_count, total_count }] } }`

#### Scenario: 无数据时返回空列表
- **WHEN** 查询条件无匹配的已审批区划
- **THEN** 响应 data.list 为空数组 `[]`，不返回错误

### Requirement: 实时统计不依赖预计算表
实时统计 SHALL 直接查询 `md_residential_area` 和 `md_administrative_division` 表，不读取 `md_division_statistics` 表。

#### Scenario: 预计算表为空时实时统计正常工作
- **WHEN** `md_division_statistics` 表中无数据（如定时任务尚未执行）
- **THEN** 实时统计端点仍能返回正确的实时数据
