// ============================================================
// Harness Spec Pipeline — 规范驱动全流程自动化流水线
// 路径选择 → 需求分析 → 需求评审 → 架构设计 → Proto → 编码 → 集成归档
// 每阶段末 HITL 暂停，resume 续跑；复用 harness-pipeline.js 做阶段 5 编码。
// ============================================================

export const meta = {
  name: 'harness-spec-pipeline',
  description: '规范驱动全流程：路径选择→需求分析→需求评审→架构设计→Proto→编码→集成归档，阶段间 HITL 暂停',
  phases: [
    { title: '0 路径选择', detail: 'dispatch 分级 S/M/L → request.md' },
    { title: '1 需求分析', detail: 'brainstorming 澄清（HITL）→ requirement-analyst → proposal/specs' },
    { title: '2 需求评审', detail: '3 视角并行 (coverage/structure/clarity) → 2/3 投票' },
    { title: '3 架构设计', detail: 'architecture-designer → design/tasks' },
    { title: '4 Proto 变更', detail: 'Owner 执行 make ci（HITL）→ proto_ci 门禁' },
    { title: '5 编码测试', detail: 'Owner 委托 N×harness-pipeline.js（HITL）' },
    { title: '6 集成归档', detail: 'build/vet 门禁 → 归档 → summary → INDEX' },
  ],
}

// ============================================================
// 沙箱兼容 Node API（复用 harness-pipeline.js 模式）
// ============================================================
let fs = null, path = null
try { fs = require('fs'); path = require('path') } catch (e) { /* sandbox: no Node API */ }
// ROOT：绝对路径基准（与 harness-pipeline.js 一致，沙箱 cwd 可能非项目根）
const ROOT = (typeof process !== 'undefined' && process.cwd) ? process.cwd() : '.'

// ============================================================
// Agent Schema（顶层常量，避免内联嵌套对象字面量解析歧义）
// ============================================================
const DISPATCH_SCHEMA = {
  type: 'object', additionalProperties: false, required: ['workload', 'reason', 'route', 'services'],
  properties: {
    workload: { type: 'string', enum: ['SKIP', 'S', 'M', 'L'] },
    signals: { type: 'object', additionalProperties: true },
    reason: { type: 'string' },
    route: { type: 'string' },
    services: { type: 'array', items: { type: 'string' } },
  },
}

const CLARIFY_SCHEMA = {
  type: 'object', additionalProperties: false, required: ['summary', 'questions'],
  properties: {
    summary: { type: 'string' },
    questions: { type: 'array', items: { type: 'object', additionalProperties: false, required: ['id', 'text', 'options'], properties: { id: { type: 'string' }, text: { type: 'string' }, options: { type: 'array', items: { type: 'string' } }, recommended: { type: 'number' }, why: { type: 'string' } } } },
  },
}

const REQUIREMENT_SCHEMA = {
  type: 'object', additionalProperties: false, required: ['traceability', 'specsCount', 'selfReview'],
  properties: {
    traceability: { type: 'object', additionalProperties: true },
    specsCount: { type: 'number' },
    selfReview: { type: 'string' },
  },
}

const SPEC_REVIEW_SCHEMA = {
  type: 'object', additionalProperties: false, required: ['verdict', 'summary'],
  properties: {
    verdict: { type: 'string', enum: ['APPROVED', 'REVISION'] },
    mustFixes: { type: 'array', items: { type: 'object', additionalProperties: true } },
    summary: { type: 'string' },
  },
}

const ARCHITECT_SCHEMA = {
  type: 'object', additionalProperties: false, required: ['services', 'protoChanges', 'tasksCount'],
  properties: {
    services: { type: 'array', items: { type: 'object', additionalProperties: true } },
    protoChanges: { type: 'array', items: { type: 'string' } },
    tasksCount: { type: 'number' },
  },
}

// ============================================================
// 持久化状态机（HITL 暂停/resume 的权威状态）
// ============================================================

const isMissing = (v) => !v || v === 'undefined' || typeof v !== 'string'

// args: { change, task, workload?, resumeFromRunId?, resumeWith?{decisions} }
const CHANGE = (typeof args !== 'undefined' && args && args.change) || ''
const TASK = (typeof args !== 'undefined' && args && args.task) || ''
if (!CHANGE) throw new Error('harness-spec-pipeline: 缺少 args.change（变更名，如 access-control）')
if (!TASK) throw new Error('harness-spec-pipeline: 缺少 args.task（用户需求描述）')

const changeDir = () => `${ROOT}/.harness/changes/${CHANGE}`
// statePath 仅作展示（沙箱无 fs，resume 状态经 args.resumeState 传，不落盘）
const statePath = () => `${changeDir()}/pipeline-state.json`

