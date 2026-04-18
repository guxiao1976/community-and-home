// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package file

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/guxiao/community-and-home/common/errorx"
	"github.com/guxiao/community-and-home/services/identity/api/internal/svc"
	"github.com/guxiao/community-and-home/services/identity/api/internal/types"
	"github.com/guxiao/community-and-home/services/identity/model"
	"github.com/minio/minio-go/v7"
	"github.com/zeromicro/go-zero/core/logx"
)

type UploadFileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

// Upload file to MinIO
func NewUploadFileLogic(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *UploadFileLogic {
	return &UploadFileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

func (l *UploadFileLogic) UploadFile(req *types.UploadFileReq) (resp *types.UploadFileResp, err error) {
	// Get user ID from context
	userId, ok := l.ctx.Value("userId").(int64)
	if !ok {
		return nil, errorx.NewDefaultError("unauthorized")
	}

	// Parse multipart form
	err = l.r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		logx.Errorf("failed to parse multipart form: %v", err)
		return nil, errorx.NewDefaultError("failed to parse form")
	}

	// Get file from form
	file, header, err := l.r.FormFile("file")
	if err != nil {
		logx.Errorf("failed to get file from form: %v", err)
		return nil, errorx.NewDefaultError("file is required")
	}
	defer file.Close()

	// Generate unique filename
	ext := filepath.Ext(header.Filename)
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	objectName := fmt.Sprintf("%s/%d/%s", req.EntityType, req.EntityId, filename)

	// Upload to MinIO
	_, err = l.svcCtx.MinIOClient.PutObject(
		l.ctx,
		l.svcCtx.Config.MinIO.BucketName,
		objectName,
		file,
		header.Size,
		minio.PutObjectOptions{
			ContentType: header.Header.Get("Content-Type"),
		},
	)
	if err != nil {
		logx.Errorf("failed to upload to MinIO: %v", err)
		return nil, errorx.NewDefaultError("failed to upload file")
	}

	// Generate file URL
	fileUrl := fmt.Sprintf("https://%s/%s/%s", l.svcCtx.Config.MinIO.Endpoint, l.svcCtx.Config.MinIO.BucketName, objectName)

	// Save file record to database
	now := time.Now()
	uploadedFile := &model.AuthUploadedFile{
		UserId:      userId,
		EntityType:  req.EntityType,
		EntityId:    req.EntityId,
		FileName:    header.Filename,
		FilePath:    fileUrl,
		FileSize:    header.Size,
		FileType:    header.Header.Get("Content-Type"),
		BucketName:  l.svcCtx.Config.MinIO.BucketName,
		UploadTime:  now,
		CreatedTime: now,
	}

	result, err := l.svcCtx.AuthUploadedFileModel.Insert(l.ctx, uploadedFile)
	if err != nil {
		logx.Errorf("failed to save file record: %v", err)
		return nil, errorx.NewDefaultError("failed to save file record")
	}

	fileId, err := result.LastInsertId()
	if err != nil {
		logx.Errorf("failed to get last insert id: %v", err)
		return nil, errorx.NewDefaultError("failed to save file record")
	}

	return &types.UploadFileResp{
		FileId:  fileId,
		FileUrl: fileUrl,
	}, nil
}
