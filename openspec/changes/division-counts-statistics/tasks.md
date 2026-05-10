## 1. 后端

- [x] 1.1 types.go 新增 DivisionCountsReq、DivisionCountItem、DivisionCountsResp 类型
- [x] 1.2 model 新增 CountByDivisionParent 方法（SQL 一次聚合子级区划+社区/村数量）
- [x] 1.3 新增 logic getDivisionCountsLogic
- [x] 1.4 新增 handler getDivisionCountsHandler
- [x] 1.5 注册路由 GET /api/masterdata/statistics/division-counts

## 2. 前端

- [x] 2.1 masterdata.ts 新增 getDivisionCounts API 函数
- [x] 2.2 masterdata.d.ts 新增 DivisionCountItem 类型
- [x] 2.3 新建 statistics/division-counts/Index.vue（4 面板 + 面包屑）
- [x] 2.4 router 注册路由 /masterdata/statistics/division-counts
- [x] 2.5 侧边栏菜单新增"小区数据统计"入口

## 3. 验证

- [x] 3.1 后端编译通过
- [ ] 3.2 前端页面可访问，省级列表按合计降序显示
- [ ] 3.3 逐级下钻数据正确，面包屑可回退
