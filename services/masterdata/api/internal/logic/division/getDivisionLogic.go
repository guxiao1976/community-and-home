// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package division

import (
	"context"

	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetDivisionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Get single division details
func NewGetDivisionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDivisionLogic {
	return &GetDivisionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetDivisionLogic) GetDivision(req *types.GetDivisionReq) (resp *types.GetDivisionResp, err error) {
	// todo: add your logic here and delete this line

	return
}
