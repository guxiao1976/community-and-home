# review

## 触发条件

对服务代码进行 12 维度审查。触发词：`审查 <服务名>`、`review <服务名>`、`全量审查 <服务名>`、`检查 <服务名> 的代码`。

## 角色

你是 Reviewer — 三种模式：**计划评审**（审 spec）、**设计评审**（审 design+tasks）、**执行评审**（审代码）。只审查、不修改。**完整角色定位 / 权限边界 / 上下文加载清单 / 服务名映射 / 工具熔断见 `.harness/agents/subagents/reviewer.md`**。

## 分级统一（两模式共用）

计划评审与执行评审的分级语义对齐，避免跨阶段换算成本：

| 层级 | 计划评审 | 执行评审 | 核心定义 |
|:---:|:---:|:---:|------|
| 阻塞级 | MUST FIX | CRITICAL | 不修复导致架构错误、安全风险、数据丢失、业务不可用 |
| 建议级 | SHOULD FIX | WARNING | 不修复不影响核心，但有逻辑缺陷、性能隐患、规范违反 |
| 提示级 | INFO | NOTE | 优化建议、风格、补充说明 |

## 模式一：计划评审（Plan Review）— 4 视角并行

在阶段 2（需求评审）使用。Owner Agent 同时启动 4 个 Reviewer 子 Agent，各负责不同维度，并行审查 **spec**（不含 tasks，tasks 是阶段 3 架构设计产物，由「设计评审」审）。投票 ≥2/3 APPROVED 即通过（CRITICAL 级 MUST FIX 一票否决）。

### 四个视角

| Lens                | 负责维度              | 审查焦点                                                                                                   |
| ------------------- | ----------------- | ------------------------------------------------------------------------------------------------------ |
| **coverage** 覆盖完整性  | 需求覆盖, 场景完整性, 边界识别 | 每个 Requirement 是否覆盖需求决策点？每个 Requirement ≥1正向+1异常 Scenario？边界条件是否考虑？`[NEEDS CLARIFICATION]` 是否遗漏？ |
| **structure** 结构合理性 | 职责边界, 一致性        | proposal 影响范围 ↔ specs 职责边界是否一致？各 capability 职责是否清晰无重叠？                                                 |
| **clarity** 清晰可执行   | 粒度, 歧义, 一致性       | spec 中 SHALL/MUST 是否有唯一解释？Scenario 是否具体到让实现者得出相同行为？术语是否一致？               |
| **validity** 业务有效性   | 业务自洽, 非功能, 合规    | 需求是否符合核心业务规则？安全/性能/兼容性是否在 spec 明确？是否合规？有无架构冲突/技术债/依赖风险？ |

### 输入

**评审对象（磁盘）**：
- `.harness/changes/<name>/request.md` — 对照原始需求，检查评审是否偏离用户初衷
- `.harness/changes/<name>/proposal.md`
- `.harness/changes/<name>/specs/*/spec.md`

> 完整上下文加载清单（含预加载基准上下文）见 `.harness/agents/subagents/reviewer.md`。

### 产出（每视角独立）

```
.harness/changes/<name>/review/
├── spec_review_coverage_v1.md    # 覆盖完整性视角
├── spec_review_structure_v1.md   # 结构合理性视角
├── spec_review_clarity_v1.md     # 清晰可执行视角
└── spec_review_validity_v1.md    # 业务有效性视角
```

版本递增: v1 → v2 → v3，旧版不删。

### 审查报告格式

```markdown
# Plan Review — <change-name>（<视角名>视角）

**审查维度**: <负责维度>

## 摘要
- 🔴 MUST FIX: N / 🟡 SHOULD FIX: N / 🔵 INFO: N

## 发现

### 🔴 MUST FIX
| # | 文件:行号/章节 | 问题 | 修复建议 |
|---|-------------|------|---------|

### 🟡 SHOULD FIX
| # | 文件:行号/章节 | 问题 | 建议 |
|---|-------------|------|------|

### 🔵 INFO
| # | 建议 |
|---|------|

---
VERDICT: APPROVED / REVISION
---
```

