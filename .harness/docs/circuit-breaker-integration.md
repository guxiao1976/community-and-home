# Circuit Breaker 集成指南

## 背景

这是一个针对 **Claude Code 本身** 的改进建议，而非用户项目代码的集成。

### 问题
- Agent 可能陷入重复失败的工具调用死循环
- 案例：2026-06-21 任务中，30+ 次调用 `TaskList(command=...)` 导致相同错误
- 浪费 27K tokens（13.5% 预算），严重影响用户体验

### 解决方案
在 Claude Code 的工具执行层添加熔断器，自动检测并阻止重复失败。

---

## 集成位置（Claude Code 内部）

### 架构层次
```
Claude Opus 4.8
    ↓
Claude Code (CLI/Desktop/Web)
    ↓
Tool Execution Layer  ← 在这里集成熔断器
    ↓
实际工具（TaskList, Bash, Read, etc.）
```

### 具体位置（推测）

Claude Code 的工具执行可能在：
```
claude-code/
├── src/
│   ├── tools/
│   │   ├── executor.ts          ← 工具执行入口
│   │   ├── registry.ts          ← 工具注册表
│   │   └── validation.ts        ← 参数验证
│   └── circuit-breaker.ts       ← 新增：熔断器
```

---

## 集成伪代码

### 1. 添加熔断器模块

```typescript
// src/circuit-breaker.ts

interface ToolCallSignature {
  toolName: string;
  errorType: string;
  paramsHash: string;
}

class CircuitBreaker {
  private failureWindow: ToolCallSignature[] = [];
  private readonly threshold = 3;
  private readonly windowSize = 10;

  checkAndRecord(
    toolName: string,
    errorType: string,
    params: Record<string, any>,
    errorMessage: string
  ): string | null {
    // 跳过瞬态错误
    const transientErrors = ['NetworkError', 'TimeoutError', 'FileNotFoundError'];
    if (transientErrors.includes(errorType)) {
      return null;
    }

    // 计算签名
    const signature = this.computeSignature(toolName, errorType, params);

    // 记录失败
    this.failureWindow.push(signature);
    if (this.failureWindow.length > this.windowSize) {
      this.failureWindow.shift();
    }

    // 检查最近 N 次是否相同
    const recent = this.failureWindow.slice(-this.threshold);
    if (recent.length === this.threshold && 
        recent.every(s => JSON.stringify(s) === JSON.stringify(signature))) {
      return this.buildDiagnostic(toolName, errorType, params, errorMessage);
    }

    return null;
  }

  private computeSignature(
    toolName: string,
    errorType: string,
    params: Record<string, any>
  ): ToolCallSignature {
    const paramsJson = JSON.stringify(params, Object.keys(params).sort());
    const paramsHash = this.hashString(paramsJson).substring(0, 8);
    return { toolName, errorType, paramsHash };
  }

  private hashString(str: string): string {
    // Simple hash function
    let hash = 0;
    for (let i = 0; i < str.length; i++) {
      hash = ((hash << 5) - hash) + str.charCodeAt(i);
      hash = hash & hash;
    }
    return Math.abs(hash).toString(16);
  }

  private buildDiagnostic(
    toolName: string,
    errorType: string,
    params: Record<string, any>,
    errorMessage: string
  ): string {
    const paramList = Object.entries(params)
      .map(([k, v]) => `  - ${k}: ${v}`)
      .join('\n');

    return `
╔═══════════════════════════════════════════════════════════════════════╗
║                  🔴 CIRCUIT BREAKER ACTIVATED                         ║
╚═══════════════════════════════════════════════════════════════════════╝

You have called ${toolName}() ${this.threshold} times with IDENTICAL errors.
This is NOT a transient failure - it's a LOGIC ERROR in your approach.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Error Pattern:
  Tool: ${toolName}
  Error: ${errorType}
  Message: ${errorMessage}

Your Parameters:
${paramList}

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Required Actions:

1. ⛔ STOP - Do NOT retry this call again
2. 📖 READ the tool definition for ${toolName}
3. 🔍 DIAGNOSE why your parameters don't match the tool definition
4. 🔄 TRY a completely different approach

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Hint: If the error says "unexpected parameter", the tool likely takes
FEWER parameters than you're providing. Check if it accepts NO parameters.

