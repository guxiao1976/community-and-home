# Harness 文档中心

> 开发流水线文档索引 · 人类开发者参考

---

## 📚 架构与设计

### 核心文档（流水线体系）

| 文档 | 用途 | 目标读者 |
|------|------|---------|
| [pipeline-architecture.md](./pipeline-architecture.md) | 流水线架构设计（三层体系、21 项检查、Agent 机制） | 新成员、架构师 |
| [pipeline-evolution.md](./pipeline-evolution.md) | 演进历史（11 个 Phase、改进效果） | 架构师、维护者 |
| [pipeline-patterns.md](./pipeline-patterns.md) | 最佳实践（扩展指南、代码模板、故障排查） | 开发者、扩展者 |

**使用场景**：
- 🎯 **新成员入职**：按顺序阅读三份文档，了解流水线全貌
- 🔧 **添加新检查**：查阅 `pipeline-patterns.md` § 2
- 🔍 **故障排查**：查阅 `pipeline-patterns.md` § 6
- 📖 **了解历史**：查阅 `pipeline-evolution.md`

---

## 🛠️ 流程与规范

| 文档 | 用途 | 更新频率 |
|------|------|:---:|
| [tool-usage-rules.md](./tool-usage-rules.md) | Agent 工具使用规则（7 个场景映射） | 发现反模式时 |
| [circuit-breaker.md](./circuit-breaker.md) | 熔断器机制设计 | 稳定 |
| [circuit-breaker-integration.md](./circuit-breaker-integration.md) | 熔断器集成方案 | 稳定 |

---

## 📝 阶段完成记录（历史存档）

这些文档记录了流水线各阶段的实施过程，主要用于追溯：

| 文档 | 内容 |
|------|------|
| [PHASE1-COMPLETE.md](./PHASE1-COMPLETE.md) | Phase 1: 机械化检查基础（8 项） |
| [PHASE1-2-COMPLETE.md](./PHASE1-2-COMPLETE.md) | Phase 1.2: TDD 证据检查 |
| [PHASE1-3-COMPLETE.md](./PHASE1-3-COMPLETE.md) | Phase 1.3: 知识图谱集成 |
| [PHASE2-COMPLETE.md](./PHASE2-COMPLETE.md) | Phase 2: 前端检查集成 |
| [PHASE3-COMPLETE.md](./PHASE3-COMPLETE.md) | Phase 3: Pipeline 编排优化 |
| [phase1-implementation-summary.md](./phase1-implementation-summary.md) | Phase 1 总结 |

---

## 🔗 关联资源

### Agent 执行文档（Agent 工作时使用）

这些文档**不在本目录**，由 Agent 在执行时加载：

| 文档 | 路径 | 用途 |
|------|------|------|
| Owner Agent 调度规则 | `/.harness/agents/owner-agent.md` | 路径选择、阶段编排 |
| QA 技能定义 | `/.harness/skills/qa/SKILL.md` | QA Agent 执行步骤 |
| Review 技能定义 | `/.harness/skills/review.md` | Review 9 维度定义 |
| Go 检查脚本 | `/.harness/skills/qa/scripts/harness-checks.sh` | 15 项机械化检查 |
| 前端检查脚本 | `/.harness/skills/qa/scripts/harness-checks-frontend.sh` | 6 项前端检查 |
| Pipeline 工作流 | `/.harness/workflows/harness-pipeline.js` | Generator→QA→Debug→Review 循环 |

### 项目规范

| 文档 | 路径 | 用途 |
|------|------|------|
| 项目编码规范 | `/.harness/rules/项目编码规范.md` | Snowflake ID、配置、文档约定 |
| Proto 管理规范 | `/.harness/rules/Proto管理规范.md` | Proto 变更流程 |
| 工程结构 | `/.harness/rules/工程结构.md` | 服务分层、中间件 |

### 知识与记忆

| 资源 | 路径 | 用途 |
|------|------|------|
| 知识图谱索引 | `/.harness/knowledge/INDEX.md` | 架构/业务/数据知识 |
| 项目记忆 | `/.harness/knowledge/memory/MEMORY.md` | 踩坑经验、技术决策 |
| 变更追溯 | `/.harness/changes/INDEX.md` | 历史变更索引 |
| 任务管理 | `/.harness/tasks/BACKLOG.md` | 当前待办事项 |

---

## 🎯 快速导航

### 我想了解...

| 需求 | 推荐阅读 |
|------|---------|
| 流水线是如何工作的？ | [pipeline-architecture.md](./pipeline-architecture.md) § 2-4 |
| 为什么会有这个检查项？ | [pipeline-evolution.md](./pipeline-evolution.md) 找对应 Phase |
| 如何添加新的检查？ | [pipeline-patterns.md](./pipeline-patterns.md) § 2 |
| 如何扩展 Review 维度？ | [pipeline-patterns.md](./pipeline-patterns.md) § 3 |
| QA 检查失败了怎么办？ | [pipeline-patterns.md](./pipeline-patterns.md) § 6.3 |
| 流水线的演进历史？ | [pipeline-evolution.md](./pipeline-evolution.md) 全文 |
| Agent 工具使用不当？ | [tool-usage-rules.md](./tool-usage-rules.md) |
| 工具调用死循环？ | [circuit-breaker.md](./circuit-breaker.md) |

---

## 📌 文档维护

### 更新规则

| 文档 | 更新时机 | 维护者 |
|------|---------|--------|
| `pipeline-architecture.md` | 架构变更时 | Owner Agent |
| `pipeline-evolution.md` | 新 Phase 完成时 | Owner Agent |
| `pipeline-patterns.md` | 发现新模式时 | 开发团队 |
| `tool-usage-rules.md` | 发现反模式时 | Owner Agent |

### 一致性检查

每月审查文档与代码的一致性：
- [ ] 检查项数量是否匹配（当前 15 Go + 6 前端）
- [ ] 脚本路径是否正确
- [ ] 数据指标是否更新
- [ ] 新增的最佳实践是否记录

---

## 💡 贡献指南

发现有价值的模式或遇到典型问题？欢迎更新文档：

1. **新的最佳实践** → 更新 `pipeline-patterns.md`
2. **重大架构改进** → 更新 `pipeline-evolution.md` + 追加新 Phase
3. **工具使用反模式** → 更新 `tool-usage-rules.md`
4. **常见故障案例** → 更新 `pipeline-patterns.md` § 6.3

---

**文档版本**: v1.0  
**最后更新**: 2026-06-22  
**维护者**: Owner Agent & 开发团队
