// Test AI Judgment Validator
const validator = require('./ai-judgment-validator.js');

console.log('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
console.log('Testing AI Judgment Validator');
console.log('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
console.log('');

// Test 1: AI says PASS, but compile fails
console.log('Test 1: AI PASS + Compile FAIL');
const test1 = validator.validateAIJudgment(
  { status: 'PASS', reason: 'Code looks good' },
  {
    checks: [
      { check: 'compile', status: 'FAIL', detail: 'syntax error' },
      { check: 'tests', status: 'PASS' },
      { check: 'dependencies', status: 'PASS' },
    ]
  }
);
console.log('Result:', test1.final_status, '(overridden:', test1.overridden, ')');
console.log('Conflicts:', test1.conflicts.length);
console.log('');

// Test 2: AI says FAIL, but all pass
console.log('Test 2: AI FAIL + All Deterministic PASS');
const test2 = validator.validateAIJudgment(
  { status: 'FAIL', reason: 'Code quality issues' },
  {
    checks: [
      { check: 'compile', status: 'PASS' },
      { check: 'tests', status: 'PASS' },
      { check: 'dependencies', status: 'PASS' },
      { check: 'coverage', status: 'PASS' },
    ]
  }
);
console.log('Result:', test2.final_status, '(overridden:', test2.overridden, ')');
console.log('Human review:', test2.human_review_required);
console.log('Conflicts:', test2.conflicts.length);
console.log('');

// Test 3: AI and deterministic agree
console.log('Test 3: AI PASS + All PASS (Agreement)');
const test3 = validator.validateAIJudgment(
  { status: 'PASS', reason: 'All checks pass' },
  {
    checks: [
      { check: 'compile', status: 'PASS' },
      { check: 'tests', status: 'PASS' },
      { check: 'dependencies', status: 'PASS' },
    ]
  }
);
console.log('Result:', test3.final_status, '(overridden:', test3.overridden, ')');
console.log('Conflicts:', test3.conflicts.length);
console.log('');

// Test 4: AI says PASS, tests fail
console.log('Test 4: AI PASS + Tests FAIL');
const test4 = validator.validateAIJudgment(
  { status: 'PASS', reason: 'Implementation complete' },
  {
    checks: [
      { check: 'compile', status: 'PASS' },
      { check: 'tests', status: 'FAIL', detail: '3 tests failed' },
      { check: 'dependencies', status: 'PASS' },
    ]
  }
);
console.log('Result:', test4.final_status, '(overridden:', test4.overridden, ')');
console.log('Conflicts:', test4.conflicts.length);
console.log('');

console.log('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
console.log('All tests completed');
console.log('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
