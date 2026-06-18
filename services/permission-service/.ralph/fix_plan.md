# Ralph Fix Plan — permission-service

## Immediate
- [x] 修复 ListPermissionsRequest 字段编号跳号（reserved 1）
- [x] 验证 RBAC 权限模型完整性
- [x] 修复错误码：60001→060001, 60004→060004, 60006→060006（7处使用，5个文件）
- [x] ListPermissions RPC 支持 type/status 过滤（当前忽略请求参数）

## Soon
- [x] CLAUDE.md 更新：确认 RPC-only 还是 API+RPC
- [x] 错误码审查（确保使用 06xxxx 前缀）
- [x] RevokeRole 缓存失效补全 building/unit/grid scope
- [x] API types RoleInfo.Id / PermissionInfo.Id/ParentId 添加 json:"...,string" 标签（Snowflake 兼容）

## Completed
- [x] Ralph 初始化
- [x] RBAC 验证：4表模型完整、13个RPC全部实现、缓存策略可用、API层正确代理gRPC
  - 发现：错误码位数不对(5位vs6位)、ListPermissions忽略过滤参数、缓存失效不完整、Snowflake json tag缺失
