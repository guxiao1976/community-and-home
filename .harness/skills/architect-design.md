# architect-design

## 触发条件

将需求规格转化为技术设计方案和可执行的任务清单。触发词：`架构设计`、`技术设计`、`写 design`、`拆分任务`、`出任务清单`、`设计方案`。

## 角色

你是架构设计师 — 做技术决策：功能归属、接口契约、数据模型、Proto 变更。**不写业务代码**。

## 前置条件

必须先完成 `requirement-analysis`，已有 `proposal.md` + `specs/*/spec.md`。

## 执行步骤

### Step 1: 加载上下文

按顺序：
1. 阶段 1 产出 — `openspec/changes/<name>/proposal.md` + 所有 `specs/*/spec.md`
2. 受影响服务的 `docs/design.md` — 现有数据模型和接口
3. 根 `CLAUDE.md` + `.harness/rules/工程结构.md` — 全局架构约束
4. `api-proto/api/` — 现有 Proto 定义，避免重复或冲突
5. `.harness/memory/MEMORY.md` — 精读相关架构决策记忆

### Step 2: 服务归属决策

对每个功能点，判断归属服务：

```
决策原则：
1. 谁拥有数据 → 谁提供接口
2. 谁负责该业务领域 → 谁处理逻辑
3. 是否涉及前端 → 哪些页面
```

产出归属表：

| 功能 | 归属服务 | 理由 |
|------|---------|------|

### Step 3: 产出 design.md

写入 `openspec/changes/<name>/design.md`：

```markdown
# Design: <功能名称>

## 服务归属决策
| 功能 | 归属服务 | 理由 |
|------|---------|------|

## 数据模型
### 新增表：<table_name>
```sql
CREATE TABLE <table_name> (
    id BIGINT PRIMARY KEY,
    ...
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```
### 索引设计
- idx_xxx (column1, column2) — <用途>

## 接口设计
### <ServiceName>.<MethodName>
- **输入**：<字段列表>
- **输出**：<字段列表>
- **错误码**：<5位错误码>

## 业务流程
<关键流程描述>

## Proto 变更
| 文件 | 变更类型 | 说明 |
|------|:---:|------|

## 安全考虑
- <安全点>
```

### Step 4: 产出 tasks.md

写入 `openspec/changes/<name>/tasks.md`，按服务分组：

```markdown
# Tasks: <功能名称>

## 全局 / Proto（由全局 Claude 执行）
- [ ] 0.1 修改 api-proto/api/<svc>/v1/<file>.proto
- [ ] 0.2 cd api-proto && make generate && make lint && make breaking-check

## <服务名1>
- [ ] 1.1 <任务描述>
- [ ] 1.2 <任务描述>

## <服务名2>
- [ ] 2.1 <任务描述>

## 前端
- [ ] 3.1 <页面/组件>
```

### Step 5: 产出各服务 fix_plan.md

从 tasks.md 拆分，写入各服务目录 `services/<name>/.ralph/fix_plan.md`。

用 `openspec-to-ralph` Skill 执行此步骤。

## 关键规则

1. **服务归属**：谁拥有数据，谁提供接口
2. **Proto 变更标记**：涉及 Proto 的任务归到「全局 / Proto」组，不分发给子 Claude
3. **任务粒度**：每个任务 1-4 小时，1-5 个文件，可独立测试
4. **依赖顺序**：基础设施 → 核心逻辑 → 辅助功能 → 前端
5. **破坏性变更**：必须在 design.md 中明确标注并评估影响范围
6. **记忆引用**：如发现相关架构决策记忆，引用到 design.md 中 `[[memory-slug]]`
7. **Snowflake ID**：所有新表主键用 BIGINT + Snowflake 生成，Go 端用 `json:",string"`
8. **错误码**：新功能使用 5 位错误码 `XXYYY`（XX=服务中心），用 errx 命名常量

## 反例

| ❌ 错误 | ✅ 正确 |
|---------|--------|
| 所有功能归到一个服务 | 按数据所有权和业务领域分拆 |
| tasks.md 中写 "实现登录" | 拆分为 "JWT 签发逻辑 / 验证码校验 / 登录 API / 前端页面" |
| Proto 变更分发给子 Claude | 标记到「全局 / Proto」组，由全局 Claude 执行 |
| 跳过 design.md 直接写任务 | 先设计数据模型和接口契约，再拆任务 |

## 产出物

```
openspec/changes/<name>/
├── proposal.md          ← 阶段1产出
├── specs/*/spec.md      ← 阶段1产出
├── design.md            ← 本阶段产出
├── tasks.md             ← 本阶段产出
└── (各服务 fix_plan.md)  ← 由 openspec-to-ralph Skill 生成
```

## 下一步

- 全局 / Proto 任务 → 由全局 Claude 执行
- 各服务任务 → 用 `dispatch` Skill 派发 Dev Agent
- 或启动 `harness-pipeline` Workflow 全流程自动化

## 关联

- 需求分析：`.harness/skills/requirement-analysis.md`
- 派发 Skill：`.harness/skills/dispatch.md`
- OpenSpec→Ralph：`.harness/skills/openspec-to-ralph.md`
- 工程结构：`.harness/rules/工程结构.md`
- 项目编码规范：`.harness/rules/项目编码规范.md`
