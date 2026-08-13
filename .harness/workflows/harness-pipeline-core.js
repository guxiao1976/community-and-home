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

// ============================================================
// Service Registry Loader (replaces hardcoded VALID_SERVICES)
// ============================================================
// Sandbox-compat: the Workflow runtime provides no Node.js API (no require/fs/path/process).
// Guard all Node API usage so the pipeline degrades to args-driven mode instead of dying at load.
let fs = null, path = null
try { fs = require('fs'); path = require('path') } catch (e) { /* sandbox: no Node API */ }

function loadServiceRegistry() {
  // Fallback for sandboxed runs (no fs/process): derive a minimal registry from args,
  // so Step-C validation accepts the explicitly-passed service. Full-Node runs load services.json.
  const sandbox = () => {
    const bare = (typeof args !== 'undefined' && args && args.serviceDir) ? String(args.serviceDir).split('/').pop() : ''
    const name = (typeof args !== 'undefined' && args && args.serviceName) || ''
    return {
      services: bare ? [bare] : [],
      web: [],
      getService: () => (bare ? { name: bare, module: bare } : null),
      getServiceModule: () => bare || null,
    }
  }
  try {
    if (!fs || !path) return sandbox()
    const registryPath = path.join(process.cwd(), '.harness/registry/services.json')
    if (!fs.existsSync(registryPath)) {
      throw new Error(`Service registry not found. Run: bash .harness/scripts/build-service-registry.sh`)
    }
    const registry = JSON.parse(fs.readFileSync(registryPath, 'utf-8'))
    return {
      services: registry.services.map(s => s.name),
      web: registry.web.map(w => w.name),
      getService: (name) => registry.services.find(s => s.name === name),
      getServiceModule: (name) => registry.services.find(s => s.name === name)?.module || null,
    }
  } catch (e) {
    return sandbox()
  }
}

const ServiceRegistry = loadServiceRegistry()

// ── Pipeline metrics logger (best-effort, never blocks) ──
function logMetrics(record) {
  try {
    const logDir = path.join(process.cwd(), '.harness/logs/pipeline')
    if (!fs.existsSync(logDir)) fs.mkdirSync(logDir, { recursive: true })
    fs.appendFileSync(path.join(logDir, 'metrics.jsonl'), JSON.stringify(record) + '\n')
  } catch (e) { /* silent */ }
}
const VALID_SERVICES = ServiceRegistry.services
const VALID_WEB = ServiceRegistry.web
const ALL_VALID = [...VALID_SERVICES, ...VALID_WEB]
// ============================================================

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

