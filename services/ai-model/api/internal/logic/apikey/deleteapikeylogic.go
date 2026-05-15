// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package apikey

import (
	"context"

	"github.com/guxiao/community-and-home/services/ai-model/api/internal/svc"
	"github.com/guxiao/community-and-home/services/ai-model/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteAPIKeyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除 API Key
func NewDeleteAPIKeyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteAPIKeyLogic {
	return &DeleteAPIKeyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteAPIKeyLogic) DeleteAPIKey(req *types.DeleteAPIKeyRequest) (resp *types.BaseResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
