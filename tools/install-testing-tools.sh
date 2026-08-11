#!/usr/bin/env bash
#
# install-testing-tools.sh — 安装测试工具
#
# 用途：一键安装 API 测试所需的工具
#

set -euo pipefail

echo "🔧 安装 API 测试工具..."
echo ""

# 1. 安装 gomock
echo "[1/3] 安装 gomock (Mock 生成器)..."
if ! command -v mockgen &> /dev/null; then
    go install github.com/golang/mock/mockgen@latest
    echo "✅ gomock 安装完成"
else
    echo "✅ gomock 已安装: $(which mockgen)"
fi

# 2. 安装 testify (断言库)
echo ""
echo "[2/3] 安装 testify (断言库)..."
cd /home/jiaoxh/my-project/community-and-home
go get github.com/stretchr/testify/assert
go get github.com/stretchr/testify/require
echo "✅ testify 安装完成"

# 3. 安装 golang/mock
echo ""
echo "[3/3] 安装 golang/mock 依赖..."
go get github.com/golang/mock/gomock
echo "✅ golang/mock 安装完成"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ 所有工具安装完成"
echo ""
echo "验证:"
echo "  mockgen: $(which mockgen 2>/dev/null || echo '未找到')"
echo ""
echo "下一步:"
echo "  bash tools/generate-mocks.sh"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
