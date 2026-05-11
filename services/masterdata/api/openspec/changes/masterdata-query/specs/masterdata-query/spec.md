## ADDED Requirements

### Requirement: 小区&村查询 API
系统 SHALL 提供 `GET /api/masterdata/query/residential-areas` 接口，支持按行政区划和关键词查询已批准的小区/村数据，返回包含行政区划名称的结果列表。

请求参数：
- `city_id`（可选，int64）
- `county_id`（可选，int64）
- `street_id`（可选，int64）
- `community_div_id`（可选，int64）
- `keyword`（可选，string）：小区名称关键词
- `community_type`（可选，int32）：1=住宅小区，2=村庄，3=混合型
- `page`（可选，int32，默认1）
- `page_size`（可选，int32，默认20，最大50）

响应字段在现有 ResidentialArea 基础上额外包含：
- `city_name`（string）：城市名称
- `county_name`（string）：区县名称
- `street_name`（string）：街道名称，可为空
- `community_name`（string）：社区名称，可为空

默认只返回 submission_status=2（已批准）的记录。

#### Scenario: 按城市查询小区
- **WHEN** 用户传入 city_id=283989
- **THEN** 系统返回该城市下所有已批准小区，每条记录包含 city_name="东营市"、county_name 等行政区划名称

#### Scenario: 按关键词搜索
- **WHEN** 用户传入 keyword="阳光小区"
- **THEN** 系统返回 name LIKE '%阳光小区%' 的已批准小区，包含行政区划名称

#### Scenario: 按小区类型筛选
- **WHEN** 用户传入 community_type=2
- **THEN** 系统只返回村庄类型的已批准小区

#### Scenario: 行政区划名称缺失
- **WHEN** 查询到的小区 street_id 为空
- **THEN** 响应中 street_name 为空字符串

### Requirement: 主数据查询页面
系统 SHALL 提供"数据查询"页面，使用 tab 页签组织不同查询类型，默认显示第 1 个 tab"小区&村"。

#### Scenario: 小区&村 tab 查询条件
- **WHEN** 用户打开"小区&村" tab
- **THEN** 页面上部显示查询条件：省/市/区县/街道/社区五级级联下拉框、小区名称输入框、小区类型下拉框、搜索按钮、重置按钮

#### Scenario: 结果展示
- **WHEN** 用户点击搜索
- **THEN** 页面下部展示表格，列依次为：ID、小区名称、小区编码、城市名称（ID）、县区名称（ID）、街道/乡镇名称（ID）、社区名称（ID）、地址、小区类型、操作（详情）

#### Scenario: 名称和 ID 合并展示
- **WHEN** 表格中显示"城市名称（ID）"列
- **THEN** 格式为"东营市（283989）"，ID 缺失时只显示名称

#### Scenario: 只读操作
- **WHEN** 用户在结果表格中查看某一行
- **THEN** 操作列只有"详情"按钮，点击后跳转到现有小区详情页面

#### Scenario: 空字段展示
- **WHEN** 某小区的街道或社区信息为空
- **THEN** 对应列显示"-"
