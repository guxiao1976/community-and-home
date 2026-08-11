#!/usr/bin/env bash
#
# quick-deploy.sh - 一键部署 Harness Pipeline 到新项目
#
# Usage:
#   bash quick-deploy.sh <source-template-dir> [target-project-dir]
#
# Example:
#   bash quick-deploy.sh /path/to/community-and-home/.harness/templates /path/to/new-project
#

set -euo pipefail

TEMPLATE_DIR="${1:-}"
TARGET_DIR="${2:-.}"

if [[ -z "$TEMPLATE_DIR" ]]; then
    echo "Usage: $0 <template-dir> [target-dir]"
    echo ""
    echo "Example:"
    echo "  $0 /path/to/old-project/.harness/templates /path/to/new-project"
    exit 1
fi

if [[ ! -d "$TEMPLATE_DIR" ]]; then
    echo "❌ 模板目录不存在: $TEMPLATE_DIR"
    exit 1
fi

echo "╔═══════════════════════════════════════════════════════════════╗"
echo "║                                                               ║"
echo "║         🚀 Harness Pipeline 一键部署工具                      ║"
echo "║                                                               ║"
echo "╚═══════════════════════════════════════════════════════════════╝"
echo ""
echo "源目录: $TEMPLATE_DIR"
echo "目标目录: $TARGET_DIR"
echo ""

# ==================== Step 1: 创建目录结构 ====================

echo "[1/8] 创建目录结构..."

mkdir -p "$TARGET_DIR/.harness/agents/subagents"
mkdir -p "$TARGET_DIR/.harness/skills/qa/scripts"
mkdir -p "$TARGET_DIR/.harness/workflows"
mkdir -p "$TARGET_DIR/.harness/knowledge/memory"
mkdir -p "$TARGET_DIR/.harness/knowledge/docs"
mkdir -p "$TARGET_DIR/.harness/changes/TEMPLATE"
mkdir -p "$TARGET_DIR/.harness/scripts"
mkdir -p "$TARGET_DIR/.harness/tasks"
mkdir -p "$TARGET_DIR/.harness/templates"
mkdir -p "$TARGET_DIR/tools"
mkdir -p "$TARGET_DIR/.github/workflows"

echo "  ✅ 目录结构创建完成"

# ==================== Step 2: 复制核心脚本 ====================

echo ""
echo "[2/8] 复制核心脚本..."

# 流水线脚本
if [[ -f "$TEMPLATE_DIR/harness-pipeline.template.js" ]]; then
    cp "$TEMPLATE_DIR/harness-pipeline.template.js" \
       "$TARGET_DIR/.harness/workflows/harness-pipeline.js"
    echo "  ✅ harness-pipeline.js"
else
    echo "  ⚠️  harness-pipeline.template.js 不存在，跳过"
fi

# 门禁检查脚本
if [[ -f "$TEMPLATE_DIR/harness-gate-check.template.sh" ]]; then
    cp "$TEMPLATE_DIR/harness-gate-check.template.sh" \
       "$TARGET_DIR/.harness/scripts/harness-gate-check-v2.sh"
    chmod +x "$TARGET_DIR/.harness/scripts/harness-gate-check-v2.sh"
    echo "  ✅ harness-gate-check-v2.sh"
else
    echo "  ⚠️  harness-gate-check.template.sh 不存在，跳过"
fi

# Pre-commit hook
if [[ -f "$TEMPLATE_DIR/../tools/pre-commit.sh" ]]; then
    cp "$TEMPLATE_DIR/../tools/pre-commit.sh" \
       "$TARGET_DIR/tools/pre-commit.sh"
    chmod +x "$TARGET_DIR/tools/pre-commit.sh"
    echo "  ✅ pre-commit.sh"
fi

# Install hooks script
if [[ -f "$TEMPLATE_DIR/../tools/install-hooks.sh" ]]; then
    cp "$TEMPLATE_DIR/../tools/install-hooks.sh" \
       "$TARGET_DIR/tools/install-hooks.sh"
    chmod +x "$TARGET_DIR/tools/install-hooks.sh"
    echo "  ✅ install-hooks.sh"
fi

# ==================== Step 3: 复制 Agent 定义 ====================

