package logic

import (
	"context"

	"github.com/guxiao1976/community-auth/api/internal/svc"
	"github.com/guxiao1976/community-auth/api/internal/types"
	"github.com/guxiao1976/community-common/v2/pkg/crypto"
	"github.com/guxiao1976/community-common/v2/pkg/errx"
	"github.com/zeromicro/go-zero/core/logx"
)

type PublicKeyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPublicKeyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PublicKeyLogic {
	return &PublicKeyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// PublicKey 获取 RSA 公钥 PEM 字符串，供前端加密登录手机号和密码
func (l *PublicKeyLogic) PublicKey() (*types.PublicKeyResp, error) {
	publicKey, err := crypto.GetRSAPublicKey()
	if err != nil {
		l.Errorf("GetRSAPublicKey failed: %v", err)
		return nil, errx.NewDefaultError("RSA 密钥未初始化")
	}

	return &types.PublicKeyResp{
		PublicKey: publicKey,
	}, nil
}
