---
triggers: ["UI", "页面", "设计", "样式", "视觉", "mockup", "brainstorming", "Visual Companion"]
service: web/mobile
type: process
severity: should-follow
status: active
created: 2026-06-06
updated: 2026-06-09
last_applied: null
apply_count: 0
---

# 前端开发优先使用浏览器可视化伴侣进行设计评审

## 为什么会有这条经验

2026-06-06 开发移动端公告信息页面重设计时首次使用 Visual Companion，极大提升了设计迭代效率。
传统"文字描述→编码→看效果→再改"的循环被缩短为"mockup可视化→实时调整→定稿→编码"，减少了大量返工。

## 怎么做

1. 启动 Visual Companion 服务器，通过 HTML mockup 在浏览器中展示设计方案
2. 迭代调整布局、配色、字号、间距等视觉细节，用户直接在浏览器中看到效果
3. 设计定稿后，写出设计文档（spec），再编写实现计划（plan），最后编码实现

## 怎么验证

- 设计定稿有对应的 HTML mockup 文件（位于 `.superpowers/brainstorm/`）
- 设计文档（spec）引用了 mockup 的视觉方案
- 编码实现与 mockup 视觉一致

## 适用场景

- 适用：页面重设计、新页面开发、组件样式调整
- 不适用：纯逻辑修改、API 调整、后端开发

## 关联经验

- [[verify-api-before-calling]]
