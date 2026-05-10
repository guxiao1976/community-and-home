## Context

基层组织模块（grassroots/Index.vue）已实现完整的三 Tab 审批工作流界面。行政区划模块（division/Index.vue）目前仅有简单树形展示和基础 CRUD，缺少提交管理、撤回、发起删除等操作。两个模块共用同一套后端 API（getAdministrativeDivisions、submitDivision、batchSubmitDivisions、withdrawDivision、requestDeleteDivision、cancelDeleteDivision、getMySubmissionRecords）。

## Goals / Non-Goals

**Goals:**
- 重写 division/Index.vue，对齐 grassroots/Index.vue 的三 Tab 结构和操作逻辑
- 复用已有的 API 函数和类型定义，不新增后端接口
- 行政区划管理 level 1-3 的审批工作流与基层组织 level 4-5 保持一致

**Non-Goals:**
- 不修改后端代码
- 不修改共享的 API 函数和类型定义
- 不修改基层组织模块

## Decisions

### 1. 直接重写 division/Index.vue 而非增量修改
基层组织模块是成熟参考实现，增量修改容易遗漏细节。直接以 grassroots/Index.vue 为蓝本重写，调整数据范围（level 1-3 替代 level 4-5）和树形加载逻辑（加载所有子级而非固定 level）。

### 2. 查询编辑 Tab 不使用级联选择器
基层组织用省/市/区级联选择器筛选街道，因为街道属于区县。行政区划模块本身管理的就是省/市/区数据，无需级联筛选，直接展示完整树形结构即可。

### 3. 不需要代码校验规则
基层组织的"代码以父级开头、长度与同级一致"校验适用于街道/社区级别。省/市/区代码有国家标准编码规则，暂不在前端强制校验。

### 4. 不需要添加下级时的父级代码预填
基层组织创建社区时需要预填街道代码，行政区划模块不适用此逻辑。

## Risks / Trade-offs

- [树形数据量] level 1-3 的行政区划数据量较大（数千条），使用懒加载树形展示 → 与基层组织一致，已验证可行
- [store vs 本地 ref] 当前 division 模块使用 Pinia store，基层组织使用本地 ref → 统一改为本地 ref，保持一致性，减少 store 依赖
