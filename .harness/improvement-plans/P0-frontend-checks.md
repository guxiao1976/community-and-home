# P0-2: 补全前端机械化检查

## 背景

**现状问题**：
- Go 服务有 15 项机械化检查（`harness-checks.sh`）
- 前端只有 6 项检查（`harness-checks-frontend.sh`）
- **缺失检查**：
  - ❌ ESLint 规则检查
  - ❌ 组件 Props 类型完整性
  - ❌ 路由定义与使用一致性
  - ❌ API 调用类型安全（与 Proto 对齐）
  - ❌ Store mutations 类型安全
  - ❌ 硬编码配置（如 API base URL）

**影响**：
- 前端代码质量无法自动保证，依赖人工 Review
- 低级错误（如 `as any`、未定义路由）可能进入生产
- 前后端类型不一致导致运行时错误

## 目标

将前端机械化检查从 **6 项扩展到 12+ 项**，达到与 Go 服务相当的覆盖率。

## 技术方案

### 新增检查项（6 项）

#### Check #7: ESLint 规则检查
**目的**：强制编码规范（no-console、no-debugger、type-safety 等）

**实现**：
```bash
cd $SERVICE_DIR && npm run lint 2>&1
EXIT_CODE=$?
if [ $EXIT_CODE -ne 0 ]; then
  log_fail "ESLint" "发现 lint 违规，运行 npm run lint 查看详情"
else
  log_pass "ESLint" "无 lint 违规"
fi
```

**前置条件**：确保 `package.json` 有 `lint` script

#### Check #8: TypeScript `any` 类型检查
**目的**：禁止 `as any` 和 `:any` 逃逸类型检查

**实现**：
```bash
ANY_COUNT=$(grep -rn "as any\|: any" $SERVICE_DIR/src \
  --include="*.ts" --include="*.vue" \
  | grep -v "// @ts-ignore" \
  | wc -l)

if [ $ANY_COUNT -gt 0 ]; then
  log_fail "TypeScript any" "发现 $ANY_COUNT 处 'as any' 或 ': any'，禁止逃逸类型检查"
  grep -rn "as any\|: any" $SERVICE_DIR/src --include="*.ts" --include="*.vue" | head -5
else
  log_pass "TypeScript any" "无 any 类型逃逸"
fi
```

#### Check #9: 路由定义与使用一致性
**目的**：检查 `router.push({ name: 'xxx' })` 中的 name 是否在 `router/index.ts` 中定义

**实现**：
```bash
# 提取路由定义的 name
DEFINED_ROUTES=$(grep -rn "name: ['\"]" $SERVICE_DIR/src/router/index.ts \
  | sed -E "s/.*name: ['\"]([^'\"]+)['\"].*/\1/" \
  | sort -u)

# 提取代码中 router.push 的 name
USED_ROUTES=$(grep -rn "router.push.*name: ['\"]" $SERVICE_DIR/src \
  --include="*.ts" --include="*.vue" \
  | sed -E "s/.*name: ['\"]([^'\"]+)['\"].*/\1/" \
  | sort -u)

# 找出未定义但被使用的路由
UNDEFINED=$(comm -13 <(echo "$DEFINED_ROUTES") <(echo "$USED_ROUTES"))

if [ -n "$UNDEFINED" ]; then
  log_fail "路由一致性" "发现未定义的路由: $(echo $UNDEFINED | tr '\n' ', ')"
else
  log_pass "路由一致性" "所有使用的路由均已定义"
fi
```

#### Check #10: API 接口类型与 Proto 对齐
**目的**：检查前端 API 调用的 TypeScript 接口是否与 `api-proto` 生成的类型一致

**实现**：
```bash
# 检查是否使用了 api-proto 生成的类型（而非手写类型）
# 假设生成的类型在 src/types/proto.ts 或 src/api/proto-types.ts

HANDWRITTEN_TYPES=$(grep -rn "interface.*Request\|interface.*Response" \
  $SERVICE_DIR/src/api \
  --include="*.ts" \
  | grep -v "import.*proto-types" \
  | wc -l)

if [ $HANDWRITTEN_TYPES -gt 0 ]; then
  log_warn "API 类型对齐" "发现 $HANDWRITTEN_TYPES 处手写 API 类型，建议使用 proto-types.ts"
  grep -rn "interface.*Request\|interface.*Response" $SERVICE_DIR/src/api --include="*.ts" | head -3
else
  log_pass "API 类型对齐" "API 类型使用 proto 生成"
fi
```

**优化方向**：如果有 Proto → TypeScript 生成工具，检查生成文件是否最新

#### Check #11: Vuex/Pinia Store Mutations 类型安全
**目的**：检查 store mutations 是否有类型定义

