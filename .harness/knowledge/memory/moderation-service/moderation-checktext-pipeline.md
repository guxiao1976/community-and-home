---
triggers: ["CheckText", "管线", "pipeline", "审核接口", "RPC审核", "生产管线", "is_production", "内容审核调用", "moderation"]
service: moderation-service
severity: must-follow
type: decision
status: active
created: 2026-06-17
---

# CheckText RPC 已接入管线配置

## 背景

2026-06-16 之前，moderation-service 的 `CheckText` RPC 使用硬编码的 `TextEngine` 执行审核，管线配置只在测试页面（`POST /api/moderation/pipeline/test`）使用。

## 现在的行为

```
CheckText RPC 调用
  → 查 mod_pipeline_config WHERE is_production=1 AND is_active=1
  → 有生产管线 → PipelineExecutor.Execute(config, content) → AC引擎→大模型
  → 无生产管线 → 回退旧 TextEngine（兼容已有调用方）
```

## 对调用方的影响

**无任何影响。** gRPC 接口签名（`CheckTextRequest` / `CheckTextResponse`）不变，只是内部执行逻辑变了。

调用方示例（community-hub-service 等）：
```go
resp, err := moderationClient.CheckText(ctx, &moderationv1.CheckTextRequest{
    Content: "用户输入内容",
})
// resp.Pass, resp.RiskLevel, resp.Reason  — 字段不变
```

## 管线生效步骤

1. 在「内容审核管线配置」页面新建/编辑管线
2. 配置大模型模板
3. 点「设为生产」
4. 此后所有 `CheckText` RPC 调用自动使用该管线

## 配置位置

- 生产管线配置存在 `moderation_db.mod_pipeline_config` 表
- `is_production=1` 的那条即为生效配置
- RPC ServiceContext 通过 `PipelineModel.FindProduction()` 加载

## 关联

- [[api-response-single-wrap]] — PipelineExecutor 输出需符合响应格式约定
- [[grpc-timeout-layers]] — 管线中大模型调用需独立 context 超时
