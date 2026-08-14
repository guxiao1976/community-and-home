# Role List Sort（角色列表排序）Specification

## Purpose

为 permission-service 的角色列表接口（RPC `ListRoles` / REST `GET /api/perm/roles`）提供可选的单列排序能力，使管理员能按 `id / role_code / role_name / sort_order / status / created_at / updated_at` 七个白名单字段升序或降序浏览角色，同时保持无排序时的默认行为与分页契约完全不变，并对非法字段/方向给出明确、可机器判别的参数错误。

## 术语

- **白名单（whitelist）**：本 spec 规定的可排序字段集合 `{id, role_code, role_name, sort_order, status, created_at, updated_at}`。这些键即 `sys_role` 表的 DB 列名，RPC `sort.field` 与 REST `sortBy` 使用同一套键。
- **参数错误（invalid-param error）**：指业务错误码 `errx.CodeInvalidParam`（99400，语义 Bad Request）。按项目统一响应约定（HTTP 200 + 响应体 `{code, msg, data}`），其呈现为 `code=99400`、`msg` 指明具体非法字段/方向、`data=null`。
- **平局决胜（tiebreaker）**：除请求字段外，`ORDER BY` 恒以 `id asc` 收尾，保证排序全序确定、分页稳定。

## Requirements

### Requirement: 排序入参契约（单列排序）
系统 SHALL 在 `ListRolesRequest` 中提供单列排序入参 `optional common.v1.SortField sort = 3`；`SortField` 包含至多一组 `field` + `order`，即系统 SHALL 支持单列排序、不支持多列组合排序。

#### Scenario: 单列排序生效
- **GIVEN** 调用方构造 `ListRolesRequest` 且 `sort.field = "role_name"`、`sort.order = "desc"`
- **WHEN** RPC `ListRoles` 处理该请求
- **THEN** 响应按 `role_name desc` 排序，且以 `id asc` 收尾

#### Scenario: 排序入参缺省不影响现有调用
- **GIVEN** 调用方未携带 `sort`（或仅带 `page`/`status`）
- **WHEN** RPC `ListRoles` 处理该请求
- **THEN** 以 `sort_order asc` 为首键、`id asc` 平局决胜返回（主排序键与变更前一致）

### Requirement: 字段白名单校验
系统 SHALL 对非空的 `sort.field` 进行白名单校验（匹配 `{id, role_code, role_name, sort_order, status, created_at, updated_at}`），匹配 MUST 大小写不敏感（字段统一转小写后比对）；非空且不在白名单内的字段 MUST 被拒绝，返回参数错误，且 `msg` MUST 指明非法字段名。

#### Scenario: 白名单字段被接受
- **GIVEN** `sort.field = "role_code"`、`sort.order = "asc"`
- **WHEN** RPC `ListRoles` 处理该请求
- **THEN** 校验通过，结果按 `role_code asc, id asc` 排序返回

#### Scenario: 非法字段被拒绝并指明
- **GIVEN** `sort.field = "role_name_typo"`
- **WHEN** RPC `ListRoles` 处理该请求
- **THEN** 返回参数错误（`code=99400`、`data=null`），`msg` 包含非法字段名 `role_name_typo`，不执行任何查询

#### Scenario: 字段大小写不敏感
- **GIVEN** `sort.field = "ROLE_CODE"`、`sort.order = "asc"`
- **WHEN** RPC `ListRoles` 处理该请求
- **THEN** 校验通过，等效于 `role_code asc, id asc`

### Requirement: 排序方向校验
系统 SHALL 将 `sort.order` 按大小写不敏感方式规范化；空值 MUST 默认 `asc`；规范化后非 `asc` 或 `desc` 的方向 MUST 被拒绝，返回参数错误，且 `msg` MUST 指明非法方向值。

#### Scenario: desc 方向（大小写不敏感）
- **GIVEN** `sort.field = "created_at"`、`sort.order = "DESC"`
- **WHEN** RPC `ListRoles` 处理该请求
- **THEN** 校验通过，结果按 `created_at desc, id asc` 排序返回

#### Scenario: 空方向默认 asc
- **GIVEN** `sort.field = "role_name"`、`sort.order` 为空字符串或未设置
- **WHEN** RPC `ListRoles` 处理该请求
- **THEN** 按 `role_name asc, id asc` 排序返回

#### Scenario: 非法方向被拒绝并指明
- **GIVEN** `sort.field = "status"`、`sort.order = "sideways"`
- **WHEN** RPC `ListRoles` 处理该请求
- **THEN** 返回参数错误（`code=99400`、`data=null`），`msg` 指明非法方向值 `sideways`

### Requirement: 默认排序保持兼容
系统 SHALL 在排序入参缺省、或 `sort.field` 为空时，返回以 `sort_order asc` 为首键、`id asc` 收尾的默认顺序（主排序键与变更前一致）；空字段即使携带了非空方向，也 MUST 视为未指定排序，不得报错。

#### Scenario: 无排序入参保持现状
- **GIVEN** 请求未携带 `sort`
- **WHEN** RPC `ListRoles` 处理该请求
- **THEN** 结果以 `sort_order asc` 为首键、`id asc` 为平局决胜返回（主排序键与变更前一致）

#### Scenario: 空字段 + 方向不报错
- **GIVEN** `sort.field` 为空字符串、`sort.order = "desc"`
- **WHEN** RPC `ListRoles` 处理该请求
- **THEN** 视为未指定排序，按 `sort_order asc, id asc` 返回，不返回参数错误

### Requirement: id 平局决胜
系统 SHALL 在所有排序场景（含默认排序与任意请求字段/方向）将 `id asc` 作为 `ORDER BY` 的最后一级，以保证全序确定与分页稳定。

