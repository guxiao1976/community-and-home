// ============================================================
// Review Prompts & Schema — Multi-Perspective Code Review
// ============================================================

const REVIEW_SCHEMA = {
  type: 'object',
  properties: {
    verdict: { type: 'string', enum: ['PASS', 'FAIL'] },
    criticalCount: { type: 'number' },
    warningCount: { type: 'number' },
    summary: { type: 'string' },
    criticalIssues: {
      type: 'array',
      items: { type: 'object', properties: { file: { type: 'string' }, issue: { type: 'string' } } },
    },
    memorySuggestions: {
      type: 'array',
      description: '可跨服务复用的模式性问题，建议沉淀为记忆',
      items: {
        type: 'object',
        properties: {
          slug: { type: 'string', description: 'kebab-case slug, 如 json-tag-path-vs-body' },
          title: { type: 'string' },
          category: { type: 'string', enum: ['pitfall', 'guideline', 'process'] },
          severity: { type: 'string', enum: ['must-follow', 'should-follow'] },
          triggers: { type: 'string', description: '空格分隔的触发关键词' },
          content: { type: 'string', description: '记忆正文 (markdown)' },
        },
        required: ['slug', 'title', 'category', 'severity', 'triggers', 'content'],
      },
    },
  },
  required: ['verdict', 'summary'],
}


// ── 多视角审查：三个 Lens，并行执行 ──

const isFrontend = (SVC_DIR || '').startsWith('web/')

const REVIEW_LENSES = [
  {
    key: 'security-arch',
    label: '安全架构',
    dimensions: '架构一致性(#1)、安全性(#5)、变更完整性(#8)',
    focus: isFrontend
      ? '你关注前端架构的正确性和安全风险。检查组件分层合理性、API 调用权限校验、Token 存储安全（localStorage/cookie）、XSS 防护（v-html）、CORS 配置、硬编码密钥、敏感信息泄露到前端、CHANGELOG 完整性。'
      : '你关注架构决策的正确性和安全风险。检查 Proto/gRPC 规范、服务边界、跨服务 DB 访问、硬编码密钥、SQL 注入、输入校验、CHANGELOG 完整性。',
  },
  {
    key: 'standards-eng',
    label: '规范工程',
    dimensions: '规范遵循(#3)、复用性(#6)、测试覆盖(#7)、记忆遵守(#9)',
    focus: isFrontend
      ? '你关注前端编码规范和工程质量。检查 Snowflake ID string 类型、no `as any`（type-safety）、no console.log/debugger、hardcoded secrets、web/common/ 复用（勿重复定义类型）、API 响应直接使用（勿 res.data 双解包）、Vue 模板勿嵌套 {{ }}、测试覆盖(新增组件/函数)、记忆遵守(M1-M4)。'
      : '你关注编码规范和工程质量。检查 Snowflake ID 序列化(jstype/json:\\",string\\")、错误码格式(5位)、API 响应格式、代码复用、测试覆盖(新增函数是否有测试)、记忆遵守(M1-M4)。',
  },
  {
    key: 'design-biz',
    label: '设计业务',
    dimensions: '设计一致性(#2)、代码质量(#4)、Migration(#8部分)',
    focus: isFrontend
      ? '你关注前端业务逻辑的正确性和设计一致性。检查与 design.md 一致性、API 字段名与 api-proto 对齐、组件状态管理合理性、边界条件处理(loading/empty/error 状态)、错误处理完善性(ElMessage 用户提示)、表单验证完整性、性能(大列表虚拟滚动/懒加载)。'
      : '你关注业务逻辑的正确性和设计一致性。检查与 design.md 一致性、数据模型正确性、业务流程正确性、边界条件处理(null/零值/错误路径)、错误处理完善性、资源泄露、Migration 安全性(回滚方案/锁表/影响现有数据)。',
  },
]