### VERDICT 规则

- APPROVED — 该视角无 MUST FIX
- REVISION — 存在 ≥1 MUST FIX

### 投票规则（Owner Agent 裁决）

```
任意视角出现 CRITICAL 级 MUST FIX（架构违反/安全漏洞/数据丢失/业务不可用）→ 直接 REVISION（一票否决）
无 CRITICAL 时：3/3 APPROVED → 进入阶段 3
                2/3 APPROVED → 进入阶段 3（少数 REVISION 的普通 MUST FIX 记录到 summary）
                ≤1/3 APPROVED → 回阶段 1，最多 3 轮
```

**异议闭环**：
- 所有 REVISION 视角的问题，报告须含「问题跟踪表」（状态：待修复/已修复/已验证）
- 下一轮评审优先校验上一轮问题修复情况
- 票决通过后的遗留 SHOULD FIX/INFO，同步到 BACKLOG 技术债条目，不石沉大海

每条评审意见必须包含：问题描述、修改建议、优先级（MUST FIX / SHOULD FIX / INFO）。

## 模式一.5：设计评审（Design Review）— 审 design + tasks

在阶段 3（架构设计）产出 `design.md` + `tasks.md` 后使用。审「设计正确性」——需求评审（模式一）审「需求拆得对不对」，设计评审审「设计对不对」。

### 审查对象（磁盘）
- `.harness/changes/<name>/design.md` — 数据模型 / 接口设计 / 业务流程 / Proto 变更 / 安全考虑
- `.harness/changes/<name>/tasks.md` — 任务拆分 / 依赖顺序

### 视角分工（2 视角并行，降低单 agent 确认偏差）

| 视角 | 负责维度 |
|------|---------|
| **data-model** 数据模型 | 数据模型、服务归属 |
| **interface-proto** 接口契约+Proto | 接口契约、Proto 破坏性、依赖顺序 |

### 审查焦点

| 维度 | 审查焦点 |
|------|---------|
| **服务归属** | tasks 分服务是否合理？谁拥有数据谁提供接口？ |
| **数据模型** | 数据模型是否满足 spec 需求？字段/关系是否正确？ |
| **接口契约** | gRPC/Proto 接口是否自洽？有无契约冲突？ |
| **Proto 破坏性** | Proto 变更是否向后兼容（不删/改字段号、不改字段类型）？ |
| **依赖顺序** | 基础设施→核心→辅助→前端？ |

### 任务粒度刚性校验（替代主观的「1-4h/1-5 文件」）

- 单任务不得跨多个服务
- 单任务不得同时包含「数据模型变更 + 业务逻辑 + 前端页面」三类工作
- 涉及 Proto 变更 / Migration 变更必须独立成任务
- 单任务测试用例数 1~10 个（超出需拆）

> 命中任一违反 = MUST FIX（任务粒度不合理）。

### VERDICT
- APPROVED — 设计正确，无阻塞问题
- REVISION — 存在 ≥1 MUST FIX，回架构设计师修订（带反馈，≤3 轮）

## 模式二：执行评审（Execution Review）

在阶段 5（编码+测试）QA PASS 后使用，审查代码实现。**以下为完整 SOP**。

## 执行步骤

> 服务名映射见 `.harness/agents/subagents/reviewer.md`（权威源 `.harness/registry/services.json`）。

### Step 1: 确定审查范围

```bash
# 基线可配置：默认 main，worktree/分支模式从 change 记录读分支/commit
git diff ${BASE_BRANCH:-main}...HEAD -- services/<name>/
```

或用户指定的文件/PR/commit。

### Step 2: 加载上下文

完整上下文加载清单见 `.harness/agents/subagents/reviewer.md`「上下文加载清单（执行评审）」。

