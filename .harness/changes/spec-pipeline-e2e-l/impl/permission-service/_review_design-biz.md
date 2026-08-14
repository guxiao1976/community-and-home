# Code Review — 权限服务（设计业务视角）

**审查时间**: 2026-06-19 09:30  
**审查维度**: 设计一致性(#2)、代码质量(#4)、Migration(#8部分)

## 摘要
- 🔴 CRITICAL: 1 / 🟡 WARNING: 2 / 🔵 NOTE: 0

## 发现

### 🔴 CRITICAL
| # | 文件:行号 | 维度 | 问题 | 修复建议 |
|---|----------|------|------|---------|
| C1 | rpc/internal/logic/permission/checkpermissionlogic.go:76 | 代码质量 | 权限匹配逻辑错误：代码动态拼接 `code = "GET:/api/users"`，但数据库中 `sys_permission.path` 为 `/api/user/v1/GetUser`（Proto 定义），不可能匹配成功。实际权限码应为 `user:read`（`sys_permission.code`），而非 `action:path` 拼接。| 修改第76行为：`code := p.Code`，移除 `fmt.Sprintf` 拼接逻辑。或根据 design.md 明确权限码定义规则：若使用 `action:path` 则数据库初始化时必须预填 `GET:/api/user/v1/GetUser` 格式的权限码；若使用 `code` 字段则调用方传 `user:read` 而非 `GET:/api/users`。两种方案必须选一，当前实现两者混杂导致权限检查失效。|

### 🟡 WARNING
| # | 文件:行号 | 维度 | 问题 | 建议 |
|---|----------|------|------|------|
| W1 | rpc/internal/logic/permission/checkpermissionlogic_test.go:186-188 | 设计一致性 | 测试用例 Mock 数据与 CRITICAL C1 同源问题：Mock 的 Permission 使用 `Path: "/api/users"`，但实际匹配逻辑在第76行拼接 `action:path`，与第78行的 `p.Code == needle` 形成双重判断。测试通过仅因 Mock 数据人为对齐，掩盖了生产逻辑不一致。| 修复 C1 后，同步修改测试用例 Mock 数据，确保与生产数据模型一致（design.md 表明 `sys_permission.code` 为 `user:read`，`path` 为 `/api/user/v1/GetUser`）。|
| W2 | model/rel.go:160-167 | 代码质量 | BatchInsertUserRoles 循环逐条插入，未使用批量 INSERT，在大量角色分配时性能低下（N次 DB 往返）。| 改为单条 SQL：`INSERT IGNORE INTO rel_user_role (user_id, role_id, scope_type, scope_id) VALUES (?,?,?,?), (?,?,?,?), ...` 拼接占位符，一次执行。参考 MySQL 批量插入最佳实践。|

### 🔵 NOTE
（无）

---

## 详细分析

### CRITICAL C1 — 权限匹配逻辑不一致

**根本原因**：design.md 定义权限模型时，`sys_permission` 表同时存在 `code` 和 `path` 字段：
- `code`：权限码，如 `user:read`
- `path`：API 路径，如 `/api/user/v1/GetUser`（Proto RPC 路径）

但 `CheckPermissionLogic` 第76行实现时：
```go
code := fmt.Sprintf("%s:%s", in.Action, p.Path.String)  // 拼接为 "GET:/api/user/v1/GetUser"
```

第78行判断：
```go
if code == needle || p.Code == needle {  // needle = "GET:/api/users"
```

**问题**：
1. `in.ApiPath` 来自调用方（如 API Gateway），传入的是简化路径 `/api/users`
2. `p.Path` 来自数据库，存储的是 Proto 完整路径 `/api/user/v1/GetUser`
3. 两者无法匹配，导致权限检查失效
4. `p.Code == needle` 分支也无法生效，因为 `p.Code` 为 `user:read`，而 `needle` 为 `GET:/api/users`

**验证方法**：
- 查看 `sys_permission` 表初始化数据（如 `docs/sql/permission_init.sql`）
- 查看 API Gateway 或其他服务调用 `CheckPermission` 时传入的 `action` 和 `api_path` 实际值

**修复方案**（二选一）：

**方案A — 使用 code 字段（推荐）**：
- 数据库中 `sys_permission.code` 存储 `user:read`、`user:write` 等语义化权限码
- 调用方传入 `action="user"`, `api_path="read"`
- CheckPermissionLogic 拼接 `needle = "user:read"`，与 `p.Code` 匹配
- 优点：语义清晰，前端可直接用 code 做按钮级权限控制

**方案B — 使用 path 字段**：
- 数据库中 `sys_permission.path` 存储 `GET:/api/users`（简化路径）
- 调用方传入 `action="GET"`, `api_path="/api/users"`
- CheckPermissionLogic 拼接 `needle = "GET:/api/users"`，与拼接后的 `p.Path` 匹配
- 缺点：需要维护 Proto RPC 路径与 REST 路径的映射关系

**当前实现的矛盾**：代码试图同时支持两种方案（第78行 OR 判断），但两边都无法匹配成功。

### WARNING W1 — 测试用例掩盖设计缺陷

`checkpermissionlogic_test.go:186-188` 的 Mock 数据：
```go
mockPerm.On("FindByIds", ...).Return([]*model.SysPermission{
    {Id: 101, Code: "user:read", Path: sql.NullString{String: "/api/users", Valid: true}},
}, nil)
```

这里 `Path: "/api/users"` 与调用方传入的 `ApiPath: "/api/users"` 一致，**人为对齐**使得测试通过，但掩盖了以下问题：
1. 真实数据库中 `sys_permission.path` 为 Proto 路径 `/api/user/v1/GetUser`（design.md 表明是 Proto 定义）
2. 测试环境与生产环境数据模型不一致
3. 测试通过 ≠ 生产逻辑正确

**建议**：修复 C1 后，测试 Mock 数据应反映真实数据库结构：
```go
// 如采用方案A（code 匹配）
{Code: "user:read", Path: sql.NullString{String: "/api/user/v1/GetUser", Valid: true}}
// 调用方传入 Action: "user", ApiPath: "read"

// 如采用方案B（path 匹配）
{Code: "user:read", Path: sql.NullString{String: "GET:/api/users", Valid: true}}
// 调用方传入 Action: "GET", ApiPath: "/api/users"
```

### WARNING W2 — 批量插入性能问题

`BatchInsertUserRoles` 循环逐条插入，每次调用 `ExecCtx`，在为 100 个用户批量分配角色时会产生 100 次 DB 往返（RTT）。

**改进方案**：
```go
func (m *defaultRelUserRoleModel) BatchInsertUserRoles(ctx context.Context, records []*RelUserRole) error {
    if len(records) == 0 {
        return nil
    }
    
    placeholders := make([]string, len(records))
    args := make([]interface{}, 0, len(records)*4)
    
    for i, r := range records {
        placeholders[i] = "(?, ?, ?, ?)"
        args = append(args, r.UserId, r.RoleId, r.ScopeType, r.ScopeId)
    }
    
    query := fmt.Sprintf("insert ignore into %s (user_id, role_id, scope_type, scope_id) values %s",
        m.table, strings.Join(placeholders, ","))
    _, err := m.conn.ExecCtx(ctx, query, args...)
    return err
}
```

**影响评估**：当前实现功能正确（INSERT IGNORE 保证幂等性），但性能为 O(N) 次网络往返，批量优化后降为 O(1)。

---

## 设计一致性检查

| 检查项 | 结果 | 说明 |
|--------|:----:|------|
| 与 design.md 一致性 | ⚠️ 部分不一致 | 权限匹配逻辑（code vs path）未明确，实际实现与设计文档描述不匹配 |
| 数据模型正确性 | ✅ 通过 | Model 层结构与 design.md 定义的 4 张表一致 |
| 业务流程正确性 | ⚠️ 部分不一致 | CheckPermission 核心流程框架正确（缓存→系统角色→权限匹配），但匹配细节有误 |
| 边界条件处理 | ✅ 良好 | 测试覆盖空列表、分页边界、无角色等场景 |
| 错误处理完善性 | ✅ 通过 | DB 查询失败返回 `Allowed: false`，符合 Fail-Safe 原则 |
| 资源泄露 | ✅ 无泄露 | Redis 连接使用 go-redis 连接池，测试中正确使用 miniredis |

---

## 变更完整性检查

| 检查项 | 结果 | 说明 |
|--------|:----:|------|
| CHANGELOG 更新 | ✅ 完整 | 2026-06-19 条目详细记录测试补充、TDD 流程、覆盖场景 |
| Proto CHANGELOG | N/A | 本次变更无 Proto 改动 |
| Migration | N/A | 本次变更无数据库 schema 改动 |
| design.md | ⚠️ 需补充 | 需明确权限匹配规则（code 字段 vs path 字段 vs 拼接逻辑） |

---

## 记忆遵守检查

✅ 所有 `// SEE: [[testing-discipline]]` 引用已验证（共 11 处）  
✅ 测试遵循 testing-discipline.md 要求：Unit Tests 完整覆盖 Model 层和 CheckPermission 核心场景  
✅ 测试命名符合 `Test<Function>_<Condition>_<ExpectedResult>` 规范  
✅ 使用 Table-driven 测试（如 `TestSysRoleModel_FindList_Pagination`）  

**未遗漏 must-follow 记忆**：本次变更为测试补充，不涉及 Proto、gRPC、Migration、密钥管理等 must-follow 场景。

---

VERDICT: **FAIL**

**原因**：存在 1 个 CRITICAL 问题（C1 — 权限匹配逻辑错误），该问题会导致权限检查在生产环境失效，属于核心业务逻辑缺陷，必须修复后重新审查。

**修复路径**：
1. 明确权限模型设计决策（code 字段 vs path 字段）
2. 修改 `CheckPermissionLogic.CheckPermission` 第76-78行逻辑
3. 同步修改测试用例 Mock 数据，反映真实数据库结构
4. 更新 design.md 明确权限匹配规则
5. 验证现有数据库初始化脚本与新逻辑一致
6. 重新运行所有测试，确保通过

**2个 WARNING 可在后续迭代中优化**（不阻塞本轮修复）。
