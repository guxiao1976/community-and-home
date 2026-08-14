# 开发管线完善方案（Pipeline Improvement Plan）

> 来源：2026-08-14 role-platforms-save（L 级 spec-pipeline 实战检验）+ 行业最佳实践调研。
> 目标：把「评审盲循环、无收敛预算、无成本护栏、门禁与真实提交门不一致、状态脆弱」等实战暴露的缺陷，改造成有闭合反馈环、有收敛判据、有预算护栏、有确定性门禁的管线。

---

## 一、实战缺陷盘点（role-platforms-save 实证）

| # | 缺陷 | 实战证据 | 后果 |
|---|------|---------|------|
| D1 | **需求评审盲循环**：stage2 REVISION 回阶段 1 时分析师用相同 prompt 重跑，mustFixes **未注入** | 3 轮评审逐字相同 prompt，同类缺口（base-check 审计范围）每轮被不同评审重复发现 | 3 轮不收敛 → 人工 escalate；该小任务 spec 评审烧 **~200 万 token**，全管线 ~**300 万 token / 数小时** |
| D2 | **决策状态未消费即回环**：stage2_done/stage2_escalate 处理后未 `delete` 决策，主循环重进 stage2 再次触发 | 实测 escalate 分支回环，靠 rollbackCount 上限兜底而非收敛 | 死循环/误判全局 escalate；本次靠手工补丁修复 |
| D3 | **resume 缓存毒化**：评审 Agent 以 (prompt, opts) 为键缓存，spec 文件变了但 prompt 相同 → 返回**旧 REVISION** | 同一评审 cycle 第 1 轮 prompt 与上轮 cycle 相同 → 缓存命中旧结论 | 修复可能被"伪重审"吞掉；本次靠手工加 cycle 标记规避 |
| D4 | **无收敛判据 / 无预算**：manual-fix cycle 各自重置 rounds，rollbackCount 累计到全局上限才停 | 3 轮人工修正 cycle 才收敛；无"findings 增量归零即收敛"的退出判据 | 无法自动判断"已收敛"，靠人肉判断 |
| D5 | **无 token 预算护栏**：无 soft/hard 预算、无模型分级降级、无成本告警 | 全管线 ~300 万 token 无任何预警 | 成本不可控，昂贵模型用于机械环节 |
| D6 | **门禁与真实提交门不一致**：harness-checks 未查 gofmt → PASS，pre-commit 却拒 | QA 16 PASS 后提交被 gofmt 拒绝（4 文件）；model Update SQL 改动无 SQL 断言（评审指出） | "通过"是假象；门禁缺确定性锚点 |
| D7 | **resume 状态脆弱**：ctx 经 args.resumeState 手传，缺完整性校验 | 本次精简 ctx 致 stage4 proto 变更被误判"无 proto 变更"跳过 | 状态丢失静默改变流程 |
| D8 | **"模型评判模型"**：spec 评审全部是 agent 对 agent，无确定性 ground-truth | 评审依赖 3 个 LLM 视角共识，无结构化自检兜底 | 评审质量受 LLM 漂移影响 |

## 二、行业最佳实践调研（含来源）

