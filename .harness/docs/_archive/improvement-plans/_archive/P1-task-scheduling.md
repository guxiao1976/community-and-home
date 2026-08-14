# P1-1: 任务调度策略优化

## 背景

**现状问题**：
- 当前调度策略：按 P0→P1→P2→P3 线性排序（`owner-agent.md:382-395`）
- **长尾任务饥饿**：P2/P3 任务永远排在队尾，如果 P0/P1 持续产生，P2/P3 永远不会被执行
- **依赖关系未考虑**：任务间有 `blocked_by` 关系，但调度器未识别依赖链
- **年龄未考虑**：创建 30 天的 P2 任务 vs 创建 1 天的 P1 任务，前者优先级反而更低

**影响**：
- 技术债务累积（P2/P3 多为 debt 类型）
- 代码质量逐渐下降（规范问题、重构需求被搁置）
- 团队士气下降（长期任务无进展）

## 目标

实现 **加权优先级调度 + 依赖感知**，解决长尾任务饥饿问题，同时确保高优先级任务不被延误。

## 技术方案

### 1. 加权优先级公式

**基础公式**：
```
score = priority_weight × (1 + age_factor) × urgency_multiplier × dependency_boost

其中：
- priority_weight: P0=100, P1=50, P2=20, P3=10
- age_factor: min(age_days × 0.05, 2.0)  # 年龄加成，最多翻倍
- urgency_multiplier: 
    - source=qa|review 且 severity=CRITICAL → 1.5
    - source=sensor 且类型=security → 2.0
    - 其他 → 1.0
- dependency_boost:
    - 被其他任务 blocked_by → 1.2（解除阻塞能释放下游）
    - 无依赖 → 1.0
```

**示例计算**：

| 任务 | Priority | Age | Source | 依赖者数 | 计算 | 最终分数 |
|------|---------|-----|--------|---------|------|---------|
| A | P0 | 1天 | human | 0 | 100 × (1+0.05) × 1.0 × 1.0 | **105** |
| B | P1 | 30天 | qa, CRITICAL | 0 | 50 × (1+1.5) × 1.5 × 1.0 | **187.5** |
| C | P2 | 60天 | debt | 2 | 20 × (1+2.0) × 1.0 × 1.2 | **72** |
| D | P3 | 90天 | sensor | 0 | 10 × (1+2.0) × 1.0 × 1.0 | **30** |

**排序结果**：B (187.5) > A (105) > C (72) > D (30)

**效果**：30 天高严重性 P1 任务超过新建 P0，60 天 P2 任务超过新建 P1，避免饥饿。

### 2. 依赖链分析

**问题**：任务 A blocked_by B，B blocked_by C → 必须先执行 C，再 B，再 A

**解决方案**：构建任务 DAG（有向无环图），拓扑排序

**实现**：`.harness/scripts/harness-tasks.sh` 新增 `--order-by-deps` 选项

```bash
# 伪代码
function build_dependency_graph() {
  # 读取所有 open 任务
  TASKS=$(list_all_open_tasks)
  
  # 解析 blocked_by 关系
  declare -A GRAPH  # task_id -> [blocked_by_ids]
  for task in $TASKS; do
    BLOCKED_BY=$(grep "^blocked_by:" "$task" | cut -d: -f2)
    GRAPH[$task]="$BLOCKED_BY"
  done
  
  # 拓扑排序（Kahn 算法）
  # 1. 找出所有入度为 0 的节点（无依赖）
  # 2. 输出这些节点
  # 3. 删除这些节点及其出边
  # 4. 重复直到所有节点输出
  
  # 有环 → 检测并报告 circular dependency
}

# 使用
bash .harness/scripts/harness-tasks.sh list --status open --order-by-deps
```

**输出示例**：
```
# 拓扑排序后的任务队列
task-2026-06-16-003  # 无依赖，优先
task-2026-06-16-005  # blocked_by 003，等待 003 完成
task-2026-06-16-008  # blocked_by 005，等待 005 完成
```

### 3. 调度器状态机

**Owner Agent 调度逻辑**（`owner-agent.md:382-395` 替换）

