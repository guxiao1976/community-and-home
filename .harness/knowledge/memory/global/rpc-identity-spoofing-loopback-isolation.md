---
triggers: "RPC 身份伪造 0.0.0.0 无鉴权 mTLS user_id metadata 数据权限 网络隔离 回环 8088 ListenOn 服务凭据 拦截器"
status: active
severity: should-follow
type: pitfall
created: 2026-08-12
updated: 2026-08-12
last_applied: null
apply_count: 0
---

# RPC 身份伪造风险：metadata 盲信 + 0.0.0.0 无鉴权（仓库级模式）

## 场景

2026-08-12 多视角安全评审（community-hub-service access-data-permission 阶段④）发现：

1. 数据权限身份经 gRPC metadata 传输（`scope.UserIDFromCtx` 盲信入站 `user_id`），无任何服务间认证。
2. RPC 绑定 `0.0.0.0:8088`（`rpc/etc/communityhub.yaml`），无鉴权拦截器注册（common 仅有
   RequestID/Logger/Recovery/ErrorHandler）。
3. 任何能连通 8088 的对端（宿主机局域网、Docker 桥接网络内其他容器）都可注入 `user_id=<任意用户>`
   伪造身份，使 `AssertPublishScope`（写）与 `GetDataScopes` 过滤（读）全部失效——数据权限特性
   的"fail-closed/防伪造"只对 REST 网关一跳成立，RPC 边界处身份完全可伪造。
4. 仓库级模式：**9 个服务全部 `ListenOn: 0.0.0.0` 且无 mTLS/服务凭据**（user 8082 / auth 8083 /
   perm 8084 / file 8085 / moderation 8086 / masterdata 8087 / community-hub 8088 / ai-model 8080）。

## 修复（community-hub 落地）

1. **RPC 绑定回环**：`ListenOn: 127.0.0.1:8088`。go-zero `figureOutListenOn` 对非 `0.0.0.0` 的 host
   原样注册 etcd（含 `127.0.0.1`），单机拓扑（scripts/start.sh 全部 `go run` 在宿主）下 API 网关与
   moderation 回调可正常发现。只允许宿主机可信进程连通，阻断局域网/Docker 桥接对端。
2. **配置不变式测试**：`rpc/internal/config/config_test.go` 断言 `ListenOn` 必须回环，
   防回退（0.0.0.0 时 FAIL）。
3. **信任边界文档化**：`scope/userctx.go` 注释明确"metadata 盲信的安全前提 = 网络隔离"。
4. **沉淀记忆**（本文件）：供后续 8 个服务统一治理。

## 实践规则

1. **新增/修改带身份 gRPC 接口的服务**：若身份经 metadata 传输，必须保证 RPC 只被可信网络可达——
   要么绑定回环/内网（单机），要么 Docker 内网隔离，要么叠加服务凭据/mTLS + unary 拦截器校验。
2. **仓库级根治方向**（Owner 协调，涉及 common/ 与全部 9 服务）：服务凭据（共享 token 或 mTLS）
   + 在 common 增加统一鉴权 unary 拦截器；`UserIDFromCtx` 应叠加调用方身份校验。
3. **不要只靠代码层"防伪造"**：代码只能保证"从 body 取身份被忽略"，网络层不做隔离时 metadata
   身份一样可伪造。
4. **回环绑定注意**：若未来服务拆分多主机/K8s 多副本，回环绑定不可用，须切换为内网绑定 +
   网络策略或 mTLS。
