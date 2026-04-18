// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package file

import (
	"context"

	"github.com/guxiao/community-and-home/services/identity/api/internal/svc"
	"github.com/guxiao/community-and-home/services/identity/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetFileUrlLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Get file URL
func NewGetFileUrlLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFileUrlLogic {
	return &GetFileUrlLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetFileUrlLogic) GetFileUrl(req *types.GetFileUrlReq) (resp *types.GetFileUrlResp, err error) {
	// Get file record from database
	file, err := l.svcCtx.AuthUploadedFileModel.FindOne(l.ctx, req.FileId)
	if err != nil {
		logx.Errorf("failed to get file: %v", err)
		return nil, err
	}

	return &types.GetFileUrlResp{
		FileUrl: file.FilePath,
	}, nil
}
