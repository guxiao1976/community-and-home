// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package auth

import (
	"context"

	"github.com/guxiao/community-and-home/services/identity/api/internal/svc"
	"github.com/guxiao/community-and-home/services/identity/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginSmsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// User login with SMS code
func NewLoginSmsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginSmsLogic {
	return &LoginSmsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginSmsLogic) LoginSms(req *types.LoginSmsReq) (resp *types.LoginSmsResp, err error) {
	// todo: add your logic here and delete this line

	return
}
