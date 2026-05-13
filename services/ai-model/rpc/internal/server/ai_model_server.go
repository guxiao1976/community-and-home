package server

import (
	"context"

	"community-and-home/services/ai-model/rpc/internal/logic"
	"community-and-home/services/ai-model/rpc/internal/svc"
	"community-and-home/services/ai-model/rpc/pb"
)

type AiModelServer struct {
	svcCtx *svc.ServiceContext
	pb.UnimplementedAiModelServer
}

func NewAiModelServer(svcCtx *svc.ServiceContext) *AiModelServer {
	return &AiModelServer{
		svcCtx: svcCtx,
	}
}

func (s *AiModelServer) ModerateText(ctx context.Context, in *pb.TextModerationRequest) (*pb.TextModerationResponse, error) {
	l := logic.NewTextModerationLogic(ctx, s.svcCtx)
	return l.ModerateText(in)
}

func (s *AiModelServer) ModerateImage(ctx context.Context, in *pb.ImageModerationRequest) (*pb.ImageModerationResponse, error) {
	l := logic.NewImageModerationLogic(ctx, s.svcCtx)
	return l.ModerateImage(in)
}

func (s *AiModelServer) HealthCheck(ctx context.Context, in *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	l := logic.NewHealthCheckLogic(ctx, s.svcCtx)
	return l.HealthCheck(in)
}

func (s *AiModelServer) GetModelInfo(ctx context.Context, in *pb.ModelInfoRequest) (*pb.ModelInfoResponse, error) {
	l := logic.NewGetModelInfoLogic(ctx, s.svcCtx)
	return l.GetModelInfo(in)
}
