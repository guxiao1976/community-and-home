# Community-Home 开发流水线全景图

## 📐 总体架构

```
┌─────────────────────────────────────────────────────────────────────┐
│                     Owner Agent（全局协调层）                          │
│                      纯编排器 + 质量把关                               │
├─────────────────────────────────────────────────────────────────────┤
│  路径选择 → 子Agent派发 → 产出验收 → Go/No-Go 裁决 → Proto管理 → 归档  │
└─────────────────────────────────────────────────────────────────────┘
                              ↓
    ┌─────────────┬──────────┴──────────┬─────────────┐
    ↓             ↓                      ↓             ↓
【需求分析】   【架构设计】          【编码管线】    【集成归档】
 子Agent       子Agent           Workflow × N      Owner内联
```

此文档已移至独立文件，请查看完整内容：
**[docs/pipeline-overview.md](pipeline-overview.md)**

---

## 🎯 快速导航

### 核心流程
- [阶段 0：工具选择](pipeline-overview.md#阶段-0工具选择路径判断) - 决策树和路径判断
- [阶段 1：需求分析](pipeline-overview.md#阶段-1需求分析) - 子Agent分析 + 追溯表
- [阶段 2：需求评审](pipeline-overview.md#阶段-2需求评审) - 3视角并行评审
- [阶段 3：架构设计](pipeline-overview.md#阶段-3架构设计) - 记忆驱动设计
- [阶段 4：Proto变更](pipeline-overview.md#阶段-4proto-变更如需要) - Owner亲自执行
- [阶段 5：编码+测试](pipeline-overview.md#阶段-5编码--测试管线核心) - **核心管线**
- [阶段 6：集成归档](pipeline-overview.md#阶段-6集成验证--归档) - 验证和文档

### 质量保障
- [三层防护体系](pipeline-overview.md#️-质量保障体系三层防护) - 机械化 + CI/CD + AI审查
- [测试金字塔](pipeline-overview.md#-测试体系) - 单元/集成/E2E测试
- [15项机械化检查](pipeline-overview.md#52-qa-agent机械化检查---15-项) - 详细清单

### 技术栈
- [支撑技术](pipeline-overview.md#-支撑技术栈) - Neo4j图谱、Memory系统、MCP工具
- [性能指标](pipeline-overview.md#-流水线性能指标) - 反馈速度、质量指标
- [核心优势](pipeline-overview.md#-核心优势) - 确定性验证、AI+自动化融合

---

## 📊 关键指标一览

| 维度 | 数值 |
|------|------|
| **阶段数** | 6 个（工具选择→需求分析→评审→设计→编码→归档）|
| **质量检查** | 15 项机械化 + 3 视角 AI 审查 |
| **测试覆盖** | 30% 门禁，Logic 文件 100% 有测试 |
| **反馈速度** | Pre-commit 5-15秒，QA 10-30秒 |
| **自动化率** | 格式/测试/分析/QA 100%，审查 95% |
| **CI 通过率** | >95%（有 pre-commit hook 后）|

---

## 🚀 立即开始

```bash
# 1. 安装开发环境 Git Hooks
bash scripts/install-git-hooks.sh

# 2. 验证安装
bash scripts/verify-hooks.sh

# 3. 查看完整文档
cat docs/pipeline-overview.md
```

---

**相关文档**：
- [README.md](../README.md) - 项目快速开始
- [CLAUDE.md](../CLAUDE.md) - Claude Code 协作指南
- [部署指南](deployment-guide.md) - Docker 和 K8s 部署
- [项目编码规范](../.harness/rules/项目编码规范.md) - 硬性约束
