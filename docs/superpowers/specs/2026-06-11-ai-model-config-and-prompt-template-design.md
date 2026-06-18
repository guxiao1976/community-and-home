# AI 模型配置 & 提示词模板 — 页面设计方案

## 一、需求概要

AI 模型管理模块下新增两个增强页面：

1. **模型配置页**：引入不可变的 `config_key` 作为唯一标识，支持连接测试（保存前验证 API 可用性）
2. **提示词模板页**：模板绑定模型配置（通过 `config_key`），支持 `{{变量}}` 自动识别和即时测试

## 二、数据模型变更

### 2.1 `am_model_config` 表新增

| 字段 | 类型 | 说明 |
|------|------|------|
| `config_key` | `varchar(64)` UNIQUE NOT NULL | 不可变唯一标识，格式 `[a-z0-9-]+`，创建后锁定 |
| `temperature` | `decimal(3,2)` DEFAULT 0.7 | 模型温度参数 |

### 2.2 `am_prompt_template` 表新增

| 字段 | 类型 | 说明 |
|------|------|------|
| `config_key` | `varchar(64)` NOT NULL | 关联的模型配置 Key |
| `variables` | `text` | 从 content 自动解析的变量列表（JSON 数组），保存时自动刷新 |

## 三、模型配置页设计

### 3.1 页面结构

基于现有 `ModelForm.vue` 增强，保持两区域布局：

```
┌──────────────────────────────────────────────┐
│  ← 返回         创建模型 / 编辑模型           │
├──────────────────────────────────────────────┤
│  基本配置表单（现有字段 + 新增）               │
│  ┌──────────────────────────────────────────┐│
│  │ config_key *  [________] (创建后不可改)   ││
│  │ 显示名称 *    [________]                 ││
│  │ 提供商 *      [OpenAI ▼]                ││
│  │ 模型类型      [Chat ▼]                  ││
│  │ API端点 *     [________]                ││
│  │ 最大Token     [4096]                    ││
│  │ 温度          [0.7] (新增)              ││
│  │ 描述          [________]                ││
│  └──────────────────────────────────────────┘│
│                                              │
│  连接测试                                    │
│  ┌──────────────────────────────────────────┐│
│  │ [测试连接]  结果：✅ 成功 / ❌ 失败+原因   ││
│  └──────────────────────────────────────────┘│
│                                              │
│  API 密钥管理（仅编辑模式，同现有实现）         │
│                                              │
│       [保存]  [取消]                         │
└──────────────────────────────────────────────┘
```

### 3.2 关键交互

- **config_key**：创建时必填，格式校验 `[a-z0-9-]+`，唯一性校验（失焦时后端查重），编辑模式下 Input 置灰 + 提示"创建后不可修改"
- **连接测试**：点击后用表单的 endpoint + 选择的 API Key 发一个最小化请求。测试通过才能保存。展示：成功→绿色+延迟；失败→红色+错误信息
- **API 密钥选择**：新建模式下，密钥从现有的独立管理改为通过下拉框选择已配置的密钥（或输入新密钥并自动保存）

## 四、提示词模板页设计

### 4.1 页面结构

基于现有 `TemplateList.vue` 增强弹窗表单：

```
┌─ 新增/编辑模板 —————————————————————————─────┐
│                                              │
│  模板名称 *   [________________]             │
│  分类 *       [审核模板 ▼] (审核/通用/...)     │
│  使用模型 *   [gpt4-moderation ▼]            │
│              (下拉：已启用的模型 config_key)   │
│                                              │
│  模板内容 *   ┌──────────────────────────┐   │
│              │ 请判断用户的输入是否为...   │   │
│              │ 用户输入：{{user_input}}    │   │
│              └──────────────────────────┘   │
│                                              │
│  已识别变量   [user_input]  (自动提取+标签展示)│
│                                              │
│  ┌─ 即时测试 ─────────────────────────────┐  │
│  │  为每个变量生成输入框：                   │  │
│  │  user_input *  [____________________]  │  │
│  │  [测试运行]                             │  │
│  │  ┌─ 模型返回 ──────────────────────┐    │  │
│  │  │ (原始 JSON 或格式化展示)          │    │  │
│  │  └────────────────────────────────┘    │  │
│  └───────────────────────────────────────┘  │
│                                              │
│         [保存]  [取消]                        │
└──────────────────────────────────────────────┘
```

