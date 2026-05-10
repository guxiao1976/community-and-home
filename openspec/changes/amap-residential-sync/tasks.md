## 1. 数据模型变更

- [x] 1.1 执行 ALTER TABLE 为 `md_residential_area` 新增 `longitude DECIMAL(10,7)`、`latitude DECIMAL(10,7)` 和 `data_source TINYINT NOT NULL DEFAULT 0` 字段
- [x] 1.2 更新 `scripts/sql/masterdata_schema.sql` 中 `md_residential_area` 表定义，在 `address` 后添加 longitude、latitude、data_source 字段
- [x] 1.3 更新 `.api` 文件中 `ResidentialArea` 类型，新增 `Longitude`、`Latitude` 和 `DataSource` 字段
- [x] 1.4 运行 goctl 重新生成 `types.go`，确认 Go struct 中包含新字段
- [x] 1.5 更新 `mdResidentialAreaModel_gen.go` 中 `MdResidentialArea` struct 手动添加 `Longitude sql.NullFloat64`、`Latitude sql.NullFloat64` 和 `DataSource int64` 字段

## 2. 后端配置

- [x] 2.1 在 `services/masterdata/api/etc/masterdata-api.yaml` 添加 `AMapKey` 配置项
- [x] 2.2 在 `services/masterdata/api/internal/config/config.go` 的 Config struct 中添加 `AMapKey string` 字段

## 3. 后端同步引擎

- [x] 3.1 创建 `services/masterdata/api/internal/sync/sync_engine.go`，定义 `SyncEngine`、`SyncProgress` 结构体和 `NewSyncEngine` 构造函数
- [x] 3.2 实现 `StartSync` 方法：生成 task ID，创建 SyncProgress，启动 goroutine，返回 task ID
- [x] 3.3 实现 `GetProgress` 方法：从 tasks map 返回进度快照
- [x] 3.4 实现 `runSync` 主循环：遍历街道 → 调用高德 District API → 调用 Place Polygon API → 去重 → 插入数据库
- [x] 3.5 实现 AMap HTTP 请求逻辑：District API 调用和 polyline 提取，polyline 超过 200 点时抽稀
- [x] 3.6 实现 Place Polygon API 调用和 POI 列表解析
- [x] 3.7 实现去重逻辑：`FindByNameAndCountyId` 检查，已存在则 TotalSkipped++
- [x] 3.8 实现编码生成：`streetCode + 4位序号`，`FindByCode` 唯一性校验
- [x] 3.9 实现 POI location 解析：`"lng,lat"` 拆分为 longitude/latitude
- [x] 3.10 实现数据库插入：构建 MdResidentialArea struct（submission_status=2, community_type=1, submitter_id=0, data_source=1），调用 Insert
- [x] 3.11 实现随机延迟 1-3 秒（最后一个街道不等待）
- [x] 3.12 实现错误处理：单个街道失败时 TotalFailed++ 并 continue，不中断整个任务

## 4. 后端 API 端点

- [x] 4.1 在 `masterdata.api` 中添加 SyncResidentialAreasReq/Resp、GetSyncProgressReq/Resp 类型和 amap_sync group 端点声明
- [x] 4.2 运行 goctl 生成 handler 和 logic 骨架代码（handler/amap_sync/、logic/amap_sync/）
- [x] 4.3 确认 `routes.go` 中新增了 amap_sync 的路由注册
- [x] 4.4 填充 `syncResidentialAreasLogic.go`：参数校验、调用 SyncEngine.StartSync
- [x] 4.5 填充 `getSyncProgressLogic.go`：参数校验、调用 SyncEngine.GetProgress
- [x] 4.6 在 `serviceContext.go` 中添加 SyncEngine 字段并初始化

## 5. 后端编译验证

- [x] 5.1 `go build ./...` 编译通过
- [ ] 5.2 重启 masterdata 服务，确认无启动错误

## 6. 前端 API

- [x] 6.1 在 `web/pc/src/api/masterdata.ts` 中添加 `syncResidentialAreas` 和 `getSyncProgress` 接口函数

## 7. 前端同步页面

- [x] 7.1 创建 `web/pc/src/views/amap-sync/Index.vue` 基础页面结构
- [x] 7.2 实现省→市→区县→街道四级联动选择器（街道 el-select 多选）
- [x] 7.3 实现同步按钮和确认对话框
- [x] 7.4 实现进度面板：el-progress 进度条 + 状态文本（"正在处理第 N/M 个街道：街道名"）
- [x] 7.5 实现轮询逻辑：每 2 秒调用 getSyncProgress，completed/failed 时停止
- [x] 7.6 实现结果汇总展示：同步完成/失败提示和最终统计数据

## 8. 前端路由和菜单

- [x] 8.1 在 `web/pc/src/router/index.ts` 添加 `/masterdata/amap-sync` 路由
- [x] 8.2 在 `web/pc/src/components/layout/AppSidebar.vue` 添加"高德地图同步"菜单项和 Download 图标

## 9. 端到端验证

- [ ] 9.1 选择"东营区"→ 多选街道 → 触发同步 → 进度条正常推进 → 完成后查看数据库有新记录
- [ ] 9.2 重复同步同一街道 → 已存在小区被跳过（TotalSkipped > 0）
- [ ] 9.3 验证 md_residential_area 中 code 格式正确（如 `3705020010001`）且 longitude/latitude 有值
- [ ] 9.4 验证 submission_status = 2
