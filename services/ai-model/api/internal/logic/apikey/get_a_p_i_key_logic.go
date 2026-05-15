// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package apikey

import (
	"context"

	"community-and-home/services/ai-model/api/internal/svc"
	"community-and-home/services/ai-model/api/internal/types"

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
	// 从数据库查询API Key
	apiKey, err := l.svcCtx.ApiKeyModel.FindOne(l.ctx, req.Id)
	if err != nil {
		return &types.APIKeyResponse{
			BaseResponse: types.BaseResponse{
				Code:    500,
				Message: err.Error(),
			},
		}, nil
	}

	// 检查是否已删除
	if apiKey.DeleteTime.Valid {
		return &types.APIKeyResponse{
			BaseResponse: types.BaseResponse{
				Code:    404,
				Message: "API Key not found",
			},
		}, nil
	}

	return &types.APIKeyResponse{
		BaseResponse: types.BaseResponse{
			Code:    0,
			Message: "success",
		},
		Data: types.APIKeyInfo{
			Id:          apiKey.Id,
			ModelId:     apiKey.ModelId,
			KeyName:     apiKey.KeyName,
			Status:      apiKey.Status,
			Description: "", // TODO: 需要添加description字段到数据库
			CreatedAt:   apiKey.CreatedTime.Format("2006-01-02 15:04:05"),
			UpdatedAt:   apiKey.UpdatedTime.Format("2006-01-02 15:04:05"),
		},
	}, nil
}
