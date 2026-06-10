export const meta = {
  name: 'harness-pipeline',
  description: 'Harness 开发管线：Generator → QA → Reviewer，失败自动回到 Generator 修复，直到 PASS',
  phases: [
    { title: 'Develop', detail: 'Generator 实现/修复代码' },
    { title: 'QA', detail: 'QA Agent 编译+测试验证' },
    { title: 'Review', detail: 'Code Reviewer 8 维度审查' },
  ],
}

// args: { serviceName: "审核服务", serviceDir: "services/moderation-service", task: "实现 gRPC 层" }

const ROOT_CLAUDE = '/home/jiaoxh/my-project/community-home/CLAUDE.md'
const SVC_DIR = `/home/jiaoxh/my-project/community-home/${args.serviceDir}`

// ============================================================
// Agent prompts
// ============================================================

function generatorPrompt(iteration, fixContext) {
  const base = `你是 ${args.serviceName} 的开发 Agent。

## 启动上下文（必须先读，顺序重要）
1. 阅读 /home/jiaoxh/my-project/community-home/CLAUDE.md — 全局索引/地图，了解 rules/ memory/ changes/ 位置
2. 阅读 .harness/rules/项目编码规范.md — 编码规范、硬性约束（Snowflake、gRPC、提交前检查等）
3. 阅读 ${SVC_DIR}/CLAUDE.md — 角色定位、关键规则、全局公约、常用命令
4. 阅读 ${SVC_DIR}/docs/design.md — 数据模型、业务流程（如存在）
5. 阅读 ${SVC_DIR}/CHANGELOG.md — 变更历史
6. **读取 .harness/knowledge/memory/MEMORY.md** — 加载全局历史经验，避免重复已知错误
7. 根据任务关键词，精读匹配的记忆文件内容

## 记忆驱动编码（编码前必须执行）

在开始编写代码之前，你必须完成以下步骤：

### Step A: 搜索相关记忆（两级匹配）
1. 从任务描述中提取关键技术关键词（如 gRPC、Proto、数据库、JWT、Snowflake 等）
2. **第一级：triggers 精确匹配（优先）**
   - 读取 .harness/knowledge/memory/MEMORY.md 索引，获取所有记忆的 triggers 列表（格式：\`记忆标题, type, severity, keyword1 keyword2...\`）
   - 用任务关键词精确匹配索引中的 triggers 关键词
   - 命中 triggers 的记忆 → **高置信度**，直接列入候选
3. **第二级：正文关键词匹配（降权，需人工判断）**
   - 仅当第一级匹配结果 < 2 个时，才使用 Grep 搜索正文
   - 正文命中的记忆 → **低置信度**，需检查其 \`type\` 和 \`severity\` 是否与任务相关
   - 过滤规则：
     - \`type: pitfall\` 且 triggered 关键词不在任务技术栈中 → 排除
     - \`type: guideline\` 且 service 范围不匹配 → 排除
4. 列出找到的相关记忆及其 severity 等级（must-follow / should-follow / info）和 type（pitfall / guideline / process）

### Step B: 应用记忆
1. 对于每个 must-follow 记忆，确保生成的代码严格遵守其指导
2. 在应用记忆的代码位置，添加注释标记：
   \`\`\`
   // SEE: [[memory-slug]] — <简短说明为什么这条记忆适用于此处>
   \`\`\`
   其中 memory-slug 是记忆文件名（不含 .md 扩展名）
3. 对于 should-follow 记忆，判断是否适用当前任务，适用则同样标记

### Step C: 编码总结
在编码完成后，输出记忆应用报告：
\`\`\`
### 记忆应用报告
- 搜索关键词: <关键词列表>
- 找到相关记忆: <数量>
- 已应用:
  - [[memory-slug-1]] — 应用于 <文件名:行号> — <原因>
  - [[memory-slug-2]] — 应用于 <文件名:行号> — <原因>
- 未应用（不适用当前任务）:
  - [[memory-slug-3]] — <不适用的原因>
\`\`\`

## 全局公约提醒
- Proto 变更必须在 api-proto/ 中操作，告知用户切换到全局 Claude 执行 make generate
- 服务间通信仅通过 gRPC，禁止直连其他服务数据库
- 不修改 common/ 和 api-proto/

## 任务`

  if (iteration === 1) {
    return base + `\n${args.task}

## 完成标准
- go build ./... 通过
- go test ./... 通过
- 更新 ${SVC_DIR}/CHANGELOG.md
- 输出记忆应用报告`
  } else {
    return base + `\n修复上一阶段发现的问题。请阅读以下报告中的失败项并逐一修复：

${fixContext}

## 完成标准
- 所有失败项修复完成
- go build ./... 通过
- go test ./... 通过
- 更新 ${SVC_DIR}/CHANGELOG.md`
  }
}

