# Harness Pipeline Patterns

> 流水线最佳实践 · 如何添加新检查、如何扩展技能 · 最后更新 2026-08-13

---

## 1. 概述

本文档提供 Harness Pipeline 的扩展指南和最佳实践，帮助开发者：
- 添加新的机械化检查项
- 扩展 Review 维度
- 新增 Agent 类型
- 优化检查性能
- 调试 Pipeline 问题

---

## 2. 添加新的机械化检查

### 2.1 Go 服务检查扩展

**场景**: 需要新增一个检查项，例如"检测 API Handler 是否有速率限制"。

#### Step 1: 编写检查函数

在 `.harness/skills/qa/scripts/harness-checks.sh` 中添加：

```bash
# ─── Check 17: API Rate Limiting ──────────────────────────────

check_api_rate_limiting() {
  echo "[17/17] API rate limiting" >&2
  local target="$PROJECT_ROOT/services"
  [[ -n "$SERVICE_NAME" ]] && target="$PROJECT_ROOT/services/$SERVICE_NAME"

  if [[ ! -d "$target" ]]; then
    log_pass "api_rate_limiting" "no service directory (skipped)"
    return
  fi

  local violations=()
  
  # 查找所有 Handler 函数
  while IFS= read -r handler_file; do
    [[ -z "$handler_file" ]] && continue
    [[ ! -f "$handler_file" ]] && continue
    
    # 检查是否有 ratelimit.Limit 调用
    if ! grep -q 'ratelimit\.Limit' "$handler_file"; then
      local rel="${handler_file#$PROJECT_ROOT/}"
      violations+=("$rel")
    fi
  done < <(find "$target" -path '*/api/internal/handler/*.go' -not -name 'routes.go' -type f)

  if [[ ${#violations[@]} -eq 0 ]]; then
    log_pass "api_rate_limiting" "all API handlers have rate limiting"
  else
    local detail
    detail="$(printf '%s; ' "${violations[@]}" | head -c 2000)"
    detail="$(json_escape "$detail")"
    log_warn "api_rate_limiting" "${#violations[@]} handlers without rate limiting: $detail"
  fi
}
```

#### Step 2: 注册到主流程

在 `main()` 函数中调用：

```bash
main() {
  # ... 现有检查 ...
  check_api_rate_limiting  # ← 新增
  
  # ... 统计和输出 ...
}
```

#### Step 3: 更新检查项计数

```bash
# 顶部注释
# Checks:
#   1. go build ./...
#   ...
#  17. API rate limiting  — ← 新增

# main() 中的输出
local labels=("go build" "go vet" ... "API rate limiting")  # ← 新增
```

#### Step 4: 更新文档

1. **SKILL.md** - 更新检查项表格
2. **pipeline-architecture.md** - 更新 Layer 1 检查项列表
3. **pipeline-evolution.md** - 添加新的 Phase 记录改进

#### Step 5: 测试

```bash
# 测试单个服务
bash .harness/skills/qa/scripts/harness-checks.sh --service user-service

# 测试 JSON 输出
bash .harness/skills/qa/scripts/harness-checks.sh --service user-service --json | jq .

# 测试差分扫描（先修改代码）
bash .harness/skills/qa/scripts/harness-checks.sh --service user-service

# 测试全量扫描
bash .harness/skills/qa/scripts/harness-checks.sh --service user-service --full
```

### 2.2 前端检查扩展

**场景**: 添加 a11y (无障碍) 检查。

#### Step 1: 编写检查函数

在 `.harness/skills/qa/scripts/harness-checks-frontend.sh` 中添加：

```bash
# ─── Check 7: Accessibility ───────────────────────────────────

check_a11y() {
  echo "[7/7] Accessibility check" >&2
  
  local search_dir="${TARGET_DIR:-$WEB_DIR}"
  local violations=()
  
  # 检测常见 a11y 问题
  local patterns=(
    "<img[^>]*(?!alt=)"  # img 缺少 alt
    "<button[^>]*(?!aria-label=).*>\s*<[^>]+>\s*</button>"  # 图标按钮缺 aria-label
    "role=\"button\"[^>]*(?!tabindex=)"  # role=button 缺 tabindex
  )
  
  for pattern in "${patterns[@]}"; do
    while IFS= read -r match; do
      [[ -z "$match" ]] && continue
      local file="${match%%:*}"
      [[ "$file" == *"node_modules"* ]] && continue
      
      local rel="${file#$PROJECT_ROOT/}"
      local line_num
      line_num=$(echo "$match" | cut -d: -f2)
      violations+=("$rel:$line_num")
    done < <(grep -rnPE "$pattern" "$search_dir/src" --include='*.vue' --include='*.tsx' 2>/dev/null || true)
  done
  
  if [[ ${#violations[@]} -eq 0 ]]; then
    log_pass "a11y" "no accessibility issues detected"
  else
    local detail
    detail="$(printf '%s; ' "${violations[@]}" | head -c 2000)"
    detail="$(json_escape "$detail")"
    log_warn "a11y" "${#violations[@]} potential issues: $detail"
  fi
}
```

