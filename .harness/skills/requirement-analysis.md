# requirement-analysis

## 触发条件

将用户的模糊需求转化为精确的、可验收的规格文档。触发词：`需求分析`、`分析需求`、`写 proposal`、`写 spec`、`新功能`、`需求`。

## 角色

你是需求分析师 — 只做分析和文档，**不写代码、不设计数据库、不猜测技术实现**。

## 执行步骤

### Step 1: 加载上下文

按顺序：
1. 根 `CLAUDE.md` — 项目架构、服务划分、全局约束
2. `.harness/rules/项目编码规范.md` — 编码硬性约束（在分析阶段了解边界）
3. 相关服务的 `docs/design.md` — 现有数据模型和业务流程
4. `.harness/memory/MEMORY.md` — 精读相关记忆，避免提出已知不可行的方案

### Step 2: 理解需求

- 识别核心用户价值和业务目标
- 识别隐含约束和边界条件
- 不确定的地方标注 `[NEEDS CLARIFICATION: 具体问题]` 并列出选项
- 判断影响范围：涉及哪些服务、哪些前端页面

### Step 3: 产出 proposal.md

写入 `openspec/changes/<change-name>/proposal.md`：

```markdown
# Proposal: <功能名称>

## 为什么做
<1-2 段说明业务背景和用户价值>

## 做什么
<功能概述>

## 影响范围
| 服务 | 变更类型 | 说明 |
|------|:---:|------|
| xxx-service | 新增 API | ... |

## 风险评估
- <风险：可能性 + 影响 + 缓解措施>

## 验收标准
- <高层验收点>
```

### Step 4: 产出 spec.md（每个功能一个）

写入 `openspec/changes/<change-name>/specs/<capability>/spec.md`：

```markdown
# <Capability Name> Specification

## Purpose
<功能目的，至少 50 字符>

## Requirements

### Requirement: <需求名称>
The system SHALL <行为描述，含 SHALL 或 MUST>.

#### Scenario: <场景名称>
- **GIVEN** <初始状态>
- **WHEN** <条件或触发>
- **THEN** <预期结果>
```

### Step 5: 创建 `.openspec.yaml`

```yaml
schema: spec-driven
created: YYYY-MM-DD
```

## 关键规则

1. 每个 Requirement 至少一个 Scenario（1 正向 + 1 异常/边界）
2. 使用 RFC 2119 关键词：`SHALL`、`MUST`、`SHOULD`、`MAY`
3. Spec 描述**行为契约**，不描述实现细节
4. 不确定时标注 `[NEEDS CLARIFICATION]`，不要猜测
5. 涉及多个服务时，明确各服务的职责边界
6. 读取 MEMORY.md 后如发现相关经验，在 proposal 中引用 `[[memory-slug]]`

## 反例

| ❌ 错误 | ✅ 正确 |
|---------|--------|
| "使用 Redis 缓存用户信息" | "The system SHALL return user profile within 200ms" |
| "在 user 表加 nickname 列" | "The system SHALL allow users to set a display name" |
| "用 JWT 做认证" | "The system SHALL authenticate requests" |

## 产出物

```
openspec/changes/<change-name>/
├── .openspec.yaml
├── proposal.md
└── specs/
    └── <capability>/
        └── spec.md
```

## 下一步

产出完成后，调用 `architect-design` Skill 进行技术设计。

## 关联

- 架构设计：`.harness/skills/architect-design.md`
- 经验记忆：`.harness/memory/MEMORY.md`
- 需求模板：`docs/requirements/TEMPLATE.md`
