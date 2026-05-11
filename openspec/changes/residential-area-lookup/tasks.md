## 1. 后端 Model 层

- [x] 1.1 在 `mdResidentialAreaModel.go` 新增 `SearchByKeyword` 方法 — 按 name LIKE 或 code LIKE 查询，不限制区域
- [x] 1.2 在 `mdResidentialAreaModel.go` 新增 `CountByKeyword` 方法 — 返回匹配总数用于分页
- [x] 1.3 在 `MdResidentialAreaModel` interface 中添加上述两个方法签名

## 2. 后端 API 定义

- [x] 2.1 在 `masterdata.api` 中新增 lookup 路由 `GET /api/masterdata/residential-areas/lookup`，定义 `LookupResidentialAreasReq` 和 `LookupResidentialAreasResp` 类型
- [x] 2.2 手动同步更新 `types.go`（添加请求/响应类型，resp 中包含 province_name/city_name/county_name/street_name/community_name）

## 3. 后端 Handler 和 Logic

- [x] 3.1 用 goctl 生成或手动创建 handler 文件 `lookupResidentialAreasHandler.go`
- [x] 3.2 创建 logic 文件 `lookupResidentialAreasLogic.go` — 调用 SearchByKeyword 查小区，用 county_id 查 division 表获取 path 解析省市，用 street_id/community_div_id 查街道和社区名称

## 4. 前端 API

- [x] 4.1 在 `masterdata.ts` 中新增 `LookupResidentialAreaItem` 接口和 `lookupResidentialAreas` API 函数

## 5. 前端页面

- [x] 5.1 创建 `web/pc/src/views/residential-areas/Lookup.vue` — 搜索框 + el-table 结果列表（列：小区名、代码、省、市、区县、街道、社区）+ 分页
- [x] 5.2 在 `router/index.ts` 中添加路由 `masterdata/residential-areas/lookup`

## 6. 编译验证

- [x] 6.1 后端编译通过：`cd services/masterdata/api && go build ./...`
- [x] 6.2 前端编译通过：`cd web/pc && npm run build`（Lookup.vue 无新增错误，已有 TS 错误均非本次引入）
