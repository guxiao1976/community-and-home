## Context

住宅小区列表页 `List.vue` 当前使用一个 `el-cascader` 组件加载全部行政区划数据（5000 条），从区县（level 3）开始选择。后端 `GET /api/masterdata/divisions` 已支持 `level` 和 `parent_id` 参数进行层级筛选，`GET /api/masterdata/residential-areas` 已支持 `county_id` 参数筛选。无需后端改动。

## Goals / Non-Goals

**Goals:**
- 将筛选区改为三个独立的 `el-select` 下拉：省份 → 城市 → 区县，逐级联动
- 选择省份后加载对应城市，选择城市后加载对应区县
- 未选择城市时禁用搜索按钮，防止无效的全量查询

**Non-Goals:**
- 不修改后端 API
- 不修改新建/编辑小区表单页的行政区划选择器（保持 cascader）
- 不增加街道（level 4）筛选

## Decisions

### 1. 使用三个独立 el-select 而非 el-cascader

**选择**: 拆为三个 `el-select`，每次切换上级时清空下级并重新加载。

**替代方案**: 继续用 `el-cascader` 但从 level 1 开始显示。

**理由**: 用户明确要求"增加省份、城市下拉列表"，三个独立下拉视觉更清晰，且可以实现"必须选择到城市"的校验逻辑。

### 2. 按需加载子级数据

**选择**: 页面加载时仅请求 level=1（省份）；选择省份后请求 `parent_id=X, level=2`（城市）；选择城市后请求 `parent_id=Y, level=3`（区县）。

**替代方案**: 一次性加载全部行政区划，前端构建树。

**理由**: 按需加载减少首次请求数据量，且已有后端 API 直接支持。

### 3. 搜索校验方式

**选择**: 当城市未选择时，搜索按钮 `:disabled="!filters.city_id"`。

**理由**: 最直观的用户体验，避免发送无意义请求。

## Risks / Trade-offs

- **[风险] 区划数据量大时按需加载有延迟** → 行政区划数据量有限（省 30+、市 300+），响应很快
- **[风险] 选择省份后切换导致城市列表闪烁** → 使用 `loading` 状态提示用户
