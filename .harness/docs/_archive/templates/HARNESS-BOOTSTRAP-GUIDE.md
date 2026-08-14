# Harness Pipeline 建设操作手册

**版本**: 1.0  
**日期**: 2026-06-23  
**适用范围**: 任何技术栈的软件项目  
**目标读者**: AI Agent (Claude Code / Cursor / etc.)

---

## 📋 手册说明

本手册是一份**可操作的指导文档**，AI Agent 阅读本手册后，可以在任何项目中自动建立完整的开发流水线（Harness Pipeline）。

### 手册结构

1. **概念篇** - 理解 Harness Pipeline 的核心理念
2. **架构篇** - 了解整体架构和目录结构
3. **实施篇** - 逐步建立流水线的操作步骤
4. **模板篇** - 可复用的文件模板
5. **验证篇** - 验证流水线是否正确运行

---

## 第一部分：概念篇

### 1.1 什么是 Harness Pipeline？

Harness Pipeline 是一个**AI 驱动的开发流水线管理系统**，核心理念是：

> "让 AI 按照严格的工程纪律执行开发任务，而不是随意编码"

### 1.2 四大支柱

```
┌─────────────────────────────────────────────────────────────┐
│                    Harness Pipeline                         │
│                                                             │
│  1. Mechanization    - 机械化：标准化、可重复、自动化      │
│  2. Accountability   - 可追溯：每个决策都有记录             │
│  3. Composition      - 可组合：模块化、可扩展               │
│  4. Memory-Driven    - 记忆驱动：从错误中学习               │
└─────────────────────────────────────────────────────────────┘
```

### 1.3 为什么需要 Harness？

**问题**:
- AI 容易"偷懒"，跳过测试直接写代码
- 代码变更缺乏追溯，难以回溯决策
- 多个 Agent 协作时缺乏协调机制
- 经验无法沉淀，重复犯同样的错误

**解决方案**:
- 通过 6 个 Phase 强制 AI 遵循工程纪律
- 变更追踪机制记录所有决策
- Workflow 系统协调多 Agent 分工
- Memory 系统持久化经验教训

---

## 第二部分：架构篇

### 2.1 目录结构

```
project-root/
├── .harness/                          # Harness 核心目录
│   ├── agents/                        # Agent 定义
│   │   ├── owner-agent.md            # 主 Agent（你的入口）
│   │   └── subagents/                # 子 Agent 定义
│   │       ├── architecture-designer.md
│   │       ├── code-reviewer.md
│   │       ├── requirement-analyst.md
│   │       └── test-engineer.md
│   │
│   ├── skills/                        # 技能脚本（Slash Commands）
│   │   ├── qa/                       # QA 相关技能
│   │   │   └── scripts/
│   │   │       ├── harness-checks.sh
│   │   │       └── tdd-evidence-validator.sh
│   │   └── design-sync/              # 设计同步技能
│   │
│   ├── workflows/                     # 工作流脚本
│   │   ├── harness-pipeline.js       # 核心流水线
│   │   ├── code-review-workflow.js   # Code Review 流程
│   │   └── tdd-workflow.js           # TDD 流程
│   │
│   ├── knowledge/                     # 知识库
│   │   ├── memory/                   # 持久化记忆
│   │   │   ├── MEMORY.md            # 记忆索引
│   │   │   ├── .memory-index.json   # 记忆索引（机器可读）
│   │   │   └── *.md                 # 具体记忆文件
│   │   └── docs/                    # 项目文档
│   │       └── graph-context.md     # 知识图谱上下文
│   │
│   ├── changes/                       # 变更追踪
│   │   ├── INDEX.md                  # 变更索引
│   │   ├── TEMPLATE/                 # 变更模板
│   │   │   ├── phase1-requirement.md
│   │   │   ├── phase2-design.md
│   │   │   ├── phase3-api-design.md
│   │   │   ├── phase4-db-design.md
│   │   │   ├── phase5-implementation.md
│   │   │   └── phase6-integration.md
│   │   └── <change-name>/           # 具体变更目录
│   │       ├── summary.md
│   │       ├── phase*.md
│   │       └── _qa.md
│   │
│   ├── scripts/                       # 工具脚本
│   │   ├── harness-gate-check-v2.sh  # 门禁检查
│   │   └── tdd-evidence-validator.sh # TDD 证据验证
│   │
│   ├── tasks/                         # 任务追踪
│   │   └── P0-*.md                   # 高优先级任务
│   │
│   └── templates/                     # 模板文件
│       ├── BOOTSTRAP-GUIDE.md        # 本手册
│       └── *.template.js             # 可复用脚本模板
│
├── tools/                             # 开发工具
│   ├── install-testing-tools.sh      # 安装测试工具
│   ├── generate-mocks.sh             # 生成 Mock
│   ├── generate-grpc-mocks.sh        # 生成 gRPC Mock
│   ├── new-service-with-tests.sh     # 创建新服务（含测试）
│   ├── install-hooks.sh              # 安装 Git hooks
│   └── pre-commit.sh                 # Pre-commit hook
│
├── .github/                           # GitHub 配置
│   └── workflows/
│       └── test.yml                  # CI/CD 配置
│
└── [项目源代码目录]
```