#### Step 2-5: 同 Go 服务流程

---

## 3. 扩展 Review 维度

### 3.1 添加新的审查维度

**场景**: 需要新增"性能优化"维度（Performance）。

#### Step 1: 定义维度

在 `.harness/skills/review.md` 中添加：

```markdown
## 维度 #10: 性能优化

**关注点**:
- 数据库查询效率（N+1 问题、缺少索引）
- 算法复杂度（O(n²) 可优化为 O(n)）
- 缓存策略（热数据未缓存）
- 资源泄露（goroutine 泄露、连接未关闭）
- 前端性能（大列表未虚拟滚动、重渲染）

**检查清单**:
- [ ] 数据库查询是否有 N+1 问题
- [ ] 循环中是否有重复的网络调用
- [ ] 是否使用了合适的数据结构
- [ ] 是否有缓存策略
- [ ] 前端是否有性能优化（虚拟滚动、懒加载）
```

#### Step 2: 添加 Review Lens

在 `.harness/workflows/harness-pipeline.js` 的 `REVIEW_LENSES` 中添加：

```javascript
const REVIEW_LENSES = [
  // ... 现有 3 个视角 ...
  {
    key: 'performance',
    label: '性能优化',
    dimensions: '性能优化(#10)',
    focus: isFrontend
      ? '你关注前端性能问题。检查大列表虚拟滚动、懒加载、重渲染优化、API 调用合并、缓存策略。'
      : '你关注后端性能问题。检查数据库 N+1、缺少索引、算法复杂度、缓存策略、goroutine 泄露。',
  },
]
```

#### Step 3: 更新投票规则

如果需要 4 个视角，更新 Pipeline 投票逻辑：

```javascript
// 3/4 PASS → SUCCESS
const passCount = reviewResults.filter(r => r.verdict === 'PASS').length
if (passCount >= 3) {
  return { status: 'SUCCESS', ... }
}
```

#### Step 4: 更新文档

1. **review.md** - 添加维度定义
2. **pipeline-architecture.md** - 更新 Review Agent 表格
3. **pipeline-evolution.md** - 记录扩展原因

### 3.2 调整视角关注点

**场景**: "安全架构"视角需要增加对 CORS 配置的检查。

直接修改 `REVIEW_LENSES` 中的 `focus` 字段：

```javascript
{
  key: 'security-arch',
  label: '安全架构',
  dimensions: '架构一致性(#1)、安全性(#5)、变更完整性(#8)',
  focus: isFrontend
    ? '你关注前端架构的正确性和安全风险。检查组件分层合理性、API 调用权限校验、Token 存储安全、XSS 防护、CORS 配置、硬编码密钥、敏感信息泄露、CHANGELOG 完整性。'  // ← 新增 CORS
    : '...',
}
```

---

## 4. 添加新的 Agent 类型

### 4.1 创建子 Agent 定义

**场景**: 需要一个专门的"迁移评估 Agent"（Migration Evaluator）。

#### Step 1: 创建 Agent 定义文件

`.harness/agents/subagents/migration-evaluator.md`:

