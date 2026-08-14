# Tasks: Role List Sort（角色列表排序）

> **对执行 Agent 的指令**: 每个 Task 独立可测，按 TDD 执行（先写测试→看失败→写实现→看通过）。精确到文件路径。
> 依赖顺序：Proto → model → rpc 校验 → rpc 集成 → api → 前端。
> 说明：Task 1.1 的 `FindList` 签名变更会让 rpc 包在 Task 2.2 恢复前编译失败（属签名变更的正常中间态）。每个 Task 以 `go build ./model/...` 或 `go build ./rpc/...` 等**包级**编译验证。
> 相关设计见 `design.md`。

## 全局 / Proto（由全局 Claude 执行，不分发子 Claude）

### Task 0.1: 定义 ListRolesRequest.sort 字段
- **修改**: `api-proto/api/permission/v1/permission.proto`
- [ ] `ListRolesRequest` 新增字段 `optional common.v1.SortField sort = 3;`（`common.v1.SortField` 已存在，不改 common.proto）
- [ ] 确认 `permission.proto` 已 import `common/v1/common.proto`（已有 BaseResp/PageResponse 引用，通常已 import）
- [ ] 确认无新增 int64 字段（SortField 为 string 字段，无需 `[jstype=JS_STRING]`）
- [ ] 更新 `api-proto/CHANGELOG.md`（新增 ListRoles 排序入参，兼容变更）
- SEE: [[grpc-only-comms]]、[[proto-jstype]]

### Task 0.2: 生成代码 + CI
- [ ] `cd api-proto && make ci`（lint + breaking-check + generate）
- [ ] `make lint` → 0 errors
- [ ] `make breaking-check` → 无破坏性变更（本变更为新增 optional 字段，应为 PASS）
- [ ] 确认生成的 `gen/go/permission/v1/permission.pb.go` 含 `Sort` 字段
- [ ] 确认 `git status` 中 go.mod/go.sum 无意外变更

## permission-service（model）

### Task 1.1: FindList 签名扩展 + ORDER BY 组装（含 tiebreaker）
- **修改**: `services/permission-service/model/permission.go`
- **修改**: `services/permission-service/model/permission_test.go`
- **修改**: `services/permission-service/rpc/internal/logic/permission/listroleslogic.go`（FindList 调用点，暂传空排序串）
- **修改**: `services/permission-service/rpc/internal/logic/permission/assignrolelogic_test.go`（MockRoleModel.FindList 方法签名）
- **修改**: `services/permission-service/rpc/internal/logic/permission/listroleslogic_test.go`（现有 `.On("FindList", ...)` 断言签名同步）
- [ ] **RED**: 在 `permission_test.go` 新增 `TestSysRoleModel_FindList_WithSortField`（期望 SQL `order by role_name desc, id asc`，用 sqlmock 断言 `ExpectQuery`）与 `TestSysRoleModel_FindList_DefaultSort`（空排序 → `order by sort_order asc, id asc`）→ 看到 FAIL
- [ ] **确认 RED**: `cd services/permission-service && go test ./model/ -run FindList` → 看到 FAIL
- [ ] **GREEN**: 
  - 接口签名：`FindList(ctx, status *int64, page, pageSize int64, sortField, sortOrder string) ([]*SysRole, int64, error)`
  - 新增包内纯函数 `orderByClause(sortField, sortOrder string) string`：白名单字面量二次防御（非白名单回落 `sort_order`）、方向仅 `asc`/`desc`（非 `desc` 回落 `asc`）、空字段默认 `sort_order`、恒追加 `, id asc`
  - `FindList` 组装 `fmt.Sprintf("select * from %s %s %s limit %d offset %d", m.table, where, orderByClause(sortField, sortOrder), pageSize, offset)`
  - 同步 rpc 侧：MockRoleModel.FindList 方法签名 +2 参数；listroleslogic.go 调用点补 `""`, `""`；listroleslogic_test.go 所有 `On("FindList", ...)` 断言补 2 个 `""` 参数
- [ ] **确认 GREEN**: `go test ./model/ -run FindList` → PASS
- [ ] **REFACTOR**: 提取白名单 `map[string]struct{}` 常量，清理魔法字符串
- ⚠️ 安全：`ORDER BY` 字段仅拼接白名单字面量，严禁拼接 `sortField` 原始值；方向仅二值分支。SEE: design.md「安全考虑 §1」

## permission-service（rpc）

