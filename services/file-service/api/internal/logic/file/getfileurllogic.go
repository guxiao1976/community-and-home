package file

import (
	"context"
	"fmt"

	filev1 "github.com/guxiao1976/api-proto/gen/go/file/v1"
	"github.com/guxiao1976/community-common/v2/pkg/errx"
	"github.com/guxiao1976/community-file/api/internal/svc"
	"github.com/guxiao1976/community-file/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetFileUrlLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetFileUrlLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFileUrlLogic {
	return &GetFileUrlLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetFileUrlLogic) GetFileUrl(userId int64, req *types.GetFileUrlReq) (*types.GetFileUrlResp, error) {
	rpcReq := &filev1.GetFileUrlRequest{
		FileId: req.Id,
	}

	rpcResp, err := l.svcCtx.FileRpc.Client().GetFileUrl(l.ctx, rpcReq)
	if err != nil {
		l.Errorf("GetFileUrl gRPC failed: %v", err)
		return nil, fmt.Errorf("get file URL failed: %w", err)
	}

	// 权限检查：仅文件所有者可以获取下载 URL
	if rpcResp.File != nil && rpcResp.File.UserId != userId {
		return nil, errx.NewCodeError(ErrCodeFileOperationFailed, "permission denied: not the file owner")
	}

	return &types.GetFileUrlResp{
		DownloadUrl: rpcResp.DownloadUrl,
		File:        toFileInfo(rpcResp.File),
	}, nil
}
