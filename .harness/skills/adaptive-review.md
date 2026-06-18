# Skill: Adaptive Review — 置信度自适应审查

## 职责

在 Owner Agent 阶段 5 验收环节，根据 Workflow 返回的 confidence 评分，自动决定审查深度。

## 输入

从 Workflow 返回值中获取：
```javascript
{
  status: 'pass',
  confidence: 0.85,  // 0.0-1.0
  serviceName: '审核服务',
  serviceDir: 'services/moderation-service',
  qaSummary: '...',
  reviewSummary: '...',
  iterations: 1
}
```

## 三级审查模式

### Level 1: 高置信（confidence ≥ 0.80）— 摘要审查

**审查内容**：
1. 读取 `{serviceDir}/_qa.md` 的前 50 行（摘要章节）
2. 读取 `{serviceDir}/_review_*.md` 的「摘要」章节和「🔴 CRITICAL」表格
3. 检查点：
   - QA verdict = PASS？
   - Review 无 CRITICAL？
   - CHANGELOG 已更新？

**通过条件**：3 个检查点全 ✅

**失败处理**：任一检查点 ❌ → 降级到「抽查」模式

**执行命令**：
```bash
# 提取 QA 摘要
bash .harness/scripts/extract-summary.sh --file {serviceDir}/_qa.md --lines 50

# 提取 Review 摘要
for review_file in {serviceDir}/_review_*.md; do
  bash .harness/scripts/extract-summary.sh --file "$review_file" --section "摘要"
done

# 检查 CHANGELOG
git diff HEAD -- {serviceDir}/CHANGELOG.md | grep "^+" | wc -l
```

**耗时**：~5 分钟

---

### Level 2: 中置信（0.50 ≤ confidence < 0.80）— 抽查

**抽样策略**：
```bash
# 获取变更文件并随机抽样
bash .harness/scripts/spot-check.sh --service {serviceName} --seed {taskId} --ratio 0.3 --min 2
```

**审查内容**（每个抽样文件）：
1. **全文阅读代码**
2. **检查 Memory 遵守**：有 `[[slug]]` 引用？引用正确？
3. **检查 TDD 证据**：新增函数有对应测试？
4. **检查边界条件**：nil/zero/empty 处理？
5. **检查错误处理**：error 是否正确传播？

**记录**：写入 `.harness/changes/{change}/impl/{service}/spot_check.md`
```markdown
## 抽查报告

**抽查时间**: 2026-06-18 15:30  
**Confidence**: 0.65  
**抽样文件**: 3/10 个文件（30%）

### 抽查结果

#### ✅ services/moderation-service/internal/logic/check_logic.go
- Memory 引用: [[grpc-timeout-layers]] 正确使用
- TDD 证据: CheckText_TableDriven 测试覆盖
- 边界条件: nil 校验完整
- 错误处理: 正确传播 grpc status

#### ✅ services/moderation-service/internal/handler/check_handler.go
- Memory 引用: [[api-response-single-wrap]] 正确应用
- TDD 证据: handler 测试覆盖
- 边界条件: 请求参数校验完整
- 错误处理: 正确映射到 HTTP 状态码

#### ⚠️ services/moderation-service/internal/types/types.go
- WARNING: 新增字段 `ModelType` 缺少 JSON tag
- 建议: 添加 `json:"model_type"`

### 总结

- **抽查发现**: 1 个 WARNING，0 个 CRITICAL
- **审查结果**: PASS（无 CRITICAL 阻塞项）
- **建议修复**: ModelType 字段补充 JSON tag
```

**通过条件**：无 CRITICAL 发现

**失败处理**：发现 ≥1 CRITICAL → 强制回退到阶段 5 编码

**耗时**：~15 分钟

---

### Level 3: 低置信（confidence < 0.50）— 全文审查 + 人工确认

**审查内容**：
1. **阅读所有变更文件**（全文）
2. **9 维度审查**：
   - 架构：分层是否清晰？依赖方向正确？
   - 设计：接口设计合理？可扩展性如何？
   - 规范：命名、注释、代码风格
   - 质量：边界条件、错误处理、资源清理
   - 安全：SQL 注入、XSS、权限校验
   - 复用：是否重复造轮？common/ 是否利用？
   - 测试：覆盖率、测试质量、TDD 证据
   - 变更：变更范围是否最小？是否有过度重构？
   - 记忆：相关记忆是否遵守？

