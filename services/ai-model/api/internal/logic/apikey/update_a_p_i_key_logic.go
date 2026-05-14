// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package apikey

import (
	"context"
	"time"

	"community-and-home/services/ai-model/api/internal/svc"
	"community-and-home/services/ai-model/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateAPIKeyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新 API Key
func NewUpdateAPIKeyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateAPIKeyLogic {
	return &UpdateAPIKeyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateAPIKeyLogic) UpdateAPIKey(req *types.UpdateAPIKeyRequest) (resp *types.APIKeyResponse, err error) {
	// 查询现有记录
	existing, err := l.svcCtx.ApiKeyModel.FindOne(l.ctx, req.Id)
	if err != nil {
		return &types.APIKeyResponse{
			BaseResponse: types.BaseResponse{
				Code:    404,
				Message: "API Key 不存在",
			},
		}, nil
	}

	// 更新字段
	if req.KeyName != "" {
		existing.KeyName = req.KeyName
	}
	if req.ApiKey != "" {
		existing.ApiKey = req.ApiKey
	}
	if req.Status > 0 {
		existing.Status = req.Status
	}
	existing.UpdatedTime = time.Now()

	// 更新数据库
	err = l.svcCtx.ApiKeyModel.Update(l.ctx, existing)
	if err != nil {
		return &types.APIKeyResponse{
			BaseResponse: types.BaseResponse{
				Code:    500,
				Message: "更新 API Key 失败: " + err.Error(),
			},
		}, nil
	}

	description := ""
	if req.Description != "" {
		description = req.Description
	}

	return &types.APIKeyResponse{
		BaseResponse: types.BaseResponse{
			Code:    200,
			Message: "success",
		},
		Data: types.APIKeyInfo{
			Id:          existing.Id,
			KeyName:     existing.KeyName,
			Status:      existing.Status,
			Description: description,
			CreatedAt:   existing.CreatedTime.Format("2006-01-02 15:04:05"),
			UpdatedAt:   existing.UpdatedTime.Format("2006-01-02 15:04:05"),
		},
	}, nil
}
