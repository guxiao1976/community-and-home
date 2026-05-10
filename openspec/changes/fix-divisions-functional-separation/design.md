## Context

当前 `views/division/Index.vue` 使用 `el-table` 的树形懒加载展示全部五个层级的行政区划（省→市→区县→街道→社区），所有 CRUD 操作共享同一套 API 和 Pinia store（`stores/division.ts`）。路由在 `/masterdata/divisions`，侧边栏菜单位于"主数据管理"分组下。

后端 API `getAdministrativeDivisions` 已支持 `parent_id` 和 `level` 过滤参数，无需修改即可按层级查询。store 的 `loadChildren` 方法也基于 `parent_id` 加载子节点。

## Goals / Non-Goals

**Goals:**

- 将行政区划管理拆分为两个独立页面，按变更频率分离关注点
- 行政区划页面只展示和管理省（level 1）、市（level 2）、区县（level 3）
- 基层组织页面通过区县级联选择器定位，管理街道（level 4）和社区（level 5）
- 两个页面复用现有 API 层和 store，不引入新的后端接口

**Non-Goals:**

- 不修改后端 API 或数据模型
- 不修改审核中心的工作流程
- 不修改住宅小区等下游模块的区划选择逻辑
- 不调整 store 内部实现（`stores/division.ts` 保持不变）

## Decisions

### 1. 复用现有 store，而非新建独立 store

**选择**: 基层组织页面和行政区划页面共用 `useDivisionStore`。

**理由**: 两个页面操作的是同一张表、同一套 API，区别仅在于查询参数（level/parent_id 过滤）。新建 store 会引入重复代码和状态同步问题。

**替代方案**: 为基层组织新建 `useGrassrootsStore` — 放弃，因为 CRUD 逻辑完全相同，仅过滤条件不同。

### 2. 基层组织页面使用级联选择器 + 树形表格

**选择**: 页面顶部放置省-市-区县三级 `el-cascader`，选中区县后查询其下属街道，街道展开后懒加载社区。

**理由**: 区县数量可能很多（全国 3000+），不适合一次性加载全部。级联选择器让用户快速定位目标区县，减少数据量。选中后只加载该区县下的街道和社区，性能好且聚焦。

**实现**: 初始加载省级列表（level=1），选中省后加载市级（parent_id=省id），选中市后加载区县级（parent_id=市id）。点击【搜索】后用选中区县的 id 调用 `getAdministrativeDivisions({ parent_id: districtId })` 获取街道列表，社区通过 `loadChildren` 懒加载。

### 3. 行政区划页面限制为三级

**选择**: 修改 `views/division/Index.vue`，在初始化查询时只获取 level 1（省级），懒加载子节点时限制到 level 3。移除 level 4/5 的"添加下级"按钮，新增对话框的级别选项只保留省级/市级/区县级。

**理由**: 简单直接，无需创建新组件。通过过滤条件实现功能隔离，用户不会在行政区划页面看到街道/社区数据。

### 4. 新增/编辑对话框按页面区分级别范围

**选择**: 行政区划页面的新增对话框级别选项为 1-3，基层组织页面的新增街道对话框级别固定为 4、新增社区对话框级别固定为 5。

**理由**: 级别在对应页面内是确定的，不需要用户手动选择，减少出错可能。

## Risks / Trade-offs

- **[两个页面共用同一个 store 实例]** → 两个页面同时打开时，store 状态可能互相覆盖。缓解：当前项目每个页面独立使用，不需要同时打开两个页面管理区划；如果将来需要，可以为基层组织页面创建独立 store 实例（`useDivisionStore('grassroots')`）。
- **[级联选择器加载体验]** → 省→市→区县逐级加载可能有短暂等待。缓解：数据量小（省~34、每省市~20、每市区~15），响应很快；使用 `el-cascader` 的动态加载（lazy）模式，选中即加载下一级。
- **[行政区划页面已有街道/社区数据]** → 如果数据库中已有 level 4/5 数据关联到 level 3 节点，行政区划页面的 `hasChildren` 判断需要调整。缓解：修改 `buildTree` 中的 `hasChildren` 逻辑为 `item.level < 3`（而非当前的 `< 5`），并在 `loadChildren` 中过滤只返回 level ≤ 3 的子节点。
