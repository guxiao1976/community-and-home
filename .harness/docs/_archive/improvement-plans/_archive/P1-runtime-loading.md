# P1-3: harness-pipeline.js 运行时加载

## 背景

**现状问题**：
- `harness-pipeline.js` 是 892 行的单文件，包含 Generator/QA/Review/Debug 所有 Prompt
- **构建流程**：`.harness/agents/prompts/{generator,qa,review,debug}.js` → `build-pipeline.sh` 打包 → `harness-pipeline.js`
- **问题**：
  1. **修改一个 Prompt 需要重新构建整个文件**：开发体验差
  2. **Prompt 版本管理困难**：无法单独追溯某个 Prompt 的变更历史
  3. **无法热更新**：运行中的 Workflow 无法加载新 Prompt
  4. **调试困难**：错误堆栈指向打包后的行号，不是源文件行号

**影响**：
- Prompt 迭代周期长（修改 → 构建 → 测试）
- 团队协作冲突（多人同时修改不同 Prompt → 合并冲突）
- 无法快速响应需求变化

## 目标

将 `harness-pipeline.js` 从 **编译时打包** 改为 **运行时动态加载**，实现：
1. 修改 Prompt 无需重新构建
2. 每个 Prompt 独立版本管理
3. 支持热更新（调试模式）
4. 清晰的错误堆栈

## 技术方案

### 架构对比

**旧架构（编译时打包）**：
```
.harness/agents/prompts/generator.js
.harness/agents/prompts/qa.js
.harness/agents/prompts/review.js
.harness/agents/prompts/debug.js
          ↓
bash .harness/scripts/build-pipeline.sh
          ↓
.harness/workflows/harness-pipeline.js (892行)
```

**新架构（运行时加载）**：
```
.harness/workflows/harness-pipeline.js (核心逻辑，~200行)
  ├─ 运行时 import('./prompts/generator.js')
  ├─ 运行时 import('./prompts/qa.js')
  ├─ 运行时 import('./prompts/review.js')
  └─ 运行时 import('./prompts/debug.js')
```

### 1. 目录结构重构

**新结构**：
```
.harness/workflows/
├── harness-pipeline.js          # 核心编排（~200行）
├── prompts/
│   ├── generator.js             # Generator Prompt + 参数处理
│   ├── qa.js                    # QA Prompt + Schema
│   ├── review.js                # Review Prompt + Lenses + Schema
│   ├── debug.js                 # Debug Prompt + Schema
│   └── shared.js                # 共享常量（SVC_DIR, ROOT_CLAUDE等）
├── schemas/
│   ├── qa-schema.js             # QA 验收 Schema
│   ├── review-schema.js         # Review 审查 Schema
│   └── debug-schema.js          # Debug 根因 Schema
└── utils/
    ├── confidence.js            # 置信度计算逻辑
    └── service-resolver.js      # 服务名解析
```

### 2. 运行时加载机制

**核心逻辑**（`harness-pipeline.js`）：

```javascript
export const meta = {
  name: 'harness-pipeline',
  description: 'Harness 开发管线：Generator → QA → Debug → Reviewer',
  phases: [
    { title: 'Develop', detail: 'Generator 实现/修复代码' },
    { title: 'QA', detail: 'QA Agent 编译+测试验证' },
    { title: 'Debug', detail: '根因分析（QA FAIL 时触发）' },
    { title: 'Review', detail: '3 视角并行审查' },
  ],
}

// ── 动态导入 Prompt 模块 ──
const PROMPTS_DIR = '.harness/workflows/prompts'
const SCHEMAS_DIR = '.harness/workflows/schemas'

// 使用动态 import 而非 require（支持 ESM）
async function loadPrompts() {
  const {generatorPrompt} = await import(`${PROMPTS_DIR}/generator.js`)
  const {qaPrompt, QA_SCHEMA} = await import(`${PROMPTS_DIR}/qa.js`)
  const {reviewLensPrompt, REVIEW_LENSES, REVIEW_SCHEMA} = await import(`${PROMPTS_DIR}/review.js`)
  const {debuggingPrompt, DEBUG_SCHEMA} = await import(`${PROMPTS_DIR}/debug.js`)
  const {setSharedContext} = await import(`${PROMPTS_DIR}/shared.js`)
  
  return {generatorPrompt, qaPrompt, reviewLensPrompt, REVIEW_LENSES, debuggingPrompt,
          QA_SCHEMA, REVIEW_SCHEMA, DEBUG_SCHEMA, setSharedContext}
}

// ── 主流程 ──
const prompts = await loadPrompts()

// 设置共享上下文（SVC_DIR, SVC_NAME, ROOT_CLAUDE 等）
prompts.setSharedContext({
  SVC_DIR: args.serviceDir,
  SVC_NAME: args.serviceDir.split('/').pop(),
  ROOT_CLAUDE: 'CLAUDE.md',
  args: args
})

// 使用 Prompt
let iteration = 1
while (iteration <= MAX_ITERATIONS) {
  phase('Develop')
  await agent(prompts.generatorPrompt(iteration, fixContext, TASK_TYPE), {
    label: `${args.serviceName}: 开发`,
    isolation: 'worktree'
  })
  
  phase('QA')
  const qaResult = await agent(prompts.qaPrompt(), {
    label: `QA: ${args.serviceName}`,
    schema: prompts.QA_SCHEMA
  })
  
  // ... 其余逻辑
}
```

