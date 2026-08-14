# P2-3: 成本监控

## 背景

**现状问题**：
- **Token 使用无追踪**：不知道每次 Pipeline 消耗多少 tokens
- **成本不可控**：无预算限制，超支无预警
- **优化无依据**：不知道哪个环节最耗 tokens（Generator? Review?）
- **无成本归属**：无法按服务/任务类型/团队成员归类成本

**影响**：
- AI 成本不透明（黑盒支出）
- 无法做成本优化（不知道瓶颈在哪）
- 预算超支风险（无提前预警）
- ROI 无法量化（投入产出比不清晰）

## 目标

建立 **Token 使用跟踪 + 成本预算管理** 系统，实现：
1. 实时追踪每次 Pipeline 的 token 使用
2. 按维度归类成本（服务/任务类型/Agent 类型/阶段）
3. 设置预算并告警（超过阈值自动通知）
4. 生成成本报告（日/周/月）

预期效果：Token 成本优化 **20-30%**

## 技术方案

### 1. Token 使用追踪

#### 追踪点设计

| 追踪点 | 数据 | 来源 |
|--------|------|------|
| Agent 调用 | input_tokens, output_tokens, model | Agent 工具返回值 |
| Workflow 执行 | 所有 Agent 的 token 总和 | Workflow 汇总 |
| Pipeline 完成 | 总 token 使用 + 按阶段分解 | harness-pipeline.js 汇总 |

#### 数据格式

**文件**：`.harness/runtime/token-usage.jsonl`

**格式**：
```jsonl
{"timestamp":"2026-06-18T15:30:15Z","event":"agent_call","agent":"generator","service":"moderation-service","iteration":1,"model":"claude-opus-4","input_tokens":1200,"output_tokens":850,"cost_usd":0.024}
{"timestamp":"2026-06-18T15:33:00Z","event":"agent_call","agent":"qa","service":"moderation-service","model":"claude-sonnet-4","input_tokens":500,"output_tokens":300,"cost_usd":0.004}
{"timestamp":"2026-06-18T15:35:30Z","event":"agent_call","agent":"review","lens":"security-arch","service":"moderation-service","model":"claude-sonnet-4","input_tokens":600,"output_tokens":400,"cost_usd":0.005}
{"timestamp":"2026-06-18T15:38:00Z","event":"pipeline_complete","service":"moderation-service","task_type":"feature","total_input_tokens":4500,"total_output_tokens":2800,"total_cost_usd":0.065,"breakdown":{"generator":0.024,"qa":0.004,"review":0.037}}
```

#### 成本计算

**定价模型**（Claude 4 系列，2026 年 6 月）：

| Model | Input (per 1M tokens) | Output (per 1M tokens) |
|-------|-----------------------|------------------------|
| claude-opus-4 | $15 | $75 |
| claude-sonnet-4 | $3 | $15 |
| claude-haiku-4 | $0.25 | $1.25 |

**计算公式**：
```javascript
function calculateCost(inputTokens, outputTokens, model) {
  const pricing = {
    'claude-opus-4': {input: 15, output: 75},
    'claude-sonnet-4': {input: 3, output: 15},
    'claude-haiku-4': {input: 0.25, output: 1.25}
  }
  const p = pricing[model] || pricing['claude-sonnet-4']  // fallback
  return (inputTokens * p.input + outputTokens * p.output) / 1_000_000
}
```

### 2. Token 追踪实现

#### harness-pipeline.js 注入

```javascript
// ── Token 追踪器 ──
const fs = require('fs')
const TOKEN_FILE = '.harness/runtime/token-usage.jsonl'

function trackTokens(event, data) {
  const entry = {
    timestamp: new Date().toISOString(),
    event: event,
    ...data
  }
  fs.appendFileSync(TOKEN_FILE, JSON.stringify(entry) + '\n')
}

// ── Agent 调用包装 ──
async function trackedAgent(prompt, opts) {
  const result = await agent(prompt, opts)
  
  // 从 result 提取 token 使用（假设 Agent 工具返回 usage 字段）
  const usage = result._usage || {input_tokens: 0, output_tokens: 0, model: 'unknown'}
  const cost = calculateCost(usage.input_tokens, usage.output_tokens, usage.model)
  
  trackTokens('agent_call', {
    agent: opts.label || 'unknown',
    service: args.serviceName,
    iteration: iteration,
    model: usage.model,
    input_tokens: usage.input_tokens,
    output_tokens: usage.output_tokens,
    cost_usd: cost
  })
  
  return result
}

// ── 使用 ──
const qaResult = await trackedAgent(qaPrompt(), {label: 'qa', schema: QA_SCHEMA})

// ── Pipeline 完成时汇总 ──
const totalCost = /* 从 TOKEN_FILE 读取本次 Pipeline 的所有记录求和 */
trackTokens('pipeline_complete', {
  service: args.serviceName,
  task_type: TASK_TYPE,
  total_input_tokens: totalInput,
  total_output_tokens: totalOutput,
  total_cost_usd: totalCost,
  breakdown: {generator: 0.024, qa: 0.004, review: 0.037}
})
```

