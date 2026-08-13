# 变更目录结构规范

> 规范化的变更管理目录结构 · 版本化的评审追溯  
> 创建时间: 2026-07-11  
> 状态: ✅ 正式规范

---

## 一、目录结构标准

每个变更在 `.harness/changes/` 下创建独立目录：

```
.harness/changes/{变更类型}-{需求名称}-{YYYYMMDD}/
├── summary.md                         # 一页纸总结（必需）
│
├── 01_requirement/                    # 需求阶段
│   ├── proposal.md                   # 需求提案
│   ├── specs/                        # 技术规格
│   │   ├── backend/spec.md
│   │   └── frontend/spec.md
│   └── review/                       # 🆕 评审版本化
│       ├── review_v1.md              # 第 1 轮评审
│       ├── review_v2.md              # 第 2 轮评审
│       └── review_v3_APPROVED.md     # 最终通过版本
│
├── 02_architecture/                   # 架构阶段
│   ├── design.md                     # 架构设计
│   ├── tasks.md                      # 任务拆解
│   └── review/                       # 🆕 3 视角评审版本化
│       ├── iteration_1/
│       │   ├── security-arch.md      # 安全架构视角
│       │   ├── standard-eng.md       # 工程规范视角
│       │   └── design-biz.md         # 业务设计视角
│       └── iteration_2/
│           ├── security-arch_APPROVED.md
│           ├── standard-eng_APPROVED.md
│           └── design-biz_APPROVED.md
│
├── 03_coding/                         # 编码阶段
│   ├── {service}/
│   │   ├── coding_report_v1.md      # 🆕 编码报告
│   │   └── coding_report_v2.md      # 修复后的版本
│   └── review/
│       ├── code_review_v1.md        # 🆕 代码评审
│       └── code_review_v2_APPROVED.md
│
├── 04_testing/                        # 测试阶段
│   ├── {service}/
│   │   └── _qa.md                   # QA 报告
│   └── review/
│       └── test_review_v1.md        # 测试评审（如需要）
│
├── 05_ci/                             # CI 阶段
│   └── ci_result.json               # 🆕 结构化 CI 结果
│
├── 06_deployment/                     # 部署阶段
│   ├── deploy_plan.md               # 部署计划
│   ├── deploy_report.md             # 🆕 部署报告
│   └── rollback_plan.md             # 回滚方案
│
└── pipeline-evaluation.md            # 流水线评估报告
```

---

## 二、版本命名规范

### 2.1 评审文件版本

**格式**: `{type}_v{N}_{STATUS}.md`

- `{type}`: 评审类型（review / code_review / test_review）
- `{N}`: 轮次编号（1, 2, 3...）
- `{STATUS}`: 状态标记（可选）

**状态标记**:
- `APPROVED` - 评审通过
- `REVISION_REQUIRED` - 需要修改
- `ESCALATED` - 升级人工决策
- （无状态） - 进行中

**示例**:
```
review_v1.md                    # 第 1 轮评审（需要修改）
review_v2.md                    # 第 2 轮评审（需要修改）
review_v3_APPROVED.md           # 第 3 轮评审通过

code_review_v1.md               # 第 1 轮代码评审
code_review_v2_APPROVED.md      # 第 2 轮通过
```

### 2.2 多视角评审版本

架构评审有 3 个视角，使用迭代目录：

```
review/
├── iteration_1/
│   ├── security-arch.md        # 视角 1
│   ├── standard-eng.md         # 视角 2
│   └── design-biz.md           # 视角 3
└── iteration_2/
    ├── security-arch_APPROVED.md
    ├── standard-eng_APPROVED.md
    └── design-biz_APPROVED.md
```

### 2.3 其他产物版本

```
coding_report_v1.md             # 第 1 次编码
coding_report_v2.md             # 修复后
deploy_report_v1.md             # 第 1 次部署
deploy_report_v2.md             # 重新部署
```

---

## 三、核心原则

### 3.1 永不删除旧版本

✅ **DO**:
- 创建新版本文件
- 保留所有历史版本
- 完整的 Audit Trail

❌ **DON'T**:
- 覆盖旧文件
- 删除历史版本
- 在原文件上修改

### 3.2 状态标记明确

最终通过的版本必须有 `_APPROVED` 标记：
```
review_v3_APPROVED.md          # ✅ 清晰
review_v3.md                   # ❌ 不清楚是否通过
```

### 3.3 summary.md 记录版本历史

在 `summary.md` 中记录评审轮次：

```markdown
## 评审历史

### 需求评审
- 第 1 轮 (2026-07-10): REVISION_REQUIRED
  - 文件: `01_requirement/review/review_v1.md`
  - 主要问题: 缺少风险评估
- 第 2 轮 (2026-07-11): APPROVED
  - 文件: `01_requirement/review/review_v2_APPROVED.md`

### 架构评审
- 第 1 轮 (2026-07-11): REVISION_REQUIRED (2/3 通过)
  - security-arch: APPROVED
  - standard-eng: REVISION_REQUIRED
  - design-biz: APPROVED
- 第 2 轮 (2026-07-12): APPROVED (3/3 通过)
  - 文件: `02_architecture/review/iteration_2/`
```

