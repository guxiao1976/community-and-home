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

type UpdateDivisionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateDivisionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateDivisionLogic {
	return &UpdateDivisionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateDivisionLogic) UpdateDivision(in *pb.UpdateDivisionReq) (*pb.UpdateDivisionResp, error) {
	existing, err := l.svcCtx.MdAdministrativeDivisionModel.FindOne(l.ctx, in.Id)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, fmt.Errorf("division not found")
		}
		return nil, err
	}
	if existing.DeleteTime.Valid {
		return nil, fmt.Errorf("division is deleted")
	}

	var parentPath string
	var parentId sql.NullInt64

	if in.ParentId > 0 {
		if in.ParentId == in.Id {
			return nil, fmt.Errorf("division cannot be its own parent")
		}
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

	if in.Code != existing.Code {
		codeCheck, err := l.svcCtx.MdAdministrativeDivisionModel.FindOneByCode(l.ctx, in.Code)
		if err != nil && err != model.ErrNotFound {
			return nil, err
		}
		if codeCheck != nil && !codeCheck.DeleteTime.Valid {
			return nil, fmt.Errorf("division code already exists")
		}
	}

	oldPath := existing.Path
	oldParentId := existing.ParentId

	existing.ParentId = parentId
	existing.Level = int64(in.Level)
	existing.Name = in.Name
	existing.Code = in.Code
	existing.Path = fmt.Sprintf("%s%d/", parentPath, in.Id)
	existing.SortOrder = int64(in.SortOrder)
	existing.Status = int64(in.Status)

	err = l.svcCtx.MdAdministrativeDivisionModel.Update(l.ctx, existing)
	if err != nil {
		return nil, err
	}

	// If parent changed, update all descendants' paths
	if oldParentId != parentId {
		err = l.svcCtx.MdAdministrativeDivisionModel.UpdatePathForDescendants(l.ctx, oldPath, existing.Path)
		if err != nil {
			return nil, err
		}
	}

	return &pb.UpdateDivisionResp{}, nil
}
