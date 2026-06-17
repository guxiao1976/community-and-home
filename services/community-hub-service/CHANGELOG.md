# CHANGELOG — community-hub-service

## 2026-06-06 — 服务初始化

### 做了什么
- 创建 community-hub-service 微服务，实现社区枢纽功能
- 实现 RPC + REST API 双层架构（端口 8087/8887）
- 实现 NoticeService（通知公告 CRUD + 软删除）
- 实现 ContactService（便民联络列表 + 批量更新）
- 实现 LostFoundService（寻失互助 CRUD + 标记解决）
- 创建 4 张数据库表（notices, notice_attachments, community_contacts, lost_found_items）
- 使用 go-zero sqlx 风格数据模型（4 个 Model）
- 所有 int64 ID 使用 json:",string" 标签
- 使用 configx.MustLoad 加载配置（支持 ${ENV_VAR}）
- 使用 Snowflake 生成分布式唯一 ID

### 为什么
社区平台需要小区信息汇聚中心，提供通知公告、便民联络、寻失互助等社区内容场景

### 影响
- Proto: api-proto/api/community/v1/community.proto（已定义）
- 调用方: 无（新服务，暂无外部 gRPC 调用方）
- 数据库: 新增 community_hub_db 库，4 张表
- 关联: go.work 已添加 services/community-hub-service
