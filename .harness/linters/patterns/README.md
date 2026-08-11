# Harness Linter Patterns — 正确模式索引

> **工具即反馈（Tools as Feedback）**  
> 这些模式文档配合 `harness-checks.sh` 的错误消息，为 Agent 提供即时的修复指令。

## 核心理念

当 Agent 遇到架构违规时，Linter 错误消息应该包含：

1. **WHY** — 为什么违规（架构原理）
2. **FIX** — 如何修复（具体步骤）
3. **EXAMPLE** — 项目中的正确实现（文件路径:行号）
4. **REFERENCE** — 相关文档（架构规范、Memory）

这创造了一个**零人工干预的反馈回路**：

```
Agent 写代码 → Linter 检测违规 → 
错误消息 = 修复指令（注入 Agent 上下文）→ 
Agent 读取对应的 pattern 文档 → 
Agent 按文档修复 → 
Linter 再次运行 → 通过 ✓
```

## 模式文档列表

| 模式文档 | 对应检查 | 适用场景 |
|---------|---------|---------|
| [proto-jstype.md](./proto-jstype.md) | Check 4: Proto int64 jstype | Proto 定义中的 ID/时间戳字段 |
| [json-string.md](./json-string.md) | Check 5: Go json:",string" | Go REST API types.go 中的 ID 字段 |
| [cross-service-rpc.md](./cross-service-rpc.md) | Check 6: Cross-service DB import | 跨服务调用（禁止直接 DB 访问）|
| [response-wrap.md](./response-wrap.md) | Check 13: Response single-wrap | REST API 响应格式（禁止双层嵌套）|

## 使用方式

### 1. Agent 遇到检查失败

```bash
[FAIL] 6. cross-service DB import — 3 violations: 
  services/auth-service/api/internal/logic/login_logic.go imports user-service/model

WHY: 服务间通信必须通过 gRPC。直接访问其他服务的数据库破坏服务边界，造成紧耦合。

FIX: 
  1. 移除跨服务的 model 包导入
  2. 在 svcCtx 中添加对应的 RPC 客户端（如 UserRpc）
  3. 通过 RPC 调用获取数据：svcCtx.UserRpc.GetUserInfo(ctx, req)
  4. 将 RPC 响应映射到 Logic 返回类型

EXAMPLE: services/auth-service/api/internal/logic/verify_token_logic.go:28-35

REFERENCE: .harness/rules/项目编码规范.md §1 | .harness/linters/patterns/cross-service-rpc.md
```

### 2. Agent 读取对应的模式文档

```
Agent 看到 REFERENCE 字段中的 cross-service-rpc.md
  ↓
读取 .harness/linters/patterns/cross-service-rpc.md
  ↓
看到完整的：
  - ❌ 错误模式（带解释）
  - ✅ 正确模式（带代码示例）
  - 参考实现（项目中已有的正确代码）
  - 步骤清单
```

### 3. Agent 按照文档修复

```go
// Agent 学到：
// 1. 移除 import "github.com/guxiao1976/community-user/model"
// 2. 在 svcCtx 中添加 UserRpc userclient.User
// 3. 在 config 中添加 UserRpc zrpc.RpcClientConf
// 4. 在 YAML 中添加 UserRpc.Etcd 配置
// 5. 在 Logic 中调用 l.svcCtx.UserRpc.GetUserInfo(...)
```

### 4. Agent 重新运行检查

```bash
bash .harness/skills/qa/scripts/harness-checks.sh --service auth-service

[PASS] 6. cross-service DB import — no violations
```

## 模式文档编写规范

每个模式文档应包含以下章节：

### 必需章节

1. **核心原则** — 一句话总结规则
2. **为什么** — 架构原理、技术原因、反例后果
3. **❌ 错误模式** — 违规代码示例 + 问题说明
4. **✅ 正确模式** — 正确代码示例 + 步骤分解
5. **完整示例** — 项目中已有的正确实现（文件路径:行号）
6. **检查清单** — 修复时的逐项确认
7. **相关文档** — 链接到规范、Memory、其他模式

