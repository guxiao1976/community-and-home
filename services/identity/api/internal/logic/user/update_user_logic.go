// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package user

import (
	"context"
	"database/sql"

	"github.com/guxiao/community-and-home/common/errorx"
	"github.com/guxiao/community-and-home/services/identity/api/internal/svc"
	"github.com/guxiao/community-and-home/services/identity/api/internal/types"
	"github.com/guxiao/community-and-home/services/identity/model"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
)

type UpdateUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Update user
func NewUpdateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserLogic {
	return &UpdateUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateUserLogic) UpdateUser(req *types.UpdateUserReq) (resp *types.UpdateUserResp, err error) {
	// Get existing user
	user, err := l.svcCtx.AuthUserModel.FindOne(l.ctx, req.Id)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errorx.NewDefaultError("user not found")
		}
		return nil, errorx.NewDefaultError("failed to get user")
	}

	// Update fields
	if req.Nickname != "" {
		user.Nickname = sql.NullString{String: req.Nickname, Valid: true}
	}
	if req.AvatarUrl != "" {
		user.AvatarUrl = sql.NullString{String: req.AvatarUrl, Valid: true}
	}
	if req.UserType != 0 {
		user.UserType = int64(req.UserType)
	}
	if req.Status != 0 {
		user.Status = int64(req.Status)
	}
	if req.ScopeId != nil {
		user.ScopeId = sql.NullInt64{Int64: *req.ScopeId, Valid: true}
	}
	if req.Password != "" {
		hashedBytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			logx.Errorf("failed to hash password: %v", err)
			return nil, errorx.NewDefaultError("failed to hash password")
		}
		user.PasswordHash = string(hashedBytes)
	}

	// Update user
	err = l.svcCtx.AuthUserModel.Update(l.ctx, user)
	if err != nil {
		logx.Errorf("failed to update user: %v", err)
		return nil, errorx.NewDefaultError("failed to update user")
	}

	return &types.UpdateUserResp{
		Success: true,
	}, nil
}