If you believe this is a false positive, EXPLAIN YOUR REASONING before
making another attempt.
`;
  }
}

export const circuitBreaker = new CircuitBreaker();
```

### 2. 修改工具执行器

```typescript
// src/tools/executor.ts

import { circuitBreaker } from '../circuit-breaker';

export async function executeTool(
  toolName: string,
  params: Record<string, any>
): Promise<ToolResult> {
  try {
    // 执行工具
    const result = await actualToolExecution(toolName, params);
    return result;
  } catch (error) {
    const errorType = error.constructor.name;
    const errorMessage = error.message;

    // 检查是否触发熔断
    const diagnostic = circuitBreaker.checkAndRecord(
      toolName,
      errorType,
      params,
      errorMessage
    );

    if (diagnostic) {
      // 触发熔断，返回诊断信息
      throw new CircuitBreakerError(diagnostic);
    }

    // 正常失败，继续抛出原始错误
    throw error;
  }
}

class CircuitBreakerError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'CircuitBreakerError';
  }
}
```

---

## 配置选项

### 环境变量

```bash
# 触发阈值（默认 3）
CLAUDE_CODE_CIRCUIT_BREAKER_THRESHOLD=3

# 滑动窗口大小（默认 10）
CLAUDE_CODE_CIRCUIT_BREAKER_WINDOW=10

# 启用/禁用（默认启用）
CLAUDE_CODE_CIRCUIT_BREAKER_ENABLED=true
```

### 用户配置文件

```json
// ~/.claude/config.json
{
  "circuitBreaker": {
    "enabled": true,
    "threshold": 3,
    "windowSize": 10,
    "whitelist": ["WebSearch", "WebFetch"]  // 允许更多重试的工具
  }
}
```

---

## 测试策略

### 单元测试

```typescript
import { circuitBreaker } from '../circuit-breaker';

describe('CircuitBreaker', () => {
  beforeEach(() => {
    circuitBreaker.reset();
  });

  test('triggers after 3 identical failures', () => {
    const params = { command: 'bash ...', description: '...' };
    
    // 第 1-2 次：不触发
    expect(circuitBreaker.checkAndRecord('TaskList', 'InputValidationError', params, 'msg')).toBeNull();
    expect(circuitBreaker.checkAndRecord('TaskList', 'InputValidationError', params, 'msg')).toBeNull();
    
    // 第 3 次：触发
    const result = circuitBreaker.checkAndRecord('TaskList', 'InputValidationError', params, 'msg');
    expect(result).toContain('CIRCUIT BREAKER ACTIVATED');
  });

  test('does not trigger for different parameters', () => {
    // 参数不同
    circuitBreaker.checkAndRecord('TaskList', 'InputValidationError', { command: 'a' }, 'msg');
    circuitBreaker.checkAndRecord('TaskList', 'InputValidationError', { command: 'b' }, 'msg');
    const result = circuitBreaker.checkAndRecord('TaskList', 'InputValidationError', { command: 'c' }, 'msg');
    expect(result).toBeNull();
  });

  test('does not trigger for transient errors', () => {
    const params = { url: 'http://api.example.com' };
    
    // 网络错误：允许重试
    circuitBreaker.checkAndRecord('WebFetch', 'NetworkError', params, 'msg');
    circuitBreaker.checkAndRecord('WebFetch', 'NetworkError', params, 'msg');
    const result = circuitBreaker.checkAndRecord('WebFetch', 'NetworkError', params, 'msg');
    expect(result).toBeNull();
  });
});
```

### 集成测试

模拟真实场景：
```typescript
test('real scenario: TaskList with wrong parameters', async () => {
  const executor = new ToolExecutor();
  
  // 第 1-2 次：正常失败
  await expect(executor.execute('TaskList', { command: 'bash ...' }))
    .rejects.toThrow('InputValidationError');
  await expect(executor.execute('TaskList', { command: 'bash ...' }))
    .rejects.toThrow('InputValidationError');
  
  // 第 3 次：熔断
  await expect(executor.execute('TaskList', { command: 'bash ...' }))
    .rejects.toThrow('CIRCUIT BREAKER ACTIVATED');
});
```

