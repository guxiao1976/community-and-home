#!/usr/bin/env python3
"""
Circuit Breaker for Tool Call Failures

防止 Agent 陷入重复失败的工具调用死循环。
在 harness 工具执行层检测连续相同的失败模式。
"""

import hashlib
import json
from collections import deque
from typing import Dict, Optional, Tuple


class ToolCallCircuitBreaker:
    """检测并阻止重复的工具调用失败"""

    # 触发熔断的错误类型
    DETERMINISTIC_ERRORS = {
        "InputValidationError",
        "ToolNotFoundError",
        "ParameterError",
        "SchemaValidationError",
    }

    # 不触发熔断的错误类型（瞬态错误，需要重试）
    TRANSIENT_ERRORS = {
        "NetworkError",
        "TimeoutError",
        "FileNotFoundError",  # 文件可能正在生成
        "TemporaryError",
    }

    def __init__(self, threshold: int = 3, window_size: int = 10):
        """
        Args:
            threshold: 连续相同失败多少次触发熔断（默认 3）
            window_size: 记录最近多少次调用（默认 10）
        """
        self.threshold = threshold
        self.failure_window = deque(maxlen=window_size)

    def check_and_record(
        self,
        tool_name: str,
        error_type: str,
        params: Dict,
        error_message: str
    ) -> Optional[str]:
        """
        检查是否应该触发熔断。

        Args:
            tool_name: 工具名称
            error_type: 错误类型（如 InputValidationError）
            params: 调用参数
            error_message: 原始错误信息

        Returns:
            如果触发熔断，返回诊断信息；否则返回 None
        """
        # 跳过瞬态错误
        if error_type in self.TRANSIENT_ERRORS:
            return None

        # 只处理确定性错误
        if error_type not in self.DETERMINISTIC_ERRORS:
            return None

        # 计算失败签名
        signature = self._compute_signature(tool_name, error_type, params)

        # 记录本次失败
        self.failure_window.append(signature)

        # 检查最近 N 次是否完全相同
        recent = list(self.failure_window)[-self.threshold:]
        if len(recent) == self.threshold and all(s == signature for s in recent):
            return self._build_diagnostic_message(
                tool_name, error_type, params, error_message
            )

        return None

    def _compute_signature(
        self, tool_name: str, error_type: str, params: Dict
    ) -> Tuple[str, str, str]:
        """计算失败签名（用于检测重复）"""
        # 参数 hash（忽略顺序）
        params_json = json.dumps(params, sort_keys=True)
        params_hash = hashlib.md5(params_json.encode()).hexdigest()[:8]
        return (tool_name, error_type, params_hash)

    def _build_diagnostic_message(
        self,
        tool_name: str,
        error_type: str,
        params: Dict,
        error_message: str
    ) -> str:
        """构建熔断诊断信息"""
        param_list = "\n".join(f"  - {k}: {v}" for k, v in params.items())

        return f"""
╔═══════════════════════════════════════════════════════════════════════╗
║                  🔴 CIRCUIT BREAKER ACTIVATED                         ║
╚═══════════════════════════════════════════════════════════════════════╝

You have called {tool_name}() {self.threshold} times with IDENTICAL errors.
This is NOT a transient failure - it's a LOGIC ERROR in your approach.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Error Pattern:
  Tool: {tool_name}
  Error: {error_type}
  Message: {error_message}

Your Parameters:
{param_list}

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Required Actions:

1. ⛔ STOP - Do NOT retry this call again
2. 📖 READ the tool definition for {tool_name}
3. 🔍 DIAGNOSE why your parameters don't match the tool definition
4. 🔄 TRY a completely different approach

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Hint: If the error says "unexpected parameter", the tool likely takes
FEWER parameters than you're providing. Check if it accepts NO parameters.

If you believe this is a false positive, EXPLAIN YOUR REASONING before
making another attempt.
"""

    def reset(self):
        """重置熔断器状态"""
        self.failure_window.clear()


# 全局单例
_circuit_breaker = ToolCallCircuitBreaker()


def check_tool_call_failure(
    tool_name: str,
    error_type: str,
    params: Dict,
    error_message: str
) -> Optional[str]:
    """
    检查工具调用失败是否触发熔断。

    在 harness 工具执行层调用此函数。
    """
    return _circuit_breaker.check_and_record(
        tool_name, error_type, params, error_message
    )


def reset_circuit_breaker():
    """重置熔断器（用于测试）"""
    _circuit_breaker.reset()


if __name__ == "__main__":
    # 测试示例
    breaker = ToolCallCircuitBreaker(threshold=3)

    # 模拟 3 次相同失败
    for i in range(3):
        result = breaker.check_and_record(
            tool_name="TaskList",
            error_type="InputValidationError",
            params={"command": "bash ...", "description": "..."},
            error_message="unexpected parameter 'command'"
        )
        if result:
            print(f"第 {i+1} 次调用触发熔断：")
            print(result)
        else:
            print(f"第 {i+1} 次调用：记录失败，允许继续")

    print("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
    print("测试不同参数的调用（不触发熔断）：")
    breaker.reset()

    for i in range(3):
        result = breaker.check_and_record(
            tool_name="TaskList",
            error_type="InputValidationError",
            params={"command": f"bash script-{i}.sh"},  # 不同参数
            error_message="unexpected parameter 'command'"
        )
        if result:
            print(f"触发熔断（不应该发生）")
        else:
            print(f"第 {i+1} 次调用：参数不同，允许继续")
