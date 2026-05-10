## MODIFIED Requirements

### Requirement: 住宅小区数据模型
`md_residential_area` 表 SHALL 包含 `longitude`（DECIMAL(10,7)）、`latitude`（DECIMAL(10,7)）和 `data_source`（TINYINT DEFAULT 0）字段。`data_source` 标识数据来源：0=人工维护，1=高德接口。

#### Scenario: 新增字段存储经纬度
- **WHEN** 通过高德同步或其他方式写入小区数据时提供经纬度
- **THEN** longitude 和 latitude 字段被正确存储，精度为小数点后 7 位

#### Scenario: 字段可为空
- **WHEN** 手动创建小区数据未提供经纬度
- **THEN** longitude 和 latitude 字段为 NULL，不影响其他功能

#### Scenario: 高德同步数据标记来源
- **WHEN** 通过高德地图同步写入小区数据
- **THEN** data_source 字段值为 1

#### Scenario: 人工创建数据默认来源
- **WHEN** 手动创建小区数据
- **THEN** data_source 字段值为 0（默认值）

### Requirement: 住宅小区 API 类型定义
`.api` 文件中的 `ResidentialArea` 类型 SHALL 包含 `longitude`、`latitude` 和 `data_source` 字段。

#### Scenario: API 返回经纬度和数据来源
- **WHEN** 查询住宅小区详情或列表
- **THEN** 响应中包含 longitude、latitude 字段（可能为 null）和 data_source 字段（0 或 1）
