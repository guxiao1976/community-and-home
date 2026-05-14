package template

import (
	"context"
	"database/sql"
	"time"

	"community-and-home/services/ai-model/api/internal/svc"
	"community-and-home/services/ai-model/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateTemplateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新提示词模板
func NewUpdateTemplateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateTemplateLogic {
	return &UpdateTemplateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateTemplateLogic) UpdateTemplate(req *types.UpdateTemplateRequest) (resp *types.TemplateResponse, err error) {
	// 查询现有记录
	existing, err := l.svcCtx.PromptTemplateModel.FindOne(l.ctx, req.Id)
	if err != nil {
		return &types.TemplateResponse{
			BaseResponse: types.BaseResponse{
				Code:    404,
				Message: "模板不存在",
			},
		}, nil
	}

	// 更新字段
	if req.Name != "" {
		existing.TemplateName = req.Name
	}
	if req.Content != "" {
		existing.Content = req.Content
	}
	if req.Category != "" {
		existing.Category = sql.NullString{
			String: req.Category,
			Valid:  true,
		}
	}
	if req.Description != "" {
		existing.Description = sql.NullString{
			String: req.Description,
			Valid:  true,
		}
	}
	existing.UpdatedTime = time.Now()

	// 更新数据库
	err = l.svcCtx.PromptTemplateModel.Update(l.ctx, existing)
	if err != nil {
		return &types.TemplateResponse{
			BaseResponse: types.BaseResponse{
				Code:    500,
				Message: "更新模板失败: " + err.Error(),
			},
		}, nil
	}

	category := ""
	if existing.Category.Valid {
		category = existing.Category.String
	}

	description := ""
	if existing.Description.Valid {
		description = existing.Description.String
	}

	return &types.TemplateResponse{
		BaseResponse: types.BaseResponse{
			Code:    200,
			Message: "success",
		},
		Data: types.TemplateInfo{
			Id:          existing.Id,
			Name:        existing.TemplateName,
			Content:     existing.Content,
			Category:    category,
			Description: description,
			CreatedAt:   existing.CreatedTime.Format("2006-01-02 15:04:05"),
			UpdatedAt:   existing.UpdatedTime.Format("2006-01-02 15:04:05"),
		},
	}, nil
}
