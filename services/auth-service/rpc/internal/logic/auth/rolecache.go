package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-auth/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

// roleEntry JWT roles 数组中的单个条目（精简编码）
type roleEntry struct {
	R string `json:"r"` // role_code
	C int64  `json:"c"` // community_id
}

// getUserRolesWithCache 获取用户已认证角色（Cache-Aside 模式）。
//
//  1. Redis GET auth:roles:{userId} → 命中则返回
//  2. 未命中 → gRPC GetUserRoles(userId, verf_status=2)
//  3. 写入 Redis（TTL 5 分钟）
func getUserRolesWithCache(ctx context.Context, svcCtx *svc.ServiceContext, userId int64) ([]roleEntry, error) {
	cacheKey := fmt.Sprintf("auth:roles:%d", userId)

	// 1. 尝试从 Redis 读取
	cached, err := svcCtx.RedisClient.Get(ctx, cacheKey).Result()
	if err == nil && cached != "" {
		var roles []roleEntry
		if err := json.Unmarshal([]byte(cached), &roles); err == nil {
			return roles, nil
		}
		// JSON 解析失败，忽略缓存，穿透拉取
		logx.WithContext(ctx).Infof("getUserRolesWithCache: unmarshal cache failed for userId=%d, fallback to gRPC", userId)
	}

	// 2. 缓存未命中，调用 user-service gRPC
	resp, err := svcCtx.UserServiceRpc.GetUserRoles(ctx, &userv1.GetUserRolesRequest{
		UserId:     userId,
		VerfStatus: ptrInt32(2), // 只取已认证通过的角色
	})
	if err != nil {
		return nil, fmt.Errorf("GetUserRoles gRPC failed: %w", err)
	}
	if resp == nil || resp.Base.GetCode() != 0 {
		code := int32(0)
		msg := ""
		if resp != nil && resp.Base != nil {
			code = resp.Base.GetCode()
			msg = resp.Base.GetMsg()
		}
		return nil, fmt.Errorf("GetUserRoles business error: code=%d, msg=%s", code, msg)
	}

	// 3. 转换为精简格式
	roles := make([]roleEntry, 0, len(resp.Roles))
	for _, r := range resp.Roles {
		roles = append(roles, roleEntry{
			R: r.RoleCode,
			C: r.CommunityId,
		})
	}

	// 4. 写入 Redis 缓存（TTL 可配置，默认 5 分钟）
	data, _ := json.Marshal(roles)
	ttl := 300
	if svcCtx.SysConfig != nil {
		if v, err := svcCtx.SysConfig.GetInt(ctx, "auth.cache.roles_ttl_seconds"); err == nil {
			ttl = v
		}
	}
	svcCtx.RedisClient.Set(ctx, cacheKey, string(data), time.Duration(ttl)*time.Second)

	return roles, nil
}

// ptrInt32 返回 int32 指针
func ptrInt32(v int32) *int32 {
	return &v
}
