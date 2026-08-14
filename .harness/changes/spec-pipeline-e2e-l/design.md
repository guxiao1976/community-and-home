# Design: Role List Sort（角色列表排序）

> 对应 Change：`spec-pipeline-e2e-l`
> 依据：`.harness/changes/spec-pipeline-e2e-l/proposal.md` + `specs/role-list-sort/spec.md`

## 服务归属决策

| 功能 | 归属服务 | 理由 |
|------|---------|------|
| `ListRolesRequest` 增 `sort` 入参 | api-proto（全局） | Proto 单一来源，`common.v1.SortField` 已存在、common.proto 不变 |
| 白名单/方向校验 + 透传排序 | permission-service（rpc） | 谁处理该业务领域谁校验（单层权威校验） |
| `FindList` 排序组装（ORDER BY） | permission-service（model） | 谁拥有 `sys_role` 表数据谁生成查询 |
| REST `sortBy`/`sortOrder` 透传 + Base 检查 | permission-service（api） | REST 网关层职责，转 gRPC 并呈现统一错误 |
| 列排序交互 | web/pc | 展示层交互，权威校验在后端 |

> 单一服务职责：本次变更只涉及 permission-service 一个业务服务，无跨服务接口新契约（仅复用 `common.v1.SortField`）。

## 数据模型

无新增表、无字段变更、无 Migration。排序白名单字段均为 `sys_role` 既有列名：

| 白名单 key | DB 列 | 备注 |
|-----------|-------|------|
| `id` | `id` | 恒追加为 `ORDER BY` 末级 tiebreaker |
| `role_code` | `role_code` | |
| `role_name` | `role_name` | |
| `sort_order` | `sort_order` | 默认首键 |
| `status` | `status` | |
| `created_at` | `created_at` | |
| `updated_at` | `updated_at` | |

### 数据访问层契约变更（`SysRoleModel.FindList`）

```
现状：  FindList(ctx, status *int64, page, pageSize int64) ([]*SysRole, int64, error)
目标：  FindList(ctx, status *int64, page, pageSize int64, sortField, sortOrder string) ([]*SysRole, int64, error)
```

- `sortField`/`sortOrder` 均为**已校验**值（RPC 层权威校验），可为空。
- 空 `sortField` → 默认首键 `sort_order`；方向仅 `asc`/`desc`，空默认 `asc`。
- 恒追加 `id asc` 平局决胜：`ORDER BY {field} {dir}, id asc`。
- 二次防御：`orderByClause(sortField, sortOrder)` 内建白名单 map，非白名单字段回落 `sort_order`、非 `desc` 方向回落 `asc`，**杜绝任何原始输入拼入 SQL**（REQ-6 注入安全，纵深防御，即使 RPC 校验被绕过也安全）。

## 接口设计

### RPC `PermissionService.ListRoles`（permission-service/rpc）

- **入参新增**：`optional common.v1.SortField sort = 3`（`field`/`order` 均为 string，见 common.proto）
- **校验（`validateSort`，纯函数，单层权威）**：
  - `sort.field` 非空时统一小写后比对白名单 `{id, role_code, role_name, sort_order, status, created_at, updated_at}`；未命中 → 参数错误，`msg` 指明非法字段名。
  - `sort.field` 为空 → 视为未指定排序，**跳过方向校验不报错**（REQ-4「空字段+方向不报错」）。
  - `sort.order` 统一小写规范化；空值默认 `asc`；非 `asc`/`desc` → 参数错误，`msg` 指明非法方向值。
- **错误呈现**：校验失败以 `responsex.NewBaseRespWithError(int32(errx.CodeInvalidParam), msg)` 写入响应 `Base`，**返回 `nil` error**（项目 gRPC 层约定：业务错误走 Base，见 `common/pkg/responsex/grpc.go` 与 `createrolelogic.go` 同款模式）。不执行任何查询。
- **透传**：校验通过后以 `sortField`/`sortOrder` 调用 `RoleModel.FindList`。
- 错误码使用 `errx.CodeInvalidParam`（99400）常量，禁止裸数字。SEE: [[error-code-literal-bypasses-qa-gate]]、[[error-code-collision-and-namespace-alignment]]

### REST `GET /api/perm/roles`（permission-service/api）

- **入参新增**（`types.ListRolesReq`，camelCase query 参数）：
  - `SortBy *string form:"sortBy,optional"`
  - `SortOrder *string form:"sortOrder,optional"`
