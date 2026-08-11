---
triggers: ["gRPC", "ResourceExhausted", "message larger than max", "消息上限", "4MB", "敏感词", "MaxCallRecvMsgSize", "grpc.Dial", "MustNewClient"]
service: moderation-service
severity: must-follow
type: pitfall
status: active
created: 2026-06-17
---

# gRPC 默认 4MB 消息上限无法传输 5万+ 敏感词

## 现象

```text
rpc error: code = ResourceExhausted desc = grpc: received message larger than max (4524111 vs. 4194304)
```

master-data-service 单次 gRPC 返回 51,080 条敏感词时消息体 4.5MB，超过 gRPC 默认 4MB 接收限制，导致 moderation-service 启动时 `FATAL: word store load failed`。

## 修复

创建 gRPC 客户端时必须设置 `MaxCallRecvMsgSize`：

```go
conn, err := grpc.NewClient(endpoint,
    grpc.WithTransportCredentials(insecure.NewCredentials()),
    grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(20*1024*1024)), // 20MB
)
```

## 影响范围

- moderation-service RPC → master-data-service（获取敏感词）
- 任何需要接收大消息的 gRPC 客户端

## 关联

- [[grpc-only-comms]] — 服务间通信仅 gRPC 的规则