### 2.1 收敛性（Convergence）
- **有界轮次 + 新鲜评审退出门**：plan→review→build 循环必须有界（≤3 轮），**评审 scratch 放在 worktree 之外**防前轮结论泄漏给新评审（[Convergo](https://github.com/gomilesf/convergo)）
- **评审反馈必须闭环**：gap-finding 收敛循环（implement→audit→find gaps→re-implement），**audit 无新发现即早停**（[Spec Kit `/spec.converge`](https://github.com/tikalk/agentic-sdlc-spec-kit)、[specd](https://github.com/nhalm/specd)）
- **模型不能自判"通过"**：测试/确定性工具才是 ground-truth；评审通过仍需人工 merge 门（[LangGraph 示例](https://developers.openai.com/cookbook/examples/agents_sdk/agent_improvement_loop)、[Convergo](https://github.com/gomilesf/convergo)）
- **评审视角差异化 + 否决权**：正确性/回归风险视角应持否决权重，而非等权多数（[Adversarial judge panels](https://developers.openai.com/cookbook/examples/agents_sdk/agent_improvement_loop)）

### 2.2 Spec 驱动（SDD）
- **显式批准门**：spec 未批准不进实现；歧义解析循环（self-critique + 用户反馈）（[OpenSpec](https://www.thoughtworks.com/zh-cn/radar/tools/openspec)、[DeepSpec](https://github.com/godrix/spec.md)）
- **四支柱评估**：Spec Compliance / Code Quality / Test Adequacy / Risk & Evidence，各 ≥70 才 VERIFIED（[Spec Kit](https://github.com/tikalk/agentic-sdlc-spec-kit)）
- **Delta-spec 回填**：增量 spec 而非一次性完整 spec，适配既有系统（[OpenSpec](https://www.thoughtworks.com/zh-cn/radar/tools/openspec)）

### 2.3 评估驱动 / 质量门禁（EDD）
- **Evals 作为可执行测试**：golden dataset + LLM-as-Judge（多数共识 + 定期校准）+ 确定性 evaluator（[LangChain Evals](https://www.langchain.com/resources/how-to-evaluate-llms)、[Practical LLM Evaluation](https://www.packtpub.com/en-tw/product/practical-llm-evaluation-for-production-systems-9781807423889)）
- **CI 回归门**：每 PR 跑 eval 套件，metric 下降超阈值即 block merge（[Eval-First AI](https://botscrew.com/blog/eval-first-ai-llm-evaluation-guide/)）
- **独立 SLO 阈值**：不要一个聚合分（会掩盖 40% 失败）；分维度门禁（task 完成率/工具成功率/恢复率/p99/护栏触发率）（[Practical LLM Evaluation](https://store.accuristech.com/standards/practical-llm-evaluation-for-production-systems-1st-edition)）

### 2.4 成本 / 预算 / 早停
- **Token 预算分档**：soft(80%)/hard(100%)，超限降级模型或拒绝（[tokenfence](https://www.npmjs.com/package/tokenfence)、[budget-gatekeeper](https://lobehub.com/skills/kennedym-ds-copilot_orchestrator-budget-gatekeeper)）
- **语义早停**：草稿嵌入余弦距离 + 耐心窗口，停更即停；**judge 型早停反效果**（评判本身耗 token）（[Semantic Early-Stopping](https://arxiv.org/abs/2606.27009)）
- **checkpoint/resume 到失败步**（非从 step1）（[animus-core](https://libraries.io/pypi/animus-core)）

---

## 三、完善方案（分阶段，每项含「依据 → 做法 → 验证」）

### Phase 1 — 收敛闭环（最高优先，直接解决 ~200 万 token 浪费）

**P1.1 评审反馈注入（D1 根因）**
- 依据：2.1「评审反馈必须闭环」+ 2.2「歧义解析循环」
- 做法：stage2 REVISION 回阶段 1 时，把本轮 mustFixes 结构化拼入 analyst prompt（`## 上轮评审反馈` 段，含 section/issue/fix），要求分析师逐条对照修订并标注「已解决」
- 验证：评审轮次分布收敛（中位数从 3+ 降到 ≤2）；重复发现同类缺口次数归零

**P1.2 语义收敛早停（D4）**
- 依据：2.1「audit 无新发现即早停」+ 2.4「语义早停」
- 做法：评审轮次间对比 mustFixes 集合，**连续一轮无新 mustFixes → 判定收敛**（直接 APPROVED 或带剩余 WARNING 前进）；有界 ≤3 轮 + 人工修正 cycle ≤2 次后强制 escalate
- 验证：spec 评审不再靠人肉判断何时停；早停轮次 ≥50% 场景 ≤2 轮

**P1.3 新鲜评审隔离 + 缓存确定性失效（D3）**
- 依据：2.1「评审 scratch 在 worktree 外」+ 2.3「评审输入版本化」
- 做法：评审 prompt 嵌入**被审 spec 文件内容哈希**（非 cycle 序号）；文件变则 prompt 变 → resume 缓存必然失效；评审 scratch（mustFixes 收集）存 worktree 外
- 验证：resume 后评审必读最新文件，缓存命中旧结论为 0

**P1.4 决策状态即消费即清（D2）**
- 依据：状态机幂等
- 做法：统一 `consumeDecision(ctx, checkpoint)` helper——读取后立即 `delete ctx.decisions[cp]`；主循环以「决策是否已消费」判分支，杜绝重进回环
- 验证：escalate/approve 分支单次触发，回环计数为 0

### Phase 2 — 成本预算护栏（防 D5）

**P2.1 Token 预算分档 + 模型分级路由**
- 依据：2.4「soft/hard 预算 + 模型降级」
- 做法：管线级 token 预算（soft 80% / hard 100%）；**模型分级**——机械环节（格式/编译/字段映射）用低价模型，评审/收敛/门禁裁决用高能力模型；超 soft 降级，超 hard 强制 escalate
- 验证：同类任务成本下降 ≥40%；昂贵模型调用占比可观测

**P2.2 成本可观测**
- 依据：2.4「不确定用量上报而非假精确」
- 做法：METRIC 行补 token/cost 字段（已有点位）；每次 HITL 暂停点报告累计成本 + 预估剩余
- 验证：交付确认时用户能看到"本次变更消耗 X 万 token / 成本 Y"

### Phase 3 — 确定性门禁（D6，把 ground-truth 移进门）

**P3.1 harness-checks 对齐真实提交门**
- 依据：2.3「CI 回归门」+「独立 SLO 阈值」
- 做法：harness-checks 增 `gofmt/gofumpt` 检查（与 pre-commit 同一命令）；增 model 层 SQL 断言（`ExpectExec update sys_role`）覆盖字段映射类改动；错误码魔数检查已内置
- 验证：harness-checks PASS 后提交 0 被拒；SQL 改动有落库断言

**P3.2 Spec 确定性自检（D8）**
- 依据：2.2「四支柱评估」+ 2.3「确定性 evaluator」
- 做法：spec 评审前先跑**确定性自检**（非 LLM）：追溯表全✅、跨 spec 引用解析、错误码登记一致、字段号无冲突、mustFixes 引用已解决；自检 FAIL 直接 REVISION，不耗 LLM 评审
- 验证：spec 评审 token 降（机械检查不烧 LLM）；自检拦下可机械化判定的缺陷

### Phase 4 — 评估与反馈飞轮

**P4.1 管线自身 eval 语料**
- 依据：2.3「eval 语料复合」+「反馈飞轮」
- 做法：把本 session 的摩擦案例（盲循环、决策回环、缓存毒化、gofmt 门不一致、ctx 脆弱）固化为 `pipeline/evals/` 回归用例；管线改动后跑 eval 防回归
- 验证：管线改动不破坏已修复行为（如回归断言"评审 mustFixes 必须注入"）

**P4.2 评审发现 → 结构化回填（D9）**
- 依据：2.1「捕获每个反馈信号，左移修复」
- 做法：评审 WARNING/memory-suggestion 自动路由：WARNING → 对应服务的 backlog task（source=review）；memory-suggestion → 记忆系统；CRITICAL → 阻塞
- 验证：评审产出不再"报告即死"，进入可执行闭环

### Phase 5 — 状态可靠性（D7）

**P5.1 ctx 版本化 + resume 完整性校验**
- 依据：2.4「checkpoint/resume」
- 做法：ctx 增 `schema` 版本（已有 1）；resume 时校验**当前 stage 的必需字段**（如 stage3 后必需 protoChanges），缺失则拒绝 resume 并提示完整 ctx
- 验证：resume 缺字段静默跳过（本次 stage4 误跳）不再发生

**P5.2 状态落盘（fs 可用时）**
- 做法：非沙箱环境把 pipeline-state.json 落盘，resume 优先读盘，ctx 传参仅作沙箱 fallback
- 验证：resume 不再依赖手传完整 ctx

---

## 四、优先级与投入

| 优先级 | 项 | 投入 | 直接收益 |
|:---:|---|---|---|
| P0 | P1.1 评审反馈注入 / P1.2 收敛早停 / P1.4 决策即消费 | 中 | 消灭盲循环 ~200 万 token 浪费 + 评审收敛 |
| P0 | P3.1 门禁对齐（gofmt/SQL 断言） | 小 | "PASS 即真 PASS"，提交 0 被拒 |
| P1 | P1.3 评审隔离/缓存失效 / P2.1 预算护栏 / P5.1 ctx 校验 | 中 | 状态可靠 + 成本可控 |
| P2 | P2.2 成本上报 / P3.2 spec 自检 / P4.1 eval 语料 / P4.2 发现回填 / P5.2 状态落盘 | 中-大 | 评估飞轮 + 长期自进化 |

> 已建立：BACKLOG task-2026-08-14-002（盲循环）；本方案为完整路线图，建议拆分为独立 dispatch 任务逐项实施（每项走对应分级管线）。
