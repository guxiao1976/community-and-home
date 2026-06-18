package file

import (
	"context"

	filev1 "github.com/guxiao1976/api-proto/gen/go/file/v1"
	"github.com/guxiao1976/community-file/rpc/internal/svc"
	"github.com/guxiao1976/community-common/v2/pkg/errx"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"

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

func (l *DeleteFileLogic) DeleteFile(in *filev1.DeleteFileRequest) (*filev1.DeleteFileResponse, error) {
	f, err := l.svcCtx.FileModel.FindOne(l.ctx, in.FileId)
	if err != nil {
		return nil, errx.NewCodeError(70001, "file not found")
	}

	if err := l.svcCtx.MinioCli.Delete(l.ctx, f.FilePath); err != nil {
		l.Errorf("delete from minio failed: %v", err)
		// MinIO 删除失败不阻塞，记录日志继续 DB 软删除
	}

	if err := l.svcCtx.FileModel.Delete(l.ctx, in.FileId); err != nil {
		l.Errorf("delete file metadata failed: %v", err)
		return nil, errx.NewCodeError(70002, "delete file failed")
	}

	return &filev1.DeleteFileResponse{Base: responsex.NewBaseResp()}, nil
}
