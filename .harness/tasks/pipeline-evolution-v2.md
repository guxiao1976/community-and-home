# Pipeline 自我进化 V2

**创建日期**: 2026-08-10
**预计启动**: 2026-08-24（2 周后，Incident ≥ 5 条）
**状态**: pending

## 当前状态（V1 已完成）

- [x] Incident 记录模板 (`_template.yml`)
- [x] 首条 Incident (`2026-08-10-menu-refactor.yml`)
- [x] evolve-pipeline.sh（grep 方案）
- [x] gate-engine.js verify 门禁
- [x] verify-before-deliver skill

## V2 要做

1. [ ] Incident 解析从 grep 升级为完整 YAML（Python/Node.js）
2. [ ] Pattern 权重排序（`pain_rounds × 出现次数`）
3. [ ] 门禁自动升级建议（≥3 次 → WARN→BLOCK）
4. [ ] 新 pattern → 生成 gate-engine.js 新增代码
5. [ ] 输出结构化 PR body（Markdown）

## 触发条件

`.harness/logs/incidents/` 中积累 ≥ 5 条记录（不含 `_template.yml`），或手动触发：
```bash
bash .harness/scripts/evolve-pipeline.sh
```
