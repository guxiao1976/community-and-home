# P0-1: HITL 置信度自适应审查

## 背景

**现状问题**：
- `owner-agent.md:229-246` 已定义置信度分级审查规则（≥0.80 摘要 / 0.50-0.79 抽查 / <0.50 全文）
- `harness-pipeline.js:708-715` 计算 confidence 评分（0.0-1.0）
- 但 Owner Agent 在阶段 5 验收时，**未实现分级审查逻辑**，所有产出都需人工全文审查

**影响**：
- 人工审查时间浪费：高置信产出（confidence ≥0.80）仍需人工逐行审查
- 人类疲劳：大量机械审查降低对真正问题的敏感度
- 预估节省：实现后可减少 **50% 人工审查时间**

## 目标

在 Owner Agent 阶段 5 验收环节，根据 Workflow 返回的 confidence 评分，自动决定审查深度：
- **≥0.80**：仅读 QA summary + Review summary，确认无异常
- **0.50-0.79**：随机抽查 max(2, totalFiles × 30%) 个变更文件全文
- **<0.50**：全文审查所有变更文件 + 人工确认

## 技术方案

### 1. Owner Agent Skill 增强

**文件**：`.harness/agents/subagents/owner-agent.md` 或新建 `.harness/skills/adaptive-review.md`

**新增逻辑**：

```markdown
## 阶段 5 验收：置信度自适应审查

### Step 1: 收集 Workflow 返回值
每个 Workflow 返回：
\`\`\`javascript
{
  status: 'pass',
  confidence: 0.85,  // ← 关键字段
  serviceName: '审核服务',
  qaSummary: '...',
  reviewSummary: '...',
  memorySuggestions: [...]
}
\`\`\`

### Step 2: 按 confidence 分级处理

#### ≥0.80 — 摘要审查（Trust & Verify）
**审查内容**：
1. 读取 `services/<name>/_qa.md` 的「摘要」章节（前 50 行）
2. 读取 `services/<name>/_review_*.md` 的「摘要」和「🔴 CRITICAL」表格
3. 检查点：
   - QA verdict = PASS？
   - Review 无 CRITICAL？
   - CHANGELOG 已更新？

**通过条件**：3 个检查点全✅ → 标记 PASS
**失败处理**：任一检查点 ❌ → 降级到「抽查」模式

#### 0.50-0.79 — 抽查模式（Spot Check）
**抽样策略**：
1. 获取变更文件列表：`git diff --name-only HEAD`
2. 计算抽查数量：`sample_size = max(2, total_files × 30%)`
3. 随机抽样（种子：任务 ID hash，确保可复现）

**审查内容**（每个抽样文件）：
- 全文阅读代码
- 检查是否遵守 Memory 指导（有 `[[slug]]` 引用？）
- 检查 TDD 证据（新增函数有对应测试？）
- 检查边界条件处理

**记录**：写入 `.harness/changes/<change>/impl/<service>/spot_check.md`：
\`\`\`markdown
## 抽查报告
- 抽查文件: [file1.go, file2.go]
- 抽查发现: 
  - ✅ file1.go: Memory 引用正确，测试覆盖完整
  - ⚠️ file2.go: 缺少空值校验，已记录 WARNING
\`\`\`

**通过条件**：无 CRITICAL 发现 → PASS
**失败处理**：发现 ≥1 CRITICAL → 强制回退到阶段 5 编码

#### <0.50 — 全文审查 + 人工确认（Deep Dive）
**审查内容**：
- 阅读所有变更文件（全文）
- 按 9 维度审查（架构/设计/规范/质量/安全/复用/测试/变更/记忆）
- 人工确认：输出问题清单 + 建议，**暂停等待用户决策**

**用户决策**：
- 继续交付（接受风险）→ 标记为「人工批准」
- 回退修复 → 返回阶段 5 编码

### Step 3: 更新 summary.md

在 `.harness/changes/<change>/summary.md` 记录审查决策：
\`\`\`markdown
## 阶段 5 验收
- Confidence: 0.85
- 审查模式: 摘要审查
- 审查结果: PASS
- 审查人: Owner Agent (自动)
- 审查时间: 2026-06-18 14:30
\`\`\`
```

### 2. 实现示例（Owner Agent Prompt 片段）

**位置**：`.harness/agents/subagents/owner-agent.md` 或内联到 Owner Agent 主循环

