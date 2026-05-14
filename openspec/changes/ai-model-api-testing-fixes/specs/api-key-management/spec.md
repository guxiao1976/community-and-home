## MODIFIED Requirements

### Requirement: Create API key
系统SHALL允许用户创建新的API密钥，用于访问第三方AI模型服务。

#### Scenario: 成功创建API密钥
- **WHEN** 用户提交有效的API密钥信息（名称、提供商、密钥值）
- **THEN** 系统加密存储密钥并返回密钥ID

#### Scenario: 验证密钥格式
- **WHEN** 用户提交格式不正确的密钥
- **THEN** 系统返回格式验证错误

#### Scenario: 验证密钥名称唯一性
- **WHEN** 用户创建重复名称的密钥
- **THEN** 系统返回名称冲突错误

### Requirement: List API keys
系统SHALL提供API密钥列表查询，支持分页和筛选，密钥值应脱敏显示。

#### Scenario: 查询密钥列表
- **WHEN** 用户请求密钥列表
- **THEN** 系统返回密钥列表，密钥值显示为`***...后4位`

#### Scenario: 按提供商筛选
- **WHEN** 用户筛选特定提供商的密钥
- **THEN** 系统只返回该提供商的密钥

#### Scenario: 按状态筛选
- **WHEN** 用户筛选有效或已过期的密钥
- **THEN** 系统返回符合状态条件的密钥

### Requirement: Get API key detail
系统SHALL允许用户查询单个API密钥的详细信息，密钥值脱敏显示。

#### Scenario: 查询密钥详情
- **WHEN** 用户请求有效密钥ID的详情
- **THEN** 系统返回密钥信息，密钥值脱敏

#### Scenario: 查看完整密钥值
- **WHEN** 用户点击"显示完整密钥"
- **THEN** 系统验证权限后返回完整密钥值

### Requirement: Update API key
系统SHALL允许用户更新API密钥的名称、备注等信息（不包括密钥值本身）。

#### Scenario: 更新密钥元数据
- **WHEN** 用户更新密钥名称或备注
- **THEN** 系统更新信息并保持密钥值不变

#### Scenario: 轮换密钥值
- **WHEN** 用户提交新的密钥值进行轮换
- **THEN** 系统更新密钥值并记录轮换历史

### Requirement: Delete API key
系统SHALL允许用户删除API密钥（软删除）。

#### Scenario: 删除未使用的密钥
- **WHEN** 用户删除未被模型配置引用的密钥
- **THEN** 系统标记密钥为已删除

#### Scenario: 删除被引用的密钥
- **WHEN** 用户尝试删除正在被使用的密钥
- **THEN** 系统返回错误并列出引用该密钥的模型配置

### Requirement: Validate API key
系统SHALL提供API密钥有效性验证功能。

#### Scenario: 测试密钥连接
- **WHEN** 用户点击"测试连接"
- **THEN** 系统使用该密钥调用提供商API并返回测试结果

#### Scenario: 检测密钥过期
- **WHEN** 系统检测到密钥已过期或失效
- **THEN** 系统标记密钥状态并通知用户
