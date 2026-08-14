# Proposal: permission-service 角色列表接口新增排序支持

## 为什么做

角色列表接口（REST `GET /api/perm/roles` → RPC `ListRoles`）当前在数据访问层固定按 `sort_order asc` 排序（`services/permission-service/model/permission.go` FindList 硬编码 `order by sort_order asc`）。管理员无法按角色名称、角色编码、创建时间等字段排序浏览。随着角色数量增长（系统角色 + 自定义角色），固定排序降低可读性与检索效率，管理员需要按业务字段升/降序查看。

本变更在保持现有默认排序、分页与响应契约完全稳定的前提下，为角色列表引入**可选单列排序**：支持 `id, role_code, role_name, sort_order, status, created_at, updated_at` 七个白名单字段，方向 `asc`/`desc` 大小写不敏感，非法输入返回参数错误并指明非法字段/方向，排序恒追加 `id asc` 平局决胜保证分页稳定。

## 做什么

- **Proto（api-proto/permission/v1）**：`ListRolesRequest` 新增 `optional common.v1.SortField sort = 3`（单列排序）。`common.v1.SortField` 已存在于 `api-proto/api/common/v1/common.proto`（`field`/`order` 两个 string 字段），当前无消费方，**common.proto 无需变更**。
- **RPC（permission-service/rpc）**：`ListRolesLogic` 统一校验（单层校验，API 层不重复）：
  - `sort.field` 非空时须命中白名单，大小写不敏感；未命中 → 参数错误，消息指明非法字段。
  - `sort.order` 大小写不敏感；空值默认 `asc`；非 `asc`/`desc` → 参数错误，消息指明非法方向。
  - 校验失败以 `responsex.NewBaseRespWithError(errx.CodeInvalidParam /* 99400 */, …)` 写入响应 `Base`（项目 gRPC 层约定，业务错误走 Base 而非 Go error，见 `common/pkg/responsex/grpc.go`）。
  - 校验通过后将 `sort.field`/`sort.order` 透传给 `RoleModel.FindList`。
- **模型（permission-service/model）**：扩展 `FindList` 签名以接收排序字段与方向（决策 q9）；生成 `ORDER BY {field} {dir}, id asc`。`field` 仅允许拼接白名单字面量，杜绝 SQL 注入；无排序/字段为空时保持默认首键 `sort_order asc`，恒追加 `id asc` 平局决胜（现状 SQL 为 `order by sort_order asc`，主排序键保持一致，仅补齐平局稳定）。
- **REST（permission-service/api）**：`ListRolesReq` 新增 `sortBy`/`sortOrder`（camelCase，optional），透传至 gRPC `SortField`；API 层通过 `responsex.ToError(grpcResp.Base)` 将业务错误转换为 Go error（修复性前置：当前 ListRoles API 层未检查 Base，需补 `ToError`，符合经验记忆 [[rpc-callback-must-check-response-base]]），使非法排序在统一响应体中以 code=99400、data=null 呈现。
- **前端（web/pc）**：`views/roles/List.vue` 对白名单列（ID / 角色名称 / 角色编码 / 状态 / 创建时间）开启 `sortable`，排序时发送 `sortBy`（白名单 key）+ `sortOrder`；列 → 白名单 key 映射：`id→id`、`角色名称→role_name`、`角色编码→role_code`、`状态→status`、`创建时间→created_at`。`sortOrder` 列当前未在表格渲染，不入列。

## 影响范围

| 服务 | 变更类型 | 说明 |
|------|:---:|------|
| api-proto | Proto 变更 | `permission.proto` `ListRolesRequest` 增 `optional common.v1.SortField sort = 3`（common/v1.SortField 已存在，不改 common.proto）。需全局 Claude 执行 `make ci` 生成 |
| permission-service（rpc） | 修改 | `listroleslogic.go` 排序白名单校验 + 透传；`model/permission.go` FindList 签名扩展 + ORDER BY 组装 |
| permission-service（api） | 修改 | `api/internal/types/types.go` `ListRolesReq` 增 sortBy/sortOrder；`api/internal/logic/perm/listroleslogic.go` 透传 + `ToError(grpcResp.Base)` |
| web/pc | 修改 | `views/roles/List.vue` 白名单列 sortable；`src/api/identity.ts` `getRoles` 参数类型扩展 |

> 不改 common/、无数据迁移、无缓存变更（ListRoles 无缓存）。RPC 校验层在 permission-service 内部，单一服务职责。