---

## 四、迁移指南

### 4.1 现有变更目录迁移

对于已有的变更目录（如 `test-pipeline-work-records/`）：

1. **不需要立即重构** — 旧结构仍然有效
2. **新变更采用新规范** — 从下一个变更开始
3. **可选择性迁移** — 对重要变更可以重新组织

### 4.2 从旧结构到新结构

**旧结构**:
```
test-pipeline-work-records/
├── proposal.md
├── specs/
├── design.md
├── tasks.md
├── _review.md                  # 单个文件
├── _review_design-biz.md       # 单个文件
└── services/master-data-service/_qa.md
```

**新结构**（建议）:
```
test-pipeline-work-records/
├── summary.md
├── 01_requirement/
│   ├── proposal.md
│   ├── specs/
│   └── review/
│       └── review_v1_APPROVED.md
├── 02_architecture/
│   ├── design.md
│   ├── tasks.md
│   └── review/
│       └── iteration_1/
│           ├── security-arch_APPROVED.md
│           ├── standard-eng_APPROVED.md
│           └── design-biz_APPROVED.md
└── 04_testing/
    └── master-data-service/_qa.md
```

---

## 五、工具支持

### 5.1 创建变更目录脚本

```bash
# .harness/scripts/create-change.sh

#!/bin/bash
TYPE=$1      # feat / fix / refactor
NAME=$2      # 需求名称
DATE=$(date +%Y%m%d)

CHANGE_DIR=".harness/changes/${TYPE}-${NAME}-${DATE}"

mkdir -p "$CHANGE_DIR"/{01_requirement/{specs,review},02_architecture/review,03_coding/review,04_testing,05_ci,06_deployment}

# 创建 summary.md 模板
cat > "$CHANGE_DIR/summary.md" <<EOF
# Summary — ${NAME}

**变更名称**: ${NAME}
**创建时间**: $(date +%Y-%m-%d)
**状态**: 🚧 进行中
**类型**: ${TYPE}

---

## Executive Summary

（待补充）

---

## 评审历史

### 需求评审
- （待补充）

### 架构评审
- （待补充）

### 代码评审
- （待补充）

---

**创建人**: Owner Agent
**最后更新**: $(date +%Y-%m-%d)
EOF

echo "✅ 变更目录已创建: $CHANGE_DIR"
```

### 5.2 评审版本生成脚本

```bash
# .harness/scripts/next-review-version.sh

#!/bin/bash
REVIEW_DIR=$1    # review/ 目录路径
TYPE=${2:-review}  # review / code_review / test_review

# 查找最新版本
LATEST=$(ls "$REVIEW_DIR" | grep "^${TYPE}_v" | sort -V | tail -1)

if [ -z "$LATEST" ]; then
  NEXT_VERSION=1
else
  CURRENT_VERSION=$(echo "$LATEST" | sed -E "s/${TYPE}_v([0-9]+).*/\1/")
  NEXT_VERSION=$((CURRENT_VERSION + 1))
fi

echo "${REVIEW_DIR}/${TYPE}_v${NEXT_VERSION}.md"
```

---

## 六、检查清单

### 6.1 新变更创建时

- [ ] 使用标准目录结构
- [ ] 创建 `summary.md`
- [ ] 各阶段子目录齐全
- [ ] `review/` 目录已创建

### 6.2 评审轮次开始时

- [ ] 确认当前是第几轮
- [ ] 创建对应版本的文件（如 `review_v2.md`）
- [ ] 不覆盖旧版本

### 6.3 评审通过时

- [ ] 文件名添加 `_APPROVED` 标记
- [ ] 在 `summary.md` 中记录
- [ ] 保留所有历史版本

### 6.4 变更完成时

- [ ] 所有阶段目录完整
- [ ] 最终版本都有状态标记
- [ ] `summary.md` 包含完整历史
- [ ] `pipeline-evaluation.md` 已生成

---

## 七、FAQ

### Q1: 旧的变更目录需要重新组织吗？

**A**: 不需要。旧结构仍然有效，新变更采用新规范即可。可选择性地重新组织重要变更。

### Q2: 如果只有 1 轮评审就通过了？

**A**: 也要创建版本文件：`review_v1_APPROVED.md`

### Q3: 多视角评审如果只有 1 轮？

**A**: 仍然使用 `iteration_1/` 目录，保持结构一致性。

### Q4: 评审文件可以是其他格式吗（如 JSON）？

**A**: 可以，但建议使用 Markdown 以便人类阅读。如需结构化数据，可以同时生成 `.json` 和 `.md`。

---

## 八、参考资料

- `.harness/docs/_archive/harness-engineering-comparison.md` - Harness 工程对比分析
- `.harness/docs/pipeline-flow-complete.md` - 完整流程图
- `.harness/workflows/harness-pipeline.js` - 流程实现

---

**版本**: 1.0.0  
**创建时间**: 2026-07-11  
**状态**: ✅ 正式规范  
**维护者**: Harness Engineering Team
