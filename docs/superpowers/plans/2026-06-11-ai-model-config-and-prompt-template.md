# AI 模型配置 & 提示词模板增强 — 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 增强模型配置页（新增 config_key + 连接测试）和提示词模板页（新增模型绑定 + 变量识别 + 即时测试）

**Architecture:** REST API 网关 (Go, :8891) → gRPC 客户端 → gRPC 服务 (Go, :8084) → MySQL + 适配器。前端 Vue 3 + Element Plus + TypeScript 表单页面

**Tech Stack:** Go (go-zero), Vue 3, Element Plus, TypeScript, MySQL, Playwright (测试)

**Context:** 现有服务已运行 — API :8891, RPC :8084; 前端 :3003。PromptTemplate 表已有 `variables` 字段

---

### Task 1: 数据库迁移 — 新增 config_key 和 temperature 列

**Files:**
- Create: `services/ai-model-service/sql/migration_004_config_key.sql`

- [ ] **Step 1: 创建迁移 SQL**

```sql
-- Migration 004: Add config_key and temperature to am_model_config
-- Add config_key to am_prompt_template

ALTER TABLE `am_model_config`
  ADD COLUMN `config_key` VARCHAR(64) NOT NULL DEFAULT '' AFTER `model_name`,
  ADD COLUMN `temperature` DECIMAL(3,2) NOT NULL DEFAULT 0.70 AFTER `max_retries`,
  ADD UNIQUE INDEX `idx_config_key` (`config_key`);

-- Update existing rows: set config_key = model_name for existing data
UPDATE `am_model_config` SET `config_key` = `model_name` WHERE `config_key` = '';

ALTER TABLE `am_prompt_template`
  ADD COLUMN `config_key` VARCHAR(64) NOT NULL DEFAULT '' AFTER `template_name`,
  ADD INDEX `idx_config_key` (`config_key`);
```

- [ ] **Step 2: 执行迁移**

```bash
docker exec -i mysql mysql -uroot -proot123456 ai_model_db < services/ai-model-service/sql/migration_004_config_key.sql
```

- [ ] **Step 3: 验证**

```bash
docker exec -i mysql mysql -uroot -proot123456 ai_model_db -e "DESC am_model_config;" | grep -E "config_key|temperature"
docker exec -i mysql mysql -uroot -proot123456 ai_model_db -e "DESC am_prompt_template;" | grep config_key
```

---

### Task 2: 更新 Go 数据模型 — AmModelConfig

**Files:**
- Modify: `services/ai-model-service/rpc/model/ammodelconfigmodel_gen.go`

- [ ] **Step 1: 添加 ConfigKey 和 Temperature 字段到 struct**

在 `AmModelConfig` struct 的 `ModelName` 字段后添加：

```go
ConfigKey             string         `db:"config_key"`              // 不可变唯一标识，如 gpt4-moderation
```

在 `MaxRetries` 字段后添加：

```go
Temperature           float64        `db:"temperature"`             // 模型温度，默认 0.7
```

- [ ] **Step 2: 添加 FindOneByConfigKey 到接口和实现**

在 `ammodelconfigmodel.go` 的 `AmModelConfigModel` 接口中添加：

```go
FindOneByConfigKey(ctx context.Context, configKey string) (*AmModelConfig, error)
```

并实现该方法（参照 `FindOneByModelName` 模式，使用 `config_key` 列）。

- [ ] **Step 3: 更新 Insert 方法** — 在 values 中添加 `data.ConfigKey` 和 `data.Temperature`，调整占位符数量

- [ ] **Step 4: 更新 Update 方法** — 在 set 值中添加 `newData.ConfigKey` 和 `newData.Temperature`

- [ ] **Step 5: 验证编译**

```bash
cd services/ai-model-service && go build ./rpc/...
```

---

### Task 3: 更新 Go 数据模型 — AmPromptTemplate

**Files:**
- Modify: `services/ai-model-service/rpc/model/amprompttemplatemodel_gen.go`

