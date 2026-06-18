package logic

import (
	"context"

	authv1 "github.com/guxiao1976/api-proto/gen/go/auth/v1"
	"github.com/guxiao1976/community-auth/api/internal/svc"
	"github.com/guxiao1976/community-auth/api/internal/types"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginReq) (*types.LoginResp, error) {
	// 调用 gRPC Login
	resp, err := l.svcCtx.AuthRpc.Login(l.ctx, &authv1.LoginRequest{
		EncryptedPhone:    req.EncryptedPhone,
		EncryptedPassword: req.EncryptedPassword,
		DeviceId:          req.DeviceId,
		DeviceType:        req.DeviceType,
	})
	if err != nil {
		return nil, err
	}
	if !responsex.IsSuccess(resp.Base) {
		return nil, responsex.ToError(resp.Base)
	}

	return &types.LoginResp{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresAt:    resp.ExpiresAt,
		UserId:       resp.UserId,
	}, nil
}