function reviewLensPrompt(lens) {
  return `你是 Code Reviewer Agent — ${lens.label}视角。

## 角色定义（必须先读）
阅读 .harness/skills/review.md — 了解完整的 9 维度审查规则。但你只需要聚焦在 ${lens.dimensions} 维度。

## 审查目标
从 **${lens.label}** 视角审查 ${SVC_DIR}/ 的代码变更（QA 已通过，_qa.md 可供参考）。

## 你的审查焦点
${lens.focus}

## 审查步骤
1. 阅读 ${ROOT_CLAUDE} — 全局规则
2. 阅读 ${SVC_DIR}/CLAUDE.md — 服务规则
3. 阅读 ${SVC_DIR}/docs/design.md — 设计文档（如存在）
4. 阅读 ${SVC_DIR}/CHANGELOG.md — 变更历史
5. 阅读 ${SVC_DIR}/_qa.md — QA 报告
6. 获取变更内容（git diff 或审查变更文件）
7. **仅审查你负责的维度**，不越界审查其他视角的维度
8. 写入 ${SVC_DIR}/_review_${lens.key}.md
9. 输出 VERDICT

## 记忆遵守检查（仅 standards-eng 视角需要做）
${lens.key === 'standards-eng' ? `
审查时必须验证代码是否遵守项目记忆系统中的经验：

### M1: 收集代码中的记忆引用
- Grep 变更文件中的 \`// SEE: [[\` 注释，提取所有被引用的 memory-slug

### M2: 验证引用准确性
对每个引用：
- slug 文件不存在 → 🔴 CRITICAL
- 代码未遵守记忆指导 → 🔴 CRITICAL
- 虚假匹配 → 🟡 WARNING

### M3: 检查遗漏的记忆
- 从变更描述和 git diff 提取技术关键词
- 用关键词精确匹配 MEMORY.md 索引中的 triggers
- must-follow 记忆遗漏 → 🔴 CRITICAL，should-follow → 🟡 WARNING
` : '本视角不负责记忆遵守检查（由规范工程视角负责）。'}

## 记忆建议（M4，所有视角都做）

如果审查中发现的问题符合以下条件，在 VERDICT 的 memorySuggestions 字段中建议创建 Memory：
- **模式性**：该问题在其他服务/代码区域也可能存在（非孤立个案）
- **可复用**：不是一次性配置错误或命名拼写错误（有教学价值）
- **未覆盖**：查阅 .harness/knowledge/memory/MEMORY.md 索引，该场景尚未被记忆覆盖

每个 memorySuggestion 需提供完整的记忆内容（包含 slug/title/category/severity/triggers/content）。无符合条件的发现时返回空数组。

## 与其他 Reviewer 的分工

| 视角 | 你（${lens.label}） | 其他 Reviewer |
|------|:---:|:---:|
| 架构一致性 (#1) | ${lens.key === 'security-arch' ? '✅ 审查' : '—'} | ${lens.key !== 'security-arch' ? '✅' : '—'} |
| 设计一致性 (#2) | ${lens.key === 'design-biz' ? '✅ 审查' : '—'} | ${lens.key !== 'design-biz' ? '✅' : '—'} |
| 规范遵循 (#3) | ${lens.key === 'standards-eng' ? '✅ 审查' : '—'} | ${lens.key !== 'standards-eng' ? '✅' : '—'} |
| 代码质量 (#4) | ${lens.key === 'design-biz' ? '✅ 审查' : '—'} | ${lens.key !== 'design-biz' ? '✅' : '—'} |
| 安全性 (#5) | ${lens.key === 'security-arch' ? '✅ 审查' : '—'} | ${lens.key !== 'security-arch' ? '✅' : '—'} |
| 复用性 (#6) | ${lens.key === 'standards-eng' ? '✅ 审查' : '—'} | ${lens.key !== 'standards-eng' ? '✅' : '—'} |
| 测试覆盖 (#7) | ${lens.key === 'standards-eng' ? '✅ 审查' : '—'} | ${lens.key !== 'standards-eng' ? '✅' : '—'} |
| 变更完整性 (#8) | ${lens.key === 'security-arch' || lens.key === 'design-biz' ? '✅ 审查' : '—'} | ${lens.key !== 'security-arch' && lens.key !== 'design-biz' ? '✅' : '—'} |
| 记忆遵守 (#9) | ${lens.key === 'standards-eng' ? '✅ 审查' : '—'} | ${lens.key !== 'standards-eng' ? '✅' : '—'} |

## 审查报告格式

写入 ${SVC_DIR}/_review_${lens.key}.md：

\`\`\`markdown
# Code Review — ${args.serviceName}（${lens.label}视角）

**审查时间**: <时间>
**审查维度**: ${lens.dimensions}

## 摘要
- 🔴 CRITICAL: N / 🟡 WARNING: N / 🔵 NOTE: N

## 发现

### 🔴 CRITICAL
| # | 文件:行号 | 维度 | 问题 | 修复建议 |
|---|----------|------|------|---------|

### 🟡 WARNING
| # | 文件:行号 | 维度 | 问题 | 建议 |
|---|----------|------|------|------|

### 🔵 NOTE
| # | 文件:行号 | 建议 |
|---|----------|------|

---
VERDICT: PASS / FAIL
---
\`\`\`

## VERDICT 规则
- PASS — 你负责的维度中无 CRITICAL 问题
- FAIL — 存在 ≥1 个 CRITICAL（列出具体问题和修复建议）

## 约束
- 只读权限：Read、Grep、Glob、Bash
- 严禁 Write、Edit
- 只审查你负责的维度，不要越界
- 对于不负责的维度，如果偶然注意到问题，可以写在「🔵 NOTE」中供其他 Reviewer 参考
`
}
