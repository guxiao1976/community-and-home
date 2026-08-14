---
name: pipeline-review
description: 流水线检视与完善 —— 定期或按需检查开发流水线自身健康度（文档新鲜度、应用率、门禁健康、进化机制、原则符合性），跑样例验链路，产出报告并将问题回 BACKLOG。
---

# 流水线检视与完善

> 用 `.harness/docs/harness-design-principles.md` 作为标尺，逐项检查流水线自身健康。触发方式：定期（harness-loop，建议每周）或手动 `/pipeline-review`。

## Step 1: 加载标尺

读 `.harness/docs/harness-design-principles.md`（16 条原则 + 环节映射 + 目录规范）。

## Step 2: 检视 5 个维度

### 维度 1 · 文档新鲜度（提示性检查，人工判断）

```bash
# 列出各 harness 文档最后更新时间 + 最近引擎实质改动，供人工判断。
# 注：设计先行原则（第15条，文档先提交）使「文档与代码同批交付」时文档时间必然早于代码，
# 故本维度不自动 FAIL，改为列出信息由人工判断「哪些实质改动改变了文档描述的内容却未同步」。
for f in .harness/docs/pipeline-architecture.md .harness/docs/pipeline-flow-complete.md .harness/docs/pipeline-patterns.md .harness/docs/pipeline-evolution.md .harness/docs/harness-design-principles.md .harness/agents/owner-agent.md; do
  echo "--- $f"
  git log -1 --format='  最后更新: %cd  %s' --date=format:'%F %T' -- "$f" 2>/dev/null
done
echo "--- 最近引擎实质改动（workflows/scripts，供对比）"
git log -10 --format='  %cd  %s' --date=format:'%F %T' -- .harness/workflows .harness/scripts 2>/dev/null
```

判定：人工判断「文档最后更新后，是否有实质引擎改动改变了文档描述的内容却未同步」→ 是 = FAIL；否 = PASS。

### 维度 2 · 应用率

抽查 `.harness/tasks/` 最近 N 个已完成任务（completed），确认是否走了 dispatch 分级 + spec-pipeline（全流程自动化）或 harness-pipeline（编码）。判定：应用率 ≥ 90%；S 级内联需有 dispatch 分级记录 + 门禁。

### 维度 3 · 门禁健康（4 子项）

1. **机械化检查**：`bash .harness/skills/qa/scripts/harness-checks.sh --service <各服务>` → 0 FAIL；WARN 有记录且递减
2. **配置漂移**：对比 `config/quality-gates.yml` 规则数 vs `workflows/gate-engine.js` 执行数 → 无「定义了没执行」漂移
3. **pre-commit 生效**：确认 `.git/hooks/pre-commit` 挂载且能拦截
4. **门禁日志**：`ls -la .harness/logs/gates/` → 最近有记录

### 维度 4 · 进化机制运转

```bash
ls .harness/logs/incidents/*.yml | grep -v _template | wc -l   # Incident 条数
```

判定：Incident 处理率 100%；达阈值（≥5 条）则已触发 `evolve-pipeline.sh`。

### 维度 5 · 原则符合性（元检视）

对照 16 条原则 + 环节→skills/tools 映射，抽查近期 `docs/devlog/`、`.harness/changes/`、git log，确认各环节用了正确 skill/tool、无「主 agent 越界写代码」等违规。

## Step 3: 测试（正向 + 负向）

**正向**：跑一个最小无害样例任务（给某服务 model 加一行 `// Deprecated`），用 `harness-spec-pipeline.js` 走全流程（S 级轻量：dispatch → 阶段 5 编码 harness-pipeline → 阶段 6 归档），验证不卡死、产出 PASS。

**负向**：故意改坏一个测试，验证 QA 门禁正确拦截（返回 FAIL 而非假装成功）。

判定：正向全链路走通 + 负向被门禁正确拦截。

## Step 4: 闭环

1. 产出报告 `.harness/docs/pipeline-review-report-<日期>.md`（记录 5 维度 + 测试结果 + 结论）
2. 每个问题 → `bash .harness/scripts/harness-tasks.sh create ...` 回 BACKLOG
3. Incident 达阈值 → `bash .harness/scripts/evolve-pipeline.sh`
4. 报告归档作为下次检视基线
