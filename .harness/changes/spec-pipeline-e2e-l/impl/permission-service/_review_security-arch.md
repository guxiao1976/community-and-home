# Code Review — 权限服务（安全架构视角）

**审查时间**: 2026-06-19 09:45  
**审查维度**: 架构一致性(#1)、安全性(#5)、变更完整性(#8)

## 摘要
- 🔴 CRITICAL: 0 / 🟡 WARNING: 2 / 🔵 NOTE: 3

## 发现

### 🔴 CRITICAL
无 CRITICAL 问题。

### 🟡 WARNING

| # | 文件:行号 | 维度 | 问题 | 建议 |
|---|----------|------|------|------|
| W1 | rpc/internal/logic/permission/checkpermissionlogic.go:89 | 安全性 | `Expire` TTL 使用 `time.Duration(cacheTTL)*time.Second`，但 CHANGELOG.md 第 163 行记录"Expire 用了纳秒值但 go-redis v9 期望 time.Duration，TTL 可能未正确设置" — 需验证实际运行时缓存是否生效 | 运行时验证 Redis TTL：`redis-cli TTL perm:user:100`，确认是否为 1800 秒。如失效则检查 go-redis 版本兼容性 |
| W2 | model/permission.go:97 | 架构一致性 | `FindList` 使用 `fmt.Sprintf("...limit %d offset %d", pageSize, offset)` 拼接 LIMIT/OFFSET，虽然这两个参数来自 int64 不会造成 SQL 注入，但不符合"所有参数化"的最佳实践 | 改为参数化：`fmt.Sprintf("...limit ? offset ?", ...)` + `args = append(args, pageSize, offset)` |

### 🔵 NOTE

| # | 文件:行号 | 建议 |
|---|----------|------|
| N1 | CHANGELOG.md:163-165 | 已知问题"GetDataScopes 缓存 write-only，未真正加速读取"属于性能优化项，建议在后续迭代中补充缓存读取逻辑，真正减轻 DB 负担 |
| N2 | CHANGELOG.md:165 | 已知问题"UpdateRole 用 KEYS * 全量扫描"在当前未找到实现代码（可能在其他 Logic 文件中），但 design.md 第 133 行确认了该设计。建议在大规模部署前改为 SCAN 或精确失效（使用 SET 记录需失效的 key） |
| N3 | rpc/internal/logic/permission/checkpermissionlogic.go:76 | 权限匹配逻辑 `code := fmt.Sprintf("%s:%s", in.Action, p.Path.String)`，注意 `p.Path.String` 是直接访问 `sql.NullString` 的字段，应使用 `p.Path.Valid` 检查。当前代码未检查 Valid，若 Path 为 NULL 会拼接空字符串导致误匹配 |

---

## 架构一致性检查

✅ **Proto 规范**  
- 服务内无 Proto 文件，符合全局 Proto 管理规范（api-proto/ 统一管理）
- 无 api-proto/ 目录修改，符合子 Claude 禁止修改 Proto 的约束

✅ **gRPC 通信**  
- 代码中无跨服务 DB 导入（已验证：`grep` 未发现其他服务 model 导入）
- 服务定位为"纯叶子节点"（design.md 第 202 行），无出站 gRPC 调用，符合服务隔离原则

✅ **Snowflake ID 序列化**  
- 本次变更仅涉及测试代码和 CHANGELOG，未修改业务代码的 ID 字段
- 已验证历史代码：types.go 中 ID 字段正确标注 `json:"id,string"`（QA 报告第 33 行确认）

✅ **错误码规范**  
- 无新增错误码，现有错误码已在 design.md 定义（060001-060006）

---

## 安全性检查

✅ **密钥管理**  
- api/etc/perm-api.yaml 第 7 行：`AccessSecret: ${JWT_ACCESS_SECRET}` 使用环境变量占位符
- rpc/etc/permissionservice.yaml 第 11 行：`DataSource: ${MYSQL_USER}:${MYSQL_PASSWORD}@tcp(...)` 使用环境变量
- api/perm.go 第 24 行、rpc/permissionservice.go 第 22 行：正确使用 `configx.MustLoad` 展开环境变量
- 符合全局规范（项目编码规范.md §7）

✅ **SQL 参数化**  
- model/permission.go: 所有 INSERT/UPDATE/DELETE 使用 `?` 占位符（第 39/52/103/109 行）
- model/rel.go: 所有 DML 语句使用 `?` 占位符（第 39/47/52/114/122/130/143/167/187 行）
- FindByIds 使用 `strings.Join(placeholders, ",")` 动态构建 IN 子句，但参数通过 `args...` 传递，**符合参数化要求**
- ⚠️ **唯一例外**：model/permission.go 第 97 行 LIMIT/OFFSET 直接拼接数值（见 W2）

✅ **输入校验**  
- CheckPermission 逻辑中 userId/action/apiPath 直接用于 Redis key 和 DB 查询，但：
  - Redis key 使用 `fmt.Sprintf("perm:user:%d", in.UserId)` — userId 是 int64，无注入风险
  - DB 查询全部参数化，无 SQL 注入风险
- 未发现用户输入直接拼接到 SQL/命令的情况

✅ **敏感信息日志**  
- 代码中未发现打印用户密码、Token、密钥的日志
- CheckPermission 返回 `allowed: true/false`，未泄露权限详情

✅ **Redis 操作安全**  
- 未使用 `KEYS *` 全量扫描（在当前变更代码中未发现，CHANGELOG 提到的 UpdateRole KEYS * 不在本次变更范围）
- 缓存 key 格式固定（`perm:user:{userId}`、`perm:scopes:{userId}:{scopeType}`），无命令注入风险

---

## 变更完整性检查

✅ **CHANGELOG 更新**  
- CHANGELOG.md 新增 2026-06-19 条目（第 52-115 行）
- 记录了做了什么、为什么、测试结果、影响范围，符合模板要求

✅ **影响范围声明**  
- Proto: 无变更 ✓
- 调用方: 无影响 ✓
- 数据库: 无 Migration ✓
- 测试覆盖: 从 0% 提升到核心路径覆盖 ✓

✅ **设计文档一致性**  
- 新增测试覆盖 design.md 中描述的核心流程（第 100-116 行 CheckPermission 流程、第 140-149 行 GetDataScopes 流程）
- 测试验证了 design.md 第 32-33 行的"系统角色天然拥有全部权限"设计决策

❌ **Proto CHANGELOG**  
- N/A（本次变更无 Proto 修改）

❌ **Migration**  
- N/A（本次变更无数据库 Schema 修改）

---

## 记忆遵守检查

本视角不负责记忆遵守检查（由规范工程视角负责）。

---

## 测试覆盖评估（安全视角）

✅ **关键安全场景已覆盖**  
- checkpermissionlogic_test.go:141-166 — 系统角色直接放行（防止权限绕过）
- checkpermissionlogic_test.go:211-249 — 权限拒绝场景（防止未授权访问）
- checkpermissionlogic_test.go:273-294 — 无角色拒绝（默认拒绝原则）

✅ **Model 层安全约束已验证**  
- permission_test.go:219-239 — 系统角色不可删除（is_system = 0 条件保护）
- rel_test.go:45-76 — INSERT IGNORE 幂等性（防止重复授权导致数据不一致）
- rel_test.go:259-279 — 级联删除（防止孤立权限数据）

⚠️ **未覆盖的安全场景**（后续补充）  
- 并发场景：多个请求同时修改同一角色的权限时，缓存一致性是否有竞态
- 异常场景：Redis 故障时的降级策略（当前代码 Redis 失败会 fallback 到 DB，但未有显式测试）
- 恶意输入：超长 userId/apiPath/action（虽然有类型限制，但未显式测试边界）

---

## 安全架构决策验证

✅ **RBAC + 数据范围双层模型**  
- 测试覆盖了权限检查（CheckPermission）
- 数据范围查询（GetDataScopes）测试在 CHANGELOG 第 104 行标记为"未覆盖场景"，建议后续补充

✅ **Redis 缓存策略**  
- checkpermissionlogic_test.go:251-271 — 缓存命中场景已验证
- checkpermissionlogic.go:33-36 — 缓存命中直接返回，减少 DB 负担
- ⚠️ TTL 设置存在疑问（见 W1）

✅ **系统角色特权设计**  
- checkpermissionlogic.go:44-49 — 系统角色跳过权限匹配
- 测试已覆盖（checkpermissionlogic_test.go:141-166）
- 符合 design.md 第 32 行设计决策

---

## 依赖安全性

✅ **新增测试依赖**  
- `github.com/DATA-DOG/go-sqlmock@v1.5.2` — 成熟的 SQL mock 库，无已知漏洞
- `github.com/alicebob/miniredis/v2@v2.38.0` — 内存 Redis 模拟，仅测试使用
- `github.com/stretchr/testify@v1.11.1` — 官方推荐的测试框架

所有依赖仅用于测试，不影响生产环境安全。

---

VERDICT: PASS

---

**理由**: 本次变更为纯测试代码补充，无架构违反、无安全漏洞、变更记录完整。两个 WARNING 均为代码质量改进建议，不阻塞合并。建议在后续迭代中：
1. 验证 Redis TTL 实际生效情况（W1）
2. 统一 SQL 参数化风格（W2）
3. 补充 GetDataScopes、并发场景、异常降级的测试覆盖
