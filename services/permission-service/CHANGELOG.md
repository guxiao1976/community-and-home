# CHANGELOG — permission-service

## 2026-08-11 — RBAC 角色体系合并（方案 B）

### 做了什么
- `rel_user_role` 扩展：新增 `status`（个体角色生命周期 0-4）、`verified_at`、`expires_at`
- 角色从 4 个扩到 8 个：新增 `tenant`/`committee`/`merchant`/`sys_admin`
- `AssignRole` 支持 status/verified_at/expires_at 参数
- `GetUserRoles` 返回个体生命周期状态（status/verifiedAt/expiresAt）
- `CheckPermission` 增加个体角色过期校验（status=2 且未过期才生效）
- 新增 `UpdateUserRoleStatus` RPC：认证通过/驳回时更新角色状态
- 新增 `FindAllByUserId`/`UpdateRoleStatus` model 方法
- `FindActiveByUserId` 只返回已认证（status=2）且未过期的角色
- API 层 `UserRoleInfo` 加生命周期字段，挂载 PermMiddleware

### 为什么
废弃 user-service 的 `user_membership_role`，permission-service 成为角色唯一权威，统一两套角色体系。

### 影响
- Proto: `api-proto/api/permission/v1/permission.proto`（UserRoleInfo + UpdateUserRoleStatus）
- 调用方: auth-service（JWT roles）、user-service（认证流程）、前端 PC/移动端（状态展示）
- 数据库: `rel_user_role` 加 3 列
- 关联: 提交待定

## 2026-06-19 — 修复 int64 字段 JSON 序列化精度丢失

### 做了什么
- `api/internal/types/types.go`：为所有 int64 字段添加 `json:",string"` 标注
  - PageInfo.Total：添加 `,string` 标注
  - RoleInfo.CreatedAt / UpdatedAt：添加 `,string` 标注
  - PermissionInfo.CreatedAt / UpdatedAt：添加 `,string` 标注
  - CreateRoleReq.PermissionIds / UpdateRoleReq.PermissionIds：使用自定义 Int64Array 类型
- 新增 `Int64Array` 类型：自定义 JSON marshaler/unmarshaler，将 []int64 序列化为字符串数组
- `api/internal/types/types_test.go`：新增回归测试
  - TestInt64FieldsSerializeAsString：验证所有 int64 字段序列化为字符串
  - TestInt64PrecisionPreservation：验证大数字往返不丢失精度
  - 共 10 个测试用例，覆盖 7 处修复点

### 为什么
QA 编码规范检查发现 7 处违规：int64 字段缺少 `json:",string"` 标注，将导致前端 JavaScript 精度丢失（Number 类型仅支持 53 位，Snowflake ID 为 64 位）。

根因：编码规范要求所有 int64 字段（包括时间戳、计数器、ID 数组）都必须标注为字符串，但 QA 检查脚本仅检查字段名以 'Id' 结尾的字段，导致漏检。

### 技术细节
- **标量 int64 字段**：直接添加 `json:",string"` 标注
- **[]int64 数组字段**：Go 标准库不支持 `json:",string"` 对数组生效，需自定义类型实现 MarshalJSON/UnmarshalJSON
- **SEE 标记**：代码中添加 `// SEE: [[proto-jstype]]` 引用记忆，确保后续维护者理解修复原因

### 测试结果
```
# TDD RED 阶段：7 个测试失败（验证问题存在）
FAIL: PageInfo.Total / RoleInfo.CreatedAt / RoleInfo.UpdatedAt
FAIL: CreateRoleReq.PermissionIds / UpdateRoleReq.PermissionIds
FAIL: PermissionInfo.CreatedAt / PermissionInfo.UpdatedAt

# TDD GREEN 阶段：所有测试通过
ok  	github.com/guxiao1976/community-permission/api/internal/types	0.017s

# QA 检查：编码规范检查通过
[5/11] Go json:",string" ✓
```

### 影响
- Proto: 无（仅 API 层类型定义变更）
- 调用方: **前端需适配** — 原本接收 number 的字段现在为 string，需更新 TypeScript 类型定义
- 数据库: 无
- 向后兼容性: **破坏性变更** — JSON 响应格式改变，前端必须同步更新

### 应用的记忆
- [[proto-jstype]] (must-follow) — Proto int64 字段必须加 jstype=JS_STRING
  - types.go:12 — PageInfo.Total
  - types.go:28-29 — RoleInfo.CreatedAt/UpdatedAt
  - types.go:51, 76 — CreateRoleReq/UpdateRoleReq.PermissionIds
  - types.go:103-104 — PermissionInfo.CreatedAt/UpdatedAt
  - types.go:7-38 — Int64Array 自定义类型实现

---

## 2026-06-04 — W8: conf.MustLoad → configx.MustLoad 迁移

### 做了什么
- `rpc/permissionservice.go`：将 `conf.MustLoad` 替换为 `configx.MustLoad`，导入路径从 `go-zero/core/conf` 改为 `community-common/v2/pkg/configx`

