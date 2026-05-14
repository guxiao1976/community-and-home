package logic

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"community-and-home/services/ai-model/rpc/internal/svc"
	"community-and-home/services/ai-model/rpc/model"
	"community-and-home/services/ai-model/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateModelConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateModelConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateModelConfigLogic {
	return &CreateModelConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 模型配置管理
func (l *CreateModelConfigLogic) CreateModelConfig(in *pb.CreateModelConfigReq) (*pb.ModelConfigResp, error) {
	// 构造数据模型
	// 转换 supported_features 为 JSON 数组格式
	var capabilitiesJSON string
	if in.SupportedFeatures != "" {
		// 如果是逗号分隔的字符串，转换为JSON数组
		features := strings.Split(in.SupportedFeatures, ",")
		for i := range features {
			features[i] = strings.TrimSpace(features[i])
		}
		capabilitiesBytes, err := json.Marshal(features)
		if err != nil {
			l.Errorf("marshal capabilities failed: %v", err)
			return nil, err
		}
		capabilitiesJSON = string(capabilitiesBytes)
	}

	modelConfig := &model.AmModelConfig{
		ModelName:             in.Name,
		ModelType:             in.Type,
		Provider:              in.Provider,
		DisplayName:           sql.NullString{String: in.DisplayName, Valid: in.DisplayName != ""},
		Description:           sql.NullString{String: in.Description, Valid: in.Description != ""},
		Endpoint:              sql.NullString{String: in.Endpoint, Valid: in.Endpoint != ""},
		MaxRetries:            2,
		Timeout:               30000,
		CostPer1KInputTokens:  in.CostPer_1KInputTokens,
		CostPer1KOutputTokens: in.CostPer_1KOutputTokens,
		Status:                1,
		HealthStatus:          "unknown",
		Priority:              100,
		CreatedTime:           time.Now(),
		UpdatedTime:           time.Now(),
	}

	// 如果提供了 supported_features，设置 capabilities
	if capabilitiesJSON != "" {
		modelConfig.Capabilities = sql.NullString{String: capabilitiesJSON, Valid: true}
	}

	// 插入数据库
	result, err := l.svcCtx.ModelManager.CreateModelConfig(l.ctx, modelConfig)
	if err != nil {
		l.Errorf("create model config failed: %v", err)
		return nil, err
	}

	// 获取插入的ID
	id, err := result.LastInsertId()
	if err != nil {
		l.Errorf("get last insert id failed: %v", err)
		return nil, err
	}

	// 返回创建的模型配置
	return &pb.ModelConfigResp{
		Id:                     id,
		Name:                   in.Name,
		DisplayName:            in.DisplayName,
		Provider:               in.Provider,
		Type:                   in.Type,
		Endpoint:               in.Endpoint,
		MaxTokens:              in.MaxTokens,
		SupportedFeatures:      in.SupportedFeatures,
		CostPer_1KInputTokens:  in.CostPer_1KInputTokens,
		CostPer_1KOutputTokens: in.CostPer_1KOutputTokens,
		Status:                 1,
		Description:            in.Description,
	}, nil
}
