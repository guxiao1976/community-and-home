# 系统配置参数 Redis 缓存 — 实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 md_configuration 表改造为无审批、Redis 缓存的系统参数表，并通过 common/pkg/sysconfig 工具库让所有微服务按 Key 读取。

**Architecture:** 单 Redis Hash (`sys_config`) 存储全量参数，common 提供 `sysconfig.Get(key)` 读取，master-data API 负责写入时同步 DB + Redis，Redis 未命中时通过 gRPC GetConfig 降级。

**Tech Stack:** Go 1.25, go-zero 1.10.2, Redis Hash, Proto (buf)

**Spec:** `docs/superpowers/specs/2026-06-09-system-config-redis-design.md`

---

### Task 1: Proto — 新增 GetConfig RPC

**Files:**
- Modify: `api-proto/api/masterdata/v1/masterdata.proto`
- Regenerated: `api-proto/gen/go/masterdata/v1/masterdata.pb.go` (via `make generate`)

**Dependencies:** None

- [ ] **Step 1: 编辑 proto 文件**

在 `api-proto/api/masterdata/v1/masterdata.proto` 末尾（`service MasterdataService` 的 `}` 之前），新增消息定义和 RPC：

```protobuf
// === System Configuration ===

message GetConfigReq {
    string config_key = 1;
}

message GetConfigResp {
    string config_value = 1;
    string value_type  = 2;
    string description = 3;
}
```

在 `service MasterdataService` 块内（最后的 `}` 前）新增：

```protobuf
    // GetConfig returns a system configuration value by key.
    // Used as a fallback when the Redis cache misses.
    rpc GetConfig(GetConfigReq) returns (GetConfigResp);
```

- [ ] **Step 2: 生成代码 + Lint**

```bash
cd api-proto && make generate && make lint
# Expected: 全部 PASS，无 lint 错误
```

- [ ] **Step 3: 验证生成**

```bash
grep -n "GetConfig" api-proto/gen/go/masterdata/v1/masterdata_grpc.pb.go
# 应包含 GetConfig 相关的 client/server 方法
```

- [ ] **Step 4: Commit**

```bash
git add api-proto/api/masterdata/v1/masterdata.proto api-proto/gen/
git commit -m "feat(masterdata): add GetConfig RPC for system config retrieval

Adds GetConfig RPC to MasterdataService. Used as gRPC fallback
when Redis cache misses during system config reads.

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

### Task 2: Migration SQL

**Files:**
- Create: `services/master-data-service/migration/003_system_config_refactor.sql`

**Dependencies:** None (can run in parallel with Task 1)

- [ ] **Step 1: 创建 migration 文件**

```sql
-- migration/003_system_config_refactor.sql
-- Refactor md_configuration: remove approval workflow, module column, is_public.
-- Simplify to a lightweight system parameter table.

-- Drop approval-related columns
ALTER TABLE md_configuration
    DROP COLUMN IF EXISTS approval_status,
    DROP COLUMN IF EXISTS submission_status,
    DROP COLUMN IF EXISTS submission_type,
    DROP COLUMN IF EXISTS change_snapshot,
    DROP COLUMN IF EXISTS submitter_id,
    DROP COLUMN IF EXISTS submit_time,
    DROP COLUMN IF EXISTS reviewer_id,
    DROP COLUMN IF EXISTS review_time,
    DROP COLUMN IF EXISTS review_notes,
    DROP COLUMN IF EXISTS is_public;

-- Drop old composite unique index on (module, config_key) — goctl naming pattern
ALTER TABLE md_configuration DROP INDEX IF EXISTS idx_module_config_key;

-- Drop module column
ALTER TABLE md_configuration DROP COLUMN IF EXISTS module;

-- Add unique constraint on config_key alone
ALTER TABLE md_configuration ADD UNIQUE KEY uk_config_key (config_key);

-- Resize config_key if needed
ALTER TABLE md_configuration MODIFY COLUMN config_key VARCHAR(128) NOT NULL
    COMMENT '配置键，如 user.max_community_join_count';
```

- [ ] **Step 2: Commit**

```bash
git add services/master-data-service/migration/003_system_config_refactor.sql
git commit -m "feat(master-data): add migration to refactor md_configuration table

Remove approval workflow fields, module column, is_public.
Add unique constraint on config_key alone.

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

### Task 3: common/pkg/sysconfig — 新工具库

**Files:**
- Create: `common/pkg/sysconfig/sysconfig.go`
- Create: `common/pkg/sysconfig/sysconfig_test.go`

**Dependencies:** None (does not depend on api-proto — gRPC fallback is injected via interface)

- [ ] **Step 1: 创建 sysconfig.go**

