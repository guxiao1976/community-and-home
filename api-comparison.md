# AI Model API 前后端对比分析

## 1. 模型配置管理 (Model Config)

### 1.1 获取模型列表
- **前端**: `GET /api/v1/models` - ✅ 匹配
- **后端**: `GET /api/v1/models` 
- **返回结构**: 
  - 后端: `{ code, message, data: { models: [], total } }`
  - 前端期望: `{ models: [], total }` (响应拦截器已处理)
- **状态**: ✅ 已修复 (ModelList.vue 已更新)

### 1.2 获取单个模型
- **前端**: `GET /api/v1/model/:id` - ✅ 匹配
- **后端**: `GET /api/v1/model/:id`
- **返回结构**: `{ code, message, data: ModelInfo }`
- **状态**: ✅ 正常

### 1.3 创建模型
- **前端**: `POST /api/v1/model` - ✅ 匹配
- **后端**: `POST /api/v1/model`
- **请求参数对比**:
  - ✅ name, display_name, provider, type, endpoint
  - ✅ max_tokens, supported_features
  - ✅ cost_per_1k_input_tokens, cost_per_1k_output_tokens
  - ✅ description
- **状态**: ✅ 正常

### 1.4 更新模型
- **前端**: `PUT /api/v1/model` - ✅ 匹配
- **后端**: `PUT /api/v1/model`
- **请求参数**: ✅ 所有字段匹配
- **状态**: ✅ 正常

### 1.5 删除模型
- **前端**: `DELETE /api/v1/model/:id` - ✅ 匹配
- **后端**: `DELETE /api/v1/model/:id`
- **状态**: ✅ 正常

## 2. API Key 管理

### 2.1 获取 API Keys 列表
- **前端**: `GET /api/v1/apikeys` - ✅ 匹配
- **后端**: `GET /api/v1/apikeys`
- **返回结构**: 
  - 后端: `{ code, message, data: { keys: [], total } }`
  - 前端期望: `PaginatedResponse<ApiKey>`
- **问题**: ⚠️ 字段名不匹配
  - 后端返回 `keys`
  - 前端可能期望 `list` 或 `items`

### 2.2 创建 API Key
- **前端**: `POST /api/v1/apikey`
- **后端**: `POST /api/v1/apikey`
- **参数对比**:
  - 前端: `model_id, provider, key_name, api_key`
  - 后端: `model_id, key_name, api_key, description`
- **问题**: ⚠️ 前端多了 `provider` 字段，后端没有

### 2.3 其他 API Key 操作
- **状态**: ✅ 路径和参数基本匹配

## 3. 模板管理 (Template)

### 3.1 获取模板列表
- **前端**: `GET /api/v1/templates` - ✅ 匹配
- **后端**: `GET /api/v1/templates`
- **返回结构**:
  - 后端: `{ code, message, data: { templates: [], total } }`
  - 前端期望: `PaginatedResponse<PromptTemplate>`
- **问题**: ⚠️ 字段名不匹配
  - 后端返回 `templates`
  - 前端可能期望 `list` 或 `items`

### 3.2 其他模板操作
- **状态**: ✅ 路径和参数基本匹配

## 4. 统计相关

### 4.1 使用统计
- **前端**: `GET /api/v1/statistics` - ✅ 匹配
- **后端**: `GET /api/v1/statistics`
- **返回结构**:
  - 后端: `{ code, message, data: { statistics: [], total } }`
- **问题**: ⚠️ 字段名可能不匹配

## 5. 响应码问题

### 5.1 成功码
- **后端返回**: 
  - 部分接口返回 `code: 0`
  - 部分接口返回 `code: 200`
- **前端期望**: 
  - `ErrorCode.Success = 0`
  - 响应拦截器已兼容 `0` 和 `200`
- **状态**: ✅ 已修复

## 需要修复的问题

### 高优先级
1. ✅ ModelList.vue - 已修复 (res.models 替代 res.data.list)
2. ⚠️ API Keys 列表 - 需要检查前端是否正确处理 `keys` 字段
3. ⚠️ Templates 列表 - 需要检查前端是否正确处理 `templates` 字段
4. ⚠️ Statistics 列表 - 需要检查前端是否正确处理 `statistics` 字段

### 中优先级
5. ⚠️ CreateAPIKey - 前端传递的 `provider` 字段后端不接受

### 低优先级
6. 统一后端响应码为 `0` (可选)
