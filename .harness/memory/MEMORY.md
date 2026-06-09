# 全局经验索引

> Agent 启动时读取本文件，根据当前任务上下文精读相关记忆文件。
> 格式：`- [标题](文件.md) — 适用范围, 严重程度, 类型, 触发关键词`
>
> 类型说明：`pitfall`=踩过的坑 | `guideline`=编码/架构规范 | `process`=流程约束 | `decision`=技术决策 | `model`=数据模型

## 必须遵守 (must-follow)

- [Proto int64 字段必须加 jstype=JS_STRING](proto-jstype.md) — api-proto, must-follow, `guideline`, `proto int64 jstype JS_STRING Snowflake`
- [服务间通信仅通过 gRPC](grpc-only-comms.md) — all, must-follow, `guideline`, `gRPC 服务间调用 直连数据库`
- [API 调用前验证路由存在 + 禁止静默吞错](verify-api-before-calling.md) — all, must-follow, `pitfall`, `API 路由 404 catch 空 静默`
- [提交前必须通过机械化检查](pre-commit-checks.md) — all, must-follow, `process`, `提交 commit 门禁 gate harness-checks QA 检查`
- [手机号加密存储读取必须解密](phone-encryption.md) — all, must-follow, `pitfall`, `phone 手机号 AES 加密 解密 乱码`
- [Migration 文件提交后必须执行](migration-must-execute.md) — all, must-follow, `pitfall`, `migration DDL ALTER TABLE column schema 1054 unknown column`

## 应该遵守 (should-follow)

- [前端可视化开发流程](frontend-visual-development-workflow.md) — web/mobile, should-follow, `process`, `UI 页面 设计 样式 视觉`

## 参考信息 (info)

