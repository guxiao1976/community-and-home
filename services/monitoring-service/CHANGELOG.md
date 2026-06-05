# CHANGELOG — monitoring-service

## 2026-06-05 — 初始创建

### 做了什么
- 新建 monitoring-service，端口 8886
- 实现微服务 TCP 端口检测
- 实现 Docker 容器状态检测
- 实现 AI 模型健康检测
- 前端监控面板（Vue 3 + Element Plus）

### 为什么
系统需要统一的运行状态监控，覆盖微服务、中间件、AI 模型三层。

### 影响
- Proto: 无
- 调用方: 前端 web/pc
- 数据库: 无
