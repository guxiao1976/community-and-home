#!/bin/bash

# Phase 4 测试脚本 - 行政区划管理

BASE_URL="http://localhost:8889"
IDENTITY_URL="http://localhost:8888"

echo "=========================================="
echo "Phase 4: 行政区划管理功能测试"
echo "=========================================="
echo ""

# 1. 登录获取token
echo "1. 登录获取访问令牌..."
LOGIN_RESPONSE=$(curl -s -X POST "${IDENTITY_URL}/api/identity/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "13800000000",
    "password": "Admin@123456"
  }')

echo "登录响应: $LOGIN_RESPONSE"
TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
  echo "❌ 登录失败，无法获取token"
  exit 1
fi

echo "✅ 登录成功，Token: ${TOKEN:0:20}..."
echo ""

# 2. 获取行政区划列表
echo "2. 获取行政区划列表..."
DIVISIONS_RESPONSE=$(curl -s -X GET "${BASE_URL}/api/masterdata/divisions" \
  -H "Authorization: Bearer $TOKEN")

echo "行政区划列表响应: $DIVISIONS_RESPONSE"
echo ""

# 3. 创建省级区划
echo "3. 创建省级行政区划（测试省）..."
CREATE_PROVINCE=$(curl -s -X POST "${BASE_URL}/api/masterdata/divisions" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "level": 1,
    "name": "测试省",
    "code": "999000",
    "sort_order": 99
  }')

echo "创建省级区划响应: $CREATE_PROVINCE"
PROVINCE_ID=$(echo $CREATE_PROVINCE | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)
echo "省级区划ID: $PROVINCE_ID"
echo ""

# 4. 创建市级区划
if [ ! -z "$PROVINCE_ID" ]; then
  echo "4. 创建市级行政区划（测试市）..."
  CREATE_CITY=$(curl -s -X POST "${BASE_URL}/api/masterdata/divisions" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
      \"parent_id\": $PROVINCE_ID,
      \"level\": 2,
      \"name\": \"测试市\",
      \"code\": \"990100\",
      \"sort_order\": 1
    }")

  echo "创建市级区划响应: $CREATE_CITY"
  CITY_ID=$(echo $CREATE_CITY | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)
  echo "市级区划ID: $CITY_ID"
  echo ""
fi

# 5. 按父级ID查询
echo "5. 查询省级下的子区划..."
CHILDREN_RESPONSE=$(curl -s -X GET "${BASE_URL}/api/masterdata/divisions?parent_id=$PROVINCE_ID" \
  -H "Authorization: Bearer $TOKEN")

echo "子区划列表响应: $CHILDREN_RESPONSE"
echo ""

# 6. 按ID查询单个区划
echo "6. 按ID查询区划详情..."
DETAIL_RESPONSE=$(curl -s -X GET "${BASE_URL}/api/masterdata/divisions/$PROVINCE_ID" \
  -H "Authorization: Bearer $TOKEN")

echo "区划详情响应: $DETAIL_RESPONSE"
echo ""

# 7. 更新区划
echo "7. 更新区划信息..."
UPDATE_RESPONSE=$(curl -s -X PUT "${BASE_URL}/api/masterdata/divisions/$PROVINCE_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "测试省（已更新）",
    "sort_order": 100
  }')

echo "更新区划响应: $UPDATE_RESPONSE"
echo ""

# 8. 删除区划（先删除市级，再删除省级）
echo "8. 删除测试数据..."
if [ ! -z "$CITY_ID" ]; then
  echo "删除市级区划..."
  DELETE_CITY=$(curl -s -X DELETE "${BASE_URL}/api/masterdata/divisions/$CITY_ID" \
    -H "Authorization: Bearer $TOKEN")
  echo "删除市级响应: $DELETE_CITY"
fi

if [ ! -z "$PROVINCE_ID" ]; then
  echo "删除省级区划..."
  DELETE_PROVINCE=$(curl -s -X DELETE "${BASE_URL}/api/masterdata/divisions/$PROVINCE_ID" \
    -H "Authorization: Bearer $TOKEN")
  echo "删除省级响应: $DELETE_PROVINCE"
fi

echo ""
echo "=========================================="
echo "✅ Phase 4 测试完成"
echo "=========================================="
