// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package apikey

import (
	"context"

	"github.com/guxiao/community-and-home/services/ai-model/api/internal/svc"
	"github.com/guxiao/community-and-home/services/ai-model/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListAPIKeysLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取 API Key 列表
func NewListAPIKeysLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAPIKeysLogic {
	return &ListAPIKeysLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListAPIKeysLogic) ListAPIKeys(req *types.ListAPIKeysRequest) (resp *types.APIKeysResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
