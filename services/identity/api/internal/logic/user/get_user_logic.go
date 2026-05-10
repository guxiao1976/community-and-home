// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package user

import (
	"context"

	"github.com/guxiao/community-and-home/common/errorx"
	"github.com/guxiao/community-and-home/services/identity/api/internal/svc"
	"github.com/guxiao/community-and-home/services/identity/api/internal/types"
	"github.com/guxiao/community-and-home/services/identity/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Get user details
func NewGetUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserLogic {
	return &GetUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserLogic) GetUser(req *types.GetUserReq) (resp *types.GetUserResp, err error) {
	user, err := l.svcCtx.AuthUserModel.FindOne(l.ctx, req.Id)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errorx.NewDefaultError("user not found")
		}
		return nil, errorx.NewDefaultError("failed to get user")
	}

	var scopeId *int64
	if user.ScopeId.Valid {
		scopeId = &user.ScopeId.Int64
	}

	return &types.GetUserResp{
		User: types.User{
			Id:                 user.Id,
			Phone:              user.Phone,
			Nickname:           user.Nickname.String,
			AvatarUrl:          user.AvatarUrl.String,
			UserType:           int32(user.UserType),
			Status:             int32(user.Status),
			VerificationStatus: int32(user.VerificationStatus),
			ScopeId:            scopeId,
			CreatedTime:        user.CreatedTime.Format("2006-01-02 15:04:05"),
			UpdatedTime:        user.UpdatedTime.Format("2006-01-02 15:04:05"),
		},
	}, nil
}
