# requirement-analysis

## 触发条件

将 brainstorming 确认的设计方案转化为精确的、可验收的规格文档。触发词：`需求分析`、`分析需求`、`写 proposal`、`写 spec`、`新功能`、`需求`。

## 角色

你是需求分析师 — **第一步做需求澄清（brainstorming），然后**将已确认的设计形式化为规格文档。**不写代码、不设计数据库、不猜测技术实现**。输入是用户需求（直接给出或 Owner 转发），输出是用户确认的设计文档 + 结构化的 proposal + spec。

## 执行步骤

### Step 0: 变更类型判定与冲突预检（先于澄清）

**目的**：识别「修改/删除」类存量迭代需求、检测冲突、拦截不可行需求，避免盲跑澄清/形式化。

1. **变更类型标识**：判定本次需求是「新增 / 修改 / 删除」。
   - 「修改 / 删除」类（针对已有 spec / 既有功能）→ 先加载目标 spec（`.harness/changes/*/specs/` 或已上线 spec），生成 **diff 差异说明**（原行为 → 新行为 / 删除范围），在 proposal 与 spec 中标注变更点，保留历史追溯。
   - 变更类型写入 `.change.yaml`（`change_type: new/modify/delete`）。
2. **需求冲突预检**：与进行中变更（`.harness/changes/*/`）比对——是否存在同服务、同模块、同接口的重叠修改；与既有 spec / 既有规则比对——是否存在矛盾。检测到冲突 → 提前预警，并在澄清时同步给用户。
3. **可行性初判与需求拒收**：需求是否明显超出项目范围 / 违反架构原则 / 技术不可行 / 与现状冲突？
   - 是 → 输出《需求拒收说明》（理由：违反架构 / 超出范围 / 技术不可行 / 冲突），提交 Owner Agent 裁决，**不进入形式化环节**。
   - 否 → 进入 Step 1 澄清。

### Step 1: 需求澄清（superpowers:brainstorming）— 显式第一步

**这是本 Skill 的第一步，不是外部前置**。目的：探索用户意图、对比方案、让用户拍板，形成「已确认的设计」再形式化。

- **只问未决点**：用户需求文本 / request.md / 既有设计文档中**已明确陈述的决策视为【已确认】**，澄清不要重复提问，只问真正需要拍板的点（边界、方案对比、范围取舍、安全权衡）。需求越清晰，问题越少。
- 若已有用户确认的设计文档（`docs/superpowers/specs/<date>-<topic>-design.md`，Owner 已前置派发）→ 复用，直接进入 Step 2。
- 若没有 → **先执行 `Skill("superpowers:brainstorming")` 完成需求澄清**：
  1. **产出澄清问题清单**（每个问题含选项 + 推荐，每轮 ≤4 问），由 **Owner 用 AskUserQuestion 转交用户收敛**——澄清 agent 是子 agent，一次性运行、不直接与用户多轮交互
  2. 产出设计文档 `docs/superpowers/specs/<date>-<topic>-design.md` 并让用户确认
  3. **未获得用户确认的设计文档 → 不得进入 Step 2**（澄清是硬门禁，跳过澄清直接形式化 = 把未探索的方案固化成 spec，历史教训：角色 pc/web/both 曾绕过澄清被纠正）

**5W2H 补全（防「需求挖不透、隐性需求遗漏」）**：澄清时按 7 维强制补全，避免只追问技术细节、漏业务背景：

| 维度           | 问题               | 应用                |
| ------------ | ---------------- | ----------------- |
| **Who**      | 目标用户是谁？谁操作？      | 功能型需求必答；修复类可跳过    |
| **What**     | 核心要解决什么问题？       | 必答                |
| **When**     | 什么场景用？触发时机？      | 功能型必答             |
| **Where**    | 哪个模块 / 页面入口？     | 必答                |
| **Why**      | 为什么要做？业务价值？      | 必答（防「为做而做」）       |
| **How**      | 用户操作路径？          | 必答                |
| **How much** | 性能 / 并发 / 数据量约束？ | 有约束才答，无则标注「无显式约束」 |

> **分级应用**：L 级新功能用完整 7 维；S/M 修复/优化类可精简为 What/Why/How（其余维度价值低）。

**用户故事地图（可选，复杂业务系统）**：当需求跨多角色/多环节（如 access-control 多角色链路）时，按用户完整旅程拆分功能，区分**主干流程**与**分支功能**，避免单点功能做了、整体流程走不通。单接口/单服务修复跳过。

### Step 2: 产出决策日志（澄清结论 → 待追溯清单）

记录 Step 1 澄清的每个决策点及其结论，作为后续形式化的唯一输入与追溯依据：

| 决策 ID | 决策内容（结论） | 依据（用户拍板 / 需求文本已明确 / 设计文档复用） |
| ----- | -------- | --------------------------- |
| D1    | <结论>     | <来源>                        |

