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
| A12 | Step 3 + subagent 清单精准性修正：删 graph-context（HOW 层）/ design.md 只读数据模型+业务流程章节 / 补「先定位服务再加载」/ 两处清单同步 | `requirement-analysis.md` + `requirement-analyst.md` | 贯彻「WHAT 不 HOW」到加载清单，必要的不缺、HOW 层不碰（graph-context 留架构/编码阶段） |

**未做（另立任务）**：需求冲突检测**确定性脚本**（现为 Skill 步骤，机械化版本见 BACKLOG task-2026-08-15-003）；P3 级小缺口（历史变更参考 / 隐私合规段落）记 backlog 待办，未立即改。
**验证**：`node --check` 语法过；`run-evals.sh` 41 项全绿（workflow 改动无回归）；skill 文档 Step 0-8 + 转换追溯 + DoD 结构连贯、无残留旧措辞；harness 自检 8 PASS。

---

## 六、Review Skill 完善（2026-08-15，第一批：文档精度）

> 来源：两轮 review 意见评审。聚焦**计划评审 + 执行评审的文档精度**（权限矛盾、分级混乱、检查清单泛化），纯改 `review.md`，不涉及 spec-pipeline.js 的流程层（validity 视角 / 设计评审阶段为第二、三批，另议）。

| # | 改动 | 位置 |
|---|------|------|
| R1 | 权限边界重写：只读 + 仅允许写评审报告，严禁改被审对象 | `review.md` 角色 |
| R2 | 分级统一：MUST FIX↔CRITICAL / SHOULD FIX↔WARNING / INFO↔NOTE 三档对照表 | `review.md` 分级统一 |
| R3 | 9 维度→12 维度：补二级检查项 + 新增可观测性/依赖变更/配置变更 | `review.md` Step 3 |
| R4 | Migration + 跨服务兼容性 + 废弃删除三类专项检查清单 | `review.md` 特殊规则 |
| R5 | M3 打分召回（knowledge-load.sh）+ 冲突优先级 + M4 由 Owner 落库 | `review.md` Step 4 |
| R6 | 报告自检（定位文件+行号 / 建议可落地 / 维度逐项标注） | `review.md` |
| R7 | 增量复审范围 + 强制回归（公共函数/核心链路扩调用方） | `review.md` VERDICT |
| R8 | 任务粒度刚性规则（不跨服务 / Proto+Migration 独立 / 测试 1-10） | `review.md` 模式一 |
| R9 | 计划评审预加载 CLAUDE.md + design.md（structure 判断基准） | `review.md` 模式一输入 |
| R10 | 工具熔断（复用需求分析 Agent 的 2 次失败熔断 + 空结果/格式不符） | `review.md` |
| R11 | QA 联动（加载 QA 报告 / 审未覆盖分支 / 覆盖率<80% WARNING / 不重跑构建） | `review.md` Step 2 |
| R12 | 服务名映射标注「以 registry/services.json 为准」 | `review.md` |

### 第二三批（2026-08-15，流程层 + 架构层）

| # | 改动 | 位置 |
|---|------|------|
| R13 | 加 validity 视角（4 视角并行，+33% 评审 token） | `harness-spec-pipeline.js` lenses + prompt + `review.md` |
| R14 | CRITICAL 一票否决（mustFixes 加 severity=critical）+ stage2_done 呈现少数 MUST FIX | `harness-spec-pipeline.js` 投票逻辑 |
| R15 | 2/3 阈值动态化：写死 `>=2` → `ceil(totalReviews*2/3)`（3视角=2，4视角=3） | `gate-engine.js` + `builtinGate` |
| R16 | 阶段 2 去 tasks.md（只审 specs）+ 阶段 3 加「设计评审」循环（architect→design-review→REVISION 回 architect，≤3 轮） | `review.md`（模式一.5）+ `harness-spec-pipeline.js` stage3 |

**验证**：`node --check` 语法过；`run-evals.sh` 41 项全绿；harness 自检 8 PASS；无「3 视角」残留。设计评审复用了 stage2 的 SPEC_REVIEW_SCHEMA，REVISION 反馈注入 architect prompt（同 P1.1 收敛闭环）。