### 可选章节

- **常见问题** — FAQ，覆盖常见误区
- **前端配套** — 如果涉及全链路，提供前端对应的实现
- **对比表** — 错误 vs. 正确的对比
- **原理深入** — 技术细节、历史背景

### 示例格式

```markdown
# 正确模式：XXX

## 核心原则

**一句话总结规则。**

## 为什么

### 子标题（如果需要）

解释架构原理...

## ❌ 错误模式

```go
// 错误代码示例
```

**问题**：列出具体问题

## ✅ 正确模式

### 步骤 1：XXX

```go
// 正确代码示例
```

### 步骤 2：XXX

...

## 完整示例

**文件**: `services/xxx/yyy.go:28-35`

## 检查清单

- [ ] 步骤 1
- [ ] 步骤 2
...

## 相关文档

- [规范名称](路径)
```

## 待补充的模式文档

以下模式尚未编写文档，按优先级排序：

| 优先级 | 模式名称 | 对应检查 | 说明 |
|:-----:|---------|---------|------|
| P0 | hardcoded-secrets.md | Check 8 | 硬编码密钥检测 + .env 使用规范 |
| P1 | api-stubs.md | Check 12 | goctl TODO 桩处理 |
| P1 | error-codes.md | Check 7 | 错误码规范（使用 errx 常量）|
| P2 | test-coverage.md | Check 3 | 测试覆盖率 + TDD 流程 |
| P2 | proto-breaking.md | Proto CI | 破坏性变更检测 + 兼容性原则 |

## 贡献指南

### 新增模式文档

1. 在 `.harness/linters/patterns/` 创建新的 `.md` 文件
2. 按照上述编写规范编写内容
3. 在本 README 的"模式文档列表"表格中添加索引
4. 在 `harness-checks.sh` 对应检查的 `log_fail` 中添加 REFERENCE 字段

### 更新现有文档

1. 每次遇到新的违规模式或修复方式，补充到对应文档
2. 在"完整示例"章节引用项目中新的正确实现
3. 在"常见问题"章节添加新的 FAQ

### 质量标准

- **可操作性** — Agent 读完后知道具体怎么做
- **自包含** — 无需跳转太多链接，核心信息在本文档内
- **示例真实** — 引用的文件路径、行号必须真实存在且正确
- **持续更新** — 每次项目演进后，更新示例和路径

## 相关资源

- [项目编码规范](../../rules/项目编码规范.md) — 全局硬性约束
- [Proto 管理规范](../../rules/Proto管理规范.md) — Proto 变更流程
- [工程结构规范](../../rules/工程结构.md) — 分层架构约束
- [harness-checks.sh](../../skills/qa/scripts/harness-checks.sh) — 机械化检查脚本
- [OpenAI Constraints Analysis](../../docs/openai-constraints-analysis.md) — 设计理念深度分析

## 设计理念

这套模式索引系统的设计灵感来自 **OpenAI "工具即反馈"（Tools as Feedback）** 理念：

> "传统的 Linter 是为人类设计的，假设读者有经验、有上下文、能推断修复方式。为 Agent 设计的 Linter 需要把修复步骤说清楚，就像写给一个聪明但缺乏代码库特定知识的新人看的一样。"

**核心洞见**：

1. **Linter 错误消息 = Agent 的临时教练**  
   不只是告诉你"错了"，而是告诉你"为什么错"、"怎么改"、"参考哪里"

2. **模式文档 = Agent 的知识库**  
   每个违规对应一份详细的修复指南，Agent 无需猜测或搜索

3. **零人工干预的反馈回路**  
   Agent → Linter → Pattern Doc → Agent 修复 → Linter 验证 → 通过

详见：[OpenAI Constraints Analysis](../../docs/openai-constraints-analysis.md)
