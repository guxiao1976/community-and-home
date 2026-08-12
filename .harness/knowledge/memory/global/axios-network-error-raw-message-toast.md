---
triggers: ["axios", "网络错误", "Network Error", "错误提示", "toast", "error.message", "拦截器", "离线"]
service: all
severity: should-follow
type: pitfall
status: active
created: 2026-08-12
updated: 2026-08-12
---

# axios 网络错误 e.message 为英文 'Network Error'，直接展示会覆盖拦截器友好中文提示

## 为什么会有这条经验

项目 request.ts 响应拦截器已对网络错误统一 `showToast('网络连接失败，请检查网络')` 并 reject 原始 AxiosError（其 message 为英文 'Network Error'）。业务 catch 中若用 `e?.message || e?.msg || 默认文案` 兜底展示，对网络错误会把友好中文提示替换为英文原文（uni.showToast / ElMessage 后者覆盖前者）。

## 怎么做

优先取 `e?.msg` 或 `error.response?.data?.msg`，无 response 且无业务消息时依赖拦截器 toast 或使用中文案。web/pc 与 web/mobile 拦截器模式相同，均适用。

## 怎么验证

- 断网触发请求，观察提示为中文「网络连接失败」而非英文 'Network Error'
- grep catch 分支确认优先取 `e?.msg` / `response.data.msg`
