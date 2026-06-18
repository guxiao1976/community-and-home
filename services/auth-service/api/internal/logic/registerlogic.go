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

type RegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterLogic) Register(req *types.RegisterReq) (*types.RegisterResp, error) {
	// RSA 解密手机号
	phone, err := crypto.RSADecrypt(req.EncryptedPhone)
	if err != nil {
		l.Infof("Register: RSA decrypt phone failed: %v", err)
		return nil, errx.NewDefaultError("手机号解密失败")
	}

	// 调用 gRPC Register
	resp, err := l.svcCtx.AuthRpc.Register(l.ctx, &authv1.RegisterRequest{
		Phone:             phone,
		SmsCode:           req.SmsCode,
		EncryptedPassword: req.EncryptedPassword,
		Nickname:          req.Nickname,
		DeviceId:          req.DeviceId,
		DeviceType:        req.DeviceType,
	})
	if err != nil {
		return nil, err
	}
	if !responsex.IsSuccess(resp.Base) {
		return nil, responsex.ToError(resp.Base)
	}

	return &types.RegisterResp{
		UserId:       resp.UserId,
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresAt:    resp.ExpiresAt,
	}, nil
}
