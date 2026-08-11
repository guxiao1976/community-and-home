# CHANGELOG — user-service

## 2026-08-11 — RBAC 角色体系合并 + 认证 REST API

### 做了什么
- **废弃 `user_membership_role`**：角色授予迁移到 permission-service 的 `rel_user_role`
- `ApplyRole` 改调 permission-service AssignRole（写入 rel_user_role，status=0）
- `SubmitCertification` 改走 permission-service（提交时 UpdateUserRoleStatus status=1）
- `ReviewCertification` 改走 permission-service（通过 status=2+expires，驳回 status=3）
- `GetUserRoles`/`CheckAccess` 改为代理 permission-service
- 新增 `role_mapper.go`：role_code↔role_id 映射（调 permission-service ListRoles 缓存）
- 新增认证 REST API：
  - `POST /api/users/certifications`（提交认证材料）
  - `GET /api/users/certifications`（我的认证记录）
  - `GET /api/verifications`（管理员列表认证申请）
  - `POST /api/verifications/:id/review`（管理员审核）
- 移动端 `my.vue` hasOwnerRole 改为真实查询

### 为什么
permission-service 成为角色唯一权威，认证流程从 user-service 自管角色改为调用 permission-service。

### 影响
- Proto: 无（复用现有 RPC）
- 调用方: auth-service（JWT roles 经代理获取）、移动端（applyRole/getUserRoles）
- 数据库: 废弃 `user_membership_role` 表，rel_user_role 承载角色
- 关联: 提交待定

## 2026-06-04 — 错误码 6 位 → 5 位统一

### 做了什么
- 所有错误码从 6 位 `10X00Y` 改为 5 位 `10X0Y`（去掉中间多余的 0）
- 更新文件：`rpc/internal/logic/user/` 下 12 个 .go 文件，`docs/design.md` 错误码表

### 映射
| 旧码 | 新码 | 含义 |
|------|------|------|
| 100001 | 10001 | 用户不存在 |
| 100002 | 10002 | 手机号已注册 |
| 100003 | 10003 | 重复提交认证 |
| 100004 | 10004 | 信用分不足 |
| 100005 | 10005 | 小区成员不存在/退出 |
| 100006 | 10006 | 最多加入5个小区 |
| 100007 | 10007 | 认证申请不存在 |
| 100008 | 10008 | 角色已存在 |
| 100009 | 10009 | 角色已过期 |
| 100010 | 10010 | 权限不足 |
| 100400 | 10040 | 参数校验失败 |

### 影响
- Proto: `api-proto/api/user/v1/user.proto` 注释中的错误码需同步更新（告知全局 Claude）
- 调用方: auth-service 需关注错误码变化
- 数据库: 无

---

## 2026-06-04 — 全局公约与设计文档迁移

### 做了什么
- `CLAUDE.md` 新增 `## 全局公约` 章节，引用根 CLAUDE.md
- 设计文档迁移：`docs/specs/user-design.md` → `services/user-service/docs/design.md`
- 添加 `docs/CHANGELOG.md`（本文件）

### 为什么
项目规范化——统一文件布局，子 Claude 启动时能感知全局架构规则。

### 影响
- Proto: 无
- 调用方: 无
- 数据库: 无
