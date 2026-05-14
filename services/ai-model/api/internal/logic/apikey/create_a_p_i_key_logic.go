package apikey

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"time"

	"community-and-home/services/ai-model/api/internal/svc"
	"community-and-home/services/ai-model/api/internal/types"
	"community-and-home/services/ai-model/rpc/model"

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
	// 使用用户提供的 API Key 或生成新的
	apiKey := req.ApiKey
	if apiKey == "" {
		// 生成随机 API Key
		keyBytes := make([]byte, 32)
		if _, err := rand.Read(keyBytes); err != nil {
			return &types.APIKeyResponse{
				BaseResponse: types.BaseResponse{
					Code:    500,
					Message: "生成 API Key 失败: " + err.Error(),
				},
			}, nil
		}
		apiKey = "sk-" + hex.EncodeToString(keyBytes)
	}

	// 生成脱敏显示
	maskedKey := apiKey[:10] + "***" + apiKey[len(apiKey)-6:]

	// 查询模型信息获取 provider
	conn, _ := l.svcCtx.DB.RawDB()

	var provider string
	err = conn.QueryRowContext(l.ctx, "SELECT provider FROM am_model_config WHERE id = ?", req.ModelId).Scan(&provider)
	if err != nil {
		provider = "unknown"
	}

	// 构建数据库记录
	now := time.Now()
	record := &model.AmApiKey{
		KeyName:  req.KeyName,
		Provider: provider,
		ApiKey:   apiKey, // 实际应该加密存储
		MaskedKey: sql.NullString{
			String: maskedKey,
			Valid:  true,
		},
		DailyQuota:   0, // 0=无限制
		MonthlyQuota: 0,
		Priority:     100,
		Status:       1, // 1=enabled
		FailureCount: 0,
		CreatedTime:  now,
		UpdatedTime:  now,
	}

	// 插入数据库
	result, err := l.svcCtx.ApiKeyModel.Insert(l.ctx, record)
	if err != nil {
		return &types.APIKeyResponse{
			BaseResponse: types.BaseResponse{
				Code:    500,
				Message: "创建 API Key 失败: " + err.Error(),
			},
		}, nil
	}

	id, _ := result.LastInsertId()

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
			Id:          id,
			ModelId:     req.ModelId,
			KeyName:     req.KeyName,
			Status:      1,
			Description: description,
			CreatedAt:   now.Format("2006-01-02 15:04:05"),
			UpdatedAt:   now.Format("2006-01-02 15:04:05"),
		},
	}, nil
}
