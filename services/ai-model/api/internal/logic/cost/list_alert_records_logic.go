// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package cost

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"community-and-home/services/ai-model/api/internal/svc"
	"community-and-home/services/ai-model/api/internal/types"
	"community-and-home/services/ai-model/rpc/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListAlertRecordsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取预警记录
func NewListAlertRecordsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAlertRecordsLogic {
	return &ListAlertRecordsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListAlertRecordsLogic) ListAlertRecords(req *types.ListAlertRecordsRequest) (resp *types.AlertRecordsResponse, err error) {
	// 构建查询条件
	var conditions []string
	var args []interface{}

	if req.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, req.Status)
	}
	if req.StartDate != "" {
		conditions = append(conditions, "alert_time >= ?")
		args = append(args, req.StartDate)
	}
	if req.EndDate != "" {
		conditions = append(conditions, "alert_time <= ?")
		args = append(args, req.EndDate)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	conn, _ := l.svcCtx.DB.RawDB()

	// 查询总数
	var total int32
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM am_alert_record %s", whereClause)
	err = conn.QueryRowContext(l.ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return &types.AlertRecordsResponse{
			BaseResponse: types.BaseResponse{
				Code:    500,
				Message: "查询预警记录总数失败: " + err.Error(),
			},
		}, nil
	}

	// 查询记录列表
	query := fmt.Sprintf(`
		SELECT id, alert_type, threshold, actual_value, message,
		       model_id, model_name, status, alert_time
		FROM am_alert_record %s
		ORDER BY alert_time DESC
		LIMIT 100
	`, whereClause)

	rows, err := conn.QueryContext(l.ctx, query, args...)
	if err != nil {
		return &types.AlertRecordsResponse{
			BaseResponse: types.BaseResponse{
				Code:    500,
				Message: "查询预警记录失败: " + err.Error(),
			},
		}, nil
	}
	defer rows.Close()

	var records []types.AlertRecordInfo
	for rows.Next() {
		var record model.AmAlertRecord
		var modelId sql.NullInt64
		var modelName sql.NullString
		var message sql.NullString

		err = rows.Scan(
			&record.Id,
			&record.AlertType,
			&record.Threshold,
			&record.ActualValue,
			&message,
			&modelId,
			&modelName,
			&record.Status,
			&record.AlertTime,
		)
		if err != nil {
			continue
		}

		info := types.AlertRecordInfo{
			Id:          record.Id,
			AlertType:   record.AlertType,
			Threshold:   record.Threshold,
			ActualValue: record.ActualValue,
			Status:      record.Status,
			AlertTime:   record.AlertTime.Format("2006-01-02 15:04:05"),
		}

		if message.Valid {
			info.Message = message.String
		}
		if modelId.Valid {
			info.ModelId = modelId.Int64
		}
		if modelName.Valid {
			info.ModelName = modelName.String
		}

		records = append(records, info)
	}

	return &types.AlertRecordsResponse{
		BaseResponse: types.BaseResponse{
			Code:    200,
			Message: "success",
		},
		Data: types.AlertRecordsData{
			Records: records,
			Total:   total,
		},
	}, nil
}
