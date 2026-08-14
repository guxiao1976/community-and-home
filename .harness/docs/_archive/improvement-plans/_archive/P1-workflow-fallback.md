# P1-2: Workflow 降级方案

## 背景

**现状问题**：
- 阶段 5（编码+测试）强依赖 Claude Code 的 `Workflow` 工具
- **硬性禁令**（`owner-agent.md:129-136`）：禁止使用外部技能替代 `harness-pipeline.js`
- **单点故障**：
  - Workflow 工具不可用（API 限流、工具 bug、版本不兼容）→ 整个流水线瘫痪
  - 无降级方案 → 无法交付任何代码变更
- **供应商锁定**：过度依赖 Claude Code 专有工具

**影响**：
- 系统可用性风险（SLA 无保障）
- 业务连续性受威胁
- 无法在 Workflow 不可用期间处理紧急修复

## 目标

实现 **3 级降级策略**，确保在 Workflow 不可用时，流水线仍能以降级模式运行，保障业务连续性。

目标可用性：**99.5%**（年停机时间 < 44 小时）

## 技术方案

### 架构设计

```
Tier 1: Workflow 并行模式（正常）
  ↓ Workflow 不可用
Tier 2: Agent 串行模式（降级）
  ↓ Agent 隔离失败
Tier 3: 内联模式（紧急）
  ↓ 全面故障
手动模式（人工接管）
```

### Tier 1: Workflow 并行模式（正常）

**特点**：
- 多服务并行执行（无依赖的服务同时启动）
- 每个服务独立 worktree 隔离
- Generator → QA → Review 完整流程
- 最大并发度：5（受 Claude API 限流约束）

**适用场景**：
- Workflow 工具可用
- 非紧急变更
- 跨服务功能开发

**代码示例**（当前实现）：
```javascript
const results = await parallel([
  () => workflow({scriptPath: '.harness/workflows/harness-pipeline.js', 
                  args: {serviceName:'审核服务', serviceDir:'services/moderation-service', task:'...'}}),
  () => workflow({scriptPath: '.harness/workflows/harness-pipeline.js', 
                  args: {serviceName:'社区枢纽', serviceDir:'services/community-hub-service', task:'...'}})
])
```

### Tier 2: Agent 串行模式（降级）

**特点**：
- 使用 `Agent()` 工具替代 `Workflow()`
- 串行执行（一个服务完成后再启动下一个）
- 每个服务仍有 worktree 隔离
- Generator → QA → Review 流程保留
- 并发度：1（串行）

**触发条件**：
1. Workflow 工具不可用（调用返回错误 ≥3 次）
2. 用户显式指定 `--fallback-mode=agent`
3. 单服务变更（无需并行）

**实现**：Owner Agent 新增降级逻辑

```javascript
// 伪代码
let mode = 'workflow'  // 默认 Workflow 模式

async function dispatchService(serviceName, serviceDir, task) {
  if (mode === 'workflow') {
    try {
      const result = await workflow({
        scriptPath: '.harness/workflows/harness-pipeline.js',
        args: {serviceName, serviceDir, task}
      })
      return result
    } catch (err) {
      if (err.message.includes('Workflow tool unavailable')) {
        log('⚠️ Workflow 不可用，降级到 Agent 串行模式')
        mode = 'agent'  // 永久降级（本次会话）
        // 重试当前任务
        return dispatchService(serviceName, serviceDir, task)
      }
      throw err
    }
  }
  
  if (mode === 'agent') {
    // 串行执行：Generator → QA → Review
    const generatorResult = await agent(generatorPrompt(1, '', 'feature'), {
      label: `${serviceName}: 开发`,
      isolation: 'worktree'
    })
    
    const qaResult = await agent(qaPrompt(), {
      label: `QA: ${serviceName}`,
      schema: QA_SCHEMA
    })
    
    if (qaResult.verdict === 'FAIL') {
      // 修复循环（最多 3 轮）
      for (let i = 0; i < 3; i++) {
        const fixResult = await agent(generatorPrompt(i+2, qaResult.summary, 'feature'), {
          label: `${serviceName}: 修复轮${i+1}`,
          isolation: 'worktree'
        })
        const retryQA = await agent(qaPrompt(), {label: `QA: ${serviceName}`, schema: QA_SCHEMA})
        if (retryQA.verdict === 'PASS') {
          qaResult = retryQA
          break
        }
      }
    }
    
    // Review（3 视角串行）
    const reviews = []
    for (const lens of REVIEW_LENSES) {
      const r = await agent(reviewLensPrompt(lens), {
        label: `Review:${lens.key}`,
        schema: REVIEW_SCHEMA
      })
      reviews.push(r)
    }
    
    const passCount = reviews.filter(r => r.verdict === 'PASS').length
    return {
      status: passCount >= 2 ? 'pass' : 'fail',
      serviceName,
      qaSummary: qaResult.summary,
      reviewSummary: reviews.map(r => `${r.label}: ${r.verdict}`).join(', ')
    }
  }
}

// 跨服务调度
const results = []
for (const svc of services) {
  const result = await dispatchService(svc.name, svc.dir, svc.task)
  results.push(result)
}
```

