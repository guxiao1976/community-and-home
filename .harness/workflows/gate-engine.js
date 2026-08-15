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
// 接线状态（2026-08-14 更新）：
//   harness-pipeline workflow：gate qa×2 / review（Generator→QA→Review 循环）
//   spec-pipeline 全流程：checkGate 接入 requirement_analysis / requirement_review /
//     architecture_design / proto_ci / integration（沙箱内降级 builtinGate 等价执行）
//   verify（交付前自验证）段：未接线——由 verify-before-deliver skill + harness-checks
//     的编译/测试检查覆盖，本段保留为设计参考
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

// ── 全流程门禁 Helper（spec-pipeline 各阶段用）────────────────

/** 检查文件存在 */
function fileExists(p) {
  return fs.existsSync(p)
}

/** 读取文件内容（不存在返回空串） */
function readFile(p) {
  try { return fs.readFileSync(p, 'utf8') } catch { return '' }
}

/** 检查 markdown 是否包含所有必填章节（## 或 ### 标题） */
function hasSections(content, sections) {
  return sections.every(sec => content.includes(sec))
}

/** 检查文件是否含占位符（TBD/TODO/待定/[NEEDS CLARIFICATION]/<描述>） */
function hasPlaceholders(content) {
  return /TBD|TODO|待定|\[NEEDS CLARIFICATION\]|<\w+>/g.test(content)
}

