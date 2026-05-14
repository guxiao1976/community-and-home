# AI模型管理模块 API 清单

## 1. 模型配置 (Model Config)

### 前端API调用
| 方法 | 路径 | 前端函数 | 说明 |
|------|------|----------|------|
| GET | `/api/v1/models` | `getModelConfigs()` | 获取模型配置列表 |
| GET | `/api/v1/model/:id` | `getModelConfigById(id)` | 获取单个模型配置 |
| POST | `/api/v1/model` | `createModelConfig(data)` | 创建模型配置 |
| PUT | `/api/v1/model` | `updateModelConfig(data)` | 更新模型配置 |
| DELETE | `/api/v1/model/:id` | `deleteModelConfig(id)` | 删除模型配置 |

### 后端API定义
| 方法 | 路径 | Handler | 说明 |
|------|------|---------|------|
| GET | `/api/v1/models` | `getModels` | 获取可用模型列表 ✓ |
| POST | `/api/v1/model` | `createModel` | 创建模型配置 ✓ |
| PUT | `/api/v1/model` | `updateModel` | 更新模型配置 ✓ |
| DELETE | `/api/v1/model/:id` | `deleteModel` | 删除模型配置 ✓ |

### 不匹配问题
- ❌ **前端调用 `GET /api/v1/model/:id` 获取单个模型详情，但后端未定义此端点**

---

## 2. API密钥管理 (API Key)

### 前端API调用
| 方法 | 路径 | 前端函数 | 说明 |
|------|------|----------|------|
| GET | `/api/v1/apikeys` | `getApiKeys()` | 获取API密钥列表 |
| GET | `/api/v1/apikey/:id` | `getApiKeyById(id)` | 获取单个API密钥 |
| POST | `/api/v1/apikey` | `createApiKey(data)` | 创建API密钥 |
| PUT | `/api/v1/apikey` | `updateApiKey(data)` | 更新API密钥 |
| DELETE | `/api/v1/apikey/:id` | `deleteApiKey(id)` | 删除API密钥 |

### 后端API定义
| 方法 | 路径 | Handler | 说明 |
|------|------|---------|------|
| GET | `/api/v1/apikeys` | `listAPIKeys` | 获取API Key列表 ✓ |
| POST | `/api/v1/apikey` | `createAPIKey` | 创建API Key ✓ |
| PUT | `/api/v1/apikey` | `updateAPIKey` | 更新API Key ✓ |
| DELETE | `/api/v1/apikey/:id` | `deleteAPIKey` | 删除API Key ✓ |

### 不匹配问题
- ❌ **前端调用 `GET /api/v1/apikey/:id` 获取单个API密钥详情，但后端未定义此端点**

### 字段不匹配
- ⚠️ 前端字段：`model_config_id`, `provider`, `key_name`, `api_key`
- ⚠️ 后端字段：`model_id`, `key_name`, `api_key`, `description`
- **问题：前端使用 `model_config_id` 和 `provider`，后端使用 `model_id`，字段名不一致**

---

## 3. 提示模板 (Prompt Template)

### 前端API调用
| 方法 | 路径 | 前端函数 | 说明 |
|------|------|----------|------|
| GET | `/api/v1/templates` | `getTemplates()` | 获取模板列表 |
| GET | `/api/v1/template/:id` | `getTemplateById(id)` | 获取单个模板 |
| POST | `/api/v1/template` | `createTemplate(data)` | 创建模板 |
| PUT | `/api/v1/template` | `updateTemplate(data)` | 更新模板 |
| DELETE | `/api/v1/template/:id` | `deleteTemplate(id)` | 删除模板 |

### 后端API定义
| 方法 | 路径 | Handler | 说明 |
|------|------|---------|------|
| GET | `/api/v1/templates` | `listTemplates` | 获取提示词模板列表 ✓ |
| POST | `/api/v1/template` | `createTemplate` | 创建提示词模板 ✓ |
| PUT | `/api/v1/template` | `updateTemplate` | 更新提示词模板 ✓ |
| DELETE | `/api/v1/template/:id` | `deleteTemplate` | 删除提示词模板 ✓ |

### 不匹配问题
- ❌ **前端调用 `GET /api/v1/template/:id` 获取单个模板详情，但后端未定义此端点**

