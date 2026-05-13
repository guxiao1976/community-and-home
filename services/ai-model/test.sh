#!/bin/bash

set -e

echo "=========================================="
echo "  AI Model Service - Test Suite"
echo "=========================================="
echo ""

BASE_URL="http://localhost:8001"

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 测试函数
test_endpoint() {
    local name=$1
    local method=$2
    local endpoint=$3
    local data=$4

    echo -n "Testing $name... "

    if [ "$method" = "GET" ]; then
        response=$(curl -s -w "\n%{http_code}" "$BASE_URL$endpoint")
    else
        response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL$endpoint" \
            -H "Content-Type: application/json" \
            -d "$data")
    fi

    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n-1)

    if [ "$http_code" = "200" ]; then
        echo -e "${GREEN}✓ PASS${NC}"
        echo "  Response: $body" | head -c 100
        echo ""
    else
        echo -e "${RED}✗ FAIL${NC} (HTTP $http_code)"
        echo "  Response: $body"
    fi
    echo ""
}

# 1. 健康检查
echo "1. Health Check"
test_endpoint "Health" "GET" "/health" ""

# 2. 模型信息
echo "2. Model Info"
test_endpoint "Model Info" "GET" "/model/info" ""

# 3. 安全内容测试
echo "3. Safe Content Test"
test_endpoint "Safe Content" "POST" "/api/moderate/text" '{
  "content": "今天天气真好，适合出去散步"
}'

# 4. 政治敏感测试
echo "4. Political Sensitive Test"
test_endpoint "Political" "POST" "/api/moderate/text" '{
  "content": "打倒共产党"
}'

# 5. 色情低俗测试
echo "5. Pornographic Test"
test_endpoint "Pornographic" "POST" "/api/moderate/text" '{
  "content": "卖淫嫖娼，提供特殊服务"
}'

# 6. 人身攻击测试
echo "6. Personal Attack Test"
test_endpoint "Personal Attack" "POST" "/api/moderate/text" '{
  "content": "你妈死了，傻逼东西"
}'

# 7. 暴力恐怖测试
echo "7. Violence Test"
test_endpoint "Violence" "POST" "/api/moderate/text" '{
  "content": "制造炸弹，恐怖袭击"
}'

# 8. 指定检查维度测试
echo "8. Specific Categories Test"
test_endpoint "Specific Categories" "POST" "/api/moderate/text" '{
  "content": "习近平",
  "check_categories": ["政治敏感"]
}'

# 9. 性能测试
echo "9. Performance Test (10 requests)"
echo -n "Running... "
start_time=$(date +%s%3N)
for i in {1..10}; do
    curl -s -X POST "$BASE_URL/api/moderate/text" \
        -H "Content-Type: application/json" \
        -d '{"content": "测试内容'$i'"}' > /dev/null
done
end_time=$(date +%s%3N)
duration=$((end_time - start_time))
avg_latency=$((duration / 10))
echo -e "${GREEN}✓ DONE${NC}"
echo "  Total time: ${duration}ms"
echo "  Avg latency: ${avg_latency}ms"
echo ""

echo "=========================================="
echo "  Test Summary"
echo "=========================================="
echo "All tests completed!"
echo ""
echo "Next steps:"
echo "  1. Check logs: docker logs -f ai-model-python"
echo "  2. Monitor GPU: nvidia-smi"
echo "  3. Start Go RPC: cd rpc && go run ai_model.go"