```markdown
# Migration Evaluator Agent

你是数据库迁移评估专家，负责审查 Migration 脚本的安全性和正确性。

## 职责

1. **安全性评估**
   - 是否有锁表风险（ALTER TABLE 在生产大表）
   - 是否有数据丢失风险（DROP COLUMN 未备份）
   - 是否有回滚方案

2. **正确性检查**
   - DDL 语法正确性
   - 字段类型和约束合理性
   - 索引设计合理性

3. **影响分析**
   - 预估迁移时间
   - 评估对现有数据的影响
   - 识别依赖此表的服务

## 执行步骤

1. 阅读 Migration 文件（`model/migrations/*.sql`）
2. 检查 DDL 语句
3. 查询表现有数据量（通过 MCP mysql）
4. 评估锁表时间
5. 输出评估报告

## 产出

写入 `.harness/changes/<change>/migration-assessment.md`:

\`\`\`markdown
# Migration Assessment

## 安全性
- 🟢 / 🟡 / 🔴 锁表风险: <评估>
- 🟢 / 🟡 / 🔴 数据丢失风险: <评估>
- 🟢 / 🟡 / 🔴 回滚方案: <评估>

## 影响分析
- 预估迁移时间: <时间>
- 影响的表: <列表>
- 影响的服务: <列表>

## 建议
<具体建议>

---
VERDICT: SAFE / UNSAFE
---
\`\`\`
```

#### Step 2: 注册到 Owner Agent

在 `.harness/agents/owner-agent.md` 的调度表中添加：

```markdown
| # | 阶段 | 执行方式 | 产出 | 门禁 |
|---|------|:---:|------|------|
| ... |
| 3.5 | **迁移评估** | 子 Agent (migration-evaluator) | `migration-assessment.md` | VERDICT = SAFE |
```

#### Step 3: Owner Agent 调用

```javascript
// 检测是否有 Migration 文件
const hasMigration = await detectMigrationFiles(changeDir)

if (hasMigration) {
  const assessment = await agent({
    subagent_type: 'general-purpose',
    prompt: `阅读 .harness/agents/subagents/migration-evaluator.md，执行迁移评估任务。`,
  })
  
  if (assessment.includes('VERDICT: UNSAFE')) {
    // 阻塞流程，要求修正
  }
}
```

### 4.2 创建 Workflow 子步骤

**场景**: 在 harness-pipeline.js 中添加"性能测试"步骤。

#### Step 1: 定义 Schema

```javascript
const PERF_TEST_SCHEMA = {
  type: 'object',
  properties: {
    verdict: { type: 'string', enum: ['PASS', 'FAIL'] },
    benchmarks: {
      type: 'array',
      items: {
        type: 'object',
        properties: {
          name: { type: 'string' },
          ns_per_op: { type: 'number' },
          status: { type: 'string', enum: ['PASS', 'WARN', 'FAIL'] },
        },
      },
    },
  },
  required: ['verdict', 'benchmarks'],
}
```

#### Step 2: 编写 Prompt 函数

```javascript
function perfTestPrompt() {
  return `你是 Performance Test Agent，负责运行性能测试并分析结果。

## 执行步骤
1. 运行 Benchmark: \`go test ./... -bench=. -benchmem -benchtime=3s\`
2. 对比 baseline（如存在）
3. 分析性能退化
4. 输出结构化结果

## 阻塞规则
- >100% 退化 → FAIL
- 50-100% 退化 → WARN
- <50% 退化 → PASS
`
}
```

#### Step 3: 集成到 Pipeline

```javascript
// 在 Review 之后添加性能测试
phase('PerfTest')
const perfResult = await agent(perfTestPrompt(), {
  schema: PERF_TEST_SCHEMA,
  label: 'Performance Test',
})

if (perfResult.verdict === 'FAIL') {
  // 触发修复循环
}
```

---

## 5. 性能优化模式

### 5.1 差分扫描优化

**问题**: 全量扫描耗时长（Go 服务 ~50s）

**方案**: 默认使用差分扫描，仅检查变更文件

```bash
# 获取变更文件
changed_files() {
  local pattern="${1:-*.go}"
  (git diff --name-only HEAD; git diff --cached --name-only; git ls-files --others --exclude-standard) \
    | sort -u | grep -E "\.(${pattern//\*/})\$" 2>/dev/null || true
}

# 在检查函数中使用
if $FULL_SCAN; then
  go_files=$(find "$search_dir" -name '*.go' -not -name '*_test.go' -not -path '*/vendor/*' | sort)
else
  go_files=$(cd "$PROJECT_ROOT" && changed_files 'go' | sed "s|^|$PROJECT_ROOT/|" | sort)
  if [[ -z "$go_files" ]]; then
    log_pass "json_string" "no Go changes in diff (skipped)"
    return
  fi
fi
```

**效果**: 增量检查时间从 50s → 5-8s

### 5.2 并行检查

**问题**: 18 项检查串行执行，总耗时 = Σ 单项耗时

**方案**: 独立检查项并行执行（需要重构脚本）

```bash
# 使用后台任务
check_go_build &
check_go_vet &
check_proto_jstype &
check_json_string &

