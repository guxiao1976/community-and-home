// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package division

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/guxiao/community-and-home/common/errorx"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"
	"github.com/guxiao/community-and-home/services/masterdata/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateDivisionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Create new administrative division
func NewCreateDivisionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateDivisionLogic {
	return &CreateDivisionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateDivisionLogic) CreateDivision(req *types.CreateDivisionReq) (resp *types.CreateDivisionResp, err error) {
	// Check if code already exists
	_, err = l.svcCtx.MdAdministrativeDivisionModel.FindOneByCode(l.ctx, req.Code)
	if err == nil {
		return nil, errorx.NewDefaultError("该区划代码已被使用，请重新输入")
	}

	// Get user ID from context (default to 1 if not found)
	userId := int64(1)
	if uid := l.ctx.Value("userId"); uid != nil {
		userId = uid.(int64)
	}

	// Build parent_id
	var parentId sql.NullInt64
	if req.ParentId > 0 {
		parentId = sql.NullInt64{Int64: req.ParentId, Valid: true}
	}

	// Calculate path
	path := "/"
	if parentId.Valid {
		parent, err := l.svcCtx.MdAdministrativeDivisionModel.FindOne(l.ctx, parentId.Int64)
		if err != nil {
			return nil, err
		}
		path = parent.Path
	}

	now := time.Now()

	// Insert division
	result, err := l.svcCtx.MdAdministrativeDivisionModel.Insert(l.ctx, &model.MdAdministrativeDivision{
		ParentId:    parentId,
		Level:       int64(req.Level),
		Name:        req.Name,
		Code:        req.Code,
		Path:        path,
		SortOrder:        int64(req.SortOrder),
			SubmissionType:    sql.NullInt64{Int64: 1, Valid: true},
			SubmissionStatus: 0,
		Status:      1,
		CreatedBy:   userId,
		CreatedTime: now,
		UpdatedTime: now,
	})
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Update path to include new ID
	newPath := fmt.Sprintf("%s%d/", path, id)
	division, err := l.svcCtx.MdAdministrativeDivisionModel.FindOne(l.ctx, id)
	if err != nil {
		return nil, err
	}
	division.Path = newPath
	err = l.svcCtx.MdAdministrativeDivisionModel.Update(l.ctx, division)
	if err != nil {
		return nil, err
	}

	return &types.CreateDivisionResp{
		Id: id,
	}, nil
}