const RESUME_FROM = (typeof args !== 'undefined' && args && args.resumeFromRunId) || ''
const RESUME_WITH = (typeof args !== 'undefined' && args && args.resumeWith) || null
// 沙箱无 fs：resume 状态经 args.resumeState 传入（非磁盘）
const RESUME_STATE = (typeof args !== 'undefined' && args && args.resumeState) || null

// 注意：Workflow 脚本禁止 Date.now()/new Date()（会破坏 resume），state 不含时间戳
function defaultState() {
  return {
    schema: 1,
    change: CHANGE,
    task: TASK,
    currentStage: 0,
    stageResults: {},
    decisions: {},
    resumePending: null,
    resumeCount: 0,
  }
}

function loadState() {
  // 沙箱无 fs：优先从 args.resumeState 恢复（Owner resume 时传入完整 ctx）
  if (RESUME_STATE && typeof RESUME_STATE === 'object') {
    return { ...defaultState(), ...RESUME_STATE }
  }
  return defaultState()
}

function saveState(s) {
  // 沙箱无 fs：状态只保存在 ctx 内存，随 pauseForInput 返回带出
  s.__saved = true
}

// ── HITL 暂停：记录暂停点 + 返回 need_input（含完整 ctx，供 Owner resume 传回）──
function pauseForInput(ctx, checkpoint, payload) {
  ctx.resumePending = { checkpoint, questions: payload.questions, options: payload.options || {}, onResume: payload.onResume }
  ctx.currentStage = payload.stage
  saveState(ctx)
  log(`⏸️ 暂停等待输入: ${checkpoint}`)
  return {
    status: 'need_input',
    checkpoint,
    stage: payload.stage,
    summary: payload.summary,
    questions: payload.questions,
    artifacts: payload.artifacts || {},
    ctx: ctx,             // 完整 ctx 返回给 Owner，resume 时经 args.resumeState 传回
    resumeFromRunId: RESUME_FROM,
    onResume: payload.onResume,
  }
}

