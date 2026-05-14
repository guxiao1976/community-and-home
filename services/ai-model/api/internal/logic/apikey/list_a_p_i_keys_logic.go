// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package apikey

import (
	"context"
	"fmt"

	"community-and-home/services/ai-model/api/internal/svc"
	"community-and-home/services/ai-model/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListAPIKeysLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取 API Key 列表
func NewListAPIKeysLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAPIKeysLogic {
	return &ListAPIKeysLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListAPIKeysLogic) ListAPIKeys(req *types.ListAPIKeysRequest) (resp *types.APIKeysResponse, err error) {
	// 构建查询条件
	whereClause := "WHERE delete_time IS NULL"
	args := []interface{}{}

	if req.Status > 0 {
		whereClause += " AND status = ?"
		args = append(args, req.Status)
	}

	// 查询总数
	var total int32
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM am_api_key %s", whereClause)
	conn, _ := l.svcCtx.DB.RawDB()
	err = conn.QueryRowContext(l.ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return &types.APIKeysResponse{
			BaseResponse: types.BaseResponse{
				Code:    500,
				Message: "查询 API Key 总数失败: " + err.Error(),
			},
		}, nil
	}

	// 查询列表
	query := fmt.Sprintf(`
		SELECT id, key_name, status, created_time, updated_time
		FROM am_api_key %s
		ORDER BY created_time DESC
		LIMIT 100
	`, whereClause)

	rows, err := conn.QueryContext(l.ctx, query, args...)
	if err != nil {
		return &types.APIKeysResponse{
			BaseResponse: types.BaseResponse{
				Code:    500,
				Message: "查询 API Key 列表失败: " + err.Error(),
			},
		}, nil
	}
	defer rows.Close()

	var keys []types.APIKeyInfo
	for rows.Next() {
		var key types.APIKeyInfo

		err = rows.Scan(
			&key.Id,
			&key.KeyName,
			&key.Status,
			&key.CreatedAt,
			&key.UpdatedAt,
		)
		if err != nil {
			continue
		}

		keys = append(keys, key)
	}

	return &types.APIKeysResponse{
		BaseResponse: types.BaseResponse{
			Code:    200,
			Message: "success",
		},
		Data: types.APIKeysData{
			Keys:  keys,
			Total: total,
		},
	}, nil
}
