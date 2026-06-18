# Skill: Agent Serial Mode — Agent 串行执行模式（Workflow 降级）

## 职责

当 Workflow 工具不可用时，使用 Agent 工具串行执行 Generator → QA → Review 流程。

## 输入

```javascript
{
  serviceName: '审核服务',
  serviceDir: 'services/moderation-service',
  task: '实现敏感词检测功能',
  taskType: 'feature'
}
```

## 执行流程

### Step 1: Generator（开发）

```javascript
const generatorResult = await agent(generatorPrompt(1, '', taskType), {
  label: `${serviceName}: 开发`,
  isolation: 'worktree',
  model: 'claude-opus-4'  // Generator 需要推理能力
})
```

### Step 2: QA（质量检查）

```javascript
const qaResult = await agent(qaPrompt(), {
  label: `QA: ${serviceName}`,
  schema: QA_SCHEMA,
  model: 'claude-sonnet-4'  // QA 机械检查，Sonnet 足够
})

if (qaResult.verdict === 'FAIL') {
  // 修复循环（最多 3 轮）
  for (let i = 0; i < 3; i++) {
    log(`🔧 修复轮 ${i+1}...`)
    
    await agent(generatorPrompt(i+2, qaResult.summary, taskType), {
      label: `${serviceName}: 修复轮${i+1}`,
      isolation: 'worktree'
    })
    
    const retryQA = await agent(qaPrompt(), {
      label: `QA: ${serviceName} (重试)`,
      schema: QA_SCHEMA,
      model: 'claude-sonnet-4'
    })
    
    if (retryQA.verdict === 'PASS') {
      qaResult = retryQA
      break
    }
  }
}

if (qaResult.verdict !== 'PASS') {
  log('❌ QA 连续 3 轮失败，终止流程')
  return {status: 'fail', reason: 'qa_failed_after_3_retries'}
}
```

### Step 3: Review（3 视角串行）

```javascript
const REVIEW_LENSES = [
  {key: 'security-arch', label: '安全架构'},
  {key: 'testing-qa', label: '测试质量'},
  {key: 'design-pattern', label: '设计模式'}
]

const reviews = []
for (const lens of REVIEW_LENSES) {
  log(`🔍 Review: ${lens.label}...`)
  
  const reviewResult = await agent(reviewLensPrompt(lens), {
    label: `Review: ${lens.label}`,
    schema: REVIEW_SCHEMA,
    model: 'claude-sonnet-4'
  })
  
  reviews.push(reviewResult)
}

// 2/3 通过规则
const passCount = reviews.filter(r => r.verdict === 'PASS').length

if (passCount < 2) {
  log('❌ Review 未通过（需至少 2/3）')
  return {status: 'fail', reason: 'review_failed', reviews}
}
```

### Step 4: 汇总结果

```javascript
return {
  status: passCount >= 2 ? 'pass' : 'fail',
  confidence: computeConfidence({
    iterations: iteration,
    qaPass: qaResult.verdict === 'PASS',
    reviewPass: passCount >= 2,
    memoryApplied: /* 从 Generator 提取 */
  }),
  serviceName,
  serviceDir,
  qaSummary: qaResult.summary,
  reviewSummary: reviews.map(r => `${r.label}: ${r.verdict}`).join(', ')
}
```

## 性能对比

| 模式 | 3 个服务耗时 | 并发度 | 可用性要求 |
|------|------------|--------|-----------|
| Workflow 并行 | ~20 分钟 | 3-5 | Workflow 可用 |
| **Agent 串行** | **~60 分钟** | 1 | Workflow 不可用 |
| 内联模式 | ~15 分钟 | 1 | 紧急情况 |

**结论**: Agent 串行模式牺牲性能（3x），换取可用性。

## 使用方式

在 Owner Agent 中检测 Workflow 不可用时，切换到此模式：

```javascript
const mode = await bash('bash .harness/scripts/workflow-fallback.sh select --services 3')

if (mode === 'agent') {
  log('⚠️  Workflow 不可用，降级到 Agent 串行模式')
  
  // 逐个服务串行执行
  for (const svc of services) {
    const result = await executeAgentSerialMode(svc)
    results.push(result)
  }
} else {
  // 正常 Workflow 并行
  const results = await parallel([
    () => workflow({scriptPath: '...', args: svc1}),
    () => workflow({scriptPath: '...', args: svc2})
  ])
}
```

## 降级日志

记录到 `.harness/runtime/fallback-events.log`：

```
[2026-06-18T15:30:00Z] WORKFLOW_UNAVAILABLE → degraded to AGENT_SERIAL
[2026-06-18T15:30:05Z] Agent serial: moderation-service started
[2026-06-18T15:45:20Z] Agent serial: moderation-service completed (15.3 min)
[2026-06-18T15:45:25Z] Agent serial: community-hub-service started
...
```

## 恢复机制

定期检查 Workflow 健康度，自动恢复：

```javascript
// 每次任务开始前检查
const health = await bash('bash .harness/scripts/workflow-fallback.sh check')

if (health === 'healthy' && currentMode === 'agent') {
  log('✅ Workflow 已恢复，切换回正常模式')
  currentMode = 'workflow'
}
```
