---
triggers: ["Vue 模板", "{{", "插值", "Unterminated string constant", "template parse error", "花括号"]
service: web
severity: must-follow
status: active
created: 2026-06-12
updated: 2026-06-12
---

# Vue 模板中不能出现嵌套的 {{ 字面量

## 为什么会有这条经验

提示词模板的"已识别变量"展示中写了 `{{ '{{' + v + '}}' }}`，Vue 模板编译器将内层的 `'{{'` 解析为嵌套插值起始标记，报 `Unterminated string constant` 错误，触发 Vite error overlay。

## 怎么做

将 `{{` 拆开为 `{'{'`，`}}` 拆开为 `'}'`：

```vue
<!-- ❌ 错误 -->
<el-tag>{{ '{{' + v + '}}' }}</el-tag>

<!-- ✅ 正确 -->
<el-tag>{{ '{' + '{' + v + '}' + '}' }}</el-tag>
```

如果只是展示变量名（不需要花括号装饰），直接 `{{ v }}` 更简洁。

## 怎么验证

- Vue dev server 启动后不应有 template parse error overlay
- 包含 `{{ }}` 字面量的模板在浏览器中正常渲染

## 关联经验

- [[api-response-single-wrap]]
