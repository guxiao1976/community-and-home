/**
 * AI Judgment Validator
 *
 * Validates AI judgment results against deterministic checks
 * to detect and resolve conflicts.
 */

const fs = require('fs');
const path = require('path');

/**
 * Conflict types
 */
const ConflictType = {
  AI_PASS_COMPILE_FAIL: 'ai-pass-compile-fail',
  AI_PASS_TESTS_FAIL: 'ai-pass-tests-fail',
  AI_PASS_DEPENDENCIES_FAIL: 'ai-pass-dependencies-fail',
  AI_FAIL_ALL_PASS: 'ai-fail-all-deterministic-pass',
  AI_FAIL_CRITICAL_PASS: 'ai-fail-critical-pass',
};

/**
 * Conflict severity
 */
const Severity = {
  CRITICAL: 'critical',
  HIGH: 'high',
  MEDIUM: 'medium',
  LOW: 'low',
};

/**
 * Recommended actions
 */
const Action = {
  OVERRIDE_TO_FAIL: 'override-to-fail',
  OVERRIDE_TO_PASS: 'override-to-pass',
  HUMAN_REVIEW: 'human-review',
  LOG_WARNING: 'log-warning',
};

/**
 * Validate AI judgment against deterministic results
 *
 * @param {Object} aiResult - AI judgment result { status: 'PASS'|'FAIL', reason: string }
 * @param {Object} deterministicResult - Deterministic check results
 * @returns {Object} Validation result with conflicts
 */
function validateAIJudgment(aiResult, deterministicResult) {
  const conflicts = [];

  // Extract deterministic check statuses
  const checks = deterministicResult.checks || [];
  const compile = checks.find(c => c.check === 'compile')?.status === 'PASS';
  const tests = checks.find(c => c.check === 'tests')?.status === 'PASS';
  const dependencies = checks.find(c => c.check === 'dependencies')?.status === 'PASS';

  // All critical checks pass
  const allCriticalPass = compile && tests && dependencies;

  // All checks pass (including warnings)
  const allPass = checks.every(c => c.status === 'PASS' || c.status === 'WARN' || c.status === 'SKIP');

  // Conflict 1: AI says PASS but compilation failed
  if (aiResult.status === 'PASS' && !compile) {
    conflicts.push({
      type: ConflictType.AI_PASS_COMPILE_FAIL,
      severity: Severity.CRITICAL,
      action: Action.OVERRIDE_TO_FAIL,
      message: 'AI判断为PASS，但编译失败（确定性错误）',
      detail: 'Compilation is deterministic and must pass',
      recommendation: '自动推翻AI判断，结果改为FAIL',
    });
  }

  // Conflict 2: AI says PASS but tests failed
  if (aiResult.status === 'PASS' && !tests) {
    conflicts.push({
      type: ConflictType.AI_PASS_TESTS_FAIL,
      severity: Severity.CRITICAL,
      action: Action.OVERRIDE_TO_FAIL,
      message: 'AI判断为PASS，但测试失败（确定性错误）',
      detail: 'Test failures are deterministic and must be fixed',
      recommendation: '自动推翻AI判断，结果改为FAIL',
    });
  }

  // Conflict 3: AI says PASS but dependencies failed
  if (aiResult.status === 'PASS' && !dependencies) {
    conflicts.push({
      type: ConflictType.AI_PASS_DEPENDENCIES_FAIL,
      severity: Severity.HIGH,
      action: Action.OVERRIDE_TO_FAIL,
      message: 'AI判断为PASS，但依赖验证失败',
      detail: 'Dependency verification is deterministic',
      recommendation: '自动推翻AI判断，结果改为FAIL',
    });
  }

  // Conflict 4: AI says FAIL but all deterministic checks pass
  if (aiResult.status === 'FAIL' && allPass) {
    conflicts.push({
      type: ConflictType.AI_FAIL_ALL_PASS,
      severity: Severity.MEDIUM,
      action: Action.HUMAN_REVIEW,
      message: 'AI判断为FAIL，但所有确定性检查都通过',
      detail: 'Possible false negative - AI may be too strict',
      recommendation: '建议人工审查AI的判断理由',
      ai_reason: aiResult.reason,
    });
  }

  // Conflict 5: AI says FAIL but all critical checks pass
  if (aiResult.status === 'FAIL' && allCriticalPass && !allPass) {
    conflicts.push({
      type: ConflictType.AI_FAIL_CRITICAL_PASS,
      severity: Severity.LOW,
      action: Action.LOG_WARNING,
      message: 'AI判断为FAIL，关键检查通过但有警告',
      detail: 'AI might be flagging non-critical issues (warnings)',
      recommendation: '记录但不推翻AI判断',
    });
  }

  // Determine final result
  let finalResult = aiResult.status;
  let overridden = false;

  const criticalConflicts = conflicts.filter(c =>
    c.action === Action.OVERRIDE_TO_FAIL || c.action === Action.OVERRIDE_TO_PASS
  );

  if (criticalConflicts.length > 0) {
    // Override AI judgment
    if (criticalConflicts[0].action === Action.OVERRIDE_TO_FAIL) {
      finalResult = 'FAIL';
      overridden = true;
    } else if (criticalConflicts[0].action === Action.OVERRIDE_TO_PASS) {
      finalResult = 'PASS';
      overridden = true;
    }
  }

  return {
    original_ai_status: aiResult.status,
    final_status: finalResult,
    overridden,
    conflicts,
    deterministic_summary: {
      compile,
      tests,
      dependencies,
      all_critical_pass: allCriticalPass,
      all_pass: allPass,
    },
    human_review_required: conflicts.some(c => c.action === Action.HUMAN_REVIEW),
  };
}

