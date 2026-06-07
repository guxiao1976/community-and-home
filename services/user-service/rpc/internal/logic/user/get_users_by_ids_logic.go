package user

import (
	"context"

	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-user/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
)

type GetUsersByIdsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUsersByIdsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUsersByIdsLogic {
	return &GetUsersByIdsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUsersByIdsLogic) GetUsersByIds(in *userv1.GetUsersByIdsRequest) (*userv1.GetUsersByIdsResponse, error) {
	if len(in.Ids) == 0 {
		return &userv1.GetUsersByIdsResponse{
			Base:  responsex.NewBaseResp(),
			Users: nil,
		}, nil
	}

	users, err := l.svcCtx.UserBaseModel.FindByIds(l.ctx, in.Ids)
	if err != nil {
		l.Errorf("find users by ids error: %v", err)
		return nil, err
	}

	return &userv1.GetUsersByIdsResponse{
		Base:  responsex.NewBaseResp(),
		Users: toProtoUsers(users),
	}, nil
}
