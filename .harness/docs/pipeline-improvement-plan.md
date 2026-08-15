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

### Phase 1 — 收敛闭环（最高优先，直接解决 ~200 万 token 浪费）✅ 已实施（commit e45fe59）

**P1.1 评审反馈注入（D1 根因）** ✅
- 依据：2.1「评审反馈必须闭环」+ 2.2「歧义解析循环」
- 做法：stage2 REVISION 回阶段 1 时，把本轮 mustFixes 结构化拼入 analyst prompt（`## 上轮评审反馈` 段，含 section/issue/fix），要求分析师逐条对照修订并标注「已解决」
- 验证：评审轮次分布收敛（中位数从 3+ 降到 ≤2）；重复发现同类缺口次数归零
- 状态：已实现（`stage1Requirement` 注入 `ctx.stageResults[2].reviewRounds` 末轮 mustFixes）

**P1.2 语义收敛早停（D4）** ✅
- 依据：2.1「audit 无新发现即早停」+ 2.4「语义早停」
- 做法：评审轮次间对比 mustFixes 集合，**连续一轮无新 mustFixes → 判定收敛**（直接 APPROVED 或带剩余 WARNING 前进）；有界 ≤3 轮 + 人工修正 cycle ≤2 次后强制 escalate
- 验证：spec 评审不再靠人肉判断何时停；早停轮次 ≥50% 场景 ≤2 轮
- 状态：已实现（`mustFixKey` 签名跨轮对比，`rounds>=2 && newKeys==0` → 提前 escalate）

**P1.3 新鲜评审隔离 + 缓存确定性失效（D3）** ✅
- 依据：2.1「评审 scratch 在 worktree 外」+ 2.3「评审输入版本化」
- 做法：评审 prompt 嵌入**被审 spec 文件内容哈希**（非 cycle 序号）；文件变则 prompt 变 → resume 缓存必然失效；评审 scratch（mustFixes 收集）存 worktree 外
- 验证：resume 后评审必读最新文件，缓存命中旧结论为 0
- 状态：已实现（`specContentHash` FNV-1a 确定性哈希进评审 prompt）

**P1.4 决策状态即消费即清（D2）** ✅
- 依据：状态机幂等
- 做法：统一 `consumeDecision(ctx, checkpoint)` helper——读取后立即 `delete ctx.decisions[cp]`；主循环以「决策是否已消费」判分支，杜绝重进回环
- 验证：escalate/approve 分支单次触发，回环计数为 0
- 状态：已实现（stage2/3/4/5/6 全部决策消费；`stage1_clarify` 为持久输入不消费）

> 附：P1 顺带固化会话验证的 `manualSpecFix`（Owner 人工修正后跳分析师直接重审）与 manual-fix cycle `rollbackCount` 重置。
> 行为测试：`.harness/workflows/p1-convergence.test.js`（14 项，`node .harness/workflows/p1-convergence.test.js`）。

### Phase 2 — 成本预算护栏（防 D5）✅ 已实施

**P2.1 Token 预算分档 + 模型分级路由** ✅
- 依据：2.4「soft/hard 预算 + 模型降级」
- 做法：管线级 token 预算（soft 150 万 / hard 250 万输出 token，可经 `args.budget` 覆盖）；`spentTokens()` 读 Workflow 运行时 `budget.spent()`；主循环 hard 超限 → 升级人工（Owner 可确认继续）；**模型分级** `routeModel(key)` 经 `args.models` 配置（默认继承会话模型），已挂接 dispatch/clarify/analysis/review/architecture 各 agent
- 验证：同类任务成本下降 ≥40%；昂贵模型调用占比可观测
- 状态：已实现（`PIPELINE_BUDGET` / `budgetLevel()` / `costSummary()` / `routeModel()`）

**P2.2 成本可观测** ✅
- 依据：2.4「不确定用量上报而非假精确」
- 做法：每次 HITL 暂停 `pauseForInput` 的 summary 追加 `[成本护栏: 累计~X万输出token / soft Y万 / hard Z万]`；最终 return 补 `cost` 字段（budgetLevel 结果）
- 验证：交付确认时用户能看到"本次变更消耗 X 万 token / 成本 Y"
- 状态：已实现（pauseForInput + 最终 return）

### Phase 3 — 确定性门禁（D6，把 ground-truth 移进门）✅ 已实施

