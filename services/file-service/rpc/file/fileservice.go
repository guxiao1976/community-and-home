package file

import (
	filev1 "github.com/guxiao1976/api-proto/gen/go/file/v1"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

type FileServiceClient struct {
	cli filev1.FileServiceClient
}

func NewFileServiceClient(c zrpc.RpcClientConf) (*FileServiceClient, error) {
	conn := zrpc.MustNewClient(c)
	return &FileServiceClient{cli: filev1.NewFileServiceClient(conn.Conn())}, nil
}

func (c *FileServiceClient) Client() filev1.FileServiceClient {
	return c.cli
}

func (c *FileServiceClient) Conn() *grpc.ClientConn {
	return c.cli.(interface{ Conn() *grpc.ClientConn }).Conn()
}
