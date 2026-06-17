# 内容审核全链路集成 — 流水线复盘

> 2026-06-17，对比 Harness 规范（owner-agent.md §4 阶段表）与实际执行

## 一、逐阶段对比

| # | 阶段 | Harness 规范要求 | 实际执行 | 偏差 |
|---|------|-----------------|---------|:---:|
| 0 | **路径选择** | 首条响应输出路径结论（OpenSpec/DevAgent/Edit）→ 写入 request.md | 直接进入 brainstorming，未做路径选择 | 🔴 完全跳过 |
| 1 | **需求分析** | 派发独立 `requirement-analyst` 子 Agent，产出 proposal.md + specs | Owner 内联做了 brainstorming（superpowers 技能），产出 design.md | 🔴 绕过了项目子 Agent |
| 2 | **需求评审** | 3 子 Agent 并行（coverage/structure/clarity），2/3 APPROVED 才通过 | 完全跳过 | 🔴 完全跳过 |
| 3 | **架构设计** | 派发独立 `architecture-designer` 子 Agent | Owner 内联在 brainstorming 中完成 | 🔴 绕过了项目子 Agent |
| 4 | **Proto 变更** | Owner 亲自执行 + `make ci`（lint+breaking+generate） | ✅ Owner 亲自执行 | 🟢 符合 |
| 5 | **编码+测试** | `Workflow({scriptPath:".harness/workflows/harness-pipeline.js"})` → Generator→QA→Debug→Reviewer 循环，最多 3 轮。每服务独立并行。 | 使用 superpowers `subagent-driven-development`，直接 dispatch implementer，**跳过了 QA、Review 门禁**，全部串行 | 🔴 核心偏差 |
| 6 | **集成归档** | 全链路 build+test、harness-smoke.sh、mv QA/Review 到 impl/、更新 INDEX | 手动 build 3 个服务，未执行 smoke test，未归档 | 🟡 部分执行 |

## 二、阶段 5（编码）是最严重的偏差

### Harness 规范要求的 Generator→QA→Review 循环：

```
Generator(写代码) → QA(11项机械化检查) 
  ├── QA PASS → Reviewer(3视角并行评审)
  │              ├── 2/3 PASS → ✅ 完成
  │              └── <2/3 → 回到 Generator 修复
  └── QA FAIL → Debug(根因分析) → 回到 Generator 修复
  (最多 3 轮)
```

### 实际执行：

```
Implementer(写代码) → ✅ 直接标记完成  ← 没有 QA，没有 Review
```

结果：4 个低级问题（Vue template `as` 语法、`.d.ts` 运行时值、旧进程未杀、go-zero 指针切片）全部直达用户。

## 三、出问题的根本原因（3 层）

### 第 1 层：Owner 没有做路径选择（阶段 0）

收到需求后，应该按 owner-agent.md 判定：跨 4 服务 + Proto 变更 → **OpenSpec 路径**。实际直接用了 superpowers 的 brainstorming 技能，跳过了整个 Harness 阶段表。

### 第 2 层：用错了编码执行引擎

Harness 要求用 `Workflow({scriptPath:".harness/workflows/harness-pipeline.js"})`，这个 Workflow 脚本内置了 QA 机械化检查 + 3 视角并行 Reviewer。实际用了 superpowers 的 `subagent-driven-development`，这个技能虽然也有 Review 步骤，但：
- 它是通用技能，不了解项目的 11 项 QA 检查
- 它的 Review 是"建议"而非"硬门禁"
- 它的 implementer prompt 不注入项目记忆（memory/）

### 第 3 层：即使用了 subagent-driven-development，也跳过了它的 Review 步骤

`subagent-driven-development` 技能明确要求每个 task 走三步：
```
Implementer → Spec Reviewer → Code Quality Reviewer
```
实际只执行了第一步，后两步被"自行优化"掉了。

## 四、流水线需要完善的 5 个点

### 1. Owner Agent 必须强制执行阶段 0 路径选择（高优先级）

**当前问题**：CLAUDE.md 和 owner-agent.md 都要求路径选择，但没有机制强制执行。Claude 可以不输出路径选择就直接动手。

**建议**：在 CLAUDE.md 的硬性约束中加一条：
```
| 7 | 收到需求后首条响应必须输出路径选择（路径+理由+涉及服务），否则不可进入下一阶段 | owner-agent.md §路径选择 |
```

### 2. Harness Pipeline 的入口应该是 Workflow 而非 subagent-driven-development（高优先级）

**当前问题**：superpowers 的 brainstorming/writing-plans/subagent-driven-development 是通用技能，不了解项目的 QA 检查、Memory 注入、服务边界。

**建议**：
- 涉及多服务开发时，阶段 5 必须使用 `Workflow({scriptPath:".harness/workflows/harness-pipeline.js"})`
- 在 owner-agent.md 中加入："禁止用外部技能替代 Harness Pipeline 的核心循环"
- 如果要用 superpowers 技能辅助（如 brainstorming），应作为 Harness 阶段内的子步骤，而非替代整个流程

### 3. 编码阶段必须包含 E2E 冒烟测试（高优先级）

**当前问题**：阶段 5 的 QA 只做静态检查（build/vet/jstype/等），不做运行时冒烟。导致"新路由注册了但旧进程占端口"这类问题漏过。

**建议**：在 `harness-pipeline.js` 的 QA 步骤中增加第 12 项检查：
```
[12/12] API smoke test — curl 每个新注册的端点，验证返回非 404
```
或利用已有的 `harness-smoke.sh`（阶段 6），将服务重启+端点验证作为阶段 5 的最后一个 QA 步骤。

### 4. Reviewer 角色应该用独立的、对抗性的 prompt（中优先级）

**当前问题**：后续补上的 Spec Reviewer 和 Code Quality Reviewer 效果很好（发现了 3 BLOCKER + 3 CRITICAL + 7 HIGH），但它们是用通用 Explore agent 跑的。Harness 的 reviewer 有 3 个独立视角（correctness/security/performance）。

**建议**：
- Harness pipeline 的 Reviewer 使用 `.harness/agents/prompts/review.js` 中定义的 REVIEW_LENSES
- 3 个 Reviewer 应该并行运行，各自从不同视角审查
- 每个 Reviewer 的 prompt 应该注入对应服务的 memory/ 目录中的经验

### 5. 流水线失败应该自动回退而非手动修复（低优先级）

**当前问题**：Review 发现问题后，是我手动 dispatch fix agent 修复的。Harness 规范要求 Debug→Generator 自动循环。

**建议**：
- 确保 `harness-pipeline.js` 的 orchestration loop 正常工作
- QA FAIL → 自动触发 Debug agent（根因分析）→ Generator agent（修复）
- Review FAIL → 自动回到 Generator 修复
- 只在超过 3 轮时升级给人

## 五、总结

```
本次实际执行路径:
  brainstorming(inline) → writing-plans(inline) → subagent-driven(无Review) → 手动修复

Harness 规范路径:
  阶段0(路径选择) → 阶段1(需求分析子Agent) → 阶段2(3人评审) → 阶段3(架构设计子Agent)
  → 阶段4(Proto变更/Owner) → 阶段5(Workflow×4并行,每服务QA+Review)
  → 阶段6(冒烟测试+归档)
```

差距不是某一个步骤的问题，而是**整个流程框架被替换了**。下次跨服务需求开发，应该走 Harness 规范的 OpenSpec 路径，用 `harness-pipeline.js` Workflow 作为编码引擎。