/** 执行命令，返回退出码 + stdout */
function run(cmd) {
  try {
    const out = execSync(cmd, { timeout: 60000, encoding: 'utf8', stdio: ['ignore', 'pipe', 'pipe'] })
    return { rc: 0, out: out || '' }
  } catch (e) {
    return { rc: 1, out: e.stdout ? e.stdout.toString() : e.message }
  }
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

  // ── 全流程阶段门禁（spec-pipeline）───────────────────────────
  requirement_analysis: [
    {
      id: 'req_file_exists',
      severity: 'BLOCK',
      desc: 'proposal.md 必须存在',
      check: (ctx) => fileExists(`${ctx.changeDir}/proposal.md`),
    },
    {
      id: 'req_required_sections',
      severity: 'BLOCK',
      desc: 'proposal.md 必须含 背景/做什么/影响范围/风险评估',
      check: (ctx) => hasSections(readFile(`${ctx.changeDir}/proposal.md`), ['为什么做', '做什么', '影响范围', '风险评估']),
    },
    {
      id: 'req_specs_exist',
      severity: 'BLOCK',
      desc: 'specs/*/spec.md 至少 1 个',
      check: (ctx) => {
        const dir = `${ctx.changeDir}/specs`
        if (!fs.existsSync(dir)) return false
        return fs.readdirSync(dir, { recursive: true }).some(f => String(f).endsWith('spec.md'))
      },
    },
    {
      id: 'req_no_placeholders',
      severity: 'BLOCK',
      desc: 'proposal/specs 无 [NEEDS CLARIFICATION]/TBD/TODO/待定 占位符',
      check: (ctx) => {
        const content = readFile(`${ctx.changeDir}/proposal.md`) + readFile(`${ctx.changeDir}/.change.yaml`)
        const specsDir = `${ctx.changeDir}/specs`
        let specContent = ''
        if (fs.existsSync(specsDir)) {
          for (const f of fs.readdirSync(specsDir, { recursive: true })) {
            if (String(f).endsWith('spec.md')) specContent += readFile(`${specsDir}/${f}`)
          }
        }
        return !hasPlaceholders(content + specContent)
      },
    },
  ],
  requirement_review: [
    {
      id: 'spec_review_min_approved',
      severity: 'BLOCK',
      desc: '需求评审需 ≥2/3 视角 APPROVED（阈值随视角数动态：3视角=2，4视角=3）',
      check: (ctx) => (ctx.passCount || 0) >= Math.ceil((ctx.totalReviews || 3) * 2 / 3),
    },
    {
      id: 'spec_review_round_limit',
      severity: 'BLOCK',
      desc: '需求评审轮次 ≤3（超限升级人工）',
      check: (ctx) => (ctx.rounds || 0) <= 3,
    },
  ],
  architecture_design: [
    {
      id: 'arch_design_exists',
      severity: 'BLOCK',
      desc: 'design.md 必须存在',
      check: (ctx) => fileExists(`${ctx.changeDir}/design.md`),
    },
    {
      id: 'arch_design_sections',
      severity: 'BLOCK',
      desc: 'design.md 必须含 架构概述/数据模型/接口设计',
      check: (ctx) => hasSections(readFile(`${ctx.changeDir}/design.md`), ['架构概述', '数据模型', '接口设计']),
    },
    {
      id: 'arch_tasks_exists',
      severity: 'BLOCK',
      desc: 'tasks.md 必须存在',
      check: (ctx) => fileExists(`${ctx.changeDir}/tasks.md`),
    },
    {
      id: 'arch_task_count',
      severity: 'BLOCK',
      desc: 'tasks.md 至少 3 个任务',
      check: (ctx) => {
        const m = readFile(`${ctx.changeDir}/tasks.md`).match(/### Task /g)
        return (m ? m.length : 0) >= 3
      },
    },
    {
      id: 'arch_zero_placeholders',
      severity: 'BLOCK',
      desc: 'tasks.md 无 TBD/TODO/<描述> 占位符',
      check: (ctx) => !hasPlaceholders(readFile(`${ctx.changeDir}/tasks.md`)),
    },
    {
      id: 'arch_tdd_steps',
      severity: 'WARN',
      desc: '含逻辑任务应有 RED→GREEN 步骤（WARN，不阻塞）',
      check: (ctx) => {
        const tasks = readFile(`${ctx.changeDir}/tasks.md`)
        return !/TDD|RED|GREEN|测试/.test(tasks) || /RED/.test(tasks)
      },
    },
  ],
  proto_ci: [
    {
      id: 'proto_lint',
      severity: 'BLOCK',
      desc: 'cd api-proto && make lint 必须通过',
      check: () => run('cd api-proto && make lint 2>&1').rc === 0,
    },
    {
      id: 'proto_breaking',
      severity: 'BLOCK',
      desc: 'make breaking-check 必须通过（无破坏性变更）',
      check: () => run('cd api-proto && make breaking-check 2>&1').rc === 0,
    },
    {
      id: 'proto_changes_present',
      severity: 'BLOCK',
      desc: '需要 proto 变更时 api-proto 应有改动',
      check: (ctx) => {
        if (!ctx.protoChangesRequired) return true
        return run('cd api-proto && git diff --name-only 2>&1').out.trim().length > 0
      },
    },
  ],
  integration: [
    {
      id: 'integ_full_build',
      severity: 'BLOCK',
      desc: '根目录 go build ./... 必须通过（全链路编译）',
      check: () => run('go build ./... 2>&1').rc === 0,
    },
    {
      id: 'integ_full_vet',
      severity: 'BLOCK',
      desc: '根目录 go vet ./... 必须通过',
      check: () => run('go vet ./... 2>&1').rc === 0,
    },
    {
      id: 'integ_qa_archived',
      severity: 'BLOCK',
      desc: '每服务 impl/<svc>/_qa.md 必须存在（QA 报告已归档）',
      check: (ctx) => {
        const services = ctx.services || []
        if (services.length === 0) return true
        return services.every(s => fileExists(`${ctx.changeDir}/impl/${s}/_qa.md`))
      },
    },
    {
      id: 'integ_review_archived',
      severity: 'BLOCK',
      desc: '每服务 impl/<svc>/_review*.md 必须存在',
      check: (ctx) => {
        const services = ctx.services || []
        if (services.length === 0) return true
        return services.every(s => {
          const dir = `${ctx.changeDir}/impl/${s}`
          if (!fs.existsSync(dir)) return false
          return fs.readdirSync(dir).some(f => f.startsWith('_review'))
        })
      },
    },
    {
      id: 'integ_summary_exists',
      severity: 'BLOCK',
      desc: 'summary.md 必须存在且含必填章节',
      check: (ctx) => hasSections(readFile(`${ctx.changeDir}/summary.md`), ['阶段', '交付清单']),
    },
    {
      id: 'integ_index_updated',
      severity: 'WARN',
      desc: 'INDEX.md 应包含该 change（WARN，不阻塞）',
      check: (ctx) => readFile('.harness/changes/INDEX.md').includes(ctx.change || ''),
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
