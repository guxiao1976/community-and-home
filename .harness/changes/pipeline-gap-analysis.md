# 流水线执行差距分析

> 基于 2026-06-17 内容审核全链路集成任务的全程回溯

## 逐阶段追踪

| # | 阶段 | 规范要求 | 实际执行 | 差距类型 |
|---|------|---------|---------|:---:|
| 0 | 路径选择 | 收到需求首条响应输出路径结论 → 写入 request.md | 直接进入 brainstorming，未做路径选择。后期补写 | 🔴 约束 |
| 1 | 需求分析 | 派发独立 `requirement-analyst` 子 Agent | Owner 内联用 superpowers brainstorming 完成 | 🔴 约束 |
| 2 | 需求评审 | 3 子 Agent 并行（coverage/structure/clarity），2/3 APPROVED | 完全跳过 | 🔴 约束 |
| 3 | 架构设计 | 派发独立 `architecture-designer` 子 Agent，产出含 TDD 步骤的 tasks.md | Owner 内联在 brainstorming 中完成 | 🔴 约束 |
| 4 | Proto 变更 | `make ci`（lint + breaking-check + generate） | 只执行了 `make generate`，跳过了 lint 和 breaking-check | 🟡 约束 |
| 5 | 编码管线 | N×Workflow 并行 `harness-pipeline.js`，Generator→QA→Review 循环 | 使用 superpowers `subagent-driven-development`，无 QA 无 Review，全部串行 | 🔴 约束 |
| 5.1 | TDD | RED→GREEN→REFACTOR，每个新函数先写测试 | 0 个新测试。后来补了 20 个 | 🔴 约束 |
| 5.2 | QA | 15 项机械化检查（含 API 冒烟测试） | 第一次执行完全跳过。后用 ad-hoc Workflow 补做 | 🔴 约束 |
| 5.3 | Review | 3 视角并行评审（correctness/security/robustness） | 第一次跳过。后补做了 2 轮，发现 30+ 问题 | 🔴 约束 |
| 6 | 集成归档 | 全链路编译、smoke test、归档 impl/ | 只手动 build 了 3 个服务 | 🟡 约束 |

## 差距归因

### 🔴 约束不到位（8 项）—— 占了绝大多数

这些阶段**规范本身是正确的**，但执行时被跳过了。原因是：

**1. 缺少硬性门禁检查**

每个阶段都有明确的"门禁"列（如 2/3 APPROVED、QA PASS + Review 2/3 PASS），但这些门禁**只在文档中描述，没有自动化检查**。Agent 可以声称"完成了阶段 5"而实际上根本没启动 Workflow。

```
现状: 门禁 = 文档中的一句话，依赖 Agent 自觉遵守
问题: Agent 可以选择性忽略，无后果
根因: 门禁没有被编码为可执行的检查
```

**2. owner-agent.md 是"指南"而非"合约"**

- `owner-agent.md` 使用自然语言描述流程
- 没有机制验证"你是否按 §4 阶段表执行了？"
- CLAUDE.md 第 7 条（路径选择强制）是在本次复盘后才加的

**3. 外部技能可以无感知地替代内部流程**

- superpowers `brainstorming` → 替代了 Phase 1 需求分析子 Agent
- superpowers `writing-plans` → 替代了 Phase 3 架构设计子 Agent
- superpowers `subagent-driven-development` → 替代了 Phase 5 Workflow

Pipeline 没有检测"你用的不是 harness-pipeline.js"的能力。

**4. Phase 5 禁令是文字，不是代码**

```
owner-agent.md 写了: "❌ 禁止使用 subagent-driven-development 替代 harness-pipeline.js"
实际发生: 用了 subagent-driven-development，没有任何拦截
```

### 🟡 流水线规范可改进（2 项）

**5. Phase 4 Proto 变更的 `make ci` 未严格执行**

规范说 `make ci`（lint+breaking+generate），但 Breaking Check 在本地可能因缺少 `buf breaking` 的 baseline 而失败。这不是 Agent 的问题，是 `make ci` 在本地环境可能不适用。

**建议**: `make ci` 拆分为 `make lint generate`（本地）+ `make breaking-check`（CI only）。

**6. QA 检查 #3 的 0/0 检测存在盲区**

（本次已修复：添加了"最近 7 天新增包无测试"检测）

## 措施建议

### 短期（可立即执行）

| # | 措施 | 类型 |
|---|------|:---:|
| 1 | CLAUDE.md 第 7 条已加 ✅ | 约束 |
| 2 | owner-agent.md Phase 5 禁令已加 ✅ | 约束 |
| 3 | QA #3 新包检测已加 ✅ | 流水线 |
| 4 | QA #15 API 冒烟测试已加 ✅ | 流水线 |

### 中期（需开发）

| # | 措施 | 说明 |
|---|------|------|
| 5 | **门禁检查脚本** | 写一个 `harness-gate-check.sh`，在每个阶段完成后运行，检查：Phase 1 后有 proposal.md？Phase 2 后有 3 份 review？Phase 5 后有 _qa.md + _review.md？输出 PASS/FAIL |
| 6 | **harness-pipeline.js 作为唯一入口** | 在 CLAUDE.md 中声明：涉及编码的阶段 5 只能通过 Workflow 工具 + harness-pipeline.js 启动。在 owner-agent.md 中加"如果检测到使用了 subagent-driven-development 等外部技能，阶段 6 门禁自动 FAIL" |
| 7 | **阶段 0 检查前置** | 收到需求后，在进入任何其他阶段前，先运行 `harness-gate-check.sh --phase 0` 确认 request.md 已创建。FAIL → 阻塞 |

### 长期（架构层面）

| # | 措施 | 说明 |
|---|------|------|
| 8 | **Phase 阶段状态机** | 用一个 JSON/YAML 文件跟踪 `current_phase`，每个阶段完成后更新。下一阶段启动前校验。类似 CI/CD pipeline 的 stage 概念 |
| 9 | **产出物完整性校验** | 阶段 6 自动收集所有产出文件路径，与规范要求的产出清单对比，缺失的标记为 FAIL |
| 10 | **外部技能拦截** | 在 Workflow 或 Agent 调用层面，检测是否使用了非 harness-pipeline.js 的编码执行方式。如果检测到，写入 WARN 到 _qa.md |

## 结论

**8/10 个差距是约束不到位，不是流水线写得有问题。** owner-agent.md 的阶段表本身是正确的，但它是一份"操作手册"而非"可执行合约"。Agent 可以不遵守而无后果。

核心矛盾：**自然语言规范 vs 自动化执行**。用文档告诉 Agent "你应该这样做"不够——必须让"不这样做就会 FAIL"。
