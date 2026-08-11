---
triggers: "模拟数据 测试数据 管线 数据可信度 loop-run 63次 时间戳 数据质量 untracked"
status: active
severity: must-follow
type: pitfall
created: 2026-06-17
updated: 2026-06-17
last_applied: null
apply_count: 0
---

# 管线数据可信度：模拟数据 vs 真实数据

## 场景

2026-06-17 管线审查时发现：
- 3 个 task 各有 63 次"开始执行"记录（`open ↔ in_progress` ping-pong）
- 11 个 loop-run 的 sensor 数据连续 2 天不变（QA 3 FAIL, Review 9 WARNINGs）
- Sensor 5 报告 "1976 open issues"

## 根因

这些数据是管线开发期间的**模拟/压力测试产物**：
- 所有 loop-run 文件共享同一 mtime（`2026-06-17 15:20:41`），尽管文件名跨越多天
- 所有 task 文件均为 untracked（从未进 git）
- 63 次状态变更由 `harness-tasks.sh status` 循环调用 `sed -i` 追加造成

## 教训

**代码缺陷的真实性 ≠ 数据规模的严重性。**

- 模拟数据能暴露设计缺口（如缺少收敛检测、JSON 解析 bug）——这些是真实问题
- 但不能用模拟数据的规模来论证影响面——"63 次派发零进展"在生产中不会发生
- 管线是否"完美"不能由模拟数据的 FAIL count 判断，必须等 50+ 真实任务跑完后用 `retrospective` 命令验证

## 实践规则

1. 分析管线问题时，先区分数据是模拟还是生产：`git status .harness/tasks/ .harness/loop-runs/`
2. 用代码审查（read the code）验证问题是否真实存在，而非仅依赖数据表现
3. 生产数据跑够 50 个任务后再做首次复盘
