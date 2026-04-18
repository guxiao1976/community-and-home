// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package property

import (
	"context"
	"time"

	"github.com/guxiao/community-and-home/common/errorx"
	"github.com/guxiao/community-and-home/services/identity/api/internal/svc"
	"github.com/guxiao/community-and-home/services/identity/api/internal/types"
	"github.com/guxiao/community-and-home/services/identity/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type BindPropertyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Bind property to user
func NewBindPropertyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BindPropertyLogic {
	return &BindPropertyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BindPropertyLogic) BindProperty(req *types.BindPropertyReq) (resp *types.BindPropertyResp, err error) {
	// Get user ID from context
	userId, ok := l.ctx.Value("userId").(int64)
	if !ok {
		return nil, errorx.NewDefaultError("unauthorized")
	}

	// Check if user is verified
	user, err := l.svcCtx.AuthUserModel.FindOne(l.ctx, userId)
	if err != nil {
		logx.Errorf("failed to get user: %v", err)
		return nil, errorx.NewDefaultError("user not found")
	}

	if user.VerificationStatus != 1 {
		return nil, errorx.NewDefaultError("user not verified")
	}

	// Check if already bound
	existing, err := l.svcCtx.AuthPropertyBindingModel.FindByUserAndProperty(l.ctx, userId, req.PropertyUnitId)
	if err != nil && err != model.ErrNotFound {
		logx.Errorf("failed to check existing binding: %v", err)
		return nil, errorx.NewDefaultError("failed to bind property")
	}

	if existing != nil && existing.BindStatus == 1 {
		return nil, errorx.NewDefaultError("property already bound")
	}

	// Create binding
	now := time.Now()
	binding := &model.AuthPropertyBinding{
		UserId:         userId,
		PropertyUnitId: req.PropertyUnitId,
		IsPrimary:      0, // Not primary by default
		BindStatus:     1, // Active
		BindTime:       now,
		CreatedTime:    now,
		UpdatedTime:    now,
	}

	result, err := l.svcCtx.AuthPropertyBindingModel.Insert(l.ctx, binding)
	if err != nil {
		logx.Errorf("failed to insert binding: %v", err)
		return nil, errorx.NewDefaultError("failed to bind property")
	}

	id, err := result.LastInsertId()
	if err != nil {
		logx.Errorf("failed to get last insert id: %v", err)
		return nil, errorx.NewDefaultError("failed to bind property")
	}

	return &types.BindPropertyResp{
		Id: id,
	}, nil
}
