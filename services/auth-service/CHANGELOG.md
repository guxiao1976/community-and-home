# CHANGELOG — auth-service

## 2026-08-13 — 端准入判定（access-control）

### 做了什么
- 新增 `PermissionServiceRpc` gRPC 客户端（`config.go` + `servicecontext.go` + `etc/authservice.yaml`），auth 首次依赖 permission-service
- 新增 `platform.go`：`classifyDeviceType`（web/admin→pc、ios/android/miniapp→mobile、空/未知→空）+ `checkPlatformAccess`（读 permission `GetUserRoles` 取 `platforms` 做端判定，fail-open：空 platforms/未知端/零角色/GetUserRoles 失败均放行）
- `Login`/`LoginSms`/`Register`/`RefreshToken` 在签发 Token 前挂载端判定，不匹配返回 `50007`（该账号为移动端用户，请使用移动端 APP）；refresh 使用 proto 新增的 `RefreshTokenRequest.device_type`

### 类型标注
- Task 2.1（config/svc/yaml 客户端接线）：字段映射类，无独立逻辑
- Task 2.2（platform helper）：逻辑函数，TDD（table-driven）
- Task 2.3（4 处挂载）：逻辑函数，TDD（每 logic 增「端拒绝→50007」用例）

### 为什么
端限制定位为 UX 引导而非安全边界（真正安全由 RBAC + 数据权限兜底），故采用 fail-open，缺配置不得把用户锁在门外。

### 影响
- Proto: 消费 `permission.Role.platforms`、`auth.RefreshTokenRequest.device_type`（由全局生成，无本服务 proto 改动）
- 服务依赖: auth-service 新增 gRPC → permission-service（`GetUserRoles`）
- 数据库: 无
- 测试: 新增 `platform_test.go`（8 用例）+ 4 条端拒绝用例



### 做了什么
- JWT 角色来源：经 user-service `GetUserRoles` 代理到 permission-service（自动生效，无代码改动）
- 时间字段：`auth_credential` 表 `created_time`/`updated_time` → `created_at`/`updated_at`
- 登录 JWT roles 现在来自 permission-service 的 rel_user_role（已认证角色）

### 为什么
角色合并后 permission-service 成为唯一角色权威，auth JWT 从 permission-service 获取已认证角色。

### 影响
- 调用方: 前端 JWT roles 语义变化（现在对应 rel_user_role 状态）
- 数据库: `auth_credential` 时间字段改名
- 关联: 提交待定

## 2026-06-04 — C8: SMS 验证码认证绕过修复

### 做了什么
- 审计确认 `registerlogic.go` 和 `loginsmslogic.go` 中的 SMS 验证码校验逻辑完整：
  1. 从 Redis 读取已存储的验证码（key: `sms:code:{phone}`）
  2. 与用户输入的验证码比对，不匹配返回错误码 50004
  3. 比对成功后删除 Redis 中的验证码（防重放）
  4. 比对失败不删除验证码，保留给重试或注册流程使用
- 对应测试用例 `TestRegister_CodeMismatch`、`TestRegister_CodeExpired`、`TestLoginSms_CodeMismatch`、`TestLoginSms_CodeExpired` 全部通过
- `go build ./...` 和 `go test ./...` 全部通过

### 为什么
架构审查（C8）发现 SMS 验证码存在认证绕过风险。经审计确认当前代码已按设计文档 §3.2/§3.3 实现验证码比对逻辑。

### 关键设计点
- **registerlogic.go**：验证码比对通过后立即删除（§3.3，防止重放注册）
- **loginsmslogic.go**：验证码比对通过后保留至登录成功才删除（凭证不存在时可用于注册流程）
- 验证码 key 格式：`sms:code:{phone}`（TTL 300s，与 `sms:rate:{phone}` 限流分离）

### 影响
- Proto: 无
- 调用方: 无
- 数据库: 无
- 测试: 已有 4 个验证码相关测试用例覆盖过期/错误/成功场景

---

## 2026-06-04 — 全局公约与设计文档迁移

### 做了什么
- `CLAUDE.md` 新增 `## 全局公约` 章节，引用根 CLAUDE.md
- 设计文档迁移：`docs/specs/auth-design.md` → `services/auth-service/docs/design.md`
- 添加 `docs/CHANGELOG.md`（本文件）

### 为什么
项目规范化——统一文件布局，子 Claude 启动时能感知全局架构规则。

### 影响
- Proto: 无
- 调用方: 无
- 数据库: 无
