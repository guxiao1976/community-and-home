package file

import (
	"context"
	"io"
	"time"

	filev1 "github.com/guxiao1976/api-proto/gen/go/file/v1"
	"github.com/guxiao1976/community-common/v2/pkg/errx"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-file/model"
	"github.com/guxiao1976/community-file/rpc/internal/svc"

	gominio "github.com/minio/minio-go/v7"
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
	if l.svcCtx.RawMinio == nil {
		return nil, errx.NewCodeError(ErrCodeFileOperationFailed, "MinIO not available")
	}

	// L2：回读 MinIO 实际对象前若干字节做 magic-bytes 校验（REQ-CAS-3）
	obj, err := l.svcCtx.RawMinio.GetObject(l.ctx, l.svcCtx.Bucket, in.ObjectKey, gominio.GetObjectOptions{})
	if err != nil {
		l.Errorf("get object failed: %v", err)
		return nil, errx.NewCodeError(ErrCodeFileOperationFailed, "read uploaded object failed")
	}
	resp, err := l.confirmWithReader(in, obj)
	obj.Close()
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// confirmWithReader 执行 L2 嗅探校验 + 落库（reader 注入便于单测）。
func (l *ConfirmUploadLogic) confirmWithReader(in *filev1.ConfirmUploadRequest, r io.Reader) (*filev1.ConfirmUploadResponse, error) {
	sniffedExt, err := verifySniffedContent(in.FileName, r)
	if err != nil {
		return nil, err
	}

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
		FileType:   sniffedExt,
		Confirmed:  true,
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