function qaPrompt() {
  return `你是 QA Engineer Agent。

## 角色定义（必须先读）
阅读 .harness/skills/qa.md — 你的角色定义、验证步骤和产出格式。

## 验证目标
验证 ${SVC_DIR}/ 的代码质量。

## 验证步骤
1. 阅读 ${SVC_DIR}/CLAUDE.md — 服务规则
2. 阅读 ${SVC_DIR}/CHANGELOG.md — 变更历史
3. **运行机械化检查**: \`bash .harness/skills/qa/scripts/harness-checks.sh --service ${args.serviceName} --json\`
   - 解析 JSON 输出，将各项检查结果整合到 QA 报告的「机械化检查结果」章节
   - FAIL 项在报告中标注具体违规（文件名:行号:字段名）
   - WARN 项作为 WARNING 级别记录
4. 运行 go build ./... — 编译检查（机械化检查已覆盖，确认结果）
5. 运行 go vet ./... — 静态分析（机械化检查已覆盖，确认结果）
6. 运行 go test ./... -count=1 — 单元测试（机械化检查已覆盖，确认结果）
7. 检查新增代码的测试覆盖
8. 写入 ${SVC_DIR}/_qa.md
9. 输出 VERDICT

## QA 报告必须包含机械化检查结果
在 _qa.md 中新建「机械化检查结果」章节，格式：
\`\`\`markdown
## 机械化检查结果 (harness-checks.sh)

| # | 检查项 | 结果 | 详情 |
|---|--------|------|------|
| 1 | go build | ✅/❌ | <详情> |
| 2 | go vet | ✅/❌ | <详情> |
| 3 | go test | ✅/❌ | <详情，含测试数量> |
| 4 | Proto int64 jstype | ✅/❌ | <违规数量> violations |
| 5 | json:",string" | ✅/❌ | <违规数量> violations |
| 6 | 跨服务DB导入 | ✅/❌ | <详情> |
| 7 | 错误码格式 | ✅/⚠️ | <详情> |
| 8 | 硬编码密钥 | ✅/❌ | <详情> |
\`\`\`

## 约束
- 只读权限：Read、Grep、Glob、Bash（go build、go vet、go test、bash .harness/skills/qa/scripts/harness-checks.sh）
- 严禁 Write、Edit

## 记忆记录
如果 QA 判定为 FAIL，你必须：
1. 分析根本原因（不是表面错误信息）
2. 检查 .harness/knowledge/memory/MEMORY.md 是否已有相关经验
3. 如果有 → 更新该记忆文件（增加新的复现场景）
4. 如果没有 → 创建新的记忆文件到 .harness/knowledge/memory/<slug>.md
5. 更新 MEMORY.md 索引
记忆文件格式：参见 .harness/knowledge/memory/MEMORY.md 中的说明
`
}

function reviewPrompt() {
  return `你是 Code Reviewer Agent。

## 角色定义（必须先读）
阅读 .harness/skills/review.md — 你的角色定义、9维度审查规则和产出格式。

## 审查目标
审查 ${SVC_DIR}/ 的代码变更（QA 已通过，_qa.md 可供参考）。

## 审查步骤
1. 阅读 ${ROOT_CLAUDE} — 全局规则
2. 阅读 ${SVC_DIR}/CLAUDE.md — 服务规则
3. 阅读 ${SVC_DIR}/docs/design.md — 设计文档
4. 阅读 ${SVC_DIR}/CHANGELOG.md — 变更历史
5. 阅读 ${SVC_DIR}/_qa.md — QA 报告（了解测试结果）
6. 获取变更内容（git diff 或审查变更文件）
7. 按 9 个维度审查（含记忆遵守）
8. 写入 ${SVC_DIR}/_review.md
9. 输出 VERDICT

## 第 9 维度：记忆遵守（Memory Compliance）

审查时必须验证代码是否遵守项目记忆系统中的经验：

### M1: 收集代码中的记忆引用
- Grep 变更文件中的 \`// SEE: [[\` 注释，提取所有被引用的 memory-slug

### M2: 验证引用准确性
对每个 \`// SEE: [[memory-slug]]\` 注释：
- 读取对应的 \`.harness/knowledge/memory/<slug>.md\` 或 \`${SVC_DIR}/.harness/knowledge/memory/<slug>.md\`
- slug 文件不存在 → 🔴 CRITICAL
- 代码未遵守记忆指导 → 🔴 CRITICAL
- 虚假匹配（记忆不适用于此上下文）→ 🟡 WARNING

### M3: 检查遗漏的记忆（两级匹配）
- 从变更描述和 git diff 提取技术关键词
- **第一级**：用关键词精确匹配 MEMORY.md 索引中的 triggers 列表
- **第二级**：仅当第一级匹配结果 < 2 个时，才 Grep 正文
- 关键词命中 triggers 但代码未引用且应适用 → 🔴 CRITICAL
- should-follow 记忆遗漏 → 🟡 WARNING
- 注意：使用记忆的 \`type\` 字段过滤（pitfall 类仅当技术栈匹配时才告警）

### M4: 更新记忆元数据
对于被正确应用的记忆，在 _review.md 报告中标注需要更新 last_applied / apply_count。

## 约束
- 只读权限：Read、Grep、Glob、Bash（go build、go vet、go test、git diff 等）
- 严禁 Write、Edit

## QA VERDICT 解析
- QA PASS — 继续到 Review
- QA FAIL — 回到 Generator 修复

## 审查 VERDICT 解析
- Review PASS — 管线完成！
- Review FAIL — 回到 Generator 修复
- 如果 QA 通过但 Review 发现有测试相关问题，Generator 修复后需重新走 QA

## 记忆记录
如果 Review 发现 CRITICAL 问题，你必须：
1. 判断这是否是一个新的规范/踩坑（而非一次性的代码错误）
2. 如果是新经验 → 创建记忆文件到 .harness/knowledge/memory/<slug>.md
3. 如果已有相关记忆 → 更新它
4. 更新 MEMORY.md 索引
`

}

