#!/usr/bin/env python3
"""
监控 Claude 会话日志，检测工具调用循环

用法：
  python3 monitor-tool-loops.py [log-file]

如果检测到循环，输出警告信息。
"""

import sys
import json
import re
from collections import deque
from typing import Dict, List, Optional


def parse_tool_calls(log_content: str) -> List[Dict]:
    """从日志中提取工具调用"""
    tool_calls = []

    # 匹配工具调用模式（根据实际日志格式调整）
    # 示例：<invoke name="TaskList">
    pattern = r'<invoke name="(\w+)">(.*?)</invoke>'

    for match in re.finditer(pattern, log_content, re.DOTALL):
        tool_name = match.group(1)
        params_xml = match.group(2)

        # 提取参数
        params = {}
        param_pattern = r'<parameter name="(\w+)">(.*?)