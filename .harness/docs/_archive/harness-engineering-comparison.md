# Harness Engineering 对比分析

> 文章理念 vs 当前流水线的详细对比分析  
> 创建时间: 2026-07-11  
> 分析目标: 识别可借鉴的改进点

---

## 一、核心理念对比

### 文章核心观点
- Harness 的价值在于让 Agent 的错误变得**可控、可发现、可修复**
- 外部化的约束机制，而非依赖模型能力
- 四根支柱：上下文架构、Agent 专业化、持久化记忆、结构化执行

### 我们的现状
✅ **已实现**:
- `.harness/` 完整目录结构（agents/rules/skills/workflows/changes）
- Workflow 编排系统（harness-pipeline.js）
- 多 Agent 角色（Owner/Generator/QA/Reviewer/Debug）
- 变更追溯机制（changes/）

✅ **优势**:
- Workflow 自动化程度更高（代码驱动，而非文档驱动）
- 并行能力更强（多服务并行开发、3 视角并行评审）
- 工具集成更丰富（MCP、GitHub、自动化脚本）

---

## 二、逐项对比分析

### 2.1 Agent 角色定义

#### 文章方案
- **Application Owner Agent** 作为中枢
- 角色定义文件 ~400 行
- 包含：
  - 角色与项目背景
  - 配置中枢索引（Index & Map）
  - 七项核心职责
  - **10 阶段工作流调度指令**
  - 沟通原则与硬性约束

#### 我们的现状
**文件**: `.harness/agents/owner-agent.md` (已读取)

✅ **已有**:
- Owner Agent 定义存在
- 明确的职责划分
- Generator/QA/Reviewer/Debug 专业化角色

❌ **差距**:
1. **缺少明确的阶段定义和流转规则**
   - 文章有 10 个严格有序的阶段，每个阶段有明确的：
     - 触发条件（Entry Criteria）
     - Skill 加载（Skill Injection）
     - 质量门禁（Quality Gate）
     - 回退路径（Rollback Routes）
   
   - 我们的流水线有阶段概念，但在 **Agent 定义文件** 中未明确体现
   - 阶段流转逻辑在 workflow JS 代码中，不在 Agent 定义中

2. **Owner Agent 的调度指令不够显性**
   - 文章每个阶段都有详细的 "如何执行" 指令
   - 我们依赖 workflow.js 代码逻辑，Agent 定义中缺少这部分

3. **缺少 Human-in-the-Loop 确认点的明确标注**
   - 文章有 5 个明确的人工确认点
   - 我们虽然实际有（如评审后），但未在 Agent 定义中显性声明

💡 **可借鉴**:
```markdown
## Owner Agent 应补充的内容

### 阶段流转指令（Stage Orchestration）

**阶段 1: 需求分析**
- 触发条件: 用户输入需求
- 加载 Skill: requirement-analysis
- 产出: proposal.md + specs/
- 质量门禁: proposal 必须包含 7 个必填章节
- Human 确认点: ⚠️ 需求待决议时暂停
- 通过 → 阶段 2 | 失败 → 返回用户澄清

**阶段 2: 架构设计**
- 触发条件: 需求分析通过
- 加载 Skill: architect-design
- 产出: design.md + tasks.md
- 质量门禁: 
  - design.md > 5000 字
  - tasks.md 至少 3 个任务
- Review: 3 视角并行评审
- 最多 3 轮评审，超出升级人工
- 通过 → 阶段 3 | REVISION_REQUIRED → 返回阶段 1

... (其余阶段类似)

### Human-in-the-Loop 确认点

明确标注 5 个关键决策点：
1. 需求待决议确认
2. 架构设计评审后确认
3. 编码完成评审后确认
4. 部署参数确认
5. 最终交付确认
```

---

### 2.2 十阶段流程设计

#### 文章方案
```
需求分析 → 需求评审 → 编码实现 → 编码评审 → 单元测试编写
  → 单元测试评审 → 代码推送 → CI 验证 → 部署验证 → 用户确认
```

特点：
- **严格线性**，每个阶段有明确的质量门禁
- **评审与执行分离**（需求评审、编码评审、测试评审独立）
- **回退路径精确**（CI 失败 → 测试 0/0 回退阶段 5，编译错误回退阶段 3）
- **循环上限**（需求评审最多 3 轮，编码/测试评审最多 2 轮）