- [ ] **Step 1: 在 AmPromptTemplate struct 的 TemplateName 后添加 ConfigKey 字段**

```go
ConfigKey             string         `db:"config_key"`              // 关联的模型配置 Key
```

- [ ] **Step 2: 更新 Insert 方法** — 添加 `data.ConfigKey`，占位符 12→13

- [ ] **Step 3: 更新 Update 方法** — 添加 `newData.ConfigKey`，占位符 12→13

- [ ] **Step 4: 验证编译**

```bash
cd services/ai-model-service && go build ./rpc/...
```

---

### Task 4: 更新 RPC 层 — 模型配置 CRUD 适配新字段

**Files:**
- Modify: `services/ai-model-service/rpc/internal/logic/createmodelconfiglogic.go`
- Modify: `services/ai-model-service/rpc/internal/logic/updatemodelconfiglogic.go`
- Modify: `services/ai-model-service/rpc/internal/logic/getmodelconfiglogic.go`
- Modify: `services/ai-model-service/rpc/internal/logic/getavailablemodelslogic.go`

- [ ] **Step 1: CreateModelConfigLogic — 添加 ConfigKey 校验和写入**

在 Insert 调用前添加 ConfigKey 唯一性校验：

```go
// 校验 config_key 格式
if in.ConfigKey == "" {
    return nil, errs.New(errs.CodeInvalidArgument, "config_key is required").ToGrpcStatus()
}
if !regexp.MustCompile(`^[a-z0-9-]+$`).MatchString(in.ConfigKey) {
    return nil, errs.New(errs.CodeInvalidArgument, "config_key must be [a-z0-9-]+").ToGrpcStatus()
}
// 唯一性检查
existing, err := l.svcCtx.ModelConfigModel.FindOneByConfigKey(l.ctx, in.ConfigKey)
if err == nil && existing != nil {
    return nil, errs.New(errs.CodeInvalidArgument, "config_key already exists").ToGrpcStatus()
}
```

在 `model.AmModelConfig{}` 初始化中添加：
```go
ConfigKey:   in.ConfigKey,
Temperature: float64(in.Temperature),
```

- [ ] **Step 2: UpdateModelConfigLogic — 排除 config_key 更新（不可变）**

确保更新时不修改 `config_key`。若请求中包含 `ConfigKey` 则忽略或报错。

- [ ] **Step 3: GetModelConfigLogic / GetAvailableModelsLogic — 响应包含新字段**

在构建响应 pb 对象时添加：
```go
ConfigKey:   record.ConfigKey,
Temperature: float32(record.Temperature),
```

- [ ] **Step 4: 验证编译和测试**

```bash
cd services/ai-model-service && go build ./rpc/... && go vet ./rpc/...
```

---

### Task 5: 新增 RPC 层 — 连接测试和模板测试 Logic

**Files:**
- Create: `services/ai-model-service/rpc/internal/logic/testmodelconnectionlogic.go`
- Create: `services/ai-model-service/rpc/internal/logic/testtemplatelogic.go`

- [ ] **Step 1: TestModelConnectionLogic — 发送最小化请求验证连接**

