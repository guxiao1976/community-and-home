# 社区枢纽服务 — community-hub-service

## 角色定位

这是 **社区枢纽服务**（`github.com/guxiao1976/community-hub`），小区信息汇聚中心。**RPC + API 双层**，负责通知公告发布、便民联络维护、寻失互助。

## 启动上下文

按顺序读取：

1. `docs/graph-context.md` — Neo4j 自动生成的服务上下文
2. `docs/design.md` — 数据模型、业务流程
3. `../../.harness/rules/项目编码规范.md` — 编码硬性约束
4. `../../.harness/knowledge/memory/MEMORY.md` — 全局经验，按触发词匹配
5. `.claude/memory/` — 本服务经验记忆

## 关键规则

以下仅列出本服务特有规则，通用约束见 `.harness/rules/`：

1. **int64 ID 全部使用 `json:",string"`** — 确保前端 Snowflake ID 精度不丢失
2. **RPC 响应首字段 `BaseResp base`**
3. **通知/寻失的审核和权限尚未集成** — 修改这些模块时需评估 moderation-service 和 permission-service 的对接
4. **业务错误码 08xxxx** — 08=社区枢纽：080001(不存在) 080002(无权限) 080003(超限) 080004(不存在) 080005(参数无效)

## 全局公约

全局约束见根 [`CLAUDE.md`](../../CLAUDE.md) §7条硬性约束。提交前 `bash ../../.harness/skills/qa/scripts/harness-checks.sh --service community-hub-service`。
## 常用命令

```bash
go build ./...        # 构建
go test ./...         # 测试
cd rpc && go run communityhub.go -f etc/communityhub.yaml   # 运行 RPC (8088)
cd api && go run communityhub.go -f etc/communityhub-api.yaml # 运行 API (8887)
```

## 架构

```
rpc/                    # gRPC（NoticeService, ContactService, LostFoundService）
api/                    # REST API 网关（/api/community/*）
model/                  # go-zero sqlx 数据模型（notice, contact, lost_found）
```
