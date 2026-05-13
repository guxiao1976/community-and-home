package logic

import (
	"context"

	"community-and-home/services/ai-model/rpc/internal/svc"
	"community-and-home/services/ai-model/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetModelInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetModelInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetModelInfoLogic {
	return &GetModelInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetModelInfoLogic) GetModelInfo(in *pb.ModelInfoRequest) (*pb.ModelInfoResponse, error) {
	models := []*pb.ModelDetail{}

	// 文本模型
	if l.svcCtx.Config.Models.Text.Enabled {
		models = append(models, &pb.ModelDetail{
			Name:       l.svcCtx.Config.Models.Text.Name,
			Version:    l.svcCtx.Config.Models.Text.Version,
			Type:       "text",
			Parameters: 7_000_000_000,
			Capabilities: []string{
				"政治敏感检测",
				"色情低俗检测",
				"暴力恐怖检测",
				"违法犯罪检测",
				"人身攻击检测",
				"广告营销检测",
			},
		})
	}

	// 图片模型（预留）
	if l.svcCtx.Config.Models.Image.Enabled {
		models = append(models, &pb.ModelDetail{
			Name:       l.svcCtx.Config.Models.Image.Name,
			Version:    l.svcCtx.Config.Models.Image.Version,
			Type:       "image",
			Parameters: 0,
			Capabilities: []string{
				"色情图片检测",
				"暴力血腥检测",
				"违禁物品检测",
			},
		})
	}

	return &pb.ModelInfoResponse{
		Models: models,
	}, nil
}