```go
// testmodelconnectionlogic.go
package logic

import (
    "context"
    "fmt"
    "time"
    
    "github.com/guxiao/community-and-home/services/ai-model/rpc/internal/manager"
    "github.com/guxiao/community-and-home/services/ai-model/rpc/internal/svc"
    "github.com/guxiao/community-and-home/services/ai-model/rpc/model"
    "github.com/guxiao/community-and-home/services/ai-model/rpc/pkg/errs"
    pb "github.com/guxiao1976/api-proto/gen/go/aimodel/v1"
    "github.com/zeromicro/go-zero/core/logx"
)

type TestModelConnectionLogic struct {
    ctx    context.Context
    svcCtx *svc.ServiceContext
    logx.Logger
}

func NewTestModelConnectionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TestModelConnectionLogic {
    return &TestModelConnectionLogic{
        ctx:    ctx,
        svcCtx: svcCtx,
        Logger: logx.WithContext(ctx),
    }
}

func (l *TestModelConnectionLogic) TestModelConnection(in *pb.TestModelConnectionReq) (*pb.TestModelConnectionResp, error) {
    // 获取模型配置
    config, err := l.svcCtx.ModelConfigModel.FindOneByConfigKey(l.ctx, in.ConfigKey)
    if err != nil {
        return nil, errs.New(errs.CodeModelConfigNotFound, "config_key not found").ToGrpcStatus()
    }
    
    // 获取 API Key
    apiKey, err := l.svcCtx.ApiKeyModel.FindOne(l.ctx, config.ApiKeyId.Int64)
    if err != nil {
        return nil, errs.New(errs.CodeApiKeyNotFound, "api key not found").ToGrpcStatus()
    }
    
    // 构建临时适配器进行连接测试
    start := time.Now()
    err = l.svcCtx.ModelManager.TestConnection(l.ctx, config, apiKey.ApiKey)
    latency := time.Since(start).Milliseconds()
    
    if err != nil {
        return &pb.TestModelConnectionResp{
            Base:      successResp(),
            Success:   false,
            LatencyMs: latency,
            Error:     err.Error(),
        }, nil
    }
    
    return &pb.TestModelConnectionResp{
        Base:      successResp(),
        Success:   true,
        LatencyMs: latency,
    }, nil
}
```

- [ ] **Step 2: TestTemplateLogic — 变量替换 + 模型调用**

```go
// testtemplatelogic.go
package logic

import (
    "context"
    "encoding/json"
    "regexp"
    "strings"
    
    "github.com/guxiao/community-and-home/services/ai-model/rpc/internal/svc"
    "github.com/guxiao/community-and-home/services/ai-model/rpc/pkg/errs"
    pb "github.com/guxiao1976/api-proto/gen/go/aimodel/v1"
    "github.com/zeromicro/go-zero/core/logx"
)

type TestTemplateLogic struct {
    ctx    context.Context
    svcCtx *svc.ServiceContext
    logx.Logger
}

func NewTestTemplateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TestTemplateLogic {
    return &TestTemplateLogic{
        ctx:    ctx,
        svcCtx: svcCtx,
        Logger: logx.WithContext(ctx),
    }
}

func (l *TestTemplateLogic) TestTemplate(in *pb.TestTemplateReq) (*pb.TestTemplateResp, error) {
    // 1. 变量替换
    rendered := in.Content
    if in.Variables != nil {
        for k, v := range in.Variables {
            placeholder := "{{" + k + "}}"
            rendered = strings.ReplaceAll(rendered, placeholder, v)
        }
    }
    
    // 2. 检查是否有未替换的变量
    unreplaced := regexp.MustCompile(`\{\{(\w+)\}\}`).FindAllString(rendered, -1)
    
    // 3. 调用模型
    callResp, err := NewCallModelLogic(l.ctx, l.svcCtx).CallModel(&pb.ModelCallRequest{
        ConfigKey: in.ConfigKey,
        Prompt:    rendered,
    })
    if err != nil {
        return nil, err
    }
    
    return &pb.TestTemplateResp{
        Base:        successResp(),
        Rendered:    rendered,
        Unreplaced:  unreplaced,
        Content:     callResp.Content,
        InputTokens: callResp.InputTokens,
        OutputTokens: callResp.OutputTokens,
        Cost:       callResp.Cost,
        Latency:    callResp.Latency,
    }, nil
}
```

- [ ] **Step 3: 在 RPC server 中注册新 handler**

在 `rpc/internal/server/` 中更新 server 实现，将新的 Logic 注册为 gRPC 方法。

- [ ] **Step 4: 验证编译**

```bash
cd services/ai-model-service && go build ./rpc/...
```

---

### Task 6: 更新 API 层类型定义和 Handler

