package user

import (
	"context"
	"fmt"

	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-user/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

// invalidateUserPermissionCache 用户状态变更后的缓存处理
//   - 调 permission-service InvalidateUserCache 删除 perm:user / perm:scopes 缓存
//   - 禁用(status=2)时写 user:disabled:{userId} 标记，CheckPermission 据此拦截
//   - 启用(status=1)时删除禁用标记
//
// 与 permission-service 共享同一 Redis，禁用标记直接读写
func invalidateUserPermissionCache(ctx context.Context, svcCtx *svc.ServiceContext, logger logx.Logger, userId, status int64) {
	// 1. 调用 permission-service 删除权限缓存
	if svcCtx.PermissionClient != nil {
		if _, err := svcCtx.PermissionClient.InvalidateUserCache(ctx, &permissionv1.InvalidateUserCacheRequest{
			UserId: userId,
		}); err != nil {
			logger.Errorf("invalidate permission cache failed for user=%d: %v", userId, err)
		}
	}

	// 2. 写/删禁用标记（与 permission-service 共享 Redis）
	// RedisClient 可能未初始化（测试环境），需判空保护
	if svcCtx.RedisClient == nil {
		logger.Infof("RedisClient is nil, skip disabled marker for user=%d", userId)
		return
	}
	disabledKey := fmt.Sprintf("user:disabled:%d", userId)
	if status == 2 {
		// 禁用：写标记，TTL 24h（覆盖 refresh token 生命周期）
		if err := svcCtx.RedisClient.SetexCtx(ctx, disabledKey, "1", 24*3600); err != nil {
			logger.Errorf("set disabled marker failed for user=%d: %v", userId, err)
		}
		logger.Infof("user %d disabled, permission cache invalidated", userId)
	} else if status == 1 {
		// 启用：删除标记
		if _, err := svcCtx.RedisClient.DelCtx(ctx, disabledKey); err != nil {
			logger.Errorf("delete disabled marker failed for user=%d: %v", userId, err)
		}
		logger.Infof("user %d enabled, disabled marker removed", userId)
	}
}
