// P1 收敛闭环行为测试 — 从源文件提取纯函数，验证不变量
// 运行: node .harness/workflows/p1-convergence.test.js
const fs = require('fs')

const SRC = fs.readFileSync(__dirname + '/../../workflows/harness-spec-pipeline.js', 'utf8')

// 提取顶层函数定义（consumeDecision / specContentHash / mustFixKey）
function extractFn(name) {
  const re = new RegExp(`function ${name}\\([^)]*\\) \\{[\\s\\S]*?\\n\\}`, 'm')
  const m = SRC.match(re)
  if (!m) throw new Error(`未找到函数 ${name}`)
  return m[0]
}

// 用 stub 上下文 eval 提取的函数（注入模块级闭包依赖：changeDir / fs / path / log）
function loadFn(name, ctxStub) {
  const body = extractFn(name)
  // 组装可执行函数：注入 changeDir（模块闭包依赖）+ fs/path/log
  const wrapper = new Function('fs', 'path', 'ROOT', 'CHANGE', 'log',
    `const changeDir = () => \`\${ROOT}/.harness/changes/\${CHANGE}\`;\n` +
    `${body.replace(new RegExp(`function ${name}\\(`), 'return function ' + name + '(')}\nreturn ${name};`)
  return wrapper(ctxStub.fs, ctxStub.path, ctxStub.ROOT, ctxStub.CHANGE, ctxStub.log)
}

let pass = 0, fail = 0
function assert(name, cond, detail = '') {
  if (cond) { pass++; console.log(`  ✅ ${name}`) }
  else { fail++; console.log(`  ❌ ${name} ${detail}`) }
}

// ── 1. consumeDecision: 读后即删（P1.4）──
{
  console.log('\n[P1.4] consumeDecision')
  const fn = loadFn('consumeDecision', { fs: null, path: null, ROOT: '/x', CHANGE: 'c', log: () => {} })
  const ctx = { decisions: { stage2_escalate: { escalate: '放宽阈值' }, other: { v: 1 } } }
  const d = fn(ctx, 'stage2_escalate')
  assert('读到决策', d && d.escalate === '放宽阈值')
  assert('读后即删', !('stage2_escalate' in ctx.decisions))
  assert('不影响其他决策', ctx.decisions.other.v === 1)
  assert('未消费时返回 undefined', fn({ decisions: {} }, 'nope') === undefined)
}

// ── 2. specContentHash: 内容敏感 + 确定性（P1.3）──
{
  console.log('\n[P1.3] specContentHash')
  // 自包含夹具：构造 ROOT/.harness/changes/<CHANGE>/ 结构（changeDir 依赖此路径），不依赖真实变更目录
  const root = '/tmp/p1-hash-root'
  const changeDirPath = `${root}/.harness/changes/fixture`
  fs.rmSync(root, { recursive: true, force: true })
  fs.mkdirSync(changeDirPath + '/specs/cap1', { recursive: true })
  fs.writeFileSync(changeDirPath + '/request.md', 'req v1')
  fs.writeFileSync(changeDirPath + '/proposal.md', 'proposal v1')
  fs.writeFileSync(changeDirPath + '/.change.yaml', 'yaml v1')
  fs.writeFileSync(changeDirPath + '/specs/cap1/spec.md', 'spec v1')
  const fn = loadFn('specContentHash', { fs, path: require('path'), ROOT: root, CHANGE: 'fixture', log: () => {} })
  const h1 = fn()
  assert('返回非空哈希', typeof h1 === 'string' && h1.length > 0, h1)
  const h2 = fn()
  assert('确定性（同内容同哈希）', h1 === h2)
  // 修改一个 spec 文件再验证哈希变化
  const specFile = `${changeDirPath}/specs/cap1/spec.md`
  fs.appendFileSync(specFile, '\n<!-- P1-hash-test-touch -->\n')
  const h3 = fn()
  assert('内容变则哈希变', h1 !== h3)
  fs.writeFileSync(specFile, 'spec v1') // 还原
  const h4 = fn()
  assert('还原后哈希恢复', h1 === h4)
  fs.rmSync(root, { recursive: true, force: true })
}

// ── 3. mustFixKey: 确定性签名（P1.2）──
{
  console.log('\n[P1.2] mustFixKey')
  const fn = loadFn('mustFixKey', { fs: null, path: null, ROOT: '/x', CHANGE: 'c', log: () => {} })
  const a = fn('coverage', { section: 'REQ-PLAT-8', issue: '问题A问题A问题A问题A问题A问题A问题A问题A问题A问题A问题A问题A' })
  const b = fn('coverage', { section: 'REQ-PLAT-8', issue: '问题A问题A问题A问题A问题A问题A问题A问题A问题A问题A问题A问题A' })
  const c = fn('structure', { section: 'REQ-PLAT-8', issue: '问题A问题A问题A问题A问题A问题A问题A问题A问题A问题A问题A问题A' })
  assert('同 lens+section+issue 同签名', a === b)
  assert('不同 lens 不同签名', a !== c)
  assert('签名含 lens+section', a.includes('coverage') && a.includes('REQ-PLAT-8'))
}

// ── 4. 早停判据逻辑模拟（P1.2 收敛停滞）──
{
  console.log('\n[P1.2] 收敛早停判据')
  const mfk = loadFn('mustFixKey', { fs: null, path: null, ROOT: '/x', CHANGE: 'c', log: () => {} })
  // 模拟 stage2: round1 REVISION 出 mustFix A；round2 REVISION 出同样的 A（无新）→ 应判停滞
  const round1Keys = [mfk('coverage', { section: 'S1', issue: 'issue1' })]
  const round2Keys = [mfk('coverage', { section: 'S1', issue: 'issue1' })] // 重复
  const history1 = new Set([])
  const new1 = round1Keys.filter(k => !history1.has(k))
  assert('round1 全部为新', new1.length === 1)
  const history2 = new Set(round1Keys)
  const new2 = round2Keys.filter(k => !history2.has(k))
  assert('round2 无新 → 停滞触发（rounds>=2 且 newKeys==0）', new2.length === 0)
  // round2 出现新 mustFix B → 非停滞（有进展）
  const round2New = [mfk('structure', { section: 'S2', issue: 'issue2' })]
  const new3 = round2New.filter(k => !history2.has(k))
  assert('round2 有新 → 继续（非停滞）', new3.length === 1)
}

console.log(`\n==== P1 行为测试: ${pass} 通过 / ${fail} 失败 ====`)
process.exit(fail ? 1 : 0)
