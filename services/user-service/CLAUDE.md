# 用户服务 — user-service

## 角色定位

这是 **用户服务**（`github.com/guxiao1976/community-user`），提供统一的用户资料管理、小区归属、角色认证。**RPC-only**（api/ 层仅有脚手架，未实现）。被 auth-service 调用。

## 启动上下文

按顺序读取：

1. `docs/graph-context.md` — Neo4j 自动生成的服务上下文（依赖/路由/表/血缘）
2. `docs/design.md` — 数据模型、业务流程、权限模型
3. `../../.harness/rules/项目编码规范.md` — 编码硬性约束
4. `../../.harness/knowledge/memory/MEMORY.md` — 全局经验，按触发词匹配
5. `.claude/memory/` — 本服务经验记忆

## 关键规则

以下仅列出本服务特有规则，通用约束见 `.harness/rules/`：

1. **手机号必须 AES 加密存储** — 写入加密、读取解密。sqlx 需手动调用 `crypto.AESDecrypt`，非 GORM 自动
2. **写入前校验关联实体存在** — JoinCommunity/ApplyRole 前 FindOne user，ReviewCertification 前校验 reviewer_id
3. **被 auth-service 调用** — 修改接口需评估对 auth 的兼容性
4. **Migration 必须三步闭环** — 写 SQL → 提交 → 在数据库执行并验证

## 全局公约

全局约束见根 [`CLAUDE.md`](../../CLAUDE.md) §7条硬性约束。提交前 `bash ../../.harness/skills/qa/scripts/harness-checks.sh --service user-service`。
## 常用命令

```bash
go build ./...        # 构建
go test ./...         # 测试
cd rpc && go run userservice.go -f etc/userservice.yaml   # 运行 RPC
```

## 架构

```
rpc/                    # gRPC 服务（主要实现）
  internal/
    logic/user/         # 业务逻辑（CRUD + 认证流程）
model/                  # GORM 数据模型（user_base, community_membership, role, certification, residence）
api/                    # REST 层（脚手架，未实现）
```
