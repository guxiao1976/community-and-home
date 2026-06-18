# P0-3: 记忆检索效率优化

## 背景

**现状问题**：
- Generator 启动时读取 `MEMORY.md` 全文索引（37 条记忆）
- 逐条匹配 triggers 关键词（线性扫描）
- 命中后读取完整记忆文件（平均 50 行/条）
- **当前开销**：37 条 × 50 行 = ~1850 行，启动时延 ~2-3 秒

**增长趋势**：
- 每月新增记忆：~5 条（基于 QA/Review 反馈）
- 1 年后：37 + 60 = 97 条 → ~4850 行
- 3 年后：37 + 180 = 217 条 → ~10850 行（**超出合理范围**）

**影响**：
- Generator 启动时间线性增长
- 上下文窗口被索引占用（影响代码分析空间）
- 精确匹配效率低下（O(N×M)，N=记忆数，M=关键词数）

## 目标

将记忆检索效率从 **O(N×M)** 优化到 **O(K)**，其中 K 是命中记忆数（通常 2-5 条）。

启动时延从当前 2-3 秒降低到 **<0.5 秒**，且不随记忆增长而退化。

## 技术方案

### 架构设计

```
旧架构（线性扫描）:
  MEMORY.md (37条索引) → 全文加载 → 逐条匹配 triggers → 命中后读记忆文件
  时间复杂度: O(N×M)，空间: O(N)

新架构（倒排索引）:
  .memory-index.json (trigger→slug映射) → 关键词查表 → 直接读命中的记忆文件
  时间复杂度: O(K)，空间: O(1)
```

### 1. 倒排索引设计

**索引文件**：`.harness/knowledge/memory/.memory-index.json`

**格式**：
```json
{
  "version": "1.0",
  "generated_at": "2026-06-18T14:30:00Z",
  "total_memories": 37,
  "index": {
    "proto": ["proto-jstype", "grpc-only-comms", "api-response-single-wrap"],
    "gRPC": ["grpc-only-comms", "grpc-timeout-layers", "grpc-max-msg-size-sensitive-words"],
    "int64": ["proto-jstype"],
    "jstype": ["proto-jstype"],
    "测试": ["testing-discipline", "pre-commit-checks"],
    "test": ["testing-discipline", "llm-connection-test"],
    "migration": ["migration-must-execute"],
    "手机号": ["phone-encryption"],
    "AES": ["phone-encryption"],
    "加密": ["phone-encryption"],
    "响应": ["api-response-single-wrap"],
    "response": ["api-response-single-wrap"]
  },
  "memories": {
    "proto-jstype": {
      "title": "Proto int64 字段必须加 jstype=JS_STRING",
      "file": "proto-jstype.md",
      "severity": "must-follow",
      "type": "guideline",
      "service": "api-proto",
      "triggers": ["proto", "int64", "jstype", "JS_STRING", "Snowflake"]
    },
    "grpc-only-comms": {
      "title": "服务间通信仅通过 gRPC",
      "file": "grpc-only-comms.md",
      "severity": "must-follow",
      "type": "guideline",
      "service": "all",
      "triggers": ["gRPC", "服务间调用", "直连数据库"]
    }
  }
}
```

**索引字段说明**：
- `index`: `trigger → [memory-slug]` 倒排映射（一对多）
- `memories`: `slug → metadata` 记忆元数据（避免读 MEMORY.md）

### 2. 索引构建脚本

**文件**：`.harness/scripts/memory-index-build.sh`

**功能**：
1. 遍历 `.harness/knowledge/memory/*.md` 文件
2. 解析每个文件的 frontmatter（triggers、severity、type、service）
3. 构建倒排索引
4. 写入 `.memory-index.json`

