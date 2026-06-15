---
triggers: ["gRPC timeout", "DeadlineExceeded", "context deadline", "超时", "模型调用超时", "LLM timeout", "大模型调用"]
service: all
severity: must-follow
status: active
created: 2026-06-12
updated: 2026-06-12
---

# gRPC 调用链需要三层超时对齐

## 为什么会有这条经验

调用 DeepSeek 模型时频繁出现 `DeadlineExceeded`，根因是超时控制分散在三层，每层独立截断：

| 层 | 配置位置 | 默认值 | 影响 |
|----|---------|--------|------|
| REST Server | `api/etc/*.yaml` `Timeout` | ~3s（go-zero 默认） | HTTP 请求被截断，返回 503 |
| gRPC Client | `api/etc/*.yaml` `RpcClientConf.Timeout` | 2000ms（go-zero 默认） | API→RPC 调用超时 |
| RPC Server | `rpc/etc/*.yaml` `Timeout` | 0 或很短 | RPC 端处理超时 |
| gRPC Context | `l.ctx`（HTTP request context） | 继承 REST Server 超时 | 即使外层设长，内层 Context 先到期 |

大模型单次调用需 2-10 秒，三层默认超时都不够。

## 怎么做

1. **所有三层都显式设超时**：

```yaml
# api/etc/xxx.yaml
Timeout: 120000  # REST 层

# api/etc/xxx.yaml RPC client
AiModelRpc:
  Timeout: 60000  # gRPC 客户端

# rpc/etc/xxx.yaml
Timeout: 120000  # RPC 服务端
```

2. **模型调用用独立 Context**（绕开 HTTP request context）：

```go
callCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
defer cancel()
rpcResp, err := l.svcCtx.AiModelRpc.CallModel(callCtx, &aimodel.ModelCallRequest{...})
```

3. **超时值从大到小**：RPC server ≥ REST server > gRPC client > external API timeout

## 怎么验证

- 用长文本调用大模型，不应在 10 秒内返回 `DeadlineExceeded`
- curl 直接测试，观察 `time_total` 和 HTTP 返回码

## 关联经验

- [[llm-connection-test]]
