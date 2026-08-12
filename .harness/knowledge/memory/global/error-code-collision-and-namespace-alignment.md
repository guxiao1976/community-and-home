---
triggers: ["错误码", "error code", "60006", "060007", "冲突", "复用", "语义", "NewBaseRespWithError", "同码异义", "namespace"]
type: guideline
severity: should-follow
service: all
status: active
created: 2026-08-12
updated: 2026-08-12
apply_count: 0
---
# 错误码必须语义唯一，禁止同码异义；新增 RPC 错误码需跨服务对齐命名空间

## 为什么会有这条经验

1. permission-service 新增 `assertpublishscopelogic.go` 时用错误码 60006 表示「目标小区超出发布者数据范围」，但该服务既有 `createrolelogic.go` 已用 60006 表示「角色编码已存在」。同一服务内同一个错误码对应两种不同业务语义，调用方（community-hub 等）收到 60006 时无法区分错误类型，排查与前端提示都会混淆。
2. 更隐蔽的是 **proto 注释、代码、design.md 三方命名空间不一致**：permission.proto:266 注释写 060006=无数据权限，design.md §八 登记 060006=角色编码已存在，代码实际两处都用 60006 —— 三处各说各话，评审只能靠人工 grep 发现。

## 怎么做

1. **新增错误时先全仓 grep**：`grep -rn "NewBaseRespWithError" services/<name>/rpc/`，确认要用的 5 位错误码未被其他 Logic 占用。
2. **一码一义**：同一服务内错误码严格唯一，不同语义用不同 code；重复语义则复用既有 code。
3. **新增语义 → 分配新码段**：如 permission-service 60001-60006 已用，新增用 60007+。本变更即分配 60007「目标小区超出发布者数据范围」，与 60006「角色编码已存在」解耦。
4. **三方同步**：修改错误码必须同时更新（a）service 代码常量/调用处、（b）api-proto 对应 `.proto` 头部错误码注释 + 相关 message 注释、（c）design.md / spec 错误码表。proto 变更由 Owner 执行（硬约束 #2），代码侧由子 Agent 执行后通知 Owner 同步 proto。
5. **消费方映射**：跨服务消费（如 community-hub 把 permission 的 060007 映射为 080006）需在映射处注释来源，便于追溯。

## 触发场景

- 新增/修改 RPC Logic 时引入 `responsex.NewBaseRespWithError(code, msg)`。
- 审查时核对既有错误码是否被新代码重复使用。
- 修改 proto 错误码注释但未同步 design.md / spec。

## 关联经验

- [[grpc-only-comms]] — 错误码经 BaseResp 跨服务传递，语义唯一性直接影响调用方判断
- [[insert-ignore-swallows-errors]] — 授权/分配类错误返回的错误码语义
