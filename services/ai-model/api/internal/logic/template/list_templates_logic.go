package template

import (
	"context"
	"fmt"

	"community-and-home/services/ai-model/api/internal/svc"
	"community-and-home/services/ai-model/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListTemplatesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取提示词模板列表
func NewListTemplatesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTemplatesLogic {
	return &ListTemplatesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListTemplatesLogic) ListTemplates(req *types.ListTemplatesRequest) (resp *types.TemplatesResponse, err error) {
	// 构建查询条件
	whereClause := "WHERE delete_time IS NULL"
	args := []interface{}{}

	if req.Category != "" {
		whereClause += " AND category = ?"
		args = append(args, req.Category)
	}
	if req.Status > 0 {
		whereClause += " AND is_active = ?"
		args = append(args, req.Status)
	}

	// 查询总数
	var total int32
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM am_prompt_template %s", whereClause)
	conn, _ := l.svcCtx.DB.RawDB()
	err = conn.QueryRowContext(l.ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return &types.TemplatesResponse{
			BaseResponse: types.BaseResponse{
				Code:    500,
				Message: "查询模板总数失败: " + err.Error(),
			},
		}, nil
	}

	// 查询模板列表
	query := fmt.Sprintf(`
		SELECT id, template_name, content, category, description, created_time, updated_time
		FROM am_prompt_template %s
		ORDER BY created_time DESC
		LIMIT 100
	`, whereClause)

	rows, err := conn.QueryContext(l.ctx, query, args...)
	if err != nil {
		return &types.TemplatesResponse{
			BaseResponse: types.BaseResponse{
				Code:    500,
				Message: "查询模板列表失败: " + err.Error(),
			},
		}, nil
	}
	defer rows.Close()

	var templateList []types.TemplateInfo
	for rows.Next() {
		var t types.TemplateInfo
		var category, description *string

		err = rows.Scan(
			&t.Id,
			&t.Name,
			&t.Content,
			&category,
			&description,
			&t.CreatedAt,
			&t.UpdatedAt,
		)
		if err != nil {
			continue
		}

		if category != nil {
			t.Category = *category
		}
		if description != nil {
			t.Description = *description
		}

		templateList = append(templateList, t)
	}

	return &types.TemplatesResponse{
		BaseResponse: types.BaseResponse{
			Code:    200,
			Message: "success",
		},
		Data: types.TemplatesData{
			Templates: templateList,
			Total:     total,
		},
	}, nil
}
