## Why

当前小区数据统计模块仅展示预计算数据（每日凌晨2点定时任务生成），用户无法获取当天最新的行政区划-小区/村统计数据。需要增加实时统计能力，让用户可以随时查看最新的小区分布情况，同时保留昨日数据作为历史对比基准。

## What Changes

- 新增实时统计 API 端点 `GET /api/masterdata/statistics/division-counts/realtime`，直接查询 residential_areas 和 divisions 表计算实时数据
- 实时接口返回格式与现有 `GET /api/masterdata/statistics/division-counts` 完全一致（DivisionCountItem），便于前端复用同一展示组件
- 实时查询同样只统计 `submission_status=2`（已审批）且未软删除的记录，与预计算逻辑保持一致
- 现有预计算统计接口和定时任务不受影响

## Capabilities

### New Capabilities
- `realtime-division-stats`: 实时行政区划小区/村统计查询能力，提供与昨日数据格式一致的实时统计 API

### Modified Capabilities
- `division-counts-ui`: 前端页面改造为双 Tab 结构（昨日数据 + 实时数据），仅影响前端展示层，后端接口不变

## Impact

- **后端代码**：新增 handler、logic 文件；.api 文件新增路由和类型定义；无需修改 model 层和定时任务
- **API**：新增 1 个 GET 端点，无 breaking changes
- **前端**：单独出改造方案，由其他 worktree 执行（不在本次实现范围）