### 3. Prompt 模块设计

**Generator Prompt**（`.harness/workflows/prompts/generator.js`）：

```javascript
// 导出共享状态（由 harness-pipeline.js 注入）
let ctx = {}
export function setContext(context) {
  ctx = context
}

export function generatorPrompt(iteration, fixContext, taskType) {
  taskType = taskType || 'feature'
  const isChore = taskType === 'chore'
  const strictTdd = !isChore && taskType !== 'debt'
  const isFrontend = (ctx.SVC_DIR || '').startsWith('web/')
  
  // ... Prompt 逻辑（保持不变）
  
  const base = `你是 ${ctx.args.serviceName} 的开发 Agent。

## 启动上下文
1. 阅读 ${ctx.SVC_DIR}/CLAUDE.md
2. 阅读 ${ctx.SVC_DIR}/docs/design.md
3. 阅读 ${ctx.SVC_DIR}/CHANGELOG.md
...`

  if (iteration === 1) {
    return base + `\n${ctx.args.task}\n\n## 完成标准\n...`
  } else {
    return base + `\n修复问题：\n${fixContext}\n...`
  }
}
```

**优点**：
- 单文件职责单一（只负责 Generator Prompt）
- 修改后无需构建，直接生效
- Git 历史清晰（只改 generator.js，不影响其他文件）

### 4. 热更新支持（可选）

**场景**：调试 Prompt 时，修改后立即重新加载，无需重启 Workflow

**实现**：缓存失效机制

```javascript
// 开发模式：每次调用前清除缓存
const DEV_MODE = process.env.HARNESS_DEV_MODE === 'true'

async function loadPrompts() {
  if (DEV_MODE) {
    // 清除 Node.js 模块缓存
    Object.keys(require.cache).forEach(key => {
      if (key.includes('/prompts/')) {
        delete require.cache[key]
      }
    })
  }
  
  // 重新加载
  const {generatorPrompt} = await import(`${PROMPTS_DIR}/generator.js?t=${Date.now()}`)
  // ... 其他 Prompt
}

// 每轮迭代前重新加载（仅开发模式）
while (iteration <= MAX_ITERATIONS) {
  if (DEV_MODE) {
    prompts = await loadPrompts()
    prompts.setSharedContext(ctx)
  }
  
  // ... 执行逻辑
}
```

**用法**：
```bash
HARNESS_DEV_MODE=true claude workflow run .harness/workflows/harness-pipeline.js \
  --args '{"serviceName":"审核服务", "serviceDir":"services/moderation-service", "task":"..."}'
