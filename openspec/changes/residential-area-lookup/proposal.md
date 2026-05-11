## Why

用户需要通过小区名称或代码快速反向查询其行政区划归属链路（省→市→区县→街道→社区）。当前系统只有正向筛选（选省→市→区县→小区），缺少反向查询能力。该功能是数据管理和运维的常用需求。

## What Changes

- 新增后端 API：输入关键词（小区/村名或代码），返回匹配的小区列表及其完整行政区划归属链路
- 新增前端页面：提供搜索输入框，展示查询结果，每个小区显示一行归属信息（省份、城市、区县、街道、社区）
- 复用现有的 `SearchByName`、`FindByCode`、`FindChildren` 等模型方法，无需新建数据库表

## Capabilities

### New Capabilities
- `residential-area-lookup`: 住宅小区反向查询 — 输入小区名称或代码，返回小区信息及行政区划归属链路

### Modified Capabilities

## Impact

- 后端：`masterdata.api` 新增路由、`types.go` 新增类型、新增 handler 和 logic
- 后端 model：可能需新增按名称/代码模糊查询的方法（现有 `SearchByName` 可复用）
- 前端：新增页面和路由、`masterdata.ts` 新增 API 调用
- 无数据库 schema 变更
