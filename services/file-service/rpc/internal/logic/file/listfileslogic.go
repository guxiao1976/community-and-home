package file

import (
	"context"

	filev1 "github.com/guxiao1976/api-proto/gen/go/file/v1"
	commonv1 "github.com/guxiao1976/api-proto/gen/go/common/v1"
	"github.com/guxiao1976/community-file/rpc/internal/svc"
	"github.com/guxiao1976/community-common/v2/pkg/errx"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"

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

func (l *ListFilesLogic) ListFiles(in *filev1.ListFilesRequest) (*filev1.ListFilesResponse, error) {
	page := in.Page.GetPage()
	if page < 1 {
		page = 1
	}
	pageSize := in.Page.GetPageSize()
	maxPageSize := int32(100)
	if l.svcCtx.SysConfig != nil {
		if v, err := l.svcCtx.SysConfig.GetInt(l.ctx, "file.list.max_page_size"); err == nil {
			maxPageSize = int32(v)
		}
	}
	defaultPageSize := int32(20)
	if l.svcCtx.SysConfig != nil {
		if v, err := l.svcCtx.SysConfig.GetInt(l.ctx, "file.list.default_page_size"); err == nil {
			defaultPageSize = int32(v)
		}
	}
	if pageSize < 1 || pageSize > maxPageSize {
		pageSize = defaultPageSize
	}

	var userID *int64
	if in.UserId != nil {
		userID = in.UserId
	}
	var entityType *string
	if in.EntityType != nil && *in.EntityType != "" {
		entityType = in.EntityType
	}
	var entityID *int64
	if in.EntityId != nil {
		entityID = in.EntityId
	}

	files, total, err := l.svcCtx.FileModel.FindPage(l.ctx, userID, entityType, entityID, int64(page), int64(pageSize))
	if err != nil {
		l.Errorf("list files failed: %v", err)
		return nil, errx.NewCodeError(70002, "query files failed")
	}

	totalPages := int32(total / int64(pageSize))
	if total%int64(pageSize) > 0 {
		totalPages++
	}

	pbFiles := make([]*filev1.FileInfo, 0, len(files))
	for _, f := range files {
		pbFiles = append(pbFiles, toProtoFile(f))
	}

	return &filev1.ListFilesResponse{
		Base:  responsex.NewBaseResp(),
		Files: pbFiles,
		Page: &commonv1.PageResponse{
			Page:       int32(page),
			PageSize:   int32(pageSize),
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}
