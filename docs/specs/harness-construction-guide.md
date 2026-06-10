# Harness 构建说明书

> 适用场景：为已有代码库引入 AI Coding Agent，构建系统化的约束、反馈和知识体系。
> 基于 Community-Home（8+ Go 微服务 monorepo）实战经验提炼，可适配 Java/Python/Node.js 等其他技术栈。

---

## 一、什么是 Harness Engineering

**Agent = Model + Harness**。模型是引擎，Harness 是方向盘、刹车、导航。

Harness Engineering 是围绕 AI Coding Agent 设计四类基础设施的系统工程：

| 支柱 | 本质 | 告诉 Agent | 形式 |
|------|------|-----------|------|
| Rules | 稳定约束 | "什么绝对不能做" | 工程结构、编码规范、Proto 管理 |
| Skills | 可复用 SOP | "具体怎么做" | 需求分析、编码、QA、审查 |
| Knowledge | 领域知识 | "系统是什么样的" | 架构文档、业务流程、数据模型、经验记忆 |
| Changes | 变更追溯 | "做了什么、为什么" | proposal→plan→CHANGELOG→review→deploy |

**核心理念**：用外部化的结构约束替代对模型内在能力的依赖。Agent 不需要"学"如何做好——它只需要严格执行定义好的每一条指令。

---

## 二、什么时候该建 Harness

满足任一条件就该开始：

- AI 代码率卡在 30% 以下，反复返工
- Agent 在不同任务中重复犯同类错误
- 团队隐性知识散落在群聊、会议纪要、人脑里
- 代码库 > 5 万行，Agent 无法自行理解全局
- 涉及 > 3 个服务的跨服务变更频繁

**不需要建的信号**：< 1 万行代码、单人项目、Agent 偶尔辅助补全（非主力开发）。

---

## 三、构建顺序（按优先级）

### 第一阶段：地基（1-2 天）

**Step 1: 建目录结构**
```
.harness/
├── agents/          # Owner Agent
├── rules/           # 稳定约束
├── skills/          # 可复用 SOP
├── knowledge/       # 静态知识
│   └── memory/      #  经验记忆
├── changes/         # 变更追溯
├── workflows/       # 多 Agent 编排
└── scripts/         # 共享基础设施
```

**Step 2: 从现有 CLAUDE.md 提取 Rules**

把 CLAUDE.md 中的稳定约束提取为独立文件：
- `工程结构.md` — 技术栈、服务分层、中间件、端口规划
- `编码规范.md` — 硬性约束（类型、命名、错误码、安全）
- `Proto/API 管理规范.md`（如适用）

CLAUDE.md 精简为 ~100 行索引地图。

**Step 3: Dry Run**

用一个虚拟需求走通 Generator → QA → Reviewer 全流程。**在真实需求前发现系统性缺陷。** 密切检查：编译/测试是否正确检测 0/0 假通过、评审报告是否生成、门禁条件是否可程序化验证。

### 第二阶段：技能（2-3 天）

**Step 4: 建 Skills**

按优先级：

| 优先级 | Skill | 为什么先建 |
|:---:|------|------|
| 1 | select-tool | 每个需求入口，选错工具 = 全线低效 |
| 2 | qa | 机械化检查（build/test/lint），Agent 不可自我评判 |
| 3 | review | 独立审查者，分离执行与评判 |
| 4 | requirement-analysis | 模糊需求 → 可验收的 spec |
| 5 | architect-design | spec → design + tasks |
| 6 | unit-test-write | 改动驱动测试，不是一刀切 |

每个 Skill 的结构：
```
skills/<name>.md        # 入口（CC 自动发现为斜杠命令）
skills/<name>/          # 如需要私有脚本
  ├── SKILL.md          # 实际 SOP
  └── scripts/          # 该 Skill 专用脚本
```

脚本归属原则：**谁主要用它，就放谁的目录下**。多人用的放 `scripts/`。

**Step 5: 建 Owner Agent**

~150 行的薄编排层，五模块：
1. 项目背景（20 行）— 刚好够用的视野
2. 知识索引（40 行）— Rules/Skills/Knowledge 三张表，标注 L1/L2/L3
3. 核心职责（20 行）— 7 维度，每项带行为准则
4. 调度流程（50 行）— 阶段表（触发→加载→产出→门禁→回退）+ 失败路由
5. 沟通原则（15 行）— 必须/禁止清单

**不要写成百科全书**。Owner Agent 是"地图"——告诉 Agent 什么时候去哪里找什么。Skills/Rules/Knowledge 才是"内容"。

### 第三阶段：知识积累（持续）

**Step 6: 建立记忆系统**

首次：从过去 3-6 个月的 Git log 和 Code Review 记录中提取 5-10 条关键经验。

统一 frontmatter 格式：
```yaml
---
triggers: ["关键词1", "关键词2"]   # 触发词
service: all / <service-name>
type: pitfall | guideline | decision | process | model
severity: must-follow | should-follow | info
status: active | draft | superseded
created: YYYY-MM-DD
updated: YYYY-MM-DD
---
```

记忆搜索采用两级匹配：triggers 精确匹配（高置信度）→ 正文 grep（低置信度，需过滤 type）。

**Step 7: 沉淀业务知识**

