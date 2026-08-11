# CLAUDE.md — monitoring-service

## 角色定位

这是 **运行监控服务**（`github.com/guxiao1976/community-monitoring`），API-only，负责聚合检测系统各组件的健康状态。无状态、无数据库依赖。

## 启动上下文

按顺序读取：

1. `docs/graph-context.md` — Neo4j 自动生成的服务上下文
2. `docs/design.md` — 检测机制、配置结构
3. `../../.harness/rules/项目编码规范.md` — 编码硬性约束
4. `../../.harness/knowledge/memory/MEMORY.md` — 全局经验，按触发词匹配
5. `.claude/memory/` — 本服务经验记忆

## 关键规则

以下仅列出本服务特有规则，通用约束见 `.harness/rules/`：

1. **只做探测和聚合** — 不修改外部系统状态
2. **配置驱动** — 检测目标全在 YAML 中配置，增减服务无需改代码
3. **无 Proto** — 本服务无 gRPC 接口，不需要 api-proto/

## 全局公约

全局约束见根 [`CLAUDE.md`](../../CLAUDE.md) §7条硬性约束。提交前 `bash ../../.harness/skills/qa/scripts/harness-checks.sh --service monitoring-service`。
## 常用命令

```bash
go build ./...        # 构建
go test ./...         # 测试
cd api && go run monitoring.go -f etc/monitoring-api.yaml   # 运行 (8886)
```
