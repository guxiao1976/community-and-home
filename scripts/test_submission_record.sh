#!/bin/bash
# 提交记录留痕测试

API="http://localhost:8889/api/masterdata"
PARENT_ID=3633

RED='\033[31m'
GREEN='\033[32m'
NC='\033[0m'
FAILED=0

pass() { echo -e "  ${GREEN}✓ $1${NC}"; }
fail() { echo -e "  ${RED}✗ $1${NC}"; FAILED=1; }

api_post() { curl -s -X POST "${API}$1" -H "Content-Type: application/json" -d "$2"; }

check_record() {
  local entity_id=$1 expected_result=$2 label=$3
  local row
  row=$(docker exec -i mysql mysql -u root -p123456 masterdata_db -N -e \
    "SELECT review_result FROM md_submission_record WHERE entity_id=$entity_id ORDER BY id DESC LIMIT 1;" 2>/dev/null)
  if [ -z "$row" ]; then
    fail "$label: 无记录"
  elif [ "$row" = "$expected_result" ]; then
    pass "$label: review_result=$row"
  else
    fail "$label: 期望=$expected_result 实际=$row"
  fi
}

check_record_field() {
  local entity_id=$1 field=$2 expected=$3 label=$4
  local row
  row=$(docker exec -i mysql mysql -u root -p123456 masterdata_db -N -e \
    "SELECT $field FROM md_submission_record WHERE entity_id=$entity_id ORDER BY id DESC LIMIT 1;" 2>/dev/null)
  if [ "$row" = "$expected" ]; then
    pass "$label: $field=$row"
  else
    fail "$label: $field 期望=$expected 实际=$row"
  fi
}

echo "======================================================"
echo "  提交记录留痕测试"
echo "======================================================"
echo ""

# ---------- 场景1: 新增→提交→通过 ----------
echo "【场景1】新增 → 提交 → 审核通过"
RESP=$(api_post "/divisions" "{\"parent_id\":$PARENT_ID,\"level\":5,\"name\":\"留痕测试-通过\",\"code\":\"REC0000000101\",\"sort_order\":10}")
ID1=$(echo "$RESP" | grep -o '"id":[0-9]*' | grep -o '[0-9]*')
echo "  创建ID=$ID1"
[ -z "$ID1" ] && { fail "创建失败"; } || check_record "$ID1" "" "1.1 提交前(无记录)"

api_post "/divisions/$ID1/submit" > /dev/null
check_record "$ID1" 0 "1.2 提交后(review_result=0)"
check_record_field "$ID1" "submission_type" "1" "1.2 操作类型=新增"
check_record_field "$ID1" "entity_name" "留痕测试-通过" "1.2 实体名称快照"

api_post "/approval/administrative_division/$ID1/review" '{"action":"approve","review_notes":"测试通过"}' > /dev/null
check_record "$ID1" 1 "1.3 审核通过后(review_result=1)"

# 清理
docker exec -i mysql mysql -u root -p123456 masterdata_db --default-character-set=utf8mb4 -e "DELETE FROM md_administrative_division WHERE id=$ID1;" 2>/dev/null
echo ""

# ---------- 场景2: 新增→提交→拒绝 ----------
echo "【场景2】新增 → 提交 → 审核拒绝（物理删除）"
RESP=$(api_post "/divisions" "{\"parent_id\":$PARENT_ID,\"level\":5,\"name\":\"留痕测试-拒绝\",\"code\":\"REC0000000102\",\"sort_order\":11}")
ID2=$(echo "$RESP" | grep -o '"id":[0-9]*' | grep -o '[0-9]*')
echo "  创建ID=$ID2"
api_post "/divisions/$ID2/submit" > /dev/null
api_post "/approval/administrative_division/$ID2/review" '{"action":"reject","review_notes":"名称不符合规范"}' > /dev/null
check_record "$ID2" 2 "2.1 拒绝后(review_result=2)"
check_record_field "$ID2" "review_notes" "名称不符合规范" "2.2 审核备注保留"

