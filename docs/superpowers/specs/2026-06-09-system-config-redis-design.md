# 系统配置参数 Redis 缓存方案

**日期**: 2026-06-09
**状态**: 设计已确认，待实施

## 1. 背景与动机

当前 `md_configuration` 表存在以下问题：

1. **审批流程冗余**：系统级运行参数（如"每个用户最多加入 3 个小区"）由超级管理员设定后极少变更，走完整的草稿→提交→待审→通过流程毫无必要
2. **`module` 字段鸡肋**：配置时往往不清楚参数属于哪个模块，强制要求填写反而造成困扰
3. **其他服务无法读取**：配置仅通过 REST API 暴露，没有 gRPC 接口，其他微服务无法在运行时获取参数值
4. **无应用级缓存**：每次查询走 DB 或 go-zero 单行缓存，热点参数没有共享内存层

## 2. 设计目标

- 去掉审批流程，超级管理员修改即生效
- 去掉 `module` 字段，用点分隔 Key 前缀替代（隐式命名空间）
- 启动时全量加载到 Redis Hash，所有微服务通过 `common/` 工具库按 Key 单条读取
- 修改时同步更新 DB + Redis

## 3. 数据库设计

### 3.1 新表结构

```sql
CREATE TABLE md_configuration (
    id           BIGINT PRIMARY KEY,       -- Snowflake ID
    config_key   VARCHAR(128) NOT NULL,     -- 配置键，如 "user.max_community_join_count"
    config_value TEXT NOT NULL,             -- 配置值（统一字符串存储）
    value_type   VARCHAR(16) NOT NULL,      -- string | number | boolean | json
    description  VARCHAR(512),              -- 描述说明
    created_by   BIGINT,
    created_time DATETIME NOT NULL,
    updated_time DATETIME NOT NULL,
    delete_time  DATETIME,                  -- 软删除标记
    UNIQUE KEY uk_config_key (config_key)
);
```

### 3.2 与旧表对比

| 字段 | 旧表 | 新表 | 说明 |
|------|:---:|:---:|------|
| `id` | ✓ | ✓ | |
| `module` | ✓ | ✗ | 去掉，改用 Key 前缀 |
| `config_key` | ✓ | ✓ | |
| `config_value` | ✓ | ✓ | |
| `value_type` | ✓ | ✓ | |
| `description` | ✓ | ✓ | |
| `is_public` | ✓ | ✗ | 去掉，系统参数默认内部使用 |
| `approval_status` | ✓ | ✗ | 去掉 |
| `submission_status` | ✓ | ✗ | 去掉 |
| `submission_type` | ✓ | ✗ | 去掉 |
| `change_snapshot` | ✓ | ✗ | 去掉 |
| `submitter_id` | ✓ | ✗ | 去掉 |
| `submit_time` | ✓ | ✗ | 去掉 |
| `reviewer_id` | ✓ | ✗ | 去掉 |
| `review_time` | ✓ | ✗ | 去掉 |
| `review_notes` | ✓ | ✗ | 去掉 |
| `created_by` | ✓ | ✓ | |
| `created_time` | ✓ | ✓ | |
| `updated_time` | ✓ | ✓ | |
| `delete_time` | ✓ | ✓ | |

### 3.3 Key 命名规范

采用**点分隔前缀**，格式：`<domain>.<param_name>`

| 前缀 | 含义 | 示例 |
|------|------|------|
| `user.` | 用户相关 | `user.max_community_join_count` |
| `auth.` | 认证相关 | `auth.token_expire_minutes` |
| `moderation.` | 审核相关 | `moderation.sensitive_word_limit` |
| `community.` | 社区相关 | `community.max_members_per_community` |
| `file.` | 文件相关 | `file.max_upload_size_mb` |
| `general.` | 通用参数 | `general.maintenance_mode` |

前缀由超级管理员在创建参数时自由填写，不做枚举约束。

## 4. Redis 设计

### 4.1 数据结构

```
Key:   sys_config
Type:  Hash

Field → Value (JSON):
  "user.max_community_join_count" → {
      "value": "3",
      "type": "number",
      "desc": "每个用户最多加入的小区数",
      "updated_at": "2026-06-09T10:00:00Z"
  }
  "moderation.sensitive_word_limit" → {
      "value": "500",
      "type": "number",
      "desc": "敏感词库上限",
      "updated_at": "2026-06-09T10:00:00Z"
  }
```

### 4.2 为什么是单 Hash

- **一条 Redis Key**：运维方便，`HGETALL sys_config` 一目了然
- **O(1) 按 Key 读取**：`HGET sys_config <field>` 精准获取
- **内存效率**：Hash 的 ziplist 编码对小字段极省内存（参数总量预计 < 100 个）
- **原子操作**：`HSET` / `HDEL` 天然支持单字段更新