**P3.1 harness-checks 对齐真实提交门** ✅
- 依据：2.3「CI 回归门」+「独立 SLO 阈值」
- 做法：harness-checks 增 `check_go_fmt`（对变更 git diff HEAD + --cached 的 Go 文件跑 `gofmt -l`，非空即 FAIL，与 pre-commit 同一命令）；错误码魔数检查已内置。负向验证：未格式化文件 → FAIL/EXIT 1
- 验证：harness-checks PASS 后提交 0 被拒；SQL 改动有落库断言（SQL 断言为测试实践，由 QA 人工分诊）
- 状态：已实现（check_go_fmt + labels 对齐 + 头部说明；harness-checks 17 PASS 验证）

**P3.2 Spec 确定性自检（D8）** ✅
- 依据：2.2「四支柱评估」+ 2.3「确定性 evaluator」
- 做法：stage2 评审前跑 `specDeterministicCheck()`（非 LLM）：①追溯表条目数≥决策数 + 全✅ ②错误码登记意图（新码 6000x 必须声明登记，比对其在 proto 头注释已登记集 + spec 语境含登记关键词）③REQ 引用解析（REQ-XXX 引用 → 定义存在）；FAIL 直接回阶段 1（发现作反馈），不耗 LLM 评审，detRounds≥4 升级人工
- 验证：spec 评审 token 降（机械检查不烧 LLM）；自检拦下可机械化判定的缺陷
- 状态：已实现（specDeterministicCheck + stage2 接线 + detRounds 上限）

> 行为测试：`.harness/workflows/p2p3-guards.test.js`（12 项，`node` 运行）。P1 回归 14 项通过。

### Phase 4 — 评估与反馈飞轮 ✅ 已实施

**P4.1 管线自身 eval 语料** ✅
- 依据：2.3「eval 语料复合」+「反馈飞轮」
- 做法：`pipeline/evals/` 语料库——`p1-convergence.eval.js`/`p2p3-guards.eval.js`/`p4p5-evals.eval.js` + `run-evals.sh` 运行器 + README（用例↔实战缺陷映射）；**已接入 harness-checks #20「pipeline evals」**（每次 QA 顺带验证管线自身不回归）
- 验证：管线改动后 `bash .harness/pipeline/evals/run-evals.sh` 防回归（41 项全绿）；harness-checks 18 PASS
- 状态：已实施（3 eval 文件迁移/新增 + 运行器 + 文档 + harness-checks 门禁；eval 夹具自包含，不依赖真实变更目录）

**P4.2 评审发现 → 结构化回填（D9）** ✅
- 依据：2.1「捕获每个反馈信号，左移修复」
- 做法：管线评审把 WARNING（mustFixes/确定性自检发现）写 `.harness/review-feedback/<change>.warnings.jsonl`；memory 建议写 `<service>.memory.jsonl`；`backfill-review-feedback.sh` 把 WARNING 聚合为 backlog task（source=review）+ memory 追加 pending-suggestions.md（处理完归档 processed/，幂等）
- 验证：评审产出不再"报告即死"，进入可执行闭环（已实测建任务 + memory 落盘）
- 状态：已实施（spec-pipeline + harness-pipeline 写入 + 回填脚本）

### Phase 5 — 状态可靠性（D7）✅ 已实施

**P5.1 ctx 版本化 + resume 完整性校验** ✅
- 依据：2.4「checkpoint/resume」
- 做法：`REQUIRED_CTX_FIELDS` 每 stage 必需字段（stage2 需 traceability / stage4 需 protoChanges（D7）/ stage5 需 services 任一）+ `validateResumeState` 在 loadState 后校验，缺失拒绝 resume 并提示完整 ctx
- 验证：resume 缺字段静默跳过（本次 stage4 误跳）不再发生
- 状态：已实现（getPath 支持 a.b[3].c + 校验接线）

**P5.2 状态落盘（fs 可用时）** ✅
- 做法：`saveState` 落盘 `pipeline-state.json`（schema+change 匹配）；`loadState` 优先 args.resumeState（既有机制），未传且 fs 可用时读盘；沙箱无 fs 时保持 ctx 传参 fallback
- 验证：resume 不再依赖手传完整 ctx（fs 环境）；schema 不匹配拒绝盘状态
- 状态：已实现（防御性落盘/读盘，沙箱优雅降级）