#### Agent 工具返回值增强

**问题**：当前 `agent()` 工具可能不返回 token usage

**解决方案**：
1. **方案 A**：修改 Agent 工具，返回 `{result, usage: {input_tokens, output_tokens, model}}`
2. **方案 B**（fallback）：从 API 响应头估算（如果 Claude API 返回 `X-RateLimit-*` headers）
3. **方案 C**（粗略）：基于 Prompt 长度估算（1 token ≈ 4 chars）

**推荐**：方案 A + 方案 C fallback

### 3. 成本预算管理

#### 预算配置

**文件**：`.harness/config/budget.json`

```json
{
  "monthly_budget_usd": 500,
  "daily_budget_usd": 20,
  "per_pipeline_budget_usd": 1.0,
  "alerts": [
    {"threshold": 0.7, "level": "warning", "message": "已使用 70% 月度预算"},
    {"threshold": 0.9, "level": "critical", "message": "已使用 90% 月度预算"},
    {"threshold": 1.0, "level": "emergency", "message": "月度预算耗尽，暂停自动调度"}
  ],
  "cost_limits": {
    "generator": 0.5,    // 单次 Generator 最多 $0.50
    "review": 0.3,       // 单次 Review（3视角）最多 $0.30
    "pipeline": 1.0      // 单次 Pipeline 最多 $1.00
  }
}
```

#### 预算检查逻辑

```javascript
// ── 预算检查器 ──
function checkBudget() {
  const budget = JSON.parse(fs.readFileSync('.harness/config/budget.json', 'utf8'))
  const usageThisMonth = calculateMonthlyUsage()  // 从 token-usage.jsonl 统计本月总成本
  
  const utilizationRate = usageThisMonth / budget.monthly_budget_usd
  
  for (const alert of budget.alerts) {
    if (utilizationRate >= alert.threshold && !isAlertSent(alert.level)) {
      sendAlert(alert.level, alert.message, {
        used: usageThisMonth,
        budget: budget.monthly_budget_usd,
        rate: (utilizationRate * 100).toFixed(1) + '%'
      })
      markAlertSent(alert.level)
    }
  }
  
  if (utilizationRate >= 1.0) {
    throw new Error('Monthly budget exhausted. Halting pipeline.')
  }
}

// ── Pipeline 启动前检查 ──
checkBudget()

// ── Agent 调用前检查单次限额 ──
function checkAgentBudget(agentType, estimatedCost) {
  const budget = JSON.parse(fs.readFileSync('.harness/config/budget.json', 'utf8'))
  const limit = budget.cost_limits[agentType]
  
  if (estimatedCost > limit) {
    throw new Error(`${agentType} estimated cost $${estimatedCost} exceeds limit $${limit}`)
  }
}
```

### 4. 成本报告

#### 日报

**脚本**：`bash .harness/scripts/token-report.sh --period daily`

**输出示例**：
```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Token 使用报告 - 2026-06-18
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

总成本: $12.45

按服务分组:
  moderation-service:    $5.20 (41.8%)
  community-hub-service: $3.80 (30.5%)
  user-service:          $2.15 (17.3%)
  web/pc:                $1.30 (10.4%)

按阶段分组:
  Generator: $6.50 (52.2%)
  Review:    $4.20 (33.7%)
  QA:        $1.75 (14.1%)

按任务类型分组:
  feature: $8.30 (66.7%)
  bug:     $2.80 (22.5%)
  debt:    $1.35 (10.8%)

预算使用:
  本日预算: $20.00
  已使用:   $12.45 (62.3%)
  剩余:     $7.55

本月预算:
  月度预算: $500.00
  已使用:   $285.30 (57.1%)
  剩余:     $214.70
  预计月底总成本: $498.15 (在预算内 ✅)

Top 5 最耗 Token 的 Pipeline:
  1. moderation-service (feature) - $2.30
  2. community-hub-service (feature) - $1.85
  3. moderation-service (bug) - $1.20
  4. user-service (feature) - $1.10
  5. web/pc (feature) - $0.95
```

