// ============================================================
// Generator Prompt — Development Agent (TDD + Memory-Driven)
// Uses template: .harness/agents/prompts/templates/generator.md
// ============================================================

import { renderFile } from './template-renderer.js'

function generatorPrompt(iteration, fixContext, taskType) {
  taskType = taskType || 'feature'
  const isChore = taskType === 'chore'
  const isDebt = taskType === 'debt'
  const strictTdd = !isChore && !isDebt  // full TDD only for feature/bug
  const isFrontend = (SVC_DIR || '').startsWith('web/')
  const langTool = isFrontend ? 'TypeScript' : 'Go'
  const buildCmd = isFrontend ? 'npm run build' : 'go build ./...'
  const testCmd = isFrontend ? 'npm run test:unit' : 'go test ./...'
  const vetCmd = isFrontend ? 'npm run type-check' : 'go vet ./...'

  const context = {
    serviceName: args.serviceName,
    serviceDir: SVC_DIR,
    taskType: taskType,
    isChore: isChore,
    isDebt: isDebt,
    strictTdd: strictTdd,
    isFrontend: isFrontend,
    langTool: langTool,
    buildCmd: buildCmd,
    testCmd: testCmd,
    vetCmd: vetCmd,
    iteration: iteration,
    iteration1: iteration === 1,
    hasFixContext: !!fixContext,
    fixContext: fixContext || '',
    modulePrefix: isFrontend ? '@' : 'github.com/guxiao1976/community-user',
  }

  return renderFile('generator', context)
}

// Export for build script
if (typeof module !== 'undefined' && module.exports) {
  module.exports = { generatorPrompt }
}
