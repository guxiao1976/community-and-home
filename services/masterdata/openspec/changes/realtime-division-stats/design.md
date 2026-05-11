## Context

当前小区数据统计模块通过预计算方式工作：每日凌晨2点定时任务清空 `md_division_statistics` 表并重新计算省/市/区/街道四级的 community_count、village_count、total_count。前端通过 `GET /api/masterdata/statistics/division-counts?parent_id=X` 查询最新日期的统计数据。

用户需要在不等待定时任务的情况下查看当天最新数据。需要新增一个实时统计 API，直接查询 `md_residential_area` 和 `md_administrative_division` 表，返回格式与现有接口完全一致。

## Goals / Non-Goals

**Goals:**
- 新增实时统计 API `GET /api/masterdata/statistics/division-counts/realtime`，返回格式与现有 `division-counts` 接口一致
- 实时查询逻辑与预计算逻辑保持一致：仅统计 `submission_status=2` 且未软删除的记录
- 为前端提供完整的接口文档和前端改造方案

**Non-Goals:**
- 不修改现有预计算统计接口和定时任务
- 不做前端 web 代码改造（由其他 worktree 执行）
- 不引入缓存层（实时数据不应被缓存）
- 不新增数据库表

## Decisions

### 1. 实时查询复用同一 Model 层

**决定**：在 `mdDivisionStatisticsModel` 中新增 `FindRealtimeCountsByParentId` 方法，直接查询 `md_residential_area` + `md_administrative_division`。

**理由**：统计数据的核心逻辑（按 level 聚合、按 parent_id 过滤、审批状态过滤）集中在 Model 层。复用同一 Model 接口保持一致性，Logic 层只需调用不同方法。

**替代方案**：创建独立的 RealtimeStatisticsModel。但考虑到查询的是相同表，且方法签名一致（返回 `[]DivisionCountRow`），分开反而增加维护成本。

### 2. Level 1/2 聚合使用子查询而非多步查询

**决定**：省/市级统计通过 SQL 子查询从 `md_residential_area` 聚合，而非从区/街道级统计再聚合。

**理由**：预计算方式因为要写入中间表所以分步进行；实时查询可以直接用子查询一步到位，减少数据库交互次数，代码更简洁。

**SQL 策略**：
- **Level 3（区县）**：`GROUP BY county_id`，从 residential_area 按 county_id 分组
- **Level 4（街道）**：`GROUP BY street_id`，从 residential_area 按 street_id 分组
- **Level 2（市）**：先按 county_id 分组得到区县统计，再按区县的 parent_id 聚合到市级
- **Level 1（省）**：通过两级子查询聚合到省级

### 3. 新增独立 handler 而非在现有 handler 中分支

**决定**：新增 `getDivisionCountsRealtimeHandler`，放在同一 `statistics` group 下。

**理由**：符合 go-zero 的一个路由一个 handler 的约定，保持代码清晰。.api 文件中路由定义 `get /statistics/division-counts/realtime` 也自然对应独立 handler。

### 4. 请求/响应类型复用

**决定**：实时接口复用现有的 `DivisionCountsReq` 和 `DivisionCountsResp` 类型。

**理由**：参数格式和返回格式完全一致，无需定义新类型。前端可以用同一组件渲染两个 Tab 的数据。

## Risks / Trade-offs

- **[性能]** 实时查询在高数据量下可能比预计算慢 → 限制只查询已审批数据（submission_status=2），且同一时间只有一个用户操作统计页，并发压力低
- **[数据一致性]** 实时数据与昨日数据可能不一致（审批中的记录不计入实时数据）→ 这正是设计意图，两个 Tab 展示不同时间维度的数据
- **[SQL 复杂度]** 省/市级的嵌套子查询较复杂 → 已有预计算逻辑作为参考，SQL 结构一致；且实际行政区划层级有限（~34省、~330市），查询结果集小