**Files:**
- Modify: `services/ai-model-service/api/internal/types/types.go`
- Modify: `services/ai-model-service/api/internal/handler/routes.go`
- Create: `services/ai-model-service/api/internal/handler/model/testconnectionhandler.go`
- Create: `services/ai-model-service/api/internal/logic/model/testconnectionlogic.go`
- Create: `services/ai-model-service/api/internal/handler/template/testtemplatehandler.go`
- Create: `services/ai-model-service/api/internal/logic/template/testtemplatelogic.go`

- [ ] **Step 1: 更新 ModelInfo 类型**

在 `ModelInfo` struct 中添加：
```go
ConfigKey   string  `json:"config_key"`
Temperature float64 `json:"temperature"`
```

- [ ] **Step 2: 更新 CreateModelRequest / UpdateModelRequest**

添加 ConfigKey 和 Temperature 字段。

- [ ] **Step 3: 更新 TemplateInfo 类型**

在 `TemplateInfo` struct 中添加：
```go
ConfigKey string   `json:"config_key"`
Variables []string `json:"variables"`
```

- [ ] **Step 4: 新增连接测试和模板测试的 Request/Response 类型**

- [ ] **Step 5: 新增 API Handler 和 Logic（代理到 gRPC）**

参照现有 handler 模式（如 `createmodelhandler.go` + `createmodellogic.go`），创建 test-connection 和 template/test handler+logic。

- [ ] **Step 6: 在 routes.go 中注册新路由**

在 model 路由组添加：
```go
{
    Method:  http.MethodPost,
    Path:    "/model/test-connection",
    Handler: model.TestConnectionHandler(serverCtx),
},
```

在 template 路由组添加：
```go
{
    Method:  http.MethodPost,
    Path:    "/template/test",
    Handler: template.TestTemplateHandler(serverCtx),
},
```

- [ ] **Step 7: 验证编译**

---

### Task 7: 更新 ModelManager — 连接测试方法

**Files:**
- Modify: `services/ai-model-service/rpc/internal/manager/manager.go`

- [ ] **Step 1: 添加 TestConnection 方法**

```go
// TestConnection sends a minimal request to verify the model API is reachable
func (m *ModelManager) TestConnection(ctx context.Context, config *model.AmModelConfig, apiKey string) error {
    adapter, err := m.getOrCreateAdapter(config, apiKey)
    if err != nil {
        return fmt.Errorf("failed to create adapter: %w", err)
    }
    return adapter.Ping(ctx)
}
```

- [ ] **Step 2: 在各 adapter 中添加 Ping 方法**

在 Claude/OpenAI/Ollama adapter 中实现 `Ping(ctx context.Context) error`。OpenAI 调用 `models.list`，Claude 发一个最小化 messages 请求，Ollama 调用 `/api/tags`。

- [ ] **Step 3: 验证编译**

```bash
cd services/ai-model-service && go build ./rpc/...
```

---

### Task 8: 前端 — 更新 API 模块

**Files:**
- Modify: `web/pc/src/api/aimodel.ts`

- [ ] **Step 1: 更新 ModelConfig 接口**

```typescript
export interface ModelConfig {
  id: string;
  config_key: string;          // 新增，不可变唯一标识
  name: string;
  display_name: string;
  provider: string;
  type: string;
  endpoint?: string;
  max_tokens: number;
  temperature: number;         // 新增
  supported_features: string;
  cost_per_1k_input_tokens: number;
  cost_per_1k_output_tokens: number;
  status: number;
  description?: string;
  created_at: string;
  updated_at: string;
}
```

- [ ] **Step 2: 更新 PromptTemplate 接口**

```typescript
export interface PromptTemplate {
  id: string;
  name: string;
  config_key: string;          // 新增，关联模型
  content: string;
  category: string;
  variables: string[];         // 新增，变量列表
  created_at: string;
  updated_at: string;
}
```

- [ ] **Step 3: 新增 API 函数**

