## Why

需要新增社区/村数量统计模块，以逐级下钻方式展示省→市→区县→街道各级别的小区和村注册数量，方便直观比较不同区域数据。

## What Changes

- **后端**：新增 `/api/masterdata/statistics/division-counts` API，按父级行政区划聚合统计下属小区/村数量
- **前端**：新增统计页面，4 面板逐级下钻展示，4 列（名称、小区数、村数、合计）
- **菜单**：主数据侧边栏新增"小区数据统计"菜单入口

## Capabilities

### New Capabilities
- `division-counts-api`: 后端统计 API（按父级聚合社区/村数量）
- `division-counts-ui`: 前端 4 面板逐级下钻统计页面

## Impact

- 后端：新增 handler/logic/types/routes
- 前端：新增页面、路由、API 函数、侧边栏菜单
