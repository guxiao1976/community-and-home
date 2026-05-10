package logic

import (
	"context"

	"github.com/guxiao/community-and-home/services/identity/rpc/internal/svc"
	"github.com/guxiao/community-and-home/services/identity/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserVerificationStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserVerificationStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserVerificationStatusLogic {
	return &GetUserVerificationStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Verification and property related methods
func (l *GetUserVerificationStatusLogic) GetUserVerificationStatus(in *pb.GetUserVerificationStatusReq) (*pb.GetUserVerificationStatusResp, error) {
	verification, err := l.svcCtx.AuthHomeownerVerificationModel.FindOneByUserIdPropertyUnitId(l.ctx, in.UserId, in.PropertyUnitId)
	if err != nil {
		return &pb.GetUserVerificationStatusResp{}, nil
	}

	resp := &pb.GetUserVerificationStatusResp{
		VerificationId:     verification.Id,
		VerificationStatus: int32(verification.VerificationStatus),
		VerificationType:   "homeowner",
		SubmittedTime:      verification.SubmitTime.Format("2006-01-02 15:04:05"),
	}

	if verification.ReviewTime.Valid {
		resp.ReviewedTime = verification.ReviewTime.Time.Format("2006-01-02 15:04:05")
	}
	if verification.ReviewNotes.Valid {
		resp.ReviewerNotes = verification.ReviewNotes.String
	}

	return resp, nil
}
