# CLAUDE.md

## 角色定位

这是 **Vue 3 管理后台**（`web/pc/`），Element Plus + TypeScript + Pinia + Vue Router。前端层入口见 [web/CLAUDE.md](../CLAUDE.md)。

## 关键规则

1. **Snowflake ID 处理**：
   - 所有 ID 字段 TypeScript 类型为 `string`（`web/common/types/`）
   - axios 使用 `lossless-json` 作为响应解析器（`transformResponse`）
   - 不再使用 `Number(route.params.id)`，直接使用 `route.params.id as string`
   - 详细信息见根目录 `CLAUDE.md` 第 5 条
2. **禁止在前端定义业务逻辑** — 所有业务规则属于后端服务
3. **API 接口必须与 api-proto 一致** — 接口契约定义在 `../../api-proto/`
4. **API 响应直接使用** — 后端遵循单层包装（`.harness/rules/项目编码规范.md §9`），axios 拦截器已剥掉外层 data，`res` 直接就是业务数据。**不需要** `res.data` 再取一层
5. **Vue 模板避免嵌套 `{{ }}`** — `{{` 在模板插值内会被解析为新插值起始。要展示花括号字面量，拆为 `{'{' + v + '}'}`。详见 `.harness/knowledge/memory/vue-template-nested-interpolation.md`
6. **API 调用禁止静默吞错** — catch 块至少要 `console.error`，关键操作需要 `ElMessage.error` 提示用户

## 架构

```
web/pc/
  src/
    api/           # API 调用层（identity.ts, masterdata.ts, aimodel.ts）
    stores/        # Pinia 状态管理（auth, permission, user, division）
    views/         # 页面组件
    components/    # 通用组件（business/ 业务组件）
    utils/         # 工具函数（request.ts — axios 实例 + lossless-json）
    router/        # Vue Router 配置
  tests/
    unit/          # Vitest 单元测试
    e2e/           # Playwright E2E 测试
```

## 常用命令

```bash
npm run dev           # Vite dev server (port 3003)
npm run build         # vue-tsc + vite build
npm run lint          # vue-tsc type check
npm run test:unit     # Vitest
npm run test:e2e      # Playwright
```

## 可用技能 (Skills)

本项目已集成以下 AI 技能，可在开发流水线中自动调用，或手动使用：

### 前端开发技能

#### 1. `/frontend-design` - UI 界面设计
**用途**: 生成现代化的 Vue 3 UI 组件和页面布局

**何时使用**:
- 开发新功能页面时
- 需要设计复杂表单或数据展示界面时
- 重构现有组件以提升用户体验时

**示例**:
```bash
# 在 Claude Code 中调用
/frontend-design 为用户管理模块设计一个用户列表页面，包含搜索、筛选、分页和批量操作功能
```

**自动调用**: 在 Harness 流水线的开发阶段，针对 feature 类型任务自动调用

---

#### 2. `/webapp-testing` - E2E 自动化测试
**用途**: 执行端到端测试，验证用户流程和 UI 交互

**何时使用**:
- 完成新功能开发后
- 修复 UI bug 需要回归测试时
- 发版前进行全流程验证时

**示例**:
```bash
# 在 Claude Code 中调用
/webapp-testing 测试用户登录流程：输入手机号和密码 → 点击登录 → 验证跳转到首页
```

**自动调用**: 在 Harness 流水线的 QA 阶段，单元测试通过后自动调用

**测试覆盖**:
- 关键用户流程（登录、核心功能操作）
- UI 交互正确性
- 错误处理和边界条件
- 跨浏览器兼容性 (Chrome, Firefox)

---

#### 3. `/audit-website` - 前端质量审计
**用途**: 全面审计前端质量（性能、可访问性、SEO、最佳实践）

**何时使用**:
- 功能开发完成，准备上线前
- 性能优化需要基线数据时
- 确保符合 WCAG 可访问性标准时

**示例**:
```bash
# 在 Claude Code 中调用
/audit-website 审计当前应用的性能和可访问性，生成 Lighthouse 报告
```

