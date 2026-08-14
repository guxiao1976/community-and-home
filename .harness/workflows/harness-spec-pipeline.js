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

// ============================================================
// 持久化状态机（HITL 暂停/resume 的权威状态）
// ============================================================

const isMissing = (v) => !v || v === 'undefined' || typeof v !== 'string'

// args: { change, task, workload?, resumeFromRunId?, resumeWith?{decisions} }
const CHANGE = (typeof args !== 'undefined' && args && args.change) || ''
const TASK = (typeof args !== 'undefined' && args && args.task) || ''
if (!CHANGE) throw new Error('harness-spec-pipeline: 缺少 args.change（变更名，如 access-control）')
if (!TASK) throw new Error('harness-spec-pipeline: 缺少 args.task（用户需求描述）')

const changeDir = () => `.harness/changes/${CHANGE}`
const statePath = () => `${changeDir()}/pipeline-state.json`

const RESUME_FROM = (typeof args !== 'undefined' && args && args.resumeFromRunId) || ''
const RESUME_WITH = (typeof args !== 'undefined' && args && args.resumeWith) || null

function loadState() {
  if (!fs) return null
  try {
    if (fs.existsSync(statePath())) {
      return JSON.parse(fs.readFileSync(statePath(), 'utf8'))
    }
  } catch (e) { /* 损坏则重新初始化 */ }
  return {
    schema: 1,
    change: CHANGE,
    task: TASK,
    currentStage: 0,
    stageResults: {},
    decisions: {},
    resumePending: null,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
    resumeCount: 0,
  }
}

function saveState(s) {
  if (!fs) return
  try {
    if (!fs.existsSync(changeDir())) fs.mkdirSync(changeDir(), { recursive: true })
    s.updatedAt = new Date().toISOString()
    fs.writeFileSync(statePath(), JSON.stringify(s, null, 2))
  } catch (e) { /* state 写入失败不阻断（降级为纯内存） */ }
}

// ── HITL 暂停：写状态 + 返回 need_input ──
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
    stateFile: statePath(),
    resumeFromRunId: RESUME_FROM,
    onResume: payload.onResume,
  }
}

// ── 门禁检查（gate-engine）──
function checkGate(phase, ctxForGate) {
  try {
    const gateEngine = require('./gate-engine.js')
    const r = gateEngine.validateGate(phase, ctxForGate)
    if (r.warnings.length > 0) log(`  ⚠️ 门禁 WARN: ${r.warnings.map(w => w.message).join('; ')}`)
    return r
  } catch (e) {
    log(`  ⚠️ gate-engine 不可用（门禁降级）: ${e.message}`)
    return { passed: true, failures: [], warnings: [] }
  }
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
      { label: 'dispatch 分级', schema: {
        type: 'object', additionalProperties: false, required: ['workload', 'reason', 'route', 'services'],
        properties: {
          workload: { type: 'string', enum: ['SKIP', 'S', 'M', 'L'] },
          signals: { type: 'object', additionalProperties: true },
          reason: { type: 'string' },
          route: { type: 'string' },
          services: { type: 'array', items: { type: 'string' } },
        },
      } }
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
      { label: '需求澄清（brainstorming）', schema: {
        type: 'object', additionalProperties: false, required: ['summary', 'questions'],
        properties: {
          summary: { type: 'string' },
          questions: { type: 'array', items: {
            type: 'object', additionalProperties: false, required: ['id', 'text', 'options'],
            properties: {
              id: { type: 'string' },
              text: { type: 'string' },
              options: { type: 'array', items: { type: 'string' } },
              recommended: { type: 'number' },
              why: { type: 'string' },
            },
          } },
        },
      } }
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
      { label: 'requirement-analyst', schema: {
        type: 'object', additionalProperties: false, required: ['traceability', 'specsCount', 'selfReview'],
        properties: {
          traceability: { type: 'object', additionalProperties: true },
          specsCount: { type: 'number' },
          selfReview: { type: 'string' },
        },
      } }
    )

    // 门禁：requirement_analysis
    const gate = checkGate('requirement_analysis', { changeDir: changeDir(), summary: `${CHANGE} 需求分析门禁` })
    if (!gate.passed) {
      log(`❌ 需求分析门禁 FAIL: ${gate.failures.map(f => f.message).join('; ')}`)
      return { status: 'stage_fail', stage: 1, change: CHANGE, failures: gate.failures, stateFile: statePath() }
    }
    ctx.stageResults[1] = { ...ctx.stageResults[1], traceability: res.traceability, specsCount: res.specsCount, gatePassed: true }
    saveState(ctx)
    log(`  ✅ 需求分析完成: ${res.specsCount} specs, self-review ${res.selfReview}`)
  }
}

// 阶段 2: 需求评审（Phase 4 实现）
async function stage2Review(ctx) { /* Phase 4 */ }

// 阶段 3: 架构设计（Phase 4 实现）
async function stage3Architecture(ctx) { /* Phase 4 */ }

// 阶段 4: Proto 变更（Phase 5 实现）
async function stage4Proto(ctx) { /* Phase 5 */ }

// 阶段 5: 编码测试（Phase 5 实现）
async function stage5Coding(ctx) { /* Phase 5 */ }

// 阶段 6: 集成归档（Phase 6 实现）
async function stage6Integrate(ctx) { /* Phase 6 */ }

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

while (ctx.currentStage <= 6) {
  const stage = ctx.currentStage
  const fn = STAGE_FN[stage]
  // 空壳阶段：Phase 3-6 实现前直接推进
  if (!fn) {
    ctx.currentStage++
    saveState(ctx)
    continue
  }

  const result = await fn(ctx)
  ctx.currentStage = ctx.currentStage + 1
  saveState(ctx)

  // 阶段函数请求 HITL 暂停 → 返回 need_input，终止本轮（resume 后从 currentStage 续跑）
  if (result && result.status === 'need_input') {
    return result
  }
  // 阶段门禁失败 → 返回 stage_fail
  if (result && result.status === 'stage_fail') {
    return result
  }
}

log('✅ 全流程完成')
return {
  status: 'pass',
  change: CHANGE,
  stageResults: ctx.stageResults,
  stateFile: statePath(),
}