**性能对比**：
- Workflow 并行：3 个服务，耗时 ~20 分钟（并行）
- Agent 串行：3 个服务，耗时 ~60 分钟（串行）
- **可用性 > 性能**，降级模式可接受

### Tier 3: 内联模式（紧急）

**特点**：
- Owner Agent 直接内联执行（不启动子 Agent）
- 无 worktree 隔离（直接在主分支操作）
- 简化流程：Generator + QA（跳过 Review）
- 并发度：1，手动执行

**触发条件**：
1. Agent 隔离失败（worktree 创建错误）
2. 紧急修复（用户指定 `--emergency-mode`）
3. 单文件小改动（≤10 行）

**实现**：Owner Agent 内联逻辑

```javascript
if (mode === 'inline') {
  log('🚨 紧急内联模式 — 跳过隔离，直接操作主分支')
  
  // 1. 直接 Edit/Write 修改代码（Owner Agent 自己执行）
  // 2. 运行 build + test
  const buildResult = await bash('cd services/moderation-service && go build ./...')
  const testResult = await bash('cd services/moderation-service && go test ./...')
  
  if (buildResult.exitCode !== 0 || testResult.exitCode !== 0) {
    log('❌ 构建或测试失败，回滚变更')
    await bash('git checkout -- services/moderation-service')
    return {status: 'fail', reason: 'build or test failed'}
  }
  
  log('✅ 内联模式完成（跳过 Review，人工验收）')
  return {status: 'pass', mode: 'inline', requiresHumanReview: true}
}
```

**风险提示**：
- 无隔离 → 可能污染主分支
- 无 Review → 质量风险
- **仅用于紧急情况**（生产故障、安全漏洞）

### Tier 4: 手动模式（人工接管）

**触发条件**：
- 所有自动化模式失败
- 系统全面故障

**流程**：
1. Owner Agent 输出当前状态 + 待办任务清单
2. 生成手动操作指南（Step-by-step）
3. 用户接管，手动执行
4. 完成后更新任务状态

**输出示例**：
```markdown
## 🚨 系统降级到手动模式

所有自动化工具不可用，请按以下步骤手动执行：

### 待办任务
1. 修改 services/moderation-service/internal/logic/check_logic.go
   - 添加 nil 校验（第 45 行）
   - 参考设计：.harness/changes/moderation-fix/design.md

2. 运行测试
   \`\`\`bash
   cd services/moderation-service
   go build ./...
   go test ./...
   \`\`\`

3. 更新 CHANGELOG.md

4. 提交
   \`\`\`bash
   git add .
   git commit -m "fix: 修复 CheckText nil pointer"
   \`\`\`

5. 通知 Owner Agent 完成
   在下一次对话中输入："任务 task-2026-06-18-010 已手动完成"
```

### 降级决策树

```
收到变更请求
  ↓
是否紧急？
  是 → 单文件小改？
    是 → Tier 3（内联）
    否 → Tier 2（Agent 串行）
  否 ↓
单服务 or 跨服务？
  单服务 → 尝试 Tier 1（Workflow）
    成功 → 完成
    失败 → Tier 2（Agent 串行）
  跨服务 → 尝试 Tier 1（Workflow 并行）
    成功 → 完成
    失败 → Tier 2（Agent 串行，逐个服务）
      成功 → 完成
      失败 → Tier 3（内联，需人工确认）
        成功 → 完成
        失败 → Tier 4（手动）
```

## 实施步骤

### Phase 1: Agent 串行模式（2 天）

**Task 1.1**: 实现降级逻辑
- 位置：Owner Agent 内联或新 Skill `.harness/skills/fallback-dispatch.md`
- 实现 `dispatchService()` 函数（支持 workflow / agent 两种模式）
- 模式切换逻辑（自动检测 Workflow 不可用）

**Task 1.2**: Worktree 隔离保留
- Agent 串行模式仍使用 `isolation: 'worktree'`
- 每个服务独立分支（避免冲突）

**Task 1.3**: 状态持久化
- 降级事件记录到 `.harness/runtime/fallback-events.log`
- 格式：`[timestamp] WORKFLOW_UNAVAILABLE → degraded to AGENT_SERIAL`

### Phase 2: 内联模式（1 天）

**Task 2.1**: 实现内联逻辑
- Owner Agent 直接执行 Edit/Write
- 运行 build + test
- 失败 → 自动回滚（`git checkout --`）

**Task 2.2**: 人工确认机制
- 内联模式完成后 → 输出变更摘要
- 询问用户："内联模式已完成，变更未经 Review，是否继续？"
- 拒绝 → 回滚

**Task 2.3**: 触发条件配置
- 环境变量：`HARNESS_ALLOW_INLINE_MODE=true`（默认 false）
- 用户显式传参：`--emergency-mode`

### Phase 3: 降级决策引擎（1 天）

**Task 3.1**: 实现决策树
- 函数：`selectDispatchMode(urgency, fileCount, serviceCount)`
- 返回：`{mode: 'workflow|agent|inline', reason: '...'}`