#### 我们的现状
**文件**: `.harness/docs/pipeline-flow-complete.md` + `harness-pipeline.js`

我们的流程（基于已读文件）：
```
路径选择 → 需求分析 → 需求评审 → 架构设计 → 架构评审 
  → 编码实现(Workflow并行) → QA验证 → Debug(条件触发) → Review(3视角)
  → 集成验证 → 文档交付
```

✅ **我们的优势**:
1. **并行能力更强**
   - 多服务并行开发（Workflow × N）
   - 3 视角并行评审（security-arch/standard-eng/design-biz）

2. **自动化根因分析**
   - QA FAIL 自动触发 Debug Agent
   - systematic-debugging 机制

3. **更灵活的工具集成**
   - MCP 工具链
   - 自动化脚本（harness-checks.sh）

❌ **差距**:
1. **测试环节未独立成阶段**
   - 文章: 单元测试编写 → 单元测试评审（独立阶段）
   - 我们: 测试在 QA 阶段内完成，未拆分

2. **缺少明确的循环上限控制**
   - 文章: 需求评审最多 3 轮，编码评审最多 2 轮
   - 我们: 在 summary.md 中看到有轮次记录，但未见硬性上限

3. **回退路径不够精确**
   - 文章: CI 失败根据错误类型精确回退（测试 0/0 → 阶段 5，编译错误 → 阶段 3）
   - 我们: QA FAIL → Debug → Generator 修复，回退逻辑在代码中但未文档化

💡 **可借鉴**:
```javascript
// harness-pipeline.js 可增强的部分

const STAGE_CONFIG = {
  QA: {
    maxRetries: 2,
    rollbackRules: [
      {
        condition: 'total_tests === 0 && status === "FAIL"',
        rollbackTo: 'TEST_WRITE',
        reason: '测试用例数为 0，回退到测试编写阶段'
      },
      {
        condition: 'compileError === true',
        rollbackTo: 'CODING',
        reason: '编译错误，回退到编码阶段'
      },
      {
        condition: 'testsFailed > 0 && retries < maxRetries',
        rollbackTo: 'DEBUG',
        reason: '测试失败，触发根因分析'
      }
    ]
  },
  REVIEW: {
    maxIterations: 3, // 评审最多 3 轮
    escalateToHuman: true, // 超出后升级人工
  }
}
```

---

### 2.3 质量门禁设计

#### 文章方案
**可程序化验证** — "If it can't be mechanically enforced, the agent will drift."

例子：
```
❌ 不够具体: "检查 CI 是否通过"
✅ 可验证: status == SUCCESS && total_tests > 0 && passed == total

❌ 不够具体: "生成评审报告"
✅ 可验证: 文件存在 && 包含必填章节 && 字数 > 500
```

#### 我们的现状
**文件**: `.harness/skills/qa/scripts/harness-checks.sh`

✅ **已有**:
- AST 级别的代码检查（JSON.parse/stringify 检查、类型安全检查）
- 编译验证
- 测试覆盖率检查
- Proto 规范验证

❌ **差距**:
1. **门禁条件散落在脚本中，未集中管理**
   - 文章建议: 在 Agent 定义或流程配置中集中声明所有门禁
   - 我们: 门禁逻辑在 harness-checks.sh、workflow.js 等多处

2. **缺少文档层面的门禁检查**
   - 例如: spec.md 是否包含必填章节？字数是否达标？
   - design.md 是否足够详细（> 5000 字）？

3. **缺少 summary.md 的结构化验证**
   - 是否记录了关键信息（评审轮次、CI 结果、部署状态）？

💡 **可借鉴**:
```yaml
# .harness/config/quality-gates.yml（新增配置文件）

gates:
  requirement_analysis:
    - type: file_exists
      path: "proposal.md"
    - type: content_check
      file: "proposal.md"
      required_sections:
        - "## 背景"
        - "## 目标"
        - "## 技术方案"
        - "## 风险评估"
    - type: word_count
      file: "proposal.md"
      min: 500
  
  architecture_design:
    - type: file_exists
      path: "design.md"
    - type: word_count
      file: "design.md"
      min: 5000
    - type: task_count
      file: "tasks.md"
      min: 3
  
  coding:
    - type: compilation
      expect: SUCCESS
    - type: ast_check
      rules:
        - no_json_string_literals
        - no_unsafe_any
  
  qa:
    - type: ci_check
      conditions:
        - status == "SUCCESS"
        - total_tests > 0
        - passed == total
        - coverage >= 30  # 我们今天设定的标准
    
  review:
    - type: file_exists
      path: "review/*.md"
    - type: review_verdict
      allowed: ["APPROVED", "REVISION_REQUIRED"]
      max_iterations: 3
```