# 等待所有后台任务
wait

# 合并结果
```

**注意**: 需要处理 RESULTS 数组的并发写入（使用临时文件 + 合并）

### 5.3 缓存优化

**问题**: `go test` 每次重新执行所有测试

**方案**: 利用 Go test cache（默认启用）

```bash
# 不使用 -count=1（禁用缓存）
go test ./...  # ← 使用缓存

# 仅在需要 FRESH run 时禁用缓存（QA Agent）
go test ./... -count=1
```

**效果**: 测试时间从 20s → 2s（无变更时）

---

## 6. 调试与故障排查

### 6.1 机械化检查调试

**场景**: 检查项误报，需要定位原因。

#### 技巧 1: 单项检查调试

修改脚本临时只执行目标检查：

```bash
main() {
  # 注释掉其他检查
  # check_go_build
  # check_go_vet
  check_json_string  # ← 只执行这一项
  
  # 输出详细调试信息
  set -x  # 打印每条命令
}
```

#### 技巧 2: 手动复现检查逻辑

```bash
# 提取检查逻辑到临时脚本
cd services/user-service

# 复现 json:",string" 检查
grep -rnP '\w+Id\s+int64.*json:"' . --include='*.go' \
  | grep -v 'json:"[^"]*string'
```

#### 技巧 3: JSON 输出调试

```bash
# 使用 jq 过滤特定检查项
bash .harness/skills/qa/scripts/harness-checks.sh --service user-service --json \
  | jq '.results[] | select(.check == "json_string")'
```

### 6.2 Pipeline 调试

**场景**: harness-pipeline.js 卡在某个阶段。

#### 技巧 1: 查看 Workflow 日志

```bash
# Workflow 输出文件位于
ls -lt .harness/.claude/workflow-outputs/

# 读取最新的输出
cat .harness/.claude/workflow-outputs/wf_*.output
```

#### 技巧 2: 手动触发 Agent

```javascript
// 临时脚本测试 Generator Prompt
const prompt = generatorPrompt(1, '', 'feature')
console.log(prompt)

// 手动调用 Agent 测试
await agent(prompt, { label: 'Test Generator' })
```

#### 技巧 3: 分阶段执行

修改 Pipeline 只执行特定阶段：

```javascript
// 跳过 Generator，直接测试 QA
// const genResult = await agent(generatorPrompt(...))
const genResult = { /* 模拟结果 */ }

phase('QA')
const qaResult = await agent(qaPrompt(), { schema: QA_SCHEMA })
// ... 继续调试 ...
```

### 6.3 常见问题诊断

#### 问题 1: QA 一直 FAIL，但本地测试通过

**原因**: QA Agent 缓存了旧的测试结果

**解决**:
```bash
# 清理 Go test cache
go clean -testcache

# 清理 build cache
go clean -cache

# 重新运行 QA
bash .harness/skills/qa/scripts/harness-checks.sh --service <name>
```

#### 问题 2: Review FAIL，但看不出问题

**原因**: Review Agent 基于过期的 design.md

**解决**:
```bash
# 检查 design.md 是否最新
cat services/<name>/docs/design.md

# 同步知识图谱
bash .harness/scripts/graph-sync.sh

# 重新运行 Review
```

#### 问题 3: Memory 搜索遗漏相关记忆

**原因**: 
1. 记忆索引过期
2. triggers 关键词不匹配

**解决**:
```bash
# 重建索引
bash .harness/scripts/memory-index-build.sh

# 测试查询
bash .harness/scripts/memory-index-query.sh --union <keyword1> <keyword2>

