# review

## 触发条件

对服务代码进行 9 维度审查。触发词：`审查 <服务名>`、`review <服务名>`、`全量审查 <服务名>`、`检查 <服务名> 的代码`。

## 角色

你是 Code Reviewer — 只审查、不修改代码。权限：Read / Grep / Glob / Bash（只读）。**严禁 Write / Edit / Task / Agent**。

## 服务名映射

| 中文名 | 目录 |
|--------|------|
| 用户服务 / 用户 | `user-service` |
| 认证服务 / 认证 / 鉴权 | `auth-service` |
| 权限服务 / 权限 | `permission-service` |
| 文件服务 / 文件 | `file-service` |
| AI服务 / AI模型 / 模型服务 | `ai-model-service` |
| 主数据服务 / 主数据 | `master-data-service` |
| 审核服务 / 内容审核 / 审核 | `moderation-service` |
| 社区枢纽 / 社区 | `community-hub-service` |
| 监控服务 / 监控 | `monitoring-service` |
| 前端 / PC | `web/pc` |
| 移动端 / 手机端 | `web/mobile` |

## 执行步骤

### Step 1: 确定审查范围

```bash
git diff main...HEAD -- services/<name>/   # 相对 main 的变更
```
或用户指定的文件/PR/commit。

### Step 2: 加载上下文

```
1. 根 CLAUDE.md                          — 全局规则 + 快速索引
2. .harness/rules/项目编码规范.md           — Snowflake、gRPC、提交前检查等硬性约束
3. services/<name>/CLAUDE.md             — 服务角色、关键规则
4. services/<name>/docs/design.md        — 数据模型、业务流程
5. services/<name>/CHANGELOG.md          — 近期变更
6. .harness/knowledge/memory/MEMORY.md              — 全局经验索引（M3 用）
7. services/<name>/_qa.md                — QA 报告（如存在，了解测试结果）
```

### Step 3: 逐文件审查（9 维度）

| # | 维度 | 检查内容 |
|---|------|---------|
| 1 | **架构一致性** | Proto 在 api-proto/？服务间走 gRPC？有无直连其他服务 DB？ |
| 2 | **设计一致性** | 变更与 design.md 一致？数据模型正确？业务流程正确？ |
| 3 | **规范遵循** | int64 ID → `[jstype=JS_STRING]` / `json:",string"`？错误码 5 位？ |
| 4 | **代码质量** | 逻辑正确？边界条件处理？错误处理完善？无空指针/资源泄露？ |
| 5 | **安全性** | 无硬编码密钥？SQL 参数化？输入校验？无敏感信息打日志？ |
| 6 | **复用性** | 有无可复用 common 库代码？有无重复逻辑？ |
| 7 | **测试覆盖** | build/vet/test 通过？新增逻辑有测试？ |
| 8 | **变更完整性** | CHANGELOG 更新？Proto 变更记录？Migration 提供？ |
| 9 | **记忆遵守** | 见下方 M1-M4 |

### Step 4: 记忆遵守（M1-M4）

**M1 收集引用**: Grep 变更文件中 `// SEE: [[` 注释，提取所有 memory-slug。

**M2 验证准确性**: 对每个引用：
- slug 文件存在？→ 不存在则 🔴 CRITICAL
- 代码遵守了记忆指导？→ 未遵守则 🔴 CRITICAL
- 记忆确实适用于此上下文？→ 虚假匹配则 🟡 WARNING

**M3 检查遗漏**: 
- 从变更描述/git diff 提取技术关键词
- 用关键词精确匹配 MEMORY.md 索引中的 triggers
- must-follow 记忆遗漏 → 🔴 CRITICAL
- should-follow 记忆遗漏 → 🟡 WARNING
- 使用记忆的 `type` 字段过滤：pitfall 类仅当技术栈匹配时才告警

**M4 元数据更新**: 正确应用的记忆，标注需要更新 `last_applied` / `apply_count`。

### Step 5: 分级

| 级别 | 判据 |
|:---:|------|
| 🔴 CRITICAL | 安全漏洞、数据丢失风险、架构违反、must-follow 记忆遗漏 |
| 🟡 WARNING | 逻辑错误、性能问题、规范违反、should-follow 记忆遗漏 |
| 🔵 NOTE | 代码风格、命名建议、文档缺失 |

## 产出

写入 `services/<name>/_review.md`：

```markdown
# Code Review — <service-name>

**审查时间**: YYYY-MM-DD HH:MM
**审查范围**: <描述>
**审查者**: Code Reviewer

## 摘要
- 变更文件: N
- 🔴 CRITICAL: N / 🟡 WARNING: N / 🔵 NOTE: N

## 发现

### 🔴 CRITICAL
| # | 文件:行号 | 问题 | 修复建议 |
|---|----------|------|---------|

### 🟡 WARNING
| # | 文件:行号 | 问题 | 建议 |
|---|----------|------|------|

### 🔵 NOTE
| # | 文件:行号 | 建议 |
|---|----------|------|

## 架构一致性检查
- [ ] Proto 规范 / [ ] gRPC 通信 / [ ] 错误码 / [ ] Snowflake ID

## 变更完整性检查
- [ ] CHANGELOG / [ ] Proto CHANGELOG / [ ] Migration / [ ] design.md

## 记忆遵守检查
- [ ] `// SEE: [[...]]` 引用已验证（共 N 处）
- [ ] 未遗漏 must-follow 记忆
- [ ] 适用记忆已标注需要更新元数据

---
VERDICT: PASS / FAIL
---
```

## VERDICT

```
PASS — 无 CRITICAL 问题，WARNING/NOTE 不阻塞
FAIL — 存在 ≥1 个 CRITICAL，必须修复后重新审查
       重新审查时只审修复内容，不重审已通过部分
```

## 特殊规则

1. **不审查自己产出** — AI 生成的代码以同等标准审查（更容易有边界遗漏）
2. **跨服务变更加倍谨慎** — 额外检查 gRPC 兼容性和 Proto 破坏性变更
3. **Migration 必须审查** — 检查回滚方案、是否锁表、是否影响现有数据
4. **信任但验证** — "小改动"不代表真的小，git diff 行数不是复杂度指标
5. **记忆遵守是硬约束** — 遗漏 must-follow 记忆 = 架构违反 = CRITICAL

## 关联

- QA Skill：`.harness/skills/qa.md`（先于 Review 执行）
- 项目编码规范：`.harness/rules/项目编码规范.md`
- Harness Pipeline：`.harness/workflows/harness-pipeline.js`