```typescript
/** 测试模型连接 */
export function testModelConnection(data: {
  config_key: string;
  endpoint: string;
  api_key: string;
}) {
  return request.post<{ success: boolean; latency_ms: number; error?: string }>(
    '/api/v1/model/test-connection', data
  );
}

/** 测试提示词模板 */
export function testTemplate(data: {
  content: string;
  variables: Record<string, string>;
  config_key: string;
}) {
  return request.post<{
    rendered: string;
    unreplaced: string[];
    content: string;
    input_tokens: number;
    output_tokens: number;
    cost: number;
    latency: number;
  }>('/api/v1/template/test', data);
}
```

- [ ] **Step 4: 更新 createModelConfig / updateModelConfig 签名**

添加 `config_key` 和 `temperature` 参数。

- [ ] **Step 5: 更新 createTemplate / updateTemplate 签名**

添加 `config_key` 参数。

---

### Task 9: 前端 — 增强 ModelForm.vue

**Files:**
- Modify: `web/pc/src/views/aimodel/ModelForm.vue`

核心改动：

- [ ] **Step 1: 新增 config_key 表单项**

```vue
<el-form-item label="配置Key" prop="config_key">
  <el-input
    v-model="formData.config_key"
    placeholder="例如: gpt4-moderation（创建后不可修改）"
    :disabled="isEdit"
  />
  <template #extra>
    <div style="color: #909399; font-size: 12px; margin-top: 4px;">
      小写字母+数字+连字符，创建后不可修改
    </div>
  </template>
</el-form-item>
```

添加校验规则：必填 → 格式 `^[a-z0-9-]+$` → 失焦唯一性检查

- [ ] **Step 2: 新增 temperature 表单项**

数值输入，范围 0.0-2.0，步长 0.1，默认 0.7

- [ ] **Step 3: 新增"连接测试"区域**

在表单底部、保存按钮上方添加测试区域：

```vue
<el-form-item label="连接测试">
  <el-button @click="handleTestConnection" :loading="testing">
    测试连接
  </el-button>
  <span v-if="testResult" :style="{ color: testResult.success ? '#67C23A' : '#F56C6C', marginLeft: '12px' }">
    {{ testResult.success ? `✅ 连接成功，响应时间 ${testResult.latency_ms}ms` : `❌ 连接失败: ${testResult.error}` }}
  </span>
</el-form-item>
```

测试未通过时禁止保存（按钮 disabled + tooltip 提示）

- [ ] **Step 4: 添加测试连接逻辑（script setup）**

```typescript
const testing = ref(false);
const testResult = ref<{ success: boolean; latency_ms: number; error?: string } | null>(null);
const testPassed = computed(() => testResult.value?.success === true);

async function handleTestConnection() {
  testing.value = true;
  try {
    testResult.value = await testModelConnection({
      config_key: formData.config_key,
      endpoint: formData.endpoint,
      api_key: selectedApiKey.value || '',
    });
  } catch(e) {
    testResult.value = { success: false, latency_ms: 0, error: String(e) };
  } finally {
    testing.value = false;
  }
}
```

---

### Task 10: 前端 — 增强 TemplateList.vue

**Files:**
- Modify: `web/pc/src/views/aimodel/TemplateList.vue`

- [ ] **Step 1: 新增"使用模型"下拉框**

表单项：加载已启用模型列表 (`getModelConfigs({ status: 1 })`)，下拉展示 `config_key` + `display_name`

```vue
<el-form-item label="使用模型" prop="config_key">
  <el-select v-model="formData.config_key" placeholder="请选择模型配置">
    <el-option
      v-for="m in availableModels"
      :key="m.config_key"
      :label="`${m.config_key} (${m.display_name})`"
      :value="m.config_key"
    />
  </el-select>
</el-form-item>
```

- [ ] **Step 2: 新增"已识别变量"展示**

模板内容输入框下方，实时展示提取的变量：

```vue
<div class="variables-display" v-if="detectedVariables.length">
  <span style="color: #909399; font-size: 12px;">已识别变量：</span>
  <el-tag v-for="v in detectedVariables" :key="v" size="small" style="margin-left: 4px;">
    {{ v }}
  </el-tag>
</div>
```