---

## 4. 使用统计 (Usage Statistics)

### 前端API调用
| 方法 | 路径 | 前端函数 | 说明 |
|------|------|----------|------|
| GET | `/api/v1/statistics` | `getUsageStatistics()` | 获取使用统计 |

### 后端API定义
| 方法 | 路径 | Handler | 说明 |
|------|------|---------|------|
| GET | `/api/v1/cost/stats` | `getCostStats` | 获取成本统计 |
| GET | `/api/v1/cost/alert/config` | `getAlertConfig` | 获取预警配置 |
| PUT | `/api/v1/cost/alert/config` | `updateAlertConfig` | 更新预警配置 |
| GET | `/api/v1/cost/alert/records` | `listAlertRecords` | 获取预警记录 |

### 不匹配问题
- ❌ **前端调用 `GET /api/v1/statistics` 获取使用统计，但后端未定义此端点**
- ❌ **后端提供的是成本统计API (`/api/v1/cost/stats`)，与前端期望的使用统计不匹配**
- ❌ **前端缺少成本预警相关的API调用**

---

## 5. 其他API

### 后端额外提供的API（前端未使用）
| 方法 | 路径 | Handler | 说明 |
|------|------|---------|------|
| POST | `/api/v1/model/call` | `callModel` | 调用模型 |
| POST | `/api/v1/model/call/batch` | `callModelBatch` | 批量调用模型 |
| GET | `/api/v1/health` | `healthCheck` | 健康检查 |

### 前端额外调用的API（后端未提供）
| 方法 | 路径 | 前端函数 | 说明 |
|------|------|----------|------|
| POST | `/api/v1/model/call` | `callModel(data)` | 调用AI模型 ✓ (后端已提供) |
| GET | `/api/v1/health-checks` | `getHealthChecks()` | 获取健康检查记录 ❌ |
| POST | `/api/v1/model/:id/health-check` | `triggerHealthCheck(id)` | 触发健康检查 ❌ |

---

---

## 后端路由注册确认

已验证 `services/ai-model/api/internal/handler/routes.go`，所有路由已正确注册：

### 已注册的路由
✓ Model: `/api/v1/models` (GET), `/api/v1/model` (POST/PUT), `/api/v1/model/:id` (DELETE), `/api/v1/model/call` (POST), `/api/v1/model/call/batch` (POST), `/api/v1/health` (GET)
✓ APIKey: `/api/v1/apikeys` (GET), `/api/v1/apikey` (POST/PUT), `/api/v1/apikey/:id` (DELETE)
✓ Template: `/api/v1/templates` (GET), `/api/v1/template` (POST/PUT), `/api/v1/template/:id` (DELETE)
✓ Cost: `/api/v1/cost/stats` (GET), `/api/v1/cost/alert/config` (GET/PUT), `/api/v1/cost/alert/records` (GET)

---

## 总结：关键问题

### 缺失的后端API端点（需要添加）
1. `GET /api/v1/model/:id` - 获取单个模型配置详情
2. `GET /api/v1/apikey/:id` - 获取单个API密钥详情
3. `GET /api/v1/template/:id` - 获取单个模板详情
4. `GET /api/v1/statistics` - 获取使用统计（或前端改用 `/api/v1/cost/stats`）
5. `GET /api/v1/health-checks` - 获取健康检查记录
6. `POST /api/v1/model/:id/health-check` - 触发健康检查

### 字段不匹配问题
1. API密钥：前端 `model_config_id` vs 后端 `model_id`
2. API密钥：前端有 `provider` 字段，后端无此字段
3. 时间字段：前端使用 `created_time/updated_time`，后端使用 `created_at/updated_at`

### 功能不匹配问题
1. 使用统计 vs 成本统计：前端期望的是调用次数、token使用等统计，后端提供的是成本统计
2. 前端缺少成本预警相关功能的调用

### 测试优先级
1. **高优先级**：修复缺失的详情查询API（model/apikey/template的单个查询）
2. **高优先级**：统一字段命名（created_time vs created_at等）
3. **中优先级**：明确使用统计与成本统计的关系，决定是合并还是分开
4. **低优先级**：补充健康检查相关API