| R17 | 分层统一：评审补 subagent 层（新建 `reviewer.md`），review.md 瘦身（角色/服务名映射/上下文清单/工具熔断抽到 subagent），workflow 三处引用统一为「先 Read subagent → subagent 内部 Skill()」 | `reviewer.md`（新建）+ `review.md` + `harness-spec-pipeline.js` |

**分层统一说明**：明确「subagent=角色层（我是谁/权限/上下文加载/服务名映射/熔断/交接），skill=流程层（SOP/视角/报告格式/分级）」。需求分析（requirement-analyst.md + requirement-analysis.md）、架构设计（architecture-designer.md + architect-design.md）、评审（reviewer.md + review.md）三环节结构对齐；workflow 三处 agent 入口统一引 subagent。

---

## 七、架构设计 Skill 完善（2026-08-15）

> 来源：架构设计意见评审。聚焦设计正确性（追溯链断裂、加载清单漂移、模板颗粒度），改 `architect-design.md` + `architecture-designer.md`。

### 必须修（6 条）

| # | 改动 | 位置 |
|---|------|------|
| AD1 | Step 1 加载清单补 graph-context.md + request.md（对齐 subagent 7 项，消除漂移） | Step 1 |
| AD2 | design.md 加「需求追溯矩阵」（spec→design 双向绑定，防遗漏/蔓延） | Step 3 模板 |
| AD3 | 新增 Step 0 输入门禁（校验 proposal/specs 齐备 + 无占位符，否则 RETURN_TO_OWNER） | Step 0 |
| AD4 | Proto 变更表加「破坏性(是/否)」列 + 影响评估段（对齐规则 7） | Step 3 模板 |
| AD5 | 数据模型补 deleted_at（软删除，对齐编码规范 §5.1）+ 字段约束补充 | Step 3 模板 |
| AD6 | 接口契约补「鉴权/幂等/性能约束/错误码语义」 | Step 3 模板 |

### 建议修（9 条）

| # | 改动 | 位置 |
|---|------|------|
| AD7 | 任务粒度量化（不跨层级/文件≤3/Proto+Migration 独立/测试≤10）替代主观「2-5 分钟」 | 关键规则 3 |
| AD8 | TDD 覆盖场景（1 正常+2 异常+错误码），前端 TDD 可选 | 关键规则 4 |
| AD9 | Step 3.5 Design Self-Review（审 design 本身：追溯/归属/规范/非功能/破坏性/记忆） | Step 3.5 |
| AD10 | 业务流程补「异常/失败路径 + 跨服务一致性」 | Step 3 模板 |
| AD11 | 记忆 slug 注入时自验存在 + 不适用项记录排除理由 | Step 1.5 |
| AD12 | 服务归属存疑标注「待定+候选+推荐」 | Step 2 |
| AD13 | fix_plan 触发条件写清（仅 Ralph 模式） | Step 6 |
| AD14 | 非功能设计精简 checklist（可靠性/性能/可观测性） | Step 3 模板 |
| AD15 | 关键规则 11 跨服务一致性（走接口禁直连 DB + 一致性方案 + 补偿任务） | 关键规则 |

**同步**：`architecture-designer.md` Step 摘要更新（Step 0 / Step 3.5 新增）。
**验证**：harness 自检 8 PASS；architect-design.md Step 0-6 结构完整。注：意见 P1.2（design 正确性真空区）已由上轮「设计评审」循环解决，本批不再重复。

---

## 八、执行评审 source of truth 收敛（2026-08-15）

> 来源：review 体系盘点（读执行代码 + 真实产物）。发现「文档 review.md 已 12 维度，但执行评审 prompt 源 review.js 还停在 9 维度」的跨文件漂移——根因是 review.md（文档）与 review.js（执行源）两条线靠 build-pipeline.sh 手动编译同步，改文档不同步源。

| # | 改动 | 位置 |
|---|------|------|
| ST1 | review.js 9→12 维度：可观测性#9→standards-eng、依赖变更#10+配置变更#11→security-arch、记忆遵守#9→#12 | `review.js` REVIEW_LENSES + 分工表 |
| ST2 | 重新编译 harness-pipeline.js（build-pipeline.sh，1006 行） | `harness-pipeline.js` |
| ST3 | 删 review-schema.js 死代码 + loader 死 fallback（loadPrompt 永不返回 undefined，`|| require(schema)` 是死分支） | `harness-pipeline-loader.js` + 删 `review-schema.js` |

