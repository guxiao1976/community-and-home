---
triggers: ["提交", "防重", "双击", "double submit", "submitting", "loading mask", "showLoading", "异步", "表单", "幂等"]
service: all
severity: should-follow
type: pitfall
status: active
created: 2026-08-12
updated: 2026-08-12
---

# 异步表单提交必须加防重复保护（loading mask 未渲染前的双击窗口）

## 为什么会有这条经验

前端异步提交（加入小区、登录、注册、下单等）在 `uni.showLoading({mask:true})` / ElMessage 真正渲染前存在双击窗口，第二次点击会再次发起请求；若后端不幂等则产生重复数据或触发上限误判。

## 怎么做

在 handler 内加 `submitting` ref 守卫：进入置 true、finally 复位、重入或校验失败直接 return。loading mask 只负责渲染后的拦截，不能替代代码层守卫。web/mobile 与 web/pc 均适用。

## 怎么验证

- 变更提交类 handler 确认存在 submitting 守卫
- 快速双击按钮仅触发一次请求（网络面板确认）

## 关联经验

[[insert-ignore-swallows-errors]]
