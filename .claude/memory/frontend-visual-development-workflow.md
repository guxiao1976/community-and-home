---
name: frontend-visual-development-workflow
description: 前端开发优先使用浏览器可视化伴侣进行设计评审，再进入编码实现
metadata:
  type: project
---

使用 brainstorming 技能的 Visual Companion（浏览器可视化伴侣）进行前端 UI 开发：

1. 启动 Visual Companion 服务器，通过 HTML mockup 在浏览器中展示设计方案
2. 迭代调整布局、配色、字号、间距等视觉细节，用户直接在浏览器中看到效果
3. 设计定稿后，写出设计文档（spec），再编写实现计划（plan），最后编码实现

**Why:** 2026-06-06 开发移动端公告信息页面重设计时首次使用该工具，极大提升了设计迭代效率。传统"文字描述→编码→看效果→再改"的循环被缩短为"mockup可视化→实时调整→定稿→编码"，减少了大量返工。

**How to apply:**
- 任何涉及 UI 视觉改动的任务，先启动 brainstorming → Visual Companion，在浏览器中 mockup 设计
- 用户确认设计后，再写 spec、plan、编码
- 适用场景：页面重设计、新页面开发、组件样式调整
- 不适用：纯逻辑修改、API 调整、后端开发
