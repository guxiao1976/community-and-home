---
triggers: ["模板", "template", "变量", "variable", "{{input}}", "{{content}}", "prompt", "提示词", "ModerateText"]
service: ai-model-service
severity: must-follow
type: pitfall
status: active
created: 2026-06-17
---

# 模板变量名必须用 {{content}}，不能用 {{input}}

## 问题

ai-model-service 的 `ModerateText` RPC 在渲染模板时注入的变量名是 `{{content}}`（来自 `in.Content`）：

```go
renderedPrompt := l.svcCtx.TemplateManager.Render(template.Content, map[string]string{
    "content": in.Content,
})
```

如果模板写 `用户输入：{{input}}`，`{{input}}` 不会被替换，模型会收到字面量 `{{input}}` 而非实际内容，导致"用户输入内容未提供，无法进行判断"。

## 正确写法

```text
用户输入：{{content}}

判断标准：...
```

## 验证

写完模板后确认 `{{content}}` 在模板正文中出现，不是 `{{input}}`、`{{user_input}}`、`{{text}}` 等其他变量名。

## 关联

- [[moderation-checktext-pipeline]] — CheckText 调用 ModerateText 的链路
- [[grpc-timeout-layers]] — 大模型调用超时配置