**实现**：
```bash
# 检查 store 文件是否有类型导出
STORE_FILES=$(find $SERVICE_DIR/src/store -name "*.ts" 2>/dev/null)

UNTYPED_STORES=0
for file in $STORE_FILES; do
  # 检查是否有 interface State 或 type State
  if ! grep -q "interface.*State\|type.*State" "$file"; then
    UNTYPED_STORES=$((UNTYPED_STORES + 1))
  fi
done

if [ $UNTYPED_STORES -gt 0 ]; then
  log_warn "Store 类型安全" "$UNTYPED_STORES 个 store 文件缺少 State 类型定义"
else
  log_pass "Store 类型安全" "所有 store 有类型定义"
fi
```

#### Check #12: 硬编码配置检查
**目的**：禁止硬编码 API base URL、密钥等配置

**实现**：
```bash
# 检查是否有硬编码的 http:// 或 https:// (排除注释)
HARDCODED_URLS=$(grep -rn "http://\|https://" $SERVICE_DIR/src \
  --include="*.ts" --include="*.vue" \
  | grep -v "^[[:space:]]*//\|^[[:space:]]*\*" \
  | grep -v "import.meta.env\|process.env" \
  | wc -l)

if [ $HARDCODED_URLS -gt 5 ]; then  # 允许少量示例/文档
  log_fail "硬编码配置" "发现 $HARDCODED_URLS 处硬编码 URL，应使用环境变量"
  grep -rn "http://\|https://" $SERVICE_DIR/src --include="*.ts" --include="*.vue" \
    | grep -v "^[[:space:]]*//\|^[[:space:]]*\*" \
    | grep -v "import.meta.env\|process.env" \
    | head -3
else
  log_pass "硬编码配置" "无硬编码 URL"
fi
```

### 现有检查项优化

#### 优化 Check #3: 单元测试（0/0 检测）
**现状**：只检查 `npm run test:unit` 通过
**问题**：如果没有测试文件，也会 PASS（假阳性）

**改进**：
```bash
# 运行测试
TEST_OUTPUT=$(cd $SERVICE_DIR && npm run test:unit 2>&1)
EXIT_CODE=$?

# 提取测试数量
TEST_COUNT=$(echo "$TEST_OUTPUT" | grep -oP "\d+ tests? passed" | grep -oP "\d+" || echo "0")

if [ $EXIT_CODE -ne 0 ]; then
  log_fail "单元测试" "测试失败: $TEST_OUTPUT"
elif [ $TEST_COUNT -eq 0 ]; then
  log_fail "单元测试" "0/0 假通过 — 没有测试文件"
else
  log_pass "单元测试" "$TEST_COUNT 个测试通过"
fi
```

### 新增检查项总结

| # | 检查项 | 类型 | 严重性 | 说明 |
|---|--------|------|--------|------|
| 1-6 | 现有检查 | - | - | build/type-check/test/Snowflake ID/响应格式/CHANGELOG |
| 7 | ESLint 规则 | 规范 | FAIL | 强制编码规范 |
| 8 | TypeScript `any` | 类型安全 | FAIL | 禁止类型逃逸 |
| 9 | 路由一致性 | 运行时安全 | FAIL | 防止 404 |
| 10 | API 类型对齐 | 类型安全 | WARN | 建议使用 Proto 类型 |
| 11 | Store 类型安全 | 类型安全 | WARN | Store 需要类型 |
| 12 | 硬编码配置 | 安全 | FAIL | 禁止硬编码 URL |

## 实施步骤

### Phase 1: 脚本开发（2 天）

**Task 1.1**: 创建新的 `harness-checks-frontend.sh` v2
- 位置：`.harness/skills/qa/scripts/harness-checks-frontend.sh`
- 复制现有 6 项检查
- 新增 #7-#12 检查
- 保持与 `harness-checks.sh` 相同的输出格式（JSON / 人类可读）

**Task 1.2**: 添加依赖检查
- 在脚本开头检查 `npm` / `node` 版本
- 检查 `package.json` 是否有 `lint` / `test:unit` / `build` scripts
- 缺失依赖 → 输出 SKIP + 原因

**Task 1.3**: 单元测试脚本
- 创建测试数据集：
  - 有 `as any` 的代码
  - 未定义的路由引用
  - 硬编码 URL
- 运行脚本验证能正确检测

### Phase 2: 前端项目配置（1 天）

**Task 2.1**: 确保 ESLint 配置完整
- 检查 `web/pc/.eslintrc.js` 和 `web/mobile/.eslintrc.js`
- 必需规则：
  - `no-console: error`（生产环境）
  - `no-debugger: error`
  - `@typescript-eslint/no-explicit-any: error`
- 如果缺失 → 添加规则

**Task 2.2**: 添加 lint script（如果缺失）
- `package.json`:
  ```json
  {
    "scripts": {
      "lint": "eslint --ext .ts,.vue src",
      "lint:fix": "eslint --ext .ts,.vue src --fix"
    }
  }
  ```