**根因**：执行评审的真实 prompt 在 `agents/prompts/review.js`（源）+ `harness-pipeline.js`（AUTO-GENERATED 编译版），`review.md` 只是文档。前几轮改 review.md 没同步 review.js → 漂移。
**验证**：`node --check` 三文件 OK；`run-evals.sh` 全绿；harness 自检 8 PASS；编译后 harness-pipeline.js 已 12 维度（:354/:361/:380）。
**第 3 步（设计评审落盘 + 拆视角）已完成**：设计评审单 agent → 2 视角并行（data-model / interface-proto），落盘 `design_review_{lens}_v{round}.md`（SHOULD FIX/INFO 不再丢失）；review.md 模式一.5 补「视角分工」表。
**遗留（待后续批）**：validity 视角真实跑测（找新 change 跑 4 视角验证）；qa-schema.js 同款死代码（loader.js:88 的 fallback 也是死分支，未处理，待确认后一并清理）。

---

## 九、真实 L 级变更暴露的 3 个流程 gap（2026-08-15，已修复）

> 来源：rel-user-role-migration-publish-fix（真实 L 级 spec-pipeline 全流程跑测，阶段 0-6）。这是「测试流水线」的核心产出——不是流水线跑通了没 bug，而是跑通了但暴露了 3 个真实 gap。

| # | gap | 修复 |
|---|-----|------|
| G1 | 评审报告/归档落盘在沙箱失效（spec-pipeline fs=null） | 评审报告（spec_review/design_review）改由评审 agent 用 Write 工具落盘（agent 在完整环境）；summary/INDEX 归档加 fs 失效提示（Owner 手动补） |
| G2 | harness-pipeline 执行评审无 CRITICAL 一票否决（design-biz 报 snake/camel CRITICAL 被 2/3 放行） | harness-pipeline-core.js 投票逻辑加 `criticalCount > 0` 一票否决（对齐需求评审） |
| G3 | spec 把「MySQL 迁移验证」定义成编码任务，Go 管线无法执行 | architect-design.md 关键规则 12：迁移/运维验证分类，标注「Owner 运维验证」不走 harness-pipeline |

**验证**：`node --check` 两文件 OK；`run-evals.sh` 全绿；harness 自检 8 PASS；harness-pipeline.js 重新编译（1010 行，含 CRITICAL 一票否决）。

---

## 十、需求冲突检测确定性脚本（task-003，2026-08-15 已实施）

> 来源：需求分析 agent 完善评审。requirement-analysis.md Step 0 冲突预检原为 LLM 判定，补确定性脚本机械化（非 LLM、便宜、客观），对齐 P3.2 specDeterministicCheck 模式。

| # | 改动 | 位置 |
|---|------|------|
| CC1 | 新建 `check-change-conflict.sh`：扫描 `.harness/changes/*/.change.yaml` 的 services（C1 同服务重叠）+ revises（C2 同接口/文件重叠），两两比对输出冲突预警；支持 `--change <name>` 单变更检测、`--json` | 新建脚本 |
| CC2 | 接入 `harness-self-check.sh`：refs 清单 + check_change_conflict（语法/可执行校验），自检 8→9 PASS | harness-self-check.sh |
| CC3 | requirement-analysis.md Step 0 冲突预检：优先跑确定性脚本，命中再核对 | requirement-analysis.md |

**验证**：脚本 `bash -n` 语法 OK；实测扫描 16 变更输出 17 项冲突（C1 服务重叠为主，C2 精确到真实文件）；`--change` 定向检测生效；harness 自检 9 PASS。
**注**：C1 服务重叠对历史变更较多（共享 permission-service 等），属「需澄清」级预警而非一定冲突；C2 已过滤 `§x.x` 章节引用、归一 revises 到路径首 token。

---

## 十一、设计一致性三级检查机制（2026-08-15，已实施）