然后在 workflow 中引用：
```javascript
const gates = loadQualityGates('.harness/config/quality-gates.yml')
const qaResult = await runQA(task)
const passed = validateGate(qaResult, gates.qa)
if (!passed) {
  // 根据失败类型精确回退
}
```

---

### 2.4 分离执行与评判

#### 文章方案
"将做事的 Agent 和评判的 Agent 分开，是一个强有力的杠杆。"

- 编码 Agent ≠ 评审 Agent
- Agent-to-Agent Review 在 Human Review 之前

#### 我们的现状
✅ **已实现** — 这是我们的强项！

- **Generator** (执行) ≠ **Reviewer** (评判)
- **QA Agent** 独立验证
- **Debug Agent** 根因分析
- **3 视角并行评审**（security-arch / standard-eng / design-biz）

✅ **我们做得更好**:
- 文章只提到"分离"，我们实现了 **多维度评审**
- 安全视角、工程规范视角、业务设计视角同时检查

💡 **可微调**:
- 确保评审 Agent 的检查清单在 `.harness/skills/review.md` 中明确
- 避免评审过于宽松或过于严格（今天测试中发现的"轮次限制"问题）

---

### 2.5 上下文架构（分层加载）

#### 文章方案
三层加载策略：
- **L1 - 会话常驻层**: Agent 定义 + Rules（< 40% 上下文窗口）
- **L2 - 阶段触发层**: 进入阶段时加载对应 Skill
- **L3 - 按需查询层**: Wiki 不主动加载，Agent 自主查阅

核心原则: **Just-enough Context**

#### 我们的现状
✅ **已有分层思想**:
- Rules 在 `.harness/rules/`
- Skills 按需加载
- Knowledge 在 `.harness/knowledge/`
- Memory Index 机制（`.harness/knowledge/memory/.memory-index.json`）

✅ **我们的优势**:
- **Memory Index 查询机制** — 比文章更进一步
  - 索引驱动的 O(K) 查询
  - 避免全文加载 MEMORY.md

❌ **差距**:
1. **缺少上下文窗口填充率的监控**
   - 文章建议: 保持在 40% 以下
   - 我们: 未见相关监控机制

2. **Agent 定义文件可能过长**
   - owner-agent.md 的实际长度需要评估
   - 是否遵循了 "Index & Map" 原则（~100 行索引，指向详细文档）

💡 **可借鉴**:
```javascript
// 在 workflow 中增加上下文监控

function monitorContextUsage(agent) {
  const currentTokens = estimateTokens(agent.context)
  const maxTokens = agent.contextWindow
  const fillRate = currentTokens / maxTokens
  
  if (fillRate > 0.4) {
    log.warn(`Context fill rate: ${(fillRate * 100).toFixed(1)}% - exceeding sweet spot`)
    // 考虑精简上下文或分批加载
  }
  
  return fillRate
}
```

---

### 2.6 变更管理（Audit Trail）

#### 文章方案
每个需求独立目录，标准化结构：
```
{变更类型}-{需求名称}-{YYYYMMDD}/
├── summary.md          # 一页纸总结
├── request_analysis/
│   ├── spec.md
│   ├── tasks.md
│   └── review/        # 版本递增 (v1, v2, v3...)
├── coding/
│   ├── coding_report_v1.md
│   └── review/
├── unit_test/
├── ci_result/
└── deployment/
```

**核心**: 评审文件版本递增，旧版本永不删除

#### 我们的现状
✅ **已有完整的变更管理**:
- `.harness/changes/` 目录
- 每个变更独立目录（如 `test-pipeline-work-records/`）
- 包含 summary.md、proposal.md、specs/、design.md、tasks.md、QA 报告等

✅ **我们做得更好**:
- **proposal.md** — 需求提案阶段（文章未提及）
- **design.md** — 架构设计文档（文章未单独提及）
- **pipeline-evaluation.md** — 流水线评估报告（文章未提及）

