---
triggers: ["TODO stub", "空壳", "goctl", "logic stub", "harness check", "API logic", "未实现", "todo: add your logic", "silent success", "假成功"]
service: all
severity: must-follow
type: pitfall
status: active
created: 2026-06-12
updated: 2026-06-12
---

# goctl 生成的 Logic 空壳必须实现后才能交付

## 为什么会有这条经验

AI Model Service 的 12 个 API Logic 文件全是 goctl 生成的 TODO 空壳（`// todo: add your logic here and delete this line`）。Handler 调用这些空壳 Logic，Logic 直接 `return (nil, nil)`，Handler 调用 `response.Success(w, nil)` 返回 `{code:0, data:null}`——前端看到 `code:0` 以为操作成功，但实际上什么都没发生。

具体表现：
- 删除模型 → 弹出"删除成功"，列表里数据还在（因为根本没调 RPC 层删数据）
- 创建模型 → 返回成功但无 ID，实际数据未入库
- 模板 CRUD → 全部假成功

**根因**：goctl 生成的 Logic stub 不应直接交付。每个 stub 必须在提交前实现真正的业务逻辑（至少调用 RPC 层）。

## 怎么做

1. goctl 生成 Logic 文件后，**立即**实现 `todo` 标记的函数体
2. 实现完成后**删除** `// todo: add your logic here and delete this line` 注释
3. 提交前运行 `harness-checks.sh`：Check 12 会扫描所有未实现的 TODO stub 并标记 FAIL

## 怎么验证

1. Harness Check 12 (`check_api_stubs`)：扫描 `api/internal/logic/` 下的 `todo: add your logic` 残留
2. 手工测试：对每个 CRUD 端点发送 curl 请求，验证数据确实入库/更新/删除
3. 集成测试：调用 create → list 验证出现 → delete → list 验证消失

## 关联经验

- [[api-response-single-wrap]]
- [[pre-commit-checks]]
