// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package property

import (
	"context"

	"community-and-home/services/identity/api/internal/svc"
	"community-and-home/services/identity/api/internal/types"

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
	// todo: add your logic here and delete this line

	return
}