# 验证实体已删除但记录仍在
ROW_COUNT=$(docker exec -i mysql mysql -u root -p123456 masterdata_db -N -e "SELECT COUNT(*) FROM md_administrative_division WHERE id=$ID2;" 2>/dev/null)
if [ "$ROW_COUNT" = "0" ]; then
  pass "2.3 实体已物理删除"
else
  fail "2.3 实体仍存在"
fi
REC_COUNT=$(docker exec -i mysql mysql -u root -p123456 masterdata_db -N -e "SELECT COUNT(*) FROM md_submission_record WHERE entity_id=$ID2;" 2>/dev/null)
if [ "$REC_COUNT" = "1" ]; then
  pass "2.4 记录保留"
else
  fail "2.4 记录数量=$REC_COUNT"
fi
echo ""

# ---------- 场景3: 发起删除→提交→通过 ----------
echo "【场景3】已批准数据 → 发起删除 → 提交 → 审核通过"
RESP=$(api_post "/divisions" "{\"parent_id\":$PARENT_ID,\"level\":5,\"name\":\"留痕测试-删除通过\",\"code\":\"REC0000000103\",\"sort_order\":12}")
ID3=$(echo "$RESP" | grep -o '"id":[0-9]*' | grep -o '[0-9]*')
echo "  创建ID=$ID3"
api_post "/divisions/$ID3/submit" > /dev/null
api_post "/approval/administrative_division/$ID3/review" '{"action":"approve","review_notes":""}' > /dev/null

api_post "/divisions/$ID3/request-delete" > /dev/null
check_record "$ID3" 0 "3.1 发起删除后(新记录review_result=0)"
check_record_field "$ID3" "submission_type" "3" "3.1 操作类型=删除"

api_post "/divisions/$ID3/submit" > /dev/null
check_record "$ID3" 0 "3.2 提交删除后(review_result=0)"

api_post "/approval/administrative_division/$ID3/review" '{"action":"approve","review_notes":"同意删除"}' > /dev/null
check_record "$ID3" 1 "3.3 审核通过(review_result=1)"

docker exec -i mysql mysql -u root -p123456 masterdata_db --default-character-set=utf8mb4 -e "DELETE FROM md_administrative_division WHERE id=$ID3;" 2>/dev/null
echo ""

# ---------- 场景4: 发起删除→提交→拒绝 ----------
echo "【场景4】已批准数据 → 发起删除 → 提交 → 审核拒绝"
RESP=$(api_post "/divisions" "{\"parent_id\":$PARENT_ID,\"level\":5,\"name\":\"留痕测试-删除拒绝\",\"code\":\"REC0000000104\",\"sort_order\":13}")
ID4=$(echo "$RESP" | grep -o '"id":[0-9]*' | grep -o '[0-9]*')
echo "  创建ID=$ID4"
api_post "/divisions/$ID4/submit" > /dev/null
api_post "/approval/administrative_division/$ID4/review" '{"action":"approve","review_notes":""}' > /dev/null

api_post "/divisions/$ID4/request-delete" > /dev/null
api_post "/divisions/$ID4/submit" > /dev/null
api_post "/approval/administrative_division/$ID4/review" '{"action":"reject","review_notes":"不同意删除"}' > /dev/null
check_record "$ID4" 2 "4.1 拒绝删除后(review_result=2)"
check_record_field "$ID4" "review_notes" "不同意删除" "4.2 拒绝原因保留"

docker exec -i mysql mysql -u root -p123456 masterdata_db --default-character-set=utf8mb4 -e "DELETE FROM md_administrative_division WHERE id=$ID4;" 2>/dev/null
echo ""