**实现**：
```bash
#!/usr/bin/env bash
set -euo pipefail

MEMORY_DIR=".harness/knowledge/memory"
INDEX_FILE="$MEMORY_DIR/.memory-index.json"

# 临时文件
TMP_INDEX="/tmp/memory-index-$$.json"
TMP_TRIGGERS="/tmp/triggers-$$.txt"

echo '{"version":"1.0","generated_at":"'$(date -u +"%Y-%m-%dT%H:%M:%SZ")'","index":{},"memories":{}}' > "$TMP_INDEX"

TOTAL=0

for md_file in "$MEMORY_DIR"/*.md; do
  # 跳过 MEMORY.md 索引文件本身
  [[ "$(basename "$md_file")" == "MEMORY.md" ]] && continue
  
  SLUG=$(basename "$md_file" .md)
  
  # 提取 frontmatter（使用 awk 提取 --- 之间的内容）
  FRONTMATTER=$(awk '/^---$/{flag=!flag; next} flag' "$md_file")
  
  # 解析字段
  TRIGGERS=$(echo "$FRONTMATTER" | grep "^triggers:" | sed 's/triggers: \[\(.*\)\]/\1/' | tr -d '"' | tr ',' '\n')
  SEVERITY=$(echo "$FRONTMATTER" | grep "^severity:" | awk '{print $2}')
  TYPE=$(echo "$FRONTMATTER" | grep "^type:" | awk '{print $2}')
  SERVICE=$(echo "$FRONTMATTER" | grep "^service:" | awk '{print $2}')
  TITLE=$(echo "$FRONTMATTER" | grep "^title:" | sed 's/title: //')
  
  # 写入 memories 元数据
  jq --arg slug "$SLUG" \
     --arg title "$TITLE" \
     --arg file "$(basename "$md_file")" \
     --arg severity "$SEVERITY" \
     --arg type "$TYPE" \
     --arg service "$SERVICE" \
     --argjson triggers "$(echo "$TRIGGERS" | jq -R -s -c 'split("\n") | map(select(length > 0))')" \
     '.memories[$slug] = {title:$title, file:$file, severity:$severity, type:$type, service:$service, triggers:$triggers}' \
     "$TMP_INDEX" > "$TMP_INDEX.tmp" && mv "$TMP_INDEX.tmp" "$TMP_INDEX"
  
  # 构建倒排索引
  for trigger in $TRIGGERS; do
    [[ -z "$trigger" ]] && continue
    # 将 trigger → slug 添加到 index
    jq --arg trigger "$trigger" --arg slug "$SLUG" \
       '.index[$trigger] = (.index[$trigger] // []) + [$slug] | .index[$trigger] |= unique' \
       "$TMP_INDEX" > "$TMP_INDEX.tmp" && mv "$TMP_INDEX.tmp" "$TMP_INDEX"
  done
  
  TOTAL=$((TOTAL + 1))
done

# 更新总数
jq --arg total "$TOTAL" '.total_memories = ($total | tonumber)' "$TMP_INDEX" > "$INDEX_FILE"

echo "✅ 记忆索引构建完成: $TOTAL 条记忆 → $INDEX_FILE"
rm -f "$TMP_INDEX" "$TMP_TRIGGERS"
```

**触发时机**：
1. 手动：`bash .harness/scripts/memory-index-build.sh`
2. 自动：Git pre-commit hook（当 `.harness/knowledge/memory/*.md` 变更时）
3. CI：每次 PR 自动构建并检查索引新鲜度

### 3. 检索逻辑优化

**位置**：`harness-pipeline.js` 的 `generatorPrompt()` 函数

**旧逻辑（线性扫描）**：
```markdown
### Step A: 搜索相关记忆
1. 读取 .harness/knowledge/memory/MEMORY.md 索引
2. 对每条记忆，检查 triggers 是否与任务关键词匹配
3. 命中的记忆 → 读取完整文件
```

**新逻辑（索引查询）**：
```markdown
### Step A: 搜索相关记忆（索引模式）
1. 从任务描述提取关键词：`keywords = ["gRPC", "Proto", "测试"]`
2. 读取索引文件：`.harness/knowledge/memory/.memory-index.json`
3. 对每个关键词，查询 `index[keyword]` → 获得 `[slug1, slug2, ...]`
4. 去重并按 severity 排序：`must-follow` > `should-follow` > `info`
5. 只读取命中的记忆文件（通常 2-5 个）
```

**Generator Prompt 更新**：
```markdown
## 记忆驱动编码（编码前必须执行）

### Step A: 搜索相关记忆（索引模式）
1. **提取任务关键词**：
   从 tasks.md 提取技术关键词（如 gRPC、Proto、数据库、JWT、Snowflake 等）
   
2. **查询索引**：
   \`\`\`bash
   # 读取索引文件
   INDEX=$(cat .harness/knowledge/memory/.memory-index.json)
   
   # 对每个关键词查询
   for keyword in gRPC Proto 测试; do
     SLUGS=$(echo "$INDEX" | jq -r ".index[\"$keyword\"][]?" 2>/dev/null)
     echo "$SLUGS"
   done | sort -u
   \`\`\`

3. **去重和排序**：
   - 合并所有命中的 slug
   - 从索引的 `memories` 字段读取 severity
   - 按 `must-follow` > `should-follow` > `info` 排序

4. **读取记忆文件**：
   只读取命中的记忆文件，不读 MEMORY.md 全文
   \`\`\`bash
   for slug in $MATCHED_SLUGS; do
     cat .harness/knowledge/memory/$slug.md
   done
   \`\`\`

5. **输出匹配报告**：
   \`\`\`
   搜索关键词: gRPC, Proto, 测试
   命中记忆: 5 条
     - [[grpc-only-comms]] (must-follow)
     - [[proto-jstype]] (must-follow)
     - [[testing-discipline]] (must-follow)
     - [[grpc-timeout-layers]] (must-follow)
     - [[pre-commit-checks]] (must-follow)
   \`\`\`
```

