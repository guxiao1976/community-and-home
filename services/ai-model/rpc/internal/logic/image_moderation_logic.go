package logic

import (
	"context"

	"community-and-home/services/ai-model/rpc/internal/svc"
	"community-and-home/services/ai-model/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ImageModerationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewImageModerationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ImageModerationLogic {
	return &ImageModerationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ModerateImage 图片审核（预留接口）
func (l *ImageModerationLogic) ModerateImage(in *pb.ImageModerationRequest) (*pb.ImageModerationResponse, error) {
	// TODO: 实现图片审核逻辑
	l.Info("Image moderation not implemented yet")

	return &pb.ImageModerationResponse{
		IsSafe:       true,
		RiskLevel:    "safe",
		Categories:   []string{},
		Reason:       "图片审核功能暂未实现",
		Confidence:   0.0,
		LatencyMs:    0,
		ModelVersion: "none",
	}, nil
}
