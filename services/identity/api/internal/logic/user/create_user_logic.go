// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package user

import (
	"context"
	"database/sql"
	"time"

	"github.com/guxiao/community-and-home/common/errorx"
	"github.com/guxiao/community-and-home/services/identity/api/internal/svc"
	"github.com/guxiao/community-and-home/services/identity/api/internal/types"
	"github.com/guxiao/community-and-home/services/identity/model"
	"golang.org/x/crypto/bcrypt"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Create user
func NewCreateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateUserLogic {
	return &CreateUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateUserLogic) CreateUser(req *types.CreateUserReq) (resp *types.CreateUserResp, err error) {
	// Check if phone already exists
	_, err = l.svcCtx.AuthUserModel.FindOneByPhone(l.ctx, req.Phone)
	if err == nil {
		return nil, errorx.NewDefaultError("phone already exists")
	}
	if err != model.ErrNotFound {
		logx.Errorf("failed to check phone: %v", err)
		return nil, errorx.NewDefaultError("failed to check phone")
	}

	// Hash password
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logx.Errorf("failed to hash password: %v", err)
		return nil, errorx.NewDefaultError("failed to hash password")
	}

	// Set default nickname if not provided
	nickname := req.Nickname
	if nickname == "" {
		nickname = "用户" + req.Phone[len(req.Phone)-4:]
	}

	// Create user
	now := time.Now()
	user := &model.AuthUser{
		Phone:        req.Phone,
		PasswordHash: string(hashedBytes),
		Nickname:     sql.NullString{String: nickname, Valid: true},
		UserType:     int64(req.UserType),
		Status:       1, // Active by default
		CreatedTime:  now,
		UpdatedTime:  now,
	}

	if req.ScopeId != nil {
		user.ScopeId = sql.NullInt64{Int64: *req.ScopeId, Valid: true}
	}

	result, err := l.svcCtx.AuthUserModel.Insert(l.ctx, user)
	if err != nil {
		logx.Errorf("failed to insert user: %v", err)
		return nil, errorx.NewDefaultError("failed to create user")
	}

	userId, err := result.LastInsertId()
	if err != nil {
		logx.Errorf("failed to get user id: %v", err)
		return nil, errorx.NewDefaultError("failed to get user id")
	}

	return &types.CreateUserResp{
		Id: userId,
	}, nil
}