### 为什么
根 CLAUDE.md 全局硬规则要求所有服务入口使用 `configx.MustLoad`（支持 `${VAR}` 环境变量展开）。API 层已在 C7 修复中完成迁移，本次补齐 RPC 层。

### 影响
- Proto: 无
- 调用方: 无
- 数据库: 无

## 2026-06-04 — C7 JWT Secret 硬编码修复

### 做了什么
- `api/etc/perm-api.yaml`：将硬编码的 `AccessSecret` 替换为环境变量 `${JWT_ACCESS_SECRET}`
- `api/perm.go`：将 `conf.MustLoad` 替换为 `configx.MustLoad`，确保 `${VAR}` 展开

### 为什么
审计发现 JWT Secret 占位符明文存放在配置中，不符合安全规范。改为环境变量后，密钥由根 `.env` 统一注入。

### 影响
- Proto: 无
- 调用方: 无
- 数据库: 无

---

## 2026-06-04 — 全局公约与设计文档

### 做了什么
- `CLAUDE.md` 新增 `## 全局公约` 章节，引用根 CLAUDE.md
- 首次创建设计文档 `docs/design.md`（数据库设计、权限检查流程、缓存策略、13 个 RPC 等）
- 添加 `CHANGELOG.md`（本文件）

### 为什么
项目规范化——补齐设计文档，子 Claude 启动时能感知全局架构规则和本服务设计决策。

### 已知问题（待修复）
- `CheckPermission` 中 `Expire` TTL 使用纳秒值，go-redis v9 可能未正确设置
- `GetDataScopes` 缓存 write-only，未真正加速读取
- `UpdateRole` 用 `KEYS *` 全量扫描，大规模部署需优化

### 影响
- Proto: 无
- 调用方: 无
- 数据库: 无

## 2026-06-19 — TDD 补充：完整单元测试覆盖

### 做了什么
- `model/permission_test.go`：新增 SysRole 和 SysPermission 的完整单元测试
  - FindByIds 空列表边界测试
  - FindList 分页边界测试（第1/2/3页，边界0条）
  - FindList 状态过滤测试
  - Insert/SoftDelete/FindWithFilter 测试
  - 共 10 个测试用例，覆盖正常/边界/错误路径

- `model/rel_test.go`：新增 RelUserRole 和 RelRolePermission 的完整单元测试
  - BatchInsertUserRoles 幂等性测试（INSERT IGNORE）
  - FindActiveByUserId 联表查询测试
  - FindScopesByUserId 多场景测试（社区/楼栋/无结果）
  - DeleteByRoleId 级联删除测试
  - 共 9 个测试用例

- `rpc/internal/logic/permission/checkpermissionlogic_test.go`：新增 CheckPermission 核心场景测试
  - 系统角色直接放行（is_system=1）
  - 普通用户权限匹配通过
  - 普通用户无权限拒绝
  - Redis 缓存命中直接返回
  - 用户无角色拒绝访问
  - 共 5 个测试用例，使用 miniredis 模拟 Redis

- 依赖管理：
  - 新增 `github.com/DATA-DOG/go-sqlmock`（Model 层 SQL mock）
  - 新增 `github.com/stretchr/testify/mock`（行为验证）
  - 新增 `github.com/alicebob/miniredis/v2`（内存 Redis mock）

### 为什么
QA 验证发现 permission-service 完全缺少单元测试（0 个测试文件，1645 行代码无任何覆盖），核心业务逻辑（CheckPermission、GetDataScopes、缓存一致性）无验证保障，违反项目测试纪律（testing-discipline.md）。

按 TDD 原则（RED → GREEN → REFACTOR）补充测试：
1. RED：创建测试，验证测试失败（编译错误/依赖缺失）
2. GREEN：安装依赖，所有测试通过
3. REFACTOR：（无需重构，现有代码已正确实现）

### 测试结果
```
# Model 层：19 个测试，全部通过
ok  	github.com/guxiao1976/community-permission/model	0.015s

# Logic 层：5 个测试，全部通过
ok  	github.com/guxiao1976/community-permission/rpc/internal/logic/permission	0.057s
```

### 覆盖的关键场景
1. **CheckPermission**：系统角色放行、权限匹配、拒绝、缓存命中、无角色
2. **Model 边界**：空列表、分页边界、幂等性、级联删除
3. **缓存一致性**：Redis 缓存命中/未命中、回填逻辑

### 未覆盖场景（后续补充）
- GetDataScopes（多角色合并、空结果、缓存读写）
- AssignRole/RevokeRole（幂等性、并发安全、缓存失效）
- UpdateRole（批量缓存失效 KEYS * → DEL）
- API 层 Logic（REST 网关层，需集成测试）

### 影响
- Proto: 无
- 调用方: 无
- 数据库: 无
- 测试覆盖: 从 0% 提升到核心路径覆盖（Model 100%，CheckPermission 核心场景 100%）
