# 全局经验索引

> Agent 启动时读取本文件，根据当前任务上下文精读相关记忆文件。
> 格式：`- [标题](文件.md) — 适用范围, 严重程度, 触发关键词`

## 必须遵守 (must-follow)

- [Proto int64 字段必须加 jstype=JS_STRING](proto-jstype.md) — api-proto, must-follow, `proto int64 jstype JS_STRING Snowflake`
- [服务间通信仅通过 gRPC](grpc-only-comms.md) — all, must-follow, `gRPC 服务间调用 直连数据库`


## 应该遵守 (should-follow)


## 参考信息 (info)

- [前端可视化开发流程](frontend-visual-development-workflow.md) — web/mobile, should-follow, `UI 页面 设计 样式 视觉`