# ---------- 场景5: 新增→提交→撤回 ----------
echo "【场景5】新增 → 提交 → 撤回"
RESP=$(api_post "/divisions" "{\"parent_id\":$PARENT_ID,\"level\":5,\"name\":\"留痕测试-撤回\",\"code\":\"REC0000000105\",\"sort_order\":14}")
ID5=$(echo "$RESP" | grep -o '"id":[0-9]*' | grep -o '[0-9]*')
echo "  创建ID=$ID5"
api_post "/divisions/$ID5/submit" > /dev/null
check_record "$ID5" 0 "5.1 提交后(review_result=0)"

api_post "/divisions/$ID5/withdraw" > /dev/null
check_record "$ID5" 3 "5.2 撤回后(review_result=3)"
check_record_field "$ID5" "review_notes" "撤回" "5.2 撤回备注"

docker exec -i mysql mysql -u root -p123456 masterdata_db --default-character-set=utf8mb4 -e "DELETE FROM md_administrative_division WHERE id=$ID5;" 2>/dev/null
echo ""

# ---------- 场景6: 批量提交 ----------
echo "【场景6】批量新增 → 批量提交 → 批量拒绝"
RESP_A=$(api_post "/divisions" "{\"parent_id\":$PARENT_ID,\"level\":5,\"name\":\"批量留痕A\",\"code\":\"REC0000000106\",\"sort_order\":15}")
ID6A=$(echo "$RESP_A" | grep -o '"id":[0-9]*' | grep -o '[0-9]*')
RESP_B=$(api_post "/divisions" "{\"parent_id\":$PARENT_ID,\"level\":5,\"name\":\"批量留痕B\",\"code\":\"REC0000000107\",\"sort_order\":16}")
ID6B=$(echo "$RESP_B" | grep -o '"id":[0-9]*' | grep -o '[0-9]*')
echo "  批量ID=$ID6A, $ID6B"

api_post "/divisions/batch-submit" "{\"ids\":[$ID6A,$ID6B]}" > /dev/null
check_record "$ID6A" 0 "6.1 批量提交后A(review_result=0)"
check_record "$ID6B" 0 "6.2 批量提交后B(review_result=0)"

api_post "/approval/batch-review" "{\"entity_type\":\"administrative_division\",\"ids\":[$ID6A,$ID6B],\"action\":\"reject\",\"review_notes\":\"批量拒绝\"}" > /dev/null
check_record "$ID6A" 2 "6.3 批量拒绝后A(review_result=2)"
check_record "$ID6B" 2 "6.4 批量拒绝后B(review_result=2)"
echo ""

# ---------- 场景7: 查询API验证 ----------
echo "【场景7】查询API接口验证"
MY_RESP=$(curl -s "${API}/submission-records/my?page_size=100")
MY_TOTAL=$(echo "$MY_RESP" | grep -o '"total":[0-9]*' | grep -o '[0-9]*')
if [ "$MY_TOTAL" -ge "6" ]; then
  pass "7.1 我的提交记录 total=$MY_TOTAL (>=6)"
else
  fail "7.1 我的提交记录 total=$MY_TOTAL (期望>=6)"
fi

REVIEWED_RESP=$(curl -s "${API}/submission-records/reviewed?page_size=100")
REVIEWED_TOTAL=$(echo "$REVIEWED_RESP" | grep -o '"total":[0-9]*' | grep -o '[0-9]*')
if [ "$REVIEWED_TOTAL" -ge "5" ]; then
  pass "7.2 我的审核记录 total=$REVIEWED_TOTAL (>=5)"
else
  fail "7.2 我的审核记录 total=$REVIEWED_TOTAL (期望>=5)"
fi
echo ""

# 清理
docker exec -i mysql mysql -u root -p123456 masterdata_db --default-character-set=utf8mb4 -e "DELETE FROM md_submission_record;" 2>/dev/null

echo "======================================================"
if [ "$FAILED" = "1" ]; then
  echo -e "  ${RED}结果: 存在失败用例${NC}"
else
  echo -e "  ${GREEN}结果: 全部通过 ✓${NC}"
fi
echo "======================================================"
