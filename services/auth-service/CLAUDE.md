# 认证服务 — auth-service

## 角色定位

这是 **认证服务**（`github.com/guxiao1976/community-auth`），社区平台的认证中心。**RPC-only**，提供 AT+RT 双 Token 认证（密码登录/SMS 登录/注册/刷新/登出/验证）。调用 user-service gRPC 获取用户数据。

## 启动上下文

按顺序读取：

1. `docs/graph-context.md` — Neo4j 自动生成的服务上下文
2. `docs/design.md` — JWT 设计、认证流程、安全机制
3. `../../.harness/rules/项目编码规范.md` — 编码硬性约束
4. `../../.harness/knowledge/memory/MEMORY.md` — 全局经验，按触发词匹配
5. `.claude/memory/` — 本服务经验记忆

## 关键规则

以下仅列出本服务特有规则，通用约束见 `.harness/rules/`：

1. **Token 安全** — RT 用 Redis 持久化、RefreshToken 先拉角色再旋转 RT（防止角色拉取失败时旧 RT 已销毁）
2. **密码安全** — Bcrypt 哈希存储、RSA 传输加密
3. **AT 黑名单** — 注销时 AT 加入 Redis 黑名单，写入失败必须返回错误
4. **调用 user-service** — 获取/操作用户数据必须走 gRPC，禁止直连 user DB

## 全局公约

全局约束见根 [`CLAUDE.md`](../../CLAUDE.md) §7条硬性约束。提交前 `bash ../../.harness/skills/qa/scripts/harness-checks.sh --service auth-service`。

## 常用命令

```bash
go build ./...        # 构建
go test ./...         # 测试
cd rpc && go run authservice.go   # 运行 RPC
```

## 架构

```
rpc/                    # gRPC 服务（唯一入口）
  internal/
    logic/              # 业务逻辑（login/logout/register/refresh/sms）
    svc/                # 依赖注入（DB, Redis, UserServiceRpc 客户端）
```