### 2.2 核心组件关系

```
┌─────────────────────────────────────────────────────────────┐
│                     Owner Agent                             │
│                  (主 Agent 入口)                            │
└──────────────┬──────────────────────────────────────────────┘
               │
               ├─→ Harness Pipeline (6 Phases)
               │   └─→ Phase Gate Checks
               │
               ├─→ Workflows (多 Agent 协作)
               │   ├─→ Subagent: Architecture Designer
               │   ├─→ Subagent: Code Reviewer
               │   └─→ Subagent: Test Engineer
               │
               ├─→ Skills (Slash Commands)
               │   ├─→ /qa - 质量检查
               │   └─→ /design-sync - 设计同步
               │
               ├─→ Memory System
               │   └─→ 持久化记忆 → 知识沉淀
               │
               └─→ Change Tracking
                   └─→ 变更追踪 → 可追溯性
```

---

## 第三部分：实施篇

### 3.1 前置准备

**检查项**:
- [ ] 确认项目根目录
- [ ] 确认技术栈（Go / Node.js / Python / Java / etc.）
- [ ] 确认测试框架（如果已有）
- [ ] 确认 Git 已初始化

### 3.2 实施步骤

#### Step 1: 创建 Harness 目录结构

```bash
# 执行此命令创建所有必需的目录
mkdir -p .harness/{agents/subagents,skills/{qa/scripts,design-sync},workflows,knowledge/{memory,docs},changes/TEMPLATE,scripts,tasks,templates}
mkdir -p tools
mkdir -p .github/workflows

# 验证
tree .harness -L 2
```

#### Step 2: 创建 Owner Agent

**文件**: `.harness/agents/owner-agent.md`

```markdown
# Owner Agent

You are the Owner Agent for this project.

## Your Responsibilities

1. **Execute Harness Pipeline**: Follow the 6-phase process strictly
2. **Delegate to Subagents**: Use workflows for complex tasks
3. **Track Changes**: Document all decisions in `.harness/changes/`
4. **Build Memory**: Save lessons learned to `.harness/knowledge/memory/`

## Harness Pipeline (6 Phases)

### Phase 1: Requirement Analysis
- Create: `changes/<name>/phase1-requirement.md`
- Gate: Requirements clear and complete

### Phase 2: Technical Design
- Create: `changes/<name>/phase2-design.md`
- Gate: Architecture reviewed

### Phase 3: API Design
- Create: `changes/<name>/phase3-api-design.md`
- Gate: API contract defined

### Phase 4: Database Design
- Create: `changes/<name>/phase4-db-design.md`
- Gate: Schema reviewed

### Phase 5: Implementation
- Create: `changes/<name>/phase5-implementation.md`
- **Gate Check**: Run `bash .harness/scripts/harness-gate-check-v2.sh --phase 5 --change <name>`
- Must have: Code + Tests + TDD Evidence

### Phase 6: Integration
- Create: `changes/<name>/phase6-integration.md`
- **Gate Check**: Run `bash .harness/scripts/harness-gate-check-v2.sh --phase 6 --change <name>`
- Must have: Integration tests + QA report

## Skills Available

- `/qa` - Run quality checks
- `/design-sync` - Sync with design system

## Workflow Usage

For complex tasks, use workflows:

\`\`\`javascript
const result = await workflow('code-review', {
  files: ['src/service.go'],
  focus: 'security'
});
\`\`\`

## Memory Management

After completing a task:
1. Identify lessons learned
2. Save to `.harness/knowledge/memory/<name>.md`
3. Update `MEMORY.md` index

## Change Tracking

Every change must have:
1. Directory: `.harness/changes/<change-name>/`
2. Summary: `summary.md`
3. Phase files: `phase*.md`
4. QA report: `_qa.md`
```