### 4. 索引新鲜度检查

**问题**：索引过期 → 新记忆未被检索到

**解决方案**：在 QA 检查中增加索引新鲜度验证

**位置**：`harness-checks.sh` 新增 Check #16

```bash
# Check 16: Memory Index Freshness
INDEX_FILE=".harness/knowledge/memory/.memory-index.json"
MEMORY_DIR=".harness/knowledge/memory"

if [ ! -f "$INDEX_FILE" ]; then
  log_fail "记忆索引新鲜度" "索引文件不存在，运行 bash .harness/scripts/memory-index-build.sh"
  EXIT_CODE=1
else
  INDEX_TIME=$(jq -r '.generated_at' "$INDEX_FILE")
  INDEX_EPOCH=$(date -d "$INDEX_TIME" +%s 2>/dev/null || date -j -f "%Y-%m-%dT%H:%M:%SZ" "$INDEX_TIME" +%s)
  
  # 找到最新的记忆文件修改时间
  NEWEST_MEMORY=$(find "$MEMORY_DIR" -name "*.md" -not -name "MEMORY.md" -type f -printf '%T@\n' | sort -rn | head -1)
  NEWEST_MEMORY_EPOCH=$(printf "%.0f" "$NEWEST_MEMORY")
  
  if [ "$NEWEST_MEMORY_EPOCH" -gt "$INDEX_EPOCH" ]; then
    log_fail "记忆索引新鲜度" "索引过期（记忆文件比索引新），运行 bash .harness/scripts/memory-index-build.sh"
    EXIT_CODE=1
  else
    log_pass "记忆索引新鲜度" "索引最新 (生成于 $INDEX_TIME)"
  fi
fi
```

### 5. 兼容性保留

**向后兼容**：保留 `MEMORY.md`，但只作为人类阅读索引

- Generator 优先使用 `.memory-index.json`（如果存在）
- 索引不存在时，降级到旧逻辑（读 MEMORY.md）
- 记忆文件格式保持不变（frontmatter + markdown）

## 实施步骤

### Phase 1: 索引构建工具（1 天）

**Task 1.1**: 实现 `memory-index-build.sh`
- 解析 frontmatter（triggers / severity / type / service）
- 构建倒排索引 JSON
- 单元测试：用测试记忆文件验证输出

**Task 1.2**: 实现 `memory-index-query.sh`（辅助工具）
- 输入：关键词列表
- 输出：匹配的 slug 列表（按 severity 排序）
- 用于调试和验证

**Task 1.3**: Git hook 集成
- 创建 `.git/hooks/pre-commit`
- 检测 `.harness/knowledge/memory/*.md` 变更 → 自动重建索引
- 索引变更 → 自动 stage 到 commit

### Phase 2: Generator Prompt 更新（1 天）

**Task 2.1**: 更新 `generatorPrompt()` 函数
- 文件：`.harness/agents/prompts/generator.js` 或 `harness-pipeline.js`
- 替换「Step A: 搜索相关记忆」逻辑
- 改为索引查询模式

**Task 2.2**: 降级逻辑
- 检查 `.memory-index.json` 是否存在
- 不存在 → 降级到旧逻辑（线性扫描 MEMORY.md）
- 存在但过期 → 警告 + 降级

**Task 2.3**: 输出格式调整
- Generator 产出中的「记忆应用报告」保持不变
- 增加「索引查询耗时」字段（用于性能监控）

### Phase 3: QA 检查集成（0.5 天）

**Task 3.1**: 新增 Check #16（索引新鲜度）
- 实现逻辑（见上文）
- 集成到 `harness-checks.sh`

**Task 3.2**: CI 集成
- 在 PR pipeline 中运行 `memory-index-build.sh`
- 检查生成的索引是否与提交的索引一致
- 不一致 → CI FAIL（防止忘记更新索引）

### Phase 4: 测试验证（1 天）

**Task 4.1**: 性能测试
- 准备测试数据集：37 条现有记忆 + 63 条模拟记忆（共 100 条）
- 对比旧逻辑 vs 新逻辑的检索耗时
- 目标：新逻辑 < 0.5 秒，旧逻辑 ~5 秒（100 条时）

