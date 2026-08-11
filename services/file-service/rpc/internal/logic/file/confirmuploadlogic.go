package file

import (
	"context"
	"time"

	filev1 "github.com/guxiao1976/api-proto/gen/go/file/v1"
	"github.com/guxiao1976/community-common/v2/pkg/errx"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-file/model"
	"github.com/guxiao1976/community-file/rpc/internal/svc"

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

func (l *ConfirmUploadLogic) ConfirmUpload(in *filev1.ConfirmUploadRequest) (*filev1.ConfirmUploadResponse, error) {
	f := &model.File{
		UserID:     in.UserId,
		EntityType: in.EntityType,
		EntityID:   in.EntityId,
		FileName:   in.FileName,
		FilePath:   in.ObjectKey,
		FileSize:   in.FileSize,
		MimeType:   in.MimeType,
		BucketName: l.svcCtx.Bucket,
		UploadTime: time.Now(),
	}

	id, err := l.svcCtx.FileModel.Insert(l.ctx, f)
	if err != nil {
		l.Errorf("insert file record failed: %v", err)
		return nil, errx.NewCodeError(ErrCodeFileAccessDenied, "save file metadata failed")
	}

	f.ID = id
	return &filev1.ConfirmUploadResponse{
		Base: responsex.NewBaseResp(),
		File: toProtoFile(f),
	}, nil
}