> 行为测试：`p4p5-evals.eval.js`（14 项）。全量 `run-evals.sh` 40 项通过。

---

## 四、优先级与投入

| 优先级 | 项 | 投入 | 直接收益 |
|:---:|---|---|---|
| P0 | P1.1 评审反馈注入 / P1.2 收敛早停 / P1.4 决策即消费 | 中 | 消灭盲循环 ~200 万 token 浪费 + 评审收敛 |
| P0 | P3.1 门禁对齐（gofmt/SQL 断言） | 小 | "PASS 即真 PASS"，提交 0 被拒 |
| P1 | P1.3 评审隔离/缓存失效 / P2.1 预算护栏 / P5.1 ctx 校验 | 中 | 状态可靠 + 成本可控 |
| P2 | P2.2 成本上报 / P3.2 spec 自检 / P4.1 eval 语料 / P4.2 发现回填 / P5.2 状态落盘 | 中-大 | 评估飞轮 + 长期自进化 |

> 已建立：BACKLOG task-2026-08-14-002（盲循环）；本方案为完整路线图，建议拆分为独立 dispatch 任务逐项实施（每项走对应分级管线）。

---

## 五、需求分析 Agent 完善（2026-08-15，已实施）

> 来源：rel-user-role-migration-publish-fix（L 级 spec-pipeline 实战）+ 外部意见评审。聚焦**需求分析 agent 质量**，与 P1-P5（编排/门禁/成本）互补。

| # | 改动 | 位置 | 理由 |
|---|------|------|------|
| A1 | 澄清 prompt 加「最小现状核验 + 只问未决点」 | `harness-spec-pipeline.js` stage1a | 澄清 agent 的 grounding 从"靠模型自觉"变为显式要求；任务文本已决项不重复问（本会话 11 问中若干已隐含在任务文本） |
| A2 | 新增 Step 0：变更类型判定（new/modify/delete）+ 冲突预检 + 需求拒收 | `requirement-analysis.md` | 覆盖存量迭代 diff、进行中变更冲突、不可行需求拒收三个真实缺口 |
| A3 | Step 1 补「只问未决点」 | `requirement-analysis.md` | 需求越清晰问题越少，降沟通成本 |
| A4 | Step 3 补「业务/非功能按需加载」 | `requirement-analysis.md` | business-flows / rbac-design 按需读，避免默认全量膨胀上下文 |
| A5 | Step 5 proposal 头加 优先级/规模/风险/变更类型 元数据 | `requirement-analysis.md` | Owner 排期免读全文 |
| A6 | Step 7 .change.yaml 标准字段（priority/change_type 等） | `requirement-analysis.md` | 字段规范化，供阶段 4/5/6 与 P4.2 消费 |
| A7 | Step 8 Self-Review 加「合规性」第 6 项 | `requirement-analysis.md` | 补齐权限/安全/非功能自检 |
| A8 | Step 2 改为「产出决策日志」，转换追溯移到 Step 8 闭环（正向防丢失 + 反向防幻觉） | `requirement-analysis.md` + subagent | 修正 skill 文本时序错误——追溯本应在 spec 产出后，而非产出前 |
| A9 | Step 1 澄清角色分工：agent 产出问题清单，Owner 用 AskUserQuestion 收敛（≤4 问/轮） | `requirement-analysis.md` | 修正文本误导——子 agent 一次性运行，不直接多轮交互 |
| A10 | Step 6 spec 模板强制稳定 ID `REQ-<capability>-<序号>` | `requirement-analysis.md` | 下游 architect/developer trace 的地基 |
| A11 | Step 8 加 Definition of Done（7 条硬性清单）+「自查非终审」声明 | `requirement-analysis.md` | 交付有明确完成判据，且不误导自查=评审 |

**未做（另立任务）**：需求冲突检测**确定性脚本**（现为 Skill 步骤，机械化版本见 BACKLOG task-2026-08-15-003）；P3 级小缺口（受影响服务定位 / graph-context 用途边界 / 历史变更参考 / 隐私合规段落）记 backlog 待办，未立即改。
**验证**：`node --check` 语法过；`run-evals.sh` 41 项全绿（workflow 改动无回归）；skill 文档 Step 0-8 + 转换追溯 + DoD 结构连贯、无残留旧措辞。
