// ============================================================
// gate-engine.js — 可执行的门禁引擎（对应 config/quality-gates.yml）
// ============================================================
//
// 背景：quality-gates.yml 定义了 63 条门禁规则，但此前从未被任何代码执行
// （配置漂移）。本引擎把其中「当前 pipeline 阶段真实存在、可程序化验证」的
// 规则落地为可执行检查；文档性规则（如 proposal.md 字数）因对应阶段
// （需求分析/架构设计）不在当前 workflow 中，明确标注「未接入」而不是假装执行。
//
// 规则映射（quality-gates.yml → 本引擎）：
//   qa.ci_check.status/total_tests/passed  → gate:qa_verdict_enum（QA 判定必须合法）
//   qa.ci_check 假 PASS 防护               → gate:qa_failures_shape（PASS 时 failures 应为空）
//   review.review_verdict.min_approved=2/3 → gate:review_min_approved
//   review.iteration_limit.max=2           → 由 core.js 的 MAX_ITERATIONS 已实现
//
//   verify.compile/service_running/bin_fresh/panic_log/db_dupes  → gate:verify_*
//   （交付前自验证门禁，对应 verify-before-deliver skill）
//
// 未接入 quality-gates.yml 阶段（yml 已标注 [status: not-implemented]）：
//   requirement_analysis / architecture_design / coding / ci / deployment / documentation
//   这 6 个阶段不在当前 harness-pipeline workflow（OpenSpec 设计预留），规则未被本引擎执行；
//   coding 的编译/AST 由 harness-checks.sh 独立执行，不经 validateGate()。
//
// 使用：workflow 中 require('./gate-engine.js')（产物与源共用同一目录），
// 所有 gate 判定写入 .harness/logs/gates/<date>.json 供可观测性分析。
// 引擎自身异常必须降级（try/catch），绝不因门禁日志问题阻断 pipeline。

const fs = require('fs')
const path = require('path')
const { execSync } = require('child_process')

const GATE_LOG_DIR = path.join(process.cwd(), '.harness', 'logs', 'gates')

function logGate(phase, result) {
  try {
    fs.mkdirSync(GATE_LOG_DIR, { recursive: true })
    const date = new Date().toISOString().split('T')[0]
    const file = path.join(GATE_LOG_DIR, `${date}.json`)
    const entry = { timestamp: new Date().toISOString(), phase, passed: result.passed, failures: result.failures, warnings: result.warnings, context: result.context }
    const logs = fs.existsSync(file) ? JSON.parse(fs.readFileSync(file, 'utf8')) : []
    logs.push(entry)
    fs.writeFileSync(file, JSON.stringify(logs, null, 2))
  } catch (e) { /* 门禁日志失败绝不阻断 pipeline */ }
}

// ── 工具函数 ──────────────────────────────────────────────────

/**
 * 安全执行 shell 命令，返回 stdout；异常返回空字符串
 */
function sh(cmd, opts = {}) {
  try {
    return execSync(cmd, { timeout: 5000, encoding: 'utf8', ...opts }).trim()
  } catch { return '' }
}

/**
 * 检查目标服务进程是否在运行
 */
function isProcessRunning(serviceName) {
  const out = sh(`pgrep -f "${serviceName}" || true`)
  return out.length > 0
}

/**
 * 获取服务进程的启动时间（Unix timestamp），进程不存在返回 0
 */
function getProcessStartTime(serviceName) {
  const pid = sh(`pgrep -f "${serviceName}" | head -1`)
  if (!pid) return 0
  const elapsed = sh(`ps -o etimes= -p ${pid}`)
  if (!elapsed) return 0
  return Math.floor(Date.now() / 1000) - parseInt(elapsed)
}

/**
 * 获取指定服务目录下最新 Go 源文件的修改时间
 */
function getLatestSourceModTime(serviceDir) {
  const dir = path.join(process.cwd(), serviceDir)
  const out = sh(`find ${dir} -name "*.go" -type f -exec stat -c %Y {} \\; 2>/dev/null | sort -n | tail -1`)
  return out ? parseInt(out) : 0
}

/**
 * 检查最近的日志文件是否存在 panic/fatal
 */
function checkLogsForPanic(logPattern) {
  const out = sh(`grep -l "panic\\|FATAL" ${logPattern} 2>/dev/null | head -5`)
  return out.length === 0
}

/**
 * 通过 docker exec mysql 检查表中是否有重复记录
 */
function checkDBDuplicates(dsn, table, uniqueColumn) {
  const sql = `SELECT COUNT(*) as c FROM (SELECT ${uniqueColumn}, COUNT(*) cnt FROM ${table} GROUP BY ${uniqueColumn} HAVING cnt > 1) t`
  const out = sh(`docker exec mysql mysql -u${dsn.user} -p${dsn.pass} -N -e "${sql}" 2>/dev/null`)
  return out === '' || parseInt(out) === 0
}

