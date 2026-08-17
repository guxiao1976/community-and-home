#!/usr/bin/env node
// token-stats.mjs — 管线任务 token 用量聚合
//
// 用途：
//   ① 管线结束的精确 per-agent 用量（读子 agent transcript 的 usage 字段，含 input/output/cache）
//   ② 与管线内置的 budget.spent 阶段差分（近似，仅 output）互补，提供 token 优化依据
//
// 用法：
//   node .harness/scripts/token-stats.mjs                 # 最近一次 workflow run
//   node .harness/scripts/token-stats.mjs --run <dir>     # 指定 run 目录
//   node .harness/scripts/token-stats.mjs --all           # 全部 run 汇总
//
// 输出：per-agent 表（agent / input / output / cache_read / cache_create / total）+ 总用量

import { readdirSync, readFileSync, existsSync, statSync } from 'node:fs'
import { join } from 'node:path'
import { homedir } from 'node:os'

const RUNS_ROOT = join(homedir(), '.claude/projects')
const arg = process.argv.slice(2)
const runDir = arg.includes('--run') ? arg[arg.indexOf('--run') + 1] : null
const all = arg.includes('--all')

function findRuns() {
  const runs = []
  const walk = (dir, depth) => {
    if (depth > 4) return
    let entries = []
    try { entries = readdirSync(dir) } catch { return }
    for (const e of entries) {
      const full = join(dir, e)
      let isDir = false
      try { isDir = statSync(full).isDirectory() } catch { continue }
      if (isDir) {
        if (existsSync(join(full, 'journal.jsonl'))) runs.push(full)
        walk(full, depth + 1)
      }
    }
  }
  walk(RUNS_ROOT, 0)
  return runs
}

function sumAgentUsage(agentFile) {
  // 每行 JSON 可能含 usage{input_tokens,cache_creation_input_tokens,cache_read_input_tokens,output_tokens}
  const t = { input: 0, output: 0, cacheRead: 0, cacheCreate: 0, calls: 0 }
  const lines = readFileSync(agentFile, 'utf8').split('\n')
  for (const l of lines) {
    if (!l.trim()) continue
    try {
      const o = JSON.parse(l)
      const u = (o.message && o.message.usage) || o.usage
      if (u) {
        t.input += u.input_tokens || 0
        t.output += u.output_tokens || 0
        t.cacheRead += u.cache_read_input_tokens || 0
        t.cacheCreate += u.cache_creation_input_tokens || 0
        t.calls++
      }
    } catch { /* skip non-JSON lines */ }
  }
  return t
}

function analyzeRun(dir) {
  const agents = readdirSync(dir).filter(f => f.endsWith('.jsonl') && !f.startsWith('journal'))
  const rows = []
  const total = { input: 0, output: 0, cacheRead: 0, cacheCreate: 0 }
  for (const f of agents) {
    const t = sumAgentUsage(join(dir, f))
    if (t.calls === 0) continue
    rows.push({ agent: f.replace('.jsonl', '').slice(0, 12), ...t })
    total.input += t.input; total.output += t.output
    total.cacheRead += t.cacheRead; total.cacheCreate += t.cacheCreate
  }
  rows.sort((a, b) => (b.input + b.output) - (a.input + a.output))
  return { rows, total }
}

function fmt(n) { return (n / 1000).toFixed(1) + 'k' }

function print(dir) {
  const { rows, total } = analyzeRun(dir)
  console.log(`\n=== Token 用量 — ${dir.split('/').slice(-2).join('/')} ===`)
  console.log(`${'agent'.padEnd(14)} ${'input'.padEnd(9)} ${'output'.padEnd(9)} ${'cache_read'.padEnd(11)} ${'cache_create'.padEnd(12)} total`)
  for (const r of rows) {
    const t = r.input + r.output + r.cacheRead + r.cacheCreate
    console.log(`${r.agent.padEnd(14)} ${fmt(r.input).padEnd(9)} ${fmt(r.output).padEnd(9)} ${fmt(r.cacheRead).padEnd(11)} ${fmt(r.cacheCreate).padEnd(12)} ${fmt(t)}`)
  }
  const gt = total.input + total.output + total.cacheRead + total.cacheCreate
  console.log(`${'TOTAL'.padEnd(14)} ${fmt(total.input).padEnd(9)} ${fmt(total.output).padEnd(9)} ${fmt(total.cacheRead).padEnd(11)} ${fmt(total.cacheCreate).padEnd(12)} ${fmt(gt)}`)
  console.log(`  输入 ${fmt(total.input)} | 输出 ${fmt(total.output)} | 缓存读 ${fmt(total.cacheRead)} | 缓存写 ${fmt(total.cacheCreate)} | 合计 ${fmt(gt)}`)
}

if (runDir) {
  print(runDir)
} else if (all) {
  const runs = findRuns()
  for (const r of runs.sort().reverse()) print(r)
} else {
  const runs = findRuns()
  if (runs.length === 0) { console.log('未找到 workflow run'); process.exit(1) }
  print(runs[runs.length - 1])
}