**Task 4.2**: 准确性测试
- 测试场景：
  - 单关键词匹配
  - 多关键词匹配（交集）
  - 中文 + 英文关键词混合
  - 关键词大小写不敏感
- 验证：新逻辑检索结果与旧逻辑 100% 一致

**Task 4.3**: 端到端测试
- 启动 Workflow，观察 Generator 启动时间
- 验证记忆应用报告中的记忆列表正确

### Phase 5: 文档和上线（0.5 天）

**Task 5.1**: 更新文档
- `CLAUDE.md` 补充索引构建说明
- `.harness/knowledge/memory/README.md` 说明索引机制

**Task 5.2**: 团队培训
- 宣讲：新增记忆后需要运行 `memory-index-build.sh`
- Git hook 会自动处理，但手动添加时需要注意

**Task 5.3**: 创建 Memory 记录
- 文件：`.harness/knowledge/memory/memory-index-optimization.md`
- 内容：索引机制、构建方法、故障排查

## 验收标准

### 功能验收

- [ ] 索引文件包含所有记忆的 triggers 映射
- [ ] 索引查询返回结果与线性扫描 100% 一致
- [ ] Git hook 自动更新索引
- [ ] QA Check #16 能检测索引过期

### 性能验收

| 记忆数量 | 旧逻辑耗时 | 新逻辑耗时 | 提升 |
|---------|----------|----------|------|
| 37 条 | ~2 秒 | <0.5 秒 | 75% ↓ |
| 100 条 | ~5 秒 | <0.5 秒 | 90% ↓ |
| 200 条 | ~10 秒 | <0.5 秒 | 95% ↓ |

**关键指标**：新逻辑耗时不随记忆数量增长（O(1) vs O(N)）

### 质量验收

- [ ] 100 次查询测试，0 次检索错误
- [ ] 索引过期场景，QA 正确检测并报告
- [ ] 降级逻辑在索引缺失时正常工作

## 风险和依赖

### 风险

**R1: 索引与记忆文件不同步**
- **描述**：手动修改记忆文件后忘记重建索引
- **缓解**：
  - Git hook 自动重建
  - QA Check #16 检测过期索引
  - CI 强制验证索引一致性

**R2: frontmatter 解析失败**
- **描述**：记忆文件格式不规范 → 解析错误
- **缓解**：
  - `memory-index-build.sh` 增加错误处理
  - 解析失败的文件 → 输出警告 + 跳过
  - CI 中运行构建验证所有文件可解析

**R3: 关键词匹配不准确**
- **描述**：用户输入的关键词与 triggers 不匹配 → 漏检
- **缓解**：
  - 定期审查 triggers 覆盖率
  - 增加同义词机制（如 "gRPC" → ["gRPC", "grpc", "RPC"]）
  - 第二级降级：精确匹配失败 → 模糊匹配（正文关键词）

### 依赖

**D1: jq 工具**
- 索引构建和查询依赖 `jq`（JSON 处理）
- 行动：在脚本开头检查 `jq` 是否安装，缺失 → 提示安装

**D2: 记忆文件 frontmatter 规范**
- 所有记忆文件必须有 frontmatter（triggers / severity / type）
- 行动：对现有 37 条记忆进行审查，补全缺失字段

## 效果预估

### 性能提升

| 场景 | 旧逻辑 | 新逻辑 | 提升 |
|------|-------|--------|------|
| Generator 启动时间（37 条记忆） | 2-3 秒 | <0.5 秒 | ↓ 75-83% |
| Generator 启动时间（200 条记忆，3 年后） | ~10 秒 | <0.5 秒 | ↓ 95% |
| 上下文窗口占用 | ~1850 行 | ~150 行 | ↓ 92% |

### 可扩展性

- **当前容量**：37 条记忆
- **3 年后容量**：~200 条记忆
- **索引查询耗时**：O(K)，K 通常 2-5 条，不随总记忆数增长
- **结论**：架构可支撑 500+ 条记忆无性能退化

## 后续优化

1. **模糊匹配**：关键词拼写错误容忍（编辑距离 ≤2）
2. **同义词扩展**：`gRPC` → `["gRPC", "grpc", "RPC", "远程调用"]`
3. **语义搜索**：使用 embedding 向量化 triggers，支持语义相似度匹配
4. **热度排序**：基于 `apply_count` 调整匹配结果排序（常用记忆优先）
5. **增量索引**：只重建变更的记忆，而非全量重建（加速 Git hook）
6. **Web 界面**：记忆检索和管理的可视化工具
