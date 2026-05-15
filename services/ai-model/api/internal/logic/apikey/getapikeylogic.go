// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package apikey

import (
	"context"

	"github.com/guxiao/community-and-home/services/ai-model/api/internal/svc"
	"github.com/guxiao/community-and-home/services/ai-model/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAPIKeyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取 API Key 详情
func NewGetAPIKeyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAPIKeyLogic {
	return &GetAPIKeyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAPIKeyLogic) GetAPIKey(req *types.GetAPIKeyRequest) (resp *types.APIKeyResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
