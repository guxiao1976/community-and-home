## Context

masterdata 服务已有住宅小区的 CRUD API 和审批流程，但数据全部依赖手动录入。现有 `md_residential_area` 表包含 county_id、street_id、code、name、address 等字段，缺少经纬度。前端有省→市→区→街道的级联查询模式（`getAdministrativeDivisions({ parent_id, level })`），可复用于同步页面的选择器。

高德 REST API 提供两个关键接口：
- **District API v3**: 按关键词查询行政区划，`extensions=all` 返回边界多边形 polyline
- **Place Polygon API v5**: 按多边形区域搜索 POI，`types=120300` 筛选住宅小区

## Goals / Non-Goals

**Goals:**
- 通过高德 API 按街道批量同步住宅小区数据到 `md_residential_area`
- 支持省→市→区县→街道多选后一键同步
- 同步过程有进度反馈
- 小区经纬度坐标持久化存储

**Non-Goals:**
- 不实现 AMap API 的分页遍历（初始版本只取 page_num=1, page_size=25）
- 不实现同步任务的持久化存储（服务重启后进度丢失）
- 不实现同步结果的撤销/回滚
- 不修改现有住宅小区的 CRUD 和审批流程

## Decisions

### 1. AMap API 调用层：后端代理

**选择**: 所有 AMap API 调用由 Go 后端发起，前端只调用内部 API。

**理由**: AMap Key 不应暴露给前端浏览器。后端代理还可统一处理速率限制、错误重试和日志记录。

**替代方案**: 前端直接调用 — 简单但 Key 暴露，且受浏览器 CORS 限制。

### 2. 同步执行模型：goroutine + 内存进度

**选择**: 后端启动 goroutine 顺序执行街道同步，进度存储在内存 `map[string]*SyncProgress` 中，前端通过轮询 GET 接口获取进度。

**理由**: 同步是长时间运行的操作（N 个街道 × 1-3秒间隔），不能用同步 HTTP 请求。goroutine 轻量且简单。进度用内存 map 即可，无需持久化——服务重启后重新同步即可。

**替代方案**: WebSocket 推送 — 实现复杂度高，轮询 2 秒间隔足够。消息队列 — 过度设计。

### 3. 编码规则：街道编码 + 4位序号

**选择**: 小区 code = `{streetDivisionCode}{4位序号}`，每个街道从 0001 开始递增，`FindByCode` 保证唯一。

**理由**: 用户明确要求此格式。streetDivisionCode 已存在于 `md_administrative_division.code`（如 `110105001`），拼接后 13 位，在 VARCHAR(100) 范围内。

**与现有编码的关系**: 现有 `createResidentialAreaLogic` 使用 `countyCode + 4位序号`，两者编码前缀不同但长度兼容，不会冲突。

### 4. 去重策略：按名称+区县去重

**选择**: 对每个 POI 调用 `FindByNameAndCountyId(poi.Name, countyId)`，已存在则跳过。

**理由**: 同一个小区名称在同一区县内基本不会重复。比按 code 去重更合理，因为 code 是本次新规则生成的。

### 5. 数据模型变更：新增 longitude/latitude/data_source

**选择**: `ALTER TABLE md_residential_area ADD COLUMN longitude DECIMAL(10,7), ADD COLUMN latitude DECIMAL(10,7), ADD COLUMN data_source TINYINT NOT NULL DEFAULT 0`。

**理由**: 高德 POI 返回 `location` 字段格式为 `"lng,lat"`，需要拆分存储。DECIMAL(10,7) 精度约 1cm，满足地址定位需求。`data_source` 字段用于区分数据来源：0=人工维护（默认），1=高德接口同步。

### 6. 同步数据状态：直接已批准

**选择**: `submission_status = 2`（已批准），`submission_type = NULL`，`data_source = 1`。

**理由**: 数据来源于高德地图，视为权威数据源，无需审批流程。`submitter_id = 0` 表示系统导入，`data_source = 1` 标识高德接口来源。

## Risks / Trade-offs

**[URL 长度限制]** → 高德 District API 返回的 polyline 可能非常长（复杂边界数百个坐标点），作为 Place Polygon API 的 URL 参数可能超长（>8000字符）。**缓解**: 对 polyline 坐标点进行抽稀，限制在 200 个点以内。

**[page_size=25 遗漏]** → 每个街道最多获取 25 个小区，大面积街道可能遗漏。**缓解**: 初始版本接受此限制，后续可扩展分页遍历。

**[AMap API 限流/失败]** → 高德 API 有 QPS 限制，网络也可能失败。**缓解**: 每个街道间隔 1-3 秒已满足限流要求；单个街道失败时跳过并记录 TotalFailed，不中断整个任务。

**[并发编码冲突]** → 两个同步任务同时运行可能生成相同 code。**缓解**: `FindByCode` 唯一性检查会兜底，且实际使用中不太会同时发起多个同步任务。

**[服务重启丢失进度]** → goroutine 和内存 map 在重启后丢失。**缓解**: 同步是幂等的（去重），重启后重新同步即可，不会产生重复数据。
