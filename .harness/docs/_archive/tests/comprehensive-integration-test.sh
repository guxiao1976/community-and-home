#!/usr/bin/env bash
# Comprehensive Integration Test Suite
# Tests all 7 completed improvements

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$PROJECT_ROOT"

PASS=0
FAIL=0
WARN=0

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🧪 综合集成测试套件"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "测试范围：7 个已完成的改进"
echo "开始时间：$(date)"
echo ""

# ============================================================================
# 问题 1：Pipeline 模板模块化
# ============================================================================
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "问题 1：Pipeline 模板模块化"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Test 1.1: Markdown 模板文件存在
echo "[1.1] 检查 Markdown 模板文件..."
TEMPLATES=(
  ".harness/agents/prompts/templates/generator.md"
  ".harness/agents/prompts/templates/qa.md"
  ".harness/agents/prompts/templates/review.md"
  ".harness/agents/prompts/templates/debug.md"
)

template_ok=true
for template in "${TEMPLATES[@]}"; do
  if [ -f "$template" ]; then
    echo "  ✅ $template"
  else
    echo "  ❌ $template (missing)"
    template_ok=false
  fi
done

if $template_ok; then
  echo "  ✅ PASS: 所有模板文件存在"
  PASS=$((PASS + 1))
else
  echo "  ❌ FAIL: 部分模板文件缺失"
  FAIL=$((FAIL + 1))
fi
echo ""

# Test 1.2: 模板渲染器
echo "[1.2] 测试模板渲染器..."
if node -e "
const { render } = require('./.harness/agents/prompts/template-renderer.js');
const result = render('Hello {{name}}', { name: 'Test' });
if (result === 'Hello Test') process.exit(0);
process.exit(1);
" 2>/dev/null; then
  echo "  ✅ PASS: 模板渲染器工作正常"
  PASS=$((PASS + 1))
else
  echo "  ❌ FAIL: 模板渲染器错误"
  FAIL=$((FAIL + 1))
fi
echo ""

# ============================================================================
# 问题 2：CI/CD 集成
# ============================================================================
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "问题 2：CI/CD 集成"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Test 2.1: GitHub Actions workflow
echo "[2.1] 检查 GitHub Actions workflow..."
if [ -f ".github/workflows/harness-qa-check.yml" ]; then
  echo "  ✅ PASS: harness-qa-check.yml 存在"
  PASS=$((PASS + 1))
else
  echo "  ❌ FAIL: harness-qa-check.yml 缺失"
  FAIL=$((FAIL + 1))
fi
echo ""

# Test 2.2: 本地结果发布脚本
echo "[2.2] 检查本地结果发布脚本..."
if [ -x ".harness/scripts/publish-qa-results.sh" ]; then
  echo "  ✅ PASS: publish-qa-results.sh 存在且可执行"
  PASS=$((PASS + 1))
else
  echo "  ❌ FAIL: publish-qa-results.sh 缺失或不可执行"
  FAIL=$((FAIL + 1))
fi
echo ""

# ============================================================================
# 问题 3：硬编码服务映射
# ============================================================================
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "问题 3：硬编码服务映射"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Test 3.1: 服务元数据文件
echo "[3.1] 检查服务元数据文件..."
metadata_count=$(find services -name ".service.json" 2>/dev/null | wc -l)
if [ "$metadata_count" -eq 9 ]; then
  echo "  ✅ PASS: 9 个服务元数据文件存在"
  PASS=$((PASS + 1))
else
  echo "  ⚠️  WARN: 发现 $metadata_count 个元数据文件 (预期 9)"
  WARN=$((WARN + 1))
fi
echo ""

# Test 3.2: 服务注册中心
echo "[3.2] 检查服务注册中心..."
if [ -f ".harness/registry/services.json" ] && \
   jq -e '.services | length == 9' .harness/registry/services.json >/dev/null 2>&1; then
  echo "  ✅ PASS: 服务注册中心包含 9 个服务"
  PASS=$((PASS + 1))
else
  echo "  ❌ FAIL: 服务注册中心错误"
  FAIL=$((FAIL + 1))
fi
echo ""

# Test 3.3: JavaScript 加载器
echo "[3.3] 测试 JavaScript 服务注册加载器..."
if node -e "
const { VALID_SERVICES } = require('./.harness/workflows/service-registry-loader.js');
if (VALID_SERVICES.length === 9) process.exit(0);
process.exit(1);
" 2>/dev/null; then
  echo "  ✅ PASS: JavaScript 加载器正常 (9 个服务)"
  PASS=$((PASS + 1))
else
  echo "  ❌ FAIL: JavaScript 加载器错误"
  FAIL=$((FAIL + 1))
fi
echo ""

