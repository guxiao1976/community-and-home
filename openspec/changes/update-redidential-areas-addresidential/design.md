## Context

当前新建小区流程：列表页点击"新建" → 跳转 Form.vue → 用户手动选择区县/街道/社区、手动填写编码/名称/地址/面积/人口/类型。后端 `createResidentialAreaLogic.go` 在创建时设置 `submission_status=1`（已提交），校验编码唯一性。

需要改进：1) 列表页已有多级区划筛选，新建时应复用已选区划；2) 编码应后端自动生成而非手动输入；3) 小区类型可根据区划层级自动推断；4) 去掉不必要的面积/人口字段。

## Goals / Non-Goals

**Goals:**
- 列表页"新建小区"按钮校验已选到行政区划叶子节点才允许新建
- 新建页面通过路由 query 携带区划信息，页面上只读展示
- 后端自动生成编码（区县编码 + 4位序号），确保唯一
- 小区类型根据是否有 `community_div_id` 自动默认（有社区→住宅小区，仅乡镇→村庄）
- 去掉编码输入框、面积、人口字段，地址改为选填

**Non-Goals:**
- 编辑页面的区划信息仍可修改（本次不涉及编辑流程改动）
- 不修改审批/审核流程
- 不引入新的数据库表或字段

## Decisions

### D1: 叶子节点校验在前端执行

**选择**: 前端利用已有的 `streetOptions` / `communityOptions` 判断是否为叶子节点。

**原因**: 列表页已加载了 `communityOptions`（街道下的社区列表）。如果 `street_id` 已选但 `communityOptions` 为空，说明该街道/乡镇没有子社区，即为叶子节点。无需新增后端接口。

**替代方案**: 后端新增 `GET /api/masterdata/divisions/:id/is-leaf` 接口 — 增加一次网络请求，不必要。

**实现**:
- 选中街道后，前端调用 `getAdministrativeDivisions({ parent_id: streetId, level: 5 })` 加载社区列表
- `handleCreate` 校验：`community_div_id` 有值（选了社区）或 `communityOptions.length === 0`（乡镇无子社区）
- 不满足时 `ElMessage.warning` 提示"请选择到社区后再新建小区"

### D2: 编码自动生成在后端执行

**选择**: 后端 `createResidentialAreaLogic` 根据区县编码自动生成。

**原因**: 编码唯一性保证需要数据库查询，后端更可靠，避免并发问题。

**编码规则**: 区县 `code`(6位) + 4位递增序号（从 0001 开始）。查询 `md_residential_area` 表中相同 `county_id` 前缀的最大编码序号，+1 生成新编码。例如区县 code `110105` → 编码 `1101050001`、`1101050002`...

**实现**:
- 新增 model 方法 `GetMaxCodeByCountyId(ctx, countyId)` 返回当前最大序号
- Logic 层在 `Insert` 前生成编码
- `CreateResidentialAreaReq` 中 `Code` 字段改为 `optional`，后端忽略前端传入的 code

### D3: 小区类型默认值在前端设置

**选择**: 前端根据路由 query 中是否携带 `community_div_id` 自动设置默认值。

**原因**: 纯 UI 层逻辑，无需后端参与。用户仍可手动修改类型。

**实现**: Form.vue 的 `onMounted` 中，根据 `route.query.community_div_id` 是否存在设置 `formData.community_type`。

### D4: 区划信息通过路由 query 传递

**选择**: `router.push` 时携带 `?county_id=XX&street_id=XX&community_div_id=XX&division_name=XX`。

**原因**: 简单直接，无需引入额外状态管理。使用 sessionStorage 已有的 `FILTER_KEY` 机制也可，但 query 参数更符合语义。

**实现**:
- List.vue `handleCreate` 中构造 query 对象（包含 county_id、street_id、community_div_id 以及对应的中文名称）
- Form.vue 从 `route.query` 读取并填充到表单，展示为只读文本

## Risks / Trade-offs

- **[编码并发]** 两个请求同时创建同一区县的小区可能生成相同编码 → 使用数据库唯一索引 `UNIQUE(code)` 作为兜底，插入失败时重试一次
- **[叶子节点判断依赖前端数据]** 如果社区列表加载失败，可能导致误判 → 加载失败时 `communityOptions` 为空，与真正的"无子社区"乡镇无法区分。处理方式：加载失败时 `communityOptions` 设为 `null`，校验时 `null` 与空数组 `[]` 区分开，`null` 时阻止新建并提示"无法确定是否为叶子节点，请重试"
