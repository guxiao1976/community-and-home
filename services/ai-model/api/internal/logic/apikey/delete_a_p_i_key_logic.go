// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package apikey

import (
	"context"

	"community-and-home/services/ai-model/api/internal/svc"
	"community-and-home/services/ai-model/api/internal/types"

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
	// 删除 API Key
	err = l.svcCtx.ApiKeyModel.Delete(l.ctx, req.Id)
	if err != nil {
		return &types.BaseResponse{
			Code:    500,
			Message: "删除 API Key 失败: " + err.Error(),
		}, nil
	}

	return &types.BaseResponse{
		Code:    200,
		Message: "success",
	}, nil
}
