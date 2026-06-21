#!/bin/bash
# 技能集成验证脚本
# 验证已安装的技能是否可用

set -e

echo "=== 技能集成验证 ==="
echo "时间: $(date)"
echo ""

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 检查函数
check_skill() {
    local skill_name="$1"
    local skill_type="$2"

    echo -n "检查技能: $skill_name ... "

    if npx skills list -g 2>/dev/null | grep -q "$skill_name"; then
        echo -e "${GREEN}✓ 已安装${NC}"
        return 0
    else
        echo -e "${RED}✗ 未安装${NC}"
        return 1
    fi
}

# 验证技能文档
check_docs() {
    local skill_name="$1"
    local doc_file=".harness/skills/templates/${skill_name}.md"

    echo -n "检查文档: $skill_name ... "

    if [ -f "$doc_file" ]; then
        echo -e "${GREEN}✓ 存在${NC}"
        return 0
    else
        echo -e "${YELLOW}⚠ 缺失${NC}"
        return 1
    fi
}

echo "【第一步】检查已安装技能状态"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

echo "Phase 1 技能:"
frontend_design_ok=0
webapp_testing_ok=0
audit_website_ok=0

if check_skill "frontend-design" "Example"; then
    frontend_design_ok=1
fi

if check_skill "webapp-testing" "Example"; then
    webapp_testing_ok=1
fi

if check_skill "audit-website" "General"; then
    audit_website_ok=1
fi

echo ""
echo "Phase 2 技能:"
writing_plans_ok=0
requesting_code_review_ok=0

if check_skill "writing-plans" "General"; then
    writing_plans_ok=1
fi

if check_skill "requesting-code-review" "General"; then
    requesting_code_review_ok=1
fi

echo ""
echo "Phase 3 技能:"
pdf_ok=0
xlsx_ok=0

if check_skill "pdf" "Document"; then
    pdf_ok=1
fi

if check_skill "xlsx" "Document"; then
    xlsx_ok=1
fi

echo ""

echo "【第二步】检查技能使用文档"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

echo "Phase 1 文档:"
check_docs "frontend-design"
check_docs "webapp-testing"
check_docs "audit-website"

echo ""
echo "Phase 2 文档:"
check_docs "writing-plans"
check_docs "requesting-code-review"

echo ""
echo "Phase 3 文档:"
check_docs "pdf"
check_docs "xlsx"

echo ""

echo "【第三步】检查流水线集成文件"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

files=(
    ".harness/workflows/skills-integration.js"
    ".harness/scripts/invoke-skill.sh"
    ".harness/docs/skills-evaluation-and-integration.md"
)

all_files_ok=1
for file in "${files[@]}"; do
    echo -n "检查文件: $file ... "
    if [ -f "$file" ]; then
        echo -e "${GREEN}✓ 存在${NC}"

        # 检查脚本可执行权限
        if [[ "$file" == *.sh ]]; then
            if [ -x "$file" ]; then
                echo "  ├─ ${GREEN}✓ 可执行${NC}"
            else
                echo "  ├─ ${YELLOW}⚠ 不可执行${NC}"
                all_files_ok=0
            fi
        fi
    else
        echo -e "${RED}✗ 缺失${NC}"
        all_files_ok=0
    fi
done

echo ""

echo "【第四步】检查 CLAUDE.md 更新"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

claude_files=(
    "web/pc/CLAUDE.md"
)

for file in "${claude_files[@]}"; do
    echo -n "检查 $file 技能文档 ... "
    if grep -q "可用技能" "$file" 2>/dev/null; then
        echo -e "${GREEN}✓ 已更新${NC}"
    else
        echo -e "${YELLOW}⚠ 未更新${NC}"
    fi
done

echo ""

echo "【总结】"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

phase1_skills=$((frontend_design_ok + webapp_testing_ok + audit_website_ok))
phase2_skills=$((writing_plans_ok + requesting_code_review_ok))
phase3_skills=$((pdf_ok + xlsx_ok))
total_skills=$((phase1_skills + phase2_skills + phase3_skills))

echo "Phase 1 技能安装: $phase1_skills/3"
echo "  - frontend-design: $([ $frontend_design_ok -eq 1 ] && echo -e "${GREEN}✓${NC}" || echo -e "${RED}✗${NC}")"
echo "  - webapp-testing: $([ $webapp_testing_ok -eq 1 ] && echo -e "${GREEN}✓${NC}" || echo -e "${RED}✗${NC}")"
echo "  - audit-website: $([ $audit_website_ok -eq 1 ] && echo -e "${GREEN}✓${NC}" || echo -e "${RED}✗${NC}")"

echo ""
echo "Phase 2 技能安装: $phase2_skills/2"
echo "  - writing-plans: $([ $writing_plans_ok -eq 1 ] && echo -e "${GREEN}✓${NC}" || echo -e "${RED}✗${NC}")"
echo "  - requesting-code-review: $([ $requesting_code_review_ok -eq 1 ] && echo -e "${GREEN}✓${NC}" || echo -e "${RED}✗${NC}")"

echo ""
echo "Phase 3 技能安装: $phase3_skills/2"
echo "  - pdf: $([ $pdf_ok -eq 1 ] && echo -e "${GREEN}✓${NC}" || echo -e "${RED}✗${NC}")"
echo "  - xlsx: $([ $xlsx_ok -eq 1 ] && echo -e "${GREEN}✓${NC}" || echo -e "${RED}✗${NC}")"

echo ""
echo "总计: $total_skills/7 技能已安装"

echo ""

if [ $phase1_skills -eq 3 ] && [ $phase2_skills -eq 2 ] && [ $phase3_skills -eq 2 ] && [ $all_files_ok -eq 1 ]; then
    echo -e "${GREEN}✅ Phase 1 & 2 & 3 技能集成验证通过！${NC}"
    echo ""
    echo "下一步："
    echo "1. 在测试项目上运行 harness-pipeline，验证技能自动调用"
    echo "2. 手动测试每个技能，确认输出符合预期"
    echo "3. 收集反馈，优化集成方式"
    echo "4. 准备 Phase 4 实施（企业级文档技能）"
    exit 0
else
    echo -e "${YELLOW}⚠️  技能集成未完全就绪${NC}"
    echo ""
    echo "需要完成："
    if [ $phase1_skills -lt 3 ]; then
        echo "- 安装缺失的 Phase 1 技能"
    fi
    if [ $phase2_skills -lt 2 ]; then
        echo "- 安装缺失的 Phase 2 技能"
    fi
    if [ $phase3_skills -lt 2 ]; then
        echo "- 安装缺失的 Phase 3 技能"
    fi
    if [ $all_files_ok -eq 0 ]; then
        echo "- 修复缺失的集成文件"
    fi
    exit 1
fi