**自动调用**: 在 Harness 流水线的 Review 阶段，代码审查通过后自动调用

**审计维度**:
- **性能**: Lighthouse Performance Score, FCP, LCP, CLS, FID
- **可访问性**: WCAG 2.1 AA 标准，键盘导航，屏幕阅读器
- **SEO**: Meta 标签，语义化 HTML，移动端友好性
- **最佳实践**: HTTPS，控制台错误，现代 API

---

### 文档生成技能

#### 4. `/pdf` - 生成 PDF 文档
**用途**: 将 Markdown 文档或测试报告转换为专业的 PDF 格式

**示例**:
```bash
/pdf 将当前的架构设计文档生成为 PDF
```

---

#### 5. `/xlsx` - 生成 Excel 报告
**用途**: 生成数据化的测试报告、API 清单、统计表格

**示例**:
```bash
/xlsx 生成本次测试的覆盖率报告，包含测试用例清单和执行结果
```

---

### 流程管理技能

#### 6. `/writing-plans` - 任务分解
**用途**: 将复杂需求分解为结构化的开发任务

**示例**:
```bash
/writing-plans 将用户权限管理功能分解为可执行的开发任务
```

---

#### 7. `/requesting-code-review` - 发起代码审查
**用途**: 生成标准化的代码审查清单和 PR 描述

**示例**:
```bash
/requesting-code-review 为本次变更创建审查请求，包含变更清单和审查要点
```

---

## 技能集成到流水线

当使用 Harness Pipeline 进行开发时，以下技能会自动调用：

```
阶段1: 需求分析
  ↓
阶段2: 架构设计
  ├─ [自动] writing-plans: 精细化任务分解
  └─ [自动] frontend-design: UI 设计（前端 feature 任务）
  ↓
阶段3: 开发
  ↓
阶段4: QA
  ├─ [自动] webapp-testing: E2E 测试（前端项目）
  └─ [自动] xlsx: 测试报告生成
  ↓
阶段5: Review
  ├─ [自动] requesting-code-review: 标准化审查
  └─ [自动] audit-website: 质量审计（前端项目）
  ↓
阶段6: 文档交付
  └─ [自动] pdf: 生成交付文档
```

**注意**: 技能调用失败不会阻塞主流程，会记录警告并继续执行。

# Auth Service 后端修改清单（历史）

> 由 web/pc 前端 Claude 整理。前端已完成类型对齐和 RSA 加密，以下 3 个后端问题需要修复。

## 背景

前端登录/注册请求已改为 RSA 加密传输：
- `LoginRequest`: `{ encryptedPhone, encryptedPassword, deviceId, deviceType }`
- `RegisterRequest`: `{ encryptedPhone, encryptedPassword?, smsCode, nickname, deviceId, deviceType }`
- 加密方式：RSA-OAEP + SHA-256 + Base64（对齐 `common/pkg/crypto/rsa.go` 的 `RSAEncrypt`）
- 公钥获取：`GET /api/auth/public-key`

## 问题 1（P0）：缺少公钥端点

**现状**：没有对外暴露 RSA 公钥的 API，前端无法加密。

**修改**：

### 1a. common/pkg/crypto/rsa.go — 新增 getter
```go
// GetRSAPublicKey 返回当前全局 RSA 公钥（PEM 格式）
func GetRSAPublicKey() string {
    rsaKeyMutex.RLock()
    defer rsaKeyMutex.RUnlock()
    return /* PEM 字符串，从 globalRSAPublicKey 反序列化，或直接存 PEM */
}
```
> 当前 `InitRSA` 存的是 `*rsa.PublicKey` 对象，需要额外存一份 PEM 字符串，或新增一个 getter 返回 PEM。

### 1b. api/internal/config/config.go — 新增配置字段
```go
RsaPublicKey string  // RSA 公钥（PEM，与 RPC 配置相同）
```

### 1c. api/etc/auth-api.yaml — 新增配置值
```yaml
RsaPublicKey: "-----BEGIN PUBLIC KEY-----\n...\n-----END PUBLIC KEY-----"
```