echo ""
echo "[3/8] 复制 Agent 定义..."

# Owner Agent
cat > "$TARGET_DIR/.harness/agents/owner-agent.md" << 'EOF'
# Owner Agent

You are the Owner Agent for this project.

## Your Responsibilities

1. Execute Harness Pipeline (6 phases)
2. Delegate to Subagents via workflows
3. Track changes in `.harness/changes/`
4. Build memory in `.harness/knowledge/memory/`

## Harness Pipeline

### Phase 1-4: Design
- Requirement → Technical Design → API Design → DB Design

### Phase 5: Implementation (Gate Check Required)
- Code + Tests + TDD Evidence
- Run: `bash .harness/scripts/harness-gate-check-v2.sh --phase 5 --change <name>`

### Phase 6: Integration (Gate Check Required)
- Integration Tests + QA Report
- Run: `bash .harness/scripts/harness-gate-check-v2.sh --phase 6 --change <name>`

## Change Tracking

Every change must have:
- Directory: `.harness/changes/<change-name>/`
- Files: `summary.md`, `phase*.md`, `_qa.md`

## Memory Management

Save lessons learned to `.harness/knowledge/memory/<name>.md`
Update `MEMORY.md` index.
EOF

echo "  ✅ owner-agent.md"

# Subagents (简化版本)
cat > "$TARGET_DIR/.harness/agents/subagents/requirement-analyst.md" << 'EOF'
# Requirement Analyst

Clarify and document requirements.

## Output Format

```markdown
# Requirements

## User Story
As a [role], I want [feature], so that [benefit].

## Acceptance Criteria
- [ ] Criterion 1
- [ ] Criterion 2
```
EOF

echo "  ✅ requirement-analyst.md"

# ==================== Step 4: 初始化 Memory 系统 ====================

echo ""
echo "[4/8] 初始化 Memory 系统..."

cat > "$TARGET_DIR/.harness/knowledge/memory/MEMORY.md" << EOF
# Memory Index

This file indexes all persistent memories for this project.

## Project Memories

(To be added)

## User Preferences

(To be added)

## Lessons Learned

(To be added)

---

Last updated: $(date +%Y-%m-%d)
EOF

echo "  ✅ MEMORY.md"

# ==================== Step 5: 初始化变更追踪 ====================

echo ""
echo "[5/8] 初始化变更追踪..."

cat > "$TARGET_DIR/.harness/changes/INDEX.md" << 'EOF'
# Change Tracking Index

All changes to this project are tracked here.

## Active Changes

(None)

## Completed Changes

(None)

---

To create a new change:
1. `mkdir .harness/changes/<change-name>`
2. Copy templates from `TEMPLATE/`
3. Add entry here
EOF

echo "  ✅ INDEX.md"

# 创建变更模板
cat > "$TARGET_DIR/.harness/changes/TEMPLATE/phase5-implementation.md" << 'EOF'
# Phase 5: Implementation

## Code Changes

### Files Modified
- `path/to/file.go` - Description

## Tests Created

### Test Files
- `path/to/file_test.go` - Test cases

## TDD Evidence

### RED (Test Failing)
```
[Paste test failure output]
```

### GREEN (Test Passing)
```
[Paste test success output]
```

## QA Report

See `_qa.md`

## Sign-off

- [ ] Code complete
- [ ] Tests written
- [ ] TDD evidence captured
- [ ] Ready for Phase 6
EOF

echo "  ✅ phase5-implementation.md 模板"

# ==================== Step 6: 检测技术栈并复制工具 ====================

echo ""
echo "[6/8] 检测技术栈..."

TECH_STACK="unknown"

if [[ -f "$TARGET_DIR/go.mod" ]]; then
    TECH_STACK="go"
elif [[ -f "$TARGET_DIR/package.json" ]]; then
    TECH_STACK="nodejs"
elif [[ -f "$TARGET_DIR/requirements.txt" ]] || [[ -f "$TARGET_DIR/setup.py" ]]; then
    TECH_STACK="python"
fi

echo "  检测到技术栈: $TECH_STACK"

