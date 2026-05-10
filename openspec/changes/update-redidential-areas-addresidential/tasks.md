## 1. 后端 — 编码自动生成

- [x] 1.1 `mdResidentialAreaModel.go` 新增 `GetMaxCodeByCountyId(ctx, countyId)` 方法：查询 `WHERE county_id = ? AND code LIKE ? ORDER BY code DESC LIMIT 1`，返回最大编码或空字符串
- [x] 1.2 `createResidentialAreaLogic.go` 新增 `generateCode` 函数：查区县 code → 查当前最大序号 → 拼接新编码，失败时重试一次
- [x] 1.3 `createResidentialAreaLogic.go` 修改 `CreateResidentialArea`：忽略前端传入的 code，调用 `generateCode` 自动生成；`Address` 改为选填（允许空字符串）

## 2. 后端 — API 类型更新

- [x] 2.1 `masterdata.api` 中 `CreateResidentialAreaReq` 的 `Code` 改为 `optional`，`Address` 改为 `optional`
- [x] 2.2 `types.go` 中 `CreateResidentialAreaReq` 同步修改 `Code` 和 `Address` 为 optional

## 3. 前端 — 列表页叶子节点校验与路由传参

- [x] 3.1 `List.vue` 新增 `communityOptionsLoadFailed` ref，在 `handleStreetChange` 中设置加载失败状态（catch 时设为 true，成功时设为 false）
- [x] 3.2 `List.vue` 重写 `handleCreate`：校验选择深度（city→district→street→community/leaf），通过后 `saveFilterState` 并 `router.push` 携带 `county_id`、`street_id`、`community_div_id`、`county_name`、`street_name`、`community_name` 等 query 参数
- [x] 3.3 校验逻辑：未选 district → 提示"请选择到区县"；未选 street → 提示"请选择街道/乡镇"；选了 street 且 `communityOptions.length > 0` 但未选 community → 提示"请选择社区"；`communityOptionsLoadFailed` → 提示"无法确认，请重试"

## 4. 前端 — 新建表单重构

- [x] 4.1 `Form.vue` 去掉 `loadDivisions`、`buildDivisionTree`、`handleDivisionChange`、区县级联选择器（el-cascader），改为从 `route.query` 读取区划信息，以只读文本展示路径（如"陕西省 > 咸阳市 > 秦都区 > XX街道 > XX社区"）
- [x] 4.2 `Form.vue` 去掉编码输入框、小区面积、人口数量字段；地址改为选填（去掉 required 规则）
- [x] 4.3 `Form.vue` `onMounted` 中根据 `route.query.community_div_id` 是否存在自动设置 `community_type`（有社区→1，无社区→2）
- [x] 4.4 `Form.vue` `handleSubmit` 中新建时不再发送 `code`、`area`、`population` 字段，`address` 为空时不发送

## 5. 构建、重启与验证

- [x] 5.1 go build 编译 masterdata API 和 RPC 服务，重启
- [ ] 5.2 验证：选到社区后新建 → 表单显示只读区划路径，编码自动生成，类型默认住宅小区，无面积/人口字段
- [ ] 5.3 验证：仅选乡镇（无子社区）后新建 → 表单显示4级路径，类型默认村庄
- [ ] 5.4 验证：选了街道但未选社区（该街道有子社区） → 提示必须选社区