```go
// Package sysconfig reads system configuration parameters from Redis Hash.
// Each Hash field stores a JSON-encoded ConfigValue with type metadata.
package sysconfig

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

// ConfigValue is the value stored in each Hash field of the sys_config key.
type ConfigValue struct {
	Value     string `json:"value"`
	Type      string `json:"type"` // string | number | boolean | json
	Desc      string `json:"desc"`
	UpdatedAt string `json:"updated_at"`
}

// FallbackFunc is called when a key is not found in Redis.
type FallbackFunc func(ctx context.Context, key string) (*ConfigValue, error)

// Client reads system configuration from a Redis Hash.
type Client struct {
	redis    *redis.Redis
	hashKey  string
	fallback FallbackFunc
}

// MustInit creates and returns a Client. Panics if Redis is unreachable.
// hashKey is the Redis Hash name; pass "" to default to "sys_config".
// fallback is optional gRPC callback for cache-miss scenarios; pass nil to disable.
func MustInit(conf redis.RedisConf, hashKey string, fallback FallbackFunc) *Client {
	if hashKey == "" {
		hashKey = "sys_config"
	}
	return &Client{
		redis:    redis.MustNewRedis(conf),
		hashKey:  hashKey,
		fallback: fallback,
	}
}

// Get returns the raw string value for the config key.
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	cv, err := c.get(ctx, key)
	if err != nil {
		return "", err
	}
	return cv.Value, nil
}

// GetInt returns the value parsed as int. Fails if value_type is not "number".
func (c *Client) GetInt(ctx context.Context, key string) (int, error) {
	cv, err := c.get(ctx, key)
	if err != nil {
		return 0, err
	}
	if cv.Type != "number" {
		return 0, fmt.Errorf("sysconfig: key %q type=%q, expected number", key, cv.Type)
	}
	var n int
	if _, e := fmt.Sscanf(cv.Value, "%d", &n); e != nil {
		return 0, fmt.Errorf("sysconfig: key %q value %q is not a valid int: %w", key, cv.Value, e)
	}
	return n, nil
}

// GetBool returns the value parsed as bool. Fails if value_type is not "boolean".
func (c *Client) GetBool(ctx context.Context, key string) (bool, error) {
	cv, err := c.get(ctx, key)
	if err != nil {
		return false, err
	}
	if cv.Type != "boolean" {
		return false, fmt.Errorf("sysconfig: key %q type=%q, expected boolean", key, cv.Type)
	}
	switch cv.Value {
	case "true", "1":
		return true, nil
	case "false", "0":
		return false, nil
	default:
		return false, fmt.Errorf("sysconfig: key %q value %q is not a valid boolean", key, cv.Value)
	}
}

// GetJSON unmarshals a json-typed config value into dest.
func (c *Client) GetJSON(ctx context.Context, key string, dest any) error {
	cv, err := c.get(ctx, key)
	if err != nil {
		return err
	}
	if cv.Type != "json" {
		return fmt.Errorf("sysconfig: key %q type=%q, expected json", key, cv.Type)
	}
	return json.Unmarshal([]byte(cv.Value), dest)
}

// GetAll returns all configuration entries. For debugging and monitoring only.
func (c *Client) GetAll(ctx context.Context) (map[string]ConfigValue, error) {
	raw, err := c.redis.HgetallCtx(ctx, c.hashKey)
	if err != nil {
		return nil, fmt.Errorf("sysconfig: HGETALL %s: %w", c.hashKey, err)
	}
	out := make(map[string]ConfigValue, len(raw))
	for k, v := range raw {
		var cv ConfigValue
		if json.Unmarshal([]byte(v), &cv) != nil {
			continue
		}
		out[k] = cv
	}
	return out, nil
}

// --- internal ---

func (c *Client) get(ctx context.Context, key string) (*ConfigValue, error) {
	raw, err := c.redis.HgetCtx(ctx, c.hashKey, key)
	if err != nil {
		if c.fallback != nil {
			cv, fbErr := c.fallback(ctx, key)
			if fbErr != nil {
				return nil, fmt.Errorf("sysconfig: key %q: redis miss and fallback failed: %w", key, fbErr)
			}
			go c.backfill(key, cv)
			return cv, nil
		}
		return nil, fmt.Errorf("sysconfig: key %q: redis miss (no fallback configured)", key)
	}

	var cv ConfigValue
	if err := json.Unmarshal([]byte(raw), &cv); err != nil {
		return nil, fmt.Errorf("sysconfig: key %q: malformed JSON: %w", key, err)
	}
	return &cv, nil
}

func (c *Client) backfill(key string, cv *ConfigValue) {
	b, _ := json.Marshal(cv)
	bgCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = c.redis.HsetCtx(bgCtx, c.hashKey, key, string(b))
}
```

- [ ] **Step 2: 创建 sysconfig_test.go**

```go
package sysconfig

import (
	"encoding/json"
	"testing"
)

func TestConfigValueRoundTrip(t *testing.T) {
	cv := ConfigValue{
		Value:     "3",
		Type:      "number",
		Desc:      "max communities per user",
		UpdatedAt: "2026-06-09T10:00:00Z",
	}
	b, err := json.Marshal(cv)
	if err != nil {
		t.Fatal("marshal:", err)
	}
	var cv2 ConfigValue
	if err := json.Unmarshal(b, &cv2); err != nil {
		t.Fatal("unmarshal:", err)
	}
	if cv2.Value != "3" || cv2.Type != "number" || cv2.Desc != "max communities per user" {
		t.Errorf("round-trip mismatch: %+v", cv2)
	}
}

func TestConfigValueMarshalAllTypes(t *testing.T) {
	types := []string{"string", "number", "boolean", "json"}
	for _, typ := range types {
		cv := ConfigValue{Value: "test", Type: typ, Desc: "desc", UpdatedAt: "now"}
		b, err := json.Marshal(cv)
		if err != nil {
			t.Errorf("type %s marshal error: %v", typ, err)
		}
		var cv2 ConfigValue
		if err := json.Unmarshal(b, &cv2); err != nil {
			t.Errorf("type %s unmarshal error: %v", typ, err)
		}
		if cv2.Type != typ {
			t.Errorf("type %s: got %s", typ, cv2.Type)
		}
	}
}
```

- [ ] **Step 3: 验证编译和测试**

```bash
cd common && go build ./pkg/sysconfig/ && go test ./pkg/sysconfig/ -v
# Expected: go build 成功，测试 PASS
```

- [ ] **Step 4: Commit**

