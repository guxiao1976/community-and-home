## 1. 后端 API 定义

- [x] 1.1 在 `masterdata.api` 中新增 `query` group，定义 `QueryResidentialAreasReq`（含 city_id/county_id/street_id/community_div_id/keyword/community_type/page/page_size）和 `QueryResidentialAreaItem`（含 city_name/county_name/street_name/community_name），路由 `GET /query/residential-areas`
- [x] 1.2 手动同步更新 `types.go`，添加请求和响应类型

## 2. 后端 Handler 和 Logic

- [x] 2.1 创建 handler 文件 `queryResidentialAreasHandler.go`
- [x] 2.2 注册路由到 `routes.go`
- [x] 2.3 创建 logic 文件 `queryResidentialAreasLogic.go` — 复用现有查询逻辑（默认 submission_status=2），遍历结果用 FindOne 查行政区划名称，做内存缓存避免重复查询

## 3. 前端 API

- [x] 3.1 在 `masterdata.ts` 中新增 `QueryResidentialAreaItem` 接口和 `queryResidentialAreas` API 函数

## 4. 前端页面

- [x] 4.1 创建 `web/pc/src/views/masterdata-query/Index.vue` — el-tabs 结构，第 1 个 tab "小区&村"，含五级级联查询条件 + el-table 结果列表（ID、名称、编码、城市名称(ID)、县区名称(ID)、街道名称(ID)、社区名称(ID)、地址、类型、操作-详情）+ 分页
- [x] 4.2 在 `router/index.ts` 中添加路由 `masterdata/query`

## 5. 编译验证

- [x] 5.1 后端编译通过：`cd services/masterdata/api && go build ./...`
