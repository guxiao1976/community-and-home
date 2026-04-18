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

type GetVerificationListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// List all verification requests (admin)
func NewGetVerificationListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetVerificationListLogic {
	return &GetVerificationListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetVerificationListLogic) GetVerificationList(req *types.GetVerificationListReq) (resp *types.GetVerificationListResp, err error) {
	// TODO: Check if user is admin

	// Set default pagination
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	// Get verifications
	verifications, err := l.svcCtx.AuthHomeownerVerificationModel.FindAll(l.ctx, page, pageSize, req.Status, nil)
	if err != nil {
		logx.Errorf("failed to get verifications: %v", err)
		return nil, errorx.NewDefaultError("failed to get verifications")
	}

	// Get total count
	total, err := l.svcCtx.AuthHomeownerVerificationModel.CountAll(l.ctx, req.Status, nil)
	if err != nil {
		logx.Errorf("failed to count verifications: %v", err)
		return nil, errorx.NewDefaultError("failed to get verifications")
	}

	// Convert to response
	list := make([]types.HomeownerVerification, 0, len(verifications))
	for _, v := range verifications {
		var reviewerId *int64
		if v.ReviewerId.Valid {
			reviewerId = &v.ReviewerId.Int64
		}

		reviewTime := ""
		if v.ReviewTime.Valid {
			reviewTime = v.ReviewTime.Time.Format("2006-01-02 15:04:05")
		}

		reviewNotes := ""
		if v.ReviewNotes.Valid {
			reviewNotes = v.ReviewNotes.String
		}

		list = append(list, types.HomeownerVerification{
			Id:                 v.Id,
			UserId:             v.UserId,
			PropertyUnitId:     v.PropertyUnitId,
			DocumentUrls:       v.DocumentUrls,
			RealName:           v.RealName,
			IdCardNumber:       v.IdCardNumber,
			VerificationStatus: v.VerificationStatus,
			ReviewerId:         reviewerId,
			ReviewTime:         reviewTime,
			ReviewNotes:        reviewNotes,
			SubmitTime:         v.SubmitTime.Format("2006-01-02 15:04:05"),
			CreatedTime:        v.CreatedTime.Format("2006-01-02 15:04:05"),
			UpdatedTime:        v.UpdatedTime.Format("2006-01-02 15:04:05"),
		})
	}

	return &types.GetVerificationListResp{
		List:  list,
		Total: total,
	}, nil
}