## 已确认设计决策 → 覆盖对照

| # | 已确认决策 | proposal 章节 | spec Requirement | 覆盖 |
|---|-----------|--------------|-----------------|:---:|
| q1 | `ListRolesRequest` 新增 `optional common.v1.SortField sort = 3` | 做什么（Proto） | REQ-1 | ✅ |
| q2 | 单列排序 | 做什么（Proto） | REQ-1 | ✅ |
| q3 | 白名单 `id, role_code, role_name, sort_order, status, created_at, updated_at` | 做什么（RPC） | REQ-2 | ✅ |
| q4 | 拒绝并返回 400，指明非法字段/方向 | 做什么（RPC/REST） | REQ-2/REQ-3/REQ-8 | ✅ |
| q5 | 方向大小写不敏感，空值默认 asc，非法报 400 | 做什么（RPC） | REQ-3 | ✅ |
| q6 | 恒追加 `id asc` | 做什么（模型） | REQ-5 | ✅ |
| q7 | RPC logic 统一校验 | 做什么（RPC） | REQ-2/REQ-3 | ✅ |
| q8 | roles/List.vue 白名单列 sortable | 做什么（前端） | REQ-9 | ✅ |
| q9 | 扩展 FindList 签名 | 做什么（模型） | REQ-7 | ✅ |
| q10 | REST 参数 camelCase `sortBy`/`sortOrder` | 做什么（REST） | REQ-8 | ✅ |

> 术语说明：q4 的「返回 400」落地为本项目统一响应约定的**业务参数错误码** `errx.CodeInvalidParam`（99400，语义 Bad Request）。项目所有 REST 响应统一为 HTTP 200 + 响应体 `{code, msg, data}`（`common/pkg/responsex/response.go` 经 `httpx.OkJson` 输出），因此「400」以响应体 `code=99400` 呈现，前端按 `code !== 0` 拦截提示。

## 风险评估

- **SQL 注入（排序字段拼接 ORDER BY）**：可能性低 / 影响高。缓解：RPC 层白名单校验前置拦截，非法字段（含注入载荷）不会进入 SQL；模型层二次防御——`ORDER BY` 字段只允许拼接白名单字面量，方向仅 `asc`/`desc` 二选一，不拼接任何原始输入。REQ-6 覆盖。
- **默认排序行为回归**：可能性中 / 影响中。缓解：无 `sort` 或 `sort.field` 为空时保持现状 `sort_order asc`；恒追加 `id asc` 平局决胜，保证分页稳定且不破坏现有调用方。REQ-4/REQ-5 覆盖，并有显式回归场景。
- **字段命名错位导致排序静默失效**：`sortBy` 使用白名单（DB 列名 `role_name` 等），前端表格列用 `name`/`code`。缓解：spec REQ-9 显式定义列 → 白名单 key 映射，前端测试断言发出的 `sortBy` 值；如前端误发 `name`，后端白名单会拒绝（而非静默忽略），行为可观测。REQ-2/REQ-9 覆盖。
- **API 层未检查 gRPC Base 导致错误被吞**：可能性中 / 影响中（符合已有经验 [[rpc-callback-must-check-response-base]]）。缓解：REQ-8 显式要求 API 层 `ToError(grpcResp.Base)` 转换，否则校验错误不会呈现给前端。已在影响范围中声明。

## 验收标准

- 七个白名单字段（id/role_code/role_name/sort_order/status/created_at/updated_at）均支持 `asc`/`desc` 排序，大小写不敏感（`ROLE_CODE`/`DESC` 等同效）。
- 非法字段（如 `name`、`role_name; drop…`）与非 `asc`/`desc` 方向被拒绝：响应体 `code=99400`、`data=null`、`msg` 指明非法字段/方向。
- 不传 `sortBy`/`sortOrder` 时，主排序键与变更前一致（`sort_order asc` 为首键，恒追加 `id asc` 平局决胜）。
- 任意排序下结果恒以 `id asc` 作为末级稳定序，分页 `page/total/total_pages` 契约不变。
- `GET /api/perm/roles?sortBy=role_name&sortOrder=desc` 返回按 `role_name desc, id asc` 排序的数据。
- 前端 `roles/List.vue` 五个列可点击排序，升/降/取消循环，重载后排序保持；`el-table-column` 触发排序时请求携带正确的 `sortBy`（白名单 key）与 `sortOrder`。