**Task 3.2**: 健康检查
- 启动时检测 Workflow 工具可用性
- `await workflow({script: 'export const meta = {name:"test"}; return {status:"ok"}'})` 
- 成功 → mode=workflow，失败 → mode=agent

**Task 3.3**: 熔断器
- Workflow 连续失败 3 次 → 熔断，30 分钟内不再尝试
- 30 分钟后 → 重试健康检查

### Phase 4: 测试验证（1.5 天）

**Task 4.1**: 模拟 Workflow 不可用
- Mock Workflow 工具返回错误
- 验证自动降级到 Agent 串行
- 检查任务正常完成

**Task 4.2**: 性能测试
- 3 服务变更，Workflow 并行 vs Agent 串行
- 记录耗时对比（预期 3x 差异）

**Task 4.3**: 内联模式测试
- 单文件修改，触发内联模式
- 故意引入错误 → 验证自动回滚
- 人工拒绝 → 验证回滚逻辑

### Phase 5: 文档和上线（0.5 天）

**Task 5.1**: 更新文档
- `owner-agent.md` 补充降级策略章节
- 运维手册：如何识别降级事件、如何恢复

**Task 5.2**: 创建 Memory
- `.harness/knowledge/memory/workflow-fallback-strategy.md`

**Task 5.3**: 监控和告警
- 降级事件 → 写入日志 + 通知用户
- 格式："⚠️ Workflow 不可用，已降级到 Agent 串行模式，性能下降 3x"

## 验收标准

### 功能验收

- [ ] Workflow 不可用时自动降级到 Agent 串行
- [ ] Agent 隔离失败时降级到内联模式
- [ ] 内联模式失败时输出手动操作指南
- [ ] 熔断器正确阻止重复失败尝试

### 可用性验收

| 故障场景 | 降级模式 | 任务完成率 | 耗时增加 |
|---------|---------|----------|---------|
| Workflow 不可用 | Agent 串行 | 100% | 3x |
| Agent 隔离失败 | 内联 | 90%（需人工确认） | 1x |
| 全面故障 | 手动 | 100%（人工） | 10x |

**目标可用性**：99.5%（假设 Workflow 可用性 99%，降级成功率 95%）

### 性能验收

- [ ] 健康检查耗时 < 5 秒
- [ ] 模式切换延迟 < 10 秒
- [ ] Agent 串行模式完成单服务变更 < 30 分钟

## 风险和依赖

### 风险

**R1: 降级模式质量下降**
- **描述**：内联模式跳过 Review → 低级错误进入主分支
- **缓解**：
  - 内联模式仅用于紧急情况
  - 完成后触发异步 Review（事后审查）
  - 要求人工确认

**R2: Agent 串行性能不可接受**
- **描述**：跨 5 个服务变更，Agent 串行耗时 >2 小时
- **缓解**：
  - 限制单次变更最多 3 个服务
  - 超过 3 个 → 拆分为多个任务

**R3: 熔断器过度保护**
- **描述**：Workflow 偶发故障 → 熔断 30 分钟 → 不必要的降级
- **缓解**：
  - 失败阈值设为 3 次（不是 1 次）
  - 熔断时间可配置（`CIRCUIT_BREAKER_TIMEOUT_MINUTES`）

### 依赖

**D1: Agent 工具稳定性**
- Agent 降级方案依赖 `Agent()` 工具可用
- 如果 Agent 也不可用 → 只能手动模式
- 行动：监控 Agent 工具健康度

**D2: Worktree 隔离依赖 Git**
- Agent 串行模式仍需 worktree
- Git worktree 故障 → 降级到内联
- 行动：定期清理 `.claude/worktrees/` 避免磁盘满

## 效果预估

### 可用性提升

| 指标 | 改进前 | 改进后 | 提升 |
|------|-------|--------|------|
| 系统可用性（假设 Workflow 可用性 99%） | 99.0% | **99.5%** | +0.5% |
| 年停机时间 | 87.6 小时 | **43.8 小时** | ↓ 50% |
| 紧急修复响应时间（Workflow 不可用时） | 无法处理 | <30 分钟（内联） | - |

### 成本分析

| 项目 | 成本 | 说明 |
|------|------|------|
| 开发成本 | 6 人日 | 一次性投入 |
| 性能损失（降级时） | 3x 耗时 | 假设降级 1% 时间 → 平均增加 2% |
| 质量风险（内联模式） | 低 | 仅紧急情况使用，<0.1% 的变更 |

**ROI**：可用性提升 0.5% → 减少业务中断损失 >> 开发成本

## 后续优化

1. **智能降级决策**：基于历史数据学习，预测 Workflow 故障概率 → 提前降级
2. **部分降级**：单个服务 Workflow 失败 → 仅该服务降级，其他服务仍并行
3. **自动恢复**：Workflow 恢复后，自动从 Agent 模式切回 Workflow 模式
4. **降级模式优化**：Agent 串行改为 Agent 受限并行（并发度 2，避免冲突）
5. **可观测性集成**：降级事件发送到监控系统（Prometheus metrics）