/**
 * Log validation result
 *
 * @param {string} service - Service name
 * @param {Object} validationResult - Validation result
 */
function logValidation(service, validationResult) {
  const logDir = path.join(process.cwd(), '.harness/logs/judgments');

  // Create log directory if not exists
  if (!fs.existsSync(logDir)) {
    fs.mkdirSync(logDir, { recursive: true });
  }

  const logEntry = {
    timestamp: new Date().toISOString(),
    service,
    ...validationResult,
  };

  // Append to daily log file
  const date = new Date().toISOString().split('T')[0];
  const logFile = path.join(logDir, `${date}.json`);

  let logs = [];
  if (fs.existsSync(logFile)) {
    logs = JSON.parse(fs.readFileSync(logFile, 'utf8'));
  }

  logs.push(logEntry);
  fs.writeFileSync(logFile, JSON.stringify(logs, null, 2));

  return logFile;
}

/**
 * Format validation result for display
 *
 * @param {Object} validationResult - Validation result
 * @returns {string} Formatted message
 */
function formatValidationResult(validationResult) {
  let message = '';

  if (validationResult.overridden) {
    message += `⚠️  AI判断已被推翻\n`;
    message += `   原始判断: ${validationResult.original_ai_status}\n`;
    message += `   最终结果: ${validationResult.final_status}\n\n`;
  }

  if (validationResult.conflicts.length > 0) {
    message += `检测到 ${validationResult.conflicts.length} 个冲突:\n\n`;

    validationResult.conflicts.forEach((conflict, i) => {
      message += `${i + 1}. [${conflict.severity.toUpperCase()}] ${conflict.message}\n`;
      message += `   原因: ${conflict.detail}\n`;
      message += `   建议: ${conflict.recommendation}\n`;
      if (conflict.ai_reason) {
        message += `   AI理由: ${conflict.ai_reason}\n`;
      }
      message += `\n`;
    });
  }

  if (validationResult.human_review_required) {
    message += `⚠️  需要人工审查\n`;
  }

  return message;
}

module.exports = {
  validateAIJudgment,
  logValidation,
  formatValidationResult,
  ConflictType,
  Severity,
  Action,
};
