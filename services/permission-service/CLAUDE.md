# CLAUDE.md — permission-service

## 角色定位

这是 **权限服务**（`github.com/guxiao1976/community-permission`），基于 RBAC 的权限和数据范围引擎。**API+RPC 双层**，提供权限检查、数据范围查询、角色管理。

## 启动上下文

按顺序读取：

1. `docs/graph-context.md` — Neo4j 自动生成的服务上下文
2. `docs/design.md` — RBAC 模型、权限检查流程
3. `../../.harness/rules/项目编码规范.md` — 编码硬性约束
4. `../../.harness/knowledge/memory/MEMORY.md` — 全局经验，按触发词匹配
5. `.claude/memory/` — 本服务经验记忆

## 关键规则

以下仅列出本服务特有规则，通用约束见 `.harness/rules/`：

1. **权限缓存一致性** — Redis 缓存与 MySQL 保持一致，修改角色/权限时必须批量刷新缓存（Redis KEYS → DEL）
2. **is_system 仅防止误删/误改** — `is_system=1` 的角色不自动获得全权限，权限由 `rel_role_permission` 配置决定
3. **数据范围安全** — `GetDataScopes` 返回的 scope_id 列表由调用方用于 WHERE IN 过滤

## 全局公约

全局约束见根 [`CLAUDE.md`](../../CLAUDE.md) §7条硬性约束。提交前 `bash ../../.harness/skills/qa/scripts/harness-checks.sh --service permission-service`。
## 常用命令

```bash
go build ./...        # 构建
go test ./...         # 测试
cd rpc && go run permissionservice.go   # 运行 RPC (8084)
cd api && go run perm.go                # 运行 API (8883)
```

## 架构

```
api/                    # REST 网关（JWT 认证，代理 gRPC）
rpc/                    # gRPC 服务（CheckPermission, GetDataScopes, Role CRUD）
model/                  # GORM 模型（sys_role, sys_permission, rel_role_permission, rel_user_role）
```

## 核心概念

- **Role**: owner / property_admin / community_admin / grid_worker，支持 `is_system` 标记
- **Permission**: API 级别，类型 menu / button / api
- **UserRole**: user_id + role_id + scope_type + scope_id
- **DataScope**: 控制用户可见的数据边界
