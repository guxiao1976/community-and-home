package deleteditems

import (
	"context"
	"errors"

	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type RestoreDeletedItemLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRestoreDeletedItemLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RestoreDeletedItemLogic {
	return &RestoreDeletedItemLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RestoreDeletedItemLogic) RestoreDeletedItem(req *types.RestoreDeletedItemReq) (resp *types.RestoreDeletedItemResp, err error) {
	switch req.EntityType {
	case "residential_area":
		err = l.svcCtx.MdResidentialAreaModel.Restore(l.ctx, req.Id)
	case "administrative_division":
		err = l.svcCtx.MdAdministrativeDivisionModel.Restore(l.ctx, req.Id)
	case "configuration":
		err = l.svcCtx.MdConfigurationModel.Restore(l.ctx, req.Id)
	case "sensitive_word":
		err = l.svcCtx.MdSensitiveWordModel.Restore(l.ctx, req.Id)
	default:
		return nil, errors.New("invalid entity type: " + req.EntityType)
	}

	if err != nil {
		return nil, err
	}

	return &types.RestoreDeletedItemResp{Success: true}, nil
}