**QA 联动**（QA 先于 Review 执行，Review 不重复跑 QA 已跑的构建）：
- 加载 QA 报告，对 QA 未覆盖的分支 / 异常场景 / 边界条件重点审查
- QA 失败的用例，校验根因是否属代码设计缺陷（而非测试问题）
- QA 覆盖率低于阈值（核心链路 < 80%）→ 标 WARNING，要求补测试
- 维度 7 只审「测试质量」（新增逻辑有测试 + 测试真断言行为），不审「测试是否通过」

### Step 3: 逐文件审查（12 维度）

| #   | 维度        | 二级检查项                                                        |
| --- | --------- | ----------------------------------------------------------- |
| 1   | **架构一致性** | Proto 在 api-proto/？服务间走 gRPC？有无直连其他服务 DB？跨服务变更是否破坏 gRPC/Proto 兼容？ |
| 2   | **设计一致性** | 变更与 design.md 一致？数据模型正确？业务流程正确？ |
| 3   | **规范遵循**  | int64 ID → `[jstype=JS_STRING]`/`json:",string"`？错误码 5 位？错误码语义一致？ |
| 4   | **代码质量**  | 逻辑正确？空指针防护？循环边界？异常捕获粒度？资源释放（连接/文件）？错误码语义一致性？ |
| 5   | **安全性**   | 无硬编码密钥？SQL 参数化？输入校验？无敏感信息打日志？权限越权校验？敏感数据脱敏？接口限流防刷？配置项加密？ |
| 6   | **复用性**   | 有无可复用 common 库代码？有无重复逻辑？ |
| 7   | **测试覆盖**  | 新增逻辑有测试？测试真正断言行为（非空跑）？边界/异常分支有断言？（不重跑 build/vet/test） |
| 8   | **变更完整性** | CHANGELOG 更新？Proto 变更记录？Migration 提供？ |
| 9   | **可观测性**  | 关键链路日志埋点？错误日志完整上下文？核心指标埋点？告警阈值配置？ |
| 10  | **依赖变更**  | 新依赖 License 合规？已知安全漏洞？版本与全局技术栈兼容？ |
| 11  | **配置变更**  | 敏感配置加密？多环境区分？默认值兜底？灰度生效策略？ |
| 12  | **记忆遵守**  | 见下方 M1-M4 |

### Step 4: 记忆遵守（M1-M4）

**M1 收集引用**: Grep 变更文件中 `// SEE: [[` 注释，提取所有 memory-slug。

**M2 验证准确性**: 对每个引用：

- slug 文件存在？→ 不存在则 🔴 CRITICAL
- 代码遵守了记忆指导？→ 未遵守则 🔴 CRITICAL
- 记忆确实适用于此上下文？→ 虚假匹配则 🟡 WARNING

**M3 检查遗漏**（降低误判，不单靠 triggers 精确匹配）:

- 从变更描述/git diff 提取技术关键词
- 用 `bash .harness/scripts/knowledge-load.sh --service <名> --keywords "<关键词>"` 做打分召回（severity + service_match + keyword_match + recency），再核对场景字段，非纯 triggers 精确匹配
- 必须匹配记忆的 `type`/场景字段：pitfall 类仅当技术栈匹配才告警
- must-follow 遗漏 → 🔴 CRITICAL；should-follow 遗漏 → 🟡 WARNING
- **冲突处理**：多条记忆冲突时，优先级「架构原则 > 踩坑经验 > 最佳实践」；无法判定 → 标 INFO，提交 Owner 裁决

**M4 元数据更新建议**（Reviewer 保持只读，不落库）:

- 正确应用的记忆，在报告「记忆更新建议」章节列出需更新的条目（slug + 建议的 last_applied/apply_count 增量）
- 由 Owner 在评审通过后调用记忆管理 Agent 统一落库，Reviewer 不直接改 MEMORY.md

### Step 5: 分级

| 级别          | 判据                                | 对齐计划评审 |
|:-----------:| --------------------------------- |:---:|
| 🔴 CRITICAL | 安全漏洞、数据丢失风险、架构违反、业务不可用、must-follow 记忆遗漏 | MUST FIX |
| 🟡 WARNING  | 逻辑错误、性能问题、规范违反、should-follow 记忆遗漏 | SHOULD FIX |
| 🔵 NOTE     | 代码风格、命名建议、文档缺失                    | INFO |

