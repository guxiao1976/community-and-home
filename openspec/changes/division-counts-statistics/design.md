## Context

复用已有的 MdResidentialAreaModel.Count 方法和 MdAdministrativeDivisionModel 的查询能力。前端使用 Element Plus 表格组件。

## Decisions

### 1. 后端 SQL 一次聚合
用一条 SQL 完成：查子级行政区划 + LEFT JOIN 统计其下 community_type=1(小区) 和 community_type=2(村) 的数量。避免 N+1 查询。

### 2. 前端 4 面板用 Flex 布局
每个面板 flex: 1，min-width: 0，自适应等分。面板间无箭头图标，通过面包屑和选中高亮体现层级关系。

### 3. 面包屑放在 panels 外部顶部
面包屑与面板平级放置，不在面板标题内，节省面板空间。
