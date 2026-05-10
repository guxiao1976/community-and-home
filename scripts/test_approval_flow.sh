#!/bin/bash
# 基层组织审批流程测试

API="http://localhost:8889/api/masterdata"
PARENT_ID=3633

RED='\033[31m'
GREEN='\033[32m'
NC='\033[0m'
FAILED=0

pass() { echo -e "  ${GREEN}✓ $1${NC}"; }
fail() { echo -e "  ${RED}✗ $1${NC}"; FAILED=1; }

api_post() {
  local path="$1" body="$2"
  curl -s -X POST "${API}${path}" -H "Content-Type: application/json" -d "$body"
}

check_db() {
  local id=$1 expected_status=$2 expected_type=$3 expected_del=$4 label=$5
  local row
  row=$(docker exec -i mysql mysql -u root -p123456 masterdata_db -N -e \
    "SELECT CONCAT(submission_status,',',IFNULL(submission_type,'NULL'),',',IF(delete_time IS NULL,'NULL','HAS_VALUE')) \
     FROM md_administrative_division WHERE id=$id;" 2>/dev/null)
  if [ -z "$row" ]; then
    if [ "$expected_status" = "GONE" ]; then pass "$label: 物理删除"; else fail "$label: 数据不存在"; fi
    return
  fi
  local s t d
  s=$(echo "$row" | cut -d, -f1)
  t=$(echo "$row" | cut -d, -f2)
  d=$(echo "$row" | cut -d, -f3)
  if [ "$s" = "$expected_status" ] && [ "$t" = "$expected_type" ]; then
    if [ -n "$expected_del" ] && [ "$d" != "$expected_del" ]; then
      fail "$label: delete_time 期望=$expected_del 实际=$d"
    else
      pass "$label: status=$s, type=$t, delete_time=$d"
    fi
  else
    fail "$label: 期望 status=$expected_status,type=$expected_type; 实际 status=$s,type=$t"
  fi
}

echo "======================================================"
echo "  基层组织审批流程测试"
echo "======================================================"
echo ""

# ---------- 场景1: 新增→提交→通过 ----------
echo "【场景1】新增 → 提交 → 审核通过"
echo "  步骤1.1: 创建..."
RESP=$(api_post "/divisions" "{\"parent_id\":$PARENT_ID,\"level\":5,\"name\":\"审批测试-场景1\",\"code\":\"TST0000000001\",\"sort_order\":1}")
ID1=$(echo "$RESP" | grep -o '"id":[0-9]*' | grep -o '[0-9]*')
echo "  ID=$ID1"
[ -z "$ID1" ] && { fail "创建失败: $RESP"; } || check_db "$ID1" 0 1 "" "创建后"

echo "  步骤1.2: 提交..."
api_post "/divisions/$ID1/submit" > /dev/null
check_db "$ID1" 1 1 "" "提交后"

echo "  步骤1.3: 审核通过..."
api_post "/approval/administrative_division/$ID1/review" '{"action":"approve","review_notes":"测试通过"}' > /dev/null
check_db "$ID1" 2 1 NULL "通过后"
echo ""

# ---------- 场景2: 新增→提交→拒绝(物理删除) ----------
echo "【场景2】新增 → 提交 → 审核拒绝（应物理删除）"
echo "  步骤2.1: 创建..."
RESP=$(api_post "/divisions" "{\"parent_id\":$PARENT_ID,\"level\":5,\"name\":\"审批测试-场景2\",\"code\":\"TST0000000002\",\"sort_order\":2}")
ID2=$(echo "$RESP" | grep -o '"id":[0-9]*' | grep -o '[0-9]*')
echo "  ID=$ID2"
[ -z "$ID2" ] && { fail "创建失败: $RESP"; } || check_db "$ID2" 0 1 "" "创建后"

echo "  步骤2.2: 提交..."
api_post "/divisions/$ID2/submit" > /dev/null
check_db "$ID2" 1 1 "" "提交后"

echo "  步骤2.3: 拒绝..."
api_post "/approval/administrative_division/$ID2/review" '{"action":"reject","review_notes":"测试拒绝"}' > /dev/null
check_db "$ID2" GONE "" "" "拒绝后"
echo ""

# ---------- 场景3: 发起删除→提交→通过(软删除) ----------
echo "【场景3】发起删除 → 提交 → 审核通过（应设delete_time）"
echo "  步骤3.1: 创建并批准..."
RESP=$(api_post "/divisions" "{\"parent_id\":$PARENT_ID,\"level\":5,\"name\":\"审批测试-场景3\",\"code\":\"TST0000000003\",\"sort_order\":3}")
ID3=$(echo "$RESP" | grep -o '"id":[0-9]*' | grep -o '[0-9]*')
echo "  ID=$ID3"
api_post "/divisions/$ID3/submit" > /dev/null
api_post "/approval/administrative_division/$ID3/review" '{"action":"approve","review_notes":"先批准"}' > /dev/null
check_db "$ID3" 2 1 NULL "批准后"

