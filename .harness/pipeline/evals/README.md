# Pipeline Evals — 管线自身回归语料库（P4.1）

> 把管线实战中暴露的摩擦案例固化为回归测试，**管线改动后跑 `run-evals.sh` 防回归**。
> 每个 eval 对应一个或多个实战缺陷，是管线"自我进化不倒退"的看门狗。

## 运行

```bash
bash .harness/pipeline/evals/run-evals.sh   # 跑全部，任一失败 exit 非 0
node .harness/pipeline/evals/<name>.eval.js  # 跑单个
```

> 已接入 `harness-checks.sh` #20「pipeline evals」——每次 QA 跑门禁时顺带验证管线自身不回归。eval 夹具全部自包含（临时目录），不依赖真实变更目录。

## 用例清单（↔ 实战缺陷）

| 文件 | 覆盖 | 对应实战缺陷 |
|------|------|------|
| `p1-convergence.eval.js` | P1 收敛闭环：consumeDecision 读删语义、specContentHash 内容敏感+确定性、mustFixKey 签名、收敛早停判据 | D1 评审盲循环 / D2 决策回环 / D3 缓存毒化 / D4 无收敛判据 |
| `p2p3-guards.eval.js` | P2 成本护栏：budgetLevel soft/hard 分级、costSummary 格式、routeModel 路由；P3 确定性门禁：specDeterministicCheck（追溯表/错误码登记意图/REQ 引用） | D5 无成本护栏 / D6 门禁不一致 / D8 模型评判模型 |
| `p4p5-evals.eval.js` | P4 反馈回填：反馈 JSONL 结构可解析；P5 状态可靠：getPath 点+方括号路径、validateResumeState 必需字段校验（D7 场景）、saveState/loadState 落盘优先 | D7 ctx 脆弱静默丢状态 |

## 变更流程

1. 改 `.harness/workflows/harness-spec-pipeline.js` 或 `.harness/skills/qa/scripts/harness-checks.sh`
2. `bash .harness/pipeline/evals/run-evals.sh` 确认全绿
3. 新缺陷 → 新增 eval 用例（先红后绿），固化进语料库
