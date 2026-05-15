package logic

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"github.com/guxiao/community-and-home/services/ai-model/rpc/internal/svc"
	"github.com/guxiao/community-and-home/services/ai-model/rpc/model"
	"github.com/guxiao/community-and-home/services/ai-model/rpc/pb"

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
	l.Logger.Infof("CreateModelConfig called with: name=%s, provider=%s, type=%s", in.Name, in.Provider, in.Type)

	// 将 supported_features 转换为 JSON 数组
	var capabilitiesJSON sql.NullString
	if in.SupportedFeatures != "" {
		features := strings.Split(in.SupportedFeatures, ",")
		for i := range features {
			features[i] = strings.TrimSpace(features[i])
		}
		jsonBytes, err := json.Marshal(features)
		if err != nil {
			l.Logger.Errorf("Failed to marshal capabilities: %v", err)
			return nil, err
		}
		capabilitiesJSON = sql.NullString{String: string(jsonBytes), Valid: true}
	}

	// 插入模型配置到数据库
	now := time.Now()
	result, err := l.svcCtx.ModelConfigModel.Insert(l.ctx, &model.AmModelConfig{
		ModelName:             in.Name,
		DisplayName:           sql.NullString{String: in.DisplayName, Valid: in.DisplayName != ""},
		Provider:              in.Provider,
		ModelType:             in.Type,
		Endpoint:              sql.NullString{String: in.Endpoint, Valid: in.Endpoint != ""},
		Capabilities:          capabilitiesJSON,
		CostPer1KInputTokens:  in.CostPer_1KInputTokens,
		CostPer1KOutputTokens: in.CostPer_1KOutputTokens,
		Status:                1, // 默认启用
		Description:           sql.NullString{String: in.Description, Valid: in.Description != ""},
		HealthStatus:          "unknown",
		Timeout:               30000, // 默认30秒超时
		MaxRetries:            3,     // 默认重试3次
		Priority:              100,   // 默认优先级
		CreatedTime:           now,
		UpdatedTime:           now,
	})
	if err != nil {
		l.Logger.Errorf("Failed to insert model config: %v", err)
		return nil, err
	}

	// 获取插入的 ID
	id, err := result.LastInsertId()
	if err != nil {
		l.Logger.Errorf("Failed to get last insert id: %v", err)
		return nil, err
	}

	l.Logger.Infof("Model config created successfully with id=%d", id)

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
