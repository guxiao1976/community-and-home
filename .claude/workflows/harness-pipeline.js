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

const REV_DIR = '/home/jiaoxh/my-project/community-home/reviewers'
const ROOT_CLAUDE = '/home/jiaoxh/my-project/community-home/CLAUDE.md'
const SVC_DIR = `/home/jiaoxh/my-project/community-home/${args.serviceDir}`

// ============================================================
// Agent prompts
// ============================================================

function generatorPrompt(iteration, fixContext) {
  const base = `你是 ${args.serviceName} 的开发 Agent。

## 启动上下文（必须先读，顺序重要）
1. 阅读 ${SVC_DIR}/CLAUDE.md — 角色定位、关键规则、全局公约、常用命令
2. 阅读 ${SVC_DIR}/docs/design.md — 数据模型、业务流程（如存在）
3. 阅读 ${SVC_DIR}/CHANGELOG.md — 变更历史
4. **读取 .claude/memory/MEMORY.md** — 加载全局历史经验，避免重复已知错误
5. 根据任务关键词，精读匹配的记忆文件内容

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
- 更新 ${SVC_DIR}/CHANGELOG.md`
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
阅读 ${REV_DIR}/qa-engineer/CLAUDE.md — 你的角色定义、验证步骤和产出格式。

## 验证目标
验证 ${SVC_DIR}/ 的代码质量。

## 验证步骤
1. 阅读 ${SVC_DIR}/CLAUDE.md — 服务规则
2. 阅读 ${SVC_DIR}/CHANGELOG.md — 变更历史
3. 运行 go build ./... — 编译检查
4. 运行 go vet ./... — 静态分析
5. 运行 go test ./... -count=1 — 单元测试
6. 检查新增代码的测试覆盖
7. 写入 ${SVC_DIR}/_qa.md
8. 输出 VERDICT

## 约束
- 只读权限：Read、Grep、Glob、Bash（go build、go vet、go test）
- 严禁 Write、Edit

## 记忆记录
如果 QA 判定为 FAIL，你必须：
1. 分析根本原因（不是表面错误信息）
2. 检查 .claude/memory/MEMORY.md 是否已有相关经验
3. 如果有 → 更新该记忆文件（增加新的复现场景）
4. 如果没有 → 创建新的记忆文件到 .claude/memory/<slug>.md
5. 更新 MEMORY.md 索引
记忆文件格式：参见 .claude/memory/MEMORY.md 中的说明
`
}

function reviewPrompt() {
  return `你是 Code Reviewer Agent。

## 角色定义（必须先读）
阅读 ${REV_DIR}/code-reviewer/CLAUDE.md — 你的角色定义、审查规则和产出格式。

## 审查目标
审查 ${SVC_DIR}/ 的代码变更（QA 已通过，_qa.md 可供参考）。

## 审查步骤
1. 阅读 ${ROOT_CLAUDE} — 全局规则
2. 阅读 ${SVC_DIR}/CLAUDE.md — 服务规则
3. 阅读 ${SVC_DIR}/docs/design.md — 设计文档
4. 阅读 ${SVC_DIR}/CHANGELOG.md — 变更历史
5. 阅读 ${SVC_DIR}/_qa.md — QA 报告（了解测试结果）
6. 获取变更内容（git diff 或审查变更文件）
7. 按 8 个维度审查
8. 写入 ${SVC_DIR}/_review.md
9. 输出 VERDICT

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
2. 如果是新经验 → 创建记忆文件到 .claude/memory/<slug>.md
3. 如果已有相关记忆 → 更新它
4. 更新 MEMORY.md 索引
`

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
