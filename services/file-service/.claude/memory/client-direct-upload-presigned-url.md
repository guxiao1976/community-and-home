---
triggers: ['MinIO', '预签名', '上传', '直传', 'presigned', 'file', '文件']
type: guideline
severity: must-follow
service: file-service
status: active
created: 2026-08-09
updated: 2026-08-09
apply_count: 0
---

# 客户端直传模式 — 预签名 URL，文件流不经过本服务

上传使用预签名 URL，客户端直传 MinIO，文件流不经过本服务。本服务只生成预签名 URL 和确认上传完成。不要在服务端接收文件流。

## 参考

详见 `services/file-service/CLAUDE.md` 和 `services/file-service/docs/design.md`。
