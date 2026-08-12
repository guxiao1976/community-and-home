package permission

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// invalidateUserCaches 失效用户的所有权限/范围缓存（T1.6 失效收敛）
//
//	DEL perm:user:{userId}（Hash 能力分层缓存）
//	SCAN-DEL perm:scopes:{userId}:*（三态读穿缓存，SCAN 限定单用户前缀，安全）
//
// 由 AssignRole/RevokeRole/UpdateUserRoleStatus/InvalidateUserCache 四处理器统一调用，
// 不依赖调用方（user-service/community-hub）记得失效。
// SEE: [[redis-cache-soft-delete]] — 失效与软删除联动，收敛到 grant 变更处理器内部
func invalidateUserCaches(ctx context.Context, rdb *redis.Client, userId int64) {
	if rdb == nil {
		return
	}
	rdb.Del(ctx, fmt.Sprintf("perm:user:%d", userId))

	// SCAN-DEL 前缀为 perm:scopes:{userId}:* 的所有 key
	pattern := fmt.Sprintf("perm:scopes:%d:*", userId)
	var cursor uint64
	for {
		keys, next, err := rdb.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return
		}
		if len(keys) > 0 {
			rdb.Del(ctx, keys...)
		}
		cursor = next
		if cursor == 0 {
			return
		}
	}
}
