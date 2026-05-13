## 1. Model 层 - 实时统计查询方法

- [x] 1.1 在 `MdDivisionStatisticsModel` 接口和 `customMdDivisionStatisticsModel` 中新增 `FindRealtimeCountsByParentId(ctx, parentId *int64) ([]DivisionCountRow, error)` 方法
- [x] 1.2 实现 Level 3（区县）实时查询 SQL：按 `county_id` 聚合 `md_residential_area`，关联 `md_administrative_division` 获取名称/级别
- [x] 1.3 实现 Level 4（街道）实时查询 SQL：按 `street_id` 聚合
- [x] 1.4 实现 Level 2（市）实时查询 SQL：通过子查询从区县级统计按 `parent_id` 聚合
- [x] 1.5 实现 Level 1（省）实时查询 SQL：通过两级子查询从区县级统计聚合到省级
- [x] 1.6 无 parent_id 时查询顶级（level=1）区划逻辑

## 2. API 定义 - 路由和类型

- [x] 2.1 在 `masterdata.api` 的 `statistics` group 中新增路由 `get /statistics/division-counts/realtime (DivisionCountsReq) returns (DivisionCountsResp)`
- [x] 2.2 运行 `goctl api go` 重新生成 handler 和 types 代码

## 3. Logic 层 - 实时统计业务逻辑

- [x] 3.1 实现 `getDivisionCountsRealtimeLogic.go`：调用 `FindRealtimeCountsByParentId`，转换为 `DivisionCountItem` 返回
- [x] 3.2 错误处理：查询失败返回空列表（与现有逻辑一致）

## 4. 编译验证

- [x] 4.1 `go build ./...` 编译通过，无错误
- [x] 4.2 新增 handler 在路由中正确注册

## 5. 前端改造方案（交付文档，不在本 worktree 实施）

- [x] 5.1 输出前端改造方案文档，包含：Tab 结构设计、新增 API 接口说明、组件复用方案