# 如果仍未命中，更新记忆文件的 triggers
```

---

## 7. 最佳实践清单

### 7.1 编写机械化检查

- ✅ 提供清晰的 echo 提示（`echo "[N/Total] Check Name" >&2`）
- ✅ 支持差分扫描（`$FULL_SCAN` 标志）
- ✅ 使用 `log_pass/log_fail/log_warn` 记录结果
- ✅ 详细错误信息（文件名:行号:字段名）
- ✅ JSON 输出友好（`json_escape` 转义）
- ✅ 跳过不适用场景（返回 `skipped` 而非 FAIL）
- ❌ 避免硬编码路径（使用 `$PROJECT_ROOT` 变量）
- ❌ 避免阻塞式检查（超时保护）

### 7.2 编写 Agent Prompt

- ✅ 清晰的角色定义（"你是 XX Agent，负责 YY"）
- ✅ 明确的执行步骤（1-2-3 列表）
- ✅ 具体的产出格式（markdown 示例）
- ✅ 明确的 VERDICT 规则（PASS/FAIL 条件）
- ✅ 约束说明（只读权限、禁止行为）
- ❌ 避免模糊指令（"合理"、"适当"）
- ❌ 避免过长 prompt（>2000 字需拆分）

### 7.3 扩展 Review 维度

- ✅ 维度职责明确（不重叠）
- ✅ 检查清单具体可执行
- ✅ 优先级明确（CRITICAL vs WARNING）
- ✅ 与现有维度分工清晰
- ❌ 避免"全能"维度（职责过多）
- ❌ 避免主观判断（"代码优雅"）

### 7.4 Pipeline 性能优化

- ✅ 优先使用差分扫描
- ✅ 独立检查项可并行
- ✅ 利用工具缓存（go test cache）
- ✅ 提前终止（第一个 FAIL 即停止）
- ❌ 避免全量扫描作为默认
- ❌ 避免重复检查（DRY）

---

## 8. 代码模板

### 8.1 机械化检查模板

```bash
# ─── Check N: <检查名称> ──────────────────────────────────────