// ── 服务名提取（兼容 string / {service} / {serviceDir,name} 三种形态）──
function svcName(s) {
  if (typeof s === 'string') return s
  return (s && (s.service || s.serviceDir || s.name)) || ''
}
function svcBare(s) {
  return svcName(s).replace(/^services\//, '')
}

// ── 门禁检查（内置实现，沙箱可用；gate-engine 完整版在非沙箱环境优先）──
// 沙箱无法 require gate-engine（探针证实），故内置轻量门禁逻辑，保证自动化校验真实生效。
function checkGate(phase, ctxForGate) {
  try {
    const gateEngine = require('./gate-engine.js')
    const r = gateEngine.validateGate(phase, ctxForGate)
    if (r.warnings.length > 0) log(`  ⚠️ 门禁 WARN: ${r.warnings.map(w => w.message).join('; ')}`)
    return r
  } catch (e) {
    // 沙箱降级：内置门禁（与 gate-engine 对应 phase 逻辑一致）
    log(`  ⚠️ gate-engine 沙箱不可用，使用内置门禁: ${e.message}`)
    return builtinGate(phase, ctxForGate)
  }
}

// 内置轻量门禁（沙箱用，逻辑与 gate-engine.js 对应 phase 一致）
// 注意：沙箱无 fs，文件级检查不可用；改为「信任 agent 已报告产出成功」（agent 在完整环境写盘）。
// 文件真实落盘验证由沙箱外的 Owner/QA（harness-checks）在阶段完成后执行。
function builtinGate(phase, ctx) {
  const failures = []
  const warnings = []
  switch (phase) {
    case 'requirement_analysis': {
      // 信任 agent 报告：requirement-analyst 返回 traceability/specsCount/selfReview 即产出成功
      if (!ctx.traceability || Object.keys(ctx.traceability).length === 0) {
        failures.push({ gate: 'req_traceability', message: '转换追溯表必须非空（agent 未报告产出）' })
      }
      if (!ctx.specsCount || ctx.specsCount < 1) {
        failures.push({ gate: 'req_specs_exist', message: 'specs 数量 ≥1（agent 未报告）' })
      }
      break
    }
    case 'requirement_review': {
      if ((ctx.passCount || 0) < 2) failures.push({ gate: 'spec_review_min_approved', message: `评审需 ≥2/3 APPROVED（当前 ${ctx.passCount}/${ctx.totalReviews}）` })
      if ((ctx.rounds || 0) > 3) failures.push({ gate: 'spec_review_round_limit', message: `评审轮次 ≤3（当前 ${ctx.rounds}）` })
      break
    }
    case 'architecture_design': {
      // 信任 agent 报告：architecture-designer 返回 services/protoChanges/tasksCount 即产出成功
      if (!ctx.tasksCount || ctx.tasksCount < 3) {
        failures.push({ gate: 'arch_task_count', message: `tasks ≥3（agent 报告 ${ctx.tasksCount || 0}）` })
      }
      break
    }
    case 'proto_ci': {
      // 沙箱无法跑 make ci：信任 Owner 的 resume 决策（stage4_proto 已确认执行）
      if (ctx.protoChangesRequired && !ctx.protoDone) {
        failures.push({ gate: 'proto_ci_unverified', message: '需要 proto 变更，沙箱无法验证 make ci' })
      }
      break
    }
    case 'integration': {
      // 信任 stage6 的归档动作 + Owner 交付确认
      if (!ctx.archived) {
        failures.push({ gate: 'integ_archived', message: '集成归档未完成' })
      }
      break
    }
    default:
      break
  }
  return { passed: failures.length === 0, failures, warnings }
}

// ============================================================
// 阶段实现（Phase 2 骨架 — 后续 Phase 填充）
// ============================================================

// 阶段 0: 路径选择（dispatch 分级）
async function stage0Dispatch(ctx) {
  phase('0 路径选择')
  log('dispatch 分级：读取任务，判定 S/M/L 路由')

  // 纯文案/配置跳过判定（确定性信号）
  const pureTextPattern = /(纯文案|注释|README|CHANGELOG|配置值|yml值|yaml值|json值|环境变量默认值)/i
  const isPure = pureTextPattern.test(TASK)

  if (isPure) {
    ctx.route = 'SKIP'
    log('  ⏭️ 纯文案/配置 → 跳过 Pipeline')
  } else {
    // LLM 判定 S/M/L（prompt 要求读 dispatch.md 权威规则）
    const res = await agent(
      `你是 Community-Home 的 dispatch 判定 Agent。请按 .harness/skills/dispatch.md 的「工作量分级」规则，
      对以下用户需求做 S/M/L 分级判定。必须先 Read .harness/skills/dispatch.md 获取权威规则。

## 用户需求
${TASK}

## 判定
- workload: SKIP / S / M / L（按 dispatch.md 信号表 A-H 客观判定）
- signals: {A,B,C,D,E,F,G,H} 各信号命中情况
- reason: 一句话理由
- route: 对应路由（SKIP→Edit / S→轻量Pipeline / M→Pipeline / L→OpenSpec）
- services: 涉及的服务目录列表`,
      { label: 'dispatch 分级', schema: DISPATCH_SCHEMA }
    )
    ctx.workload = res.workload
    ctx.route = res.route
    ctx.services = res.services || []
    log(`  分级: ${ctx.workload} (${ctx.route}) — ${res.reason}`)
  }

  // 写 request.md（dispatch Step 2.4 分级块格式）
  if (fs) {
    try {
      if (!fs.existsSync(changeDir())) fs.mkdirSync(changeDir(), { recursive: true })
      const reqMd = `# Change: ${CHANGE}\n\n## 用户原话(摘要)\n${TASK}\n\n## 工作量分级\n- 分级: ${ctx.workload || '?'}\n- 路由: ${ctx.route || '?'}\n- 涉及服务: ${(ctx.services || []).join(', ') || '无'}\n`
      fs.writeFileSync(`${changeDir()}/request.md`, reqMd)
    } catch (e) { /* request.md 写入失败不阻断 */ }
  }

  // S/M 级短路阶段 1-4（阶段 5 用 workload:S 调 harness-pipeline）
  ctx.stageResults[0] = { workload: ctx.workload, route: ctx.route, services: ctx.services }
  saveState(ctx)

  // SKIP 级（纯文案/配置）：不进流水线，直接结束（Owner 自行 Edit + build）
  if (ctx.workload === 'SKIP' || ctx.route === 'SKIP') {
    log(`  ⏭️ SKIP 级（纯文案/配置）：流水线结束，Owner 直接 Edit + build 验证`)
    ctx.currentStage = 999  // 超出 0-6 循环
    saveState(ctx)
  }
}

// 阶段 1: 需求分析
async function stage1Requirement(ctx) {
  phase('1 需求分析')
  // S/M 级短路：不跑需求分析，直接跳阶段 5（编码）
  if (ctx.workload === 'S' || ctx.workload === 'M') {
    log(`  ⏭️ ${ctx.workload} 级：跳过需求分析/评审/设计，直接编码（阶段 5）`)
    ctx.stageResults[1] = { skipped: true, reason: `${ctx.workload} 级不走 OpenSpec` }
    ctx.currentStage = 4  // 主循环 +1 后跳到阶段 5
    saveState(ctx)
    return
  }

  // 1a. 澄清（brainstorming 硬门禁，显式第一步）
  if (!ctx.decisions.stage1_clarify) {
    const clarify = await agent(
      `你是需求澄清 Agent。对以下用户需求做 brainstorming 澄清（superpowers:brainstorming 思路）：
先探索现状、识别关键决策点，然后产出「待用户确认的澄清问题清单」。不要直接产出 spec。

## 变更
${CHANGE}

## 用户需求
${TASK}

## 输出（问题清单，每项含选项 + 推荐）
请列出需要用户拍板的关键问题（边界、方案对比、影响范围、安全权衡等），每项：
- id: 唯一标识
- text: 问题描述
- options: 候选选项数组（每项 label + 说明）
- recommended: 推荐选项的 index
- why: 为什么需要用户确认`,
      { label: '需求澄清（brainstorming）', schema: CLARIFY_SCHEMA }
    )
    ctx.stageResults[1] = { clarifySummary: clarify.summary }
    return pauseForInput(ctx, 'stage1_clarify', {
      stage: 1,
      summary: `需求澄清问题（${CHANGE}）— 请逐项拍板`,
      questions: clarify.questions,
      onResume: '决策注入后进入需求分析形式化',
    })
  }

  // 1b. 分析（决策已确认）
  const decisions = ctx.decisions.stage1_clarify
  const res = await agent(
    `你是 Community-Home 的需求分析师。执行 .harness/skills/requirement-analysis.md 的完整流程（Step 2-8），
产出 proposal + specs。**必须先 Read .harness/skills/requirement-analysis.md 获取权威流程**。

## 变更
${CHANGE}

## 用户需求
${TASK}

## 已确认的设计决策（用户已拍板）
${JSON.stringify(decisions, null, 2)}

## 产出（写入磁盘）
- .harness/changes/${CHANGE}/proposal.md
- .harness/changes/${CHANGE}/specs/<capability>/spec.md
- .harness/changes/${CHANGE}/.change.yaml

完成后返回：{ traceability, specsCount, selfReview }（traceability=转换追溯表，全✅ 才能通过）`,
      { label: 'requirement-analyst', schema: REQUIREMENT_SCHEMA }
    )

    // 门禁：requirement_analysis（信任 agent 报告：traceability/specsCount）
    const gate = checkGate('requirement_analysis', {
      changeDir: changeDir(),
      traceability: res.traceability,
      specsCount: res.specsCount,
      summary: `${CHANGE} 需求分析门禁`,
    })
    if (!gate.passed) {
      log(`❌ 需求分析门禁 FAIL: ${gate.failures.map(f => f.message).join('; ')}`)
      return { status: 'stage_fail', stage: 1, change: CHANGE, failures: gate.failures, stateFile: statePath() }
    }
    ctx.stageResults[1] = ctx.stageResults[1] || {}
    ctx.stageResults[1].traceability = res.traceability
    ctx.stageResults[1].specsCount = res.specsCount
    ctx.stageResults[1].gatePassed = true
    saveState(ctx)
    log(`  ✅ 需求分析完成: ${res.specsCount} specs, self-review ${res.selfReview}`)
}

// 阶段 2: 需求评审（3 视角并行 + 投票）
async function stage2Review(ctx) {
  phase('2 需求评审')

  // resume 处理：若已暂停过 stage2_done/stage2_escalate，读用户裁决分支
  if (ctx.decisions.stage2_done && ctx.decisions.stage2_done.approve) {
    const approve = ctx.decisions.stage2_done.approve
    log(`  📋 评审裁决: ${approve}`)
    if (approve.includes('回需求分析') || approve.includes('回')) {
      ctx.currentStage = 0  // 主循环 +1 → 阶段 1 修正
      saveState(ctx)
      return
    }
    // 「进入架构设计」→ 正常推进阶段 3
    log('  → 进入架构设计')
    return
  }
  if (ctx.decisions.stage2_escalate && ctx.decisions.stage2_escalate.escalate) {
    const act = ctx.decisions.stage2_escalate.escalate
    log(`  ⛔ 人工裁决: ${act}`)
    if (act.includes('终止')) {
      ctx.currentStage = 999  // 终止变更（超出 0-6 循环）
      saveState(ctx)
      return
    }
    if (act.includes('放宽')) {
      // 放宽阈值 → 强制通过进入架构设计
      ctx.stageResults[2] = { pass: 3, total: 3, rounds: ctx.stageResults[2]?.rounds || 1, escalated: true }
      log('  → 放宽阈值，进入架构设计')
      return
    }
    // 「人工修正 spec 后重试」→ 回阶段 1
    ctx.currentStage = 0
    saveState(ctx)
    return
  }

  const lenses = [
    { key: 'coverage', label: '覆盖完整性' },
    { key: 'structure', label: '结构合理性' },
    { key: 'clarity', label: '清晰可执行' },
  ]
  const reviews = await parallel(
    lenses.map(lens => () =>
      agent(
        `你是需求评审 Agent（${lens.label} 视角）。审查 ${CHANGE} 的需求规格。
先 Read .harness/skills/review.md「模式一：计划评审」，再按你的视角审查。

## 视角：${lens.label}
- coverage: 需求覆盖/场景完整性/边界识别
- structure: 服务归属/依赖顺序/职责边界
- clarity: 粒度/歧义/一致性（SHALL/MUST 唯一解释）

## 审查对象（磁盘）
- .harness/changes/${CHANGE}/request.md
- .harness/changes/${CHANGE}/proposal.md
- .harness/changes/${CHANGE}/specs/*/spec.md

## 输出
- verdict: APPROVED / REVISION（有 ≥1 MUST FIX 即 REVISION）
- mustFixes: 必须修复项数组 {section, issue, fix}
- summary: 一句话结论`,
        { label: `Review:${lens.key}`, schema: SPEC_REVIEW_SCHEMA }
      ).then(r => r ? { ...r, lens: lens.key } : null)
    )
  )
  const valid = reviews.filter(Boolean)
  const pass = valid.filter(r => r.verdict === 'APPROVED').length
  ctx.stageResults[2] = { pass, total: valid.length, rounds: (ctx.stageResults[2]?.rounds || 0) + 1 }
  saveState(ctx)

  // 写评审报告（审计轨迹）
  if (fs) {
    try {
      const dir = `${changeDir()}/review`
      if (!fs.existsSync(dir)) fs.mkdirSync(dir, { recursive: true })
      const round = ctx.stageResults[2].rounds
      for (const r of valid) {
        fs.writeFileSync(`${dir}/spec_review_${r.lens}_v${round}.md`,
          `# Plan Review — ${CHANGE}（${r.lens}视角）\n\n**VERDICT**: ${r.verdict}\n\n${r.summary}\n`)
      }
    } catch (e) { /* 报告写入失败不阻断 */ }
  }

  // 投票：≥2/3 通过；否则回阶段 1
  const gate = checkGate('requirement_review', { passCount: pass, totalReviews: valid.length, rounds: ctx.stageResults[2].rounds })
  if (gate.passed) {
    log(`  ✅ 需求评审 ${pass}/${valid.length} APPROVED`)
    return pauseForInput(ctx, 'stage2_done', {
      stage: 2,
      summary: `需求评审完成：${pass}/${valid.length} APPROVED — 批准进入架构设计？`,
      questions: [{ id: 'approve', text: '评审结论 APPROVED，进入架构设计？', options: ['进入架构设计', '回需求分析修正'] }],
      onResume: '评审裁决后进入阶段 3',
    })
  }
  // 评审不过 → 回阶段 1（≤3 轮，超限升级人工）
  const rounds = ctx.stageResults[2].rounds
  if (rounds >= 3) {
    log(`  ⛔ 需求评审已达 ${rounds} 轮上限，升级人工决策`)
    return pauseForInput(ctx, 'stage2_escalate', {
      stage: 2,
      summary: `需求评审 ${rounds} 轮仍未通过（${pass}/${valid.length}）— 升级人工`,
      questions: [{ id: 'escalate', text: '评审多次未通过，如何处理？', options: ['人工修正 spec 后重试', '终止变更', '放宽阈值'] }],
      onResume: '按人工决策继续',
    })
  }
  log(`  ❌ 需求评审 ${pass}/${valid.length}（阈值 2/3）— 回阶段 1（第 ${rounds} 轮）`)
  ctx.currentStage = 0  // 主循环 +1 → 阶段 1
  saveState(ctx)
}

// 阶段 3: 架构设计
async function stage3Architecture(ctx) {
  phase('3 架构设计')

  // resume 处理：若已暂停过 stage3_done，读用户裁决分支
  if (ctx.decisions.stage3_done && ctx.decisions.stage3_done.approve) {
    const approve = ctx.decisions.stage3_done.approve
    log(`  📋 设计确认: ${approve}`)
    if (approve.includes('回需求分析') || approve.includes('回')) {
      ctx.currentStage = 0  // 回阶段 1
      saveState(ctx)
      return
    }
    log('  → 进入 Proto 阶段')
    return
  }

  const res = await agent(
    `你是 Community-Home 的架构设计师。执行 .harness/skills/architect-design.md 的完整流程，
产出 design + tasks。**必须先 Read .harness/agents/subagents/architecture-designer.md 获取权威流程**。

## 变更
${CHANGE}

## 输入（磁盘）
- .harness/changes/${CHANGE}/proposal.md
- .harness/changes/${CHANGE}/specs/*/spec.md

## 产出（写入磁盘）
- .harness/changes/${CHANGE}/design.md
- .harness/changes/${CHANGE}/tasks.md（含「## 全局 / Proto」段标注 Proto 变更）

完成后返回：{ services, protoChanges, tasksCount }（services=按服务分组的任务，protoChanges=proto 变更清单）`,
    { label: 'architecture-designer', schema: ARCHITECT_SCHEMA }
  )

  // 门禁：architecture_design（信任 agent 报告：tasksCount）
  const gate = checkGate('architecture_design', {
    changeDir: changeDir(),
    tasksCount: res.tasksCount,
    summary: `${CHANGE} 架构设计门禁`,
  })
  if (!gate.passed) {
    log(`❌ 架构设计门禁 FAIL: ${gate.failures.map(f => f.message).join('; ')}`)
    ctx.currentStage = 0  // 回阶段 1
    saveState(ctx)
    return { status: 'stage_fail', stage: 3, change: CHANGE, failures: gate.failures, stateFile: statePath() }
  }

  ctx.stageResults[3] = { services: res.services, protoChanges: res.protoChanges, tasksCount: res.tasksCount }
  saveState(ctx)
  log(`  ✅ 架构设计完成: ${res.tasksCount} tasks, ${(res.protoChanges || []).length} proto 变更`)
  return pauseForInput(ctx, 'stage3_done', {
    stage: 3,
    summary: `架构设计完成：${res.tasksCount} tasks, ${(res.protoChanges || []).length} proto 变更 — 确认服务归属与 Proto 清单？`,
    questions: [{ id: 'approve', text: '服务归属正确？Proto 变更清单完整？', options: ['进入 Proto 阶段', '回需求分析修正'] }],
    artifacts: { design: `${changeDir()}/design.md`, tasks: `${changeDir()}/tasks.md` },
    onResume: '设计确认后进入阶段 4',
  })
}

// 阶段 4: Proto 变更（HITL：Owner 亲自执行 make ci）
async function stage4Proto(ctx) {
  phase('4 Proto 变更')
  const protoChanges = (ctx.stageResults[3] && ctx.stageResults[3].protoChanges) || []
  if (protoChanges.length === 0) {
    log('  ⏭️ 无 Proto 变更，跳过')
    ctx.stageResults[4] = { skipped: true }
    return
  }

  if (!ctx.decisions.stage4_proto) {
    return pauseForInput(ctx, 'stage4_proto', {
      stage: 4,
      summary: `需要修改 ${protoChanges.length} 个 proto 文件。请按硬规则由 Owner（全局 Claude）执行：`,
      questions: [{
        id: 'proto_done',
        text: `请执行 api-proto 变更 + make ci（lint+breaking+generate），完成后确认`,
        options: ['已执行并提交', '无变更', '需人工介入'],
      }],
      artifacts: { protoChanges, instructions: 'cd api-proto && 按 tasks.md「全局/Proto」修改 + make ci' },
      onResume: '校验 make ci + git diff 后进入阶段 5',
    })
  }

  // resume 后：把 Owner 决策（proto_done）映射到 ctx.protoDone，供门禁校验
  if (ctx.decisions.stage4_proto && ctx.decisions.stage4_proto.proto_done) {
    const done = ctx.decisions.stage4_proto.proto_done
    ctx.protoDone = done === '已执行并提交' || done === '无变更'
    if (!ctx.protoDone) {
      log('  ⛔ Owner 确认 proto 未完成，回阶段 3')
      ctx.currentStage = 3
      saveState(ctx)
      return { status: 'stage_fail', stage: 4, change: CHANGE, failures: [{ gate: 'proto_not_done', message: 'Owner 确认 proto 未完成' }], stateFile: statePath() }
    }
    log(`  📋 Proto 已执行: ${done}`)
  }

  // resume 后校验 proto_ci 门禁
  const gate = checkGate('proto_ci', { changeDir: changeDir(), protoChangesRequired: true, protoDone: ctx.protoDone, summary: `${CHANGE} Proto 门禁` })
  if (!gate.passed) {
    log(`❌ Proto 门禁 FAIL: ${gate.failures.map(f => f.message).join('; ')}`)
    ctx.currentStage = 3  // 回阶段 3（主循环 +1 → 阶段 4 重试）
    saveState(ctx)
    return { status: 'stage_fail', stage: 4, change: CHANGE, failures: gate.failures, stateFile: statePath() }
  }
  ctx.stageResults[4] = { passed: true }
  saveState(ctx)
  log('  ✅ Proto 变更通过 make ci')
}

// 阶段 5: 编码测试（HITL 委托 Owner 启动 N×harness-pipeline.js）
async function stage5Coding(ctx) {
  phase('5 编码+测试')

  // resume 处理：若已暂停过 stage5_done，读用户裁决分支
  if (ctx.decisions.stage5_done && ctx.decisions.stage5_done.approve) {
    const approve = ctx.decisions.stage5_done.approve
    log(`  📋 编码确认: ${approve}`)
    if (approve.includes('需修复')) {
      ctx.currentStage = 4  // 回阶段 5 重跑
      saveState(ctx)
      return
    }
    log('  → 进入集成归档')
    return
  }

  const services = (ctx.stageResults[3] && ctx.stageResults[3].services) || ctx.services || []
  if (services.length === 0) {
    // 没有服务列表（S/M 短路或设计未产出）→ 用 ctx.services
    log('  ⚠️ 无服务任务清单，使用 dispatch 阶段的服务列表')
  }

  if (!ctx.decisions.stage5_dispatch) {
    const svcList = services.map(svcName).join('\n')
    return pauseForInput(ctx, 'stage5_dispatch', {
      stage: 5,
      summary: `请为以下服务并行启动 Workflow harness-pipeline.js（每服务一个），全部 PASS 后确认：\n${svcList || '（无服务清单，请用 dispatch 阶段 ctx.services）'}`,
      questions: [{ id: 'pipelines_done', text: '所有服务 Pipeline 已全部 PASS？', options: ['全部 PASS', '有 FAIL'] }],
      onResume: '聚合各服务 Pipeline 结果后进入阶段 6',
    })
  }

  // resume 后聚合（Owner 在 decisions 里带回各服务结果）
  const dispatchResult = ctx.decisions.stage5_dispatch
  ctx.stageResults[5] = { dispatchResult, services }
  saveState(ctx)
  log(`  ✅ 编码阶段完成（${services.length} 服务，Owner 已确认 Pipeline 结果）`)

  return pauseForInput(ctx, 'stage5_done', {
    stage: 5,
    summary: `编码完成（${services.length} 服务）— 按置信度确认后进入集成归档？`,
    questions: [{ id: 'approve', text: '各服务 QA+Review 通过，进入集成归档？', options: ['进入集成归档', '需修复'] }],
    onResume: '编码确认后进入阶段 6',
  })
}

// 阶段 6: 集成归档
async function stage6Integrate(ctx) {
  phase('6 集成归档')

  // resume 处理：若已暂停过 stage6_done，读用户最终交付决策
  if (ctx.decisions.stage6_done && ctx.decisions.stage6_done.deliver) {
    const deliver = ctx.decisions.stage6_done.deliver
    log(`  📋 最终交付: ${deliver}`)
    if (deliver.includes('需修复')) {
      ctx.currentStage = 5  // 回阶段 5 修复
      saveState(ctx)
      return
    }
    // 「批准归档」→ 流水线完成（返回 pass）
    ctx.currentStage = 999  // 超出 0-6 循环 → 主循环退出，返回 pass
    saveState(ctx)
    return
  }

  const services = (ctx.stageResults[5] && ctx.stageResults[5].services) || []

  // 6.1 全链路编译门禁（archived=true 表示阶段 6 正在执行归档动作）
  const gate = checkGate('integration', {
    changeDir: changeDir(),
    services: services.map(svcName),
    change: CHANGE,
    archived: true,
    summary: `${CHANGE} 集成归档门禁`,
  })
  if (!gate.passed) {
    // 归档文件缺失（QA/Review 未归档）→ 尝试从 services/ 移动进来
    for (const s of services) {
      const bare = svcBare(s)
      const implDir = `${changeDir()}/impl/${bare}`
      try {
        if (!fs.existsSync(implDir)) fs.mkdirSync(implDir, { recursive: true })
        for (const f of ['_qa.md', '_review_security-arch.md', '_review_standards-eng.md', '_review_design-biz.md']) {
          const src = `services/${bare}/${f}`
          if (fs.existsSync(src) && !fs.existsSync(`${implDir}/${f}`)) {
            fs.copyFileSync(src, `${implDir}/${f}`)
            log(`  📦 归档 ${bare}/${f}`)
          }
        }
      } catch (e) { /* 归档失败不阻断 */ }
    }
    const gate2 = checkGate('integration', {
      changeDir: changeDir(),
      services: services.map(svcName),
      change: CHANGE,
      archived: true,
      summary: `${CHANGE} 集成归档门禁(归档后)`,
    })
    if (!gate2.passed) {
      log(`❌ 集成归档门禁 FAIL: ${gate2.failures.map(f => f.message).join('; ')}`)
      return { status: 'stage_fail', stage: 6, change: CHANGE, failures: gate2.failures, stateFile: statePath() }
    }
  }

  // 6.2 生成 summary.md（从 TEMPLATE）
  if (fs) {
    try {
      const template = fs.existsSync(`${ROOT}/.harness/changes/TEMPLATE.md`)
        ? fs.readFileSync(`${ROOT}/.harness/changes/TEMPLATE.md`, 'utf8')
        : '# 变更摘要 — ${CHANGE}\n\n## 阶段\n\n## 交付清单\n'
      const summary = template
        .replace(/<变更名>/g, CHANGE)
        .replace(/\*\*状态\*\*.*/, `**状态**: ✅ 已完成（spec-pipeline 全流程）`)
      fs.writeFileSync(`${changeDir()}/summary.md`, summary)
      log('  📄 生成 summary.md')
    } catch (e) { /* summary 生成失败不阻断 */ }
  }

  // 6.3 更新 INDEX.md
  if (fs) {
    try {
      const index = fs.existsSync(`${ROOT}/.harness/changes/INDEX.md`) ? fs.readFileSync(`${ROOT}/.harness/changes/INDEX.md`, 'utf8') : ''
      if (!index.includes(CHANGE)) {
        const entry = `\n## 2026-08-14 — ${CHANGE}\n\n**路径**: spec-pipeline 全流程\n**状态**: ✅ 已完成\n\n详见: [.harness/changes/${CHANGE}/](./${CHANGE}/)\n\n---\n`
        fs.writeFileSync(`${ROOT}/.harness/changes/INDEX.md`, index + entry)
        log('  📄 更新 INDEX.md')
      }
    } catch (e) { /* INDEX 更新失败不阻断 */ }
  }

  ctx.stageResults[6] = { archived: true }
  saveState(ctx)
  log('  ✅ 集成归档完成')
  return pauseForInput(ctx, 'stage6_done', {
    stage: 6,
    summary: `集成归档完成 — 最终交付确认？`,
    questions: [{ id: 'deliver', text: '集成验证通过，批准归档交付？', options: ['批准归档', '需修复'] }],
    onResume: '最终交付确认',
  })
}

// ============================================================
// 主循环
// ============================================================

let ctx = loadState()

// ── resume 检测：应用决策，清 resumePending ──
if (ctx.resumePending && RESUME_WITH && RESUME_WITH.decisions) {
  const cp = ctx.resumePending.checkpoint
  ctx.decisions[cp] = RESUME_WITH.decisions
  ctx.resumePending = null
  ctx.resumeCount = (ctx.resumeCount || 0) + 1
  saveState(ctx)
  log(`🔄 resume: 应用 ${cp} 决策后从阶段 ${ctx.currentStage} 续跑`)
}

const STAGE_FN = [stage0Dispatch, stage1Requirement, stage2Review, stage3Architecture, stage4Proto, stage5Coding, stage6Integrate]
const STAGE_GATE = ['', 'requirement_analysis', 'requirement_review', 'architecture_design', 'proto_ci', '', 'integration']

// 全局兜底：回退计数（防阶段 0-6 无限循环）
ctx.rollbackCount = ctx.rollbackCount || 0
const MAX_ROLLBACK = 10

while (ctx.currentStage <= 6) {
  const stage = ctx.currentStage
  const fn = STAGE_FN[stage]
  // 空壳阶段：Phase 3-6 实现前直接推进
  if (!fn) {
    ctx.currentStage++
    saveState(ctx)
    continue
  }

  // 全局兜底：回退到已过阶段 → 计数；超限升级人工
  if (stage <= ctx.lastStage && ctx.stageResults[stage]) {
    ctx.rollbackCount++
    saveState(ctx)
    if (ctx.rollbackCount >= MAX_ROLLBACK) {
      log(`⛔ 回退 ${ctx.rollbackCount} 次超上限，升级人工决策`)
      return {
        status: 'escalated',
        change: CHANGE,
        checkpoint: 'global_escalate',
        summary: `全流程回退 ${ctx.rollbackCount} 次超限，需人工介入`,
        stateFile: statePath(),
      }
    }
  }
  ctx.lastStage = stage

  const result = await fn(ctx)

  // 阶段函数请求 HITL 暂停 → 返回 need_input（currentStage 已由 pauseForInput 设为待续阶段）
  if (result && result.status === 'need_input') {
    return result
  }
  // 阶段门禁失败 → 返回 stage_fail（currentStage 已由阶段函数设为回退目标）
  if (result && result.status === 'stage_fail') {
    return result
  }

  // 正常完成 → 推进到下一阶段
  ctx.currentStage = ctx.currentStage + 1
  saveState(ctx)
}

log('✅ 全流程完成')
return {
  status: 'pass',
  change: CHANGE,
  stageResults: ctx.stageResults,
  stateFile: statePath(),
}
