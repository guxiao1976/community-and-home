# P0 修复验证测试用例

## 测试目标

验证 P0-1 和 P0-2 修复是否生效。

---

## 测试用例 1: 简单 CRUD 端点（验证 P0-1）

**需求**: 在 master-data-service 添加一个测试端点 `/api/masterdata/test/hello`

**预期**:
1. Workflow 应该收到 `args.task` 参数
2. Generator Agent 应该按照任务描述生成代码
3. 不应该去修复服务中已存在的历史问题

**任务描述**:
```
在 master-data-service 添加一个简单的 Hello World 测试端点。

API 规格：
- 路径: GET /api/masterdata/test/hello
- 请求参数: name (string, query parameter, 可选)
- 响应: {"message": "Hello, {name}!"}
- 如果 name 为空，返回 "Hello, World!"

实现要求：
- 创建 Handler: api/internal/handler/test/hello_handler.go
- 注册路由到 api/internal/handler/routes.go
- 添加单元测试: api/internal/handler/test/hello_handler_test.go
- 更新 CHANGELOG.md
```

**验证步骤**:
1. 启动 Workflow
2. 检查日志中是否包含 `[DEBUG] Workflow 启动 - args 参数检查`
3. 确认 `task 长度` 和 `task 前100字符` 正确输出
4. 确认 Generator 生成了目标文件（hello_handler.go）

---

## 测试用例 2: 熔断机制（验证 P0-2）

**需求**: 故意触发工具调用失败，验证熔断机制

**方法**: 
- 无需实际执行，P0-2 修复已添加到子 Agent 系统提示
- 下次子 Agent 遇到连续失败时，应该自动诊断并切换方法

**验证点**:
- requirement-analyst.md 包含熔断规则 ✅
- architecture-designer.md 包含熔断规则 ✅

---

## 执行计划

### 阶段 1: 验证 P0-1 修复

```bash
# 1. 创建测试变更目录
mkdir -p .harness/changes/test-p0-fix-hello-endpoint

# 2. 创建 request.md
cat > .harness/changes/test-p0-fix-hello-endpoint/request.md << 'EOF'
# Request: 测试 P0-1 修复 - Hello 端点

**用户原话**: 在 master-data-service 添加一个简单的 Hello World 测试端点。

**路径**: Dev Agent（单服务，简单任务）

**服务**: master-data-service

**创建时间**: 2026-06-23 06:00
EOF

# 3. 启动 Workflow（带调试日志）
# 注意：需要手动执行 Workflow 调用
```

**Workflow 调用**:
```javascript
Workflow({
  scriptPath: ".harness/workflows/harness-pipeline.js",
  args: {
    serviceName: "主数据服务",
    serviceDir: "services/master-data-service",
    task: "在 master-data-service 添加一个简单的 Hello World 测试端点。\n\nAPI 规格：\n- 路径: GET /api/masterdata/test/hello\n- 请求参数: name (string, query parameter, 可选)\n- 响应: {\"message\": \"Hello, {name}!\"}\n- 如果 name 为空，返回 \"Hello, World!\"\n\n实现要求：\n- 创建 Handler: api/internal/handler/test/hello_handler.go\n- 注册路由到 api/internal/handler/routes.go\n- 添加单元测试: api/internal/handler/test/hello_handler_test.go\n- 更新 CHANGELOG.md"
  }
})
```

### 阶段 2: 检查日志

**检查点**:
1. Workflow 输出日志中包含 `[DEBUG] Workflow 启动 - args 参数检查`
2. 日志显示 `task 长度: 300+` 和 `task 前100字符: 在 master-data-service 添加...`
3. Generator Agent 读取了 `args.task`
4. 生成的代码文件路径正确

---

## 预期结果

### P0-1 修复成功标志

- ✅ Workflow 日志显示完整的 args 参数
- ✅ Generator 按照 task 描述生成代码
- ✅ 未出现"修复历史问题"的行为

### P0-2 修复成功标志

- ✅ 子 Agent 系统提示包含熔断规则
- ✅ 下次遇到连续失败时，会输出诊断并切换方法

---

## 备注

**为什么选择 Hello 端点测试**:
1. 简单：无复杂业务逻辑，易于验证
2. 独立：不影响现有功能
3. 快速：预计 5-10 分钟完成
4. 清晰：任务描述明确，易于判断是否按要求执行

**如果测试失败**:
1. 检查 Workflow 日志，确认 args 是否传递
2. 检查 Generator Agent 的 JSONL 日志
3. 如果 args 仍然未传递，可能需要检查 Workflow 运行时的 args 注入机制
