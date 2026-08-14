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
  ctx.currentStage = 1
  saveState(ctx)
}

// 阶段 1: 需求分析（Phase 3 实现）
async function stage1Requirement(ctx) { /* Phase 3 */ }

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
  // 空壳阶段：Phase 3-6 实现前直接推进（避免骨架测试卡住）
  if (['', null, undefined].includes(String(STAGE_FN[stage]))) {
    ctx.currentStage++
    saveState(ctx)
    continue
  }
  await STAGE_FN[stage](ctx)
  ctx.currentStage++
  saveState(ctx)
}

log('✅ 全流程骨架完成（阶段函数待 Phase 3-6 实现）')
return {
  status: 'skeleton_done',
  change: CHANGE,
  stageResults: ctx.stageResults,
  stateFile: statePath(),
}
