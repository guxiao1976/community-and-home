package auth

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/guxiao/community-and-home/common/errorx"
	"github.com/guxiao/community-and-home/services/identity/api/internal/svc"
	"github.com/guxiao/community-and-home/services/identity/api/internal/types"
	"github.com/guxiao/community-and-home/services/identity/model"
	"golang.org/x/crypto/bcrypt"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// User login with password
func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginReq) (resp *types.LoginResp, err error) {
	// Find user by phone
	user, err := l.svcCtx.AuthUserModel.FindOneByPhone(l.ctx, req.Phone)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errorx.NewDefaultError("用户不存在")
		}
		logx.Errorf("failed to find user: %v", err)
		return nil, errorx.NewDefaultError("登录失败")
	}

	// Check if user is active
	if user.Status != 1 {
		return nil, errorx.NewDefaultError("账号已被禁用")
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, errorx.NewDefaultError("密码错误")
	}

	// Generate JWT token
	now := time.Now().Unix()
	accessExpire := l.svcCtx.Config.Auth.AccessExpire
	accessToken, err := l.generateToken(user.Id, now, accessExpire)
	if err != nil {
		logx.Errorf("failed to generate token: %v", err)
		return nil, errorx.NewDefaultError("生成token失败")
	}

	// Generate refresh token (7 days)
	refreshToken, err := l.generateToken(user.Id, now, 7*24*3600)
	if err != nil {
		logx.Errorf("failed to generate refresh token: %v", err)
		return nil, errorx.NewDefaultError("生成refresh token失败")
	}

	// Build user response
	var scopeId *int64
	if user.ScopeId.Valid {
		scopeId = &user.ScopeId.Int64
	}

	userResp := types.User{
		Id:                 user.Id,
		Phone:              user.Phone,
		Nickname:           user.Nickname.String,
		AvatarUrl:          user.AvatarUrl.String,
		UserType:           int32(user.UserType),
		Status:             int32(user.Status),
		VerificationStatus: int32(user.VerificationStatus),
		ScopeId:            scopeId,
		CreatedTime:        user.CreatedTime.Format(time.RFC3339),
		UpdatedTime:        user.UpdatedTime.Format(time.RFC3339),
	}

	return &types.LoginResp{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    accessExpire,
		User:         userResp,
	}, nil
}

func (l *LoginLogic) generateToken(userId int64, iat, seconds int64) (string, error) {
	claims := make(jwt.MapClaims)
	claims["exp"] = iat + seconds
	claims["iat"] = iat
	claims["userId"] = userId
	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = claims
	return token.SignedString([]byte(l.svcCtx.Config.Auth.AccessSecret))
}
