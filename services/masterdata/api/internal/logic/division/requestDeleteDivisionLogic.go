// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package division

import (
	"context"
	"errors"

	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RequestDeleteDivisionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRequestDeleteDivisionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RequestDeleteDivisionLogic {
	return &RequestDeleteDivisionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RequestDeleteDivisionLogic) RequestDeleteDivision(req *types.SubmitReq) (resp *types.SubmitResp, err error) {
	div, err := l.svcCtx.MdAdministrativeDivisionModel.FindOne(l.ctx, req.Id)
	if err != nil {
		return nil, err
	}
	if div.SubmissionStatus != 2 {
		return nil, errors.New("仅已批准状态的数据可以发起删除")
	}
	if err := checkChildData(l.ctx, l.svcCtx, req.Id, div.Level); err != nil {
		return nil, err
	}
	if err := l.svcCtx.MdAdministrativeDivisionModel.UpdateStatusAndType(l.ctx, req.Id, 0, 3); err != nil {
		return nil, err
	}
	return &types.SubmitResp{Success: true}, nil
}
