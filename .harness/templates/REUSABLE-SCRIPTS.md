# Harness Pipeline 可复用脚本库

**版本**: 1.0  
**日期**: 2026-06-23  
**说明**: 本文档包含所有可直接复用的脚本和配置文件

---

## 📋 文件清单

### 核心脚本

| 文件 | 路径 | 说明 |
|------|------|------|
| 流水线脚本 | `harness-pipeline.template.js` | 核心 6 阶段流水线 |
| 门禁检查 | `harness-gate-check.template.sh` | Phase 5/6 门禁检查 |
| TDD 验证器 | `tdd-evidence-validator.template.sh` | TDD 证据验证 |
| CI/CD 配置 | `ci-cd.template.yml` | GitHub Actions 配置 |

### 工具脚本（技术栈特定）

| 技术栈 | 脚本 |
|--------|------|
| Go | `go-testing-tools.template.sh` |
| Node.js | `nodejs-testing-tools.template.sh` |
| Python | `python-testing-tools.template.sh` |

### Agent 定义

| Agent | 文件 |
|-------|------|
| Owner Agent | `owner-agent.template.md` |
| Requirement Analyst | `requirement-analyst.template.md` |
| Architecture Designer | `architecture-designer.template.md` |
| Code Reviewer | `code-reviewer.template.md` |
| Test Engineer | `test-engineer.template.md` |

---

## 1. 核心流水线脚本

### 1.1 harness-pipeline.template.js

**用途**: Harness 6 阶段流水线的核心实现

**位置**: `.harness/workflows/harness-pipeline.js`

**说明**: 
- 此脚本已在原项目中验证
- 可直接复制到新项目
- 需要根据技术栈调整检查逻辑

**复制方法**:
```bash
cp .harness/templates/harness-pipeline.template.js \
   <new-project>/.harness/workflows/harness-pipeline.js
```

**关键配置点**:
```javascript
// 1. 技术栈检测
const TECH_STACK = detectTechStack();

// 2. 测试命令（需要根据技术栈调整）
const TEST_COMMANDS = {
  'go': 'go test ./...',
  'nodejs': 'npm test',
  'python': 'pytest'
};

// 3. 覆盖率要求
const COVERAGE_THRESHOLD = {
  'go': 30,
  'nodejs': 70,
  'python': 80
};
```

---

### 1.2 harness-gate-check.template.sh

**用途**: Phase 5/6 门禁检查

**位置**: `.harness/scripts/harness-gate-check-v2.sh`

**功能**:
1. **Phase 5 检查**:
   - QA 报告存在
   - 测试文件存在
   - TDD 证据完整
   - 覆盖率达标

2. **Phase 6 检查**:
   - 集成测试通过
   - 文档更新
   - Code Review 完成

**使用方法**:
```bash
# Phase 5 检查
bash .harness/scripts/harness-gate-check-v2.sh --phase 5 --change <change-name>

# Phase 6 检查
bash .harness/scripts/harness-gate-check-v2.sh --phase 6 --change <change-name>
```

**技术栈适配**:
```bash
# 修改第 50 行附近
case "$TECH_STACK" in
  "go")
    TEST_PATTERN="*_test.go"
    ;;
  "nodejs")
    TEST_PATTERN="*.test.ts"
    ;;
  "python")
    TEST_PATTERN="test_*.py"
    ;;
esac
```

---

### 1.3 tdd-evidence-validator.template.sh

**用途**: 验证 TDD 证据的完整性

**位置**: `.harness/scripts/tdd-evidence-validator.sh`

**检查项**:
- RED 证据存在（测试失败截图）
- GREEN 证据存在（测试通过截图）
- 测试代码先于实现代码提交

**使用方法**:
```bash
bash .harness/scripts/tdd-evidence-validator.sh \
  .harness/changes/<change-name>/_qa.md
```

---

## 2. CI/CD 配置模板

### 2.1 GitHub Actions (Go 项目)

**文件**: `ci-cd-go.template.yml`

**位置**: `.github/workflows/test.yml`