```markdown
## 阶段 5 后：置信度自适应审查

收到 Workflow 返回值后，执行：

\`\`\`javascript
// 伪代码示例
for (const result of workflowResults) {
  const {confidence, serviceName, serviceDir} = result
  
  if (confidence >= 0.80) {
    log(`${serviceName}: 高置信 (${confidence})，执行摘要审查`)
    // 读取 _qa.md 前 50 行 + _review_*.md 摘要
    const qaPass = checkQASummary(serviceDir)
    const reviewPass = checkReviewSummary(serviceDir)
    const changelogUpdated = checkChangelog(serviceDir)
    
    if (qaPass && reviewPass && changelogUpdated) {
      log(`✅ ${serviceName} 摘要审查通过`)
      result.reviewMode = 'summary'
      result.humanReviewNeeded = false
    } else {
      log(`⚠️ ${serviceName} 摘要审查发现异常，降级到抽查`)
      confidence = 0.70  // 降级
    }
  }
  
  if (confidence >= 0.50 && confidence < 0.80) {
    log(`${serviceName}: 中置信 (${confidence})，执行抽查`)
    const changedFiles = getChangedFiles(serviceDir)
    const sampleSize = Math.max(2, Math.floor(changedFiles.length * 0.3))
    const samples = randomSample(changedFiles, sampleSize, taskIdHash)
    
    log(`抽查文件: ${samples.join(', ')}`)
    const criticals = []
    for (const file of samples) {
      const issues = deepReviewFile(file)
      criticals.push(...issues.filter(i => i.level === 'CRITICAL'))
    }
    
    if (criticals.length > 0) {
      log(`❌ ${serviceName} 抽查发现 ${criticals.length} 个 CRITICAL`)
      result.reviewMode = 'spot_check_failed'
      result.humanReviewNeeded = true
      result.criticals = criticals
    } else {
      log(`✅ ${serviceName} 抽查通过`)
      result.reviewMode = 'spot_check'
      result.humanReviewNeeded = false
    }
  }
  
  if (confidence < 0.50) {
    log(`${serviceName}: 低置信 (${confidence})，要求全文审查`)
    result.reviewMode = 'full_review'
    result.humanReviewNeeded = true
    // 暂停，输出问题清单，等待用户决策
    await askUserForApproval(result)
  }
}
\`\`\`
```

## 实施步骤

### Phase 1: 基础设施（1 天）

**Task 1.1**: 创建抽查脚本
- 文件：`.harness/scripts/spot-check.sh`
- 功能：给定服务目录 + 抽样比例 → 随机抽取文件 → 输出文件列表
- 输入：`--service moderation-service --ratio 0.3 --seed <task-id>`
- 输出：JSON 数组 `["file1.go", "file2.go"]`

**Task 1.2**: 创建摘要提取器
- 文件：`.harness/scripts/extract-summary.sh`
- 功能：从 `_qa.md` / `_review_*.md` 提取摘要章节（前 N 行或特定章节）
- 输入：`--file _qa.md --lines 50` 或 `--file _review_*.md --section "摘要"`
- 输出：Markdown 摘要文本

**Task 1.3**: 更新 `summary.md` 模板
- 在 `.harness/changes/TEMPLATE.md` 增加「阶段 5 验收」章节
- 包含字段：confidence / reviewMode / humanReviewNeeded / reviewTime

### Phase 2: Owner Agent 集成（2 天）

**Task 2.1**: 编写 Skill 文件
- 创建：`.harness/skills/adaptive-review.md`
- 内容：上述技术方案的完整 Prompt

**Task 2.2**: 更新 Owner Agent 调度流程
- 文件：`.harness/agents/owner-agent.md`
- 修改阶段表第 5 行「Owner 验证」列：
  - 旧：`跟踪各 Workflow 摘要，全部 PASS → 下一阶段`
  - 新：`根据 confidence 执行分级审查（见 adaptive-review.md）`

**Task 2.3**: 实现检查函数
- 位置：Owner Agent 内联逻辑或独立 Skill
- 函数：
  - `checkQASummary(serviceDir)` → 读 `_qa.md` 检查 VERDICT
  - `checkReviewSummary(serviceDir)` → 读 `_review_*.md` 检查 CRITICAL 数
  - `checkChangelog(serviceDir)` → 检查 CHANGELOG.md 是否有新增条目
  - `deepReviewFile(file)` → 全文审查单个文件（9 维度）

