package file

import (
	"context"
	"time"

	filev1 "github.com/guxiao1976/api-proto/gen/go/file/v1"
	"github.com/guxiao1976/community-common/v2/pkg/errx"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-file/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetFileUrlLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetFileUrlLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFileUrlLogic {
	return &GetFileUrlLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetFileUrlLogic) GetFileUrl(in *filev1.GetFileUrlRequest) (*filev1.GetFileUrlResponse, error) {
	f, err := l.svcCtx.FileModel.FindOne(l.ctx, in.FileId)
	if err != nil {
		return nil, errx.NewCodeError(ErrCodeFileNotFound, "file not found")
	}

	expireSeconds := in.ExpireSeconds
	maxExpire := int32(604800)
	if l.svcCtx.SysConfig != nil {
		if v, err := l.svcCtx.SysConfig.GetInt(l.ctx, "file.url_expiry.max_seconds"); err == nil {
			maxExpire = int32(v)
		}
	}
	defaultExpire := int32(3600)
	if l.svcCtx.SysConfig != nil {
		if v, err := l.svcCtx.SysConfig.GetInt(l.ctx, "file.url_expiry.default_seconds"); err == nil {
			defaultExpire = int32(v)
		}
	}
	if expireSeconds <= 0 || expireSeconds > maxExpire {
		expireSeconds = defaultExpire
	}

	url, err := l.svcCtx.MinioCli.GetURL(f.FilePath, time.Duration(expireSeconds)*time.Second)
	if err != nil {
		l.Errorf("get file url failed: %v", err)
		return nil, errx.NewCodeError(ErrCodeFileAccessDenied, "generate download URL failed")
	}

	return &filev1.GetFileUrlResponse{
		Base:        responsex.NewBaseResp(),
		DownloadUrl: url,
		File:        toProtoFile(f),
	}, nil
}
