## 1. 路由与菜单

- [x] 1.1 在 `src/router/index.ts` 中新增 `/masterdata/grassroots` 路由，指向 `views/grassroots/Index.vue`，title 为"基层组织"
- [x] 1.2 在 `src/components/layout/AppSidebar.vue` 的"主数据管理"菜单下新增"基层组织"菜单项（path: `/masterdata/grassroots`），位于"行政区划"之后，导入 `OfficeBuilding` 图标

## 2. 修改行政区划页面（限制到 level 1-3）

- [x] 2.1 修改 `views/division/Index.vue` 初始化查询，确保 `fetchDivisions({ level: 1 })` 只加载省级数据
- [x] 2.2 修改 `loadChildren` 回调，过滤掉 level > 3 的子节点，使树只展开到区县级
- [x] 2.3 修改"添加下级"按钮的显示条件，区县行（level 3）不再显示该按钮
- [x] 2.4 修改新增对话框的级别选项，只保留省级(1)、市级(2)、区县级(3)
- [x] 2.5 修改 `handleCreateChild`，限制只能在省级下添加市级、市级下添加区县级

## 3. 新增基层组织页面

- [x] 3.1 创建 `views/grassroots/Index.vue` 基础结构：页面容器、搜索栏（级联选择器 + 搜索按钮 + 新增按钮）、树形表格、新增/编辑对话框
- [x] 3.2 实现省-市-区县三级级联选择器（`el-cascader` lazy 模式），初始加载省级列表，逐级加载下一级
- [x] 3.3 实现搜索功能：选中区县后点击搜索，调用 `getAdministrativeDivisions({ parent_id: districtId })` 加载街道列表
- [x] 3.4 实现树形表格展示街道和社区，街道为根节点，社区通过 `loadChildren` 懒加载
- [x] 3.5 实现新增街道功能：点击新增按钮，在选中区县下创建 level 4 的街道/乡镇
- [x] 3.6 实现新增社区功能：街道行点击"添加下级"，创建 level 5 的社区/村
- [x] 3.7 实现编辑、提交审核、删除功能，复用 `useDivisionStore` 的现有方法
- [x] 3.8 添加权限控制：根据 `submission_status` 控制编辑、提交、删除按钮的显示

## 4. 验证

- [x] 4.1 验证行政区划页面：只能看到和管理省-市-区县，无法看到或操作街道/社区
- [x] 4.2 验证基层组织页面：级联选择器正常工作，搜索能正确加载街道，CRUD 操作正常
- [x] 4.3 验证侧边栏菜单：两个菜单项都能正确导航到对应页面
