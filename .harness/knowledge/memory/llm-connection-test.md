---
triggers: ["连接测试", "test connection", "大模型 API", "LLM endpoint", "测试连接", "DeepSeek", "OpenAI", "Claude", "Anthropic", "provider", "连通性"]
service: ai-model-service
severity: must-follow
status: active
created: 2026-06-12
updated: 2026-06-12
---

# 大模型 API 连接测试实现规范

## 为什么会有这条经验

首次实现连接测试时，代码用 HTTP GET 请求去探测大模型端点，所有大模型 API（OpenAI/Claude/DeepSeek/Ollama）均返回 404。DeepSeek Anthropic 兼容端点路径为 `/anthropic/v1/messages`，但地址栏填入 `https://api.deepseek.com/anthropic` 时，路径补全逻辑因路径非空（`/anthropic` ≠ 空）而跳过补全，导致持续 404。

## 怎么做

### 1. 必须用 POST

大模型 API 不接受 GET。发送最小化聊天补全请求：
```json
{"model":"ping","max_tokens":1,"messages":[{"role":"user","content":"ping"}]}
```

### 2. Auth Header 按提供商区分

- OpenAI / OpenAI 兼容 → `Authorization: Bearer <key>`
- Claude / Anthropic 兼容 → `x-api-key: <key>`

### 3. 成功判定包含 "可恢复的失败"

- HTTP 2xx → 成功
- HTTP 4xx + 响应体能解析为 JSON → **也判定成功**（网络通、认证通过，只是模型名/参数等非致命问题）
- HTTP 401/403 → 认证失败（不成功）
- HTTP 404 → 端点错误（不成功）
- 网络超时/连接拒绝 → 网络层失败（不成功）

### 4. 端点路径补全

补全判断必须检测"路径是否已包含标准 API 后缀"，**而非"路径是否为空"**：

```go
// ✅ 正确：检查路径是否已含完整 API 路径
if !strings.Contains(u.Path, "/v1/messages") {
    u.Path = strings.TrimRight(u.Path, "/") + "/v1/messages"
}

// ❌ 错误：只检查路径是否为空 — /anthropic 不是空路径但也不是完整API路径
if u.Path == "" || u.Path == "/" {
    u.Path = "/v1/messages"
}
```

### 5. 提供商路径对照

| 提供商 | 自动追加路径 |
|--------|------------|
| OpenAI | `/v1/chat/completions` |
| Claude / Anthropic | `/v1/messages` |
| Ollama | `/api/generate` |

## 怎么验证

- 填入 `https://api.deepseek.com/anthropic` + provider=claude，测试连接不应返回 404
- 填入 `https://api.openai.com` + provider=openai，测试连接不应返回 404
- 用错误的 API Key 测试，应返回 "API reachable" 而非 "endpoint not found"

## 关联经验

- [[api-response-single-wrap]]
