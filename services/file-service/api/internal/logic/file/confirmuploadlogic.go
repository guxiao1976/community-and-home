package file

import (
	"context"
	"fmt"

	filev1 "github.com/guxiao1976/api-proto/gen/go/file/v1"
	"github.com/guxiao1976/community-file/api/internal/svc"
	"github.com/guxiao1976/community-file/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ConfirmUploadLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewConfirmUploadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ConfirmUploadLogic {
	return &ConfirmUploadLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ConfirmUploadLogic) ConfirmUpload(userId int64, req *types.ConfirmUploadReq) (*types.ConfirmUploadResp, error) {
	rpcReq := &filev1.ConfirmUploadRequest{
		UserId:     userId,
		ObjectKey:  req.ObjectKey,
		EntityType: req.EntityType,
		EntityId:   req.EntityId,
		FileName:   req.FileName,
		FileSize:   req.FileSize,
		MimeType:   req.MimeType,
	}

	rpcResp, err := l.svcCtx.FileRpc.Client().ConfirmUpload(l.ctx, rpcReq)
	if err != nil {
		l.Errorf("ConfirmUpload gRPC failed: %v", err)
		return nil, fmt.Errorf("confirm upload failed: %w", err)
	}

	return &types.ConfirmUploadResp{
		File: toFileInfo(rpcResp.File),
	}, nil
}
