// P2 成本护栏 + P3 确定性门禁 行为测试
// 运行: node .harness/workflows/p2p3-guards.test.js
const fs = require('fs')

const SRC = fs.readFileSync(__dirname + '/harness-spec-pipeline.js', 'utf8')

function extractFn(name) {
  const re = new RegExp(`function ${name}\\([^)]*\\) \\{[\\s\\S]*?\\n\\}`, 'm')
  const m = SRC.match(re)
  if (!m) throw new Error(`未找到函数 ${name}`)
  return m[0]
}

// 加载函数：注入模块级依赖（budget / PIPELINE_BUDGET / MODEL_ROUTES / spentTokens / budgetLevel / changeDir 等）
function loadFn(name, deps) {
  const body = extractFn(name).replace(new RegExp(`function ${name}\\(`), 'return function ' + name + '(')
  const fn = new Function(
    'fs', 'path', 'ROOT', 'CHANGE', 'log', 'budget', 'PIPELINE_BUDGET', 'MODEL_ROUTES', 'spentTokens', 'budgetLevel', 'changeDir',
    body + '\nreturn ' + name + ';'
  )
  return fn(deps.fs, deps.path, deps.ROOT, deps.CHANGE, deps.log, deps.budget,
    deps.PIPELINE_BUDGET, deps.MODEL_ROUTES, deps.spentTokens, deps.budgetLevel, deps.changeDir)
}

let pass = 0, fail = 0
function assert(name, cond, detail = '') {
  if (cond) { pass++; console.log(`  ✅ ${name}`) }
  else { fail++; console.log(`  ❌ ${name} ${detail}`) }
}
const ROOT = fs.realpathSync(__dirname + '/../..')
const baseDeps = {
  fs, path: require('path'), ROOT, CHANGE: 'role-platforms-save', log: () => {},
  budget: null, PIPELINE_BUDGET: { soft: 1500000, hard: 2500000 }, MODEL_ROUTES: {},
  spentTokens: () => null, budgetLevel: null,
  changeDir: () => `${ROOT}/.harness/changes/role-platforms-save`,
}

// ── 1. budgetLevel: soft/hard 分级（P2.1）──
{
  console.log('\n[P2.1] budgetLevel 分级')
  // spentTokens 依赖 budget 全局；测试里直接注入固定返回值
  const mkSpent = (n) => loadFn('budgetLevel', { ...baseDeps, budget: { spent: () => n }, spentTokens: () => n })()
  const ok = mkSpent(100000)
  assert('未超预算 → ok', ok.level === 'ok' && ok.spent === 100000)
  const soft = mkSpent(1600000)
  assert('超 soft → soft 级', soft.level === 'soft')
  const hard = mkSpent(2600000)
  assert('超 hard → hard 级', hard.level === 'hard')
  const unknown = loadFn('budgetLevel', { ...baseDeps, budget: null, spentTokens: () => null })()
  assert('budget 不可用 → unknown', unknown.level === 'unknown' && unknown.spent === null)
}

// ── 2. costSummary 格式（P2.2）──
{
  console.log('\n[P2.2] costSummary 格式')
  const blFn = loadFn('budgetLevel', { ...baseDeps, budget: { spent: () => 1234567 }, spentTokens: () => 1234567 })
  const fn = loadFn('costSummary', { ...baseDeps, budget: { spent: () => 1234567 }, spentTokens: () => 1234567, budgetLevel: blFn })
  const s = fn()
  assert('含累计 token 与 soft/hard', typeof s === 'string' && s.includes('万输出token') && s.includes('soft') && s.includes('hard'), s)
  const empty = loadFn('costSummary', { ...baseDeps, budget: null, spentTokens: () => null, budgetLevel: loadFn('budgetLevel', { ...baseDeps, budget: null, spentTokens: () => null }) })
  assert('budget 不可用 → 空串', empty() === '')
}

// ── 3. routeModel: 模型路由（P2.1）──
{
  console.log('\n[P2.1] routeModel 模型路由')
  const inherited = loadFn('routeModel', { ...baseDeps, MODEL_ROUTES: {} })
  assert('未配置 → undefined（继承会话模型）', inherited('review') === undefined)
  const configured = loadFn('routeModel', { ...baseDeps, MODEL_ROUTES: { review: 'sonnet', analysis: 'opus' } })
  assert('配置 review → sonnet', configured('review') === 'sonnet')
  assert('未配置键 → undefined', configured('architecture') === undefined)
}

// ── 4. specDeterministicCheck: 真实 spec 应通过；构造缺陷应检出（P3.2）──
{
  console.log('\n[P3.2] specDeterministicCheck')
  const det = loadFn('specDeterministicCheck', { ...baseDeps })
  // 真实 role-platforms-save specs（已通过 3 轮评审）→ 应无发现
  const realCtx = {
    stageResults: { 1: { traceability: {
      design_d1: 'proposal D1 → REQ-UPDATE-4 → ✅',
      design_d2: 'proposal D2 → REQ-PLAT-6 → ✅',
      design_d3: 'proposal D3 → REQ-PLAT-2/3 → ✅',
      design_d4: 'proposal D4 → REQ-PLAT-4 → ✅',
      design_d5: 'proposal D5 → REQ-UPDATE-3 → ✅',
      design_d6: 'proposal D6 → REQ-PLAT-5 → ✅',
      design_d7: 'proposal D7 → REQ-LAYOUT-1 → ✅',
      bug1: '→ REQ-UPDATE-1/2 → ✅', bug1b: '→ REQ-PLAT-1..8 → ✅', bug1c: '→ REQ-UPDATE-4 → ✅', bug2: '→ REQ-LAYOUT-1 → ✅',
    } } },
    decisions: { stage1_clarify: { 'sys-role-edit-policy': 'A', 'http-read-path-gap': 'A', 'update-platforms-empty-semantics': 'A', 'platforms-validation': 'A', 'base-check-audit': 'A', 'sortorder-latent-bug': 'A', 'column-width-plan': 'A' } },
  }
  const realFindings = det(realCtx)
  assert('真实已评审 spec → 无确定性发现', realFindings.length === 0, JSON.stringify(realFindings))

  // 构造缺陷：traceability 缺一项 + 引用未定义 REQ
  const brokenCtx = {
    stageResults: { 1: { traceability: { design_d1: 'D1 → ✅' } } },
    decisions: { stage1_clarify: { 'sys-role-edit-policy': 'A', 'base-check-audit': 'A' } },
  }
  const brokenFindings = det(brokenCtx)
  assert('traceability 条目少 → 检出', brokenFindings.some(f => f.section === 'traceability'))

  // 错误码登记意图：往 spec 注入一个未声明登记的 06xxxx，应检出
  const ctx2 = JSON.parse(JSON.stringify(realCtx))
  const specFile = `${ROOT}/.harness/changes/role-platforms-save/specs/role-update-fix/spec.md`
  const orig = fs.readFileSync(specFile, 'utf8')
  fs.appendFileSync(specFile, '\n\n测试引用未知错误码 060099 的场景描述\n')
  const codeFindings = det(realCtx)
  fs.writeFileSync(specFile, orig)
  assert('spec 引入未登记错误码 → 检出 error-code 发现', codeFindings.some(f => f.section === 'error-code' && String(f.issue).includes('60099')), JSON.stringify(codeFindings.filter(f => f.section === 'error-code')))
}

console.log(`\n==== P2/P3 行为测试: ${pass} 通过 / ${fail} 失败 ====`)
process.exit(fail ? 1 : 0)
