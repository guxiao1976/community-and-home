// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package user

import (
	"context"

	"github.com/guxiao/community-and-home/common/errorx"
	"github.com/guxiao/community-and-home/services/identity/api/internal/svc"
	"github.com/guxiao/community-and-home/services/identity/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUsersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// List users
func NewGetUsersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUsersLogic {
	return &GetUsersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUsersLogic) GetUsers(req *types.GetUsersReq) (resp *types.GetUsersResp, err error) {
	var userType, status *int64
	if req.UserType != nil {
		v := int64(*req.UserType)
		userType = &v
	}
	if req.Status != nil {
		v := int64(*req.Status)
		status = &v
	}

	total, err := l.svcCtx.AuthUserModel.CountByFilter(l.ctx, userType, status)
	if err != nil {
		return nil, errorx.NewDefaultError("failed to count users")
	}

	users, err := l.svcCtx.AuthUserModel.FindPage(l.ctx, userType, status, int64(req.Page), int64(req.PageSize))
	if err != nil {
		return nil, errorx.NewDefaultError("failed to get users")
	}

	var userList []types.User
	for _, user := range users {
		var scopeId *int64
		if user.ScopeId.Valid {
			scopeId = &user.ScopeId.Int64
		}

		userList = append(userList, types.User{
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
		})
	}

	return &types.GetUsersResp{
		Total: total,
		List:  userList,
	}, nil
}
