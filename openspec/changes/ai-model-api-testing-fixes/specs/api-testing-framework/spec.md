## ADDED Requirements

### Requirement: API endpoint inventory
系统SHALL提供完整的API端点清单，覆盖模型配置、API密钥、提示模板、使用统计四个子模块的所有端点。

#### Scenario: 列出所有API端点
- **WHEN** 开发者查看API清单
- **THEN** 系统显示所有端点的路径、方法、参数、响应格式

### Requirement: Request/Response validation
系统SHALL验证每个API的请求参数和响应格式是否符合规范。

#### Scenario: 验证请求参数格式
- **WHEN** 发送API请求
- **THEN** 系统验证参数类型、必填项、格式约束是否正确

#### Scenario: 验证响应格式
- **WHEN** 接收API响应
- **THEN** 系统验证响应符合`{ code: 0, message: "success", data: {...} }`格式

### Requirement: Error handling verification
系统SHALL验证每个API的错误处理是否完整和一致。

#### Scenario: 验证参数错误处理
- **WHEN** 发送无效参数
- **THEN** 系统返回明确的错误信息和正确的HTTP状态码

#### Scenario: 验证业务错误处理
- **WHEN** 触发业务逻辑错误
- **THEN** 系统返回统一格式的错误响应

### Requirement: Authentication verification
系统SHALL验证需要认证的API端点的JWT token验证逻辑。

#### Scenario: 验证未认证访问
- **WHEN** 不带token访问受保护端点
- **THEN** 系统返回401未授权错误

#### Scenario: 验证token过期处理
- **WHEN** 使用过期token访问
- **THEN** 系统返回token过期错误并提示刷新

### Requirement: Test case documentation
系统SHALL为每个API端点提供测试用例文档，包括正常场景和异常场景。

#### Scenario: 记录测试用例
- **WHEN** 完成API测试
- **THEN** 系统生成包含请求示例、预期响应、测试结果的文档