> 来源：今日开发中多处发现「代码与设计/规范不一致」（user-service design.md 过期、rel_user_role 三列缺迁移、review.md vs review.js 漂移）。补齐「设计文档 ↔ 代码实现」这条此前缺失的一致性线。

| 级 | 机制 | 位置 |
|---|------|------|
| 第一级 | `check-design-consistency.sh`：比对 Go model db tag 列 vs 标准迁移源（migration/scripts/docs-specs）列覆盖，WARN 报疑似缺列/建表源不在标准位置 | 新建脚本 |
| 第一级 | 接入 `harness-checks.sh` check_design_consistency（提交时门禁，WARN 非 FAIL，labels 数组同步） | harness-checks.sh |
| 第二级 | `--all` 全服务体检 + `--json` 结构化输出 + `--backlog` 自动登记过期项 | 脚本参数 |
| 第三级 | architect-design.md 关键规则 13：改 model/Migration 需同步 design.md（源头消除漂移） | architect-design.md |

**关键设计决策（克制）**：
- **WARN 非 FAIL**：model 列缺失可能是历史手工迁移/legacy/建表源在别处，WARN 提示风险而非误伤提交
- **联表 ViewModel 别名黑名单**（ur_status/role_status）：JOIN 别名非真实列，消除误报
- **SIGPIPE 坑**：`grep -q` + 管道 + `set -o pipefail` 下会误判缺失（列集每次不同），改用 heredoc `<<<` 无管道
- **labels 数组同步**：harness-checks 汇总 label 硬编码，新增检查项需同步

**验证**：脚本连跑 5 次稳定（permission-service 只报真差异 deleted_at）；harness-checks 独立显示 `[WARN] 18. design consistency`；harness 自检 9 PASS。
**实测发现**（脚本价值）：permission-service 的 `deleted_at`、master-data 全表列未覆盖标准迁移源（建表源不在全局 docs/specs）——正是今天那类不一致，现在可机械捕获。
**历史清理（2026-08-15）**：quality-check 全量体检 + `check-design-consistency.sh --all --backlog` 批量登记 5 个服务的设计一致性欠账（auth-service 1 列 / master-data 34 列 / moderation 30 列 / permission 1 列 / user 1 列）为 P2 debt 任务（task-011~015）。同时发现并修复 quality-check skill 的 `--all` 误写（check-change-conflict.sh 无此参数，默认即全扫）。

---

## 十二、日志埋点增强（task-010，2026-08-15 已实施）

> 来源：方案 B 分析结论——现有 usage 日志只能支撑"脚本调用频率"，缺"skill 调用记录"和"checks 结果明细"，不足以判断脚本有效性。加最小 2 处埋点，数据积累后做脚本有效性分析。

| # | 埋点 | 位置 | 记录内容 |
|---|------|------|---------|
| E1 | harness-checks 打点追加 `failed_checks` | harness-checks.sh 打点处 | 本次 FAIL/WARN 的具体检查项名（如 `proto_ts_align;design_consistency;git_hygiene;`）——回答"哪些检查项常 FAIL/WARN、是否有价值" |
| E2 | workflow agent 调用记 skill 使用 | harness-pipeline-core.js（logAgentUsage helper + 4 处 agent 调用） | `{type:'agent', service, phase(develop/qa/debug/review), label}` 写入 pipeline/metrics.jsonl——回答"哪些 skill 被频繁/从不调用" |

**设计原则（克制）**：
- 复用现有 `log-usage.sh`（E1）和 `logMetrics`（E2），不引入新基础设施
- E2 的 `logAgentUsage` 是 logMetrics 的薄封装（type='agent' 区分），4 个 agent 调用点各加 1 行
- 均受 `config/tracking.yml` enabled 开关控制，可整体关闭

**验证**：E1 实测记录 `failed_checks="proto_ts_align;design_consistency;git_hygiene;"`；E2 编译进 harness-pipeline.js（4 阶段）；`run-evals.sh` 全绿；harness 自检 9 PASS。
**后续**：数据积累 2-4 周后，用 `analyze-usage.sh` + 新增分析脚本做脚本有效性分析（淘汰无用脚本/补薄弱门禁）。
