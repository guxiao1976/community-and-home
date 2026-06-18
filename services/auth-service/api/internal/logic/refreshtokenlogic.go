package logic

import (
	"context"

	authv1 "github.com/guxiao1976/api-proto/gen/go/auth/v1"
	"github.com/guxiao1976/community-auth/api/internal/svc"
	"github.com/guxiao1976/community-auth/api/internal/types"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/zeromicro/go-zero/core/logx"
)

type RefreshTokenLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRefreshTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshTokenLogic {
	return &RefreshTokenLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RefreshTokenLogic) RefreshToken(req *types.RefreshTokenReq) (*types.RefreshTokenResp, error) {
	// 调用 gRPC RefreshToken
	resp, err := l.svcCtx.AuthRpc.RefreshToken(l.ctx, &authv1.RefreshTokenRequest{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		return nil, err
	}
	if !responsex.IsSuccess(resp.Base) {
		return nil, responsex.ToError(resp.Base)
	}

	return &types.RefreshTokenResp{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresAt:    resp.ExpiresAt,
	}, nil
}
