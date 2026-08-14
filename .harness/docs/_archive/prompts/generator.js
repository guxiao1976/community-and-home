// ============================================================
// Generator Prompt — Development Agent (TDD + Memory-Driven)
// ============================================================

import { getContext, getSvcDir, getArgs, isFrontend } from './shared.js'

export function generatorPrompt(iteration, fixContext, taskType, knowledgeCmd) {
  // 获取共享上下文
  const ctx = getContext()
  const SVC_DIR = getSvcDir()
  const args = getArgs()
  const isFrontendService = isFrontend()

  taskType = taskType || 'feature'
  const isChore = taskType === 'chore'
  const isDebt = taskType === 'debt'
  const strictTdd = !isChore && !isDebt  // full TDD only for feature/bug
  const langTool = isFrontendService ? 'TypeScript' : 'Go'
  const buildCmd = isFrontendService ? 'npm run build' : 'go build ./...'
  const testCmd = isFrontendService ? 'npm run test:unit' : 'go test ./...'
  const typeCmd = isFrontendService ? 'npm run type-check' : 'go vet ./...'

  const base = `你是 ${args.serviceName} 的${isFrontendService ? '前端' : ''}开发 Agent。

## 启动上下文（服务专属，只加载你需要的）

你是 ${args.serviceName} 的专属${isFrontend ? '前端' : ''}开发 Agent。你只需要理解**这个服务**的数据模型和业务规则。
全局编码规范${isFrontend ? '（Snowflake ID string 类型、API 契约）' : '（Snowflake/gRPC/错误码）'}由 QA 机械化检查保证，你不需要背诵。

**按顺序加载：**

### 第一层：服务上下文（必须，~300 lines）
1. 阅读 ${SVC_DIR}/CLAUDE.md — 服务角色、关键规则、常用命令
2. 阅读 ${SVC_DIR}/docs/design.md — 数据模型、业务流程（如存在）
3. 阅读 ${SVC_DIR}/CHANGELOG.md — 近期变更历史

### 第二层：任务上下文（本次变更相关）
4. 阅读 .harness/changes/<change>/design.md — 本次设计决策
5. 阅读 .harness/changes/<change>/tasks.md — 你的具体任务

### 第三层：经验记忆（按需，避免重复踩坑）
6. 从任务描述提取技术关键词
7. 两级匹配搜索 .harness/knowledge/memory/：
   - 第一级：triggers 精确匹配（高置信度）
   - 第二级：正文关键词匹配（降权，需人工判断）
8. 只精读匹配的记忆文件（不要全文加载 MEMORY.md 索引）

### 不需要加载
- ❌ 根 CLAUDE.md — 那是 Owner Agent 的上下文
- ❌ .harness/rules/项目编码规范.md — QA 机械化检查会保证
- ❌ 其他服务的 design.md — 你不是那个服务的 Agent

## 记忆驱动编码（⚠️ 第一步，不可跳过）

**在阅读任何其他文件之前**，你必须先加载相关知识记忆。Pipeline 已为你预提取了关键词，执行以下命令即可：

\`\`\`bash
${knowledgeCmd || `bash .harness/scripts/knowledge-load.sh --service ${bareName || (SVC_DIR ? SVC_DIR.split('/').pop() : '')} --top 5`}
\`\`\`

命令会输出 Top-5 最相关记忆（按优先级排序）。**你必须：**
1. 执行上述命令（不超过 10 秒）
2. 逐条读取 must-follow 级别的记忆文件（用 Read 工具）
3. 在代码中用 \`// SEE: [[memory-slug]]\` 注释标记应用了哪些记忆
4. 如果命令返回空，说明该服务暂无相关记忆，正常继续即可

记忆应用后在代码中标记 \`// SEE: [[memory-slug]]\`，编码结束后输出记忆应用报告（已应用/不适用）。

## 编码纪律（任务类型: ${taskType}）

${strictTdd ? `
### TDD 编码纪律（强制执行 — superpowers:test-driven-development）

**铁律：NO PRODUCTION CODE WITHOUT A FAILING TEST FIRST**
先写实现代码再补测试 → 删掉代码，从测试重新开始。

### RED — 先写失败的测试
1. 对于每个新增/修改的${isFrontend ? '组件/函数' : '函数'}，**先写 ${isFrontend ? 'vitest' : 'table-driven'} tests**（用例：正常路径、空值/零值、错误路径）
2. 运行 \`${testCmd}\` → **必须看到测试失败**（因功能尚未实现）
3. **记录 RED 证据** — 截取 FAIL 输出的关键行
4. 测试直接通过 → 说明测试没测到东西，修正测试重来
5. ${isFrontend ? '用 @vue/test-utils mount 组件，用真实 props 和 store' : '不写 mock（除非依赖外部服务）——用真实结构和业务数据'}

### GREEN — 最小实现
1. 写**刚好够通过测试**的代码，不添加测试未覆盖的功能（YAGNI）
2. 运行测试 → 确认通过 + 全部已有测试不破

### REFACTOR — 清理
1. 消除重复、改善命名、抽取辅助${isFrontend ? '组件/composable' : '函数'}
2. 运行测试 → 保持全绿

### 禁止行为
- ❌ 先写实现再补测试 → **删掉代码，从 RED 开始**
- ❌ "调用不报错"的测试 → 必须 assert 具体${isFrontend ? '值/行为' : '返回值'}
- ❌ "太简单不需要测试" → 简单代码也会出错
` : isDebt ? `
### 简化编码纪律（债务修复 — 允许先修后测）
1. 优先修复问题，然后补回归测试
2. 为修复的代码添加至少一个测试用例验证正确行为
3. 运行 \`${testCmd}\` 确保全部通过
4. 不要求严格 RED→GREEN→REFACTOR 顺序
` : `
### 运维任务（无需 TDD）
直接执行操作，验证结果即可。不需要编写测试代码。
`}