# 根据技术栈复制 CI/CD 配置
if [[ "$TECH_STACK" != "unknown" ]] && [[ -f "$TEMPLATE_DIR/ci-cd-${TECH_STACK}.template.yml" ]]; then
    cp "$TEMPLATE_DIR/ci-cd-${TECH_STACK}.template.yml" \
       "$TARGET_DIR/.github/workflows/test.yml"
    echo "  ✅ CI/CD 配置 (${TECH_STACK})"
elif [[ -f "$TEMPLATE_DIR/ci-cd.template.yml" ]]; then
    cp "$TEMPLATE_DIR/ci-cd.template.yml" \
       "$TARGET_DIR/.github/workflows/test.yml"
    echo "  ✅ CI/CD 配置 (通用)"
fi

# ==================== Step 7: 创建快速启动指南 ====================

echo ""
echo "[7/8] 创建快速启动指南..."

cat > "$TARGET_DIR/.harness/QUICKSTART.md" << 'EOF'
# Harness Pipeline 快速启动指南

## 1. 安装依赖

```bash
# 安装测试工具
bash tools/install-testing-tools.sh

# 安装 Git hooks
bash tools/install-hooks.sh
```

## 2. 创建第一个变更

```bash
# 创建变更目录
mkdir .harness/changes/my-first-change

# 复制模板
cp .harness/changes/TEMPLATE/* .harness/changes/my-first-change/

# 开始开发（遵循 6 个 Phase）
```

## 3. 运行质量检查

```bash
# Phase 5 门禁检查
bash .harness/scripts/harness-gate-check-v2.sh --phase 5 --change my-first-change

# Phase 6 门禁检查
bash .harness/scripts/harness-gate-check-v2.sh --phase 6 --change my-first-change
```

## 4. 查看完整文档

- Bootstrap 指南: `.harness/templates/HARNESS-BOOTSTRAP-GUIDE.md`
- 可复用脚本: `.harness/templates/REUSABLE-SCRIPTS.md`
EOF

echo "  ✅ QUICKSTART.md"

# ==================== Step 8: 复制文档 ====================

echo ""
echo "[8/8] 复制文档..."

if [[ -f "$TEMPLATE_DIR/HARNESS-BOOTSTRAP-GUIDE.md" ]]; then
    cp "$TEMPLATE_DIR/HARNESS-BOOTSTRAP-GUIDE.md" \
       "$TARGET_DIR/.harness/templates/"
    echo "  ✅ HARNESS-BOOTSTRAP-GUIDE.md"
fi

if [[ -f "$TEMPLATE_DIR/REUSABLE-SCRIPTS.md" ]]; then
    cp "$TEMPLATE_DIR/REUSABLE-SCRIPTS.md" \
       "$TARGET_DIR/.harness/templates/"
    echo "  ✅ REUSABLE-SCRIPTS.md"
fi

# ==================== 完成 ====================

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "✅ Harness Pipeline 部署完成！"
echo ""
echo "📁 已创建的目录:"
echo "  .harness/          - Harness 核心目录"
echo "  .harness/agents/   - Agent 定义"
echo "  .harness/workflows/ - 工作流脚本"
echo "  .harness/knowledge/ - 知识库"
echo "  .harness/changes/  - 变更追踪"
echo "  tools/             - 开发工具"
echo ""
echo "📄 核心文件:"
echo "  .harness/agents/owner-agent.md             - 主 Agent"
echo "  .harness/workflows/harness-pipeline.js     - 流水线脚本"
echo "  .harness/scripts/harness-gate-check-v2.sh  - 门禁检查"
echo "  .harness/QUICKSTART.md                     - 快速启动指南"
echo ""
echo "🎯 下一步:"
echo ""
echo "  1. 查看快速启动指南:"
echo "     cat .harness/QUICKSTART.md"
echo ""
echo "  2. 安装测试工具:"
echo "     bash tools/install-testing-tools.sh"
echo ""
echo "  3. 安装 Git hooks:"
echo "     bash tools/install-hooks.sh"
echo ""
echo "  4. 测试门禁脚本:"
echo "     bash .harness/scripts/harness-gate-check-v2.sh --help"
echo ""
echo "  5. 阅读完整文档:"
echo "     .harness/templates/HARNESS-BOOTSTRAP-GUIDE.md"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