// ── Service name → Chinese label (single source: registry/services.json) ──
function serviceLabel(name) {
  const svc = ServiceRegistry.getService(name)
  return (svc && svc.displayName) || name
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

// ── Task type routing — automatic inference ──
// Priority: explicit arg > embedded type: marker > keyword heuristics > source-based > default
function resolveTaskType() {
  // 0. Explicit override (backward compatible)
  if (args.taskType) return args.taskType

  const t = (args.task || '').toLowerCase()

  // 0.5. Embedded type: marker (from loop dispatch or human annotation)
  if (t.includes('type: chore') || t.includes('type:chore')) return 'chore'
  if (t.includes('type: debt') || t.includes('type:debt')) return 'debt'
  if (t.includes('type: bug') || t.includes('type:bug')) return 'bug'
  if (t.includes('type: feature') || t.includes('type:feature')) return 'feature'

  // 1. Source-based inference
  // review/qa findings are overwhelmingly debt (standards violations, mechanical fixes)
  if (t.includes('source: review') || t.includes('source: qa') || t.includes('source:qa')) {
    // Unless the description clearly indicates a bug
    if (/\b(bug|broken|crash|panic|竞态|race|nil pointer|空指针|死锁)\b/.test(t)) return 'bug'
    return 'debt'
  }
  // GitHub PR changes-requested → bug (something is broken)
  if (t.includes('source: github') && (t.includes('changes requested') || t.includes('修复 pr'))) {
    return 'bug'
  }
  // Sensor: graph freshness → chore, others → debt
  if (t.includes('source: sensor')) {
    if (t.includes('graph') || t.includes('图谱') || t.includes('sync') || t.includes('同步')) return 'chore'
    return 'debt'
  }

  // 3. Keyword heuristics (Chinese + English) — order matters: most specific first

  // chore: maintenance, ops, no code logic
  if (/\b(同步|sync|图谱|graph|清理|脚本|脚本|配置|config|docker|deploy|ci|文档|doc|readme|changelog|\.md|\.yml|\.yaml|\.env|依赖|dependency|更新依赖|升级|upgrade)\b/.test(t)) {
    // chore:ops (docker/deploy/ci/dependency) → needs Review for safety
    if (/\b(docker|deploy|ci|依赖|dependency|upgrade|升级)\b/.test(t)) return 'chore:ops'
    return 'chore'
  }

  // bug: broken behavior, crash, data corruption
  if (/\b(bug|fix|修复|broken|crash|panic|竞态|race|nil pointer|空指针|死锁|deadlock|goroutine leak|泄漏|内存|溢出|overflow|数组越界|index out|类型错误|type error|panic|fatal|崩溃|报错|不工作|无效|invalid|丢失|丢失数据|错[误乱]|异常)\b/.test(t)) return 'bug'

  // debt: standards, style, naming, format, technical debt
  if (/\b(debt|规范|格式|format|命名|rename|tag|jstype|json|string|响应格式|style|lint|warning|重构|refactor|统一|标准化|对齐|align|补测试|补文档|代码质量|复用|重复|duplicate|硬编码|hardcode|命名|魔法数字|magic number|TODO|stub)\b/.test(t)) return 'debt'

  // feature: new capability — catch-all for anything mentioning new things
  if (/\b(feature|新增|添加|实现|创建|新建|开发|add|implement|create|build|支持|功能|页面|组件|接口|api|endpoint|handler|migration)\b/.test(t)) return 'feature'

  // 4. Default: unknown → feature (safest, full pipeline)
  return 'feature'
}
const TASK_TYPE = resolveTaskType()
log(`任务类型: ${TASK_TYPE} (auto-inferred)`)

// ── Type-based constants ──
// chore:doc → 1 iter, 0 review (skip)
// chore:ops → 2 iter, 1 review (deploy/ci/docker changes need review)
// debt       → 2 iter, 1 review
// feature/bug → 3 iter, 3 review
const isChoreOps = TASK_TYPE === 'chore:ops'
const MAX_ITERATIONS = (TASK_TYPE === 'chore') ? 1 : (TASK_TYPE === 'debt' || isChoreOps) ? 2 : 3
let REVIEW_LENS_COUNT = (TASK_TYPE === 'chore') ? 0 : (TASK_TYPE === 'debt' || isChoreOps) ? 1 : 3
let REVIEW_PASS_THRESHOLD = (TASK_TYPE === 'debt' || isChoreOps) ? 1 : 2
const TDD_STRICT = TASK_TYPE === 'feature' || TASK_TYPE === 'bug'
const NEED_CHANGELOG = TASK_TYPE !== 'chore'

// ── Workload routing（dispatch Step 0 传入）──
// S(轻量): 保留 QA 15 项，跳过 Review（与 owner-agent 轻量Pipeline 对齐）
const WORKLOAD = (args.workload || '').toUpperCase()
if (WORKLOAD === 'S' && REVIEW_LENS_COUNT > 0) {
  log(`轻量管线（workload=S）— 跳过 Review，保留 QA 15 项`)
  REVIEW_LENS_COUNT = 0
  REVIEW_PASS_THRESHOLD = 0
}
let qaFirstPass = true  // track whether QA passed on first try

// ── Confidence scoring (for HITL adaptive review depth) ──
function computeConfidence(iterations, passCount, totalReviews, memoryMatchCount, qaFirstPass) {
  let score = 0
  score += (1 - (iterations - 1) / MAX_ITERATIONS) * 0.30    // fewer iterations = higher confidence
  score += (totalReviews > 0 ? passCount / totalReviews : 1) * 0.30  // review consensus
  score += Math.min(memoryMatchCount / 2, 1) * 0.20           // memory coverage (max at 2+ matches)
  score += (qaFirstPass ? 1 : 0) * 0.20                       // QA first-pass bonus
  return Math.round(score * 100) / 100
}


// ============================================================
// Pipeline orchestration loop
// ============================================================

// ============================================================
// Pipeline loop
// ============================================================

let iteration = 1
let fixContext = ''

// ── 预加载 L3 知识记忆（确定性注入，不依赖 Agent 自觉）──
// 从任务描述提取关键词，预构建 knowledge-load.sh 调用命令
// Generator prompt 将此作为必须执行的第一步
const taskKeywords = (args.task || '').match(/[一-鿿]{2,4}|[a-zA-Z_]{2,}/g) || []
const knowledgeCmd = taskKeywords.length > 0
  ? `bash .harness/scripts/knowledge-load.sh --service ${bareName} --keywords "${taskKeywords.join(',').substring(0, 200)}" --top 5`
  : `bash .harness/scripts/knowledge-load.sh --service ${bareName} --top 5`
log(`知识预加载: ${knowledgeCmd}`)

while (iteration <= MAX_ITERATIONS) {
  log(`第 ${iteration} 轮`)

  // Phase 1: Generator — 直接在主工作树实现并提交到当前分支。
  // FIX: 之前用 isolation:'worktree'，实现+测试滞留隔离 worktree、无合并回主树的步骤，
  // 导致 QA 在主树看不到任何改动、TDD 检查只能退回 7 天窗口审旧代码。
  // 并行管线若目标服务不同，主树直接编辑天然互不冲突。
  phase('Develop')
  await agent(generatorPrompt(iteration, fixContext, TASK_TYPE, knowledgeCmd), { label: `${args.serviceName}: 开发/修复` })
  log(`Generator 完成 (轮次 ${iteration})`)

  // Phase 2: QA
  phase('QA')
  const qaResult = await agent(qaPrompt(), { label: `QA: ${args.serviceName}`, schema: QA_SCHEMA })
  if (!qaResult) {
    log('QA Agent 被跳过，终止管线')
    return { status: 'aborted', reason: 'QA agent skipped' }
  }

  log(`QA VERDICT: ${qaResult.verdict} — ${qaResult.summary}`)

  // ── QA 门禁（gate-engine，对应 config/quality-gates.yml qa 段）──
  // 校验 QA 判定结构合法性；畸形判定（agent 输出垃圾 JSON 造成假 PASS）按 FAIL 处理。
  let qaGateFailures = []
  try {
    const gateEngine = require('./gate-engine.js')
    const qaGate = gateEngine.validateGate('qa', { qaResult, summary: `${args.serviceName} QA 门禁` })
    qaGateFailures = qaGate.failures
    if (qaGate.warnings.length > 0) {
      log(`⚠️ QA 门禁 WARN: ${qaGate.warnings.map(w => w.message).join('; ')}`)
    }
  } catch (e) {
    log(`⚠️ gate-engine 不可用（QA 门禁降级，不阻断）: ${e.message}`)
  }
  if (qaGateFailures.length > 0) {
    log(`⛔ QA 门禁阻断: ${qaGateFailures.map(f => f.message).join('; ')} — 按 QA FAIL 处理`)
  }

  if (qaResult.verdict === 'FAIL' || qaGateFailures.length > 0) {
    qaFirstPass = false

    if (TASK_TYPE === 'feature' || TASK_TYPE === 'bug') {
      // Phase 2.5: Debug — 完整根因分析
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
    } else {
      // debt/chore: skip Debug, build simple fixContext from QA results
      log(`Debug 跳过（${TASK_TYPE} 任务 — 已知修复模式）`)
      fixContext = `### QA 失败（${TASK_TYPE} 任务）\n${qaResult.summary}\n\n失败详情：\n${JSON.stringify(qaResult.failures, null, 2)}`
    }
    iteration++
    continue
  }

  // Phase 3: Multi-Perspective Review (并行，仅 QA PASS 后)
  if (REVIEW_LENS_COUNT === 0) {
    // chore: skip review entirely
    log(`⏭️ Review 跳过（${TASK_TYPE} 任务）`)
    const confidence = computeConfidence(iteration, 3, 3, 0, qaFirstPass)
    log(`✅ Harness 管线完成！${args.serviceName} (${iteration} 轮, confidence: ${confidence})`)
    logMetrics({ timestamp: args.timestamp || 'na', service: args.serviceName, taskType: TASK_TYPE, iterations: iteration, status: 'pass', reviewSkipped: true, confidence })
    return {
      status: 'pass',
      iterations: iteration,
      serviceName: args.serviceName,
      qaSummary: qaResult.summary,
      reviewSummary: 'skipped (chore)',
      memorySuggestions: [],
      confidence,
      notifications: [{ event: 'pipeline_pass', service: args.serviceName, detail: `${iteration} 轮通过 (${TASK_TYPE}), QA: ${qaResult.summary}` }],
    }
  }

  phase('Review')

  // Select lenses based on task type
  const activeLenses = TASK_TYPE === 'debt'
    ? [REVIEW_LENSES[1]]  // standards-eng only
    : REVIEW_LENSES        // all 3 for feature/bug

  const reviewResults = await parallel(
    activeLenses.map(lens => () =>
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

  log(`多视角审查完成: ${passCount}/${validReviews.length} PASS (threshold: ${REVIEW_PASS_THRESHOLD}/${validReviews.length})`)
  for (const r of validReviews) {
    log(`  ${r.lens} (${r.label}): ${r.verdict} — ${r.summary}`)
  }

  // 投票判定 — type-aware threshold
  const failingLenses = validReviews.filter(r => r.verdict === 'FAIL').map(r => r.label).join('、')
  if (passCount >= REVIEW_PASS_THRESHOLD) {
    if (failCount > 0) {
      log(`⚠️ ${failingLenses}视角有异议，但多数通过 (${passCount}/${validReviews.length})，管线继续`)
    }
    log(`✅ 多视角 Review PASS (${passCount}/${validReviews.length})`)

    // ── Review 门禁记录（gate-engine，对应 quality-gates.yml review 段）──
    try {
      const gateEngine = require('./gate-engine.js')
      const reviewGate = gateEngine.validateGate('review', {
        passCount, totalReviews: validReviews.length, summary: `${args.serviceName} Review 门禁`,
      })
      log(`🛡️ Review 门禁: ${reviewGate.passed ? 'PASS' : `FAIL(${reviewGate.failures.length})`} (${passCount}/${validReviews.length} APPROVED)`)
    } catch (e) {
      log(`⚠️ gate-engine 不可用（Review 门禁记录跳过）: ${e.message}`)
    }

    // 汇总所有 WARNING/NOTE 供参考
    const allWarnings = validReviews.reduce((sum, r) => sum + (r.warningCount || 0), 0)
    const reviewSummary = validReviews.map(r => `${r.label}: ${r.verdict}`).join(', ')

    // 收集并去重 memory suggestions（Review → Memory 反馈闭环）
    const allSuggestions = validReviews.flatMap(r => r.memorySuggestions || [])
    const seenSlugs = new Set()
    const uniqueSuggestions = allSuggestions.filter(s => {
      if (!s || !s.slug || seenSlugs.has(s.slug)) return false
      seenSlugs.add(s.slug)
      return true
    })
    if (uniqueSuggestions.length > 0) {
      log(`💡 ${uniqueSuggestions.length} 条 Memory 建议: ${uniqueSuggestions.map(s => s.slug).join(', ')}`)
    }

    // PASS — 管线完成！
    const memoryMatchCount = validReviews.reduce((sum, r) => sum + (r.memorySuggestions || []).length, 0)
    const confidence = computeConfidence(iteration, passCount, validReviews.length, memoryMatchCount, qaFirstPass)
    log(`✅ Harness 管线完成！${args.serviceName} 通过全部检查 (${iteration} 轮, confidence: ${confidence})`)
    const passNotifications = [{ event: 'pipeline_pass', service: args.serviceName, detail: `${iteration} 轮通过, QA: ${qaResult.summary}, confidence: ${confidence}` }]
    if (failCount > 0) {
      passNotifications.push({ event: 'need_human', service: args.serviceName, detail: `多数 PASS (${passCount}/${validReviews.length}) 但 ${failingLenses} 有异议` })
    }
    return {
      status: 'pass',
      iterations: iteration,
      serviceName: args.serviceName,
      qaSummary: qaResult.summary,
      reviewSummary: `${reviewSummary} (${passCount}/${validReviews.length} PASS, ${allWarnings} WARNINGs)`,
      memorySuggestions: uniqueSuggestions,
      confidence,
      notifications: passNotifications,
    }
  } else {
    // below threshold → 管线失败
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
log(`❌ 超过最大轮次 (${MAX_ITERATIONS}/${TASK_TYPE})，管线终止`)
return {
  status: 'max_iterations_exceeded',
  iterations: iteration,
  serviceName: args.serviceName,
  taskType: TASK_TYPE,
  confidence: Math.round((1 - 0.9) * 100) / 100,  // minimum confidence
  lastFixContext: fixContext,
  notifications: [{ event: 'pipeline_fail', service: args.serviceName, detail: `超过最大轮次 (${MAX_ITERATIONS}/${TASK_TYPE}), 最后错误: ${fixContext.substring(0, 200)}` }],
}
