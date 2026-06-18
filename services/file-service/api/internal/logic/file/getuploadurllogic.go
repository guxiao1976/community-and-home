package file

import (
	"context"
	"fmt"

	filev1 "github.com/guxiao1976/api-proto/gen/go/file/v1"
	"github.com/guxiao1976/community-file/api/internal/svc"
	"github.com/guxiao1976/community-file/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUploadUrlLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUploadUrlLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUploadUrlLogic {
	return &GetUploadUrlLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUploadUrlLogic) GetUploadUrl(userId int64, req *types.GetUploadUrlReq) (*types.GetUploadUrlResp, error) {
	rpcReq := &filev1.GetUploadUrlRequest{
		UserId:     userId,
		EntityType: req.EntityType,
		FileName:   req.FileName,
		MimeType:   req.MimeType,
		FileSize:   req.FileSize,
	}

	rpcResp, err := l.svcCtx.FileRpc.Client().GetUploadUrl(l.ctx, rpcReq)
	if err != nil {
		l.Errorf("GetUploadUrl gRPC failed: %v", err)
		return nil, fmt.Errorf("generate upload URL failed: %w", err)
	}

	return &types.GetUploadUrlResp{
		UploadUrl: rpcResp.UploadUrl,
		ObjectKey: rpcResp.ObjectKey,
		ExpireAt:  rpcResp.ExpireAt,
	}, nil
}