```bash
git add common/pkg/sysconfig/
git commit -m "feat(common): add sysconfig package for Redis-based system config reads

Client backed by Redis Hash with optional gRPC fallback.
Supports Get/GetInt/GetBool/GetJSON typed accessors.

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

### Task 4: master-data 模型更新

**Files:**
- Modify: `services/master-data-service/model/mdConfigurationModel_gen.go`
- Modify: `services/master-data-service/model/mdConfigurationModel.go`

**Dependencies:** Task 2 (migration applied to DB)

- [ ] **Step 1: 更新 MdConfiguration 结构体**

在 `mdConfigurationModel_gen.go` 中，将 `MdConfiguration struct` 替换为精简版（去掉 module、is_public、所有审批字段）：

```go
MdConfiguration struct {
	Id          int64          `db:"id"`
	ConfigKey   string         `db:"config_key"`
	ConfigValue string         `db:"config_value"`
	ValueType   string         `db:"value_type"`
	Description sql.NullString `db:"description"`
	CreatedBy   int64          `db:"created_by"`
	CreatedTime time.Time      `db:"created_time"`
	UpdatedTime time.Time      `db:"updated_time"`
	DeleteTime  sql.NullTime   `db:"delete_time"`
}
```

- [ ] **Step 2: 更新缓存 Key 前缀**

替换：
```go
cacheMdConfigurationIdPrefix              = "cache:mdConfiguration:id:"
cacheMdConfigurationModuleConfigKeyPrefix = "cache:mdConfiguration:module:configKey:"
```

为：
```go
cacheMdConfigurationIdPrefix        = "cache:mdConfiguration:id:"
cacheMdConfigurationConfigKeyPrefix = "cache:mdConfiguration:configKey:"
```

- [ ] **Step 3: 更新 generated interface**

将 `mdConfigurationModel interface` 中的：
```go
FindOneByModuleConfigKey(ctx context.Context, module string, configKey string) (*MdConfiguration, error)
```

替换为：
```go
FindOneByConfigKey(ctx context.Context, configKey string) (*MdConfiguration, error)
```

- [ ] **Step 4: 更新 Insert 方法**

完整替换 `Insert` 函数（注意参数从 19 个 ? 变为 6 个字段）：

```go
func (m *defaultMdConfigurationModel) Insert(ctx context.Context, data *MdConfiguration) (sql.Result, error) {
	mdConfigurationIdKey := fmt.Sprintf("%s%v", cacheMdConfigurationIdPrefix, data.Id)
	mdConfigurationConfigKeyKey := fmt.Sprintf("%s%v", cacheMdConfigurationConfigKeyPrefix, data.ConfigKey)
	ret, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?)", m.table, mdConfigurationRowsExpectAutoSet)
		return conn.ExecCtx(ctx, query, data.ConfigKey, data.ConfigValue, data.ValueType, data.Description, data.CreatedBy, data.DeleteTime)
		// Note: mdConfigurationRowsExpectAutoSet auto-resolves to the 6 non-auto fields:
		// config_key, config_value, value_type, description, created_by, delete_time
	}, mdConfigurationIdKey, mdConfigurationConfigKeyKey)
	return ret, err
}
```

- [ ] **Step 5: 替换 FindOneByModuleConfigKey → FindOneByConfigKey**

在 `mdConfigurationModel_gen.go` 中，将 `FindOneByModuleConfigKey` 函数完整替换为：

```go
func (m *defaultMdConfigurationModel) FindOneByConfigKey(ctx context.Context, configKey string) (*MdConfiguration, error) {
	mdConfigurationConfigKeyKey := fmt.Sprintf("%s%v", cacheMdConfigurationConfigKeyPrefix, configKey)
	var resp MdConfiguration
	err := m.QueryRowIndexCtx(ctx, &resp, mdConfigurationConfigKeyKey, m.formatPrimary, func(ctx context.Context, conn sqlx.SqlConn, v any) (i any, e error) {
		query := fmt.Sprintf("select %s from %s where `config_key` = ? and `delete_time` is null limit 1", mdConfigurationRows, m.table)
		if err := conn.QueryRowCtx(ctx, &resp, query, configKey); err != nil {
			return nil, err
		}
		return resp.Id, nil
	}, m.queryPrimary)
	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}
```

- [ ] **Step 6: 更新 Delete 方法**

替换 `Delete` 函数（去掉 module 依赖）：

```go
func (m *defaultMdConfigurationModel) Delete(ctx context.Context, id int64) error {
	data, err := m.FindOne(ctx, id)
	if err != nil {
		return err
	}

	mdConfigurationIdKey := fmt.Sprintf("%s%v", cacheMdConfigurationIdPrefix, id)
	mdConfigurationConfigKeyKey := fmt.Sprintf("%s%v", cacheMdConfigurationConfigKeyPrefix, data.ConfigKey)
	_, err = m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("delete from %s where `id` = ?", m.table)
		return conn.ExecCtx(ctx, query, id)
	}, mdConfigurationIdKey, mdConfigurationConfigKeyKey)
	return err
}
```

- [ ] **Step 7: 更新 Update 方法**

替换 `Update` 函数（参数从 19 个改为 8 个字段 + 1 个 WHERE id）：

```go
func (m *defaultMdConfigurationModel) Update(ctx context.Context, newData *MdConfiguration) error {
	data, err := m.FindOne(ctx, newData.Id)
	if err != nil {
		return err
	}

	mdConfigurationIdKey := fmt.Sprintf("%s%v", cacheMdConfigurationIdPrefix, data.Id)
	mdConfigurationConfigKeyKey := fmt.Sprintf("%s%v", cacheMdConfigurationConfigKeyPrefix, data.ConfigKey)
	_, err = m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("update %s set %s where `id` = ?", m.table, mdConfigurationRowsWithPlaceHolder)
		// mdConfigurationRowsWithPlaceHolder has 6 placeholders (excludes id, created_time, updated_time)
		// Total: 6 SET + 1 WHERE = 7 args
		return conn.ExecCtx(ctx, query, newData.ConfigKey, newData.ConfigValue, newData.ValueType, newData.Description, newData.CreatedBy, newData.DeleteTime, newData.Id)
	}, mdConfigurationIdKey, mdConfigurationConfigKeyKey)
	return err
}
```

- [ ] **Step 8: 更新自定义模型 (mdConfigurationModel.go)**

完整替换 `mdConfigurationModel.go` 内容：

```go
package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ MdConfigurationModel = (*customMdConfigurationModel)(nil)