### 4.3 生命周期

| 事件 | Redis 操作 |
|------|-----------|
| master-data 启动 | `SELECT * FROM md_configuration WHERE delete_time IS NULL` → `HSET sys_config <key> <json>` 逐条写入 |
| 创建参数 | 先 `INSERT` DB → `HSET sys_config <key> <json>` |
| 更新参数 | 先 `UPDATE` DB → `HSET sys_config <key> <json>` |
| 删除参数（软删除） | 先 `UPDATE SET delete_time=NOW()` → `HDEL sys_config <key>` |
| 各服务启动 | 仅连接 Redis（`configx.MustInit`），不主动加载（已在 master-data 启动时预热） |

### 4.4 缓存一致性

策略：**DB 优先，Redis 紧跟**。不要求强一致（系统参数属于低频变更数据，最终一致可接受）。

```
写流程:
  1. BEGIN TX
  2. UPDATE md_configuration SET ... WHERE id=?
  3. COMMIT
  4. HSET sys_config <key> <new_json>   ← 事务外，失败不阻塞

读流程:
  1. HGET sys_config <key>
  2. 命中 → 返回
  3. 未命中 → gRPC GetConfig(key) 降级 → 回填 Redis → 返回
```

如果 Redis HSET 失败（极低概率），下次 `HGET` 未命中时会触发降级回填，自动修复。

## 5. common 工具库设计

### 5.1 包结构

```
common/pkg/sysconfig/
  sysconfig.go         // Client 定义、初始化、核心读取方法
  sysconfig_test.go    // 单元测试（mock Redis）
```

包路径：`github.com/guxiao1976/community-common/v2/pkg/sysconfig`

### 5.2 API 设计

```go
package sysconfig

// MustInit 初始化系统配置客户端。应在服务启动时调用，传入 Redis 配置。
// panic if Redis 连接失败。
func MustInit(c redis.RedisConf) *Client

// Client 提供系统配置读取方法
type Client struct { ... }

// Get 返回原始字符串值
func (c *Client) Get(ctx context.Context, key string) (string, error)

// GetInt 返回 int 值（value_type 必须为 number）
func (c *Client) GetInt(ctx context.Context, key string) (int, error)

// GetBool 返回 bool 值（value_type 必须为 boolean）
func (c *Client) GetBool(ctx context.Context, key string) (bool, error)

// GetJSON 将 json 类型的值反序列化到 dest
func (c *Client) GetJSON(ctx context.Context, key string, dest any) error

// GetAll 返回所有配置（调试/监控用），避免业务代码依赖此方法
func (c *Client) GetAll(ctx context.Context) (map[string]ConfigValue, error)
```

### 5.3 降级策略

```
Get(key):
  HGET sys_config <key>
    ├── 命中 → 解析 value_type 并返回
    └── 未命中 → gRPC GetConfig(key) to master-data-service
                    ├── 成功 → HSET 回填 Redis，返回
                    └── 失败 → 返回 error，调用方自行降级
```

### 5.4 gRPC 降级接口（新增）

为支持降级，需在 `api-proto/api/masterdata/v1/masterdata.proto` 中新增：

```protobuf
message GetConfigReq {
    string config_key = 1;
}

message GetConfigResp {
    string config_value = 1;
    string value_type = 2;
    string description = 3;
}

service MasterDataService {
    // ... 已有的 Division, ResidentialArea, SensitiveWord RPC ...
    rpc GetConfig(GetConfigReq) returns (GetConfigResp);
}
```

降级 gRPC 调用仅在 Redis 未命中时触发，正常路径不走 gRPC。

## 6. REST API 变更

### 6.1 变更总览

| 变更 | 说明 |
|------|------|
| ✗ 去掉审批相关端点 | `POST /configurations/:id/submit`、`POST /configurations/batch-submit` |
| ✗ 去掉审批逻辑 | `approval/` 中与 configuration 相关的 review 逻辑 |
| ✓ 保留 CRUD | `GET/POST/PUT/DELETE /configurations` |
| ✓ 简化请求体 | 去掉 `module`、`is_public` 字段 |
| ✗ 去掉审批相关表 | `md_submission_record` 中与 configuration 相关的记录 |

### 6.2 新请求/响应类型

```go
type Configuration struct {
    Id          int64  `json:"id,string"`
    ConfigKey   string `json:"config_key"`
    ConfigValue string `json:"config_value"`
    ValueType   string `json:"value_type"`
    Description string `json:"description"`
    CreatedBy   int64  `json:"created_by,string"`
    CreatedTime string `json:"created_time"`
    UpdatedTime string `json:"updated_time"`
}

type CreateConfigurationReq {
    ConfigKey   string `json:"config_key"`
    ConfigValue string `json:"config_value"`
    ValueType   string `json:"value_type"`
    Description string `json:"description"`
}

type UpdateConfigurationReq {
    Id          int64  `json:"id,string"`
    ConfigValue string `json:"config_value"`
    ValueType   string `json:"value_type"`
    Description string `json:"description"`
}
// ConfigKey 创建后不可修改
```

