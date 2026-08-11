# P1-5 修复报告：任务类型自动推断错误

**修复时间**: 2026-06-23 07:00  
**问题优先级**: P1  
**修复状态**: ✅ 已完成

---

## 问题描述

### 现象

Workflow 将新功能开发任务误判为 `debt`（技术债务），导致：
- Generator 优先修复服务中已存在的问题
- 忽略用户提供的任务描述
- 虽然 QA 通过，但未完成预期功能

### 根本原因

**原代码逻辑**（第 732-739 行）：
```javascript
// bug 检测
if (/\b(bug|fix|修复|...)\b/.test(t)) return 'bug'

// debt 检测（在 feature 之前）
if (/\b(debt|规范|格式|...)\b/.test(t)) return 'debt'

// feature 检测
if (/\b(feature|新增|添加|...)\b/.test(t)) return 'feature'
```

**问题**: `debt` 检测优先于 `feature` 检测

**触发条件**: 当任务描述中同时包含：
- feature 关键词（"创建"、"实现"）
- debt 关键词（"格式"、"重复"）

由于 `debt` 检测在前，会优先匹配，导致误判。

---

## 修复方案

### 修改内容

**文件**: `.harness/workflows/harness-pipeline.js`  
**位置**: 第 732-739 行

**修复后代码**：
```javascript
// P1-5 FIX: 调整优先级顺序 - feature 优先于 debt
// feature: new capability — 优先检测新功能开发
// 包含明确的任务列表（Task 1.1, Task 2.1）或新增关键词
if (/\b(Task \d+\.\d+|feature|新增|添加|实现|创建|新建|开发|add|implement|create|build|支持|功能|页面|组件|接口|api|endpoint|handler|migration|ddl|model|logic)\b/.test(t)) return 'feature'

// bug: broken behavior, crash, data corruption
if (/\b(bug|fix|修复|broken|crash|...)\b/.test(t)) return 'bug'

// debt: standards, style, naming, format, technical debt
if (/\b(debt|规范|格式|format|...)\b/.test(t)) return 'debt'
```

### 修复要点

1. **调整顺序**: 将 `feature` 检测移到 `debt` 之前
2. **增强特征**: 添加 `Task \d+\.\d+` 模式检测（匹配任务列表格式）
3. **扩展关键词**: 添加 `ddl`, `model`, `logic` 等明确的新功能关键词

---

## 修复效果

### 预期改善

| 场景 | 修复前 | 修复后 |
|------|--------|--------|
| 包含"Task 1.1"等任务列表 | debt | **feature** ✅ |
| 包含"创建"、"实现"关键词 | debt | **feature** ✅ |
| 包含"DDL"、"Model"、"Logic" | debt | **feature** ✅ |
| 真正的技术债务任务 | debt | debt ✅ |

### 验证计划

**测试用例**: 重新执行工作记录模块开发（v3）

**预期结果**:
```
任务类型: feature (auto-inferred)  // 而不是 debt
```

**验证检查点**:
- [ ] Workflow 日志显示 `feature` 类型
- [ ] Generator 按照 task 列表实现功能
- [ ] 生成了目标文件（pipeline_test_* 相关）
- [ ] QA 检查通过
- [ ] Review 通过

---

## 对比分析

### v2（修复前）

```javascript
// 顺序：bug → debt → feature
任务类型: debt (auto-inferred)
```

**结果**: 修复了 goctl 重复文件，未实现目标功能

### v3（修复后）预期

```javascript
// 顺序：feature → bug → debt
任务类型: feature (auto-inferred)
```

**预期**: 按照 task 列表实现工作记录模块功能

---

## 其他改进

### 最大轮次限制（基于任务类型）

```javascript
const MAX_ITERATIONS = TASK_TYPE === 'chore' ? 1 
                     : TASK_TYPE === 'debt' ? 2 
                     : 3  // feature/bug
```

**影响**:
- `debt` 任务：最多 2 轮
- `feature` 任务：最多 3 轮

**修复后**: v3 将以 `feature` 类型执行，获得 3 轮迭代机会（vs v2 的 2 轮）

---

## 修复清单

| 文件 | 修改类型 | 行数变化 | 说明 |
|------|---------|---------|------|
| `.harness/workflows/harness-pipeline.js` | 调整顺序 + 增强 | 修改 8 行 | P1-5: feature 优先于 debt |

**总计**: 1 个文件，修改 8 行

---

## 相关问题

### P1-1: 最大轮次限制过严

**现状**: 
- feature 任务：3 轮
- debt 任务：2 轮

**是否需要调整**: 
- 修复 P1-5 后，任务将正确识别为 `feature`，获得 3 轮机会
- 暂时不需要调整，观察 v3 表现

---

## 总结

### 修复状态

| 问题 | 状态 | 验证 |
|------|:----:|:----:|
| **P1-5 任务类型误判** | ✅ 已修复 | ⏳ 待 v3 验证 |

### 预期效果

**修复前**（v2）:
- 任务类型：debt
- 执行内容：修复历史问题
- 功能完成：❌

**修复后**（v3 预期）:
- 任务类型：**feature** ✅
- 执行内容：**实现任务列表** ✅
- 功能完成：**预期 ✅**

### 下一步

1. **立即执行**: 启动 v3 测试，验证 P1-5 修复
2. **观察指标**: 
   - 任务类型是否为 `feature`
   - 是否生成目标文件
   - 是否完成功能交付

---

**修复完成时间**: 2026-06-23 07:05  
**等待验证**: v3 测试执行  
**文档位置**: `.harness/changes/work-records-v2/p1-5-fix-report.md`