type (
	MdConfigurationModel interface {
		mdConfigurationModel
		FindAllActive(ctx context.Context) ([]*MdConfiguration, error)
		FindByConfigKey(ctx context.Context, configKey string) (*MdConfiguration, error)
		FindByKeyPrefix(ctx context.Context, keyword string, limit, offset int) ([]*MdConfiguration, int64, error)
		SoftDelete(ctx context.Context, id int64) error
		CountDeleted(ctx context.Context) (int64, error)
		FindDeleted(ctx context.Context, page, pageSize int64) ([]*MdConfiguration, int64, error)
		Restore(ctx context.Context, id int64) error
	}

	customMdConfigurationModel struct {
		*defaultMdConfigurationModel
	}
)

func NewMdConfigurationModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) MdConfigurationModel {
	return &customMdConfigurationModel{
		defaultMdConfigurationModel: newMdConfigurationModel(conn, c, opts...),
	}
}

// FindAllActive returns all non-deleted configurations. Used for Redis cache warming at startup.
func (m *customMdConfigurationModel) FindAllActive(ctx context.Context) ([]*MdConfiguration, error) {
	var resp []*MdConfiguration
	query := fmt.Sprintf("select %s from %s where delete_time is null order by id", mdConfigurationRows, m.table)
	err := m.QueryRowsNoCacheCtx(ctx, &resp, query)
	switch err {
	case nil:
		return resp, nil
	case sqlx.ErrNotFound:
		return nil, nil
	default:
		return nil, err
	}
}

// FindByConfigKey finds a single non-deleted config by its key.
func (m *customMdConfigurationModel) FindByConfigKey(ctx context.Context, configKey string) (*MdConfiguration, error) {
	var resp MdConfiguration
	query := fmt.Sprintf("select %s from %s where config_key = ? and delete_time is null limit 1", mdConfigurationRows, m.table)
	err := m.QueryRowNoCacheCtx(ctx, &resp, query, configKey)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// FindByKeyPrefix supports keyword search on config_key for the admin list page.
func (m *customMdConfigurationModel) FindByKeyPrefix(ctx context.Context, keyword string, limit, offset int) ([]*MdConfiguration, int64, error) {
	var resp []*MdConfiguration
	var total int64

	baseWhere := "delete_time is null"
	var countArgs []interface{}
	var listArgs []interface{}

	if keyword != "" {
		baseWhere += " and config_key like ?"
		kw := "%" + keyword + "%"
		countArgs = append(countArgs, kw)
		listArgs = append(listArgs, kw, limit, offset)
	} else {
		listArgs = append(listArgs, limit, offset)
	}

	countQuery := fmt.Sprintf("select count(*) from %s where %s", m.table, baseWhere)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countQuery, countArgs...); err != nil {
		return nil, 0, err
	}

	listQuery := fmt.Sprintf("select %s from %s where %s order by id desc limit ? offset ?", mdConfigurationRows, m.table, baseWhere)
	err := m.QueryRowsNoCacheCtx(ctx, &resp, listQuery, listArgs...)
	switch err {
	case nil:
		return resp, total, nil
	case sqlx.ErrNotFound:
		return nil, 0, nil
	default:
		return nil, 0, err
	}
}

// SoftDelete sets delete_time = now() and evicts cache keys.
func (m *customMdConfigurationModel) SoftDelete(ctx context.Context, id int64) error {
	data, err := m.FindOne(ctx, id)
	if err != nil {
		return err
	}

	mdConfigurationIdKey := fmt.Sprintf("%s%v", cacheMdConfigurationIdPrefix, id)
	mdConfigurationConfigKeyKey := fmt.Sprintf("%s%v", cacheMdConfigurationConfigKeyPrefix, data.ConfigKey)
	_, err = m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("update %s set delete_time = now() where id = ?", m.table)
		return conn.ExecCtx(ctx, query, id)
	}, mdConfigurationIdKey, mdConfigurationConfigKeyKey)
	return err
}

// CountDeleted counts soft-deleted configurations.
func (m *customMdConfigurationModel) CountDeleted(ctx context.Context) (int64, error) {
	var count int64
	query := fmt.Sprintf("select count(*) from %s where delete_time is not null", m.table)
	err := m.QueryRowNoCacheCtx(ctx, &count, query)
	return count, err
}

// FindDeleted returns paginated soft-deleted configurations.
func (m *customMdConfigurationModel) FindDeleted(ctx context.Context, page, pageSize int64) ([]*MdConfiguration, int64, error) {
	var total int64
	countQuery := fmt.Sprintf("select count(*) from %s where delete_time is not null", m.table)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countQuery); err != nil {
		return nil, 0, err
	}

	var configs []*MdConfiguration
	offset := (page - 1) * pageSize
	query := fmt.Sprintf("select %s from %s where delete_time is not null order by delete_time desc limit ? offset ?", mdConfigurationRows, m.table)
	err := m.QueryRowsNoCacheCtx(ctx, &configs, query, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	return configs, total, nil
}

// Restore clears delete_time, making the record active again.
func (m *customMdConfigurationModel) Restore(ctx context.Context, id int64) error {
	mdConfigurationIdKey := fmt.Sprintf("%s%v", cacheMdConfigurationIdPrefix, id)
	_, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		query := fmt.Sprintf("update %s set delete_time = null where id = ?", m.table)
		return conn.ExecCtx(ctx, query, id)
	}, mdConfigurationIdKey)
	return err
}
```

- [ ] **Step 9: 验证编译**

```bash
cd services/master-data-service && go build ./model/...
# Expected: 编译成功
```

- [ ] **Step 10: Commit**

```bash
git add services/master-data-service/model/mdConfigurationModel_gen.go \
        services/master-data-service/model/mdConfigurationModel.go
git commit -m "refactor(master-data): simplify MdConfiguration model

Remove approval workflow fields, module, is_public from struct.
Add FindAllActive for startup cache warming.
Replace FindByModule with FindByKeyPrefix for admin list page.

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

### Task 5: master-data API 类型更新

**Files:**
- Modify: `services/master-data-service/api/internal/types/types.go`

**Dependencies:** None (types only, no logic)

- [ ] **Step 1: 更新 Configuration 响应类型**

