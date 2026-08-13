# CLAUDE.md 补充 — 工具调用失败处理

将以下内容添加到 CLAUDE.md 的"硬性约束"部分：

---

## 8. 工具调用失败处理（防止死循环）

### 规则

**连续 2 次相同的工具调用失败后，必须停下来诊断根本原因。**

**触发条件**：
- 工具名称相同
- 错误类型相同（如 `InputValidationError`）
- 参数完全相同

### 诊断步骤

1. **停止重复**：不要第 3 次尝试相同的调用
2. **阅读工具定义**：从系统提示中查找工具的参数要求
3. **对比差异**：你的参数 vs 工具定义
4. **分析根因**：为什么失败？
5. **尝试新方法**：完全不同的工具或参数

### 示例

#### ❌ 错误（禁止）

```
第 1 次：TaskList(command="bash ...", description="...")
错误：InputValidationError - unexpected parameter 'command'

第 2 次：TaskList(command="bash ...", description="...")
错误：InputValidationError - unexpected parameter 'command'

第 3 次：TaskList(command="bash ...", description="...")  ← 禁止！
```

#### ✅ 正确

```
第 1 次：TaskList(command="bash ...", description="...")
错误：InputValidationError - unexpected parameter 'command'

第 2 次：TaskList(command="bash ...", description="...")
错误：InputValidationError - unexpected parameter 'command'

[停下来诊断]
- 阅读 TaskList 工具定义 → 发现不接受任何参数
- 根因：我一直传递了工具不支持的参数
- 新方法：分两步执行
  1. TaskList() — 查看当前任务
  2. Bash("bash .harness/scripts/harness-tasks.sh scan") — 扫描新任务

第 3 次：TaskList()
结果：成功 ✅
```

### 特殊情况

**允许重试**（不触发此规则）：
- 瞬态错误：`NetworkError`, `TimeoutError`
- 参数不同：说明在尝试不同方法
- 错误类型不同：说明问题在变化

### 案例

2026-06-21 会话中，Agent 陷入 TaskList 死循环：
- 连续 30+ 次相同调用
- 浪费 27K tokens（13.5% 预算）
- 原因：没有遵守此规则

---

## 相关资源

- **熔断器设计**: `.harness/docs/circuit-breaker.md`
- **集成指南**: `.harness/docs/circuit-breaker-integration.md`