### Task 2.1: 排序校验纯函数（白名单 + 方向规范化）
- **创建**: `services/permission-service/rpc/internal/logic/permission/sort.go`
- **创建**: `services/permission-service/rpc/internal/logic/permission/sort_test.go`
- [ ] **RED**: table-driven 测试覆盖：合法字段+desc / 非法字段 / 字段大小写不敏感（`ROLE_CODE`→`role_code`）/ 空方向默认 asc / 非法方向 / 空字段+非空方向不报错 / 注入载荷（`role_name; drop table sys_role`）被拒
- [ ] **确认 RED**: `go test ./rpc/internal/logic/permission/ -run Sort` → 看到 FAIL
- [ ] **GREEN**: 实现纯函数（如 `validateSort(field, order string) (sortField, sortOrder string, err error)`）：
  - 字段非空 → 转小写比对白名单 `{id, role_code, role_name, sort_order, status, created_at, updated_at}`；未命中 → `errx.NewCodeError(errx.CodeInvalidParam, "非法排序字段: "+field)`
  - 字段为空 → 视为未指定排序（返回空，不报错，即使带方向）
  - 方向转小写；空默认 `asc`；非 `asc`/`desc` → `errx.NewCodeError(errx.CodeInvalidParam, "非法排序方向: "+order)`
  - 返回规范化后的 field/order（大小写已归一）
- [ ] **确认 GREEN**: `go test ./rpc/internal/logic/permission/ -run Sort` → PASS
- [ ] **REFACTOR**: 白名单与 model 层共享语义（各自独立定义，注释标注两处需同步）；错误消息指明非法字段/方向
- SEE: [[error-code-literal-bypasses-qa-gate]]（用 `errx.CodeInvalidParam` 常量，禁止裸数字）、[[tdd-red-evidence-requires-fail-excerpt]]

### Task 2.2: ListRolesLogic 集成校验 + 透传（TDD 收尾）
- **修改**: `services/permission-service/rpc/internal/logic/permission/listroleslogic.go`
- **修改**: `services/permission-service/rpc/internal/logic/permission/listroleslogic_test.go`
- [ ] **RED**: 新增测试：`TestListRoles_InvalidSortField`（`sort.field="evil"` → `resp.Base.Code==99400`、`Roles` 为空、返回 nil err）、`TestListRoles_InvalidSortOrder`、`TestListRoles_SortPassthrough`（合法排序 → mock `.On("FindList", ..., "role_name", "desc")` 收到规范化参数）
- [ ] **确认 RED**: `go test ./rpc/internal/logic/permission/ -run ListRoles` → 看到 FAIL
- [ ] **GREEN**: 在 `ListRolesLogic.ListRoles` 中：
  - `if in.Sort != nil { field, order, err := validateSort(in.Sort.Field, in.Sort.Order); if err != nil { return &permissionv1.ListRolesResponse{Base: responsex.NewBaseRespWithError(int32(errx.CodeInvalidParam), err.Error())}, nil } }`
  - 将 field/order 透传 `FindList(l.ctx, status, page, pageSize, field, order)`（sort 未携带时传 `""`, `""`）
- [ ] **确认 GREEN**: `go test ./rpc/internal/logic/permission/ -run ListRoles` → PASS
- [ ] **REFACTOR**: 校验失败路径不执行查询（断言 mock 未被调用）；确认 `go build ./...` rpc 包恢复编译
- SEE: [[rpc-callback-must-check-response-base]]（错误走 Base 而非 Go error）、[[error-code-collision-and-namespace-alignment]]

## permission-service（api）

### Task 3.1: ListRolesReq 新增 sortBy/sortOrder
- **修改**: `services/permission-service/api/internal/types/types.go`
- **修改**: `services/permission-service/api/internal/types/types_test.go`（若存在类型断言测试）
- [ ] `ListRolesReq` 新增：
  - `SortBy *string form:"sortBy,optional"`
  - `SortOrder *string form:"sortOrder,optional"`
- [ ] 运行 `cd services/permission-service/api && go build ./...` → 编译通过
- SEE: [[api-required-field-marked-optional]]（sortBy/sortOrder 均为 optional，正确）

### Task 3.2: ListRoles 透传 sort + ToError(grpcResp.Base)（修复性前置）
- **修改**: `services/permission-service/api/internal/logic/perm/listroleslogic.go`
- **创建**: `services/permission-service/api/internal/logic/perm/listroleslogic_test.go`
- [ ] **RED**: 用最小 mock（struct 内嵌 `permission.PermissionService` 接口（`github.com/guxiao1976/community-permission/rpc/permission`），仅覆写 `ListRoles` 方法）新建测试：
  - `TestListRoles_SortPassThrough`：`req.SortBy="role_name"`、`req.SortOrder="desc"` → 断言 grpcReq.Sort.Field/Order 透传
  - `TestListRoles_BaseErrorToGoError`：mock 返回 `Base.Code=99400` → `ListRoles` 返回 Go error（非 nil）
