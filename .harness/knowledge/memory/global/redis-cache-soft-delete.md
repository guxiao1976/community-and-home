---
triggers: ["Redis cache", "缓存", "soft delete", "软删除", "unique index", "唯一索引", "stale cache", "QueryRowIndexCtx", "CachedConn", "go-zero"]
service: all
severity: should-follow
type: pitfall
status: active
created: 2026-06-12
updated: 2026-06-12
---

# 软删除 + Redis 索引缓存 = 唯一索引冲突

## 为什么会有这条经验

1. `AmModelConfig` 的删除逻辑先用软删除（设 `delete_time`），后改为硬删除。但遗留的 Redis 缓存还有旧记录的 `configKey → id` 映射。
2. go-zero 的 `QueryRowIndexCtx` 创建了二级缓存：`cache:configKey:xxx → id` 和 `cache:id:xxx → 完整记录`。
3. 硬删除时 `DELETE FROM` 只清了主缓存（`id:xxx`），索引缓存（`configKey:xxx`）残留。
4. 用户用同名创建模型时，`FindOneByConfigKey` 从 Redis 读到旧映射，报告"已存在"，但 MySQL 表实际是空的。

## 怎么做

对于**个位数记录的小表**（模型配置、API Key、模板）：
- 唯一性校验用 `QueryRowNoCacheCtx` 直接查 MySQL，不走 Redis
- 或者：删除时同时清所有索引缓存 key

```go
// ✅ 正确：直接查 MySQL，加 delete_time IS NULL 过滤
func (m *customAmModelConfigModel) FindOneByConfigKey(ctx context.Context, configKey string) (*AmModelConfig, error) {
    query := fmt.Sprintf("select %s from %s where `config_key` = ? and `delete_time` IS NULL limit 1", amModelConfigRows, m.table)
    var resp AmModelConfig
    err := m.QueryRowNoCacheCtx(ctx, &resp, query, configKey)
    // ...
}
```

对于**大表**（调用日志、统计数据）— 保留 Redis 缓存，但确保删除逻辑同时清索引缓存。

## 怎么验证

```bash
# 检查是否有残留缓存
docker exec redis redis-cli KEYS "cache:aiModelDb:amModelConfig:configKey:*"

# 创建→删除→再创建同名，不应报"已存在"
```

## 关联经验

- [[grpc-timeout-layers]]