替换 `Configuration` struct（行 50-62 附近）：

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
```

- [ ] **Step 2: 更新请求类型**

替换 `CreateConfigurationReq`：

```go
type CreateConfigurationReq struct {
	ConfigKey   string `json:"config_key"`
	ConfigValue string `json:"config_value"`
	ValueType   string `json:"value_type"`
	Description string `json:"description,optional"`
}
```

替换 `CreateConfigurationResp`：

```go
type CreateConfigurationResp struct {
	Id int64 `json:"id,string"`
}
```

替换 `UpdateConfigurationReq`（去掉 module、is_public）：

```go
type UpdateConfigurationReq struct {
	Id          int64  `json:"id,string"`
	ConfigValue string `json:"config_value"`
	ValueType   string `json:"value_type"`
	Description string `json:"description,optional"`
}
```

替换 `UpdateConfigurationResp`：

```go
type UpdateConfigurationResp struct {
	Success bool `json:"success"`
}
```

替换 `GetConfigurationsReq`（去掉 Module，改用 Keyword 搜索 config_key）：

```go
type GetConfigurationsReq struct {
	Keyword  string `form:"keyword,optional"`
	Page     int64  `form:"page,optional"`
	PageSize int64  `form:"page_size,optional"`
}
```

- [ ] **Step 3: 删除审批相关类型**

删除以下类型定义：
- `SubmitReq`
- `SubmitResp`
- `BatchSubmitReq`
- `BatchSubmitResp`
- `ReviewConfigurationReq`
- `GetPendingItemsReq`（如果有 configuration 相关字段）

- [ ] **Step 4: 删除旧的 CreateConfigurationReq（如有重复定义）**

确保没有残留的旧 `CreateConfigurationReq`（含 Module/IsPublic 字段的版本）。

- [ ] **Step 5: 验证编译**

```bash
cd services/master-data-service && go build ./api/internal/types/
```

- [ ] **Step 6: Commit**

```bash
git add services/master-data-service/api/internal/types/types.go
git commit -m "refactor(master-data): update Configuration API types

Remove module, is_public, approval fields from types.
Simplify list request to keyword search on config_key.

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

### Task 6: master-data API 逻辑层重写

**Files:**
- Rewrite: `services/master-data-service/api/internal/logic/configuration/createConfigurationLogic.go`
- Rewrite: `services/master-data-service/api/internal/logic/configuration/getConfigurationLogic.go`
- Rewrite: `services/master-data-service/api/internal/logic/configuration/getConfigurationsLogic.go`
- Rewrite: `services/master-data-service/api/internal/logic/configuration/updateConfigurationLogic.go`
- Rewrite: `services/master-data-service/api/internal/logic/configuration/deleteConfigurationLogic.go`
- Delete: `services/master-data-service/api/internal/logic/configuration/submitConfigurationLogic.go`
- Delete: `services/master-data-service/api/internal/logic/configuration/batchSubmitConfigurationsLogic.go`
- Modify: `services/master-data-service/api/internal/svc/serviceContext.go`

**Dependencies:** Task 4 (model), Task 5 (types)

- [ ] **Step 1: 创建 createConfigurationLogic.go**

```go
package configuration

import (
	"context"
	"time"

	"github.com/guxiao1976/community-common/v2/pkg/snowflake"
	"github.com/guxiao1976/community-master-data-service/api/internal/svc"
	"github.com/guxiao1976/community-master-data-service/api/internal/types"
	"github.com/guxiao1976/community-master-data-service/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateConfigurationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateConfigurationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateConfigurationLogic {
	return &CreateConfigurationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateConfigurationLogic) CreateConfiguration(req *types.CreateConfigurationReq) (resp *types.CreateConfigurationResp, err error) {
	// Validate value_type
	switch req.ValueType {
	case "string", "number", "boolean", "json":
	default:
		return nil, svc.ErrInvalidValueType
	}

	// Check duplicate key
	existing, _ := l.svcCtx.MdConfigurationModel.FindByConfigKey(l.ctx, req.ConfigKey)
	if existing != nil {
		return nil, svc.ErrConfigKeyExists
	}

	now := time.Now()
	userId := l.svcCtx.UserIdFromContext(l.ctx)

	cfg := &model.MdConfiguration{
		Id:          snowflake.Generate(),
		ConfigKey:   req.ConfigKey,
		ConfigValue: req.ConfigValue,
		ValueType:   req.ValueType,
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
		CreatedBy:   userId,
		CreatedTime: now,
		UpdatedTime: now,
	}

	if _, err := l.svcCtx.MdConfigurationModel.Insert(l.ctx, cfg); err != nil {
		return nil, err
	}

	// Sync to Redis
	if err := l.svcCtx.SyncConfigToRedis(l.ctx, cfg); err != nil {
		logx.WithContext(l.ctx).Errorf("failed to sync config %s to Redis: %v", cfg.ConfigKey, err)
		// Non-fatal: Redis will be backfilled on next read
	}

	return &types.CreateConfigurationResp{Id: cfg.Id}, nil
}
```

- [ ] **Step 2: 实现 getConfigurationLogic.go**

```go
package configuration

import (
	"context"

	"github.com/guxiao1976/community-master-data-service/api/internal/svc"
	"github.com/guxiao1976/community-master-data-service/api/internal/types"
	"github.com/guxiao1976/community-master-data-service/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetConfigurationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetConfigurationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetConfigurationLogic {
	return &GetConfigurationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetConfigurationLogic) GetConfiguration(req *types.GetConfigurationReq) (resp *types.GetConfigurationResp, err error) {
	cfg, err := l.svcCtx.MdConfigurationModel.FindOne(l.ctx, req.Id)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, svc.ErrConfigNotFound
		}
		return nil, err
	}

	return &types.GetConfigurationResp{
		Configuration: toConfiguration(cfg),
	}, nil
}

func toConfiguration(cfg *model.MdConfiguration) types.Configuration {
	var desc string
	if cfg.Description.Valid {
		desc = cfg.Description.String
	}
	return types.Configuration{
		Id:          cfg.Id,
		ConfigKey:   cfg.ConfigKey,
		ConfigValue: cfg.ConfigValue,
		ValueType:   cfg.ValueType,
		Description: desc,
		CreatedBy:   cfg.CreatedBy,
		CreatedTime: cfg.CreatedTime.Format("2006-01-02 15:04:05"),
		UpdatedTime: cfg.UpdatedTime.Format("2006-01-02 15:04:05"),
	}
}
```