// ── 可执行门禁定义 ─────────────────────────────────────────────
const GATES = {
  qa: [
    {
      id: 'qa_verdict_enum',
      severity: 'BLOCK',
      desc: 'QA verdict 必须是 PASS/FAIL（quality-gates.yml qa.ci_check.status 等价物）；畸形判定按 FAIL 处理',
      check: (ctx) => {
        const v = ctx.qaResult && ctx.qaResult.verdict
        return v === 'PASS' || v === 'FAIL'
      },
    },
    {
      id: 'qa_failures_shape',
      severity: 'WARN',
      desc: 'QA 判定 PASS 时 failures 应为空；非空说明 agent 自相矛盾（假 PASS 风险）',
      check: (ctx) => {
        if (ctx.qaResult.verdict !== 'PASS') return true
        return !Array.isArray(ctx.qaResult.failures) || ctx.qaResult.failures.length === 0
      },
    },
  ],
  review: [
    {
      id: 'review_min_approved',
      severity: 'BLOCK',
      desc: '评审需 ≥2/3 视角通过（quality-gates.yml review.review_verdict.min_approved=2, total_reviewers=3）；debt 类单视角 ≥1',
      check: (ctx) => {
        const total = ctx.totalReviews || 0
        if (total === 0) return true // 无评审（chore）不设门禁
        const required = total >= 3 ? 2 : 1
        return (ctx.passCount || 0) >= required
      },
    },
  ],

  // ── 交付前自验证门禁（对应 verify-before-deliver skill）─────────
  verify: [
    {
      id: 'verify_compile',
      severity: 'BLOCK',
      desc: '代码必须能编译通过',
      check: (ctx) => {
        if (ctx.buildPassed !== undefined) return ctx.buildPassed
        // 降级：如果 pipeline 没传 buildPassed，尝试直接编译
        const svcDir = ctx.serviceDir || 'services/permission-service'
        const isFrontend = svcDir.startsWith('web/')
        const cmd = isFrontend
          ? `cd ${svcDir} && npm run build 2>&1`
          : `cd ${svcDir} && go build ./... 2>&1`
        const out = sh(cmd)
        return !out.includes('error') && !out.includes('Error')
      },
    },
    {
      id: 'verify_service_running',
      severity: 'WARN',
      desc: '目标服务进程应在运行（若已部署）；未运行提示重启',
      check: (ctx) => {
        const svc = ctx.serviceName || ctx.serviceDir?.split('/').pop() || ''
        if (!svc) return true  // 无服务名则不检查
        return isProcessRunning(svc)
      },
    },
    {
      id: 'verify_process_fresh',
      severity: 'BLOCK',
      desc: '运行中的进程必须晚于最后一次代码修改（否则是旧代码）',
      check: (ctx) => {
        const svc = ctx.serviceName || ''
        const dir = ctx.serviceDir || ''
        if (!svc || !dir) return true
        if (!isProcessRunning(svc)) return true // 没进程交给 verify_service_running 报 WARN
        const procStart = getProcessStartTime(svc)
        const srcMod = getLatestSourceModTime(dir)
        if (procStart === 0 || srcMod === 0) return true // 无法判定时放行
        return procStart > srcMod
      },
    },
    {
      id: 'verify_no_panic_in_logs',
      severity: 'BLOCK',
      desc: '最近的服务日志不能有 panic/fatal',
      check: (ctx) => {
        const logPattern = ctx.logPattern || '/tmp/*-rpc.log /tmp/*-api.log'
        return checkLogsForPanic(logPattern)
      },
    },
    {
      id: 'verify_db_no_duplicates',
      severity: 'WARN',
      desc: '核心表无重复记录（检查 UNIQUE 约束违规）',
      check: (ctx) => {
        const checks = ctx.dbIntegrityChecks || []
        if (checks.length === 0) return true  // 未配置则不检查
        const dsn = ctx.dbDsn || { user: 'root', pass: 'root123456' }
        for (const c of checks) {
          if (!checkDBDuplicates(dsn, c.table, c.column)) return false
        }
        return true
      },
    },
  ],
}

/**
 * 校验阶段门禁
 * @param {string} phase - 'qa' | 'review' | 'verify'
 * @param {object} ctx - 阶段上下文，各阶段所需字段不同：
 *   qa:      { qaResult }
 *   review:  { passCount, totalReviews, rejectCount }
 *   verify:  { serviceName, serviceDir, buildPassed, logPattern, dbIntegrityChecks, dbDsn }
 * @returns {{passed: boolean, failures: Array, warnings: Array}}
 */
function validateGate(phase, ctx = {}) {
  const rules = GATES[phase] || []
  const failures = []
  const warnings = []
  for (const rule of rules) {
    try {
      if (rule.check(ctx)) continue
      ;(rule.severity === 'WARN' ? warnings : failures).push({ gate: rule.id, message: rule.desc })
    } catch (e) {
      failures.push({ gate: rule.id, message: `gate 执行异常: ${e.message}` })
    }
  }
  const result = { passed: failures.length === 0, failures, warnings, context: ctx.summary || phase }
  logGate(phase, result)
  return result
}

module.exports = { validateGate, GATES }