3. **输出问题清单**

**人工确认**：
```markdown
## 全文审查报告（需人工确认）

**Confidence**: 0.45（低置信，需全文审查）  
**审查人**: Owner Agent  
**审查时间**: 2026-06-18 15:30

### 发现问题

#### 🔴 CRITICAL
1. **缺少 nil 校验**（check_logic.go:145）
   - 问题: req.Text 可能为 nil，直接解引用会 panic
   - 建议: 添加 `if req.Text == nil { return errx.New(...) }`

2. **数据库事务未提交**（update_logic.go:89）
   - 问题: 开启事务后未调用 Commit
   - 建议: defer tx.Rollback() + 成功时 tx.Commit()

#### ⚠️ WARNING
3. **测试覆盖不完整**（check_logic_test.go）
   - 问题: 只测试了正常路径，缺少错误路径测试
   - 建议: 补充 nil 输入、gRPC 超时、敏感词匹配等测试用例

4. **硬编码配置**（handler.go:23）
   - 问题: timeout 硬编码为 30s
   - 建议: 从 config.yaml 读取

### 建议

**选项 A: 回退修复**（推荐）
- 修复 2 个 CRITICAL 问题
- 补充测试覆盖
- 预计耗时: 1-2 小时

**选项 B: 继续交付（接受风险）**
- CRITICAL 问题已记录到 BACKLOG
- 创建 follow-up 任务（task-2026-06-18-xxx）
- 需要您显式批准

---

请决策：
- 输入 "A" → 回退到阶段 5 编码
- 输入 "B + 理由" → 继续交付（需说明接受风险的理由）
```

**暂停等待用户输入**

**耗时**：~30 分钟 + 人工决策时间

---

## 实现伪代码

```javascript
// Owner Agent 阶段 5 验收逻辑
for (const result of workflowResults) {
  const {confidence, serviceName, serviceDir} = result
  
  let reviewMode = ''
  let reviewResult = 'PASS'
  let humanReviewNeeded = false
  
  // ═══ Level 1: 高置信 ≥ 0.80 ═══
  if (confidence >= 0.80) {
    log(`${serviceName}: 高置信 (${confidence})，执行摘要审查`)
    
    // 检查 QA 摘要
    const qaSummary = await bash(`bash .harness/scripts/extract-summary.sh --file ${serviceDir}/_qa.md --lines 50`)
    const qaPass = qaSummary.includes('VERDICT: PASS')
    
    // 检查 Review 摘要
    const reviewFiles = await bash(`ls ${serviceDir}/_review_*.md`)
    let reviewPass = true
    for (const file of reviewFiles) {
      const summary = await bash(`bash .harness/scripts/extract-summary.sh --file ${file} --section "摘要"`)
      if (summary.includes('CRITICAL:') && !summary.includes('CRITICAL: 0')) {
        reviewPass = false
        break
      }
    }
    
    // 检查 CHANGELOG
    const changelogDiff = await bash(`git diff HEAD -- ${serviceDir}/CHANGELOG.md | grep "^+" | wc -l`)
    const changelogUpdated = parseInt(changelogDiff) > 0
    
    if (qaPass && reviewPass && changelogUpdated) {
      log(`✅ ${serviceName} 摘要审查通过`)
      reviewMode = 'summary'
      humanReviewNeeded = false
    } else {
      log(`⚠️ ${serviceName} 摘要审查发现异常，降级到抽查`)
      confidence = 0.70  // 降级
    }
  }
  
  // ═══ Level 2: 中置信 0.50-0.79 ═══
  if (confidence >= 0.50 && confidence < 0.80) {
    log(`${serviceName}: 中置信 (${confidence})，执行抽查`)
    
    // 抽样文件
    const samples = await bash(`bash .harness/scripts/spot-check.sh --service ${serviceName} --seed ${taskId}`)
    const sampleFiles = samples.split('\n').filter(f => f.length > 0)
    
    log(`抽查文件: ${sampleFiles.join(', ')}`)
    
    const criticals = []
    for (const file of sampleFiles) {
      // 读取文件全文
      const content = await Read(file)
      
      // 9 维度审查（简化版）
      const issues = deepReviewFile(content, file)
      criticals.push(...issues.filter(i => i.level === 'CRITICAL'))
    }
    
    // 写入抽查报告
    const report = generateSpotCheckReport(sampleFiles, criticals)
    await Write(`${serviceDir}/spot_check.md`, report)
    
    if (criticals.length > 0) {
      log(`❌ ${serviceName} 抽查发现 ${criticals.length} 个 CRITICAL`)
      reviewMode = 'spot_check_failed'
      reviewResult = 'FAIL'
      humanReviewNeeded = true
    } else {
      log(`✅ ${serviceName} 抽查通过`)
      reviewMode = 'spot_check'
      humanReviewNeeded = false
    }
  }
  
  // ═══ Level 3: 低置信 < 0.50 ═══
  if (confidence < 0.50) {
    log(`${serviceName}: 低置信 (${confidence})，要求全文审查`)
    
    // 获取所有变更文件
    const changedFiles = await bash(`git diff --name-only HEAD -- ${serviceDir}`)
    
    const allIssues = []
    for (const file of changedFiles) {
      const content = await Read(file)
      const issues = deepReviewFile(content, file)  // 完整 9 维度
      allIssues.push(...issues)
    }
    
    // 生成全文审查报告
    const report = generateFullReviewReport(allIssues)
    await Write(`${serviceDir}/full_review.md`, report)
    
    // 暂停，等待用户决策
    reviewMode = 'full_review'
    humanReviewNeeded = true
    
    log(`⏸️  ${serviceName} 全文审查完成，等待人工确认`)
    log(report)
    
    const decision = await AskUserQuestion({
      questions: [{
        question: `${serviceName} 低置信审查发现 ${allIssues.filter(i => i.level === 'CRITICAL').length} 个 CRITICAL 问题。如何处理？`,
        header: "决策",
        multiSelect: false,
        options: [
          {label: "回退修复", description: "返回阶段 5 编码，修复 CRITICAL 问题"},
          {label: "继续交付", description: "接受风险，创建 follow-up 任务"}
        ]
      }]
    })
    
    if (decision.answers[0] === "回退修复") {
      reviewResult = 'FAIL'
    } else {
      reviewResult = 'PASS (人工批准)'
      // 创建 follow-up 任务
      await bash(`bash .harness/scripts/harness-tasks.sh create --title "修复 ${serviceName} 审查发现的问题"`)
    }
  }
  
  // ═══ 更新 summary.md ═══
  const summaryUpdate = `
