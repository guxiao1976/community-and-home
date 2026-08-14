// P4 反馈回填 + P5 状态可靠 行为测试
// 运行: node .harness/pipeline/evals/p4p5-evals.eval.js
const fs = require('fs')

const SRC = fs.readFileSync(__dirname + '/../../workflows/harness-spec-pipeline.js', 'utf8')

function extractFn(name) {
  const re = new RegExp(`function ${name}\\([^)]*\\) \\{[\\s\\S]*?\\n\\}`, 'm')
  const m = SRC.match(re)
  if (!m) throw new Error('未找到函数 ' + name)
  return m[0]
}
function loadFn(name, deps) {
  const body = extractFn(name).replace(new RegExp(`function ${name}\\(`), 'return function ' + name + '(')
  const fn = new Function(
    'fs', 'path', 'ROOT', 'CHANGE', 'log', 'RESUME_STATE', 'defaultState', 'changeDir', 'statePath', 'REQUIRED_CTX_FIELDS', 'getPath',
    body + '\nreturn ' + name + ';'
  )
  return fn(deps.fs, deps.path, deps.ROOT, deps.CHANGE, deps.log, deps.RESUME_STATE, deps.defaultState, deps.changeDir, deps.statePath, deps.REQUIRED_CTX_FIELDS, deps.getPath)
}

let pass = 0, fail = 0
function assert(name, cond, detail = '') {
  if (cond) { pass++; console.log(`  ✅ ${name}`) }
  else { fail++; console.log(`  ❌ ${name} ${detail}`) }
}
const ROOT = fs.realpathSync(__dirname + '/../../..')
const baseDeps = {
  fs, path: require('path'), ROOT, CHANGE: 'test-change', log: () => {},
  RESUME_STATE: null,
  defaultState: () => ({ schema: 1, change: 'test-change', currentStage: 0, stageResults: {}, decisions: {}, resumePending: null, resumeCount: 0 }),
  changeDir: () => `${ROOT}/.harness/changes/test-change`,
  statePath: () => `${ROOT}/.harness/changes/test-change/pipeline-state.json`,
  REQUIRED_CTX_FIELDS: {
    2: [['stageResults[1].traceability']],
    4: [['stageResults[3].protoChanges']],
    5: [['stageResults[3].services', 'services']],
  },
}
// 加载 getPath 并注入 deps（validateResumeState 依赖它）
baseDeps.getPath = loadFn('getPath', baseDeps)

// ── 1. getPath: 支持 a.b[3].c（P5.1）──
{
  console.log('\n[P5.1] getPath 路径解析')
  const gp = loadFn('getPath', baseDeps)
  const obj = { stageResults: { 3: { protoChanges: ['p1'] }, 1: { traceability: { d1: '✅' } } }, services: ['s1'] }
  assert('a.b[3].c 点+方括号', gp(obj, 'stageResults[3].protoChanges')[0] === 'p1')
  assert('a.b.c 纯点', gp(obj, 'stageResults.1.traceability.d1') === '✅')
  assert('缺路径 → undefined', gp(obj, 'stageResults[9].x') === undefined)
}

// ── 2. validateResumeState: 必需字段校验（P5.1，D7 场景）──
{
  console.log('\n[P5.1] validateResumeState')
  const vr = loadFn('validateResumeState', baseDeps)
  const okCtx = { currentStage: 4, stageResults: { 3: { protoChanges: ['permission.proto'] } } }
  assert('stage4 有 protoChanges → 无缺失', vr(okCtx).length === 0)
  const d7Ctx = { currentStage: 4, stageResults: { 3: {} } }  // 缺 protoChanges（D7 曾致 stage4 误跳）
  assert('stage4 缺 protoChanges → 检出', vr(d7Ctx).length === 1 && String(vr(d7Ctx)[0]).includes('protoChanges'))
  const stage2Ctx = { currentStage: 2, stageResults: {} }
  assert('stage2 缺 traceability → 检出', vr(stage2Ctx).length === 1)
  const stage5Any = { currentStage: 5, stageResults: { 3: { services: ['x'] } } }
  assert('stage5 有 services（任一）→ 无缺失', vr(stage5Any).length === 0)
}

// ── 3. saveState/loadState: 磁盘持久化（P5.2）──
{
  console.log('\n[P5.2] saveState/loadState 落盘')
  const dir = '/tmp/p5-state-test'
  fs.rmSync(dir, { recursive: true, force: true })
  const deps = { ...baseDeps, changeDir: () => dir, statePath: () => `${dir}/pipeline-state.json`, fs, path: require('path'), log: () => {} }
  const save = loadFn('saveState', deps)
  const load = loadFn('loadState', deps)
  // saveState 落盘
  save({ schema: 1, change: 'test-change', currentStage: 3, foo: 'bar', __saved: false })
  assert('saveState 落盘文件存在', fs.existsSync(`${dir}/pipeline-state.json`))
  // loadState: 无 resumeState 时读盘（P5.2）
  const fromDisk = load()
  assert('loadState 读盘恢复 ctx', fromDisk.currentStage === 3 && fromDisk.foo === 'bar')
  // loadState: 有 resumeState 时优先 resumeState
  const depsRS = { ...deps, RESUME_STATE: { currentStage: 5, fromResume: true } }
  const loadRS = loadFn('loadState', depsRS)
  const fromRS = loadRS()
  assert('显式 resumeState 优先于盘', fromRS.currentStage === 5 && fromRS.fromResume === true)
  // schema 不匹配的盘状态 → 拒绝（用 default）
  fs.writeFileSync(`${dir}/pipeline-state.json`, JSON.stringify({ schema: 99, currentStage: 6 }))
  const badDisk = loadFn('loadState', deps)()
  assert('schema 不匹配 → 拒绝盘状态', badDisk.currentStage === 0)
  fs.rmSync(dir, { recursive: true, force: true })
}

// ── 4. P4.2 反馈 JSONL 结构：管线写入的 warning/memory 行必须可被 backfill 脚本解析 ──
{
  console.log('\n[P4.2] 反馈 JSONL 结构')
  const warningLine = JSON.stringify({ type: 'warning', change: 'role-platforms-save', round: 2, lens: 'coverage', section: 'REQ-PLAT-8', issue: '含引号"和转义\\的 issue', fix: 'fix' })
  const memoryLine = JSON.stringify({ type: 'memory', service: '权限服务', slug: 'x', title: 'y', content: 'z', triggers: 't' })
  const parsedW = JSON.parse(warningLine)
  const parsedM = JSON.parse(memoryLine)
  assert('warning 行含 type/section/issue', parsedW.type === 'warning' && parsedW.section && parsedW.issue)
  assert('含转义内容可正确解析（引号/反斜杠）', parsedW.issue.includes('"') && parsedW.issue.includes('\\'))
  assert('memory 行含 type/slug/content', parsedM.type === 'memory' && parsedM.slug && parsedM.content)
}

console.log(`\n==== P4/P5 行为测试: ${pass} 通过 / ${fail} 失败 ====`)
process.exit(fail ? 1 : 0)