#### Step 3: 创建子 Agent 定义

**3.1 需求分析师**

**文件**: `.harness/agents/subagents/requirement-analyst.md`

```markdown
# Requirement Analyst

You are a Requirement Analyst. Your job is to clarify and document requirements.

## Your Tasks

1. **Interview**: Ask clarifying questions
2. **Document**: Write clear requirement specifications
3. **Validate**: Ensure requirements are testable

## Output Format

\`\`\`markdown
# Requirement Specification

## User Story
As a [role], I want [feature], so that [benefit].

## Acceptance Criteria
- [ ] Criterion 1
- [ ] Criterion 2

## Non-Functional Requirements
- Performance: < 200ms response time
- Security: OAuth 2.0 authentication
\`\`\`

## Return to Owner

When done, return structured JSON:
\`\`\`json
{
  "requirements": [...],
  "acceptance_criteria": [...],
  "risks": [...]
}
\`\`\`
```

**3.2 架构设计师**

**文件**: `.harness/agents/subagents/architecture-designer.md`

```markdown
# Architecture Designer

You design system architecture and technical solutions.

## Your Tasks

1. **Analyze**: Understand requirements
2. **Design**: Propose architecture
3. **Document**: Create design documents

## Output Format

\`\`\`markdown
# Technical Design

## Architecture Overview
[Architecture diagram in mermaid]

## Component Design
### Component A
- Responsibility: ...
- Dependencies: ...

## Data Flow
[Sequence diagram]

## Technology Choices
| Component | Technology | Rationale |
|-----------|-----------|-----------|
| ...       | ...       | ...       |
\`\`\`

## Return to Owner

\`\`\`json
{
  "architecture": {...},
  "components": [...],
  "dependencies": [...]
}
\`\`\`
```

**3.3 代码审查员**

**文件**: `.harness/agents/subagents/code-reviewer.md`

```markdown
# Code Reviewer

You review code for quality, security, and best practices.

## Your Tasks

1. **Check**: Code quality, security, performance
2. **Verify**: Tests exist and pass
3. **Suggest**: Improvements

## Review Checklist

- [ ] Code follows project conventions
- [ ] Tests exist and cover main scenarios
- [ ] No security vulnerabilities
- [ ] Performance considerations
- [ ] Documentation updated

## Output Format

\`\`\`json
{
  "verdict": "PASS" | "FAIL",
  "issues": [
    {
      "file": "path/to/file.go",
      "line": 42,
      "severity": "high",
      "message": "..."
    }
  ],
  "suggestions": [...]
}
\`\`\`
```

**3.4 测试工程师**

**文件**: `.harness/agents/subagents/test-engineer.md`

```markdown
# Test Engineer

You ensure code quality through comprehensive testing.

## Your Tasks

1. **Write Tests**: Unit, integration, e2e
2. **Verify TDD**: Check RED-GREEN evidence
3. **Check Coverage**: Ensure >= target coverage

## Test Template

\`\`\`go
func TestXxx(t *testing.T) {
    tests := []struct {
        name    string
        input   ...
        want    ...
        wantErr bool
    }{
        {name: "success", ...},
        {name: "error", ...},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test logic
        })
    }
}
\`\`\`

## Output Format

\`\`\`json
{
  "tests_created": 5,
  "coverage": "65%",
  "tdd_evidence": "complete",
  "verdict": "PASS"
}
\`\`\`
```

#### Step 4: 创建核心流水线脚本

**文件**: `.harness/workflows/harness-pipeline.js`

```javascript
// 见后续"可复用脚本"章节
// 这是完整的 harness-pipeline.js 实现
```