echo "  步骤3.2: 发起删除..."
api_post "/divisions/$ID3/request-delete" > /dev/null
check_db "$ID3" 0 3 "" "发起删除后"

echo "  步骤3.3: 提交删除..."
api_post "/divisions/$ID3/submit" > /dev/null
check_db "$ID3" 1 3 "" "提交后"

echo "  步骤3.4: 通过删除..."
api_post "/approval/administrative_division/$ID3/review" '{"action":"approve","review_notes":"同意删除"}' > /dev/null
check_db "$ID3" 2 3 HAS_VALUE "通过后(delete_time)"
docker exec -i mysql mysql -u root -p123456 masterdata_db -e "DELETE FROM md_administrative_division WHERE id=$ID3;" 2>/dev/null
echo ""

# ---------- 场景4: 发起删除→提交→拒绝(恢复) ----------
echo "【场景4】发起删除 → 提交 → 审核拒绝（恢复status=2, type清除）"
echo "  步骤4.1: 创建并批准..."
RESP=$(api_post "/divisions" "{\"parent_id\":$PARENT_ID,\"level\":5,\"name\":\"审批测试-场景4\",\"code\":\"TST0000000004\",\"sort_order\":4}")
ID4=$(echo "$RESP" | grep -o '"id":[0-9]*' | grep -o '[0-9]*')
echo "  ID=$ID4"
api_post "/divisions/$ID4/submit" > /dev/null
api_post "/approval/administrative_division/$ID4/review" '{"action":"approve","review_notes":"先批准"}' > /dev/null
check_db "$ID4" 2 1 NULL "批准后"

echo "  步骤4.2: 发起删除..."
api_post "/divisions/$ID4/request-delete" > /dev/null
check_db "$ID4" 0 3 "" "发起删除后"

echo "  步骤4.3: 提交删除..."
api_post "/divisions/$ID4/submit" > /dev/null
check_db "$ID4" 1 3 "" "提交后"

echo "  步骤4.4: 拒绝删除..."
api_post "/approval/administrative_division/$ID4/review" '{"action":"reject","review_notes":"不同意删除"}' > /dev/null
check_db "$ID4" 2 NULL NULL "拒绝后(status=2, type=NULL)"
docker exec -i mysql mysql -u root -p123456 masterdata_db -e "DELETE FROM md_administrative_division WHERE id=$ID4;" 2>/dev/null
echo ""

# ---------- 场景5: 批量新增→提交→批量拒绝(物理删除) ----------
echo "【场景5】批量新增 → 提交 → 批量拒绝（应物理删除）"
echo "  步骤5.1: 创建两条..."

RESP_A=$(api_post "/divisions" "{\"parent_id\":$PARENT_ID,\"level\":5,\"name\":\"批量测试A\",\"code\":\"TST0000000005\",\"sort_order\":5}")
ID5A=$(echo "$RESP_A" | grep -o '"id":[0-9]*' | grep -o '[0-9]*')

RESP_B=$(api_post "/divisions" "{\"parent_id\":$PARENT_ID,\"level\":5,\"name\":\"批量测试B\",\"code\":\"TST0000000006\",\"sort_order\":6}")
ID5B=$(echo "$RESP_B" | grep -o '"id":[0-9]*' | grep -o '[0-9]*')
echo "  ID=$ID5A, $ID5B"

api_post "/divisions/$ID5A/submit" > /dev/null
api_post "/divisions/$ID5B/submit" > /dev/null
check_db "$ID5A" 1 1 "" "提交后A"
check_db "$ID5B" 1 1 "" "提交后B"

echo "  步骤5.2: 批量拒绝..."
api_post "/approval/batch-review" "{\"entity_type\":\"administrative_division\",\"ids\":[$ID5A,$ID5B],\"action\":\"reject\",\"review_notes\":\"批量拒绝\"}" > /dev/null
check_db "$ID5A" GONE "" "" "批量拒绝后A"
check_db "$ID5B" GONE "" "" "批量拒绝后B"
echo ""

# 清理场景1
docker exec -i mysql mysql -u root -p123456 masterdata_db -e "DELETE FROM md_administrative_division WHERE id=$ID1;" 2>/dev/null

echo "======================================================"
if [ "$FAILED" = "1" ]; then
  echo -e "  ${RED}结果: 存在失败用例${NC}"
else
  echo -e "  ${GREEN}结果: 全部通过 ✓${NC}"
fi
echo "======================================================"
