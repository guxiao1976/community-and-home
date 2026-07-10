// ============================================================
// QA Prompt & Schema — Quality Assurance Agent
// Uses template: .harness/agents/prompts/templates/qa.md
// ============================================================

import { renderFile } from './template-renderer.js'

const QA_SCHEMA = {
  type: 'object',
  properties: {
    verdict: { type: 'string', enum: ['PASS', 'FAIL'] },
    summary: { type: 'string' },
    failures: {
      type: 'array',
      items: {
        type: 'object',
        properties: {
          step: { type: 'string' },
          error: { type: 'string' },
        },
      },
    },
  },
  required: ['verdict', 'summary'],
}

function qaPrompt(taskType) {
  taskType = taskType || 'feature'
  const strictTdd = taskType === 'feature' || taskType === 'bug'
  const isFrontend = (SVC_DIR || '').startsWith('web/')
  const buildCmd = isFrontend ? 'npm run build' : 'go build ./...'
  const vetCmd = isFrontend ? 'npm run type-check' : 'go vet ./...'
  const testCmd = isFrontend ? 'npm run test:unit' : 'go test ./... -count=1'
  const SVC_NAME = args.serviceName.replace(/服务$/, '-service')
  const checkScript = isFrontend
    ? `bash .harness/skills/qa/scripts/harness-checks-frontend.sh --service ${SVC_NAME} --json`
    : `bash .harness/skills/qa/scripts/harness-checks.sh --service ${SVC_NAME} --json`
  const checkCount = isFrontend ? '6' : '14'

  const context = {
    serviceDir: SVC_DIR,
    isFrontend: isFrontend,
    strictTdd: strictTdd,
    buildCmd: buildCmd,
    vetCmd: vetCmd,
    testCmd: testCmd,
    checkScript: checkScript,
    checkCount: checkCount,
  }

  return renderFile('qa', context)
}

// Export for build script
if (typeof module !== 'undefined' && module.exports) {
  module.exports = { QA_SCHEMA, qaPrompt }
}
