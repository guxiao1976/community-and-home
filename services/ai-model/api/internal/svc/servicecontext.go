// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package svc

import (
	"github.com/guxiao/community-and-home/services/ai-model/api/internal/config"
	"github.com/guxiao/community-and-home/services/ai-model/rpc/aimodel"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config     config.Config
	AiModelRpc aimodel.AiModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:     c,
		AiModelRpc: aimodel.NewAiModel(zrpc.MustNewClient(c.AiModelRpc)),
	}
}
