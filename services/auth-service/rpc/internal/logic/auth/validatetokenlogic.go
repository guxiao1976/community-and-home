package auth

import (
	"context"
	"fmt"

	authv1 "github.com/guxiao1976/api-proto/gen/go/auth/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-auth/rpc/internal/svc"
	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/core/logx"
)

// ValidateTokenLogic 验证 AT（spec/auth.md 网关协同逻辑）
//
//   API Gateway 调用此接口检查 AT 是否在黑名单中。
//   除了 Redis 黑名单检查外，还验证 JWT 签名和过期时间。
type ValidateTokenLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewValidateTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ValidateTokenLogic {
	return &ValidateTokenLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ValidateToken 验证 AT 是否有效
func (l *ValidateTokenLogic) ValidateToken(in *authv1.ValidateTokenRequest) (*authv1.ValidateTokenResponse, error) {
	// 1. 解析 JWT，验证签名和过期
	atToken, err := jwt.Parse(in.AccessToken, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(l.svcCtx.Config.JwtAuth.AccessSecret), nil
	})

	if err != nil || !atToken.Valid {
		return &authv1.ValidateTokenResponse{
			Base:  responsex.NewBaseResp(),
			Valid: false,
		}, nil
	}

	claims, ok := atToken.Claims.(jwt.MapClaims)
	if !ok {
		return &authv1.ValidateTokenResponse{
			Base:  responsex.NewBaseResp(),
			Valid: false,
		}, nil
	}

	jti, _ := claims["jti"].(string)
	userID, _ := claims["user_id"].(float64)
	exp, _ := claims["exp"].(float64)

	// 2. 检查 AT 黑名单（见 spec/auth.md 网关协同逻辑步骤 2）
	blacklistKey := fmt.Sprintf("auth:at:blacklist:%s", jti)
	blacklisted, err := l.svcCtx.RedisClient.Exists(l.ctx, blacklistKey).Result()
	if err == nil && blacklisted > 0 {
		l.Infof("ValidateToken: AT jti=%s found in blacklist", jti)
		return &authv1.ValidateTokenResponse{
			Base:  responsex.NewBaseResp(),
			Valid: false,
		}, nil
	}

	return &authv1.ValidateTokenResponse{
		Base:      responsex.NewBaseResp(),
		Valid:     true,
		UserId:    int64(userID),
		ExpiresAt: int64(exp),
	}, nil
}