```markdown
### 主动任务发现（Backlog 驱动）

#### Step 1: 扫描新问题
\`\`\`bash
bash .harness/scripts/harness-tasks.sh scan --auto-create
\`\`\`

#### Step 2: 计算任务优先级分数
\`\`\`bash
# 对所有 status=open 的任务计算 score
bash .harness/scripts/harness-tasks.sh score --update-index

# 输出示例（BACKLOG.md 自动更新）
# | Task ID | Priority | Age | Score | Status |
# |---------|----------|-----|-------|--------|
# | task-001 | P1 | 30天 | 187.5 | open |
# | task-002 | P0 | 1天  | 105.0 | open |
\`\`\`

#### Step 3: 拓扑排序（依赖感知）
\`\`\`bash
bash .harness/scripts/harness-tasks.sh list --status open --order-by-deps --order-by-score
\`\`\`

输出：依赖关系合法的任务队列，按 score 降序排列

#### Step 4: 选择执行任务
**规则**：
1. 取队列前 N 个任务（N = 并发度，默认 3）
2. 过滤：
   - 有 blocked_by 且依赖未关闭 → 跳过
   - source=human 且无 assignee → 询问是否执行
3. 启动 harness-pipeline 或 dispatch

#### Step 5: 更新任务状态
\`\`\`bash
bash .harness/scripts/harness-tasks.sh status --id <id> --status in_progress
\`\`\`

#### Step 6: 完成后更新
\`\`\`bash
bash .harness/scripts/harness-tasks.sh status --id <id> --status closed
# 检查是否有任务被该任务 block → 通知（任务解锁）
\`\`\`
```

### 4. 动态优先级调整

**场景**：任务执行过程中发现新的 CRITICAL 问题 → 动态插入高优先级任务

**实现**：任务状态字段增加 `boost` 标记

```markdown
---
id: task-2026-06-18-010
priority: P2
boost: true  # ← 新增，表示手动提升优先级
boost_reason: "生产环境发现相关 bug，需优先修复"
---
```

**计算**：`boost=true` → `priority_weight × 3`（临时提升到 P0 级别）

### 5. 调度可视化

**新增命令**：`bash .harness/scripts/harness-tasks.sh timeline`

**输出**：Gantt 图 ASCII 版本

```
任务时间线（预估完成时间）
════════════════════════════════════════════════════════
task-001 [P1, score=187.5] ████████░░░░░░░░░░  (2h, in_progress)
task-002 [P0, score=105.0] ░░░░░░░░░░░░░░░░░░  (等待 task-001)
task-003 [P2, score=72.0]  ░░░░░░░░░░░░░░░░░░  (队列中)

图例: █ 进行中  ░ 等待执行
```

## 实施步骤

### Phase 1: 评分系统（1.5 天）

**Task 1.1**: 实现 `harness-tasks.sh score` 子命令
- 读取任务元数据（priority / created / source / severity）
- 计算 dependency_boost（统计有多少任务 blocked_by 该任务）
- 写入任务文件的 frontmatter：`computed_score: 187.5`
- 更新 BACKLOG.md，按 score 降序排列

**Task 1.2**: 单元测试
- 准备测试任务集（不同 priority / age / source 组合）
- 验证评分公式正确性
- 边界测试（age=0、age=1000、circular dependency）

**Task 1.3**: 调优权重参数
- 收集现有 37 个任务的优先级分布
- 模拟调度顺序，与人工预期对比
- 调整 age_factor / urgency_multiplier 参数

### Phase 2: 依赖分析（1.5 天）

**Task 2.1**: 实现拓扑排序
- 解析 `blocked_by: [task-xxx, task-yyy]`
- Kahn 算法实现
- 环检测 → 输出 circular dependency 警告

**Task 2.2**: 依赖可视化
- `harness-tasks.sh deps --id task-xxx`
- 输出依赖树（ASCII 树形结构）
```
task-001
├─ blocked_by: task-002
│  └─ blocked_by: task-003 ✅ closed
└─ blocks: [task-005, task-007]
```

**Task 2.3**: 自动依赖推断
- QA/Review 发现问题 → 自动创建任务时，推断依赖关系
- 例：Review 发现"缺少单元测试" → 创建任务 A
- 同时发现"函数逻辑错误" → 创建任务 B
- 自动推断：A blocked_by B（先修逻辑再补测试）