- [ ] **Step 3: 实现 getConfigurationsLogic.go**

```go
package configuration

import (
	"context"

	"github.com/guxiao1976/community-master-data-service/api/internal/svc"
	"github.com/guxiao1976/community-master-data-service/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetConfigurationsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetConfigurationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetConfigurationsLogic {
	return &GetConfigurationsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetConfigurationsLogic) GetConfigurations(req *types.GetConfigurationsReq) (resp *types.GetConfigurationsResp, err error) {
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	configs, total, err := l.svcCtx.MdConfigurationModel.FindByKeyPrefix(l.ctx, req.Keyword, pageSize, offset)
	if err != nil {
		return nil, err
	}

	list := make([]types.Configuration, 0, len(configs))
	for _, cfg := range configs {
		list = append(list, toConfiguration(cfg))
	}

	return &types.GetConfigurationsResp{
		List:  list,
		Total: total,
	}, nil
}
```

- [ ] **Step 4: 实现 updateConfigurationLogic.go**

```go
package configuration

import (
	"context"
	"database/sql"
	"time"

	"github.com/guxiao1976/community-master-data-service/api/internal/svc"
	"github.com/guxiao1976/community-master-data-service/api/internal/types"
	"github.com/guxiao1976/community-master-data-service/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateConfigurationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateConfigurationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateConfigurationLogic {
	return &UpdateConfigurationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateConfigurationLogic) UpdateConfiguration(req *types.UpdateConfigurationReq) (resp *types.UpdateConfigurationResp, err error) {
	// Validate value_type
	switch req.ValueType {
	case "string", "number", "boolean", "json":
	default:
		return nil, svc.ErrInvalidValueType
	}

	cfg, err := l.svcCtx.MdConfigurationModel.FindOne(l.ctx, req.Id)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, svc.ErrConfigNotFound
		}
		return nil, err
	}

	cfg.ConfigValue = req.ConfigValue
	cfg.ValueType = req.ValueType
	cfg.Description = sql.NullString{String: req.Description, Valid: req.Description != ""}
	cfg.UpdatedTime = time.Now()

	if err := l.svcCtx.MdConfigurationModel.Update(l.ctx, cfg); err != nil {
		return nil, err
	}

	// Sync to Redis
	if err := l.svcCtx.SyncConfigToRedis(l.ctx, cfg); err != nil {
		logx.WithContext(l.ctx).Errorf("failed to sync config %s to Redis: %v", cfg.ConfigKey, err)
	}

	return &types.UpdateConfigurationResp{Success: true}, nil
}
```

- [ ] **Step 5: 实现 deleteConfigurationLogic.go**

```go
package configuration

import (
	"context"

	"github.com/guxiao1976/community-master-data-service/api/internal/svc"
	"github.com/guxiao1976/community-master-data-service/api/internal/types"
	"github.com/guxiao1976/community-master-data-service/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteConfigurationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteConfigurationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteConfigurationLogic {
	return &DeleteConfigurationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteConfigurationLogic) DeleteConfiguration(req *types.DeleteConfigurationReq) (resp *types.DeleteConfigurationResp, err error) {
	// Fetch first to get config_key for Redis cleanup
	cfg, err := l.svcCtx.MdConfigurationModel.FindOne(l.ctx, req.Id)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, svc.ErrConfigNotFound
		}
		return nil, err
	}

	if err := l.svcCtx.MdConfigurationModel.SoftDelete(l.ctx, req.Id); err != nil {
		return nil, err
	}

	// Remove from Redis
	if err := l.svcCtx.RemoveConfigFromRedis(l.ctx, cfg.ConfigKey); err != nil {
		logx.WithContext(l.ctx).Errorf("failed to remove config %s from Redis: %v", cfg.ConfigKey, err)
	}

	return &types.DeleteConfigurationResp{Success: true}, nil
}
```

- [ ] **Step 6: 删除审批相关 logic 文件**

```bash
rm services/master-data-service/api/internal/logic/configuration/submitConfigurationLogic.go
rm services/master-data-service/api/internal/logic/configuration/batchSubmitConfigurationsLogic.go
```

- [ ] **Step 7: 实现 helper 函数（共用）**

创建或复用现有 helper，将 `svc.ErrInvalidValueType`、`svc.ErrConfigKeyExists`、`svc.ErrConfigNotFound` 等错误定义添加到合适的包中。

在 `api/internal/svc/` 中新建或扩展现有 errors 文件：

```go
package svc

import "github.com/guxiao1976/community-common/v2/pkg/errx"

var (
	ErrInvalidValueType = errx.NewCodeError(99400, "invalid value_type, must be string/number/boolean/json")
	ErrConfigKeyExists  = errx.NewCodeError(99400, "config_key already exists")
	ErrConfigNotFound   = errx.NewCodeError(99404, "configuration not found")
)
```

- [ ] **Step 8: 验证编译**

```bash
cd services/master-data-service && go build ./api/...
# Expected: 编译成功（可能需要解决一些 import 问题）
```

- [ ] **Step 9: Commit**

```bash
git add services/master-data-service/api/internal/logic/configuration/ \
        services/master-data-service/api/internal/svc/
git commit -m "refactor(master-data): rewrite configuration logic sans approval

Replace approval workflow CRUD with immediate-effect CRUD.
Each create/update/delete syncs to Redis Hash sys_config.
Remove submit and batch-submit logic files.

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

### Task 7: master-data API 路由更新 & ServiceContext

**Files:**
- Modify: `services/master-data-service/api/internal/handler/routes.go`
- Modify: `services/master-data-service/api/internal/svc/serviceContext.go`

**Dependencies:** Task 6

- [ ] **Step 1: routes.go — 删除审批路由**

在 `routes.go` 中找到 configuration 路由组，删除 submit 和 batch-submit 两行：

```go
// 删除这两行:
// POST /configurations/:id/submit
// POST /configurations/batch-submit

