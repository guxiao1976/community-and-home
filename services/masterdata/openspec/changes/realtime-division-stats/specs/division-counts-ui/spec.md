## MODIFIED Requirements

### Requirement: 4面板逐级下钻页面
统计页面 SHALL 包含两个 Tab 页：「昨日数据」和「实时数据」。每个 Tab 包含4个等宽面板（省、市、区县、街道/乡镇），每面板展示名称、小区数、村数、合计4列，按合计降序排列。

#### Scenario: 昨日数据 Tab 默认展示
- **WHEN** 用户进入小区数据统计页面
- **THEN** 默认显示「昨日数据」Tab，调用 `GET /api/masterdata/statistics/division-counts` 接口，展示预计算的统计数据

#### Scenario: 切换到实时数据 Tab
- **WHEN** 用户点击「实时数据」Tab
- **THEN** 调用 `GET /api/masterdata/statistics/division-counts/realtime` 接口，展示实时统计数据，4面板展示组件与昨日数据 Tab 完全复用

#### Scenario: 实时数据 Tab 逐级下钻
- **WHEN** 用户在实时数据 Tab 中点击某个省
- **THEN** 第2面板加载该省下辖市的实时数据（调用 realtime 接口传入 parent_id），第3-4面板重置为空状态提示

#### Scenario: 实时数据 Tab 面包屑导航
- **WHEN** 用户在实时数据 Tab 中下钻到区县级
- **THEN** 顶部显示"省名 > 市名 > 区县名"，可点击快速回退

#### Scenario: 实时数据 Tab 列头排序
- **WHEN** 用户在实时数据 Tab 中点击"小区"/"村"/"合计"列头
- **THEN** 该列数据切换升序/降序（前端本地排序）
