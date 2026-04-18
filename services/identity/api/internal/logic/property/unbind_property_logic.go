// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package property

import (
	"context"
	"database/sql"
	"time"

	"github.com/guxiao/community-and-home/common/errorx"
	"github.com/guxiao/community-and-home/services/identity/api/internal/svc"
	"github.com/guxiao/community-and-home/services/identity/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UnbindPropertyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Unbind property from user
func NewUnbindPropertyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnbindPropertyLogic {
	return &UnbindPropertyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UnbindPropertyLogic) UnbindProperty(req *types.UnbindPropertyReq) (resp *types.UnbindPropertyResp, err error) {
	// Get user ID from context
	userId, ok := l.ctx.Value("userId").(int64)
	if !ok {
		return nil, errorx.NewDefaultError("unauthorized")
	}

	// Get binding
	binding, err := l.svcCtx.AuthPropertyBindingModel.FindOne(l.ctx, req.Id)
	if err != nil {
		logx.Errorf("failed to get binding: %v", err)
		return nil, errorx.NewDefaultError("binding not found")
	}

	// Check permission
	if binding.UserId != userId {
		return nil, errorx.NewDefaultError("permission denied")
	}

	// Update binding status
	binding.BindStatus = 0 // Inactive
	binding.RevokeTime = sql.NullTime{Time: time.Now(), Valid: true}
	binding.UpdatedTime = time.Now()

	err = l.svcCtx.AuthPropertyBindingModel.Update(l.ctx, binding)
	if err != nil {
		logx.Errorf("failed to update binding: %v", err)
		return nil, errorx.NewDefaultError("failed to unbind property")
	}

	return &types.UnbindPropertyResp{
		Success: true,
	}, nil
}
