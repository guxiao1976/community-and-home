package logic

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/guxiao/community-and-home/services/masterdata/model"
	"github.com/guxiao/community-and-home/services/masterdata/rpc/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateDivisionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateDivisionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateDivisionLogic {
	return &CreateDivisionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateDivisionLogic) CreateDivision(in *pb.CreateDivisionReq) (*pb.CreateDivisionResp, error) {
	var parentPath string
	var parentId sql.NullInt64

	if in.ParentId > 0 {
		parent, err := l.svcCtx.MdAdministrativeDivisionModel.FindOne(l.ctx, in.ParentId)
		if err != nil {
			if err == model.ErrNotFound {
				return nil, fmt.Errorf("parent division not found")
			}
			return nil, err
		}
		if parent.DeleteTime.Valid {
			return nil, fmt.Errorf("parent division is deleted")
		}
		parentPath = parent.Path
		parentId = sql.NullInt64{Int64: in.ParentId, Valid: true}
	} else {
		parentPath = ""
		parentId = sql.NullInt64{Valid: false}
	}

	existing, err := l.svcCtx.MdAdministrativeDivisionModel.FindOneByCode(l.ctx, in.Code)
	if err != nil && err != model.ErrNotFound {
		return nil, err
	}
	if existing != nil && !existing.DeleteTime.Valid {
		return nil, fmt.Errorf("division code already exists")
	}

	division := &model.MdAdministrativeDivision{
		ParentId:  parentId,
		Level:     int64(in.Level),
		Name:      in.Name,
		Code:      in.Code,
		Path:      "",
		SortOrder: int64(in.SortOrder),
		Status:    1, // Default to active
		CreatedBy: in.CreatedBy,
	}

	result, err := l.svcCtx.MdAdministrativeDivisionModel.Insert(l.ctx, division)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Update the path with the actual ID
	division.Id = id
	division.Path = fmt.Sprintf("%s%d/", parentPath, id)
	err = l.svcCtx.MdAdministrativeDivisionModel.Update(l.ctx, division)
	if err != nil {
		return nil, err
	}

	return &pb.CreateDivisionResp{
		Id: id,
	}, nil
}
