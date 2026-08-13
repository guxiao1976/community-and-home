# 问题 1 改进完成报告

## ✅ 改进目标

**原问题**：Pipeline 模板是巨型单体（976 行），Prompts 嵌在 JS 字符串里，没有语法高亮。

## 🎯 解决方案

### 架构升级：三层分离

```
旧架构（字符串嵌入）：
  prompts/*.js (559 行)
  └─ const prompt = `巨型Markdown字符串` ← 无语法高亮，难维护

新架构（模板分离）：
  templates/*.md (502 行)           ← ✅ Markdown 格式，有语法高亮
  ├─ generator.md
  ├─ qa.md
  ├─ review.md
  └─ debug.md
  
  template-renderer.js (52 行)      ← Mustache-like 模板引擎
  
  prompts/*-new.js (约 150 行)      ← 只负责变量构造和模板调用
  
  ↓ build-pipeline-new.sh
  
  harness-pipeline.js (1186 行)     ← 自动生成的完整工作流
```

## 📊 改进效果

| 指标 | 旧系统 | 新系统 | 改进 |
|------|--------|--------|------|
| **Prompt 可编辑性** | JS 字符串，无语法高亮 | Markdown 文件，完整高亮 | ✅ **质的提升** |
| **Prompt 总行数** | 559 行（混在 JS 里） | 502 行（纯 Markdown） | -10% |
| **模板渲染逻辑** | 硬编码字符串拼接 | 52 行模板引擎 | ✅ 可复用 |
| **构建产物** | 976 行 | 1186 行 | +21%（但源码更清晰） |
| **编辑流程** | 改 JS → 重新构建 | 改 MD → 重新构建 | ✅ 同样简单 |

## 🔧 技术实现

### 1. Markdown 模板（支持变量和条件）

```markdown
# Generator Agent Prompt

你是 {{serviceName}} 的{{#isFrontend}}前端{{/isFrontend}}开发 Agent。

{{#strictTdd}}
**Feature/Bug 类型**：严格 TDD
- **RED FAIL 输出摘录**（必须）
{{/strictTdd}}

{{^strictTdd}}
**Chore/Debt 类型**：简化 TDD
{{/strictTdd}}
```

### 2. 模板渲染器（Mustache-like）

```javascript
// 支持：
// {{var}}                 - 变量替换
// {{#condition}}...{{/}}  - 条件渲染（truthy）
// {{^condition}}...{{/}}  - 条件渲染（falsy）

render(template, {
  serviceName: "用户服务",
  isFrontend: false,
  strictTdd: true
})
```

### 3. 构建脚本（自动内联模板）

```bash
bash .harness/scripts/build-pipeline-new.sh

# 输出：
# ✅ harness-pipeline.js built successfully
#    📊 Total: 1186 lines
#    📄 Templates: 502 lines (editable Markdown)
#    🎯 Logic: 684 lines (auto-generated)
```

## 💡 使用方式

### 开发者编辑 Prompt

```bash
# 1. 编辑 Markdown 模板（有语法高亮）
code .harness/agents/prompts/templates/generator.md

# 2. 重新构建
bash .harness/scripts/build-pipeline.sh

# 3. 完成！新的 Prompt 已生效
```

### 对比旧方式

```bash
# 旧方式：编辑 JS 字符串（无语法高亮）
code .harness/agents/prompts/generator.js
# 找到第 16 行的反引号字符串
# 在几百行的字符串里找到要改的地方
# 改完，重新构建

# 新方式：直接编辑 Markdown
code .harness/agents/prompts/templates/generator.md
# Markdown 格式清晰，语法高亮
# 改完，重新构建
```

## 🎉 关键改进点

### ✅ 1. 语法高亮
- **旧**：JS 反引号字符串，编辑器无法识别 Markdown
- **新**：`.md` 文件，完整的 Markdown 语法高亮

### ✅ 2. 结构清晰
- **旧**：Prompt 和 JS 逻辑混在一起
- **新**：Prompt（.md）和逻辑（.js）完全分离

### ✅ 3. 模板复用
- **旧**：每个 Prompt 都是独立的巨型字符串
- **新**：模板引擎可复用，支持条件渲染

### ✅ 4. 易于维护
- **旧**：改一个 Prompt 要在 200 行字符串里找位置
- **新**：Markdown 文件有标题、章节，快速定位

## 📁 新增文件

```
.harness/agents/prompts/
├── templates/              ← 新增：Markdown 模板目录
│   ├── generator.md       (172 行)
│   ├── qa.md              (87 行)
│   ├── review.md          (126 行)
│   └── debug.md           (117 行)
├── template-renderer.js   (52 行) ← 新增：模板引擎
├── generator-new.js       (38 行) ← 新增：模板加载器
├── qa-new.js              (35 行)
├── review-new.js          (42 行)
└── debug-new.js           (33 行)

.harness/scripts/
└── build-pipeline-new.sh  ← 更新：支持模板系统
```

## 🚀 下一步

### 选项 A：直接替换（推荐）
```bash
# 备份旧文件
mv .harness/agents/prompts/generator.js .harness/agents/prompts/generator.js.bak
mv .harness/scripts/build-pipeline.sh .harness/scripts/build-pipeline.sh.bak

# 启用新文件
mv .harness/agents/prompts/generator-new.js .harness/agents/prompts/generator.js
# ... 重复其他 3 个文件
mv .harness/scripts/build-pipeline-new.sh .harness/scripts/build-pipeline.sh

# 重新构建测试
bash .harness/scripts/build-pipeline.sh
```

### 选项 B：保留旧文件，并行运行
```bash
# 保持旧系统不变
# 新系统通过 build-pipeline-new.sh 独立构建
# 团队逐步迁移
```

## ✅ 验证结果

```bash
✅ 模板渲染器测试通过
✅ 新构建脚本成功生成 harness-pipeline.js
✅ 语法检查通过（node --check）
✅ 行数统计正确（1186 行 vs 旧的 976 行）
```

---

**结论**：问题 1 已彻底解决。Prompts 不再是 JS 字符串，而是独立的 Markdown 文件，有完整的语法高亮和结构化编辑体验。