## 产出

写入 `services/<name>/_review.md`：

```markdown
# Code Review — <service-name>

**审查时间**: YYYY-MM-DD HH:MM
**审查范围**: <描述>
**审查者**: Code Reviewer

## 摘要
- 变更文件: N
- 🔴 CRITICAL: N / 🟡 WARNING: N / 🔵 NOTE: N

## 发现

### 🔴 CRITICAL
| # | 文件:行号 | 问题 | 修复建议 |
|---|----------|------|---------|

### 🟡 WARNING
| # | 文件:行号 | 问题 | 建议 |
|---|----------|------|------|

### 🔵 NOTE
| # | 文件:行号 | 建议 |
|---|----------|------|

## 架构一致性检查
- [ ] Proto 规范 / [ ] gRPC 通信 / [ ] 错误码 / [ ] Snowflake ID

## 变更完整性检查
- [ ] CHANGELOG / [ ] Proto CHANGELOG / [ ] Migration / [ ] design.md

## 记忆遵守检查
- [ ] `// SEE: [[...]]` 引用已验证（共 N 处）
- [ ] 未遗漏 must-follow 记忆
- [ ] 适用记忆已列入「记忆更新建议」章节（由 Owner 落库）

---
VERDICT: PASS / FAIL
---
```

### 报告自检（产出后必做，防空泛意见）

报告生成后逐项校验：
1. 所有 CRITICAL / MUST FIX 问题必须定位到具体「文件 + 行号/章节」
2. 所有问题必须附带可落地的修复建议，禁止「优化代码」「提升性能」等空泛表述
3. 12 个维度逐项标注是否检查完成，不得遗漏维度
4. REVISION 视角的问题标注状态（待修复/已修复/已验证），供下轮校验

## VERDICT

```
PASS — 无 CRITICAL 问题，WARNING/NOTE 不阻塞
FAIL — 存在 ≥1 个 CRITICAL，必须修复后重新审查
```

**增量复审范围**：
- 默认：仅审上一轮 CRITICAL/WARNING 对应的代码行 + 修复改动波及的相邻逻辑
- 强制回归：修复涉及公共函数/核心链路/数据模型 → 扩大到所有调用方
- 复审报告须区分「本轮新增问题」和「上一轮问题修复验证结果」，并引用上一轮问题编号

## 特殊规则

1. **不审查自己产出** — AI 生成的代码以同等标准审查（更容易有边界遗漏）
2. **跨服务变更加倍谨慎** — 额外检查 gRPC 兼容性和 Proto 破坏性变更
3. **Migration 专项检查（必审）**：
   - 脚本幂等（重复执行不报错、不产生脏数据）
   - 大表 DDL 是否在线变更、是否评估锁表时间
   - 字段变更兼容旧版本代码（新增字段有默认值、删除字段有兼容期）
   - 配套回滚脚本/数据订正脚本
   - 索引变更评估性能影响、无重复索引
4. **跨服务兼容性专项（必审）**：
   - Proto 字段向后兼容（不删除/修改字段号、不修改字段类型）
   - 接口语义变更有无版本号隔离
   - 错误码变更是否影响上游错误处理逻辑
   - 依赖的下游接口版本是否匹配
5. **废弃/删除类变更专项（必审）**：
   - 删除接口/字段前，是否检查所有调用方已下线
   - 有无兼容过渡期与灰度下线方案
   - 废弃代码有无明确下线时间节点
   - 数据删除有无备份与回滚方案
6. **信任但验证** — "小改动"不代表真的小，git diff 行数不是复杂度指标
7. **记忆遵守是硬约束** — 遗漏 must-follow 记忆 = 架构违反 = CRITICAL

## 关联

- QA Skill：`.harness/skills/qa.md`（先于 Review 执行；Review 加载 QA 报告，见「QA 联动」）
- 项目编码规范：`.harness/rules/项目编码规范.md`
- Harness Pipeline：`.harness/workflows/harness-pipeline.js`
