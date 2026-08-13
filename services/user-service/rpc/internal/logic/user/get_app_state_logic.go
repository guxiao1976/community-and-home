package user

import (
	"context"

	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-user/model"
	"github.com/guxiao1976/community-user/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetAppStateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetAppStateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAppStateLogic {
	return &GetAppStateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetAppState 读取用户当前小区应用状态（跨设备一致）；无记录返回 current_community_id=0
func (l *GetAppStateLogic) GetAppState(in *userv1.GetAppStateRequest) (*userv1.GetAppStateResponse, error) {
	state, err := l.svcCtx.UserAppStateModel.FindOne(l.ctx, in.UserId)
	if err != nil && err != model.ErrNotFound {
		l.Errorf("GetAppState: find app state error: %v", err)
		return nil, err
	}
	if err == model.ErrNotFound || state == nil {
		return &userv1.GetAppStateResponse{
			Base:               responsex.NewBaseResp(),
			CurrentCommunityId: 0,
		}, nil
	}
	return &userv1.GetAppStateResponse{
		Base:               responsex.NewBaseResp(),
		CurrentCommunityId: state.CurrentCommunityId,
		UpdatedAt:          state.UpdatedTime.Unix(),
	}, nil
}
