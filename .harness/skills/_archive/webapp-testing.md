# WebApp Testing 技能使用模板

## 技能名称
`webapp-testing`

## 功能描述
执行端到端（E2E）自动化测试，使用 Playwright 验证用户流程和 UI 交互。

## 使用场景
- 完成新功能开发后进行集成测试
- 修复 UI bug 需要回归测试
- 发版前全流程验证
- 跨浏览器兼容性测试

## 输入要求

### 必需信息
- 测试目标（要测试的功能或页面）
- 用户流程（具体的操作步骤）
- 预期结果（每步操作的预期输出）

### 可选信息
- 测试环境 URL
- 测试数据（用户名、密码等）
- 浏览器要求（Chrome, Firefox, Safari）
- 截图/视频录制需求

## 输出内容
1. 测试用例执行报告
2. 失败用例的截图和日志
3. 测试覆盖率统计
4. 性能指标（可选）

## 调用示例

### 示例 1: 用户登录流程测试
```
测试用户登录流程：

前置条件：
- 测试账号：13800138000 / Test@123456
- 应用 URL: http://localhost:3003

测试步骤：
1. 打开登录页面
2. 输入手机号：13800138000
3. 输入密码：Test@123456
4. 点击"登录"按钮
5. 验证跳转到首页（URL 包含 /dashboard）
6. 验证顶部显示用户昵称

预期结果：
- 登录成功，跳转到首页
- 用户昵称正确显示
- 无控制台错误

浏览器：Chrome, Firefox
```

### 示例 2: 表单提交测试
```
测试用户信息编辑流程：

前置条件：
- 已登录用户
- 在用户列表页面

测试步骤：
1. 点击第一条用户记录的"编辑"按钮
2. 修改昵称为"测试用户_Updated"
3. 点击"保存"按钮
4. 验证显示成功提示消息
5. 验证列表中昵称已更新

边界条件：
- 昵称为空时显示错误提示
- 昵称超过30字符时显示错误提示

预期结果：
- 正常编辑成功
- 边界条件正确处理
```

### 示例 3: 批量操作测试
```
测试用户批量禁用功能：

测试步骤：
1. 在用户列表勾选3个用户
2. 点击"批量禁用"按钮
3. 确认二次提示弹窗
4. 验证显示成功提示
5. 验证被选用户状态变为"已禁用"
6. 验证这些用户无法登录

预期结果：
- 批量操作成功
- 状态正确更新
- 功能符合预期
```

## 最佳实践

### 1. 明确测试目标
✅ 好的测试描述：
```
测试用户登录流程：
1. 输入正确的手机号和密码 → 登录成功
2. 输入错误的密码 → 显示错误提示
3. 输入未注册的手机号 → 显示错误提示
```

❌ 模糊的测试描述：
```
测试登录功能
```

### 2. 覆盖边界条件
- 正常路径（Happy Path）
- 错误路径（Error Path）
- 边界值（空值、最大值、特殊字符）

### 3. 使用测试数据
```
测试数据：
- 有效用户：13800138000 / Test@123456
- 无效用户：13800138001 / WrongPassword
- 已禁用用户：13800138002 / Test@123456
```

### 4. 等待策略
明确何时需要等待：
- 等待 API 响应
- 等待页面加载
- 等待动画完成

## 测试类型

### 功能测试
验证功能是否按预期工作。

### 集成测试
验证多个模块协同工作。

### 回归测试
确保新代码没有破坏现有功能。

### 兼容性测试
在不同浏览器和设备上测试。

## 与其他技能配合

### webapp-testing + frontend-design
```
1. frontend-design 生成 UI 组件
2. webapp-testing 测试组件交互
3. 根据测试结果优化设计
```

### webapp-testing + audit-website
```
1. webapp-testing 验证功能正确性
2. audit-website 审计性能和可访问性
3. 综合优化
```

## 注意事项

1. **测试环境隔离**
   - 使用独立的测试数据库
   - 避免污染生产数据

2. **测试稳定性**
   - 避免依赖时间敏感的断言
   - 使用合理的超时时间
   - 处理偶发的网络延迟

3. **测试可维护性**
   - 使用 Page Object 模式
   - 复用公共测试逻辑
   - 清晰的测试命名

4. **性能考虑**
   - 并行执行测试用例
   - 合理使用 beforeEach/afterEach
   - 避免不必要的等待

## 常见问题

### Q: 测试偶尔失败怎么办？
A: 检查是否有竞态条件，增加等待时间，或使用更可靠的等待策略（waitForSelector）。

### Q: 如何测试需要登录的页面？
A: 使用 beforeEach 钩子在每个测试前自动登录，或使用 storage state 保存登录状态。

### Q: 测试执行太慢怎么办？
A: 启用并行执行，减少不必要的等待，使用 headless 模式。

### Q: 如何处理动态数据？
A: 使用数据驱动测试，或在测试前准备固定的测试数据。

## Playwright 配置示例

```typescript
// playwright.config.ts
import { defineConfig } from '@playwright/test'

export default defineConfig({
  testDir: './tests/e2e',
  timeout: 30000,
  retries: 2,
  use: {
    baseURL: 'http://localhost:3003',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
  },
  projects: [
    { name: 'Chrome', use: { browserName: 'chromium' } },
    { name: 'Firefox', use: { browserName: 'firefox' } },
  ],
})
```

## 相关文档
- [Playwright 官方文档](https://playwright.dev/)
- [Vitest 文档](https://vitest.dev/)
- [测试最佳实践](https://github.com/goldbergyoni/javascript-testing-best-practices)
