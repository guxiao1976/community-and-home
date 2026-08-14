// ============================================================
// Debug Prompt & Schema — Systematic Debugging Agent
// Uses template: .harness/agents/prompts/templates/debug.md
// ============================================================

import { renderFile } from './template-renderer.js'

const DEBUG_SCHEMA = {
  type: 'object',
  properties: {
    rootCauses: {
      type: 'array',
      items: {
        type: 'object',
        properties: {
          failure: { type: 'string' },
          category: {
            type: 'string',
            enum: ['编译错误', '运行时错误', '测试失败', '规范违反', 'TDD证据不足'],
          },
          whyChain: { type: 'array', items: { type: 'string' } },
          relatedMemory: { type: 'string' },
        },
        required: ['failure', 'category', 'whyChain'],
      },
    },
    fixPlan: {
      type: 'object',
      properties: {
        steps: { type: 'array', items: { type: 'string' } },
        avoidance: { type: 'string' },
      },
      required: ['steps', 'avoidance'],
    },
    newMemory: {
      type: 'object',
      properties: {
        slug: { type: 'string' },
        description: { type: 'string' },
        triggers: { type: 'array', items: { type: 'string' } },
      },
    },
  },
  required: ['rootCauses', 'fixPlan'],
}

function debuggingPrompt() {
  const context = {
    serviceDir: SVC_DIR,
    modulePrefix: (SVC_DIR || '').startsWith('web/')
      ? '@'
      : 'github.com/guxiao1976/community-user',
  }

  return renderFile('debug', context)
}

// Export for build script
if (typeof module !== 'undefined' && module.exports) {
  module.exports = { DEBUG_SCHEMA, debuggingPrompt }
}