check_<name>() {
  echo "[N/Total] <检查名称>" >&2
  local target="$PROJECT_ROOT/services"
  [[ -n "$SERVICE_NAME" ]] && target="$PROJECT_ROOT/services/$SERVICE_NAME"

  if [[ ! -d "$target" ]]; then
    log_pass "<name>" "no service directory (skipped)"
    return
  fi

  local violations=()
  
  # 确定扫描范围
  local files
  if $FULL_SCAN; then
    files=$(find "$target" -name '*.go' -type f | sort)
    echo "  (full scan)" >&2
  else
    files=$(cd "$PROJECT_ROOT" && changed_files 'go' | grep "^services/" | sed "s|^|$PROJECT_ROOT/|" | sort)
    if [[ -z "$files" ]]; then
      log_pass "<name>" "no changes in diff (skipped)"
      return
    fi
    echo "  (diff scan: $(echo "$files" | wc -l) files)" >&2
  fi

  # 执行检查逻辑
  while IFS= read -r file; do
    [[ -z "$file" ]] && continue
    [[ ! -f "$file" ]] && continue
    
    # TODO: 实现检查逻辑
    # 如果发现问题：
    # local rel="${file#$PROJECT_ROOT/}"
    # violations+=("$rel:$line_num:$detail")
  done < <(echo "$files")

  # 输出结果
  if [[ ${#violations[@]} -eq 0 ]]; then
    log_pass "<name>" "all checks passed"
  else
    local detail
    detail="$(printf '%s; ' "${violations[@]}" | head -c 2000)"
    detail="$(json_escape "$detail")"
    log_fail "<name>" "${#violations[@]} violations: $detail"
  fi
}
```

### 8.2 Review Lens 模板

```javascript
{
  key: '<lens-key>',
  label: '<视角名称>',
  dimensions: '<维度列表>',
  focus: isFrontend
    ? '你关注<前端焦点>。检查<前端检查项列表>。'
    : '你关注<后端焦点>。检查<后端检查项列表>。',
}
```

### 8.3 Agent Prompt 模板

```javascript
function <agent>Prompt() {
  return `你是 <Agent 名称>，负责<职责>。

## 角色定义
<详细说明>

## 执行步骤
1. <步骤1>
2. <步骤2>
3. <步骤3>

## 产出
写入 <文件路径>：

\`\`\`markdown
# <产出标题>

## <章节1>
<内容>

---
VERDICT: PASS / FAIL
---
\`\`\`

## 约束
- 只读权限：Read、Grep、Glob、Bash
- 严禁 Write、Edit
`
}
```

---

## 9. 测试策略

### 9.1 单元测试（检查脚本）

```bash
# 测试脚本：test-harness-checks.sh

test_json_string_detection() {
  # 创建测试文件
  cat > /tmp/test.go <<'EOF'
type User struct {
    UserId int64 `json:"user_id"`  // ← 缺少 ,string
}
EOF

  # 运行检查
  result=$(check_json_string /tmp/test.go)
  
  # 断言
  if [[ "$result" == *"FAIL"* ]]; then
    echo "✅ json_string detection works"
  else
    echo "❌ json_string detection failed"
    exit 1
  fi
}

test_json_string_detection
```

### 9.2 集成测试（Pipeline）

```bash
# 测试完整流程
cd /tmp/test-service
git init
# ... 创建测试代码 ...

# 运行 Pipeline
bash .harness/workflows/harness-pipeline.js \
  --service test-service \
  --task "实现用户登录"

# 验证产出
[[ -f _qa.md ]] || exit 1
[[ -f _review_*.md ]] || exit 1
```

### 9.3 回归测试

```bash
# 记录当前检查结果作为 baseline
bash .harness/skills/qa/scripts/harness-checks.sh --json > baseline.json

# 修改代码后重新检查
bash .harness/skills/qa/scripts/harness-checks.sh --json > current.json

# 对比结果
diff <(jq -S . baseline.json) <(jq -S . current.json)
```

---

## 10. 治理与进化模式

> 对应 `pipeline-evolution.md` Phase 17（Incident 机制 + 流水线检视）与 Phase 14（dispatch）。

### 10.1 Incident 模式（问题驱动进化）

流水线自身的问题也按「Incident」登记，问题驱动进化，避免「进化空转」：

1. 检视/运行中发现流水线缺陷 → 记入 Incident 记录（含现象、根因、整改）。
2. `evolve-pipeline.sh` 读取 Incident，按优先级驱动整改。
3. Incident 数量低于阈值时主动触发检视，保证闭环不空转。

### 10.2 检视模式（pipeline-review）

对流水线自身做可重复的「主动检视」：

1. 以 `harness-design-principles.md` 的 16 条原则为标尺，逐环节映射。
2. 检查目录规范、引擎与策略分离、硬编码残留。
3. 检视结果闭环整改（首轮即抓出 hardcoded secrets 漏检等真实问题）。

### 10.3 引擎策略分离模式

通用引擎逻辑与项目特有策略分离（`harness-design-principles.md` 原则 16）：

- **引擎**：go build/vet/test、硬编码密钥、TODO 桩等——可跨项目复用。
- **策略**：Snowflake ID 精度、5 位 errx 错误码、跨服务仅 gRPC 等——迁移时需替换（见 `.harness/config/project-policies.md`）。
- **实践**：服务名映射外提 `registry/services.json`（单一数据源），项目策略在 `harness-checks.sh` 头部标注归属。

### 10.4 全流程自动化模式（spec-pipeline HITL/resume）

构建「规范驱动全流程自动化」流水线的模式（`harness-spec-pipeline.js` 已验证）：

- **阶段状态机**：`while(currentStage<=6)` + 每阶段一个函数，阶段间用 ctx 对象传结构化状态。
- **HITL 暂停**：每阶段末 `pauseForInput(checkpoint, payload)` → 返回 `{status:'need_input', ctx, questions}`，Owner 问用户后 resume。
- **Resume 传递状态**：沙箱无 fs → 完整 ctx 经 `args.resumeState` 传回，`resumeWith.decisions` 注入决策。
- **失败自动回退**：评审 ≤1/3 自动回退阶段 1（rounds 累加），3 轮后 `stage2_escalate` 升级人工。
- **门禁分层**：沙箱内纯逻辑门禁（投票/轮次）+ 信任 agent 报告（traceability/tasksCount）；文件级门禁由子 Agent 完整环境跑 harness-checks。
- **编码委托**：阶段 5 用 HITL 委托 Owner 启动 N×`harness-pipeline.js`（不嵌套调用——沙箱限制）。

---

## 11. 关联资源

| 资源 | 路径 |
|------|------|
| 架构设计文档 | `.harness/docs/pipeline-architecture.md` |
| 演进日志 | `.harness/docs/pipeline-evolution.md` |
| Go 检查脚本 | `.harness/skills/qa/scripts/harness-checks.sh` |
| 前端检查脚本 | `.harness/skills/qa/scripts/harness-checks-frontend.sh` |
| Pipeline 脚本 | `.harness/workflows/harness-pipeline.js` |
| QA 技能定义 | `.harness/skills/qa/SKILL.md` |
| Review 技能定义 | `.harness/skills/review.md` |
| Owner Agent | `.harness/agents/owner-agent.md` |

---

**维护者**: 开发团队  
**更新频率**: 添加新模式时更新  
**贡献指南**: 发现有效模式后提交 PR 更新本文档
