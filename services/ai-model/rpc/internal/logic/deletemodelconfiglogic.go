package logic

import (
	"context"
	"fmt"
	"time"

	"github.com/guxiao/community-and-home/services/ai-model/rpc/internal/svc"
	"github.com/guxiao/community-and-home/services/ai-model/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteModelConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteModelConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteModelConfigLogic {
	return &DeleteModelConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteModelConfigLogic) DeleteModelConfig(in *pb.DeleteModelConfigReq) (*pb.DeleteModelConfigResp, error) {
	// 1. 查询模型配置是否存在
	modelConfig, err := l.svcCtx.ModelConfigModel.FindOne(l.ctx, in.Id)
	if err != nil {
		l.Errorf("Failed to find model config: %v", err)
		return nil, fmt.Errorf("模型配置不存在")
	}

	// 2. 软删除：设置 delete_time
	modelConfig.DeleteTime.Valid = true
	modelConfig.DeleteTime.Time = time.Now()

	// 3. 更新数据库
	err = l.svcCtx.ModelConfigModel.Update(l.ctx, modelConfig)
	if err != nil {
		l.Errorf("Failed to delete model config: %v", err)
		return nil, fmt.Errorf("删除模型配置失败")
	}

	l.Infof("Model config %d deleted successfully", in.Id)

	return &pb.DeleteModelConfigResp{}, nil
}