❌ **差距**:
1. **评审文件版本管理不明确**
   - 文章: review_v1.md, review_v2.md, review_v3.md（递增保留）
   - 我们: 从 `test-pipeline-work-records/` 中看到 `_review.md`、`_review_design-biz.md` 等，但版本策略不清晰

2. **coding_report 和 unit_test 报告未独立**
   - 我们有 `_qa.md`，但未看到独立的编码报告

💡 **可借鉴**:
```markdown
## 变更目录结构优化

建议规范化为：

{变更类型}-{需求名称}-{YYYYMMDD}/
├── summary.md                    # ✅ 已有
├── 01_requirement/
│   ├── proposal.md              # ✅ 已有
│   ├── specs/                   # ✅ 已有
│   └── review/
│       ├── review_v1.md         # 🆕 版本化
│       ├── review_v2.md
│       └── review_v3_APPROVED.md
├── 02_architecture/
│   ├── design.md                # ✅ 已有
│   ├── tasks.md                 # ✅ 已有
│   └── review/
│       ├── review_v1.md         # 🆕 版本化
│       └── review_v2_APPROVED.md
├── 03_coding/
│   ├── coding_report_v1.md      # 🆕 新增
│   └── review/
│       └── code_review_v1.md    # 🆕 新增
├── 04_testing/
│   ├── {service}/_qa.md         # ✅ 已有
│   └── test_review_v1.md        # 🆕 新增
├── 05_ci/
│   └── ci_result.json           # 🆕 新增
├── 06_deployment/
│   └── deploy_report.md         # 🆕 新增
└── pipeline-evaluation.md        # ✅ 已有（优势）
```

---

### 2.7 Skill 体系

#### 文章方案
每个 Skill = 结构化 SOP

例如 `coding-skill` 包含 8 份分层编码规范：
- Controller 实现 Spec
- 接口定义/实现 Spec
- 业务逻辑 Spec
- 数据层 Spec
- 适配层 Spec
- 文档生成 Spec

#### 我们的现状
✅ **已有完整的 Skills**:
- `.harness/skills/` 目录
- architect-design.md
- requirement-analysis.md
- review.md
- qa.md
- dispatch.md（任务分发）
- 等等