### Phase 3: Owner Agent 集成（1 天）

**Task 3.1**: 更新调度流程
- `owner-agent.md` 的「主动任务发现」章节
- 替换为新的 6-Step 流程

**Task 3.2**: 并发控制
- 增加配置：`MAX_CONCURRENT_TASKS=3`
- 调度器同时执行不超过 3 个任务
- 一个完成 → 从队列取下一个

**Task 3.3**: 人工确认逻辑
- source=human 任务 → 输出任务描述 + 询问是否执行
- 用户拒绝 → 跳过，执行下一个

### Phase 4: 测试验证（1 天）

**Task 4.1**: 模拟长尾饥饿场景
- 创建 10 个 P0/P1 任务（持续产生）
- 创建 5 个 P2 任务（age=60天）
- 运行调度器 10 轮
- 验证：P2 任务在第 3-5 轮被执行（未饥饿）

**Task 4.2**: 依赖链测试
- 创建依赖链：A → B → C → D
- 运行调度器
- 验证执行顺序：D → C → B → A

**Task 4.3**: 环检测测试
- 创建环：A blocked_by B, B blocked_by A
- 运行调度器
- 验证输出警告 + 跳过这些任务

### Phase 5: 文档和上线（0.5 天）

**Task 5.1**: 更新文档
- `owner-agent.md` 补充新调度逻辑
- `MAINTENANCE.md` 补充调度策略说明

**Task 5.2**: 创建 Memory
- `.harness/knowledge/memory/task-scheduling-strategy.md`

**Task 5.3**: 上线监控
- 前 20 个任务执行：记录调度顺序
- 收集反馈：是否有不合理的任务被延后？

## 验收标准

### 功能验收

- [ ] 评分系统计算结果符合公式
- [ ] 拓扑排序输出依赖合法的执行顺序
- [ ] 环检测正确识别并报警
- [ ] 长尾任务（age>30天）不会饥饿

### 性能验收

- [ ] 100 个任务的评分计算 < 1 秒
- [ ] 拓扑排序（100 个任务，最大依赖深度 10）< 2 秒

### 质量验收

- [ ] 10 轮模拟调度，P2 任务（age=60天）在前 5 轮被执行
- [ ] 依赖链 A→B→C，执行顺序 100% 正确
- [ ] 环检测准确率 100%

## 风险和依赖

### 风险

**R1: 权重参数不合理**
- **描述**：age_factor 过大 → 所有旧任务优先，新 P0 被延误
- **缓解**：上线后持续调优，收集反馈

**R2: 循环依赖无法自动解决**
- **描述**：A blocked_by B, B blocked_by A → 需人工介入
- **缓解**：环检测 → 通知用户 + 建议断开哪条边

**R3: 评分波动频繁**
- **描述**：每天 age +1 → score 变化 → BACKLOG.md 频繁更新
- **缓解**：score 缓存 24 小时，每天只重新计算一次

### 依赖

**D1: 任务元数据完整性**
- blocked_by 字段必须准确
- 行动：审查现有任务，补全缺失的依赖关系

**D2: 并发执行能力**
- Owner Agent 需支持并发调度多个 Workflow
- 当前已支持（`parallel()` 函数）

## 效果预估

| 指标 | 现状 | 改进后 | 提升 |
|------|------|--------|------|
| P2 任务平均等待时间 | 无限（饥饿） | <7 天 | - |
| P3 任务平均等待时间 | 无限（饥饿） | <30 天 | - |
| 高优先级任务响应时间 | 即时 | <2 天（被旧 P1 超越） | 略降但可接受 |
| 技术债务累积速度 | +5 个/月 | -2 个/月（净消化） | - |

## 后续优化

1. **机器学习预测执行时间**：基于历史数据训练模型，预测任务耗时 → 优化并发调度
2. **资源感知调度**：考虑 Token 预算、人工审查时间 → 避免资源耗尽
3. **优先级衰减**：P0 任务如果 7 天未完成 → 自动降级到 P1（防止僵尸任务占用资源）
4. **智能依赖推断**：基于任务描述的 NLP 分析，自动推断 blocked_by 关系