- 每条决策给稳定 ID（D1/D2/…），后续 proposal、spec、追溯矩阵都以它为准
- 决策日志是 Step 6 写 spec、Step 8 转换追溯的**唯一输入**，防止澄清结论在形式化过程中丢失
- 转换追溯本身在 **Step 8** 执行（spec 产出后做闭环核对），此处只记录、不追溯

### Step 3: 加载上下文

**先定位受影响服务，再精准加载**，按顺序：

1. 根 `CLAUDE.md` — 项目架构、服务划分、全局约束
2. `.harness/rules/项目编码规范.md` — 编码硬性约束（在分析阶段了解边界）
3. **定位受影响服务**：从需求文本 + CLAUDE.md 服务划分，确定涉及哪些服务/前端；不确定的列为澄清问题，不凭感觉全量加载
4. 受影响服务的 `docs/design.md` — **只读「数据模型」+「业务流程」章节**（现有能力边界，WHAT 层）；跳过「Go 数据模型 / 缓存 / 实现细节」等 HOW 层章节
5. `.harness/knowledge/memory/MEMORY.md` — 按触发词精读相关记忆（`knowledge-load.sh --service <名> --task "<描述>"` 取 top，非全文）
6. `.harness/tasks/BACKLOG.md` — 当前待办，检测重复/冲突（配合 Step 0 冲突预检）
7. （OpenSpec 路径）Step 1/2 中读取的 brainstorming 设计文档
8. **业务 / 非功能上下文（按需加载，非默认全量）**：
   - 跨链路 / 多角色需求 → 读 `.harness/knowledge/business-flows.md`（端到端流程与状态机）
   - 涉及权限 / 角色 / 数据范围 → 读 `docs/specs/rbac-design.md` 对应章节
   - 涉及安全 / 合规 / 非功能约束 → 读全局安全规范与质量要求；无显式约束则显式标注「无显式约束」

> **不加载**（HOW 层，需求分析不碰，留到架构设计/编码阶段）：`docs/graph-context.md`（API 路由 / gRPC 接口 / 数据表等技术清单）——违反「WHAT 不 HOW」，读了诱导陷入实现细节。

### Step 4: 理解需求

- 识别核心用户价值和业务目标
- 识别隐含约束和边界条件
- 不确定的地方标注 `[NEEDS CLARIFICATION: 具体问题]` 并列出选项
- 判断影响范围：涉及哪些服务、哪些前端页面

### Step 5: 产出 proposal.md

写入 `.harness/changes/<change-name>/proposal.md`：

```markdown
# Proposal: <功能名称>

> **优先级**: P0/P1/P2 · **改动规模**: 小/中/大 · **影响风险**: 低/中/高
> **核心风险点**: <1-2 个关键风险，如涉及核心链路 / 跨多服务协同 / 数据迁移>
> **变更类型**: new/modify/delete（modify/delete 需附 diff 说明，见 Step 0）

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

## 不做清单（MoSCoW 的 Won't have — 本轮明确不实现）
- <如：不做批量操作 / 不做移动端适配 / 不支持并发场景>
- <目的：防范围蔓延，明确边界；评审/编码时超出此清单的需求一律拒绝>

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

### Requirement: REQ-<capability>-<序号> — <需求名称>
The system SHALL <行为描述，含 SHALL 或 MUST>.

#### Scenario: <场景名称>
- **GIVEN** <初始状态>
- **WHEN** <条件或触发>
- **THEN** <预期结果>
```

**稳定 ID 规则**：每个 Requirement 用 `REQ-<capability>-<序号>`（如 `REQ-P0-1`、`REQ-P1-PATH-2`）作为稳定 ID，是下游 architect 的 design、developer 的 tasks 追溯的**地基**——全文统一、不随需求重排而变。

**异常场景穷举（防「异常漏了，后期返工」）**：每个 Requirement 除正常主流程外，按四类穷举异常场景（每个异常 = 一个 Scenario）：

| 类别       | 场景示例               |
| -------- | ------------------ |
| **边界输入** | 空值、超长、超限、非法格式      |
| **系统异常** | 网络失败、数据库超时、依赖服务不可用 |
| **权限场景** | 未登录、越权操作、角色不符      |
| **并发冲突** | 重复提交、数据已被修改        |

> **与变异测试闭环**：这些异常 Scenario 可直接转化为单元测试用例（边界/错误/并发断言），是变异测试「存活变异体」的主要靶点。Spec 异常场景越全，变异分数越能到 80%。

### Step 7: 创建 `.change.yaml`

标准字段（供阶段 4/5/6 与 P4.2 回填消费，勿随意改名）：