// ============================================================
// VERDICT schemas for structured output
// ============================================================

const QA_SCHEMA = {
  type: 'object',
  properties: {
    verdict: { type: 'string', enum: ['PASS', 'FAIL'] },
    summary: { type: 'string' },
    failures: {
      type: 'array',
      items: {
        type: 'object',
        properties: {
          step: { type: 'string' },
          error: { type: 'string' },
        },
      },
    },
  },
  required: ['verdict', 'summary'],
}

const REVIEW_SCHEMA = {
  type: 'object',
  properties: {
    verdict: { type: 'string', enum: ['PASS', 'FAIL'] },
    criticalCount: { type: 'number' },
    warningCount: { type: 'number' },
    summary: { type: 'string' },
    criticalIssues: {
      type: 'array',
      items: { type: 'object', properties: { file: { type: 'string' }, issue: { type: 'string' } } },
    },
  },
  required: ['verdict', 'summary'],
}

// ============================================================
// Pipeline loop
// ============================================================

const MAX_ITERATIONS = 3
let iteration = 1
let fixContext = ''

while (iteration <= MAX_ITERATIONS) {
  log(`第 ${iteration} 轮`)

  // Phase 1: Generator
  phase('Develop')
  await agent(generatorPrompt(iteration, fixContext), { label: `${args.serviceName}: 开发/修复` })
  log(`Generator 完成 (轮次 ${iteration})`)

  // Phase 2: QA
  phase('QA')
  const qaResult = await agent(qaPrompt(), { label: `QA: ${args.serviceName}`, schema: QA_SCHEMA })
  if (!qaResult) {
    log('QA Agent 被跳过，终止管线')
    return { status: 'aborted', reason: 'QA agent skipped' }
  }

  log(`QA VERDICT: ${qaResult.verdict} — ${qaResult.summary}`)

  if (qaResult.verdict === 'FAIL') {
    fixContext = `### QA 测试失败\n${qaResult.summary}\n\n失败详情：\n${JSON.stringify(qaResult.failures, null, 2)}`
    iteration++
    continue
  }

  // Phase 3: Review (only if QA passed)
  phase('Review')
  const reviewResult = await agent(reviewPrompt(), { label: `Review: ${args.serviceName}`, schema: REVIEW_SCHEMA })
  if (!reviewResult) {
    log('Review Agent 被跳过，终止管线')
    return { status: 'aborted', reason: 'Review agent skipped' }
  }

  log(`Review VERDICT: ${reviewResult.verdict} — ${reviewResult.summary}`)

  if (reviewResult.verdict === 'FAIL') {
    const issues = (reviewResult.criticalIssues || []).map(i => `- ${i.file}: ${i.issue}`).join('\n')
    fixContext = `### 审查未通过 (${reviewResult.criticalCount} CRITICAL, ${reviewResult.warningCount} WARNING)\n${reviewResult.summary}\n\n需要修复的 CRITICAL:\n${issues}`
    iteration++
    continue
  }

  // PASS!
  log(`✅ Harness 管线完成！${args.serviceName} 通过全部检查 (${iteration} 轮)`)
  return {
    status: 'pass',
    iterations: iteration,
    serviceName: args.serviceName,
    qaSummary: qaResult.summary,
    reviewSummary: reviewResult.summary,
  }
}

// Max iterations exceeded
log(`❌ 超过最大轮次 (${MAX_ITERATIONS})，管线终止`)
return {
  status: 'max_iterations_exceeded',
  iterations: iteration,
  serviceName: args.serviceName,
  lastFixContext: fixContext,
}