```yaml
name: Go Tests and Coverage

on:
  push:
    branches: [ main, master, develop ]
  pull_request:
    branches: [ main, master, develop ]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    
    - name: Run tests
      run: go test ./... -cover -coverprofile=coverage.out
    
    - name: Check coverage
      run: |
        COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
        THRESHOLD=30
        if (( $(echo "$COVERAGE < $THRESHOLD" | bc -l) )); then
          echo "❌ Coverage ${COVERAGE}% < ${THRESHOLD}%"
          exit 1
        fi
        echo "✅ Coverage ${COVERAGE}% >= ${THRESHOLD}%"

  quality-gate:
    runs-on: ubuntu-latest
    needs: [test]
    steps:
    - uses: actions/checkout@v3
    
    - name: Check test files exist
      run: |
        LOGIC_FILES=$(find . -name "*_logic.go" -not -path "*/mocks/*" || true)
        for file in $LOGIC_FILES; do
          test_file="${file%_logic.go}_logic_test.go"
          if [[ ! -f "$test_file" ]]; then
            echo "❌ Missing test: $test_file"
            exit 1
          fi
        done
        echo "✅ All logic files have tests"
```

### 2.2 GitHub Actions (Node.js 项目)

**文件**: `ci-cd-nodejs.template.yml`

```yaml
name: Node.js Tests and Coverage

on:
  push:
    branches: [ main, master, develop ]
  pull_request:
    branches: [ main, master, develop ]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Node.js
      uses: actions/setup-node@v3
      with:
        node-version: '18'
    
    - name: Install dependencies
      run: npm ci
    
    - name: Run tests
      run: npm test -- --coverage
    
    - name: Check coverage
      run: |
        COVERAGE=$(cat coverage/coverage-summary.json | jq '.total.lines.pct')
        THRESHOLD=70
        if (( $(echo "$COVERAGE < $THRESHOLD" | bc -l) )); then
          echo "❌ Coverage ${COVERAGE}% < ${THRESHOLD}%"
          exit 1
        fi
        echo "✅ Coverage ${COVERAGE}% >= ${THRESHOLD}%"
```

### 2.3 GitHub Actions (Python 项目)

**文件**: `ci-cd-python.template.yml`

```yaml
name: Python Tests and Coverage

on:
  push:
    branches: [ main, master, develop ]
  pull_request:
    branches: [ main, master, develop ]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Python
      uses: actions/setup-python@v4
      with:
        python-version: '3.11'
    
    - name: Install dependencies
      run: |
        pip install -r requirements.txt
        pip install pytest pytest-cov
    
    - name: Run tests
      run: pytest --cov=. --cov-report=term-missing
    
    - name: Check coverage
      run: |
        COVERAGE=$(pytest --cov=. --cov-report=json | jq '.totals.percent_covered')
        THRESHOLD=80
        if (( $(echo "$COVERAGE < $THRESHOLD" | bc -l) )); then
          echo "❌ Coverage ${COVERAGE}% < ${THRESHOLD}%"
          exit 1
        fi
        echo "✅ Coverage ${COVERAGE}% >= ${THRESHOLD}%"
```

---

## 3. 测试工具安装脚本

### 3.1 Go 项目

**文件**: `go-testing-tools.template.sh`

```bash
#!/usr/bin/env bash
# Go 测试工具安装脚本

set -euo pipefail

echo "🔧 安装 Go 测试工具..."

# 1. 安装 mockgen
echo "[1/3] 安装 mockgen..."
go install github.com/golang/mock/mockgen@latest

# 2. 安装 testify
echo "[2/3] 安装 testify..."
go get github.com/stretchr/testify/assert
go get github.com/stretchr/testify/require

# 3. 安装 gomock
echo "[3/3] 安装 gomock..."
go get github.com/golang/mock/gomock

echo ""
echo "✅ Go 测试工具安装完成"
echo ""
echo "已安装:"
echo "  - mockgen (Mock 生成器)"
echo "  - testify (断言库)"
echo "  - gomock (Mock 框架)"
```

### 3.2 Node.js 项目

**文件**: `nodejs-testing-tools.template.sh`

```bash
#!/usr/bin/env bash
# Node.js 测试工具安装脚本

set -euo pipefail

echo "🔧 安装 Node.js 测试工具..."

# 1. 安装 Jest
echo "[1/3] 安装 Jest..."
npm install --save-dev jest @types/jest

# 2. 安装 Testing Library
echo "[2/3] 安装 Testing Library..."
npm install --save-dev @testing-library/react @testing-library/jest-dom

# 3. 安装覆盖率工具
echo "[3/3] 配置覆盖率..."
npm install --save-dev @jest/coverage

# 4. 创建 jest.config.js
cat > jest.config.js << 'EOF'
module.exports = {
  preset: 'ts-jest',
  testEnvironment: 'node',
  collectCoverageFrom: [
    'src/**/*.{ts,tsx}',
    '!src/**/*.d.ts'
  ],
  coverageThreshold: {
    global: {
      branches: 70,
      functions: 70,
      lines: 70,
      statements: 70
    }
  }
};
EOF

echo ""
echo "✅ Node.js 测试工具安装完成"
```

