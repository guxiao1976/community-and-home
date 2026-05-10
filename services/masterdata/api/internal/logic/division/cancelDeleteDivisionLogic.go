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

type CancelDeleteDivisionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCancelDeleteDivisionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelDeleteDivisionLogic {
	return &CancelDeleteDivisionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CancelDeleteDivisionLogic) CancelDeleteDivision(req *types.SubmitReq) (resp *types.SubmitResp, err error) {
	div, err := l.svcCtx.MdAdministrativeDivisionModel.FindOne(l.ctx, req.Id)
	if err != nil {
		return nil, err
	}
	if div.SubmissionStatus != 0 || !div.SubmissionType.Valid || div.SubmissionType.Int64 != 3 {
		return nil, errors.New("仅删除待提交状态的数据可以取消删除")
	}
	if err := l.svcCtx.MdAdministrativeDivisionModel.UpdateStatusAndType(l.ctx, req.Id, 2, 0); err != nil {
		return nil, err
	}
	return &types.SubmitResp{Success: true}, nil
}
