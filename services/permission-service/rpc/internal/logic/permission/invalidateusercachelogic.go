package permission

import (
	"context"
	"fmt"

	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-permission/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type InvalidateUserCacheLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewInvalidateUserCacheLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InvalidateUserCacheLogic {
	return &InvalidateUserCacheLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

// InvalidateUserCache 删除用户的所有权限相关缓存
// 触发场景：用户禁用/启用、用户状态变更
// key 前缀: perm:user:{userId}, perm:scopes:{userId}:{scopeType}
func (l *InvalidateUserCacheLogic) InvalidateUserCache(in *permissionv1.InvalidateUserCacheRequest) (*permissionv1.InvalidateUserCacheResponse, error) {
	if in.UserId <= 0 {
		return &permissionv1.InvalidateUserCacheResponse{
			Base: responsex.NewBaseResp(),
		}, nil
	}

	// 构建所有可能存在的缓存 key
	keys := []string{
		fmt.Sprintf("perm:user:%d", in.UserId),
		fmt.Sprintf("perm:scopes:%d:community", in.UserId),
		fmt.Sprintf("perm:scopes:%d:building", in.UserId),
		fmt.Sprintf("perm:scopes:%d:unit", in.UserId),
		fmt.Sprintf("perm:scopes:%d:grid", in.UserId),
	}

	deleted, err := l.svcCtx.RedisClient.Del(l.ctx, keys...).Result()
	if err != nil {
		l.Errorf("InvalidateUserCache: DEL keys failed for user=%d: %v", in.UserId, err)
		return nil, err
	}

	l.Infof("InvalidateUserCache: user=%d deleted %d keys", in.UserId, deleted)

	return &permissionv1.InvalidateUserCacheResponse{
		Base:    responsex.NewBaseResp(),
		Deleted: deleted,
	}, nil
}