### 3.3 Python 项目

**文件**: `python-testing-tools.template.sh`

```bash
#!/usr/bin/env bash
# Python 测试工具安装脚本

set -euo pipefail

echo "🔧 安装 Python 测试工具..."

# 1. 安装 pytest
echo "[1/4] 安装 pytest..."
pip install pytest

# 2. 安装覆盖率工具
echo "[2/4] 安装 pytest-cov..."
pip install pytest-cov

# 3. 安装 Mock 工具
echo "[3/4] 安装 pytest-mock..."
pip install pytest-mock

# 4. 创建 pytest.ini
echo "[4/4] 创建 pytest.ini..."
cat > pytest.ini << 'EOF'
[pytest]
testpaths = tests
python_files = test_*.py
python_classes = Test*
python_functions = test_*

# 覆盖率配置
addopts = 
    --cov=src
    --cov-report=term-missing
    --cov-report=html
    --cov-fail-under=80
EOF

echo ""
echo "✅ Python 测试工具安装完成"
```

---

## 4. Pre-commit Hook 模板

### 4.1 通用 Pre-commit Hook

**文件**: `pre-commit.template.sh`

**位置**: `tools/pre-commit.sh`

```bash
#!/usr/bin/env bash
# Pre-commit hook - 通用版本

set -e

echo "🔍 Pre-commit 检查..."

# 检测技术栈
if [[ -f "go.mod" ]]; then
  TECH_STACK="go"
elif [[ -f "package.json" ]]; then
  TECH_STACK="nodejs"
elif [[ -f "requirements.txt" ]] || [[ -f "setup.py" ]]; then
  TECH_STACK="python"
else
  echo "⚠️  无法检测技术栈，跳过检查"
  exit 0
fi

echo "检测到技术栈: $TECH_STACK"

# 获取待提交的文件
case "$TECH_STACK" in
  "go")
    FILES=$(git diff --cached --name-only --diff-filter=ACM | grep '\.go$' | grep -v '_test\.go$' || true)
    ;;
  "nodejs")
    FILES=$(git diff --cached --name-only --diff-filter=ACM | grep '\.\(ts\|js\)$' | grep -v '\.\(test\|spec\)\.' || true)
    ;;
  "python")
    FILES=$(git diff --cached --name-only --diff-filter=ACM | grep '\.py$' | grep -v '^test_' || true)
    ;;
esac

if [[ -z "$FILES" ]]; then
  echo "✅ 没有需要检查的文件"
  exit 0
fi

# 检查测试文件存在
echo ""
echo "[1/3] 检查测试文件..."

MISSING_TESTS=()

for file in $FILES; do
  case "$TECH_STACK" in
    "go")
      if [[ "$file" == *"_logic.go" ]]; then
        test_file="${file%_logic.go}_logic_test.go"
        [[ ! -f "$test_file" ]] && MISSING_TESTS+=("$file")
      fi
      ;;
    "nodejs")
      if [[ "$file" == *.ts ]] && [[ ! "$file" == *.d.ts ]]; then
        test_file="${file%.ts}.test.ts"
        [[ ! -f "$test_file" ]] && MISSING_TESTS+=("$file")
      fi
      ;;
    "python")
      if [[ "$file" != __*.py ]]; then
        dir=$(dirname "$file")
        base=$(basename "$file" .py)
        test_file="$dir/test_$base.py"
        [[ ! -f "$test_file" ]] && MISSING_TESTS+=("$file")
      fi
      ;;
  esac
done

if [[ ${#MISSING_TESTS[@]} -gt 0 ]]; then
  echo "❌ 以下文件缺少测试:"
  for file in "${MISSING_TESTS[@]}"; do
    echo "  - $file"
  done
  exit 1
fi

echo "  ✅ 所有文件都有测试"

# 运行测试
echo ""
echo "[2/3] 运行测试..."

case "$TECH_STACK" in
  "go")
    go test ./... -short
    ;;
  "nodejs")
    npm test
    ;;
  "python")
    pytest
    ;;
esac

echo "  ✅ 测试通过"

# 检查代码格式
echo ""
echo "[3/3] 检查代码格式..."

case "$TECH_STACK" in
  "go")
    UNFORMATTED=$(echo "$FILES" | xargs gofmt -l || true)
    if [[ -n "$UNFORMATTED" ]]; then
      echo "❌ 代码格式不正确:"
      echo "$UNFORMATTED"
      exit 1
    fi
    ;;
  "nodejs")
    npx prettier --check $FILES || exit 1
    ;;
  "python")
    black --check $FILES || exit 1
    ;;
esac

echo "  ✅ 代码格式正确"

echo ""
echo "✅ Pre-commit 检查全部通过"
exit 0
```

