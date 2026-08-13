#!/usr/bin/env bash
#
# new-service-with-tests.sh — 创建新服务并自动生成测试骨架
#
# Usage:
#   bash tools/new-service-with-tests.sh <service-name> <service-type>
#
# Examples:
#   bash tools/new-service-with-tests.sh payment-service api
#   bash tools/new-service-with-tests.sh notification-service rpc
#

set -euo pipefail

SERVICE_NAME="${1:-}"
SERVICE_TYPE="${2:-api}"  # api or rpc
PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"

if [[ -z "$SERVICE_NAME" ]]; then
    echo "Usage: $0 <service-name> [api|rpc]"
    echo ""
    echo "Examples:"
    echo "  $0 payment-service api"
    echo "  $0 notification-service rpc"
    exit 1
fi

SERVICE_DIR="$PROJECT_ROOT/services/$SERVICE_NAME"

if [[ -d "$SERVICE_DIR" ]]; then
    echo "❌ 服务已存在: $SERVICE_DIR"
    exit 1
fi

echo "🚀 创建新服务: $SERVICE_NAME (类型: $SERVICE_TYPE)"
echo ""

# ==================== Step 1: 使用 goctl 创建服务 ====================

echo "[1/6] 使用 goctl 创建服务骨架..."
cd "$PROJECT_ROOT/services"

if [[ "$SERVICE_TYPE" == "api" ]]; then
    goctl api new "$SERVICE_NAME" 2>&1 || {
        echo "⚠️  goctl 不可用，创建基本目录结构"
        mkdir -p "$SERVICE_DIR/api/internal/logic"
    }
elif [[ "$SERVICE_TYPE" == "rpc" ]]; then
    # RPC 服务需要 proto 文件
    echo "⚠️  RPC 服务需要先定义 proto 文件"
    mkdir -p "$SERVICE_DIR/rpc/internal/logic"
fi

cd "$SERVICE_DIR"

# ==================== Step 2: 创建测试目录 ====================

echo "[2/6] 创建测试目录结构..."

if [[ "$SERVICE_TYPE" == "api" ]]; then
    mkdir -p api/internal/logic/mocks
    TEST_DIR="api/internal/logic"
else
    mkdir -p rpc/internal/logic/mocks
    TEST_DIR="rpc/internal/logic"
fi

# ==================== Step 3: 生成测试模板 ====================

echo "[3/6] 生成测试模板..."

cat > "$TEST_DIR/example_logic_test.go" <<'EOTEST'
package logic

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExampleLogic_Example 示例测试
// TODO: 根据实际 Logic 修改此测试
func TestExampleLogic_Example(t *testing.T) {
	tests := []struct {
		name      string
		input     interface{}
		mockSetup func(*gomock.Controller) interface{}
		want      interface{}
		wantErr   bool
	}{
		{
			name:  "success - 正常场景",
			input: nil,
			mockSetup: func(ctrl *gomock.Controller) interface{} {
				// TODO: 创建 Mock 对象
				// mockRpc := xxxmocks.NewMockXxxServiceClient(ctrl)
				// mockRpc.EXPECT().Method(gomock.Any(), gomock.Any()).Return(...)
				return nil
			},
			want:    nil,
			wantErr: false,
		},
		{
			name:  "error - 参数为空",
			input: nil,
			mockSetup: func(ctrl *gomock.Controller) interface{} {
				return nil
			},
			wantErr: true,
		},
		{
			name:  "error - 业务错误",
			input: nil,
			mockSetup: func(ctrl *gomock.Controller) interface{} {
				// TODO: Mock 返回错误
				return nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// TODO: 创建 Logic 并注入 Mock
			// svcCtx := tt.mockSetup(ctrl)
			// l := NewExampleLogic(context.Background(), svcCtx)

			// Execute
			// got, err := l.Example(tt.input)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, got)
			// assert.Equal(t, tt.want, got)
		})
	}
}
EOTEST

# ==================== Step 4: 生成 QA 报告模板 ====================

echo "[4/6] 生成 QA 报告模板..."

cat > "_qa.md" <<EOQA
# QA 报告 — $SERVICE_NAME

**服务名称**: $SERVICE_NAME
**服务类型**: $SERVICE_TYPE
**QA 时间**: $(date +%Y-%m-%d)
**状态**: 🔲 待完成

---

## 一、TDD 证据检查

| 函数名 | 测试文件 | RED 确认 | GREEN 确认 | 状态 |
|--------|---------|---------|-----------|:---:|
| ExampleLogic | example_logic_test.go | 🔲 TODO | 🔲 TODO | 🔲 |

**TDD 流程**:
1. 编写测试用例（预期失败）
2. 运行测试，确认 RED（记录失败输出）
3. 编写最小实现
4. 运行测试，确认 GREEN
5. 重构（如需要）

---

## 二、单元测试

### 测试覆盖率