## 阶段 5 验收
- Confidence: ${confidence}
- 审查模式: ${reviewMode}
- 审查结果: ${reviewResult}
- 审查人: ${humanReviewNeeded ? 'Human' : 'Owner Agent (自动)'}
- 审查时间: ${new Date().toISOString()}
`
  
  await appendToSummary(summaryUpdate)
}
```

## 检查函数

### checkQASummary(serviceDir)
```bash
SUMMARY=$(bash .harness/scripts/extract-summary.sh --file ${serviceDir}/_qa.md --lines 50)
if echo "$SUMMARY" | grep -q "VERDICT: PASS"; then
  return 0  # PASS
else
  return 1  # FAIL
fi
```

### checkReviewSummary(serviceDir)
```bash
CRITICAL_COUNT=0
for review_file in ${serviceDir}/_review_*.md; do
  SUMMARY=$(bash .harness/scripts/extract-summary.sh --file "$review_file" --section "摘要")
  COUNT=$(echo "$SUMMARY" | grep -oP "CRITICAL: \K\d+" || echo "0")
  CRITICAL_COUNT=$((CRITICAL_COUNT + COUNT))
done

if [ "$CRITICAL_COUNT" -eq 0 ]; then
  return 0  # PASS
else
  return 1  # FAIL
fi
```

### checkChangelog(serviceDir)
```bash
ADDED_LINES=$(git diff HEAD -- ${serviceDir}/CHANGELOG.md | grep "^+" | wc -l)
if [ "$ADDED_LINES" -gt 0 ]; then
  return 0  # 已更新
else
  return 1  # 未更新
fi
```

---

## 使用方式

在 Owner Agent 的阶段 5 验收中，加载此 Skill：

```markdown
## 阶段 5: 验收（多服务结果汇总）

**加载 Skill**: `.harness/skills/adaptive-review.md`

对每个 Workflow 返回值，执行分级审查...
```

## 下一步

Phase 3: 测试验证
- 准备 3 个测试场景（高/中/低置信）
- 端到端测试
- 回归测试
