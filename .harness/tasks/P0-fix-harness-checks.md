---
priority: P0
status: open
created: 2026-06-23
assignee: harness-team
source: human
estimated_effort: 1d
---

# P0: 修复 harness-checks.sh 脚本问题

## 问题描述

QA 报告显示机械化检查脚本存在 2 个问题：
1. `check_frontend: command not found` — 函数调用错误
2. 硬编码密钥检查失败 — 依赖的分类器临时不可用

## 影响范围

- 阻塞 QA 阶段自动化验证
- 降级到手工检查，可靠性下降
- 违反 Harness 机械化原则

## 修复方案

### 1. check_frontend 函数问题

**根因**: harness-checks.sh 是 Go 服务专用脚本，但内部错误引用了前端检查函数

**修复**:
```bash
# 1. 检查 harness-checks.sh 中是否有对 check_frontend 的调用
grep -n "check_frontend" .harness/skills/qa/scripts/harness-checks.sh

# 2. 如果有，移除或注释掉（Go 脚本不应调用前端检查）
# 3. 前端检查使用独立脚本 harness-checks-frontend.sh
```

### 2. 硬编码密钥检查

**根因**: 依赖外部分类器服务，服务不可用时检查失败

**修复**:
```bash
# 选项 1: 使用本地正则表达式检查（降级方案）
grep -rE "(password|secret|token|api_key)\s*=\s*['\"][^'\"]{8,}" services/$SERVICE_NAME/ \
  --exclude-dir=vendor --exclude="*_test.go"

# 选项 2: 使检查可选，失败时记录 WARN 而非 FAIL
if ! check_hardcoded_secrets; then
  log_warn "check_8" "Hardcoded secrets check failed (classifier unavailable)" \
    "Recommend manual review of config files"
else
  log_pass "check_8" "No hardcoded secrets found"
fi
```

### 3. 验证修复

```bash
# 对所有服务运行检查，确保无 command not found 错误
for svc in moderation-service user-service auth-service; do
  echo "Testing $svc..."
  bash .harness/skills/qa/scripts/harness-checks.sh --service $svc --json
done
```

## 完成标准

- [ ] harness-checks.sh 在所有 8 个服务上执行无错误
- [ ] 硬编码密钥检查失败时返回 WARN 而非阻断
- [ ] 更新文档说明前端检查使用独立脚本
- [ ] 提交 PR 并合并到 master

## 关联

- QA 报告: `services/moderation-service/_qa.md`
- 脚本文件: `.harness/skills/qa/scripts/harness-checks.sh`
- 记忆: 无（新问题）
