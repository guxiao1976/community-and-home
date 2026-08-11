---
triggers: ['监控', '探测', '状态', '只读', 'monitoring']
type: guideline
severity: must-follow
service: monitoring-service
status: active
created: 2026-08-09
updated: 2026-08-09
apply_count: 0
---

# 只做探测和聚合，不修改外部系统状态

监控服务只读取和聚合状态数据，不执行任何对外部系统的修改操作。所有探测都是只读的。

## 参考

详见 `services/monitoring-service/CLAUDE.md` 和 `services/monitoring-service/docs/design.md`。