```

修改 Prompt → 保存 → 下一轮迭代自动加载新 Prompt

### 5. 向后兼容

**兼容策略**：保留 `build-pipeline.sh`，但标记为 deprecated

- 新流程：直接运行 `harness-pipeline.js`（运行时加载）
- 旧流程：运行 `build-pipeline.sh` → 生成单文件版本（供无法使用动态加载的环境）
- 过渡期：两种方式并存，逐步迁移

## 实施步骤

### Phase 1: 目录结构重构（1 天）

**Task 1.1**: 拆分 Prompt 文件
- 从现有 `harness-pipeline.js` 提取 Prompt 函数
- 创建独立文件：`prompts/generator.js`, `qa.js`, `review.js`, `debug.js`
- 保留原有逻辑，只做文件拆分

**Task 1.2**: 创建 Schema 文件
- `schemas/qa-schema.js` 导出 `QA_SCHEMA`
- `schemas/review-schema.js` 导出 `REVIEW_SCHEMA`
- `schemas/debug-schema.js` 导出 `DEBUG_SCHEMA`

**Task 1.3**: 创建工具函数
- `utils/confidence.js` 导出 `computeConfidence()`
- `utils/service-resolver.js` 导出 `resolveService()`, `serviceLabel()`

### Phase 2: 核心逻辑改造（1.5 天）

**Task 2.1**: 实现动态加载
- `harness-pipeline.js` 改为核心编排（~200 行）
- 使用 `import()` 动态加载 Prompt 模块
- 错误处理：加载失败 → 降级到打包版本

**Task 2.2**: 共享上下文机制
- `prompts/shared.js` 导出 `setSharedContext()` / `getContext()`
- 所有 Prompt 从 `getContext()` 读取 `SVC_DIR`, `args` 等

**Task 2.3**: 路径解析
- 处理相对路径（`.harness/workflows/prompts/` vs 从 repo 根目录运行）
- 使用 `import.meta.url` 或 `__dirname`（根据运行环境）

### Phase 3: 开发体验优化（1 天）

**Task 3.1**: 热更新支持
- 实现缓存失效逻辑
- 环境变量 `HARNESS_DEV_MODE` 控制

**Task 3.2**: 错误堆栈优化
- 确保错误堆栈指向源文件（`generator.js:45`）而非打包文件
- 使用 source map（如果需要）

**Task 3.3**: 开发工具
- 脚本：`bash .harness/scripts/dev-pipeline.sh` 启动开发模式
- 监听文件变化 → 自动重新加载（可选，使用 `nodemon` 或 `watchman`）

### Phase 4: 测试验证（1 天）

**Task 4.1**: 单元测试
- 每个 Prompt 文件的单元测试
- 验证输出格式正确

**Task 4.2**: 集成测试
- 运行完整 Workflow，验证动态加载正常工作
- 对比旧版本（打包）vs 新版本（动态加载）输出一致性

**Task 4.3**: 性能测试
- 动态加载的性能开销（首次加载 + 热更新）
- 目标：首次加载 <100ms，热更新 <50ms

### Phase 5: 文档和迁移（0.5 天）

**Task 5.1**: 更新文档
- `CLAUDE.md` 补充新目录结构说明
- 开发指南：如何修改 Prompt、如何调试

**Task 5.2**: 迁移指南
- 告知团队新流程
- 废弃 `build-pipeline.sh`（保留向后兼容，但不推荐使用）

**Task 5.3**: 创建 Memory
- `.harness/knowledge/memory/prompt-runtime-loading.md`

## 验收标准

### 功能验收

- [ ] 修改任一 Prompt 文件，无需构建即可生效
- [ ] 错误堆栈指向源文件（不是打包文件）
- [ ] 开发模式热更新正常工作
- [ ] 所有现有测试通过（输出与旧版本一致）

### 性能验收

| 指标 | 目标 | 实际 |
|------|------|------|
| 首次加载 Prompt 耗时 | <100ms | - |
| 热更新耗时（开发模式） | <50ms | - |
| Workflow 总耗时差异 | <5% | - |

### 开发体验验收

- [ ] 修改 Prompt → 保存 → 5 秒内看到效果（开发模式）
- [ ] Git diff 清晰（只改 generator.js，不影响其他文件）
- [ ] 团队成员可以并行修改不同 Prompt 无冲突

## 风险和依赖

### 风险

**R1: 动态 import 兼容性**
- **描述**：某些环境不支持 ESM 动态 import
- **缓解**：
  - 检测环境，不支持 → 降级到打包版本
  - 提供 CommonJS 兼容版本（`require()` 替代 `import()`）

**R2: 路径解析问题**
- **描述**：从不同目录运行 Workflow，相对路径失效
- **缓解**：
  - 使用绝对路径（基于 `process.cwd()` 或 `import.meta.url`）
  - 在 Workflow meta 中记录基准路径

**R3: 缓存失效导致状态丢失**
- **描述**：热更新清除缓存 → 共享状态丢失
- **缓解**：
  - 共享状态存储在 Workflow 作用域（不在模块作用域）
  - 每次加载后重新注入 `setSharedContext()`

### 依赖

**D1: Node.js ESM 支持**
- 需要 Node.js ≥14（支持 `import()` 动态导入）
- 行动：在脚本开头检查 Node 版本

**D2: Workflow 工具支持 ESM**
- Claude Code 的 Workflow 工具必须支持 ESM 模块
- 行动：验证当前版本兼容性，不兼容 → 使用 CommonJS

## 效果预估

### 开发效率提升

| 场景 | 改进前 | 改进后 | 提升 |
|------|-------|--------|------|
| 修改 Prompt 到生效 | 构建（~10秒）+ 测试 | 保存即生效（0秒构建） | ↓ 100% |
| 调试一个 Prompt 迭代周期 | 5 分钟/轮 | 30 秒/轮 | ↓ 90% |
| 多人协作冲突率 | 50%（单文件冲突） | <5%（多文件独立） | ↓ 90% |

### 维护性提升

| 指标 | 改进前 | 改进后 |
|------|-------|--------|
| Prompt 文件平均行数 | 892（单文件） | ~150（每个模块） |
| Git blame 可读性 | 差（打包后混杂） | 好（每个模块独立） |
| 新 Prompt 添加成本 | 修改主文件 + 重构 | 新建文件 + import |

## 后续优化

1. **Prompt 版本管理**：每个 Prompt 文件增加版本号，运行时记录使用的版本
2. **Prompt 市场**：社区贡献的 Prompt 可以直接 import（如 `import from 'harness-community/prompts/xxx'`）
3. **可视化编辑器**：Web UI 编辑 Prompt，实时预览效果
4. **A/B 测试**：同时加载两个版本的 Prompt，对比效果
5. **Prompt 缓存**：生产环境预编译 Prompt 为字节码，加速加载
