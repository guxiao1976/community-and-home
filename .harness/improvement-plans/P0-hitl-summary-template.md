# Summary 模板更新

为了支持 HITL 置信度自适应审查，`summary.md` 模板需要新增以下字段：

## 新增章节：阶段 5 验收

```markdown
## 阶段 5 验收

- **Confidence**: 0.85
- **审查模式**: 摘要审查 | 抽查 | 全文审查
- **审查结果**: PASS | FAIL
- **审查人**: Owner Agent (自动) | Human (手动)
- **审查时间**: 2026-06-18 15:30
- **审查耗时**: 5 分钟 | 15 分钟 | 30 分钟

### 审查详情
- QA 验收: PASS (15/15 检查通过)
- Review 状态: 3/3 视角 PASS
- CRITICAL 问题: 0 个

### 抽查记录（如适用）
- 抽查文件: [file1.go, file2.go]
- 抽查发现: 
  - ✅ file1.go: Memory 引用正确，测试覆盖完整
  - ✅ file2.go: 无问题
```

## 使用场景

### 高置信（≥0.80）— 摘要审查
```markdown
## 阶段 5 验收
- Confidence: 0.85
- 审查模式: 摘要审查
- 审查结果: PASS
- 审查人: Owner Agent (自动)
- 审查耗时: 5 分钟
```

### 中置信（0.50-0.79）— 抽查
```markdown
## 阶段 5 验收
- Confidence: 0.65
- 审查模式: 抽查（30% 文件）
- 审查结果: PASS
- 审查人: Owner Agent (自动)
- 审查耗时: 15 分钟

### 抽查记录
- 抽查文件: 3/10 个文件
  - services/xxx/logic.go
  - services/xxx/handler.go
  - services/xxx/types.go
- 抽查发现: 无 CRITICAL 问题
```

### 低置信（<0.50）— 全文审查
```markdown
## 阶段 5 验收
- Confidence: 0.45
- 审查模式: 全文审查 + 人工确认
- 审查结果: PASS（人工批准）
- 审查人: Human
- 审查耗时: 30 分钟

### 发现问题
- ⚠️ 缺少边界条件测试（已接受风险）
- ⚠️ 错误处理不完整（已接受风险）
```

## 集成位置

在 `.harness/changes/TEMPLATE.md` 的「阶段 6: 集成归档」之前插入此章节。

## 下一步

Phase 2 将实现 Owner Agent 的分级审查逻辑，自动填充这些字段。
