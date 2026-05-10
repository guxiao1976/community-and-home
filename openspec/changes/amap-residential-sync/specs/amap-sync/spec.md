## ADDED Requirements

### Requirement: 后端代理高德 District API 获取街道边界
系统 SHALL 通过后端调用高德 District API v3，使用 `{区县名}{街道名}` 作为关键词、`extensions=all` 参数，获取街道边界多边形 polyline 坐标串。

#### Scenario: 成功获取街道边界
- **WHEN** 后端收到街道同步请求，查询行政区划表获得区县名和街道名
- **THEN** 调用 `GET https://restapi.amap.com/v3/config/district?keywords={区县名}{街道名}&subdistrict=0&extensions=all&key={AMapKey}`，从响应 `districts[0].polyline` 中提取边界坐标

#### Scenario: 街道在高德中未找到
- **WHEN** 高德 District API 返回空结果或 polyline 为空
- **THEN** 跳过该街道，TotalFailed 计数加 1，继续处理下一个街道

### Requirement: 后端代理高德 Place Polygon API 搜索住宅小区
系统 SHALL 使用上一步获取的 polyline 坐标作为多边形参数，调用高德 Place Polygon API v5，`types=120300` 筛选住宅小区 POI。

#### Scenario: 成功搜索到住宅小区
- **WHEN** 高德 District API 返回有效 polyline
- **THEN** 调用 `GET https://restapi.amap.com/v5/place/polygon?polygon={polyline}&types=120300&region={区县名}&city_limit=true&page_size=25&page_num=1&key={AMapKey}`，获取 POI 列表

#### Scenario: polyline 过长时抽稀
- **WHEN** polyline 坐标点数量超过 200 个
- **THEN** 对坐标点进行等间距抽稀至 200 个以内，再作为 polygon 参数传入

### Requirement: 按街道顺序同步并控制请求间隔
系统 SHALL 按街道 ID 列表顺序逐个处理，每个街道处理完成后随机等待 1-3 秒再处理下一个。

#### Scenario: 多个街道顺序处理
- **WHEN** 用户选择了 5 个街道发起同步
- **THEN** 系统依次处理第 1~5 个街道，每个之间间隔 1~3 秒随机延迟

#### Scenario: 最后一个街道不等待
- **WHEN** 处理到最后一个街道
- **THEN** 完成后立即标记任务完成，不等待额外延迟

### Requirement: 小区去重后写入数据库
系统 SHALL 对每个 POI 按 `name + county_id` 去重，仅插入不存在的小区记录。

#### Scenario: 小区名称在区县内已存在
- **WHEN** POI 名称通过 `FindByNameAndCountyId` 查到已有记录
- **THEN** 跳过该 POI，TotalSkipped 计数加 1

#### Scenario: 小区名称在区县内不存在
- **WHEN** POI 名称在 county_id 下无匹配记录
- **THEN** 生成编码（streetCode + 4位序号），插入 `md_residential_area` 表，TotalSynced 计数加 1

### Requirement: 同步数据编码规则
系统 SHALL 为每个同步的小区生成编码，格式为 `{街道行政区划编码}{4位顺序号}`，每个街道从 0001 开始递增。

#### Scenario: 编码唯一性
- **WHEN** 生成编码 `streetCode + 0001`
- **THEN** 通过 `FindByCode` 验证唯一性，若已存在则递增重试，直到找到唯一编码

### Requirement: 同步数据状态和字段映射
系统 SHALL 将同步的小区写入数据库时设置 `submission_status=2`（已批准）、`community_type=1`（住宅小区）、`submitter_id=0`（系统导入）、`data_source=1`（高德接口），并从 POI 的 `location` 字段解析经纬度。

#### Scenario: POI location 解析
- **WHEN** 高德返回 POI 的 location 为 `"118.495248,37.463162"`
- **THEN** longitude 存入 `118.4952480`，latitude 存入 `37.4631620`，data_source 存入 `1`

### Requirement: 触发同步 API 端点
系统 SHALL 提供 `POST /api/masterdata/amap-sync/sync` 端点，接收 `county_id` 和 `street_ids`，启动后台同步任务并返回 `task_id`。

#### Scenario: 参数校验
- **WHEN** 请求中 `county_id` 为 0 或 `street_ids` 为空
- **THEN** 返回参数错误

#### Scenario: 成功启动同步
- **WHEN** 参数合法
- **THEN** 启动 goroutine 执行同步，立即返回 `{ task_id: "..." }`

### Requirement: 同步进度查询 API 端点
系统 SHALL 提供 `GET /api/masterdata/amap-sync/progress?task_id=xxx` 端点，返回当前同步任务的进度信息。

#### Scenario: 任务运行中
- **WHEN** 前端查询正在运行的同步任务
- **THEN** 返回 `{ status: "running", total_streets, current_street_index, current_street_name, total_synced, total_skipped, total_failed }`

#### Scenario: 任务已完成
- **WHEN** 前端查询已完成的同步任务
- **THEN** 返回 `{ status: "completed", ... }` 包含最终统计数据

#### Scenario: 任务不存在
- **WHEN** 查询的 task_id 不存在
- **THEN** 返回错误提示

### Requirement: 前端高德地图同步页面
系统 SHALL 提供同步页面，包含省→市→区县→街道四级联动选择器（街道支持多选），同步按钮和进度展示。

#### Scenario: 级联选择
- **WHEN** 用户选择省份
- **THEN** 加载该省下的城市列表；选择城市后加载区县列表；选择区县后加载街道列表

#### Scenario: 发起同步
- **WHEN** 用户选择至少一个街道并点击同步
- **THEN** 弹出确认对话框，确认后调用同步 API，展示进度条和统计信息

#### Scenario: 轮询进度
- **WHEN** 同步任务启动后
- **THEN** 前端每 2 秒轮询进度接口，展示"正在处理第 N/M 个街道：街道名"和已同步/跳过/失败数量

#### Scenario: 同步完成
- **WHEN** 轮询到 status 为 completed
- **THEN** 停止轮询，弹出成功提示，展示最终统计汇总

### Requirement: AMap Key 配置
系统 SHALL 将 AMap API Key 存储在后端 YAML 配置文件中，不暴露给前端。

#### Scenario: 配置读取
- **WHEN** masterdata 服务启动
- **THEN** 从 `masterdata-api.yaml` 读取 `AMapKey` 字段，注入 SyncEngine