**实现**：
```bash
#!/usr/bin/env bash
TOKEN_FILE=".harness/runtime/token-usage.jsonl"
PERIOD="${1:-daily}"  # daily/weekly/monthly

case $PERIOD in
  daily)
    CUTOFF=$(date -u -d '1 day ago' +"%Y-%m-%dT%H:%M:%SZ")
    ;;
  weekly)
    CUTOFF=$(date -u -d '7 days ago' +"%Y-%m-%dT%H:%M:%SZ")
    ;;
  monthly)
    CUTOFF=$(date -u -d '30 days ago' +"%Y-%m-%dT%H:%M:%SZ")
    ;;
esac

# 过滤时间范围内的记录
RECENT=$(jq -s "map(select(.timestamp > \"$CUTOFF\"))" "$TOKEN_FILE")

# 统计总成本
TOTAL_COST=$(echo "$RECENT" | jq '[.[] | select(.event == "pipeline_complete") | .total_cost_usd] | add')

# 按服务分组
echo "$RECENT" | jq -r '[.[] | select(.event == "pipeline_complete")] | group_by(.service) | map({service: .[0].service, cost: map(.total_cost_usd) | add}) | sort_by(-.cost) | .[]' \
  | jq -r '"  \(.service): $\(.cost) (\((.cost / '$TOTAL_COST' * 100) | round)%)"'

# ... 其他统计
```

### 5. 成本优化建议

#### 自动化优化策略

**策略 1: 模型降级**
- 当前：所有 Agent 默认 `claude-opus-4`
- 优化：
  - Generator/Review → `claude-opus-4`（需要推理能力）
  - QA → `claude-sonnet-4`（机械检查，Sonnet 足够）
  - Debug → `claude-sonnet-4`
- **预期节省**：15-20%

**策略 2: Prompt 压缩**
- 当前：Generator Prompt 包含完整规范（~1200 tokens）
- 优化：
  - 提取核心规则到简短清单（~300 tokens）
  - 详细规范链接到文档（Agent 按需读取）
- **预期节省**：10-15%

**策略 3: 缓存利用**
- 当前：每次 Pipeline 重新加载 CLAUDE.md / design.md
- 优化：
  - 利用 Anthropic Prompt Caching（5 分钟 TTL）
  - 静态部分标记为 cacheable
- **预期节省**：20-30%（缓存命中时）

**策略 4: 任务类型差异化**
- 当前：chore/debt 任务走完整 3 视角 Review
- 优化（已实现）：
  - chore → 跳过 Review
  - debt → 只 1 视角 Review
- **预期节省**：5-10%（已实现）

#### 优化效果监控

**指标**：
- `cost_per_pipeline_avg` — 平均每次 Pipeline 成本
- `cost_per_service_line_of_code` — 每行代码成本
- `token_efficiency` — tokens / 代码行数比率

**目标**：
- 单次 Pipeline 成本：$1.00 → $0.70（↓ 30%）
- 每行代码成本：$0.05 → $0.035（↓ 30%）

## 实施步骤

### Phase 1: Token 追踪（1.5 天）

**Task 1.1**: 实现 Token 追踪器
- 文件输出（`token-usage.jsonl`）
- 成本计算函数
- Agent 调用包装（`trackedAgent()`）

**Task 1.2**: 集成到 harness-pipeline.js
- 替换所有 `agent()` 调用为 `trackedAgent()`
- Pipeline 完成时汇总记录

**Task 1.3**: Agent 工具返回值增强
- 如果 Agent 工具不返回 usage → 实现 fallback（基于长度估算）

### Phase 2: 预算管理（1 天）

**Task 2.1**: 创建预算配置
- `budget.json` 模板
- 默认预算值（月 $500 / 日 $20 / 单次 $1）

