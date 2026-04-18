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

func NewCreateDivisionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateDivisionLogic {
	return &CreateDivisionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateDivisionLogic) CreateDivision(req *types.CreateDivisionReq) (resp *types.CreateDivisionResp, err error) {
	// 1. Check if code already exists
	_, err = l.svcCtx.MdAdministrativeDivisionModel.FindOneByCode(l.ctx, req.Code)
	if err == nil {
		return nil, errorx.NewInvalidParamError("行政代码已存在")
	}

	// 2. Validate parent and calculate path
	path := ""
	if req.ParentId != nil {
		parent, err := l.svcCtx.MdAdministrativeDivisionModel.FindOne(l.ctx, *req.ParentId)
		if err != nil {
			return nil, errorx.NewInvalidParamError("上级行政区域不存在")
		}
		if parent.Level+1 != int64(req.Level) {
			return nil, errorx.NewInvalidParamError("层级与父级不匹配")
		}
		path = fmt.Sprintf("%s%d/", parent.Path, parent.Id)
	} else {
		if req.Level != 1 {
			return nil, errorx.NewInvalidParamError("顶级行政区域层级必须为1")
		}
		path = "/"
	}

	// 3. Create model
	data := &model.MdAdministrativeDivision{
		Level:       int64(req.Level),
		Name:        req.Name,
		Code:        req.Code,
		Path:        path,
		SortOrder:   int64(req.SortOrder),
		Status:      1,
		CreatedBy:   0, // TODO: Get from JWT
		CreatedTime: time.Now(),
		UpdatedTime: time.Now(),
	}
	if req.ParentId != nil {
		data.ParentId = sql.NullInt64{Int64: *req.ParentId, Valid: true}
	}

	res, err := l.svcCtx.MdAdministrativeDivisionModel.Insert(l.ctx, data)
	if err != nil {
		return nil, errorx.NewDefaultError("创建行政区域失败")
	}

	id, _ := res.LastInsertId()
	return &types.CreateDivisionResp{Id: id}, nil
}