### Phase 3: 测试验证（1 天）

**Task 3.1**: 准备测试场景
- 场景 1：高置信（confidence 0.85），QA/Review 全 PASS → 期望摘要审查通过
- 场景 2：中置信（confidence 0.65），抽查发现 CRITICAL → 期望回退
- 场景 3：低置信（confidence 0.45） → 期望暂停等待人工

**Task 3.2**: 端到端测试
- 启动 Workflow，修改 `computeConfidence()` 返回值模拟不同场景
- 验证 Owner Agent 执行对应审查模式
- 检查 `summary.md` 记录正确

**Task 3.3**: 回归测试
- 确保现有流程（无 confidence 场景）不受影响
- 确保可以手动覆盖审查模式（传入 `--force-full-review`）

### Phase 4: 文档和上线（0.5 天）

**Task 4.1**: 更新文档
- `.harness/agents/owner-agent.md` 补充分级审查说明
- `CLAUDE.md` 快速索引中添加 adaptive-review 链接

**Task 4.2**: 创建 Memory 记录
- 文件：`.harness/knowledge/memory/hitl-adaptive-review.md`
- 内容：分级审查规则、检查点、回退条件

**Task 4.3**: 上线监控
- 前 10 个变更：人工验证审查模式是否合理
- 收集误判案例（应该全文审查但用了摘要审查）→ 调整 confidence 权重

## 验收标准

### 功能验收

- [ ] 高置信场景（≥0.80）只读摘要，不读全文
- [ ] 中置信场景（0.50-0.79）执行抽查，记录到 spot_check.md
- [ ] 低置信场景（<0.50）暂停并等待用户决策
- [ ] summary.md 正确记录审查模式和结果
- [ ] 抽查发现 CRITICAL 强制回退到阶段 5

### 性能验收

- [ ] 摘要审查耗时 ≤ 全文审查 20%
- [ ] 抽查模式耗时 ≤ 全文审查 50%

### 质量验收

- [ ] 10 个高置信变更中，0 个漏判（应该回退但未回退）
- [ ] 10 个中置信变更中，抽查准确率 ≥ 80%（抽查样本代表性）

## 风险和依赖

### 风险

**R1: 误判风险**
- **描述**：高 confidence 但代码实际有严重问题
- **缓解**：
  - 前期设置 confidence 阈值保守（如 ≥0.85 才用摘要审查）
  - 收集误判案例，调整 `computeConfidence()` 权重
  - 保留人工覆盖选项（`--force-full-review`）

**R2: 抽样偏差**
- **描述**：随机抽样可能错过关键文件
- **缓解**：
  - 引入启发式抽样：优先抽 `*_logic.go`、`*_handler.go`、`*_test.go`
  - 最小抽样数 ≥2，避免单点代表全局

### 依赖

**D1: Workflow confidence 计算准确性**
- 当前 `computeConfidence()` 权重：迭代次数 30% + 评审共识 30% + 记忆覆盖 20% + QA 首次通过 20%
- 如果权重不合理 → confidence 不可信 → 分级审查失效
- **行动**：上线后持续调优权重

**D2: Owner Agent 上下文窗口**
- 全文审查（<0.50）会加载所有变更文件到上下文
- 如果单次变更 >50 文件 → 上下文溢出
- **行动**：限制单次变更文件数（超过 50 → 强制拆分任务）

## 效果预估

| 指标 | 现状 | 改进后 | 提升 |
|------|------|--------|------|
| 人工审查时间（高置信） | 30 分钟/服务 | 5 分钟/服务 | ↓ 83% |
| 人工审查时间（中置信） | 30 分钟/服务 | 15 分钟/服务 | ↓ 50% |
| 人工审查时间（低置信） | 30 分钟/服务 | 30 分钟/服务 | - |
| 平均审查时间（假设 60% 高置信 / 30% 中置信 / 10% 低置信） | 30 分钟 | **12 分钟** | ↓ 60% |
| 漏判率 | - | <5%（目标） | - |

## 后续优化

1. **机器学习优化 confidence 计算**：基于历史数据训练模型，预测「需要人工介入」的概率
2. **动态阈值调整**：根据团队风险容忍度，允许配置 `CONFIDENCE_THRESHOLD_HIGH`（默认 0.80）
3. **抽查智能化**：基于文件变更热力图（历史 bug 密度）优先抽查高风险文件
