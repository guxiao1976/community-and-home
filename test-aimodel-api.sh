#!/bin/bash

# AI Model API Integration Test Script
# Tests all endpoints and validates request/response structure

BASE_URL="http://localhost:8891/api/v1"
RESULTS_FILE="api-test-results.txt"

echo "=== AI Model API Integration Test ===" > $RESULTS_FILE
echo "Test started at: $(date)" >> $RESULTS_FILE
echo "" >> $RESULTS_FILE

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

test_count=0
pass_count=0
fail_count=0

# Test function
test_api() {
    local test_name=$1
    local method=$2
    local endpoint=$3
    local data=$4
    local expected_fields=$5

    test_count=$((test_count + 1))
    echo -e "\n${YELLOW}Test #${test_count}: ${test_name}${NC}"
    echo "----------------------------------------" >> $RESULTS_FILE
    echo "Test #${test_count}: ${test_name}" >> $RESULTS_FILE
    echo "Method: ${method} ${endpoint}" >> $RESULTS_FILE

    if [ "$method" = "GET" ]; then
        response=$(curl -s -w "\n%{http_code}" "${BASE_URL}${endpoint}")
    elif [ "$method" = "POST" ]; then
        response=$(curl -s -w "\n%{http_code}" -X POST "${BASE_URL}${endpoint}" \
            -H "Content-Type: application/json" \
            -d "$data")
    elif [ "$method" = "PUT" ]; then
        response=$(curl -s -w "\n%{http_code}" -X PUT "${BASE_URL}${endpoint}" \
            -H "Content-Type: application/json" \
            -d "$data")
    elif [ "$method" = "DELETE" ]; then
        response=$(curl -s -w "\n%{http_code}" -X DELETE "${BASE_URL}${endpoint}")
    fi

    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    echo "HTTP Status: ${http_code}" >> $RESULTS_FILE
    echo "Response Body:" >> $RESULTS_FILE
    echo "$body" | python3 -m json.tool 2>/dev/null >> $RESULTS_FILE || echo "$body" >> $RESULTS_FILE

    # Check HTTP status
    if [ "$http_code" = "200" ] || [ "$http_code" = "201" ]; then
        # Check response structure
        code=$(echo "$body" | python3 -c "import sys, json; print(json.load(sys.stdin).get('code', 'N/A'))" 2>/dev/null)

        if [ "$code" = "0" ] || [ "$code" = "200" ]; then
            echo -e "${GREEN}✓ PASS${NC}"
            echo "Result: PASS" >> $RESULTS_FILE
            pass_count=$((pass_count + 1))
        else
            echo -e "${RED}✗ FAIL - Unexpected business code: $code${NC}"
            echo "Result: FAIL - Business code: $code" >> $RESULTS_FILE
            fail_count=$((fail_count + 1))
        fi
    else
        echo -e "${RED}✗ FAIL - HTTP $http_code${NC}"
        echo "Result: FAIL - HTTP $http_code" >> $RESULTS_FILE
        fail_count=$((fail_count + 1))
    fi

    echo "" >> $RESULTS_FILE
}

echo -e "\n${YELLOW}=== Model Configuration APIs ===${NC}\n"

# Test 1: Get Models List
test_api "Get Models List" "GET" "/models?page=1&page_size=10"

# Test 2: Get Model by ID
test_api "Get Model by ID" "GET" "/model/1"

# Test 3: Create Model
create_model_data='{
  "name": "test-integration-model",
  "display_name": "Test Integration Model",
  "provider": "openai",
  "type": "chat",
  "endpoint": "https://api.openai.com/v1",
  "max_tokens": 4096,
  "supported_features": "streaming,function_calling",
  "cost_per_1k_input_tokens": 0.03,
  "cost_per_1k_output_tokens": 0.06,
  "description": "Integration test model"
}'
test_api "Create Model" "POST" "/model" "$create_model_data"

# Get the created model ID
created_model_id=$(curl -s -X POST "${BASE_URL}/model" \
    -H "Content-Type: application/json" \
    -d "$create_model_data" | python3 -c "import sys, json; print(json.load(sys.stdin).get('data', {}).get('id', 0))" 2>/dev/null)

