// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package template

import (
	"context"

	"community-and-home/services/ai-model/api/internal/svc"
	"community-and-home/services/ai-model/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteTemplateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除提示词模板
func NewDeleteTemplateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteTemplateLogic {
	return &DeleteTemplateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteTemplateLogic) DeleteTemplate(req *types.DeleteTemplateRequest) (resp *types.BaseResponse, err error) {
	// 删除模板
	err = l.svcCtx.PromptTemplateModel.Delete(l.ctx, req.Id)
	if err != nil {
		return &types.BaseResponse{
			Code:    500,
			Message: "删除模板失败: " + err.Error(),
		}, nil
	}

	return &types.BaseResponse{
		Code:    200,
		Message: "success",
	}, nil
}