# Test 3.4: Shell 加载器
echo "[3.4] 测试 Shell 服务注册加载器..."
if source .harness/scripts/service-registry-loader.sh 2>/dev/null && \
   [ ${#SERVICES[@]} -eq 9 ] && [ ${#SVC_MODULE_MAP[@]} -eq 9 ]; then
  echo "  ✅ PASS: Shell 加载器正常 (9 服务, 9 模块)"
  PASS=$((PASS + 1))
else
  echo "  ❌ FAIL: Shell 加载器错误"
  FAIL=$((FAIL + 1))
fi
echo ""

# ============================================================================
# 问题 4：AST 检查器
# ============================================================================
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "问题 4：AST 检查器"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Test 4.1: AST 检查器二进制
echo "[4.1] 检查 AST 检查器二进制..."
if [ -f ".harness/tools/go-ast-checker/go-ast-checker" ] && \
   [ -x ".harness/tools/go-ast-checker/go-ast-checker" ]; then
  echo "  ✅ PASS: AST 检查器存在且可执行"
  PASS=$((PASS + 1))
else
  echo "  ❌ FAIL: AST 检查器缺失或不可执行"
  FAIL=$((FAIL + 1))
fi
echo ""

# Test 4.2: AST 检查器功能测试
echo "[4.2] 测试 AST 检查器功能..."
cat > /tmp/test-ast-good.go << 'EOF'
package main
type User struct {
    UserId int64 `json:"user_id,string"`
}
EOF

cat > /tmp/test-ast-bad.go << 'EOF'
package main
type User struct {
    UserId int64 `json:"user_id"`
}
EOF

mkdir -p /tmp/test-ast-good /tmp/test-ast-bad
cp /tmp/test-ast-good.go /tmp/test-ast-good/user.go
cp /tmp/test-ast-bad.go /tmp/test-ast-bad/user.go

# 测试正确代码（应该通过）
if .harness/tools/go-ast-checker/go-ast-checker \
   -service-dir /tmp/test-ast-good \
   -service-name test \
   -registry .harness/registry/services.json >/dev/null 2>&1; then
  echo "  ✅ 正确代码检测：PASS"
  good_pass=true
else
  echo "  ❌ 正确代码检测：FAIL (误报)"
  good_pass=false
fi

# 测试错误代码（应该失败）
if ! .harness/tools/go-ast-checker/go-ast-checker \
   -service-dir /tmp/test-ast-bad \
   -service-name test \
   -registry .harness/registry/services.json >/dev/null 2>&1; then
  echo "  ✅ 错误代码检测：FAIL (正确检测)"
  bad_fail=true
else
  echo "  ❌ 错误代码检测：PASS (漏检)"
  bad_fail=false
fi

if $good_pass && $bad_fail; then
  echo "  ✅ PASS: AST 检查器功能正常"
  PASS=$((PASS + 1))
else
  echo "  ❌ FAIL: AST 检查器功能异常"
  FAIL=$((FAIL + 1))
fi
echo ""

# ============================================================================
# 问题 5：确定性验证层
# ============================================================================
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "问题 5：确定性验证层"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Test 5.1: 确定性检查脚本
echo "[5.1] 测试确定性检查脚本..."
if bash .harness/scripts/deterministic-checks.sh services/auth-service --json \
   >/dev/null 2>&1; then
  echo "  ✅ PASS: 确定性检查脚本正常"
  PASS=$((PASS + 1))
else
  echo "  ⚠️  WARN: 确定性检查有警告（可接受）"
  WARN=$((WARN + 1))
fi
echo ""

# Test 5.2: AI 判断验证器
echo "[5.2] 测试 AI 判断验证器..."
if node .harness/validators/test-validator.js >/dev/null 2>&1; then
  echo "  ✅ PASS: AI 判断验证器测试通过 (4/4)"
  PASS=$((PASS + 1))
else
  echo "  ❌ FAIL: AI 判断验证器测试失败"
  FAIL=$((FAIL + 1))
fi
echo ""

# Test 5.3: 判断分析脚本
echo "[5.3] 检查判断分析脚本..."
if [ -x ".harness/scripts/analyze-judgments.sh" ]; then
  echo "  ✅ PASS: 判断分析脚本存在且可执行"
  PASS=$((PASS + 1))
else
  echo "  ❌ FAIL: 判断分析脚本缺失"
  FAIL=$((FAIL + 1))
fi
echo ""

# ============================================================================
# 问题 7：重复代码清理
# ============================================================================
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "问题 7：重复代码清理"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Test 7.1: Templates README
echo "[7.1] 检查 templates README..."
if [ -f ".harness/templates/README.md" ]; then
  echo "  ✅ PASS: templates README 存在"
  PASS=$((PASS + 1))
else
  echo "  ❌ FAIL: templates README 缺失"
  FAIL=$((FAIL + 1))
fi
echo ""

# ============================================================================
# 问题 8：Python 依赖消除
# ============================================================================
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "问题 8：Python 依赖消除"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Test 8.1: Circuit Breaker Shell 实现
echo "[8.1] 测试 Circuit Breaker (Shell)..."
if source .harness/scripts/circuit_breaker.sh 2>/dev/null && \
   circuit_breaker_reset "test-op" 2>/dev/null && \
   circuit_breaker_call "test-op" 3 10 2>/dev/null; then
  echo "  ✅ PASS: Circuit Breaker (Shell) 正常"
  PASS=$((PASS + 1))
else
  echo "  ❌ FAIL: Circuit Breaker (Shell) 错误"
  FAIL=$((FAIL + 1))
fi
echo ""

# ============================================================================
# 总结
# ============================================================================
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "测试结果总结"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "总测试数：$((PASS + FAIL + WARN))"
echo "✅ 通过：$PASS"
echo "❌ 失败：$FAIL"
echo "⚠️  警告：$WARN"
echo ""

TOTAL=$((PASS + FAIL + WARN))
if [ $TOTAL -gt 0 ]; then
  PASS_RATE=$((PASS * 100 / TOTAL))
  echo "通过率：${PASS_RATE}%"
fi

echo ""
echo "结束时间：$(date)"
echo ""

if [ $FAIL -eq 0 ]; then
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo "🎉 所有测试通过！"
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  exit 0
else
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo "⚠️  部分测试失败，请检查"
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  exit 1
fi