### Memory 引用格式（用于自动追踪）
在代码中使用 \`// SEE: [[memory-slug]]\` 格式标记记忆引用。
Pipeline 会自动解析这些引用并更新记忆的 apply_count。
在编码总结中按以下格式输出记忆应用报告：
\`\`\`
### 记忆应用报告
- 搜索关键词: <关键词>
- 已应用:
  - [[memory-slug]] — <文件名:行号>
\`\`\`

## 硬边界（防止浪费时间）
${isFrontend ? `
- 不修改 web/common/ — 那是共享层，变更需评估影响
- API 接口必须与 api-proto 一致 — 字段名和类型对齐后端 Proto 定义
- 所有 ID 字段使用 \`string\` 类型（Snowflake 精度）
- 其他规范（TypeScript 类型安全/构建/hardcoded secrets）由 QA 机械化检查保证，信任它
` : `
- 不修改 common/ 和 api-proto/ — 那是全局 Owner 的职责
- 服务间通信仅通过 gRPC — 不直连其他服务数据库
- 其他规范（Snowflake ID/json_string/错误码格式）由 QA 机械化检查保证，信任它
`}

## 任务`

  const changelogReq = !isChore ? `\n- 更新 ${SVC_DIR}/CHANGELOG.md` : ''
  const tddReq = strictTdd ? `\n- 每个新增${isFrontend ? '组件/函数' : '函数'}遵循 RED→GREEN→REFACTOR 循环，输出 TDD 证据（测试名 + **RED FAIL 输出摘录** + GREEN PASS 确认）。无 RED 摘录 = TDD 不合格` : ''
  const testReq = !isChore ? `\n- ${testCmd} 通过（含新增测试）` : ''

  if (iteration === 1) {
    return base + `\n${args.task}

## 完成标准
- ${buildCmd} 通过${testReq}${tddReq}${changelogReq}
- 输出记忆应用报告`
  } else {
    const regressReq = !isChore ? `\n- 为修复的问题补写回归测试（先看测试 FAIL 再现修复）` : ''
    return base + `\n修复上一阶段发现的问题。请阅读以下报告中的失败项并逐一修复：

${fixContext}

## 完成标准
- 所有失败项修复完成${regressReq}
- ${buildCmd} 通过
- ${testCmd} 通过${changelogReq}`
  }
}
