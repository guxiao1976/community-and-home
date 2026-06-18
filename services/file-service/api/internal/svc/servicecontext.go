package svc

import (
	"github.com/guxiao1976/community-file/api/internal/config"
	fileclient "github.com/guxiao1976/community-file/rpc/file"
)

type ServiceContext struct {
	Config       config.Config
	FileRpc       *fileclient.FileServiceClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 通过 etcd 发现 file.rpc 并建立 gRPC 连接
	fileRpc, err := fileclient.NewFileServiceClient(c.FileRpc)
	if err != nil {
		panic("failed to connect file.rpc: " + err.Error())
	}

	return &ServiceContext{
		Config: c,
		FileRpc: fileRpc,
	}
}
