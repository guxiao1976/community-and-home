# requirement-analysis

## 触发条件

将 brainstorming 确认的设计方案转化为精确的、可验收的规格文档。触发词：`需求分析`、`分析需求`、`写 proposal`、`写 spec`、`新功能`、`需求`。

## 角色

你是需求分析师 — **第一步做需求澄清（brainstorming），然后**将已确认的设计形式化为规格文档。**不写代码、不设计数据库、不猜测技术实现**。输入是用户需求（直接给出或 Owner 转发），输出是用户确认的设计文档 + 结构化的 proposal + spec。

## 执行步骤

### Step 1: 需求澄清（superpowers:brainstorming）— 显式第一步

**这是本 Skill 的第一步，不是外部前置**。目的：探索用户意图、对比方案、让用户拍板，形成「已确认的设计」再形式化。

- 若已有用户确认的设计文档（`docs/superpowers/specs/<date>-<topic>-design.md`，Owner 已前置派发）→ 复用，直接进入 Step 2。
- 若没有 → **先执行 `Skill("superpowers:brainstorming")` 完成需求澄清**：
  1. 交互式提问（AskUserQuestion），逐项澄清：业务目标、边界、方案对比、偏好
  2. 产出设计文档 `docs/superpowers/specs/<date>-<topic>-design.md` 并让用户确认
  3. **未获得用户确认的设计文档 → 不得进入 Step 2**（澄清是硬门禁，跳过澄清直接形式化 = 把未探索的方案固化成 spec，历史教训：角色 pc/web/both 曾绕过澄清被纠正）

### Step 2: 转换追溯（brainstorming 决策 → spec 覆盖）

阅读 Step 1 的设计文档（或直接给出的需求）：
1. 理解已确认的设计决策、方案对比结论、用户偏好
2. 将设计要点映射到 proposal 和 spec 的对应章节
3. **转换追溯（防信息丢失）** — 产出后立即自检，列出 brainstorming 关键决策 → spec 的覆盖表：

| brainstorming 设计决策 | proposal 章节 | spec Requirement | 覆盖 |
|----------------------|-------------|-----------------|:---:|
| <决策1>               | §X           | REQ-XX          | ✅ |
| <决策2>               | §Y           | REQ-YY          | ⚠️ (原因) |

任何 ⚠️ 项必须解释：刻意舍弃 / 移至后续迭代 / 遗漏需补充。
全部 ✅ 才能进入 Step 3。

### Step 3: 加载上下文


按顺序：
1. 根 `CLAUDE.md` — 项目架构、服务划分、全局约束
2. `.harness/rules/项目编码规范.md` — 编码硬性约束（在分析阶段了解边界）
3. 相关服务的 `docs/design.md` — 现有数据模型和业务流程
4. `.harness/knowledge/memory/MEMORY.md` — 精读相关记忆，避免提出已知不可行的方案
5. （OpenSpec 路径）Step 1/2 中读取/追溯的 brainstorming 设计文档

### Step 4: 理解需求

- 识别核心用户价值和业务目标
- 识别隐含约束和边界条件
- 不确定的地方标注 `[NEEDS CLARIFICATION: 具体问题]` 并列出选项
- 判断影响范围：涉及哪些服务、哪些前端页面

### Step 5: 产出 proposal.md

写入 `.harness/changes/<change-name>/proposal.md`：

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

### Step 6: 产出 spec.md（每个功能一个）

写入 `.harness/changes/<change-name>/specs/<capability>/spec.md`：

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

### Step 7: 创建 `.change.yaml`

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

### Step 8: Spec Self-Review（产出后自检，不跳过）

写入 proposal.md 和 spec.md 后，**立即**逐项自检（参考 brainstorming 的 Spec Self-Review）：

1. **占位符扫描** — 搜索 `TBD`、`TODO`、`待定`、`[NEEDS CLARIFICATION]`。每个未解决的占位符 = spec 不完整。要么解决，要么显式标注并列出影响。

2. **内部一致性** — proposal 的影响范围是否与 specs 中各服务职责一致？各个 spec 的 Requirement 之间有无矛盾？同一概念在不同 spec 中是否使用相同术语？

3. **范围检查** — 这个变更是否聚焦单一目标？是否包含了应该拆分为独立变更的内容？如果 scope 过大，在 proposal 中标注建议拆分。

4. **歧义检查** — 每个 Requirement 的 SHALL/MUST 是否只有一个合理的解释？Scenario 中的 GIVEN/WHEN/THEN 是否足够具体让不同实现者得出相同的行为？

5. **场景完整性** — 每个 Requirement 是否至少 1 正向 + 1 异常/边界 Scenario？

发现任何问题 → 就地修复，无需重审。修复后重新确认 1-4 全部通过。

## 产出物

```
.harness/changes/<change-name>/
├── .change.yaml
├── proposal.md
└── specs/
    └── <capability>/
        └── spec.md
```

## 下一步

产出完成后：
- 走阶段 2 需求评审（`.harness/skills/review.md` 计划评审模式），验证 spec + tasks 合理性
- 评审通过后进入 `architect-design` Skill 进行技术设计

## 关联

- **Step 1（显式第一步）**: `Skill("superpowers:brainstorming")` — 需求澄清+方案探索，无设计文档则先执行，产出用户确认的设计文档（硬门禁）
- 架构设计：`.harness/skills/architect-design.md`
- 经验记忆：`.harness/knowledge/memory/MEMORY.md`
- 需求模板：`docs/requirements/TEMPLATE.md`
