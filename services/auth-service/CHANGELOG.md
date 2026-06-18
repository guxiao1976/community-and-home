# CHANGELOG — auth-service

## 2026-06-04 — C8: SMS 验证码认证绕过修复

### 做了什么
- 审计确认 `registerlogic.go` 和 `loginsmslogic.go` 中的 SMS 验证码校验逻辑完整：
  1. 从 Redis 读取已存储的验证码（key: `sms:code:{phone}`）
  2. 与用户输入的验证码比对，不匹配返回错误码 50004
  3. 比对成功后删除 Redis 中的验证码（防重放）
  4. 比对失败不删除验证码，保留给重试或注册流程使用
- 对应测试用例 `TestRegister_CodeMismatch`、`TestRegister_CodeExpired`、`TestLoginSms_CodeMismatch`、`TestLoginSms_CodeExpired` 全部通过
- `go build ./...` 和 `go test ./...` 全部通过

### 为什么
架构审查（C8）发现 SMS 验证码存在认证绕过风险。经审计确认当前代码已按设计文档 §3.2/§3.3 实现验证码比对逻辑。

### 关键设计点
- **registerlogic.go**：验证码比对通过后立即删除（§3.3，防止重放注册）
- **loginsmslogic.go**：验证码比对通过后保留至登录成功才删除（凭证不存在时可用于注册流程）
- 验证码 key 格式：`sms:code:{phone}`（TTL 300s，与 `sms:rate:{phone}` 限流分离）

### 影响
- Proto: 无
- 调用方: 无
- 数据库: 无
- 测试: 已有 4 个验证码相关测试用例覆盖过期/错误/成功场景

---

## 2026-06-04 — 全局公约与设计文档迁移

### 做了什么
- `CLAUDE.md` 新增 `## 全局公约` 章节，引用根 CLAUDE.md
- 设计文档迁移：`docs/specs/auth-design.md` → `services/auth-service/docs/design.md`
- 添加 `docs/CHANGELOG.md`（本文件）

### 为什么
项目规范化——统一文件布局，子 Claude 启动时能感知全局架构规则。

### 影响
- Proto: 无
- 调用方: 无
- 数据库: 无
