# Circuit Breaker for Tool Calls

## 问题背景

在 2026-06-21 的任务执行中，Agent 陷入了工具调用死循环：
- 连续 30+ 次调用 `TaskList(command=..., description=...)`
- 每次都收到 `InputValidationError: unexpected parameter`
- 浪费了约 27K tokens（13.5% 预算）
- 严重影响用户体验

**根本原因**：Agent 没有遵循"失败两次后换方法"的原则。

---

## 解决方案

### 设计理念

在 harness 工具执行层添加**轻量级熔断器**，自动检测并阻止重复失败的工具调用。

### 核心逻辑

```
工具调用 → harness → [熔断器检查] → 执行工具
                            ↓
                    检测到 3 次相同失败？
                            ↓
                    是：返回诊断信息，阻止执行
                    否：记录失败，允许继续
```

### 触发条件

**触发熔断**（连续 3 次相同）：
- ✅ 工具名称相同
- ✅ 错误类型相同（`InputValidationError`, `ToolNotFoundError` 等）
- ✅ 参数 hash 相同（说明完全重复）

**不触发熔断**（正常重试）：
- ❌ 瞬态错误（`NetworkError`, `TimeoutError`）
- ❌ 参数不同（说明 Agent 在尝试不同方法）
- ❌ 错误类型不同

---

## 实现

### 1. 核心代码

文件：`.harness/scripts/circuit_breaker.py`

关键类：`ToolCallCircuitBreaker`
- `check_and_record()`: 检查并记录失败
- `_compute_signature()`: 计算失败签名
- `_build_diagnostic_message()`: 生成诊断信息

### 2. 集成点

在 harness 工具执行层集成：

```python
# 伪代码示例
def execute_tool(tool_name, params):
    try:
        result = actual_tool_execution(tool_name, params)
        return result
    except ToolError as e:
        # 检查是否触发熔断
        diagnostic = circuit_breaker.check_and_record(
            tool_name=tool_name,
            error_type=type(e).__name__,
            params=params,
            error_message=str(e)
        )
        
        if diagnostic:
            # 触发熔断，返回诊断信息
            raise CircuitBreakerError(diagnostic)
        else:
            # 正常失败，返回原始错误
            raise e
```

### 3. 配置

环境变量：
- `HARNESS_CIRCUIT_BREAKER_THRESHOLD`: 触发阈值（默认 3）
- `HARNESS_CIRCUIT_BREAKER_WINDOW`: 窗口大小（默认 10）

---

## 效果演示

### 场景 1：触发熔断

```python
# 第 1 次调用
TaskList(command="bash ...", description="...")
# 错误：InputValidationError - unexpected parameter 'command'
# 行为：记录失败，允许继续

# 第 2 次调用
TaskList(command="bash ...", description="...")
# 错误：InputValidationError - unexpected parameter 'command'
# 行为：记录失败，允许继续

# 第 3 次调用
TaskList(command="bash ...", description="...")
# 触发熔断！返回诊断信息：
```

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

### 场景 2：不触发熔断（正常重试）

```python
# 尝试不同参数
TaskList(command="bash script1.sh")  # 失败
TaskList(command="bash script2.sh")  # 失败
TaskList(command="bash script3.sh")  # 失败
# 行为：参数不同，允许继续（Agent 在尝试不同方法）

# 瞬态错误
Bash("curl http://api.example.com")  # NetworkError
Bash("curl http://api.example.com")  # NetworkError
Bash("curl http://api.example.com")  # NetworkError
# 行为：网络错误，允许重试（可能是暂时的网络问题）
```

---

## 预期收益

### Token 节省
- **原来**：30+ 次重复调用 = ~27K tokens 浪费
- **现在**：3 次后熔断 = 节省 90% token

### 用户体验
- **原来**：看到 30+ 次重复错误，挫败感
- **现在**：3 次后看到清晰的诊断信息，快速理解问题

### Agent 行为改进
- 强制 Agent 在失败 3 次后停下来思考
- 通过诊断信息引导 Agent 找到正确方法
- 符合"失败两次后换方法"的设计原则

---

## 测试

运行测试：
```bash
cd /home/jiaoxh/my-project/community-and-home
python3 .harness/scripts/circuit_breaker.py
```

预期输出：
```
第 1 次调用：记录失败，允许继续
第 2 次调用：记录失败，允许继续
第 3 次调用触发熔断：
[显示诊断信息]
```

---

## 待办事项

- [ ] 将 circuit_breaker.py 集成到 harness 工具执行层
- [ ] 添加单元测试
- [ ] 添加配置文件支持（.harness/config/circuit-breaker.yaml）
- [ ] 添加监控和日志（记录熔断事件）
- [ ] 添加白名单机制（某些工具允许更多重试）

---

## 相关资源

- 实现代码：`.harness/scripts/circuit_breaker.py`
- 任务追踪：Task #4 "为 harness 添加工具调用熔断器"
- 问题案例：2026-06-21 TaskList 死循环事件
