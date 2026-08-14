// ============================================================
// Review Prompt & Schema — Code Review Agent
// Uses template: .harness/agents/prompts/templates/review.md
// ============================================================

import { renderFile } from './template-renderer.js'

const REVIEW_SCHEMA = {
  type: 'object',
  properties: {
    verdict: { type: 'string', enum: ['PASS', 'NEEDS_IMPROVEMENT', 'REJECT'] },
    lens: { type: 'string' },
    findings: {
      type: 'array',
      items: {
        type: 'object',
        properties: {
          severity: { type: 'string', enum: ['HIGH', 'MEDIUM', 'LOW'] },
          category: { type: 'string' },
          location: { type: 'string' },
          issue: { type: 'string' },
          suggestion: { type: 'string' },
        },
        required: ['severity', 'category', 'location', 'issue', 'suggestion'],
      },
    },
    memoryCompliance: {
      type: 'object',
      properties: {
        referenced: { type: 'array', items: { type: 'string' } },
        violated: { type: 'array', items: { type: 'string' } },
        missing: { type: 'array', items: { type: 'string' } },
      },
    },
  },
  required: ['verdict', 'lens', 'findings'],
}

const REVIEW_LENSES = [
  { name: 'Security & Architecture', slug: 'security' },
  { name: 'Standards & Engineering', slug: 'standards' },
  { name: 'Design & Business', slug: 'design' },
]

function reviewLensPrompt(lensName) {
  const context = {
    lensName: lensName,
    isSecurityLens: lensName === 'Security & Architecture',
    isStandardsLens: lensName === 'Standards & Engineering',
    isDesignLens: lensName === 'Design & Business',
  }

  return renderFile('review', context)
}

// Export for build script
if (typeof module !== 'undefined' && module.exports) {
  module.exports = { REVIEW_SCHEMA, REVIEW_LENSES, reviewLensPrompt }
}