#### Scenario: 显式排序追加 id asc
- **GIVEN** `sort.field = "role_name"`、`sort.order = "desc"`
- **WHEN** 生成查询
- **THEN** 查询的 `ORDER BY` 为 `role_name desc, id asc`

#### Scenario: 默认排序含 id asc
- **GIVEN** 无排序入参
- **WHEN** 生成查询
- **THEN** 查询的 `ORDER BY` 为 `sort_order asc, id asc`

### Requirement: 排序字段注入安全
系统 MUST 仅将白名单校验通过的字段字面量拼入 `ORDER BY` 子句；MUST NOT 将任何原始用户输入直接拼入 SQL。方向 MUST 仅取规范化后的 `asc`/`desc` 之一。

#### Scenario: 注入载荷被白名单拦截
- **GIVEN** `sort.field = "role_name; drop table sys_role"`、`sort.order = "asc"`
- **WHEN** RPC `ListRoles` 处理该请求
- **THEN** 白名单校验拒绝该字段（返回参数错误），该载荷不会进入任何 SQL 语句

#### Scenario: 方向仅白名单二值
- **GIVEN** `sort.order` 经校验后仅可能为 `asc` 或 `desc`
- **WHEN** 数据访问层组装 `ORDER BY`
- **THEN** 方向分支只接受这两个字面量，不拼接其他字符串

### Requirement: 数据访问层契约
系统 SHALL 扩展角色列表数据访问层（`RoleModel.FindList`）的签名，使其接受已校验的排序字段与方向，并将该字段/方向作为 `ORDER BY` 的首键、`id asc` 作为末键；排序入参缺省时使用 `sort_order asc, id asc`。

#### Scenario: 数据访问层按字段排序
- **GIVEN** 数据访问层收到已校验的 `sortField = "created_at"`、`sortOrder = "desc"`
- **WHEN** 执行分页查询
- **THEN** 查询按 `created_at desc, id asc` 排序，分页偏移/限制逻辑不变

#### Scenario: 数据访问层默认排序
- **GIVEN** 数据访问层收到空的排序字段/方向
- **WHEN** 执行分页查询
- **THEN** 查询按 `sort_order asc, id asc` 排序

### Requirement: REST 参数契约与错误呈现
系统 SHALL 在 REST `GET /api/perm/roles` 接受可选查询参数 `sortBy` 与 `sortOrder`（camelCase），并将其透传至 RPC 的 `sort.field` / `sort.order`；API 层 MUST 将 RPC 响应 `Base` 中的业务错误转换为统一错误响应，使校验失败以 `code=99400`、`msg` 指明非法字段/方向、`data=null` 呈现。

#### Scenario: sortBy/sortOrder 透传生效
- **GIVEN** `GET /api/perm/roles?sortBy=role_name&sortOrder=desc`
- **WHEN** 请求经 API 层透传至 RPC `ListRoles`
- **THEN** 响应按 `role_name desc, id asc` 排序返回，分页字段（page/pageSize/total/totalPages）契约不变

#### Scenario: 非法字段呈现统一错误
- **GIVEN** `GET /api/perm/roles?sortBy=evil`
- **WHEN** API 层将 RPC 的 `Base` 业务错误转换为 Go error 并输出
- **THEN** 响应体为 `code=99400`、`data=null`、`msg` 指明非法字段 `evil`

#### Scenario: 无排序参数默认排序
- **GIVEN** `GET /api/perm/roles`（不带 `sortBy`/`sortOrder`）
- **WHEN** API 层透传空排序
- **THEN** 响应按 `sort_order asc, id asc` 返回，与变更前一致

### Requirement: 前端排序交互
系统 SHALL 在 `web/pc/views/roles/List.vue` 对白名单列（ID / 角色名称 / 角色编码 / 状态 / 创建时间）开启 `sortable`，排序变化时 SHALL 发送 `sortBy`（白名单 key）与 `sortOrder`；列 → 白名单 key 映射 MUST 为：`id→id`、角色名称→`role_name`、角色编码→`role_code`、状态→`status`、创建时间→`created_at`；取消排序时 SHOULD 不携带排序参数。

#### Scenario: 点击表头升序
- **GIVEN** 用户点击「角色名称」表头
- **WHEN** 表格触发排序变化
- **THEN** 前端重新请求 `GET /api/perm/roles` 并携带 `sortBy=role_name&sortOrder=asc`

#### Scenario: 再次点击切换降序
- **GIVEN** 当前已按 `role_name asc` 排序
- **WHEN** 用户再次点击「角色名称」表头
- **THEN** 前端请求携带 `sortBy=role_name&sortOrder=desc`

#### Scenario: 取消排序恢复默认
- **GIVEN** 当前已按某列排序
- **WHEN** 用户第三次点击该表头（取消排序）
- **THEN** 前端请求不再携带 `sortBy`/`sortOrder`，服务端返回默认 `sort_order asc, id asc`

## 职责边界

- **api-proto**：`permission.proto` 定义 `ListRolesRequest.sort`（`common.v1.SortField` 已存在，不改 common.proto）。Proto 变更由全局 Claude 执行并 `make ci` 生成。
- **permission-service（rpc）**：白名单/方向校验（单层统一校验）、错误经响应 `Base`（`NewBaseRespWithError`）返回、透传排序至模型。缓存策略不变（ListRoles 无缓存）。
- **permission-service（api）**：`sortBy`/`sortOrder` 透传、`ToError(grpcResp.Base)` 转换业务错误为统一响应。
- **web/pc**：`roles/List.vue` 列排序交互 + `getRoles` 参数透传；排序交互为展示层职责，权威校验在后端。
