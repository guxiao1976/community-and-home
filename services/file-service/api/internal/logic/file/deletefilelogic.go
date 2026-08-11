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

type DeleteFileLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteFileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteFileLogic {
	return &DeleteFileLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteFileLogic) DeleteFile(userId int64, req *types.DeleteFileReq) error {
	// 权限检查：先获取文件元数据，验证所有权
	fileReq := &filev1.GetFileUrlRequest{
		FileId: req.Id,
	}
	fileResp, err := l.svcCtx.FileRpc.Client().GetFileUrl(l.ctx, fileReq)
	if err != nil {
		l.Errorf("get file metadata failed: %v", err)
		return fmt.Errorf("file not found")
	}

	if fileResp.File != nil && fileResp.File.UserId != userId {
		return errx.NewCodeError(ErrCodeFileOperationFailed, "permission denied: not the file owner")
	}

	rpcReq := &filev1.DeleteFileRequest{
		FileId: req.Id,
	}

	_, err = l.svcCtx.FileRpc.Client().DeleteFile(l.ctx, rpcReq)
	if err != nil {
		l.Errorf("DeleteFile gRPC failed: %v", err)
		return fmt.Errorf("delete file failed: %w", err)
	}

	return nil
}