- **透传**：`req.SortBy`/`req.SortOrder` → `grpcReq.Sort = &commonv1.SortField{Field, Order}`；`sortOrder` 为空时透传空串，由 RPC 层默认 `asc`。
- **Base 检查（修复性前置）**：gRPC 调用后必须 `responsex.ToError(grpcResp.Base)`，非成功则 `return nil, err`。当前 `listroleslogic.go` 未检查 Base，业务校验错误会被静默吞掉。SEE: [[rpc-callback-must-check-response-base]]
- **错误呈现**：handler 经 `responsex.Response(w, nil, err)` 输出统一响应体 `{code:99400, msg, data:null}`（HTTP 200 + 响应体 code 语义 Bad Request）。

### 业务流程

```
web/pc List.vue (sortable 列 + sort-change)
  → GET /api/perm/roles?sortBy=role_name&sortOrder=desc
    → api ListRolesLogic: 透传 sort + ToError(grpcResp.Base)
      → gRPC ListRoles
        → rpc ListRolesLogic: validateSort（白名单/方向，失败写 Base=99400）
          → RoleModel.FindList(status, page, pageSize, sortField, sortOrder)
            → orderByClause: ORDER BY role_name desc, id asc
```

## Proto 变更

| 文件 | 变更类型 | 说明 |
|------|:---:|------|
| `api-proto/api/permission/v1/permission.proto` | 兼容（新增 optional 字段） | `ListRolesRequest` 新增 `optional common.v1.SortField sort = 3`。`common/v1/common.proto` 不变（`SortField` 已存在、无消费方）。int64 字段无新增，不涉 `[jstype=JS_STRING]`。 |

- 非破坏性：仅新增消息字段，不删除/不改字段号，既有调用方二进制兼容。
- 执行：全局 Claude `cd api-proto && make ci`（lint + breaking-check + generate），并更新 `api-proto/CHANGELOG.md`。

## 安全考虑

1. **SQL 注入（ORDER BY 拼接）**：RPC 白名单前置拦截 + model `orderByClause` 内建白名单 map 二次防御；方向仅 `asc`/`desc` 字面量二选一。用户原始输入（含 `; drop ...` 载荷）不进入任何 SQL。覆盖 REQ-6。
2. **业务错误不被静默吞掉**：API 层必须 `ToError(grpcResp.Base)` 转换，否则校验失败对前端不可见。覆盖 REQ-8。SEE: [[rpc-callback-must-check-response-base]]
3. **默认排序行为回归**：空 `sort`/空 `field` 时首键保持 `sort_order asc`，恒追加 `id asc`，既有调用方排序语义不变。覆盖 REQ-4/REQ-5。
4. **错误码语义**：99400 语义 Bad Request，与既有 `errx.CodeInvalidParam` 一致，禁止同码异义。SEE: [[error-code-collision-and-namespace-alignment]]

## 记忆引用（设计阶段预防性注入，Step 1.5 产出）

| 记忆 | 适用章节 | 设计决策 |
|------|---------|---------|
| [[rpc-callback-must-check-response-base]] | 接口设计（REST） | API 层必须 `ToError(grpcResp.Base)`，否则业务校验错误被静默吞掉 |
| [[error-code-literal-bypasses-qa-gate]] | 接口设计（RPC） | 用 `errx.CodeInvalidParam`(99400) 常量，禁止裸数字 |
| [[error-code-collision-and-namespace-alignment]] | 安全考虑 | 99400 语义唯一（Bad Request），不与既有错误码冲突 |
| [[grpc-only-comms]] | 接口设计 | API→RPC 仅 gRPC 调用，不直连 DB |
| [[api-response-single-wrap]] | 接口设计（REST） | 统一响应 `{code,msg,data}`，校验失败 data=null |
| [[frontend-business-rule-hardcode]] | 前端 | 白名单权威在后端；前端仅做列→key 映射与发送，不做校验 |
| [[tdd-red-evidence-requires-fail-excerpt]] | Tasks（TDD） | RPC/model 逻辑测试必须含 RED 证据（FAIL 输出摘录） |

> 记忆注入报告：匹配 9 个，注入 7 个，不适用 2 个（[[proto-jstype]] 本次无新 int64 字段；[[web-common-type-reuse-no-redefine]] 前端仅扩展查询参数类型，不复用全局分页类型）。

## 范围外观察（不修，仅记录）

- 前端 `List.vue` 目前发 `page_size`（snake_case），REST 绑定为 `pageSize`（camelCase）——存在既有字段名不匹配。**本变更不处理**；新参数必须严格按 REST 契约发送 `sortBy`/`sortOrder`（camelCase）。如后续确认分页参数失效，另开变更。
