# OpenSpec - 提案驱动开发流程

## 概述

OpenSpec 是本项目采用的提案驱动开发（Proposal-Driven Development）流程，用于规范化需求管理、设计评审和实施跟踪。

## 目录结构

```
openspec/
├── config.yaml              # 项目级配置（上下文、技术栈、规则）
├── README.md               # 本文档
├── changes/                # 提案变更目录
│   └── {proposal-name}/    # 单个提案目录
│       ├── .openspec.yaml  # 提案元数据
│       ├── proposal.md     # 提案说明（Why、What、Capabilities）
│       ├── design.md       # 设计文档（可选）
│       ├── tasks.md        # 任务拆分
│       ├── specs/          # 详细规格（可选）
│       │   └── {spec-name}/
│       │       └── spec.md
│       ├── IMPLEMENTATION_SUMMARY.md  # 实施总结（完成后）
│       └── TEST_REPORT.md  # 测试报告（完成后）
└── specs/                  # 全局规格库（可选）
```

## 工作流程

### 1. 创建提案

在 `openspec/changes/` 下创建新目录，命名使用 kebab-case：

```bash
mkdir -p openspec/changes/feature-name
cd openspec/changes/feature-name
```

### 2. 编写提案文件

#### `.openspec.yaml` - 提案元数据

```yaml
schema: spec-driven
created: 2026-05-14
status: draft          # draft | in-progress | completed | cancelled
assignee: username     # 可选
priority: high         # high | medium | low
```

#### `proposal.md` - 提案说明

```markdown
## Why
说明业务背景和问题（为什么要做这个变更）

## What Changes
列出具体的变更内容：
- 后端变更（API、数据库、服务）
- 前端变更（页面、组件、交互）
- 配置变更
- 文档变更

## Capabilities
### New Capabilities
- `capability-name`: 描述新增的能力

### Modified Capabilities
- `existing-capability`: 描述对现有能力的修改
```

#### `tasks.md` - 任务拆分

```markdown
# 任务清单

## 阶段1：准备工作
- [ ] 任务1 - 预估时间：30分钟
- [ ] 任务2 - 预估时间：1小时

## 阶段2：后端开发
- [ ] 数据库迁移 - 预估时间：1小时
- [ ] API实现 - 预估时间：2小时
- [ ] 单元测试 - 预估时间：1小时

## 阶段3：前端开发
- [ ] API接口封装 - 预估时间：30分钟
- [ ] 组件开发 - 预估时间：2小时
- [ ] 集成测试 - 预估时间：1小时
```

### 3. 实施开发

按照 `tasks.md` 的任务清单逐步实施，更新任务状态。

### 4. 完成总结

开发完成后，创建总结文档：

#### `IMPLEMENTATION_SUMMARY.md` - 实施总结

```markdown
# 实施总结

## 完成时间
2026-05-14

## 实际变更
- 列出实际完成的变更
- 与提案的差异说明

## 技术决策
- 关键技术选型
- 架构调整
- 性能优化

## 遗留问题
- 待优化项
- 技术债务
```

#### `TEST_REPORT.md` - 测试报告

```markdown
# 测试报告

## 测试范围
- 功能测试
- 集成测试
- 性能测试

## 测试结果
- 通过率：100%
- 发现问题：0个

## 测试用例
| 用例 | 预期结果 | 实际结果 | 状态 |
|------|---------|---------|------|
| ... | ... | ... | ✅ |
```

## 配置规则

### config.yaml 结构

```yaml
schema: spec-driven

# 项目上下文（技术栈、架构、约定）
context: |
  项目概述、技术栈、架构模式、代码规范等

# 提案规则
rules:
  proposal:
    - 提案应明确说明业务价值和技术影响
    - 长度控制在500字以内
    
  tasks:
    - 任务拆分遵循单一职责
    - 每个任务不超过2小时
    
  implementation:
    - 强制性开发规范
    - 测试要求
    - 代码审查标准
```

## 最佳实践

### 1. 提案命名

使用清晰的 kebab-case 命名：
- ✅ `add-user-batch-import`
- ✅ `fix-login-timeout-issue`
- ✅ `refactor-api-response-format`
- ❌ `feature1`
- ❌ `bug_fix`

### 2. 提案粒度

- **小提案**（1-2天）：单个功能、bug修复
- **中提案**（3-5天）：模块开发、重构
- **大提案**（1-2周）：新服务、架构调整

### 3. 任务拆分原则

- 每个任务独立可测试
- 任务间依赖关系明确
- 包含回滚方案（数据库变更）
- 预估时间准确

### 4. 文档维护

- 提案完成后及时更新状态
- 实施过程中的重要决策记录在 design.md
- 测试结果记录在 TEST_REPORT.md
- 经验教训记录在 IMPLEMENTATION_SUMMARY.md

## 与其他文档的关系

| 文档 | 用途 | 更新时机 |
|------|------|---------|
| `openspec/config.yaml` | 项目级规则和上下文 | 新增通用规范时 |
| `CLAUDE.md` | 快速参考指南 | 关键规范变更时 |
| `PROJECT_STRUCTURE.md` | 项目结构详解 | 目录结构变化时 |
| `openspec/changes/*/` | 具体提案和实施 | 每个提案独立维护 |

## 工具集成

### Claude Code 集成

Claude Code 会自动读取 `openspec/config.yaml` 作为项目上下文，在生成提案、拆分任务、编写代码时遵循配置的规则。

### Git 工作流

```bash
# 创建提案分支
git checkout -b proposal/feature-name

# 提交提案文档
git add openspec/changes/feature-name/
git commit -m "proposal: add feature-name"

# 实施开发
git add services/ web/
git commit -m "feat: implement feature-name"

# 完成总结
git add openspec/changes/feature-name/IMPLEMENTATION_SUMMARY.md
git commit -m "docs: add implementation summary for feature-name"
```

## 常见问题

### Q: 什么时候需要创建提案？

A: 以下情况建议创建提案：
- 新功能开发（超过半天工作量）
- 架构调整或重构
- 数据库schema变更
- API接口变更
- 跨服务协作

### Q: 小bug修复需要提案吗？

A: 不需要。简单的bug修复（1小时内）可以直接提交代码，在commit message中说明即可。

### Q: 提案可以修改吗？

A: 可以。在实施过程中发现提案不合理，应该：
1. 更新 proposal.md 说明变更原因
2. 调整 tasks.md 任务清单
3. 在 IMPLEMENTATION_SUMMARY.md 中记录差异

### Q: 如何处理紧急需求？

A: 紧急需求可以简化流程：
1. 创建最小化提案（只有 proposal.md）
2. 快速实施
3. 事后补充完整文档

## 参考示例

查看以下提案作为参考：
- `openspec/changes/ai-model-service/` - 完整的服务开发提案
- `openspec/changes/sensitive-word-ui-optimization/` - UI优化提案
- `openspec/changes/add-model-masterdata-review/` - 跨模块功能提案
