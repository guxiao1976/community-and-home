// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package auth

import (
	"context"
	"database/sql"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/guxiao/community-and-home/common/errorx"
	"github.com/guxiao/community-and-home/services/identity/api/internal/svc"
	"github.com/guxiao/community-and-home/services/identity/api/internal/types"
	"github.com/guxiao/community-and-home/services/identity/model"
	"golang.org/x/crypto/bcrypt"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// User register
func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterLogic) Register(req *types.RegisterReq) (resp *types.RegisterResp, err error) {
	// Check if phone already exists
	_, err = l.svcCtx.AuthUserModel.FindOneByPhone(l.ctx, req.Phone)
	if err == nil {
		return nil, errorx.NewDefaultError("手机号已注册")
	}
	if err != model.ErrNotFound {
		logx.Errorf("failed to check phone: %v", err)
		return nil, errorx.NewDefaultError("注册失败")
	}

	// TODO: Verify SMS code (skipped for now, would integrate with SMS service)
	// For now, we'll accept any code in development

	// Hash password
	var passwordHash string
	if req.Password != "" {
		hashedBytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			logx.Errorf("failed to hash password: %v", err)
			return nil, errorx.NewDefaultError("密码加密失败")
		}
		passwordHash = string(hashedBytes)
	}

	// Set default nickname if not provided
	nickname := req.Nickname
	if nickname == "" {
		nickname = "用户" + req.Phone[len(req.Phone)-4:]
	}

	// Create user
	user := &model.AuthUser{
		Phone:        req.Phone,
		PasswordHash: passwordHash,
		Nickname:     sql.NullString{String: nickname, Valid: true},
		UserType:     1, // Default: resident
		Status:       1, // Active
	}

	result, err := l.svcCtx.AuthUserModel.Insert(l.ctx, user)
	if err != nil {
		logx.Errorf("failed to insert user: %v", err)
		return nil, errorx.NewDefaultError("创建用户失败")
	}

	userId, err := result.LastInsertId()
	if err != nil {
		logx.Errorf("failed to get user id: %v", err)
		return nil, errorx.NewDefaultError("获取用户ID失败")
	}

	// Generate JWT token
	now := time.Now().Unix()
	accessExpire := l.svcCtx.Config.Auth.AccessExpire
	accessToken, err := l.generateToken(userId, now, accessExpire)
	if err != nil {
		logx.Errorf("failed to generate token: %v", err)
		return nil, errorx.NewDefaultError("生成token失败")
	}

	return &types.RegisterResp{
		UserId: userId,
		Token:  accessToken,
		Expire: now + accessExpire,
	}, nil
}

func (l *RegisterLogic) generateToken(userId int64, iat, seconds int64) (string, error) {
	claims := make(jwt.MapClaims)
	claims["exp"] = iat + seconds
	claims["iat"] = iat
	claims["userId"] = userId
	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = claims
	return token.SignedString([]byte(l.svcCtx.Config.Auth.AccessSecret))
}
