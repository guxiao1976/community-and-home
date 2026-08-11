---
triggers: ["提交", "commit", "代码变更", "harness-checks", "QA", "门禁", "gate"]
service: all
type: process
severity: must-follow
status: active
created: 2026-06-07
updated: 2026-06-07
last_applied: null
apply_count: 0
---

# 代码变更提交前必须通过机械化检查

## 为什么会有这条经验

一次全量代码审查发现了 2 个 CRITICAL 缺陷（API 层漏传 building/unit/room 字段），原因是子 Agent 执行后跳过了 QA 和 Reviewer 环节，未能发现跨层变更遗漏。机械化检查可以捕获这类问题，不依赖 Agent 判断力。

## 怎么做

1. 任何 `services/<name>/` 下的 Go 代码变更，提交前必须运行：
   ```bash
   bash .harness/skills/qa/scripts/harness-checks.sh --service <name>
   ```
2. 8 项检查全部 PASS 或仅有已知 WARN → 可以提交
3. 有任何 FAIL → **不能提交**，必须先修复
4. 适用于全局 Claude 和所有服务子 Claude，无一例外

## 怎么验证

```bash
# 单服务检查
bash .harness/skills/qa/scripts/harness-checks.sh --service user-service

# 检查退出码（非 0 表示有 FAIL）
echo $?   # 0 = 全部通过，可提交
```

## 关联经验
- [[proto-jstype]]
- [[grpc-only-comms]]
- [[verify-api-before-calling]]