#### Step 5: 创建质量门禁脚本

**文件**: `.harness/scripts/harness-gate-check-v2.sh`

```bash
#!/usr/bin/env bash
# 见后续"可复用脚本"章节
```

#### Step 6: 创建测试工具脚本（技术栈特定）

根据项目技术栈，创建相应的测试工具：

**Go 项目**:
```bash
# tools/install-testing-tools.sh
go install github.com/golang/mock/mockgen@latest
go get github.com/stretchr/testify
```

**Node.js 项目**:
```bash
# tools/install-testing-tools.sh
npm install --save-dev jest @testing-library/react
npm install --save-dev @types/jest
```

**Python 项目**:
```bash
# tools/install-testing-tools.sh
pip install pytest pytest-cov pytest-mock
```

#### Step 7: 创建 CI/CD 配置

**文件**: `.github/workflows/test.yml`

```yaml
# 见后续"可复用脚本"章节
```

#### Step 8: 创建变更追踪模板

**文件**: `.harness/changes/TEMPLATE/phase1-requirement.md`

```markdown
# Phase 1: Requirement Analysis

## User Story

As a [role], I want [feature], so that [benefit].

## Acceptance Criteria

- [ ] Criterion 1
- [ ] Criterion 2

## Non-Functional Requirements

- Performance: ...
- Security: ...

## Risks

1. Risk 1 - Mitigation strategy

## Sign-off

- [ ] Requirements reviewed
- [ ] Acceptance criteria defined
- [ ] Ready for Phase 2
```

（其他 phase 模板类似，见后续章节）

#### Step 9: 初始化记忆系统

**文件**: `.harness/knowledge/memory/MEMORY.md`

```markdown
# Memory Index

This file indexes all persistent memories.

## Project Memories

- [project-setup](project-setup.md) - Initial project configuration

## User Preferences

- [coding-style](coding-style.md) - Preferred coding conventions

## Lessons Learned

- [lesson-001](lesson-001.md) - First major lesson

---

Last updated: 2026-06-23
```

#### Step 10: 创建变更追踪索引

**文件**: `.harness/changes/INDEX.md`

```markdown
# Change Tracking Index

All changes to this project are tracked here.

## Active Changes

- None

## Completed Changes

- None

---

To create a new change:
1. mkdir .harness/changes/<change-name>
2. Copy templates from TEMPLATE/
3. Add entry here
```

### 3.3 验证安装

运行以下命令验证 Harness 安装是否正确：

```bash
# 检查目录结构
tree .harness -L 2

# 检查核心文件
ls -lh .harness/agents/owner-agent.md
ls -lh .harness/workflows/harness-pipeline.js
ls -lh .harness/scripts/harness-gate-check-v2.sh

# 测试门禁脚本
bash .harness/scripts/harness-gate-check-v2.sh --help
```

---

## 第四部分：可复用脚本

### 4.1 核心流水线脚本

**文件**: `.harness/workflows/harness-pipeline.js`

（此文件将在下一个响应中提供完整内容）

### 4.2 门禁检查脚本

**文件**: `.harness/scripts/harness-gate-check-v2.sh`

（此文件将在下一个响应中提供完整内容）

### 4.3 CI/CD 配置模板

**文件**: `.github/workflows/test.yml`

（此文件将在下一个响应中提供完整内容）

---

## 第五部分：使用指南

### 5.1 日常开发流程

```
开发者请求 → Owner Agent
                │
                ├─→ 创建 Change 目录
                ├─→ Phase 1: 需求分析
                ├─→ Phase 2: 技术设计
                ├─→ Phase 3: API 设计
                ├─→ Phase 4: 数据库设计
                ├─→ Phase 5: 编码实现 (Gate Check)
                ├─→ Phase 6: 集成验证 (Gate Check)
                └─→ 完成，更新记忆
```

### 5.2 Quality Gate 说明

**Phase 5 Gate**:
- 必须有代码
- 必须有测试
- 必须有 TDD 证据（RED-GREEN）
- `_qa.md` 报告完整

**Phase 6 Gate**:
- 集成测试通过
- 文档更新
- Code Review 完成

### 5.3 Memory 使用

