---
triggers: ["DeepSeek", "thinking", "Claude adapter", "response parse", "content block", "deepseek-v4"]
service: ai-model-service
severity: must-follow
type: pitfall
status: active
created: 2026-06-12
updated: 2026-06-12
---

# DeepSeek/Claude 响应可能包含 thinking 块，需遍历所有 Content

## 为什么会有这条经验

DeepSeek V4 系列（`deepseek-v4-pro`、`deepseek-v4-flash`）的 Anthropic 兼容接口会返回 `type: "thinking"` 的 content block。Claude 适配器原本只读 `Content[0].Text`，而 thinking 块没有 `.Text` 字段，导致解析出的 content 为空（即使 tokens 计数正确）。

## 怎么做

遍历所有 content blocks，拼接所有 `type: "text"` 的内容：

```go
var claudeResp struct {
    Content []struct {
        Type     string `json:"type"`
        Text     string `json:"text"`
        Thinking string `json:"thinking"`  // ← thinking 块
    } `json:"content"`
    // ...
}

var content string
for _, block := range claudeResp.Content {
    if block.Text != "" {
        content += block.Text  // 拼接所有文本块
    }
}
```

## 怎么验证

- 用 `deepseek-v4-pro` 模型调用测试，响应不应为空
- 用非 thinking 模型（如 `deepseek-chat`）测试，响应也应正常

## 关联经验

- [[llm-connection-test]]
- [[endpoint-auto-complete]]
