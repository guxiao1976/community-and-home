---
triggers: "端口 冲突 port monorepo 8087 8088 start.sh stop.sh 服务启动 smoketest ListenOn yaml"
status: active
severity: must-follow
type: pitfall
created: 2026-06-17
updated: 2026-06-17
last_applied: null
apply_count: 0
---

# Monorepo 端口管理：冲突检测与启动顺序

## 场景

2026-06-17 写 smoketest 时发现：
- `community-hub-service` 和 `master-data-service` 的 RPC 都配了 `ListenOn: 0.0.0.0:8087`
- `scripts/start.sh` 只启动了 4/8 服务（user/auth/perm/file），其他 4 个（master-data/moderation/ai-model/community-hub）需手动启动
- `scripts/stop.sh` 的进程名模式（`identity.go`）与实际不匹配

## 修复

1. community-hub RPC 端口：8087 → 8088（`services/community-hub-service/rpc/etc/communityhub.yaml`）
2. `start.sh` 重写为 Tier 0 + Tier 1 两阶段启动，按依赖顺序启动全部 8 服务
3. `stop.sh` 覆盖 9 个进程名 + 17 个端口

## 实践规则

1. **新增服务时必须检查端口分配**：`grep -rn "ListenOn\|Port:" services/*/rpc/etc/ services/*/api/etc/ | grep -oP ':\d+' | sort | uniq -d`
2. **修改端口后检查下游引用**：其他服务的 yaml 中可能有 `RpcPort` 引用（如 monitoring-service 引用 master-data 的 8087）
3. **start.sh/stop.sh 保持同步更新**：新增服务时必须同时更新两个脚本
4. **用 smoketest 验证**：`bash .harness/scripts/harness-smoke.sh` 的 L1 层会检查端口监听，第一时间暴露冲突