### 1d. api/internal/types/types.go — 新增响应类型
```go
type PublicKeyResp struct {
    PublicKey string `json:"public_key"`
}
```

### 1e. api/internal/logic/publickeylogic.go（新建）— 逻辑层
```go
func (l *PublicKeyLogic) GetPublicKey() (*types.PublicKeyResp, error) {
    return &types.PublicKeyResp{PublicKey: l.svcCtx.Config.RsaPublicKey}, nil
}
```

### 1f. api/internal/handler/routes.go — 注册路由
在公开路由组（`rest.WithPrefix("/api/auth")`）中添加：
```go
{
    Method:  http.MethodGet,
    Path:    "/public-key",
    Handler: PublicKeyHandler(svcCtx),
},
```

---

## 问题 2（P0）：RegisterReq 字段名不匹配

**现状**：后端 `RegisterReq` 的 phone 字段是 `json:"phone"`（明文），但前端现在发 `encryptedPhone`（RSA 密文）。

**修改**：

### 2a. api/internal/types/types.go
```go
// 改前
type RegisterReq struct {
    Phone    string `json:"phone"`
    ...
}

// 改后
type RegisterReq struct {
    EncryptedPhone string `json:"encrypted_phone"`
    ...
}
```

### 2b. api/internal/logic/registerlogic.go
```go
// 改前
resp, err := l.svcCtx.AuthRpc.Register(l.ctx, &authv1.RegisterRequest{
    Phone: req.Phone,
    ...
})

// 改后
resp, err := l.svcCtx.AuthRpc.Register(l.ctx, &authv1.RegisterRequest{
    Phone: req.EncryptedPhone,  // 字段名变更
    ...
})
```

> 注意：proto 中 `RegisterRequest.phone` 是明文，但这里 API 层收到的是 RSA 密文。需要在 API 层解密后传给 gRPC，或者在 gRPC 层处理。建议在 API 层 `registerlogic.go` 中调用 `crypto.RSADecrypt(req.EncryptedPhone)` 解密后传入 proto。

---

## 问题 3（P1）：短信验证码未校验

**现状**：`rpc/internal/logic/auth/registerlogic.go:47` 有 TODO，验证码在 Redis 中但从未被读取比对。任意 6 位数字都能通过注册。

**修改**：

### 3a. rpc/internal/logic/auth/registerlogic.go
在 `// 1. 校验短信验证码` 处（约 L47），增加：
```go
// 从 Redis 读取验证码
codeKey := fmt.Sprintf("sms:code:%s", in.Phone)
storedCode, err := l.svcCtx.RedisClient.Get(l.ctx, codeKey).Result()
if err != nil || storedCode == "" {
    return &authv1.RegisterResponse{
        Base: responsex.NewBaseRespWithError(50004, "验证码已过期，请重新获取"),
    }, nil
}
if storedCode != in.SmsCode {
    return &authv1.RegisterResponse{
        Base: responsex.NewBaseRespWithError(50004, "验证码错误"),
    }, nil
}
// 验证通过，删除验证码（防重放）
l.svcCtx.RedisClient.Del(l.ctx, codeKey)
```

---

## 操作顺序

```
1. 修改 common/pkg/crypto/rsa.go（如需全局 Claude 协作）
2. 修改 api/internal/config + yaml（公钥配置）
3. 实现 public-key 端点（types → logic → handler → routes）
4. 修改 RegisterReq 字段名 + 解密逻辑
5. 实现短信验证码校验
6. go build ./... 验证编译
7. 重启 auth-service API + RPC
```

## 参考

- RSA 工具：`common/pkg/crypto/rsa.go`（`InitRSA`, `RSADecrypt`, `RSAEncrypt`）
- 现有端点模式：`api/internal/handler/routes.go`（RegisterHandlers）
- 登录逻辑参考：`rpc/internal/logic/auth/loginlogic.go`（RSA 解密 → bcrypt 校验）
- 前端类型：`web/common/types/identity.ts`（LoginRequest, RegisterRequest 等）
