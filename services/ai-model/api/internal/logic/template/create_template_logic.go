package template

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"community-and-home/services/ai-model/api/internal/svc"
	"community-and-home/services/ai-model/api/internal/types"
	"community-and-home/services/ai-model/rpc/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateTemplateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建提示词模板
func NewCreateTemplateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTemplateLogic {
	return &CreateTemplateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateTemplateLogic) CreateTemplate(req *types.CreateTemplateRequest) (resp *types.TemplateResponse, err error) {
	now := time.Now()

	// 生成模板ID
	templateId := fmt.Sprintf("tpl_%d", now.Unix())

	record := &model.AmPromptTemplate{
		TemplateId:   templateId,
		TemplateName: req.Name,
		Category: sql.NullString{
			String: req.Category,
			Valid:  req.Category != "",
		},
		Content: req.Content,
		Description: sql.NullString{
			String: req.Description,
			Valid:  req.Description != "",
		},
		Version:     "1.0",
		IsActive:    1, // 1=active
		UsageCount:  0,
		CreatedTime: now,
		UpdatedTime: now,
	}

	result, err := l.svcCtx.PromptTemplateModel.Insert(l.ctx, record)
	if err != nil {
		return &types.TemplateResponse{
			BaseResponse: types.BaseResponse{
				Code:    500,
				Message: "创建模板失败: " + err.Error(),
			},
		}, nil
	}

	id, _ := result.LastInsertId()

	return &types.TemplateResponse{
		BaseResponse: types.BaseResponse{
			Code:    200,
			Message: "success",
		},
		Data: types.TemplateInfo{
			Id:          id,
			Name:        req.Name,
			Content:     req.Content,
			Category:    req.Category,
			Description: req.Description,
			CreatedAt:   now.Format("2006-01-02 15:04:05"),
			UpdatedAt:   now.Format("2006-01-02 15:04:05"),
		},
	}, nil
}