- [ ] **确认 RED**: `go test ./api/internal/logic/perm/ -run ListRoles` → 看到 FAIL
- [ ] **GREEN**:
  - 构建 grpcReq 时：`if req.SortBy != nil { sortOrder := ""; if req.SortOrder != nil { sortOrder = *req.SortOrder }; grpcReq.Sort = &commonv1.SortField{Field: *req.SortBy, Order: sortOrder} }`（SortOrder 为空串由 RPC 层默认 asc；注意防 nil 解引用）
  - gRPC 调用后立即：`if err := responsex.ToError(grpcResp.Base); err != nil { return nil, err }`
- [ ] **确认 GREEN**: `go test ./api/internal/logic/perm/ -run ListRoles` → PASS
- [ ] **REFACTOR**: 确认既有 `GET /api/perm/roles` 无排序参数时行为不变
- SEE: [[rpc-callback-must-check-response-base]]（当前 API 层未检查 Base，本任务补齐）、[[api-response-single-wrap]]

## 前端（web/pc）

### Task 4.1: getRoles 参数类型扩展
- **修改**: `web/pc/src/api/identity.ts`
- [ ] `getRoles` 参数类型从 `params?: PaginationParams` 扩展为支持 `sortBy`/`sortOrder`（camelCase），如定义局部类型 `RolesQuery extends PaginationParams { sortBy?: string; sortOrder?: string }`
- [ ] 运行 `cd web/pc && npm run lint` → 类型检查通过
- SEE: [[web-common-type-reuse-no-redefine]]（复用共享 `PaginationParams`，仅扩展角色查询特有字段，不上移全局）

### Task 4.2: roles/List.vue 白名单列 sortable + sort-change 交互
- **修改**: `web/pc/src/views/roles/List.vue`
- **创建**: `web/pc/tests/unit/views/roles/List.spec.ts`
- [ ] 列 → 白名单 key 映射（`SORT_KEY_MAP`）：`id→id`、`角色名称(name)→role_name`、`角色编码(code)→role_code`、`状态(status)→status`、`创建时间(created_at)→created_at`；`sortOrder` 列不入列
- [ ] `el-table` 上 `@sort-change="handleSortChange"`；上述列加 `sortable`（状态列需补 `prop="status"` 以便触发）
- [ ] `handleSortChange({ prop, order })`：`order='ascending'`→`sortOrder='asc'`、`'descending'`→`'desc'`、`null`（取消）→ 不携带排序参数；`loadRoles` 时按映射发送 `sortBy`（白名单 key）+ `sortOrder`，重载后排序保持
- [ ] **测试**: `List.spec.ts` 断言点击「角色名称」升序 → mock `getRoles` 收到 `{ sortBy: 'role_name', sortOrder: 'asc' }`；再点降序 → `desc`；三连击取消 → 无 sortBy/sortOrder
- [ ] 运行 `cd web/pc && npm run test:unit` → PASS；`npm run build` → 通过
- ⚠️ 白名单权威在后端，前端仅做映射与发送；若误发非白名单 key，后端会拒绝（行为可观测）。SEE: [[frontend-business-rule-hardcode]]

## 交付门禁（变更收尾，Owner 执行）

### Task 5.1: 门禁 + 回归验证
- [ ] `cd services/permission-service && go build ./... && go test ./...` → PASS
- [ ] `bash .harness/skills/qa/scripts/harness-checks.sh --service permission-service` → 无 FAIL
- [ ] `cd web/pc && npm run build && npm run test:unit` → PASS
- [ ] 手工回归：`GET /api/perm/roles`（无参 → `sort_order asc, id asc`）、`?sortBy=role_name&sortOrder=desc`、`?sortBy=evil`（→ code=99400，data=null）
- SEE: [[pre-commit-checks]]、[[change-verification-checklist]]

---

## Self-Review 记录

- **占位符扫描**: 无 TBD/TODO，所有 Task 精确到文件路径 ✅
- **TDD 覆盖**: Task 1.1/2.1/2.2/3.2/4.2 均含 RED→GREEN→REFACTOR ✅
- **依赖顺序**: Proto(0.1-0.2) → model(1.1) → rpc(2.1-2.2) → api(3.1-3.2) → 前端(4.1-4.2) → 门禁(5.1) ✅
- **记忆引用检查**: 12 个相关记忆，12 个已引用（[[grpc-only-comms]]/[[proto-jstype]]/[[error-code-literal-bypasses-qa-gate]]/[[tdd-red-evidence-requires-fail-excerpt]]/[[rpc-callback-must-check-response-base]]/[[error-code-collision-and-namespace-alignment]]/[[api-required-field-marked-optional]]/[[api-response-single-wrap]]/[[web-common-type-reuse-no-redefine]]/[[frontend-business-rule-hardcode]]/[[pre-commit-checks]]/[[change-verification-checklist]]），0 遗漏；Task 1.1 的 SQL 注入风险未引用不存在的记忆 slug，改为引用 design.md「安全考虑 §1」