---

## 监控和日志

### 记录熔断事件

```typescript
private buildDiagnostic(...): string {
  // 记录熔断事件到日志
  logger.warn('CircuitBreaker triggered', {
    toolName,
    errorType,
    paramsHash: this.computeSignature(toolName, errorType, params).paramsHash,
    timestamp: new Date().toISOString(),
    userId: getCurrentUserId(),
    sessionId: getCurrentSessionId()
  });

  return diagnostic;
}
```

### Telemetry

```typescript
// 上报到 Anthropic 后台
telemetry.track('circuit_breaker_triggered', {
  tool_name: toolName,
  error_type: errorType,
  threshold: this.threshold,
  session_id: sessionId
});
```

---

## 向 Anthropic 提交建议

### 方式 1：GitHub Issue

如果 Claude Code 是开源的：
1. 创建 GitHub Issue
2. 标题：`Feature: Add Circuit Breaker for Repetitive Tool Call Failures`
3. 附上：
   - 问题描述（本文档的"背景"部分）
   - 设计文档（`.harness/docs/circuit-breaker.md`）
   - 实现代码（`.harness/scripts/circuit_breaker.py` 的 TypeScript 版本）
   - 测试用例

### 方式 2：官方反馈渠道

通过 Claude Code 的反馈机制：
1. 打开 Claude Code
2. 点击"反馈"或"Feature Request"
3. 提交：
   - 标题：Circuit Breaker for Tool Call Loops
   - 描述：详见本文档
   - 附件：设计文档 + 示例代码

### 方式 3：Community Forum

如果有 Claude Code 社区论坛：
1. 发帖描述问题和解决方案
2. 附上实际案例（2026-06-21 TaskList 死循环）
3. 展示预期收益（节省 90% 无效调用）

---

## 项目内部使用

在 Claude Code 官方支持之前，项目可以：

### 1. 添加到 CLAUDE.md

```markdown
## Agent 自律要求

### 工具调用失败处理

**硬性约束**：
- 同一工具调用失败 2 次后，必须停下来诊断根本原因
- 禁止连续 3 次以上相同的工具调用失败
- 触发条件：工具名 + 错误类型 + 参数完全相同

**诊断步骤**：
1. 阅读工具定义（参数、返回值、约束）
2. 对比你的调用和定义的差异
3. 尝试完全不同的方法（不同工具、不同参数）

**反例**（禁止）：
```
TaskList(command="...") → InputValidationError
TaskList(command="...") → InputValidationError
TaskList(command="...") → InputValidationError  ← 禁止！
```

**正例**（允许）：
```
TaskList(command="...") → InputValidationError
TaskList(command="...") → InputValidationError
[停下来] → 阅读 TaskList 定义 → 发现不接受参数
TaskList() → 成功  ✅
```
```

### 2. 添加到 owner-agent.md

在"禁令清单"中添加：
```markdown
### 工具调用失败循环

**触发条件**：连续 2 次相同的工具调用失败

**行动**：
1. 停止当前方法
2. 输出诊断：
   - 工具名称和失败的参数
   - 工具定义（从系统提示中引用）
   - 根本原因分析
   - 新的方法
3. 尝试完全不同的方法

**不允许**：
- 连续 3 次以上相同调用
- "也许这次会成功"的幻想
- 忽略错误信息
```

### 3. 创建监控脚本

```bash
# .harness/scripts/monitor-loops.sh
# 分析 Claude 会话日志，检测工具调用循环

#!/bin/bash
# 分析最近的会话，查找重复失败模式
# 用法：bash .harness/scripts/monitor-loops.sh [session-log-file]
```

---

## 参考资源

- **设计文档**: `.harness/docs/circuit-breaker.md`
- **Python 实现**: `.harness/scripts/circuit_breaker.py`
- **问题案例**: 2026-06-21 TaskList 死循环（本会话）
- **Token 浪费**: 27K tokens（13.5% 预算）

---

## 总结

这个熔断器是一个轻量级但有效的改进：
- **检测简单**：3 次相同失败
- **影响可控**：只针对确定性错误
- **收益明显**：节省 90% 无效调用

建议 Anthropic 在 Claude Code 中实现此功能，造福所有用户。