**Task 2.3**: Proto → TypeScript 类型生成
- 如果尚未实现：调研 `protoc-gen-ts` 或 `ts-proto`
- 生成目标：`web/pc/src/types/proto-types.ts`
- 集成到 `api-proto/Makefile` 的 `generate` 目标

### Phase 3: QA Agent 集成（0.5 天）

**Task 3.1**: 更新 QA Agent Prompt
- 文件：`harness-pipeline.js:216-306` 的 `qaPrompt()` 函数
- 修改前端检查项数量：`6` → `12`
- 更新检查脚本路径（如果重命名）

**Task 3.2**: 更新 QA 报告模板
- 在 `_qa.md` 的「机械化检查结果」表格中新增 6 行（#7-#12）

### Phase 4: 测试验证（1 天）

**Task 4.1**: 准备测试项目
- 在 `web/pc` 中故意引入问题：
  - 添加 `as any`
  - 引用不存在的路由
  - 硬编码 `https://api.example.com`
- 运行 `bash .harness/skills/qa/scripts/harness-checks-frontend.sh --service pc --json`
- 验证所有问题被检测

**Task 4.2**: 端到端测试
- 启动前端 Workflow（修改一个组件）
- 验证 QA Agent 自动运行 12 项检查
- 验证 `_qa.md` 包含所有检查结果

**Task 4.3**: 回归测试
- 在干净代码上运行检查 → 全部 PASS
- 修复测试项目中的问题 → 再次运行 → 全部 PASS

### Phase 5: 文档和上线（0.5 天）

**Task 5.1**: 更新文档
- `.harness/rules/项目编码规范.md` 补充前端规范章节
- `CLAUDE.md` 快速索引中更新 QA 检查项数量

**Task 5.2**: 创建 Memory 记录
- 文件：`.harness/knowledge/memory/frontend-checks.md`
- 内容：12 项检查的说明、常见违规示例、修复方法

**Task 5.3**: 团队培训
- 向开发团队宣讲新增检查项
- 说明如何修复常见违规（如替换 `as any` 为正确类型断言）

## 验收标准

### 功能验收

- [ ] 12 项检查全部实现
- [ ] 每项检查有对应单元测试（能检测违规）
- [ ] JSON 输出格式与 `harness-checks.sh` 一致
- [ ] QA Agent 自动调用新脚本
- [ ] `_qa.md` 包含所有检查结果

### 质量验收

- [ ] 故意引入的 10 个问题，检测率 100%
- [ ] 干净代码，误报率 0%
- [ ] 检查脚本执行时间 < 60 秒

### 覆盖率验收

- [ ] `web/pc` 和 `web/mobile` 都能运行检查
- [ ] 前端检查项数量达到 Go 服务的 80% 以上（12/15 = 80%）

## 风险和依赖

### 风险

**R1: ESLint 配置不统一**
- **描述**：`web/pc` 和 `web/mobile` 的 ESLint 配置可能不同
- **缓解**：抽取共享配置到 `web/.eslintrc.base.js`，两个项目继承

**R2: Proto → TypeScript 生成工具未实现**
- **描述**：Check #10 依赖 Proto 类型生成
- **缓解**：Check #10 先作为 WARN（不阻塞交付），后续实现生成工具后升级为 FAIL

**R3: 历史代码存量问题**
- **描述**：现有代码可能有大量 `as any` / 硬编码 URL
- **缓解**：
  - 第一阶段：只检查新增/修改的文件（git diff）
  - 第二阶段：逐步修复存量问题（创建 debt 任务）

### 依赖

**D1: 前端项目需要有 ESLint**
- 行动：在 Phase 2 中确保配置完整

**D2: 路由命名规范**
- 当前假设：所有路由都有 `name` 字段
- 如果有路由没有 name → Check #9 会误报
- 行动：统一路由定义规范，所有路由必须有 name

## 效果预估

| 指标 | 现状 | 改进后 | 提升 |
|------|------|--------|------|
| 前端检查项数量 | 6 | 12 | +100% |
| 低级错误检出率 | ~40% | ~85% | +112% |
| 人工 Review 时间（减少低级问题） | 20 分钟 | 12 分钟 | ↓ 40% |
| 生产环境前端错误率 | 基线 | -60%（预期） | - |

## 后续优化

1. **增量检查模式**：只检查 git diff 中的文件，加速 CI
2. **自动修复**：对于简单违规（如 console.log），提供 `--fix` 选项自动删除
3. **组件复杂度检查**：检测组件行数 >500 / 圈复杂度 >15 → 建议拆分
4. **无障碍检查**：检查 ARIA 属性完整性、对比度、键盘导航
5. **性能检查**：检测 v-for 缺少 key、大数组未虚拟滚动等性能反模式