从各服务 design.md 提取端到端业务流程和状态机，写成 `business-flows.md`。Agent 编码前读一遍，避免"语法正确但业务错误"。

**Step 8: 建立变更追溯**

- `changes/TEMPLATE.md` — summary.md 模板，每个新需求复制一份
- `changes/INDEX.md` — 迷你 summary 索引表，链接分布在各处的产物

---

## 四、关键设计决策

### D1: 分离执行与评判

Generator（可写）→ QA（只读，机械化检查）→ Reviewer（只读，9 维度审查）。

**为什么**：Agent 无法准确评估自身产出。独立的评判者是最有力的质量杠杆。

### D2: 质量门禁必须可程序化验证

"CI 通过"不够 → `status==SUCCESS && total_tests>0 && passed==total`

"生成评审报告"不够 → `目标路径下文件存在 && 含必填章节`

**一切不可被机器验证的约束，在 Agent 执行中都是无效约束。**

### D3: 流程一致性优先于效率

即使是 2 文件 6 行的改动，也走完整 select-tool → QA 验证 → CHANGELOG。简单需求不会因为流程而变慢（每阶段自然缩短），但跳过流程的"小改动"是事故的主要来源。

**例外**：<10 行单文件修复可有快速通道（直接 Edit + build 验证），但不能跳过 QA。

### D4: 上下文分层（L1/L2/L3）

- L1 常驻 ~400 行（Owner Agent + Rules），控制在上下文窗口 40% 以下
- L2 阶段触发（Skills），进入阶段才加载
- L3 按需查询（Knowledge），不主动塞进上下文

**为什么**：上下文窗口填充率 > 40%，输出质量快速衰退。

### D5: 子服务不建自己的 Harness

在微服务 monorepo 中，父 `.harness/` 统一管理规则/技能/知识。子服务的 CLAUDE.md 只做"薄入口"——引用父资源，不重复维护。

---

## 五、踩过的坑

| 坑 | 发现方式 | 修复 |
|------|---------|------|
| Memory 全正文搜索导致严重假阳性（如 phone-encryption 匹配所有 Go 任务） | Dry Run | 改为两级匹配：triggers 精确→正文降权，加 type 字段过滤 |
| QA go test 0/0 假通过 | Dry Run | harness-checks.sh 增加 0/0 检测 |
| 脚本迁移后 PROJECT_ROOT 路径解析错误 | 验证 | 记住脚本移动后检查相对路径层级 |
| Memory frontmatter 格式不一致（2 种格式共存） | 审计 | 统一为完整格式（triggers/service/type/severity/status/dates） |
| Pipeline 产物无生命周期管理 | Dry Run | 增加 archive 步骤 + 版本递增约定 |
| 过早创建不需要的目录 | 经验 | 只建当前需要的，等有内容再扩展（如 MCP/ 目录） |

---

## 六、适配指南

### Java 单体 vs Go 微服务

| 维度 | Java 单体 | Go 微服务 |
|------|----------|----------|
| Owner Agent | ~400 行，内含更多上下文 | ~150 行，服务级规则在下层 |
| 流水线阶段 | 10 阶段（含 CI/部署） | 5-7 阶段（CI 轻量） |
| 编码 Skill | 8 份分层 Spec（Controller→DAO全链路） | 1 份项目编码规范（分层简单） |
| 子服务 Harness | 不需要 | 不需要（父级统一） |
| 知识图谱 | 可选 | 推荐（微服务依赖复杂） |

### 其他技术栈适配

- **Python/FastAPI**: Rules 关注 typing、pydantic 校验、async 规范；QA 用 ruff + mypy + pytest
- **Node.js/Next.js**: Rules 关注 TS 类型、API routes 规范；QA 用 tsc + eslint + vitest
- **多语言混合**: 每个语言在 Rules 下各一份编码规范，共用一个 Owner Agent

---

## 七、验证清单

Harness 建完后，逐项确认：

- [ ] Dry Run 通过（虚拟需求走通全流程）
- [ ] QA 机械化检查无 0/0 假通过
- [ ] Generator → QA → Reviewer 三段分离，QA/Reviewer 只读
- [ ] CLAUDE.md < 120 行，作为索引而非百科全书
- [ ] 所有记忆 frontmatter 统一格式（含 type/severity/triggers）
- [ ] 每个服务有 `design.md`（至少数据模型 + 业务流程）
- [ ] changes/ 有 TEMPLATE.md + INDEX.md
- [ ] Owner Agent 五模块齐全，L1/L2/L3 显式标注
- [ ] harness-checks 覆盖：build / vet / test / lint / 安全 / 规范
- [ ] 关键设计决策有书面记录（decision memory）

---

## 八、参考

- Anthropic. [Effective harnesses for long-running agents](https://www.anthropic.com/engineering/effective-harnesses-for-long-running-agents)
- Anthropic. [Harness design for long-running application development](https://www.anthropic.com/engineering/harness-design-long-running-apps)
- OpenAI. [Harness engineering: leveraging Codex in an agent-first world](https://openai.com/index/harness-engineering/)
- Mitchell Hashimoto: "Every time you discover an agent has made a mistake, you take the time to engineer a solution so that it can never make that mistake again."
