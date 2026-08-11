package user

import (
	"context"

	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-user/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type BatchUpdateUsersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBatchUpdateUsersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchUpdateUsersLogic {
	return &BatchUpdateUsersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *BatchUpdateUsersLogic) BatchUpdateUsers(in *userv1.BatchUpdateUsersRequest) (*userv1.BatchUpdateUsersResponse, error) {
	if len(in.UserIds) == 0 {
		return &userv1.BatchUpdateUsersResponse{
			Base:         responsex.NewBaseResp(),
			UpdatedCount: 0,
		}, nil
	}

	// 批量更新用户状态
	updatedCount := int32(0)
	for _, userId := range in.UserIds {
		err := l.svcCtx.UserBaseModel.UpdateStatus(l.ctx, userId, int64(in.Status))
		if err != nil {
			l.Errorf("batch update user %d status failed: %v", userId, err)
			continue
		}
		// 状态变更后：失效权限缓存 + 写/删禁用标记
		invalidateUserPermissionCache(l.ctx, l.svcCtx, l.Logger, userId, int64(in.Status))
		updatedCount++
	}

	l.Infof("batch update users: %d/%d succeeded", updatedCount, len(in.UserIds))

	return &userv1.BatchUpdateUsersResponse{
		Base:         responsex.NewBaseResp(),
		UpdatedCount: updatedCount,
	}, nil
}
