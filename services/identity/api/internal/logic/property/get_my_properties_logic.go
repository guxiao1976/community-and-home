// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package property

import (
	"context"
	"database/sql"

	"github.com/guxiao/community-and-home/common/errorx"
	"github.com/guxiao/community-and-home/services/identity/api/internal/svc"
	"github.com/guxiao/community-and-home/services/identity/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMyPropertiesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Get my properties
func NewGetMyPropertiesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMyPropertiesLogic {
	return &GetMyPropertiesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMyPropertiesLogic) GetMyProperties(req *types.GetMyPropertiesReq) (resp *types.GetMyPropertiesResp, err error) {
	// Get user ID from context
	userId, ok := l.ctx.Value("userId").(int64)
	if !ok {
		return nil, errorx.NewDefaultError("unauthorized")
	}

	// Get bindings
	var bindings []types.PropertyBinding
	if req.Status != nil {
		// Filter by status
		list, err := l.svcCtx.AuthPropertyBindingModel.FindByUserId(l.ctx, userId)
		if err != nil {
			logx.Errorf("failed to get bindings: %v", err)
			return nil, errorx.NewDefaultError("failed to get properties")
		}

		for _, b := range list {
			if b.BindStatus == int64(*req.Status) {
				bindings = append(bindings, types.PropertyBinding{
					Id:             b.Id,
					UserId:         b.UserId,
					PropertyUnitId: b.PropertyUnitId,
					IsPrimary:      b.IsPrimary == 1,
					BindStatus:     b.BindStatus,
					BindTime:       b.BindTime.Format("2006-01-02 15:04:05"),
					RevokeTime:     formatNullTime(b.RevokeTime),
					RevokedBy:      formatNullInt64(b.RevokedBy),
					CreatedTime:    b.CreatedTime.Format("2006-01-02 15:04:05"),
					UpdatedTime:    b.UpdatedTime.Format("2006-01-02 15:04:05"),
				})
			}
		}
	} else {
		// Get all active bindings
		list, err := l.svcCtx.AuthPropertyBindingModel.FindActiveByUserId(l.ctx, userId)
		if err != nil {
			logx.Errorf("failed to get bindings: %v", err)
			return nil, errorx.NewDefaultError("failed to get properties")
		}

		for _, b := range list {
			bindings = append(bindings, types.PropertyBinding{
				Id:             b.Id,
				UserId:         b.UserId,
				PropertyUnitId: b.PropertyUnitId,
				IsPrimary:      b.IsPrimary == 1,
				BindStatus:     b.BindStatus,
				BindTime:       b.BindTime.Format("2006-01-02 15:04:05"),
				RevokeTime:     formatNullTime(b.RevokeTime),
				RevokedBy:      formatNullInt64(b.RevokedBy),
				CreatedTime:    b.CreatedTime.Format("2006-01-02 15:04:05"),
				UpdatedTime:    b.UpdatedTime.Format("2006-01-02 15:04:05"),
			})
		}
	}

	return &types.GetMyPropertiesResp{
		List:  bindings,
		Total: int64(len(bindings)),
	}, nil
}

func formatNullTime(t sql.NullTime) string {
	if t.Valid {
		return t.Time.Format("2006-01-02 15:04:05")
	}
	return ""
}

func formatNullInt64(i sql.NullInt64) int64 {
	if i.Valid {
		return i.Int64
	}
	return 0
}
