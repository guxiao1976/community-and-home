package user

import (
	"context"

	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-user/model"
	"github.com/guxiao1976/community-user/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
)

type GetUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserLogic {
	return &GetUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserLogic) GetUser(in *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
	user, err := l.svcCtx.UserBaseModel.FindOne(l.ctx, in.Id)
	if err != nil {
		if err == model.ErrNotFound {
			return &userv1.GetUserResponse{
				Base: responsex.NewBaseRespWithError(10001, "用户不存在"),
			}, nil
		}
		l.Errorf("find user error: %v", err)
		return nil, err
	}

	return &userv1.GetUserResponse{
		Base: responsex.NewBaseResp(),
		User: toProtoUser(user),
	}, nil
}
