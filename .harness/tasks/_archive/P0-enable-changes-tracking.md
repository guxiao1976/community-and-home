---
priority: P0
status: closed
created: 2026-06-23
completed: 2026-08-14
assignee: harness-team
source: human
estimated_effort: 2h
---

# P0: 启用变更追踪机制

## 问题描述

`.harness/changes/` 目录不存在，导致：
- owner-agent.md 定义的 OpenSpec 流程无法落地
- 需求 → 设计 → 实现 → QA → Review 的追溯链断裂
- 无法回溯历史决策

## 影响范围

- 阻塞完整的 Harness Pipeline 执行
- 无法积累变更知识
- 违反 Harness 可追溯原则

## 解决方案

### 1. 创建目录结构

```bash
cd /home/jiaoxh/my-project/community-and-home

# 创建变更追踪根目录
mkdir -p .harness/changes

# 创建索引文件
cat > .harness/changes/INDEX.md <<'EOF'
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

详见: [.harness/changes/<change-name>/](./\<change-name\>/)
```

---

## 2026-06-18 — 审核服务管线配置化

**路径**: OpenSpec
**状态**: 已完成
**涉及服务**: moderation-service
**关联**: 无

实现内容审核的管线配置化功能，支持动态配置审核策略。

详见: [.harness/changes/moderation-pipeline-config/](./moderation-pipeline-config/)

---

## 2026-06-16 — AI 模型服务连接测试

**路径**: Dev Agent
**状态**: 已完成
**涉及服务**: ai-model-service
**关联**: 无

为 AI 模型服务添加连接测试功能，验证大模型 API 可达性。

详见: [.harness/changes/ai-model-connection-test/](./ai-model-connection-test/)
EOF
```

### 2. 创建模板文件

```bash
# 复制现有 TEMPLATE.md 或创建新模板
mkdir -p .harness/changes/TEMPLATE
cat > .harness/changes/TEMPLATE/summary.md <<'EOF'
# 变更摘要 — <变更名>

**创建时间**: YYYY-MM-DD HH:MM
**完成时间**: YYYY-MM-DD HH:MM
**路径**: [直接Edit / Dev Agent / OpenSpec]

---

## 阶段 0: 路径选择

- **路径**: <选择的路径>
- **理由**: <判定条件>
- **涉及服务**: <服务列表>
- **跳过阶段**: <列出跳过的阶段及理由>

---

## 阶段 1: 需求分析

- **执行方式**: <内联 / 子Agent>
- **产出**: proposal.md + specs/*/spec.md
- **状态**: ✅ 通过 / ❌ 回退

---

## 阶段 2: 需求评审

- **执行方式**: 3 子Agent 并行
- **投票结果**: coverage: APPROVED, structure: APPROVED, clarity: APPROVED
- **状态**: ✅ 通过 / ❌ 回退

---

## 阶段 3: 架构设计

- **执行方式**: 子Agent
- **产出**: design.md + tasks.md
- **状态**: ✅ 通过 / ❌ 回退

---

## 阶段 4: Proto 变更

- **执行方式**: Owner 内联
- **变更清单**: <列出 Proto 文件>
- **状态**: ✅ 通过 / ⏭️ 跳过

---

## 阶段 5: 编码+测试

- **执行方式**: N×Workflow 并行
- **服务**: <列出服务>
- **迭代次数**: <各服务迭代次数>
- **QA 结果**: <PASS/FAIL>
- **Review 结果**: <投票结果>
- **置信度**: <0.0-1.0>
- **状态**: ✅ 通过 / ❌ 回退

---

## 阶段 6: 集成归档

- **全链路编译**: ✅ 通过 / ❌ 失败
- **运行时冒烟**: ✅ 通过 / ⚠️ 部分失败 / ⏭️ 跳过
- **Memory Suggestions**: <数量>
- **状态**: ✅ 完成

---

## 关键决策

| # | 决策点 | 决策内容 | 原因 |
|---|--------|---------|------|

---

## 例外 & 未解决问题

| # | 问题 | 影响 | 后续计划 |
|---|------|------|---------|

---

## 人工抽查（置信度 < 0.80）

| 文件 | 发现 | 修复 |
|------|------|------|

---

## 交付清单

- [ ] 代码已合并到 master
- [ ] CHANGELOG 已更新
- [ ] QA/Review 报告已归档
- [ ] Memory Suggestions 已处理
- [ ] 索引已更新
EOF
```

### 3. 补充现有变更的追溯

```bash
# 为已完成的 moderation-pipeline-config 创建追溯目录
mkdir -p .harness/changes/moderation-pipeline-config/impl/moderation-service

# 移动现有 QA/Review 报告
mv services/moderation-service/_qa.md \
   .harness/changes/moderation-pipeline-config/impl/moderation-service/

mv services/moderation-service/_review_*.md \
   .harness/changes/moderation-pipeline-config/impl/moderation-service/

# 创建 summary.md（基于模板）
# ... 填充实际数据 ...
```

## 完成标准

- [ ] `.harness/changes/` 目录创建完成
- [ ] INDEX.md 包含已完成变更的索引
- [ ] TEMPLATE/ 目录包含 summary.md 模板
- [ ] 现有 QA/Review 报告归档到对应变更目录
- [ ] 更新 owner-agent.md 中的产出路径示例
- [ ] Git 提交并推送

## 关联

- 规范: `owner-agent.md` Line 100-113
- 模板: `.harness/changes/TEMPLATE.md` (待创建)