## 7. 前端变更（web/pc 管理后台）

### 7.1 变更文件清单

| 文件 | 变更类型 | 说明 |
|------|:---:|------|
| `web/common/types/masterdata.d.ts` | 修改 | `Configuration` 接口去掉 `module`、`is_public`、审批字段 |
| `web/pc/src/api/masterdata.ts` | 修改 | 去掉 `submitConfiguration`、`batchSubmitConfigurations`；更新请求类型 |
| `web/pc/src/views/config/List.vue` | **重写** | 去掉审批流程，简化为纯 CRUD |
| `web/pc/src/views/approval-center/Index.vue` | 修改 | 移除 `EntityType.Configuration` 相关逻辑 |
| `web/pc/src/views/deleted-recovery/Index.vue` | 修改 | 移除 configuration 实体类型（或保留软删除恢复） |
| `web/pc/src/config/modules/masterdata.config.ts` | 检查 | 路由和菜单项不变，确认权限配置 |

### 7.2 页面改造要点 — `List.vue`

**去掉的内容：**
- 表格中的 `module`、`is_public`、`submission_status`（审批状态）列
- 表单中的 `module` 输入框、`is_public` 开关
- "提交审核"按钮（单个 + 批量）
- 审批状态筛选器（草稿/待审/已通过/已驳回）

**保留并简化的内容：**
- 表格列：`config_key`、`config_value`、`value_type`、`description`、`updated_time`、操作
- 筛选：按 `config_key` 模糊搜索
- 操作按钮：新建、编辑、删除（删除即生效，无需提审）
- 新建/编辑弹窗：`config_key`（新建时可填，编辑时只读）、`config_value`、`value_type`（下拉：string/number/boolean/json）、`description`

**核心交互变化：**
```
旧流程：新建 → 草稿 → 提交审核 → 审批人通过/驳回 → 生效
新流程：新建 → 立即生效（写DB+Redis）  /  编辑 → 立即生效  /  删除 → 立即生效
```

### 7.3 审批中心 & 删除恢复变更

- **审批中心**（`approval-center/Index.vue`）：`EntityType.Configuration` 从实体类型列表中移除，不再展示配置审批记录
- **删除恢复**（`deleted-recovery/Index.vue`）：保留软删除恢复能力（误删可恢复），但去掉审批状态展示

### 7.4 前端实施建议

- 前端改造由 `web/pc` 子 Claude 实例负责，在 master-data API 改造完成后进行
- 类型定义 `web/common/types/masterdata.d.ts` 的修改需同步到 `web/mobile`（如果 mobile 也有引用）

## 8. 影响范围（更新）

| 组件 | 影响 |
|------|------|
| `api-proto/api/masterdata/v1/masterdata.proto` | 新增 `GetConfig` RPC |
| `services/master-data-service/model/` | 重新生成模型（字段精简），更新自定义方法 |
| `services/master-data-service/api/` | 简化 CRUD 逻辑，去掉审批，加入 Redis 写入 |
| `services/master-data-service/rpc/` | 新增 `GetConfig` gRPC 实现 |
| `common/pkg/sysconfig/` | **新包**，系统配置读取客户端 |
| `services/user-service/` | 引入 `sysconfig`，替换硬编码常量 |
| `services/auth-service/` | 同上 |
| 其他需要读系统参数的服务 | 同上 |
| `web/pc/src/views/config/List.vue` | **重写**，去掉审批，纯 CRUD |
| `web/pc/src/views/approval-center/Index.vue` | 移除 Configuration 实体类型 |
| `web/pc/src/api/masterdata.ts` | 精简配置相关 API 函数 |
| `web/common/types/masterdata.d.ts` | 更新 Configuration 接口定义 |

## 9. 迁移计划

1. **新建 migration SQL**：`services/master-data-service/migration/003_system_config_refactor.sql`
   - `ALTER TABLE md_configuration` 去掉旧字段，添加新约束
   - 旧审批字段先保留一段时间（可回滚），后续发版时清理
2. **Proto 变更**：在 `api-proto/` 新增 `GetConfig` RPC，运行 `make generate`
3. **common 新包**：实现 `sysconfig` 并编写测试
4. **master-data 改造**：更新模型、API 逻辑、RPC 实现
5. **前端改造**：更新 `web/common/types/` + `web/pc/` 配置管理页面、审批中心、删除恢复
6. **消费方接入**：各服务在需要处调用 `sysconfig.Get(key)` 替换硬编码

## 10. 未决事项

- [ ] 旧 `md_configuration` 表中已有数据的迁移策略（确认是否有生产数据）