计算属性：
```typescript
const detectedVariables = computed(() => {
  const matches = formData.content.match(/\{\{(\w+)\}\}/g) || [];
  return [...new Set(matches.map(m => m.slice(2, -2)))];
});
```

- [ ] **Step 3: 新增"即时测试"区域**

模板表单底部，变量输入 + 测试运行 + 结果展示：

```vue
<div class="test-section" v-if="detectedVariables.length && formData.config_key">
  <el-divider>即时测试</el-divider>
  <el-form-item v-for="v in detectedVariables" :key="v" :label="v">
    <el-input v-model="testVariables[v]" :placeholder="`输入 ${v} 的值`" />
  </el-form-item>
  <el-button type="success" @click="handleTestTemplate" :loading="testRunning">
    测试运行
  </el-button>
  <div v-if="testOutput" class="test-output" style="margin-top: 12px;">
    <pre>{{ JSON.stringify(testOutput, null, 2) }}</pre>
  </div>
</div>
```

- [ ] **Step 4: 保存时自动写入 variables**

```typescript
const handleSubmit = async () => {
  // ... existing validation ...
  const data = {
    ...formData,
    variables: detectedVariables.value, // 自动提取的变量列表
  };
  // ... rest of submit logic ...
};
```

---

### Task 11: 前端 — 更新 ModelList.vue

**Files:**
- Modify: `web/pc/src/views/aimodel/ModelList.vue`

- [ ] **Step 1: 表格新增 config_key 列**

在 `id` 列后添加：

```vue
<el-table-column prop="config_key" label="Config Key" min-width="150">
  <template #default="{ row }">
    <el-tag type="info" effect="plain">{{ row.config_key }}</el-tag>
  </template>
</el-table-column>
```

---

### Task 12: 启动验证

- [ ] **Step 1: 重启 ai-model-service**

```bash
# 停止旧进程
pkill -f "go run.*ai-model" 2>/dev/null

# 重新启动 RPC
cd services/ai-model-service/rpc && go run aimodel.go &>/tmp/ai-model-rpc.log &

# 等待 RPC 就绪
until ss -tlnp | grep -q 8084; do sleep 1; done

# 重新启动 API
cd services/ai-model-service/api && go run aimodelapi.go &>/tmp/ai-model-api.log &

# 等待 API 就绪
until ss -tlnp | grep -q 8891; do sleep 1; done
```

- [ ] **Step 2: 验证 API 端点**

```bash
# 验证模型列表返回 config_key
curl -s http://localhost:8891/api/v1/models | jq '.data.models[0] | {config_key, temperature}'

# 验证模板列表返回 config_key 和 variables
curl -s http://localhost:8891/api/v1/templates | jq '.data.templates[0] | {config_key, variables}'
```

- [ ] **Step 3: 验证前端页面**

```bash
# 前端应自动热更新 (Vite HMR)
# 手动验证:
# http://localhost:3003/aimodel/models/create  — 新模型表单（含 config_key + 连接测试）
# http://localhost:3003/aimodel/templates — 模板表单（含模型选择 + 变量识别 + 即时测试）
```

- [ ] **Step 4: E2E 冒烟测试（Playwright）**

使用 Playwright 访问模型配置创建页，填写 config_key，点击测试连接，验证反馈展示。

---

## 风险提示

1. **goctl 生成代码冲突**：gen 文件头部标注 DO NOT EDIT，但我们需手动添加字段。后续执行 `goctl model` 会覆盖修改。建议在 CHANGELOG 中记录此变更，考虑将来通过 goctl 的 SQL 模板重新生成。
2. **API 层逻辑体量大**：API logic 层目前是 TODO stub，需要补全 gRPC 客户端调用。
3. **Adapters Ping 方法依赖外部 API**：测试连接依赖外部真实调用，开发环境若 API Key 无效会回报失败（符合预期）。
