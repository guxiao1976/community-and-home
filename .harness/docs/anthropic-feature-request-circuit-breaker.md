# Feature Request: Circuit Breaker for Tool Call Loops

## To: Anthropic Claude Code Team

---

## Problem

Claude agents can get stuck in infinite tool call loops, wasting tokens and degrading user experience.

### Real Case (2026-06-21)

**Scenario**: Agent repeatedly called `TaskList(command="...", description="...")` despite receiving identical `InputValidationError` messages.

**Impact**:
- 30+ identical failed calls in session 1
- 15+ identical failed calls in session 2 (same conversation)
- ~30K tokens wasted (15% of budget)
- Severe user frustration

**Root Cause**: Agent failed to follow its own directive: "If an approach has failed twice, diagnose the root cause rather than making incremental patches."

---

## Proposed Solution

Add a lightweight **Circuit Breaker** in Claude Code's tool execution layer that automatically detects and interrupts repetitive failures.

### Design

#### Detection Logic

```typescript
class CircuitBreaker {
  private failureWindow: Signature[] = [];
  private threshold = 3;

  check(toolName: string, errorType: string, params: Record<string, any>): string | null {
    const signature = this.computeSignature(toolName, errorType, params);
    
    this.failureWindow.push(signature);
    
    const recent = this.failureWindow.slice(-this.threshold);
    if (recent.length === this.threshold && 
        recent.every(s => JSON.stringify(s) === JSON.stringify(signature))) {
      return this.buildDiagnostic(toolName, errorType, params);
    }
    
    return null;
  }
}
```

#### When to Trigger

**Trigger circuit breaker** (3 consecutive identical failures):
- ✅ Same tool name
- ✅ Same error type (`InputValidationError`, `ToolNotFoundError`)
- ✅ Same parameters (hash match)

**Do NOT trigger** (allow retry):
- ❌ Transient errors (`NetworkError`, `TimeoutError`)
- ❌ Different parameters (agent is trying different approaches)
- ❌ Different error types

#### Response

When triggered, return a structured diagnostic message instead of the raw error:

```
╔═══════════════════════════════════════════════════════════════════════╗
║                  🔴 CIRCUIT BREAKER ACTIVATED                         ║
╚═══════════════════════════════════════════════════════════════════════╝

You have called TaskList() 3 times with IDENTICAL errors.
This is NOT a transient failure - it's a LOGIC ERROR in your approach.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Error Pattern:
  Tool: TaskList
  Error: InputValidationError
  Message: unexpected parameter 'command'

Your Parameters:
  - command: bash ...
  - description: ...

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Required Actions:

1. ⛔ STOP - Do NOT retry this call again
2. 📖 READ the tool definition for TaskList
3. 🔍 DIAGNOSE why your parameters don't match the tool definition
4. 🔄 TRY a completely different approach

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Hint: If the error says "unexpected parameter", the tool likely takes
FEWER parameters than you're providing. Check if it accepts NO parameters.
```

---

## Expected Benefits

### Token Savings

| Scenario | Without Circuit Breaker | With Circuit Breaker | Savings |
|----------|------------------------|---------------------|---------|
| Tool call loop | 30+ attempts | 3 attempts | 90% |
| Token waste (case study) | 30K tokens | 3K tokens | 27K tokens |

### User Experience

**Before**:
- User sees 30+ identical errors scrolling by
- Frustration and doubt about agent capability
- Manual intervention required

**After**:
- Agent stops after 3 attempts
- Clear diagnostic message guides understanding
- Faster resolution

### Agent Behavior

- Forces agent to stop and reflect after 3 failures
- Diagnostic message provides structured guidance
- Aligns with "fail twice, try different approach" principle

---

## Implementation

### Location

```
claude-code/
├── src/
│   ├── tools/
│   │   ├── executor.ts          ← Integration point
│   │   └── circuit-breaker.ts   ← New module
```

### Integration Point

