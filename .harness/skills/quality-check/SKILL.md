# quality-check — 质量/一致性检查（意图触发路由）

## 触发条件

用户表达"检查质量 / 检查一致性 / 体检 / 健康检查 / 检查流水线 / 有没有问题 / 检查下 / 全量检查"等**综合检查意图**时。触发词：`检查质量`、`质量检查`、`体检`、`健康检查`、`检查流水线`、`开发流水线的质量`、`检查一致性`、`有没有问题`、`全量检查`、`run checks`。

> **与 qa skill 的区别**：`qa` 聚焦**单个服务**的机械化验证（构建/测试/规范）。本 skill 是**跨维度意图路由**——用户说"检查质量/一致性/流水线"时，识别意图决定跑哪些检查脚本，聚合多维度报告。
> 若用户明确"检查服务 X"→ 单服务走 qa（`harness-checks.sh --service X`）；综合"体检/流水线/有没有问题"→ 本 skill 全量路由。

## 角色

你是质量检查协调员。解析用户意图 → 按路由表执行对应只读检查脚本 → 聚合结构化报告。**不修改任何文件**。

## 意图 → 脚本路由表

| 用户意图关键词 | 检查内容 | 脚本 |
|---|---|---|
| 流水线 / harness / 自身 | harness 基础设施一致性（引用/命名/文档同步/配置漂移/调用链/追溯/记忆/冲突检测） | `bash .harness/scripts/harness-self-check.sh` |
| 服务代码 / 规范 | 指定服务或全服务的构建/规范/安全/测试/设计一致性 | `bash .harness/skills/qa/scripts/harness-checks.sh [--service <name>]` |
| 设计一致性 / model / 迁移 | 设计↔代码一致（model 列 vs 迁移源） | `bash .harness/scripts/check-design-consistency.sh --all` |
| 需求冲突 / 变更重叠 | 两个变更同服务/同接口重叠 | `bash .harness/scripts/check-change-conflict.sh` |
| 管线回归 / eval | 流水线自身不回归 | `bash .harness/pipeline/evals/run-evals.sh` |
| 不确定 / "全部" | 全量体检（依次跑下 4 项） | 见 Step 2 |

## 执行步骤

### Step 1: 意图解析

- "检查服务 X" → `harness-checks.sh --service X` + `check-design-consistency.sh --service X`
- "检查流水线/自身" → `harness-self-check.sh`
- "检查设计/model 一致性" → `check-design-consistency.sh --all`
- "检查冲突/重叠" → `check-change-conflict.sh --all`
- "全部/体检/有没有问题" → 全量体检（Step 2）
- 意图不明确 → 问用户：服务代码 / 流水线自身 / 设计一致性 / 需求冲突 哪类

### Step 2: 全量体检（"全部/体检/有没有问题"）

按依赖顺序执行，聚合结果：

```bash
echo "=== 1/4 harness 自身一致性 ==="
bash .harness/scripts/harness-self-check.sh

echo "=== 2/4 设计↔代码一致性 ==="
bash .harness/scripts/check-design-consistency.sh --all

echo "=== 3/4 需求冲突预检 ==="
bash .harness/scripts/check-change-conflict.sh

echo "=== 4/4 管线自身回归 ==="
bash .harness/pipeline/evals/run-evals.sh
```

### Step 3: 聚合报告

```
✅/⚠️/❌ 质量检查报告
├─ 服务代码 (harness-checks):    PASS n / WARN n / FAIL n
├─ 流水线自身 (self-check):      PASS n / WARN n / FAIL n
├─ 设计一致性 (design-check):    WARN n 项（列: ...）
├─ 需求冲突 (conflict-check):    WARN n 项（服务重叠: ...）
└─ 管线回归 (evals):             PASS n / FAIL n
```

**关键**：WARN 提示非阻塞（如 model 列缺失可能是有意 legacy）；FAIL 必须处理 + 给修复方向（脚本自带 fix 提示）。用户只问单类 → 只报该类。

## 关键规则

1. **只读**——绝不修改文件
2. **按意图路由**——精准不跑多余
3. **报告聚合**——不 dump 原始输出，结构化汇总
4. **FAIL 必报**——突出 + 修复方向

## 关联

- `.harness/scripts/harness-self-check.sh` — harness 自检
- `.harness/skills/qa/scripts/harness-checks.sh` — 服务检查
- `.harness/scripts/check-design-consistency.sh` — 设计一致性
- `.harness/scripts/check-change-conflict.sh` — 需求冲突
- `.harness/pipeline/evals/run-evals.sh` — 管线回归