✅ **我们的优势**:
- **qa/** 下有完整的自动化检查脚本
- **templates/** 提供了多种模板（frontend-design、webapp-testing 等）

❌ **差距**:
1. **缺少分层编码规范**
   - 文章的 8 层编码 Spec（Controller/Service/Domain/DAO/Adapter）
   - 我们有 `.harness/rules/项目编码规范.md`，但可能不够分层详细

2. **Skill 之间的调用关系不够清晰**
   - 哪个阶段加载哪个 Skill？
   - 应该在 Owner Agent 定义中明确，或在单独的配置文件中

💡 **可借鉴**:
```markdown
## 新增：分层编码规范

.harness/skills/coding/
├── 00-overview.md              # 总览
├── 01-controller-spec.md       # 表现层规范
├── 02-service-spec.md          # 业务层规范
├── 03-domain-spec.md           # 领域层规范
├── 04-dao-spec.md              # 数据访问层规范
├── 05-adapter-spec.md          # 适配层规范
├── 06-proto-spec.md            # Proto 定义规范
├── 07-test-spec.md             # 测试编写规范
└── 08-doc-spec.md              # 文档生成规范

每个文件包含：
- 该层的职责边界
- 命名约定
- 代码模板
- 反面案例
- 检查清单
```

---

## 三、关键启示与行动建议

### 3.1 我们已经做得很好的地方

1. ✅ **Workflow 自动化更彻底** — 代码驱动 vs 文档驱动
2. ✅ **并行能力更强** — 多服务 + 3 视角评审
3. ✅ **工具集成更丰富** — MCP + 自动化脚本
4. ✅ **Memory Index 机制** — 比文章更先进
5. ✅ **Agent 专业化已实现** — Generator/QA/Reviewer/Debug

### 3.2 可立即借鉴的改进点

#### 优先级 P0（立即可做）

1. **在 Owner Agent 定义中补充阶段流转指令**
   - 文件: `.harness/agents/owner-agent.md`
   - 增加: 10 阶段的触发条件、Skill 加载、质量门禁、回退路径
   - 明确标注 5 个 Human-in-the-Loop 确认点

2. **质量门禁配置化**
   - 新增: `.harness/config/quality-gates.yml`
   - 集中管理所有门禁条件
   - 可程序化验证（而非散落在脚本中）

3. **评审文件版本管理规范**
   - 采用递增版本号（review_v1.md, review_v2.md, review_v3.md）
   - 旧版本永不删除
   - APPROVED 标记最终版本

#### 优先级 P1（近期优化）

4. **循环上限控制**
   - 在 workflow 配置中明确：
     - 需求评审最多 3 轮
     - 编码评审最多 2 轮
     - 超出后自动升级人工决策

5. **回退路径精确化**
   - QA FAIL 根据错误类型精确回退：
     - 测试 0/0 → 测试编写阶段
     - 编译错误 → 编码阶段
     - 测试失败 → Debug 分析

6. **上下文窗口监控**
   - 在 workflow 中增加 context fill rate 监控
   - 保持在 40% 以下的 Sweet Spot

#### 优先级 P2（长期完善）

7. **分层编码规范**
   - 参考文章的 8 层 Spec
   - 按 Controller/Service/Domain/DAO/Adapter 分层

8. **独立的测试阶段**
   - 将测试编写和测试评审独立出来
   - 而非混在 QA 阶段内

9. **文档层面的门禁**
   - 验证 spec.md/design.md 的结构完整性
   - 字数达标检查（design.md > 5000 字）

---

## 四、对比总结表

| 维度 | 文章方案 | 我们现状 | 差距 | 借鉴价值 |
|------|---------|---------|------|---------|
| **Agent 角色定义** | 400 行，包含阶段调度指令 | 有定义但阶段逻辑在 JS 中 | ⚠️ 中 | ⭐⭐⭐⭐⭐ |
| **阶段流程** | 10 阶段线性 + 精确回退 | 灵活阶段 + 并行能力强 | ⚠️ 小 | ⭐⭐⭐ |
| **质量门禁** | 集中配置 + 可程序化 | 散落脚本中 | ⚠️ 中 | ⭐⭐⭐⭐⭐ |
| **执行与评判分离** | 提出理念 | 已实现 + 3 视角评审 | ✅ 领先 | ⭐ |
| **上下文管理** | 3 层加载 + 40% 填充率 | 有分层 + Memory Index | ⚠️ 小 | ⭐⭐⭐ |
| **变更管理** | 版本递增保留 | 完整但版本策略不清 | ⚠️ 小 | ⭐⭐⭐⭐ |
| **Skill 体系** | 8 层编码 Spec | 有 Skills 但不够分层 | ⚠️ 中 | ⭐⭐⭐⭐ |
| **循环控制** | 明确上限（3轮/2轮） | 未见硬性限制 | ⚠️ 中 | ⭐⭐⭐⭐ |
| **并行能力** | 未提及 | 多服务并行 + 3 视角并行 | ✅ 领先 | - |
| **自动化根因** | 未提及 | Debug Agent + systematic | ✅ 领先 | - |

**借鉴价值图例**:
- ⭐⭐⭐⭐⭐ 立即实施，高价值
- ⭐⭐⭐⭐ 近期优化，有价值
- ⭐⭐⭐ 长期完善，中等价值
- ⭐ 已领先，无需借鉴

---

## 五、核心结论

### 我们的 Harness 工程已经相当成熟

文章提出的四根支柱：
1. ✅ **上下文架构** — 已有分层 + Memory Index
2. ✅ **Agent 专业化** — 已有 5 种专业 Agent
3. ✅ **持久化记忆** — 已有 changes/ 和 knowledge/
4. ✅ **结构化执行** — 已有 workflow 编排

### 我们的独特优势

1. **并行能力** — 文章未涉及
2. **自动化根因分析** — 文章未提及
3. **代码驱动的 Workflow** — 比文档驱动更可靠
4. **Memory Index** — 比简单的 Wiki 更先进

### 最值得借鉴的 3 点

1. **⭐⭐⭐⭐⭐ 阶段流转指令显性化** — 从 JS 代码提升到 Agent 定义文件
2. **⭐⭐⭐⭐⭐ 质量门禁配置化** — 集中管理，可程序化验证
3. **⭐⭐⭐⭐ 循环控制 + 精确回退** — 防止无限循环，提升效率

---

**创建人**: Claude  
**最后更新**: 2026-07-11  
**状态**: ✅ 对比分析完成
