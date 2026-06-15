---
triggers: ["endpoint", "端点补全", "auto-complete", "API path", "base URL", "v1/chat", "v1/messages"]
service: ai-model-service
severity: must-follow
status: active
created: 2026-06-12
updated: 2026-06-12
---

# 端点自动补全逻辑必须在所有调用路径共享

## 为什么会有这条经验

用户填写 `https://api.deepseek.com/anthropic` 作为模型端点，但实际 API 在 `/anthropic/v1/messages`。补全逻辑最初只在连接测试中实现，健康检查和模型调用（CallModel）各自拷贝了一份，导致同样的 bug 修复了三次。

## 怎么做

**提取公共函数**，不要在多个 logic 文件中重复实现：

```go
// 放在 rpc/internal/logic/endpoint.go 或类似位置
func ResolveEndpoint(rawURL, provider string) string {
    u, _ := url.Parse(rawURL)
    switch strings.ToLower(provider) {
    case "openai":
        if !strings.Contains(u.Path, "/chat/completions") {
            u.Path = strings.TrimRight(u.Path, "/") + "/v1/chat/completions"
        }
    case "claude", "anthropic":
        if !strings.Contains(u.Path, "/v1/messages") {
            u.Path = strings.TrimRight(u.Path, "/") + "/v1/messages"
        }
    case "ollama":
        if !strings.Contains(u.Path, "/api/") {
            u.Path = strings.TrimRight(u.Path, "/") + "/api/generate"
        }
    }
    return u.String()
}
```

**调用点**：TestConnection、CheckModelHealth、CallModel — 三处都用同一个函数。

**判断标准**：检测路径是否已包含标准 API 后缀（如 `/v1/messages`），**不是**检测路径是否为空。因为 `/anthropic` 不是空路径但也不是完整的 API 路径。

## 关联经验

- [[llm-connection-test]]
- [[grpc-timeout-layers]]
