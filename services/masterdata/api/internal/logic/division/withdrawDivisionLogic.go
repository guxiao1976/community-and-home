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

type WithdrawDivisionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWithdrawDivisionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WithdrawDivisionLogic {
	return &WithdrawDivisionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *WithdrawDivisionLogic) WithdrawDivision(req *types.SubmitReq) (resp *types.SubmitResp, err error) {
	div, err := l.svcCtx.MdAdministrativeDivisionModel.FindOne(l.ctx, req.Id)
	if err != nil {
		return nil, err
	}
	if div.SubmissionStatus != 1 {
		return nil, errors.New("仅已提交状态的数据可以撤回")
	}
	if err := l.svcCtx.MdAdministrativeDivisionModel.UpdateStatus(l.ctx, req.Id, 0); err != nil {
		return nil, err
	}

	_ = l.svcCtx.SubmissionRecordModel.UpdateResultByEntity(l.ctx, "administrative_division", req.Id, 0, 3, "撤回")

	return &types.SubmitResp{Success: true}, nil
}
