// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package verification

import (
	"context"
	"encoding/json"

	"github.com/guxiao/community-and-home/common/errorx"
	"github.com/guxiao/community-and-home/services/identity/api/internal/svc"
	"github.com/guxiao/community-and-home/services/identity/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetVerificationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Get verification details
func NewGetVerificationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetVerificationLogic {
	return &GetVerificationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetVerificationLogic) GetVerification(req *types.GetVerificationReq) (resp *types.GetVerificationResp, err error) {
	// Get user ID from context
	userId, ok := l.ctx.Value("userId").(int64)
	if !ok {
		return nil, errorx.NewDefaultError("unauthorized")
	}

	// Query verification
	verification, err := l.svcCtx.AuthHomeownerVerificationModel.FindOne(l.ctx, req.Id)
	if err != nil {
		logx.Errorf("failed to query verification: %v", err)
		return nil, errorx.NewDefaultError("verification not found")
	}

	// Check permission (user can only view their own verifications unless admin)
	if verification.UserId != userId {
		// TODO: Check if user is admin
		return nil, errorx.NewDefaultError("permission denied")
	}

	// Parse document URLs
	var documentUrls []string
	if verification.DocumentUrls != "" {
		if err := json.Unmarshal([]byte(verification.DocumentUrls), &documentUrls); err != nil {
			logx.Errorf("failed to unmarshal document URLs: %v", err)
			documentUrls = []string{}
		}
	}

	// Convert to response
	var reviewerId *int64
	if verification.ReviewerId.Valid {
		reviewerId = &verification.ReviewerId.Int64
	}

	reviewTime := ""
	if verification.ReviewTime.Valid {
		reviewTime = verification.ReviewTime.Time.Format("2006-01-02 15:04:05")
	}

	reviewNotes := ""
	if verification.ReviewNotes.Valid {
		reviewNotes = verification.ReviewNotes.String
	}

	return &types.GetVerificationResp{
		Verification: types.HomeownerVerification{
			Id:                 verification.Id,
			UserId:             verification.UserId,
			PropertyUnitId:     verification.PropertyUnitId,
			DocumentUrls:       verification.DocumentUrls,
			RealName:           verification.RealName,
			IdCardNumber:       verification.IdCardNumber,
			VerificationStatus: verification.VerificationStatus,
			ReviewerId:         reviewerId,
			ReviewTime:         reviewTime,
			ReviewNotes:        reviewNotes,
			SubmitTime:         verification.SubmitTime.Format("2006-01-02 15:04:05"),
			CreatedTime:        verification.CreatedTime.Format("2006-01-02 15:04:05"),
			UpdatedTime:        verification.UpdatedTime.Format("2006-01-02 15:04:05"),
		},
	}, nil
}