**何时创建 Memory**:
- 发现新的最佳实践
- 遇到坑并找到解决方案
- 用户明确的偏好
- 项目特定的约定

**Memory 格式**:
```markdown
---
name: memory-name
description: One-line summary
metadata:
  type: user | feedback | project | reference
---

Content here.

**Why**: Explanation of the lesson.

**How to apply**: Actionable guidance.

Related: [[other-memory]]
```

---

## 第六部分：技术栈适配

### 6.1 Go 项目

```bash
# 测试工具
mockgen, testify, gomock

# 门禁检查重点
- 所有 *_logic.go 都有 *_logic_test.go
- 覆盖率 >= 30%
- go fmt 检查
```

### 6.2 Node.js 项目

```bash
# 测试工具
jest, @testing-library

# 门禁检查重点
- 所有 *.ts 都有 *.test.ts
- 覆盖率 >= 70%
- eslint 检查
```

### 6.3 Python 项目

```bash
# 测试工具
pytest, pytest-cov

# 门禁检查重点
- 所有 .py 都有 test_*.py
- 覆盖率 >= 80%
- pylint 检查
```

---

## 第七部分：常见问题

### Q1: 如何跳过某个 Phase？

**A**: 不建议跳过。如果必须，在 summary.md 中说明原因。

### Q2: Gate Check 失败怎么办？

**A**: 修复问题，重新运行。不要使用 `--force`。

### Q3: 如何添加新的 Subagent？

**A**: 
1. 创建 `.harness/agents/subagents/<name>.md`
2. 在 `owner-agent.md` 中引用
3. 在 workflow 中调用

### Q4: Memory 太多怎么办？

**A**: 定期整理，合并相似的记忆，归档过时的记忆。

---

## 第八部分：检查清单

### 安装完成检查清单

- [ ] `.harness/` 目录结构完整
- [ ] `owner-agent.md` 已创建
- [ ] 至少 3 个 subagent 已定义
- [ ] `harness-pipeline.js` 已配置
- [ ] 门禁脚本可执行
- [ ] CI/CD 配置已创建
- [ ] 测试工具已安装
- [ ] Pre-commit hooks 已配置
- [ ] 变更追踪模板已准备
- [ ] Memory 系统已初始化

### 运行验证

```bash
# 1. 测试门禁脚本
bash .harness/scripts/harness-gate-check-v2.sh --help

# 2. 测试工作流（模拟）
# (需要在 Claude Code 中执行)

# 3. 创建测试 Change
mkdir .harness/changes/test-change
cp .harness/changes/TEMPLATE/* .harness/changes/test-change/

# 4. 运行 QA 检查
bash .harness/skills/qa/scripts/harness-checks.sh
```

---

## 附录 A: 快速启动命令

```bash
# 一键创建 Harness 结构
bash << 'EOF'
mkdir -p .harness/{agents/subagents,skills/{qa/scripts,design-sync},workflows,knowledge/{memory,docs},changes/TEMPLATE,scripts,tasks,templates}
mkdir -p tools .github/workflows
echo "✅ Harness structure created"
EOF

# 下载模板文件
# (从本项目复制，或使用提供的模板)
```

---

## 附录 B: 术语表

| 术语 | 含义 |
|------|------|
| **Owner Agent** | 主 Agent，负责执行 Harness Pipeline |
| **Subagent** | 子 Agent，负责特定任务（设计、审查、测试等） |
| **Phase** | 流水线的一个阶段（共 6 个） |
| **Gate Check** | 阶段门禁检查，确保质量 |
| **Change** | 一次完整的变更（从需求到交付） |
| **Memory** | 持久化的经验教训 |
| **Skill** | 可调用的命令（Slash Command） |
| **Workflow** | 多 Agent 协作的工作流 |

---

## 附录 C: 参考资料

- 本项目实施案例: `community-and-home`
- Harness 理念文档: `.harness/docs/harness-philosophy.md`
- Memory 管理指南: `.harness/knowledge/memory/README.md`

---

**手册版本**: 1.0  
**最后更新**: 2026-06-23  
**维护者**: Original Project Team

---

**END OF BOOTSTRAP GUIDE**

下一步：请查看"可复用脚本"文档，获取完整的脚本实现。
