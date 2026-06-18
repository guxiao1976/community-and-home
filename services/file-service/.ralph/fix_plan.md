# Ralph Fix Plan — file-service

## Immediate
- [x] 验证 MinIO 上传/下载功能
- [x] 检查文件访问权限控制

## Soon
- [x] Snowflake ID 序列化检查 — types.go 9 个 int64 JSON 字段修复 `,string` 标签
- [x] 错误码审查（确保使用 07xxxx 前缀）— 全部通过，70003 语义漂移需更新 proto 注释

## Completed
- [x] Ralph 初始化
- [x] 验证 MinIO 上传/下载功能 — 详见下方验证报告
