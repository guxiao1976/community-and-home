package logic

import (
	"context"

	authv1 "github.com/guxiao1976/api-proto/gen/go/auth/v1"
	"github.com/guxiao1976/community-auth/api/internal/svc"
	"github.com/guxiao1976/community-auth/api/internal/types"
	"github.com/guxiao1976/community-common/v2/pkg/errx"
	"github.com/guxiao1976/community-common/v2/pkg/jwt"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/zeromicro/go-zero/core/logx"
)

type LogoutLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLogoutLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LogoutLogic {
	return &LogoutLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// Logout 注销登录
// accessToken 从 Authorization header 提取，已由 JWT 中间件验证签名
func (l *LogoutLogic) Logout(accessToken string, req *types.LogoutReq) error {
	// 从 JWT 中提取 user_id（无需验签，JWT 中间件已验证）
	claims, err := jwt.ParseTokenUnverified(accessToken)
	if err != nil {
		l.Errorf("ParseTokenUnverified failed: %v", err)
		return errx.NewUnauthorizedError("无效的访问令牌")
	}

	// 调用 gRPC Logout
	resp, err := l.svcCtx.AuthRpc.Logout(l.ctx, &authv1.LogoutRequest{
		AccessToken:    accessToken,
		UserId:         claims.UserID,
		DeviceId:       req.DeviceId,
		KickAllDevices: req.KickAllDevices,
	})
	if err != nil {
		return err
	}
	if !responsex.IsSuccess(resp.Base) {
		return responsex.ToError(resp.Base)
	}

	return nil
}