---

## 5. Scaffolding 工具模板

### 5.1 通用 Scaffolding 脚本

**文件**: `new-service-with-tests.template.sh`

**位置**: `tools/new-service-with-tests.sh`

（由于篇幅限制，此脚本在原项目中已提供，可直接复制）

---

## 6. 使用指南

### 6.1 快速部署到新项目

```bash
# 1. 创建新项目的 Harness 结构
cd <new-project>
bash <old-project>/.harness/templates/quick-deploy.sh

# 2. 根据技术栈选择模板
# Go 项目
cp .harness/templates/ci-cd-go.template.yml .github/workflows/test.yml
cp .harness/templates/go-testing-tools.template.sh tools/install-testing-tools.sh

# Node.js 项目
cp .harness/templates/ci-cd-nodejs.template.yml .github/workflows/test.yml
cp .harness/templates/nodejs-testing-tools.template.sh tools/install-testing-tools.sh

# Python 项目
cp .harness/templates/ci-cd-python.template.yml .github/workflows/test.yml
cp .harness/templates/python-testing-tools.template.sh tools/install-testing-tools.sh

# 3. 安装 Git hooks
bash tools/install-hooks.sh

# 4. 测试
bash .harness/scripts/harness-gate-check-v2.sh --help
```

### 6.2 一键部署脚本

**文件**: `quick-deploy.sh`

```bash
#!/usr/bin/env bash
# 一键部署 Harness 到新项目

set -euo pipefail

TEMPLATE_DIR="$1"
TARGET_DIR="${2:-.}"

if [[ ! -d "$TEMPLATE_DIR" ]]; then
  echo "❌ 模板目录不存在: $TEMPLATE_DIR"
  exit 1
fi

echo "🚀 部署 Harness 到: $TARGET_DIR"

# 1. 创建目录结构
echo "[1/5] 创建目录结构..."
mkdir -p "$TARGET_DIR/.harness"/{agents/subagents,skills/qa/scripts,workflows,knowledge/memory,changes/TEMPLATE,scripts,templates}
mkdir -p "$TARGET_DIR/tools"
mkdir -p "$TARGET_DIR/.github/workflows"

# 2. 复制核心文件
echo "[2/5] 复制核心文件..."
cp "$TEMPLATE_DIR/harness-pipeline.template.js" "$TARGET_DIR/.harness/workflows/harness-pipeline.js"
cp "$TEMPLATE_DIR/harness-gate-check.template.sh" "$TARGET_DIR/.harness/scripts/harness-gate-check-v2.sh"
chmod +x "$TARGET_DIR/.harness/scripts/harness-gate-check-v2.sh"

# 3. 复制 Agent 定义
echo "[3/5] 复制 Agent 定义..."
cp "$TEMPLATE_DIR"/agent-*.template.md "$TARGET_DIR/.harness/agents/subagents/"

# 4. 复制工具脚本
echo "[4/5] 复制工具脚本..."
cp "$TEMPLATE_DIR/pre-commit.template.sh" "$TARGET_DIR/tools/pre-commit.sh"
chmod +x "$TARGET_DIR/tools/pre-commit.sh"

# 5. 初始化文件
echo "[5/5] 初始化文件..."
cat > "$TARGET_DIR/.harness/knowledge/memory/MEMORY.md" << 'EOF'
# Memory Index

This file indexes all persistent memories.

## Project Memories

(To be added)

---

Last updated: $(date +%Y-%m-%d)
EOF

cat > "$TARGET_DIR/.harness/changes/INDEX.md" << 'EOF'
# Change Tracking Index

All changes to this project are tracked here.

## Active Changes

(None)

## Completed Changes

(None)
EOF

echo ""
echo "✅ Harness 部署完成！"
echo ""
echo "下一步:"
echo "  1. 根据技术栈选择 CI/CD 模板"
echo "  2. 安装测试工具: bash tools/install-testing-tools.sh"
echo "  3. 安装 Git hooks: bash tools/install-hooks.sh"
echo "  4. 开始使用: 阅读 HARNESS-BOOTSTRAP-GUIDE.md"
```

---

## 7. 版本信息

| 版本 | 日期 | 变更 |
|------|------|------|
| 1.0 | 2026-06-23 | 初始版本 |

---

**END OF REUSABLE SCRIPTS DOCUMENTATION**