\`\`\`bash
$ go test ./$TEST_DIR/... -cover

# TODO: 粘贴覆盖率结果
\`\`\`

**目标**: >= 30%

### 测试运行结果

\`\`\`bash
$ go test ./$TEST_DIR/... -v

# TODO: 粘贴测试结果
\`\`\`

---

## 三、机械化检查

\`\`\`bash
$ bash .harness/skills/qa/scripts/harness-checks.sh --service $SERVICE_NAME

# TODO: 粘贴检查结果
\`\`\`

---

## 四、QA 总结

### 检查项

- [ ] 所有测试通过
- [ ] 覆盖率 >= 30%
- [ ] TDD 证据完整
- [ ] 机械化检查通过
- [ ] 代码已 Review

### VERDICT

🔲 **PASS** / **FAIL**

---

**QA 人员**: TODO
**完成时间**: TODO
EOQA

# ==================== Step 5: 创建 README ====================

echo "[5/6] 创建 README..."

cat > "README.md" <<EOREADME
# $SERVICE_NAME

**服务类型**: $SERVICE_TYPE
**创建时间**: $(date +%Y-%m-%d)

---

## 目录结构

\`\`\`
$SERVICE_NAME/
├── $TEST_DIR/
│   ├── example_logic.go          # 业务逻辑
│   ├── example_logic_test.go     # 单元测试
│   └── mocks/                    # Mock 文件
├── _qa.md                        # QA 报告
└── README.md                     # 本文件
\`\`\`

---

## 开发指南

### 1. 编写业务逻辑

\`\`\`bash
# 创建新的 Logic
vim $TEST_DIR/xxx_logic.go
\`\`\`

### 2. 编写测试（TDD）

\`\`\`bash
# 先写测试
vim $TEST_DIR/xxx_logic_test.go

# 运行测试（预期失败 - RED）
go test ./$TEST_DIR/... -v

# 编写实现
vim $TEST_DIR/xxx_logic.go

# 运行测试（预期成功 - GREEN）
go test ./$TEST_DIR/... -v
\`\`\`

### 3. 生成 Mock

\`\`\`bash
# 为 gRPC client 生成 Mock
bash tools/generate-grpc-mocks.sh all

# 为 Model 生成 Mock
bash tools/generate-mocks.sh $SERVICE_NAME
\`\`\`

### 4. 运行 QA

\`\`\`bash
# 运行机械化检查
bash .harness/skills/qa/scripts/harness-checks.sh --service $SERVICE_NAME

# 查看覆盖率
go test ./$TEST_DIR/... -cover

# 生成覆盖率报告
go test ./$TEST_DIR/... -coverprofile=coverage.out
go tool cover -html=coverage.out
\`\`\`

### 5. 更新 _qa.md

按照 TDD 流程填写 QA 报告。

---

## 测试要求

- ✅ 每个 Logic 都有对应的测试
- ✅ 使用 Table-driven tests 模式
- ✅ 覆盖正常/边界/错误场景
- ✅ 覆盖率 >= 30%
- ✅ TDD 证据完整

---

## 参考资料

- Mock 设置指南: \`.harness/docs/_archive/MOCK-SETUP-GUIDE.md\`
- 测试模板示例: \`services/user-service/api/internal/logic/user/user_logic_test.go\`
EOREADME

# ==================== Step 6: 更新变更追踪 ====================

echo "[6/6] 更新变更追踪..."

CHANGE_DIR="$PROJECT_ROOT/.harness/changes/$SERVICE_NAME-creation"
mkdir -p "$CHANGE_DIR"

cat > "$CHANGE_DIR/summary.md" <<EOSUMMARY
# 变更追踪 — 创建 $SERVICE_NAME

**变更类型**: feature
**创建时间**: $(date +%Y-%m-%d)
**状态**: 🔲 进行中

---

## 变更概述

创建新服务: $SERVICE_NAME ($SERVICE_TYPE)

---

## 阶段 1: 需求分析

**需求文档**: TODO

---

## 阶段 2: 技术设计

**设计文档**: TODO

---

## 阶段 3: API 设计

**API 文档**: TODO

---

## 阶段 4: 数据库设计

**数据库文档**: TODO

---

## 阶段 5: 编码实现

**实现路径**: \`services/$SERVICE_NAME/\`

---

## 阶段 6: 测试验证

**QA 报告**: \`services/$SERVICE_NAME/_qa.md\`

---

**创建者**: $(git config user.name || echo "Unknown")
**创建时间**: $(date +%Y-%m-%d)
EOSUMMARY

# 更新 INDEX.md
if [[ -f "$PROJECT_ROOT/.harness/changes/INDEX.md" ]]; then
    echo "- [$SERVICE_NAME]($SERVICE_NAME-creation/summary.md) - 创建新服务 ($(date +%Y-%m-%d))" >> "$PROJECT_ROOT/.harness/changes/INDEX.md"
fi

# ==================== 完成 ====================

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "✅ 新服务创建完成: $SERVICE_NAME"
echo ""
echo "📁 服务目录: $SERVICE_DIR"
echo ""
echo "📋 生成的文件:"
echo "  ✅ $TEST_DIR/example_logic_test.go    (测试模板)"
echo "  ✅ _qa.md                             (QA 报告模板)"
echo "  ✅ README.md                          (服务文档)"
echo "  ✅ .harness/changes/$SERVICE_NAME-creation/  (变更追踪)"
echo ""
echo "🎯 下一步:"
echo ""
echo "  1. 编写业务逻辑:"
echo "     cd $SERVICE_DIR"
echo "     vim $TEST_DIR/xxx_logic.go"
echo ""
echo "  2. 编写测试 (TDD):"
echo "     vim $TEST_DIR/xxx_logic_test.go"
echo "     go test ./$TEST_DIR/... -v"
echo ""
echo "  3. 生成 Mock:"
echo "     bash tools/generate-grpc-mocks.sh all"
echo ""
echo "  4. 运行 QA:"
echo "     bash .harness/skills/qa/scripts/harness-checks.sh --service $SERVICE_NAME"
echo ""
echo "  5. 填写 QA 报告:"
echo "     vim $SERVICE_DIR/_qa.md"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
