#!/usr/bin/env bash
#
# harness-pipeline-loader.js — 动态加载器（运行时加载 Prompts）
#
# 功能：
#   - 动态加载 Prompt 模块
#   - 注入共享上下文
#   - 支持热更新（开发模式）
#

const fs = require('fs')
const path = require('path')

// Prompt 模块路径
const PROMPTS_DIR = path.join(__dirname, 'prompts')

// 共享上下文
let sharedContext = {}

// 加载的模块缓存
const moduleCache = {}

/**
 * 设置共享上下文
 */
function setSharedContext(ctx) {
  sharedContext = ctx

  // 如果已加载模块，更新其上下文
  Object.values(moduleCache).forEach(mod => {
    if (mod.setContext) {
      mod.setContext(sharedContext)
    }
  })
}

/**
 * 动态加载 Prompt 模块
 */
function loadPrompt(name) {
  const modulePath = path.join(PROMPTS_DIR, `${name}.js`)

  // 检查文件是否存在
  if (!fs.existsSync(modulePath)) {
    throw new Error(`Prompt module not found: ${name} (${modulePath})`)
  }

  // 开发模式：每次重新加载（热更新）
  if (process.env.HARNESS_DEV === 'true') {
    delete require.cache[require.resolve(modulePath)]
  }

  // 如果已缓存且非开发模式，返回缓存
  if (moduleCache[name] && process.env.HARNESS_DEV !== 'true') {
    return moduleCache[name]
  }

  try {
    // 加载模块
    const mod = require(modulePath)

    // 注入共享上下文
    if (mod.setContext) {
      mod.setContext(sharedContext)
    }

    // 缓存
    moduleCache[name] = mod

    return mod
  } catch (err) {
    console.error(`Failed to load prompt module: ${name}`)
    console.error(err.stack)
    throw err
  }
}

/**
 * 加载所有 Prompt 函数
 */
function loadAllPrompts() {
  return {
    generatorPrompt: loadPrompt('generator').generatorPrompt,
    qaPrompt: loadPrompt('qa').qaPrompt,
    reviewLensPrompt: loadPrompt('review').reviewLensPrompt,
    debuggingPrompt: loadPrompt('debug').debuggingPrompt,

    // Schema
    QA_SCHEMA: loadPrompt('qa').QA_SCHEMA || require('./schemas/qa-schema.js'),
    REVIEW_SCHEMA: loadPrompt('review').REVIEW_SCHEMA || require('./schemas/review-schema.js')
  }
}

/**
 * 初始化加载器
 */
function initLoader(ctx) {
  setSharedContext(ctx)
  return loadAllPrompts()
}

module.exports = {
  initLoader,
  setSharedContext,
  loadPrompt,
  loadAllPrompts
}
