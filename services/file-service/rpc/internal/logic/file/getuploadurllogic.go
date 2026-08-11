package file

import (
	"context"
	"fmt"
	"time"

	filev1 "github.com/guxiao1976/api-proto/gen/go/file/v1"
	"github.com/guxiao1976/community-common/v2/pkg/errx"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-file/rpc/internal/svc"

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

func (l *GetUploadUrlLogic) GetUploadUrl(in *filev1.GetUploadUrlRequest) (*filev1.GetUploadUrlResponse, error) {
	if l.svcCtx.RawMinio == nil {
		return nil, errx.NewCodeError(ErrCodeFileAccessDenied, "MinIO not available")
	}

	objectKey := fmt.Sprintf("uploads/%d/%d_%s", in.UserId, time.Now().UnixNano(), in.FileName)

	expiryMinutes := 15
	if l.svcCtx.SysConfig != nil {
		if v, err := l.svcCtx.SysConfig.GetInt(l.ctx, "file.upload_url_expiry_minutes"); err == nil {
			expiryMinutes = v
		}
	}
	expiry := time.Duration(expiryMinutes) * time.Minute
	url, err := l.svcCtx.RawMinio.PresignedPutObject(l.ctx, l.svcCtx.Bucket, objectKey, expiry)
	if err != nil {
		l.Errorf("presigned put object failed: %v", err)
		return nil, errx.NewCodeError(ErrCodeFileAccessDenied, "generate upload URL failed")
	}

	return &filev1.GetUploadUrlResponse{
		Base:      responsex.NewBaseResp(),
		UploadUrl: url.String(),
		ObjectKey: objectKey,
		ExpireAt:  time.Now().Add(expiry).Unix(),
	}, nil
}
