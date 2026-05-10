## Why

当前住宅小区数据需要手动逐条录入，效率低且数据不完整。高德地图 POI 数据库包含全国住宅小区的名称、地址、经纬度等信息，可以通过其公开 API 按行政区划批量获取，大幅提升小区基础数据的采集效率。

## What Changes

- `md_residential_area` 表新增 `longitude`、`latitude` 字段（DECIMAL(10,7)），存储小区经纬度坐标
- `md_residential_area` 表新增 `data_source` 字段（TINYINT），标识数据来源：0=人工维护，1=高德接口
- 新增后端同步引擎：接收选定的街道 ID 列表，逐街道调用高德 District API 获取边界多边形，再调用 Place Polygon API 搜索住宅小区 POI（types=120300），去重后写入数据库
- 新增两个后端 API 端点：触发同步（POST）、查询进度（GET）
- 新增前端"高德地图同步"页面：省→市→区县→街道四级联动选择（街道多选），触发同步并轮询进度
- 同步数据直接为已批准状态（submission_status=2），编码格式为街道行政区划编码+4位顺序号，data_source=1
- 每个街道查询间隔随机 1-3 秒

## Capabilities

### New Capabilities
- `amap-sync`: 高德地图小区数据同步功能，包含后端同步引擎、API 端点和前端同步页面

### Modified Capabilities
- `residential-area`: 数据模型新增经纬度字段（longitude, latitude），.api 类型定义同步更新

## Impact

- **数据库**: `md_residential_area` 表新增 3 列（longitude, latitude, data_source），需执行 ALTER TABLE
- **后端**: masterdata 服务新增 amap_sync handler/logic，ServiceContext 注册 SyncEngine，config 增加 AMapKey
- **前端**: 新增 `/masterdata/amap-sync` 路由和菜单项
- **外部依赖**: 依赖高德 REST API（District v3 + Place Polygon v5），需配置 AMap API Key
- **AMap Key**: `86a40b61232944538a345529ceaabcbe`，存储在后端 YAML 配置中
