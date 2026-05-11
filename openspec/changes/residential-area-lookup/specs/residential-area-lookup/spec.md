## ADDED Requirements

### Requirement: 小区反向查询 API
系统 SHALL 提供 `GET /api/masterdata/residential-areas/lookup` 接口，支持按关键词查询住宅小区并返回行政区划归属链路。

请求参数：
- `keyword`（必填，string）：小区名称或代码，最少2个字符
- `page`（可选，int，默认1）
- `page_size`（可选，int，默认20，最大50）

#### Scenario: 按名称模糊查询
- **WHEN** 用户传入 keyword="阳光小区"
- **THEN** 系统返回所有 name LIKE '%阳光小区%' 的小区，每个结果包含省、市、区县、街道、社区名称

#### Scenario: 按代码查询
- **WHEN** 用户传入 keyword="370502"
- **THEN** 系统返回所有 code LIKE '%370502%' 的小区，每个结果包含省、市、区县、街道、社区名称

#### Scenario: 关键词过短
- **WHEN** 用户传入 keyword="张"（少于2个字符）
- **THEN** 系统返回错误提示"关键词至少2个字符"

#### Scenario: 无匹配结果
- **WHEN** 用户传入 keyword="不存在的xyz小区"
- **THEN** 系统返回空列表，total=0

### Requirement: 行政区划归属链路返回
查询结果中每个小区 SHALL 包含其完整的行政区划归属链路：省份名称、城市名称、区县名称、街道名称（可为空）、社区名称（可为空）。

#### Scenario: 完整归属链路
- **WHEN** 查询到一条小区记录，其 county_id、street_id、community_div_id 均有值
- **THEN** 返回结果包含 province_name、city_name、county_name、street_name、community_name 五个字段

#### Scenario: 部分归属缺失
- **WHEN** 查询到一条小区记录，其 street_id 和 community_div_id 为空（如高德同步数据）
- **THEN** 返回结果中 street_name 和 community_name 为空字符串，其他层级正常显示

### Requirement: 小区查询前端页面
系统 SHALL 提供"小区查询"页面，包含搜索输入框和结果列表。

#### Scenario: 搜索并展示结果
- **WHEN** 用户在搜索框输入关键词并点击查询
- **THEN** 页面展示匹配的小区列表，每行显示：小区名称、代码、省份、城市、区县、街道、社区

#### Scenario: 分页浏览
- **WHEN** 查询结果超过一页（默认20条）
- **THEN** 页面底部显示分页组件，用户可翻页浏览

#### Scenario: 空字段展示
- **WHEN** 某小区的街道或社区信息为空
- **THEN** 对应列显示"-"