**Task 2.2**: 实现预算检查
- 月度/日度/单次预算检查
- 告警机制（warning/critical/emergency）
- 预算耗尽时暂停调度

**Task 2.3**: 预算可视化
- 命令：`bash .harness/scripts/budget-status.sh`
- 输出：当前使用率、剩余预算、预计耗尽时间

### Phase 3: 成本报告（1 天）

**Task 3.1**: 实现报告生成脚本
- `token-report.sh --period daily/weekly/monthly`
- 统计：总成本、按服务/阶段/任务类型分组、Top N

**Task 3.2**: 集成到日常流程
- 每日自动生成报告（cron）
- 报告发送到邮件/Slack

**Task 3.3**: HTML 可视化（可选）
- 图表：成本趋势、服务占比、阶段分布
- 工具：Chart.js 或 D3.js

### Phase 4: 成本优化（0.5 天）

**Task 4.1**: 实现模型降级
- QA / Debug Agent → 使用 `claude-sonnet-4`
- 配置化（允许覆盖）

**Task 4.2**: Prompt 压缩试点
- 选择一个 Prompt（如 QA）压缩 50%
- A/B 测试：对比质量和成本

**Task 4.3**: 优化效果监控
- 对比优化前后的成本指标
- 记录到 Memory

## 验收标准

### 功能验收

- [ ] Token 使用正确记录到 `token-usage.jsonl`
- [ ] 成本计算准确（与 API 账单误差 <5%）
- [ ] 预算检查正确触发告警
- [ ] 预算耗尽时暂停 Pipeline
- [ ] 报告生成成功，数据准确

### 准确性验收

- [ ] 对比 10 次 Pipeline 的 token 追踪 vs Claude API 账单，误差 <5%
- [ ] 成本分组统计与手动计算一致

### 优化验收

- [ ] 模型降级后，QA 成本降低 ≥40%（Opus → Sonnet）
- [ ] 总体成本降低 ≥20%（实施所有策略后）

## 风险和依赖

### 风险

**R1: Token 估算不准确**
- **描述**：基于长度估算的 fallback 误差较大
- **缓解**：
  - 优先使用 Agent 工具返回的精确值
  - 定期与 API 账单对账，调整估算公式

**R2: 预算过于严格影响开发**
- **描述**：预算耗尽导致无法处理紧急修复
- **缓解**：
  - 设置紧急通道（`--bypass-budget` 选项，需人工批准）
  - 预算软限制 + 硬限制（90% 警告，110% 强制停止）

**R3: 成本优化降低质量**
- **描述**：模型降级或 Prompt 压缩导致输出质量下降
- **缓解**：
  - 优化前后做 A/B 测试
  - 监控 QA 通过率、Review CRITICAL 数
  - 质量下降 → 回滚优化

### 依赖

**D1: Agent 工具支持 usage 返回**
- 需要 Agent 工具返回 token 使用数据
- 行动：如果不支持，实现 fallback

**D2: Claude API 定价稳定**
- 定价模型变化 → 成本计算失效
- 行动：定期更新 `pricing` 配置

## 效果预估

### 成本节省

| 优化策略 | 节省比例 | 实施难度 |
|---------|---------|---------|
| 模型降级（QA/Debug → Sonnet） | 15-20% | 低 |
| Prompt 压缩 | 10-15% | 中 |
| Prompt Caching | 20-30% | 低（API 特性） |
| 任务类型差异化（已实现） | 5-10% | 低 |
| **总计** | **25-35%** | - |

**假设**：
- 当前月成本：$500
- 优化后：$500 × (1 - 0.30) = $350
- **年节省**：$150 × 12 = **$1800**

### 投资回报

| 项目 | 成本 | 说明 |
|------|------|------|
| 开发成本 | 4 人日 | 一次性投入 |
| 年成本节省 | $1800 | 持续收益 |
| **ROI** | **450%** | 首年回报率 |

## 后续优化

1. **机器学习成本预测**：基于历史数据预测未来成本趋势
2. **动态预算分配**：根据任务优先级动态调整预算（P0 任务不受限）
3. **团队成本归属**：按开发者/团队统计成本，促进成本意识
4. **实时成本面板**：WebSocket 推送实时成本到监控大屏
5. **成本优化自动化**：Agent 自动选择最优模型（质量与成本平衡）
