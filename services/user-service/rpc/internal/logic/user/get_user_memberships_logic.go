package user

import (
	"context"

	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-user/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
)

type GetUserMembershipsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserMembershipsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserMembershipsLogic {
	return &GetUserMembershipsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserMembershipsLogic) GetUserMemberships(in *userv1.GetUserMembershipsRequest) (*userv1.GetUserMembershipsResponse, error) {
	memberships, err := l.svcCtx.UserCommunityMembershipModel.FindByUserId(l.ctx, in.UserId)
	if err != nil {
		l.Errorf("find memberships error: %v", err)
		return nil, err
	}

	return &userv1.GetUserMembershipsResponse{
		Base:        responsex.NewBaseResp(),
		Memberships: toProtoMemberships(memberships),
	}, nil
}
