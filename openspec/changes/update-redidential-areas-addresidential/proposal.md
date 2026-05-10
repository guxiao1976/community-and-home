## Why

当前新建小区功能存在以下问题：1) 列表页点击"新建小区"时未校验用户是否已选择到行政区划的叶子节点（乡镇可直接新建，街道必须选到社区）；2) 新建页面仍需用户手动选择区县/街道/社区，且包含不必要的面积、人口字段，小区编码需手动输入；3) 小区类型需手动选择，但可根据是否选择了社区自动判断。

## What Changes

- 列表页"新建小区"按钮增加前置校验：必须选择到 `md_administrative_division` 叶子节点（街道有子社区则必须选社区，乡镇无子社区则可直接新建）
- 新建页面通过路由 query 携带 `county_id`、`street_id`、`community_div_id`、`division_name`，页面上显示为只读信息，不可修改
- 新建页面去掉小区面积、人口数量字段；地址改为选填
- 去掉小区编码输入框，编码由后端自动生成（区县编码 + 4位自增序号，如 `110105` + `0001`），保证唯一
- 小区类型自动选择默认值：有 `community_div_id` → 住宅小区(1)，无 `community_div_id`（仅乡镇）→ 村庄(2)
- 新建后提交状态为"已提交"（已有此逻辑，确认保持）

## Capabilities

### New Capabilities
- `auto-generate-code`: 后端自动生成小区编码，基于区县编码+4位序号，确保唯一
- `leaf-node-validation`: 新建前校验行政区划是否为叶子节点（`CountByParentId` 判断）

### Modified Capabilities
（无已有 spec 需要修改）

## Impact

- **后端 API**: `POST /api/masterdata/residential-areas` — `CreateResidentialAreaReq` 去掉 `Code` 必填，改为后端生成；`Address` 改为选填
- **后端 Logic**: `createResidentialAreaLogic.go` — 新增编码自动生成逻辑（查询当前区县最大序号+1）
- **后端 API**: 需新增一个接口判断行政区划是否为叶子节点，或复用前端已有的 `getAdministrativeDivisions` 接口在前端判断
- **前端 Form.vue**: 区划改为只读展示，去掉编码/面积/人口字段，小区类型自动填充默认值
- **前端 List.vue**: "新建小区"按钮增加叶子节点校验，校验通过后路由传参
- **数据库**: 无 schema 变更
