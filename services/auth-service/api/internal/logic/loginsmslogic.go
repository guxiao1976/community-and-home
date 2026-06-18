package logic

import (
	"context"

	authv1 "github.com/guxiao1976/api-proto/gen/go/auth/v1"
	"github.com/guxiao1976/community-auth/api/internal/svc"
	"github.com/guxiao1976/community-auth/api/internal/types"
	"github.com/guxiao1976/community-common/v2/pkg/crypto"
	"github.com/guxiao1976/community-common/v2/pkg/errx"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/zeromicro/go-zero/core/logx"
)

type LoginSmsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginSmsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginSmsLogic {
	return &LoginSmsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginSmsLogic) LoginSms(req *types.LoginSmsReq) (*types.LoginResp, error) {
	// RSA 解密手机号
	phone, err := crypto.RSADecrypt(req.EncryptedPhone)
	if err != nil {
		l.Infof("LoginSms: RSA decrypt phone failed: %v", err)
		return nil, errx.NewDefaultError("手机号解密失败")
	}

	// 调用 gRPC LoginSms
	resp, err := l.svcCtx.AuthRpc.LoginSms(l.ctx, &authv1.LoginSmsRequest{
		Phone:      phone,
		SmsCode:    req.SmsCode,
		DeviceId:   req.DeviceId,
		DeviceType: req.DeviceType,
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
