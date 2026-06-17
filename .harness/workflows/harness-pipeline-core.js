// ============================================================
// Harness Pipeline Core — Orchestration Logic
// Prompts are in .harness/agents/prompts/{generator,qa,review,debug}.js
// Build: bash .harness/scripts/build-pipeline.sh
// ============================================================

export const meta = {
  name: 'harness-pipeline',
  description: 'Harness 开发管线：Generator → QA → (Debug) → Reviewer，QA FAIL 时触发根因分析，失败自动回到 Generator 修复直到 PASS',
  phases: [
    { title: 'Develop', detail: 'Generator 实现/修复代码' },
    { title: 'QA', detail: 'QA Agent 编译+测试验证' },
    { title: 'Debug', detail: '根因分析 — QA FAIL 时触发（systematic-debugging）' },
    { title: 'Review', detail: '3 视角并行审查（安全架构/规范工程/设计业务）' },
  ],
}

// args: { serviceName: "审核服务", serviceDir: "services/moderation-service", task: "实现 gRPC 层" }

// SEE: [[harness-pipeline-undefined-guard]] — 防止 args.serviceDir 未传时字符串化为 "undefined"
const VALID_SERVICES = [
  'ai-model-service', 'auth-service', 'community-hub-service',
  'file-service', 'master-data-service', 'moderation-service',
  'monitoring-service', 'permission-service', 'user-service',
]
const VALID_WEB = ['pc', 'mobile', 'common']
const ALL_VALID = [...VALID_SERVICES, ...VALID_WEB]

const isMissing = (v) => !v || v === 'undefined' || typeof v !== 'string'

// ── Auto-detect: try to resolve serviceName/serviceDir from task text ──
function resolveService(taskText) {
  if (!taskText || typeof taskText !== 'string') return null
  const lower = taskText.toLowerCase()
  for (const svc of VALID_SERVICES) {
    // Match hyphenated name (moderation-service) or space-separated (moderation service)
    const bare = svc.toLowerCase()
    if (lower.includes(bare) || lower.includes(bare.replace(/-/g, ' '))) {
      return { name: svc, dir: `services/${svc}` }
    }
  }
  for (const web of VALID_WEB) {
    if (lower.includes(`web/${web}`) || lower.includes(`前端 ${web}`) || lower.includes(`${web} 端`)) {
      return { name: `web-${web}`, dir: `web/${web}` }
    }
  }
  return null
}

// ── Service name → Chinese label ──
function serviceLabel(name) {
  const m = {
    'ai-model-service': 'AI模型服务', 'auth-service': '认证服务',
    'community-hub-service': '社区枢纽服务', 'file-service': '文件服务',
    'master-data-service': '主数据服务', 'moderation-service': '内容审核服务',
    'monitoring-service': '监控服务', 'permission-service': '权限服务',
    'user-service': '用户服务',
  }
  return m[name] || name
}

// ── Step A: auto-detect from task text ──
if (isMissing(args.serviceName) || isMissing(args.serviceDir)) {
  const detected = resolveService(args.task || '')
  if (detected) {
    if (isMissing(args.serviceName)) {
      log(`args.serviceName 缺失，从 task 自动解析 → "${serviceLabel(detected.name)}"`)
      args.serviceName = serviceLabel(detected.name)
    }
    if (isMissing(args.serviceDir)) {
      log(`args.serviceDir 缺失，从 task 自动解析 → "${detected.dir}"`)
      args.serviceDir = detected.dir
    }
  }
}

// ── Step B: explicit validation if still missing ──
if (isMissing(args.serviceName)) {
  throw new Error(
    `harness-pipeline: 缺少 args.serviceName（如 "内容审核服务"）。\n` +
    `请在 task 描述中包含服务名称，或显式传入 serviceName。\n` +
    `可用服务: ${VALID_SERVICES.join(', ')}`
  )
}
if (isMissing(args.serviceDir)) {
  throw new Error(
    `harness-pipeline: 缺少 args.serviceDir（如 "services/moderation-service"）。\n` +
    `请在 task 描述中包含服务路径，或显式传入 serviceDir。\n` +
    `可用: services/* 或 web/*`
  )
}
if (isMissing(args.task)) {
  throw new Error(
    `harness-pipeline: 缺少 args.task。请提供任务描述。`
  )
}

// ── Step C: validate the resolved dir maps to a known service ──
const bareName = args.serviceDir.split('/').pop()
if (!ALL_VALID.includes(bareName)) {
  throw new Error(
    `harness-pipeline: 无法识别的服务 "${bareName}"（路径 "${args.serviceDir}"）。\n` +
    `可用: ${ALL_VALID.join(', ')}`
  )
}

const ROOT_CLAUDE = 'CLAUDE.md'  // Agent 在 repo 根目录运行，用相对路径
const SVC_DIR = args.serviceDir   // 如 "services/moderation-service"，agent 在 repo 根目录
const SVC_NAME = bareName         // "moderation-service" — QA 脚本等需要裸目录名


// ============================================================
// Pipeline orchestration loop
// ============================================================

// ============================================================
// Pipeline loop
// ============================================================

const MAX_ITERATIONS = 3
let iteration = 1
let fixContext = ''

