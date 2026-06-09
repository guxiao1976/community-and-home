package user

import (
	"context"
	"encoding/json"
	"fmt"

	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	// rolesCacheKeyPrefix Redis 缓存 key 前缀
	rolesCacheKeyPrefix = "auth:roles:"
	// rolesCacheTTL 缓存有效期（秒），5 分钟
	rolesCacheTTL = 300
)

// rolesCacheKey 生成缓存 key
func rolesCacheKey(userId int64) string {
	return fmt.Sprintf("%s%d", rolesCacheKeyPrefix, userId)
}

// getRolesFromCache 从 Redis 读取缓存的角色数据
// 返回 nil 表示未命中
func getRolesFromCache(ctx context.Context, rds *redis.Redis, userId int64) *userv1.GetUserRolesResponse {
	if rds == nil {
		return nil
	}
	val, err := rds.GetCtx(ctx, rolesCacheKey(userId))
	if err != nil || val == "" {
		return nil
	}
	var resp userv1.GetUserRolesResponse
	if err := json.Unmarshal([]byte(val), &resp); err != nil {
		return nil
	}
	return &resp
}

// setRolesToCache 将角色数据写入 Redis 缓存
func setRolesToCache(ctx context.Context, rds *redis.Redis, userId int64, resp *userv1.GetUserRolesResponse, ttl int) {
	if rds == nil || resp == nil {
		return
	}
	data, err := json.Marshal(resp)
	if err != nil {
		return
	}
	if ttl <= 0 {
		ttl = 300
	}
	_ = rds.SetexCtx(ctx, rolesCacheKey(userId), string(data), ttl)
}

// invalidateRolesCache 主动失效指定用户的角色缓存
// 在角色状态变更时调用（审核通过/驳回、撤销、过期等）
func invalidateRolesCache(ctx context.Context, rds *redis.Redis, userId int64) {
	if rds == nil {
		return
	}
	_, _ = rds.DelCtx(ctx, rolesCacheKey(userId))
}
