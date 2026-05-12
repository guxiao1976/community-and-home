## ADDED Requirements

### Requirement: 文本审核API调用
系统SHALL提供文本审核API接口调用功能，接受文本内容作为参数，返回多层级审核结果。

#### Scenario: 调用文本审核接口
- **WHEN** 前端发起文本审核请求，传入待审核文本
- **THEN** 系统调用内容审核微服务的文本审核接口
- **THEN** 系统返回包含传统技术、小模型、大模型三个层级的审核结果

#### Scenario: 文本审核超时处理
- **WHEN** 文本审核接口调用超过30秒未返回
- **THEN** 系统返回超时错误，提示"审核服务响应超时，请稍后重试"

### Requirement: 图片审核API调用
系统SHALL提供图片审核API接口调用功能，接受图片文件作为参数，返回审核结果。

#### Scenario: 调用图片审核接口
- **WHEN** 前端发起图片审核请求，上传图片文件
- **THEN** 系统调用内容审核微服务的图片审核接口
- **THEN** 系统返回包含小模型检测结果的审核数据

#### Scenario: 图片审核超时处理
- **WHEN** 图片审核接口调用超过60秒未返回
- **THEN** 系统返回超时错误，提示"审核服务响应超时，请稍后重试"

### Requirement: API请求格式
系统SHALL按照内容审核微服务的接口规范构造请求，包括必要的认证信息和请求参数。

#### Scenario: 文本审核请求格式
- **WHEN** 构造文本审核请求
- **THEN** 请求MUST包含JWT认证token
- **THEN** 请求body MUST包含text字段（待审核文本）
- **THEN** 请求header MUST设置Content-Type为application/json

#### Scenario: 图片审核请求格式
- **WHEN** 构造图片审核请求
- **THEN** 请求MUST包含JWT认证token
- **THEN** 请求MUST使用multipart/form-data格式上传图片文件
- **THEN** 图片文件字段名MUST为image

### Requirement: API响应解析
系统SHALL正确解析内容审核微服务返回的响应数据，提取各层级的审核结果。

#### Scenario: 解析文本审核响应
- **WHEN** 收到文本审核接口响应
- **THEN** 系统解析传统技术检测结果（包括关键词匹配、敏感词检测等）
- **THEN** 系统解析小模型检测结果（包括模型置信度、分类结果等）
- **THEN** 系统解析大模型检测结果（如果返回）

#### Scenario: 解析图片审核响应
- **WHEN** 收到图片审核接口响应
- **THEN** 系统解析小模型检测结果（包括检测标签、置信度等）
- **THEN** 系统识别大模型结果是否存在

### Requirement: 错误处理
系统SHALL处理API调用过程中的各种错误情况，并返回友好的错误信息。

#### Scenario: 网络错误处理
- **WHEN** API调用因网络问题失败
- **THEN** 系统返回错误信息"网络连接失败，请检查网络后重试"

#### Scenario: 服务不可用处理
- **WHEN** 内容审核微服务返回5xx错误
- **THEN** 系统返回错误信息"审核服务暂时不可用，请稍后重试"

#### Scenario: 参数错误处理
- **WHEN** 内容审核微服务返回4xx错误
- **THEN** 系统解析错误详情并返回具体的参数错误信息

### Requirement: 响应数据结构
系统SHALL返回统一的响应数据结构，包含各层级的原始返回值和处理状态。

#### Scenario: 文本审核响应结构
- **WHEN** 文本审核完成
- **THEN** 响应MUST包含traditional字段（传统技术结果）
- **THEN** 响应MUST包含smallModel字段（小模型结果）
- **THEN** 响应MUST包含largeModel字段（大模型结果，可为null）
- **THEN** 每个层级MUST包含status（成功/失败）、data（原始返回数据）、timestamp（处理时间）

#### Scenario: 图片审核响应结构
- **WHEN** 图片审核完成
- **THEN** 响应MUST包含smallModel字段（小模型结果）
- **THEN** 响应MUST包含largeModel字段（预留，当前为null）
- **THEN** 每个层级MUST包含status（成功/失败）、data（原始返回数据）、timestamp（处理时间）

### Requirement: API端点配置
系统SHALL支持配置内容审核微服务的API端点地址，便于不同环境切换。

#### Scenario: 使用环境变量配置
- **WHEN** 系统启动时
- **THEN** 系统从环境变量读取内容审核服务的base URL
- **THEN** 如果环境变量未设置，使用默认的开发环境地址

#### Scenario: API版本管理
- **WHEN** 调用内容审核接口
- **THEN** 系统使用配置的API版本号构造完整的请求URL
- **THEN** 当前版本MUST为v1