### 4.2 关键交互

- **"使用模型"下拉**：调用 `GET /api/v1/models?status=1` 获取已启用的模型列表，展示 `config_key` + `display_name`
- **变量自动提取**：监听模板 content 变化，正则提取 `{{(\w+)}}`，实时更新变量标签列表
- **即时测试**：
  1. 替换模板中 `{{var}}` 为实际值
  2. 调用后端 `POST /api/v1/template/:id/test` (或通用测试端点)
  3. 展示模型返回（JSON 格式化）
- **保存时**：自动将提取的变量写入 `variables` 字段

## 五、后端 API 变更

### 5.1 模型配置

| 变更 | 说明 |
|------|------|
| ModelConfig 结构体新增 `config_key`, `temperature` | 数据库 + Proto + API 类型 |
| 新增 `POST /api/v1/model/test-connection` | 连接测试端点（不保存，仅验证） |
| 创建/更新接口增加 `config_key` 校验 | 格式 + 唯一性 |

### 5.2 提示词模板

| 变更 | 说明 |
|------|------|
| PromptTemplate 结构体新增 `config_key`, `variables` | 数据库 + Proto + API 类型 |
| 保存时自动提取 `{{变量}}` → variables | 后端逻辑 |
| 新增 `POST /api/v1/template/test` | 接收 template content + variable values + config_key → 替换变量 → 调用模型 → 返回结果 |

## 六、调用流程

### 6.1 连接测试流程

```
前端 [测试连接] → POST /api/v1/model/test-connection { config_key, endpoint, api_key }
  → 后端 ModelManager 使用临时配置创建适配器
  → 发送最小化推理请求 (如 "ping", max_tokens=1)
  → 返回 { success: bool, latency_ms: int, error: string }
```

### 6.2 模板即时测试流程

```
前端 [测试运行] → POST /api/v1/template/test {
    content: "请判断...{{user_input}}",
    variables: { "user_input": "加我微信xxx" },
    config_key: "gpt4-moderation"
  }
  → 后端：用 variables 替换 content 中的 {{}}
  → 查找 config_key 对应的模型配置 + API Key
  → 调用模型适配器
  → 返回模型原始响应
```

### 6.3 运行时审核调用流程

```
moderation-service 调用时：
  → 根据业务规则选择模板 (如 moderation_template)
  → 传入变量值 { "user_input": content_to_check }
  → ai-model-service：模板渲染 + 模型调用
  → 返回审核结果
```

## 七、前端实现文件

| 文件 | 说明 |
|------|------|
| `web/pc/src/views/aimodel/ModelForm.vue` | 增强：新增 config_key、temperature、连接测试、密钥选择下拉 |
| `web/pc/src/views/aimodel/ModelList.vue` | 增强：表格列增加 config_key |
| `web/pc/src/views/aimodel/TemplateList.vue` | 增强：新增模型选择、变量识别、即时测试 |
| `web/pc/src/api/aimodel.ts` | 新增：testConnection、testTemplate 接口函数 |

## 八、后端实现文件

| 文件 | 说明 |
|------|------|
| `rpc/model/ammodelconfigmodel_gen.go` | 新增 config_key, temperature 字段 |
| `rpc/model/amprompttemplatemodel_gen.go` | 新增 config_key, variables 字段 |
| `rpc/internal/logic/` | 新增/修改 CRUD + 连接测试 + 模板测试 logic |
| `api/internal/handler/` | 新增 test-connection, template/test handler |
| `rpc/internal/manager/` | 模板变量替换 + 测试调用逻辑 |
