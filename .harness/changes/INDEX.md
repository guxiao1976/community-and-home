# 变更追溯索引

> 记录所有通过 OpenSpec 流程完成的需求/功能开发，按时间倒序排列

## 格式说明

```
## YYYY-MM-DD — <变更名>

**路径**: [直接Edit / Dev Agent / OpenSpec]
**状态**: [进行中 / 已完成 / 已归档]
**涉及服务**: service-a, service-b
**关联**: [PR #123] [Issue #456]

[一句话描述]

详见: [.harness/changes/<change-name>/](./<change-name>/)
```

---

## 2026-06-18 — 审核服务管线配置化

**路径**: OpenSpec
**状态**: 已归档
**涉及服务**: moderation-service
**关联**: commit 84eadb2

实现内容审核的管线配置化功能，支持动态配置审核策略。

详见: [.harness/changes/moderation-pipeline-config/](./moderation-pipeline-config/)

---

## 2026-06-16 — AI 模型服务增强

**路径**: Dev Agent
**状态**: 已归档
**涉及服务**: ai-model-service
**关联**: 无

为 AI 模型服务添加连接测试和模板管理功能。

详见: [.harness/changes/ai-model-enhancements/](./ai-model-enhancements/)