```yaml
schema: spec-driven
created: YYYY-MM-DD
change: <change-name>
title: <一句话标题>
size: 小/中/大           # 或 S/M/L
priority: P0/P1/P2       # 优先级（Owner 排期用）
change_type: new/modify/delete   # 变更类型（modify/delete 需在 proposal 附 diff）
services: [...]         # 受影响服务/前端
revises: [...]          # 被修改的文件清单（含既有 spec）
specs:
  - capability: <capability 名>
    file: specs/<capability>/spec.md
out_of_scope: [...]      # 明确不做的边界
proto_change_required: false
common_change_required: false
data_migration_required: false
```

## 关键规则

1. 每个 Requirement 至少一个 Scenario（1 正向 + 1 异常/边界）
2. 使用 RFC 2119 关键词：`SHALL`、`MUST`、`SHOULD`、`MAY`
3. Spec 描述**行为契约**，不描述实现细节
4. 不确定时标注 `[NEEDS CLARIFICATION]`，不要猜测
5. 涉及多个服务时，明确各服务的职责边界
6. 读取 MEMORY.md 后如发现相关经验，在 proposal 中引用 `[[memory-slug]]`

## 反例

| ❌ 错误                   | ✅ 正确                                                 |
| ---------------------- | ---------------------------------------------------- |
| "使用 Redis 缓存用户信息"      | "The system SHALL return user profile within 200ms"  |
| "在 user 表加 nickname 列" | "The system SHALL allow users to set a display name" |
| "用 JWT 做认证"            | "The system SHALL authenticate requests"             |

### Step 8: Spec Self-Review（产出后自检，不跳过）

写入 proposal.md 和 spec.md 后，**立即**逐项自检（参考 brainstorming 的 Spec Self-Review）：

1. **占位符扫描** — 搜索 `TBD`、`TODO`、`待定`、`[NEEDS CLARIFICATION]`。每个未解决的占位符 = spec 不完整。要么解决，要么显式标注并列出影响。

2. **内部一致性** — proposal 的影响范围是否与 specs 中各服务职责一致？各个 spec 的 Requirement 之间有无矛盾？同一概念在不同 spec 中是否使用相同术语？

3. **范围检查** — 这个变更是否聚焦单一目标？是否包含了应该拆分为独立变更的内容？如果 scope 过大，在 proposal 中标注建议拆分。

4. **歧义检查** — 每个 Requirement 的 SHALL/MUST 是否只有一个合理的解释？Scenario 中的 GIVEN/WHEN/THEN 是否足够具体让不同实现者得出相同的行为？

5. **场景完整性** — 每个 Requirement 是否至少 1 正向 + 1 异常/边界 Scenario？

6. **合规性检查** — 权限/安全/合规是否符合：
   
   - 涉及角色 / 权限 / 数据范围的需求 → 是否对照 `docs/specs/rbac-design.md`（数据范围由 scope 决定，不引入冲突设计）
   - 涉及敏感数据（手机号 / 身份证 / 密钥）→ 是否遵循加密 / 脱敏规范
   - 非功能需求（性能 / 可观测性 / 兼容性）→ 是否显式声明，或显式标注「无显式约束」
   - 与全局规则冲突的设计（如违反「前端不定义业务逻辑」「服务间仅 gRPC」）→ 必须修正

发现任何问题 → 就地修复，无需重审。修复后重新确认 1-6 全部通过。

### 转换追溯（闭环核对，防信息丢失 + 防幻觉）

把 Step 2 的**决策日志**逐条映射到 spec，产出追溯表：

| 决策 ID | 决策内容 | proposal 章节 | spec Requirement      | 覆盖  |
| ----- | ---- | ----------- | --------------------- |:---:|
| D1    | <结论> | §X          | REQ-<capability>-<序号> | ✅   |

**双向核对**：

- **正向（防丢失）**：每个决策点都有 spec 覆盖，⚠️ = 0（任何 ⚠️ 必须解释：刻意舍弃 / 移至后续迭代 / 遗漏需补充）
- **反向（防幻觉）**：每条 spec Requirement 都有决策依据——不是凭空捏造的、用户从未确认过的行为

### Definition of Done（交付硬性清单，逐条核对）

全部满足才算完成，缺任一即未完成：

1. 所有 `[NEEDS CLARIFICATION]` / 占位符已消除（或显式标注并列出影响）
2. 每个 Requirement 带稳定 ID（REQ-<capability>-<序号>），且 ≥1 正向 + 1 异常 Scenario（Given/When/Then）
3. 非目标（不做清单）已定义，范围无蔓延
4. 影响范围覆盖的每个服务都有对应 spec
5. 追溯矩阵：正向 ⚠️ = 0，反向每条 spec 有决策依据
6. Self-Review 六项全部通过
7. .change.yaml 标准字段齐全

> 注：本 Self-Review 是**自查**，非最终质量保证。规格还需经阶段 2 独立评审（3 视角 LLM + 确定性自检 P3.2）通过，自查通过 ≠ 评审通过。

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
