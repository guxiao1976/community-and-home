// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package verification

import (
	"context"
	"encoding/json"
	"time"

	"github.com/guxiao/community-and-home/common/errorx"
	"github.com/guxiao/community-and-home/services/identity/api/internal/svc"
	"github.com/guxiao/community-and-home/services/identity/api/internal/types"
	"github.com/guxiao/community-and-home/services/identity/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type SubmitVerificationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Submit homeowner verification
func NewSubmitVerificationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubmitVerificationLogic {
	return &SubmitVerificationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SubmitVerificationLogic) SubmitVerification(req *types.SubmitVerificationReq) (resp *types.SubmitVerificationResp, err error) {
	// Get user ID from context (assuming JWT middleware sets this)
	userId, ok := l.ctx.Value("userId").(int64)
	if !ok {
		return nil, errorx.NewDefaultError("unauthorized")
	}

	// Validate document URLs
	if len(req.DocumentUrls) == 0 {
		return nil, errorx.NewDefaultError("document URLs are required")
	}
	if len(req.DocumentUrls) > 9 {
		return nil, errorx.NewDefaultError("maximum 9 documents allowed")
	}

	// Convert document URLs to JSON string
	documentUrlsJson, err := json.Marshal(req.DocumentUrls)
	if err != nil {
		logx.Errorf("failed to marshal document URLs: %v", err)
		return nil, errorx.NewDefaultError("invalid document URLs")
	}

	// Create verification record
	verification := &model.AuthHomeownerVerification{
		UserId:             userId,
		PropertyUnitId:     req.PropertyUnitId,
		DocumentUrls:       string(documentUrlsJson),
		RealName:           req.RealName,
		IdCardNumber:       req.IdCardNumber,
		VerificationStatus: 0, // Pending
		SubmitTime:         time.Now(),
		CreatedTime:        time.Now(),
		UpdatedTime:        time.Now(),
	}

	result, err := l.svcCtx.AuthHomeownerVerificationModel.Insert(l.ctx, verification)
	if err != nil {
		logx.Errorf("failed to insert verification: %v", err)
		return nil, errorx.NewDefaultError("failed to submit verification")
	}

	id, err := result.LastInsertId()
	if err != nil {
		logx.Errorf("failed to get last insert id: %v", err)
		return nil, errorx.NewDefaultError("failed to submit verification")
	}

	return &types.SubmitVerificationResp{
		Id: id,
	}, nil
}