if [ "$created_model_id" != "0" ] && [ -n "$created_model_id" ]; then
    echo "Created model ID: $created_model_id" >> $RESULTS_FILE

    # Test 4: Update Model
    update_model_data="{
      \"id\": $created_model_id,
      \"display_name\": \"Updated Test Model\",
      \"max_tokens\": 8192,
      \"description\": \"Updated description\"
    }"
    test_api "Update Model" "PUT" "/model" "$update_model_data"

    # Test 5: Get Updated Model
    test_api "Get Updated Model" "GET" "/model/$created_model_id"

    # Test 6: Delete Model
    test_api "Delete Model" "DELETE" "/model/$created_model_id"
fi

echo -e "\n${YELLOW}=== API Key Management APIs ===${NC}\n"

# Test 7: Get API Keys List
test_api "Get API Keys List" "GET" "/apikeys?page=1&page_size=10"

# Test 8: Create API Key
create_apikey_data='{
  "model_id": 1,
  "key_name": "test-api-key",
  "api_key": "sk-test-1234567890",
  "description": "Test API key"
}'
test_api "Create API Key" "POST" "/apikey" "$create_apikey_data"

# Get the created API key ID
created_apikey_id=$(curl -s -X POST "${BASE_URL}/apikey" \
    -H "Content-Type: application/json" \
    -d "$create_apikey_data" | python3 -c "import sys, json; print(json.load(sys.stdin).get('data', {}).get('id', 0))" 2>/dev/null)

if [ "$created_apikey_id" != "0" ] && [ -n "$created_apikey_id" ]; then
    echo "Created API key ID: $created_apikey_id" >> $RESULTS_FILE

    # Test 9: Get API Key by ID
    test_api "Get API Key by ID" "GET" "/apikey/$created_apikey_id"

    # Test 10: Update API Key
    update_apikey_data="{
      \"id\": $created_apikey_id,
      \"key_name\": \"updated-test-key\"
    }"
    test_api "Update API Key" "PUT" "/apikey" "$update_apikey_data"

    # Test 11: Delete API Key
    test_api "Delete API Key" "DELETE" "/apikey/$created_apikey_id"
fi

echo -e "\n${YELLOW}=== Template Management APIs ===${NC}\n"

# Test 12: Get Templates List
test_api "Get Templates List" "GET" "/templates?page=1&page_size=10"

# Test 13: Create Template
create_template_data='{
  "name": "test-template",
  "content": "You are a helpful assistant. {{user_input}}",
  "category": "general",
  "description": "Test template"
}'
test_api "Create Template" "POST" "/template" "$create_template_data"

# Get the created template ID
created_template_id=$(curl -s -X POST "${BASE_URL}/template" \
    -H "Content-Type: application/json" \
    -d "$create_template_data" | python3 -c "import sys, json; print(json.load(sys.stdin).get('data', {}).get('id', 0))" 2>/dev/null)

if [ "$created_template_id" != "0" ] && [ -n "$created_template_id" ]; then
    echo "Created template ID: $created_template_id" >> $RESULTS_FILE

    # Test 14: Get Template by ID
    test_api "Get Template by ID" "GET" "/template/$created_template_id"

    # Test 15: Update Template
    update_template_data="{
      \"id\": $created_template_id,
      \"name\": \"updated-test-template\",
      \"content\": \"Updated content: {{user_input}}\"
    }"
    test_api "Update Template" "PUT" "/template" "$update_template_data"

    # Test 16: Delete Template
    test_api "Delete Template" "DELETE" "/template/$created_template_id"
fi

echo -e "\n${YELLOW}=== Statistics APIs ===${NC}\n"

# Test 17: Get Usage Statistics
test_api "Get Usage Statistics" "GET" "/statistics?page=1&page_size=10"

# Test 18: Get Cost Stats
test_api "Get Cost Stats" "GET" "/cost/stats"

# Test 19: Health Check
test_api "Health Check" "GET" "/health"

# Summary
echo "" >> $RESULTS_FILE
echo "========================================" >> $RESULTS_FILE
echo "Test Summary" >> $RESULTS_FILE
echo "========================================" >> $RESULTS_FILE
echo "Total Tests: $test_count" >> $RESULTS_FILE
echo "Passed: $pass_count" >> $RESULTS_FILE
echo "Failed: $fail_count" >> $RESULTS_FILE
echo "Test completed at: $(date)" >> $RESULTS_FILE

echo -e "\n${YELLOW}=== Test Summary ===${NC}"
echo -e "Total Tests: $test_count"
echo -e "${GREEN}Passed: $pass_count${NC}"
echo -e "${RED}Failed: $fail_count${NC}"
echo -e "\nDetailed results saved to: $RESULTS_FILE"