while (iteration <= MAX_ITERATIONS) {
  log(`第 ${iteration} 轮`)

  // Phase 1: Generator (隔离 worktree，避免并行管线互踩文件)
  phase('Develop')
  await agent(generatorPrompt(iteration, fixContext), { label: `${args.serviceName}: 开发/修复`, isolation: 'worktree' })
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
    // Phase 2.5: Debug — 根因分析（systematic-debugging），仅 QA FAIL 时触发
    phase('Debug')
    const debugResult = await agent(debuggingPrompt(qaResult), {
      label: `Debug: ${args.serviceName}`,
      schema: DEBUG_SCHEMA,
    })

    if (debugResult) {
      log(`Debug 根因: ${debugResult.rootCause} (置信度: ${debugResult.confidence || 'N/A'})`)
      const evidence = (debugResult.evidence || []).map(e => `- ${e}`).join('\n')
      const suggestions = (debugResult.fixSuggestions || []).map(s => `- ${s}`).join('\n')
      fixContext = `### QA 测试失败（已通过 systematic-debugging 分析根因）\n${qaResult.summary}\n\n### 根因分析\n**根因**: ${debugResult.rootCause}\n**置信度**: ${debugResult.confidence || 'N/A'}\n\n**证据链**:\n${evidence}\n\n**修复建议**:\n${suggestions}\n\n失败详情：\n${JSON.stringify(qaResult.failures, null, 2)}`
    } else {
      log('Debug Agent 被跳过，使用原始错误信息')
      fixContext = `### QA 测试失败\n${qaResult.summary}\n\n失败详情：\n${JSON.stringify(qaResult.failures, null, 2)}`
    }
    iteration++
    continue
  }

  // Phase 3: Multi-Perspective Review (并行，仅 QA PASS 后)
  phase('Review')

  // 启动三个视角的 Reviewer，并行执行
  const reviewResults = await parallel(
    REVIEW_LENSES.map(lens => () =>
      agent(reviewLensPrompt(lens), { label: `Review:${lens.key}`, schema: REVIEW_SCHEMA })
        .then(r => r ? { ...r, lens: lens.key, label: lens.label } : null)
    )
  )

  const validReviews = reviewResults.filter(Boolean)
  if (validReviews.length === 0) {
    log('所有 Review Agent 被跳过，终止管线')
    return { status: 'aborted', reason: 'All review agents skipped' }
  }

  // 统计各视角结果
  const passCount = validReviews.filter(r => r.verdict === 'PASS').length
  const failCount = validReviews.filter(r => r.verdict === 'FAIL').length

  log(`多视角审查完成: ${passCount}/${validReviews.length} PASS`)
  for (const r of validReviews) {
    log(`  ${r.lens} (${r.label}): ${r.verdict} — ${r.summary}`)
  }

  // 投票判定
  if (passCount >= 2) {
    // 2/3 或 3/3 PASS → 管线通过
    if (failCount > 0) {
      const failingLenses = validReviews.filter(r => r.verdict === 'FAIL').map(r => r.label).join('、')
      log(`⚠️ ${failingLenses}视角有异议，但多数通过 (${passCount}/${validReviews.length})，管线继续`)
    }
    log(`✅ 多视角 Review PASS (${passCount}/${validReviews.length})`)

    // 汇总所有 WARNING/NOTE 供参考
    const allWarnings = validReviews.reduce((sum, r) => sum + (r.warningCount || 0), 0)
    const reviewSummary = validReviews.map(r => `${r.label}: ${r.verdict}`).join(', ')

    // PASS — 管线完成！
    log(`✅ Harness 管线完成！${args.serviceName} 通过全部检查 (${iteration} 轮)`)
    const passNotifications = [{ event: 'pipeline_pass', service: args.serviceName, detail: `${iteration} 轮通过, QA: ${qaResult.summary}` }]
    if (failCount > 0) {
      passNotifications.push({ event: 'need_human', service: args.serviceName, detail: `多数 PASS (${passCount}/${validReviews.length}) 但 ${failingLenses} 有异议` })
    }
    return {
      status: 'pass',
      iterations: iteration,
      serviceName: args.serviceName,
      qaSummary: qaResult.summary,
      reviewSummary: `${reviewSummary} (${passCount}/${validReviews.length} PASS, ${allWarnings} WARNINGs)`,
      notifications: passNotifications,
    }
  } else {
    // 0/3 或 1/3 PASS → 管线失败
    const allCriticals = validReviews
      .filter(r => r.verdict === 'FAIL')
      .flatMap(r => (r.criticalIssues || []).map(i => ({ ...i, lens: r.lens })))

    const criticalLines = allCriticals.map(c => `- [${c.lens}] ${c.file}: ${c.issue}`).join('\n')
    const failSummary = validReviews.map(r => `${r.label}: ${r.verdict} (${r.criticalCount || 0} CRITICAL)`).join('\n')

    fixContext = `### 多视角审查未通过 (${passCount}/${validReviews.length} PASS)\n${failSummary}\n\n需要修复的 CRITICAL:\n${criticalLines}`
    log(`❌ 多视角 Review FAIL (${passCount}/${validReviews.length})`)
    iteration++
    continue
  }
}

// Max iterations exceeded
log(`❌ 超过最大轮次 (${MAX_ITERATIONS})，管线终止`)
return {
  status: 'max_iterations_exceeded',
  iterations: iteration,
  serviceName: args.serviceName,
  lastFixContext: fixContext,
  notifications: [{ event: 'pipeline_fail', service: args.serviceName, detail: `超过最大轮次 (${MAX_ITERATIONS}), 最后错误: ${fixContext.substring(0, 200)}` }],
}
