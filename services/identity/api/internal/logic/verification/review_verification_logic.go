// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package verification

import (
	"context"

	"github.com/guxiao/community-and-home/common/errorx"
	"github.com/guxiao/community-and-home/services/identity/api/internal/svc"
	"github.com/guxiao/community-and-home/services/identity/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ReviewVerificationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Review verification request (admin)
func NewReviewVerificationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReviewVerificationLogic {
	return &ReviewVerificationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ReviewVerificationLogic) ReviewVerification(req *types.ReviewVerificationReq) (resp *types.ReviewVerificationResp, err error) {
	// Get reviewer ID from context
	reviewerId, ok := l.ctx.Value("userId").(int64)
	if !ok {
		return nil, errorx.NewDefaultError("unauthorized")
	}

	// TODO: Check if user is admin

	// Validate status (1: approved, 2: rejected)
	if req.Status != 1 && req.Status != 2 {
		return nil, errorx.NewDefaultError("invalid status")
	}

	// Get verification
	verification, err := l.svcCtx.AuthHomeownerVerificationModel.FindOne(l.ctx, req.Id)
	if err != nil {
		logx.Errorf("failed to get verification: %v", err)
		return nil, errorx.NewDefaultError("verification not found")
	}

	// Check if already reviewed
	if verification.VerificationStatus != 0 {
		return nil, errorx.NewDefaultError("verification already reviewed")
	}

	// Update verification status
	err = l.svcCtx.AuthHomeownerVerificationModel.UpdateStatus(l.ctx, req.Id, req.Status, reviewerId, req.ReviewNotes)
	if err != nil {
		logx.Errorf("failed to update verification status: %v", err)
		return nil, errorx.NewDefaultError("failed to review verification")
	}

	// If approved, update user verification status
	if req.Status == 1 {
		user, err := l.svcCtx.AuthUserModel.FindOne(l.ctx, verification.UserId)
		if err != nil {
			logx.Errorf("failed to get user: %v", err)
		} else {
			user.VerificationStatus = 1 // Verified
			err = l.svcCtx.AuthUserModel.Update(l.ctx, user)
			if err != nil {
				logx.Errorf("failed to update user verification status: %v", err)
			}
		}
	}

	return &types.ReviewVerificationResp{
		Success: true,
	}, nil
}
