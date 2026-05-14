// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package template

import (
	"context"

	"community-and-home/services/ai-model/api/internal/svc"
	"community-and-home/services/ai-model/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTemplateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取提示词模板详情
func NewGetTemplateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTemplateLogic {
	return &GetTemplateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTemplateLogic) GetTemplate(req *types.GetTemplateRequest) (resp *types.TemplateResponse, err error) {
	// 从数据库查询提示词模板
	template, err := l.svcCtx.PromptTemplateModel.FindOne(l.ctx, req.Id)
	if err != nil {
		return &types.TemplateResponse{
			BaseResponse: types.BaseResponse{
				Code:    500,
				Message: err.Error(),
			},
		}, nil
	}

	// 检查是否已删除
	if template.DeleteTime.Valid {
		return &types.TemplateResponse{
			BaseResponse: types.BaseResponse{
				Code:    404,
				Message: "Template not found",
			},
		}, nil
	}

	return &types.TemplateResponse{
		BaseResponse: types.BaseResponse{
			Code:    0,
			Message: "success",
		},
		Data: types.TemplateInfo{
			Id:          template.Id,
			Name:        template.TemplateName,
			Content:     template.Content,
			Category:    template.Category.String,
			Description: template.Description.String,
			CreatedAt:   template.CreatedTime.Format("2006-01-02 15:04:05"),
			UpdatedAt:   template.UpdatedTime.Format("2006-01-02 15:04:05"),
		},
	}, nil
}