// 保留:
rest.Get("/configurations", ...)       // 列表
rest.Post("/configurations", ...)      // 新建
rest.Get("/configurations/:id", ...)   // 详情
rest.Put("/configurations/:id", ...)   // 更新
rest.Delete("/configurations/:id", ...)// 删除
```

同时删除对应的 handler 变量引用（如果有 `submitConfigurationHandler`、`batchSubmitConfigurationsHandler` 变量）。

- [ ] **Step 2: serviceContext.go — 添加 Redis 支持和配置同步方法**

在 `ServiceContext` 结构体中新增 Redis 客户端：

```go
import (
	// ... existing imports ...
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type ServiceContext struct {
	// ... existing fields ...
	RedisClient *redis.Redis  // 新增
}
```

在 `NewServiceContext` 中初始化 Redis：

```go
func NewServiceContext(c config.Config) *ServiceContext {
	// ... existing code ...
	ctx.RedisClient = redis.MustNewRedis(c.Cache[0].RedisConf) // 使用 go-zero cache 配置中的 Redis
	// ... rest ...
	return ctx
}
```

添加 Redis 同步辅助方法：

```go
import (
	"encoding/json"
	"fmt"
	"time"

	sysconfig "github.com/guxiao1976/community-common/v2/pkg/sysconfig"
)

const configHashKey = "sys_config"

// SyncConfigToRedis writes a configuration to the Redis Hash.
func (sc *ServiceContext) SyncConfigToRedis(ctx context.Context, cfg *model.MdConfiguration) error {
	desc := ""
	if cfg.Description.Valid {
		desc = cfg.Description.String
	}
	cv := sysconfig.ConfigValue{
		Value:     cfg.ConfigValue,
		Type:      cfg.ValueType,
		Desc:      desc,
		UpdatedAt: time.Now().Format(time.RFC3339),
	}
	payload, err := json.Marshal(cv)
	if err != nil {
		return fmt.Errorf("marshal ConfigValue: %w", err)
	}
	return sc.RedisClient.HsetCtx(ctx, configHashKey, cfg.ConfigKey, string(payload))
}

// RemoveConfigFromRedis removes a configuration from the Redis Hash.
func (sc *ServiceContext) RemoveConfigFromRedis(ctx context.Context, configKey string) error {
	_, err := sc.RedisClient.HdelCtx(ctx, configHashKey, configKey)
	return err
}

// WarmupConfigCache loads all active configs from DB into Redis. Call at startup.
func (sc *ServiceContext) WarmupConfigCache(ctx context.Context) error {
	configs, err := sc.MdConfigurationModel.FindAllActive(ctx)
	if err != nil {
		return fmt.Errorf("FindAllActive: %w", err)
	}

	for _, cfg := range configs {
		if err := sc.SyncConfigToRedis(ctx, cfg); err != nil {
			return fmt.Errorf("sync %s: %w", cfg.ConfigKey, err)
		}
	}
	logx.Infof("Warmed up %d config entries to Redis", len(configs))
	return nil
}
```

- [ ] **Step 3: 添加启动时的缓存预热**

在主入口文件中（或通过 go-zero 的 starter），在服务启动后调用 `WarmupConfigCache`。

- [ ] **Step 4: 验证编译**

```bash
cd services/master-data-service && go build ./api/...
```

- [ ] **Step 5: Commit**

```bash
git add services/master-data-service/api/internal/handler/routes.go \
        services/master-data-service/api/internal/svc/serviceContext.go
git commit -m "feat(master-data): add Redis sync on config write + startup warmup

Wire RedisClient into API ServiceContext. Add SyncConfigToRedis,
RemoveConfigFromRedis, and WarmupConfigCache methods. Remove
submit/batch-submit routes.

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

### Task 8: master-data RPC — GetConfig 实现

**Files:**
- Create: `services/master-data-service/rpc/internal/logic/configuration/getConfigLogic.go` (new directory)
- Modify: `services/master-data-service/rpc/internal/svc/servicecontext.go`

**Dependencies:** Task 1 (proto generated), Task 4 (model)

- [ ] **Step 1: RPC ServiceContext 添加 MdConfigurationModel**

在 `rpc/internal/svc/servicecontext.go` 中：

```go
type ServiceContext struct {
	Config                        config.Config
	MdAdministrativeDivisionModel model.MdAdministrativeDivisionModel
	MdResidentialAreaModel        model.MdResidentialAreaModel
	MdSensitiveWordModel          model.MdSensitiveWordModel
	MdConfigurationModel          model.MdConfigurationModel  // 新增
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DataSource)
	cacheConf := c.Cache
	opts := cache.WithExpiry(3600)

	return &ServiceContext{
		Config:                        c,
		MdAdministrativeDivisionModel: model.NewMdAdministrativeDivisionModel(conn, cacheConf, opts),
		MdResidentialAreaModel:        model.NewMdResidentialAreaModel(conn, cacheConf, opts),
		MdSensitiveWordModel:          model.NewMdSensitiveWordModel(conn, cacheConf, opts),
		MdConfigurationModel:          model.NewMdConfigurationModel(conn, cacheConf, opts),  // 新增
	}
}
```

- [ ] **Step 2: 实现 GetConfig RPC logic**

创建目录和文件：

```bash
mkdir -p services/master-data-service/rpc/internal/logic/configuration
```

创建 `services/master-data-service/rpc/internal/logic/configuration/getConfigLogic.go`：

```go
package configuration

import (
	"context"
	"database/sql"

	"github.com/guxiao1976/community-master-data-service/model"
	"github.com/guxiao1976/community-master-data-service/rpc/internal/svc"
	masterdatav1 "github.com/guxiao1976/api-proto/gen/go/masterdata/v1"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetConfigLogic {
	return &GetConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetConfigLogic) GetConfig(in *masterdatav1.GetConfigReq) (*masterdatav1.GetConfigResp, error) {
	cfg, err := l.svcCtx.MdConfigurationModel.FindByConfigKey(l.ctx, in.ConfigKey)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, err
		}
		return nil, err
	}

	desc := ""
	if cfg.Description.Valid {
		desc = cfg.Description.String
	}

	return &masterdatav1.GetConfigResp{
		ConfigValue: cfg.ConfigValue,
		ValueType:   cfg.ValueType,
		Description: desc,
	}, nil
}
```

- [ ] **Step 3: 验证编译**

```bash
cd services/master-data-service && go build ./rpc/...
# Expected: 编译通过
```

- [ ] **Step 4: Commit**

```bash
git add services/master-data-service/rpc/internal/logic/configuration/ \
        services/master-data-service/rpc/internal/svc/servicecontext.go
git commit -m "feat(master-data): implement GetConfig gRPC handler

Add MdConfigurationModel to RPC ServiceContext.
Implement GetConfig RPC that reads config by key from DB.

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

### Task 9: 运行机械化检查

**Dependencies:** Tasks 1-8

- [ ] **Step 1: 运行 harness 检查**

```bash
bash .harness/skills/qa/scripts/harness-checks.sh --service master-data-service
# Expected: 全部 PASS，无 FAIL
```

- [ ] **Step 2: 如果有 FAIL，修复后重新运行**

---

### Task 10: 前端变更（web/pc）

**Files:**
- Modify: `web/common/types/masterdata.d.ts` — 更新 `Configuration` 接口
- Modify: `web/pc/src/api/masterdata.ts` — 去掉 submit/batch-submit API 函数
- Rewrite: `web/pc/src/views/config/List.vue` — 纯 CRUD，无审批流程
- Modify: `web/pc/src/views/approval-center/Index.vue` — 移除 `EntityType.Configuration`
- Modify: `web/pc/src/views/deleted-recovery/Index.vue` — 移除 configuration 实体类型

**实施者**: 前端 (`web/pc`) 子 Claude 实例

- [ ] **Step 1: 更新类型定义** — 见设计文档 7.2 节
- [ ] **Step 2: 精简 API 函数** — 去掉 `submitConfiguration`、`batchSubmitConfigurations`
- [ ] **Step 3: 重写 List.vue** — 去掉 module/is_public/审批状态列、提交审核按钮；保留新建/编辑/删除
- [ ] **Step 4: 清理审批中心** — 移除 Configuration 实体类型
- [ ] **Step 5: 清理删除恢复** — 移除 configuration 实体类型（或保留软删除恢复）
- [ ] **Step 6: 运行前端 lint/type-check**

```bash
cd web/pc && npm run lint && npm run build
# Expected: 无 type error，build 成功
```

- [ ] **Step 7: Commit**

```bash
git add web/common/types/masterdata.d.ts web/pc/src/
git commit -m "refactor(web): simplify config management page for new system params

Remove approval workflow UI. Show only config_key, value, type, desc, timestamps.
Direct CRUD without submit/review flow.

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

### Task 11: 消费方接入（示例：user-service）

**Files:**
- Modify: `services/user-service/api/internal/svc/serviceContext.go` — 添加 sysconfig 客户端初始化
- Modify: `services/user-service/api/internal/config/config.go` — 添加 Redis 配置
- Modify: 使用硬编码常量的 logic 文件 — 替换为 `sysconfig.Get(key)`

**Dependencies:** Task 3 (sysconfig package), Task 8 (GetConfig RPC available)

- [ ] **Step 1: 添加 Redis 配置到 Config struct**

```go
// services/user-service/api/internal/config/config.go
type Config struct {
	// ... existing fields ...
	SysConfigRedis redis.RedisConf  // 已存在类似字段则复用
}
```

- [ ] **Step 2: 初始化 sysconfig.Client**

在 ServiceContext 的 `NewServiceContext` 中：

```go
import (
	sysconfig "github.com/guxiao1976/community-common/v2/pkg/sysconfig"
	masterdatav1 "github.com/guxiao1976/api-proto/gen/go/masterdata/v1"
)

// 在 NewServiceContext 函数内:
sc.SysConfig = sysconfig.MustInit(c.SysConfigRedis, "", func(ctx context.Context, key string) (*sysconfig.ConfigValue, error) {
    // gRPC fallback to master-data
    client := masterdatav1.NewMasterdataServiceClient(grpcClient) // 需要已有的 masterdata gRPC 连接
    resp, err := client.GetConfig(ctx, &masterdatav1.GetConfigReq{ConfigKey: key})
    if err != nil {
        return nil, err
    }
    return &sysconfig.ConfigValue{
        Value: resp.ConfigValue,
        Type:  resp.ValueType,
        Desc:  resp.Description,
    }, nil
})
```

- [ ] **Step 3: 替换硬编码常量为 sysconfig 调用**

```go
// 旧:
const MaxCommunityJoinCount = 3

// 新:
count, err := l.svcCtx.SysConfig.GetInt(l.ctx, "user.max_community_join_count")
if err != nil {
    count = 3 // 默认值
}
```

- [ ] **Step 4: 运行测试确保无回归**

```bash
cd services/user-service && go test ./...
```

- [ ] **Step 5: Commit** — 各服务独立提交

每个服务独立引入 sysconfig 时单独 commit。

---

### 实施顺序 & 依赖图

```
Task 1 (Proto) ───────────────────────┐
Task 2 (Migration) ──────────────────┤
Task 3 (sysconfig pkg) ──────────────┤
                                      ├── Task 4 (Model) ── Task 5 (Types) ── Task 6 (Logic) ── Task 7 (Routes+SvcCtx) ── Task 8 (RPC) ── Task 9 (Harness) ── Task 10 (Frontend)
                                      │                                                                                         │
                                      └───────────────────────────────────────────────────────────────────────────────── Task 11 (Consumer)
```
