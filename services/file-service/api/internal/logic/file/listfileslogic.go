package file

import (
	"context"
	"fmt"

	commonv1 "github.com/guxiao1976/api-proto/gen/go/common/v1"
	filev1 "github.com/guxiao1976/api-proto/gen/go/file/v1"
	"github.com/guxiao1976/community-file/api/internal/svc"
	"github.com/guxiao1976/community-file/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListFilesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListFilesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListFilesLogic {
	return &ListFilesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListFilesLogic) ListFiles(userId int64, req *types.ListFilesReq) (*types.ListFilesResp, error) {
	// 权限控制：强制使用认证用户的 ID，防止查询其他用户的文件
	rpcReq := &filev1.ListFilesRequest{
		Page: &commonv1.PageRequest{
			Page:     req.Page,
			PageSize: req.PageSize,
		},
		UserId:     &userId,
		EntityType: req.EntityType,
		EntityId:   req.EntityId,
	}

	rpcResp, err := l.svcCtx.FileRpc.Client().ListFiles(l.ctx, rpcReq)
	if err != nil {
		l.Errorf("ListFiles gRPC failed: %v", err)
		return nil, fmt.Errorf("list files failed: %w", err)
	}

	files := make([]*types.FileInfo, 0, len(rpcResp.Files))
	for _, f := range rpcResp.Files {
		files = append(files, toFileInfo(f))
	}

	resp := &types.ListFilesResp{
		Files: files,
	}
	if rpcResp.Page != nil {
		resp.Page = &types.PageInfo{
			Page:       rpcResp.Page.Page,
			PageSize:   rpcResp.Page.PageSize,
			Total:      rpcResp.Page.Total,
			TotalPages: rpcResp.Page.TotalPages,
		}
	}

	return resp, nil
}
