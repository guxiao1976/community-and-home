// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package apikey

import (
	"context"

	"github.com/guxiao/community-and-home/services/ai-model/api/internal/svc"
	"github.com/guxiao/community-and-home/services/ai-model/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateAPIKeyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建 API Key
func NewCreateAPIKeyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateAPIKeyLogic {
	return &CreateAPIKeyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateAPIKeyLogic) CreateAPIKey(req *types.CreateAPIKeyRequest) (resp *types.APIKeyResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
