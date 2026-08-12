---
triggers: ["_qa.md", "过时", "QA FAIL", "TDD 修复后", "测试补齐", "QA 报告未刷新", "评审输入", "状态不一致", "交付物"]
service: all
severity: should-follow
type: process
status: active
created: 2026-08-12
updated: 2026-08-12
---

# QA FAIL 修复后必须重跑并刷新 _qa.md，否则交付物含过时 FAIL 报告

## 为什么会有这条经验

QA 判定 FAIL 后，修复（补齐测试/改代码）完成并确认全绿时，必须重跑 QA 并更新 _qa.md（VERDICT 同步为 PASS、测试表刷新），否则交付物携带修复前的 FAIL 快照，后续评审/管线以过时报告判读，产生误判。

社区案例（community-hub-service，2026-08-12）：修复 8 个 TDD 缺口测试后 go build/test 全绿（本审查独立复跑确认），但 _qa.md 仍为修复前 VERDICT FAIL 快照（TDD 证据表 8 项 FAIL），与代码实际状态不一致。

## 怎么做

任何针对 QA FAIL 的修复，验收闭环必须包含「重跑 QA + 更新 _qa.md + 更新 CHANGELOG」三件套。

## 怎么验证

- 修复后检查 _qa.md 的 VERDICT 是否已为 PASS、测试表是否刷新
- 交付物评审前核对 _qa.md 与代码实际状态一致

## 关联经验

[[pre-commit-checks]] [[tdd-red-evidence-requires-fail-excerpt]]