```typescript
// src/tools/executor.ts

import { circuitBreaker } from './circuit-breaker';

export async function executeTool(toolName: string, params: any): Promise<ToolResult> {
  try {
    return await actualToolExecution(toolName, params);
  } catch (error) {
    const diagnostic = circuitBreaker.check(
      toolName,
      error.constructor.name,
      params
    );
    
    if (diagnostic) {
      throw new CircuitBreakerError(diagnostic);
    }
    
    throw error;
  }
}
```

### Configuration

```json
// User config: ~/.claude/config.json
{
  "circuitBreaker": {
    "enabled": true,
    "threshold": 3,
    "windowSize": 10,
    "whitelist": ["WebSearch"]  // Tools that allow more retries
  }
}
```

---

## Reference Implementation

A complete Python implementation is available:
- Design doc: `.harness/docs/circuit-breaker.md`
- Code: `.harness/scripts/circuit_breaker.py` (191 lines)
- Integration guide: `.harness/docs/circuit-breaker-integration.md`

Can be shared upon request.

---

## Testing Strategy

### Unit Tests

```typescript
test('triggers after 3 identical failures', () => {
  const breaker = new CircuitBreaker();
  
  breaker.check('TaskList', 'InputValidationError', { cmd: 'bash' });
  breaker.check('TaskList', 'InputValidationError', { cmd: 'bash' });
  
  const result = breaker.check('TaskList', 'InputValidationError', { cmd: 'bash' });
  expect(result).toContain('CIRCUIT BREAKER ACTIVATED');
});

test('does not trigger for different parameters', () => {
  const breaker = new CircuitBreaker();
  
  breaker.check('TaskList', 'InputValidationError', { cmd: 'a' });
  breaker.check('TaskList', 'InputValidationError', { cmd: 'b' });
  const result = breaker.check('TaskList', 'InputValidationError', { cmd: 'c' });
  
  expect(result).toBeNull();
});

test('does not trigger for transient errors', () => {
  const breaker = new CircuitBreaker();
  
  breaker.check('WebFetch', 'NetworkError', { url: 'http://api.example.com' });
  breaker.check('WebFetch', 'NetworkError', { url: 'http://api.example.com' });
  const result = breaker.check('WebFetch', 'NetworkError', { url: 'http://api.example.com' });
  
  expect(result).toBeNull();
});
```

### Integration Tests

Simulate real agent scenarios:
- TaskList with wrong parameters
- Read with non-existent files (should NOT trigger - file might be generated)
- WebFetch with network errors (should NOT trigger - transient)

---

## Risks and Mitigations

### Risk: False Positives

**Scenario**: Legitimate retries blocked by circuit breaker

**Mitigation**:
- Whitelist transient error types (`NetworkError`, `TimeoutError`)
- Allow configuration per tool (some tools need more retries)
- User can disable via config

### Risk: User Confusion

**Scenario**: User doesn't understand why circuit breaker triggered

**Mitigation**:
- Clear diagnostic message explains the pattern
- Provides actionable steps
- Links to tool definition

---

## Priority

**Recommendation**: P1 (High Priority)

**Reasoning**:
- Common failure pattern (affects all Claude Code users)
- Significant token waste (15% in documented case)
- Low implementation cost (~200 lines of code)
- High user experience improvement

---

## Contact

This feature request was generated from a real production issue in a multi-service Go project using Claude Code for development automation.

Reference case: Community & Home project, 2026-06-21 session

---

## Appendix: Alternative Considered

### System Prompt Reinforcement

**Approach**: Add stronger language to system prompt about avoiding repetitive failures

**Verdict**: ❌ Insufficient
- Already exists in system prompt
- Agent failed to follow it multiple times
- Relies on agent self-discipline

### User-Side Monitoring

**Approach**: User writes scripts to monitor logs and detect loops

**Verdict**: ⚠️ Partial solution
- Requires active monitoring by user
- Detection happens after the fact
- Cannot interrupt mid-loop

### Conclusion

Circuit breaker is the only solution that:
- Works automatically without user intervention
- Interrupts loops in real-time
- Provides structured feedback to agent
