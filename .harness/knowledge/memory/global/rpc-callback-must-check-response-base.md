---
triggers: ["回调", "callback", "服务间调用", "Base", "ToError", "审核状态", "静默", "业务错误", "gRPC err", "080006", "060007"]
service: all
severity: should-follow
type: pitfall
status: active
created: 2026-08-12
updated: 2026-08-12
---

# 服务间回调只查 gRPC err 而忽略响应 Base → 业务错误静默吞掉

## 为什么会有这条经验

go-zero 约定业务错误进响应体 BaseResp（Code!=0）而非 gRPC status。服务间 gRPC 回调若只检查返回的 `err`（gRPC status）而丢弃响应体 `Base`，则 RPC 在 Base 里返回的业务错误（如 080006/80001/060007）会被静默吞掉，且常被错误记录为「成功」。

## 怎么做

所有服务间调用方（尤其审核状态回调、异步任务回调、补偿回调）调用 `responsex.ToError(resp.GetBase())` 校验并记录失败日志。

案例：moderation-service `callbackModerationStatus` 只检查 err、丢弃响应体；当 community-hub 的 UpdateNotice/UpdateLostFoundModerationStatus 返回 80006 denyResp（如 permission 系统身份种子缺失或 permission RPC 拒 060007）时，审核结果静默丢失、日志显示成功。

## 怎么验证

- 服务间回调是否 `responsex.ToError(resp.GetBase())` 校验
- 回调失败是否有明确日志，而非静默「成功」

## 关联经验

[[best-effort-compensation-must-log]] [[verify-api-before-calling]]
