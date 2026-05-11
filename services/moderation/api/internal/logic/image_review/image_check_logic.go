package image_review

import (
	"context"
	"io"
	"net/http"

	"github.com/guxiao/community-and-home/common/errorx"
	"github.com/guxiao/community-and-home/services/moderation/api/internal/svc"
	"github.com/guxiao/community-and-home/services/moderation/api/internal/types"
	"github.com/guxiao/community-and-home/services/moderation/internal/engine"

	"github.com/zeromicro/go-zero/core/logx"
)

type ImageCheckLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewImageCheckLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ImageCheckLogic {
	return &ImageCheckLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ImageCheckLogic) ImageCheck(r *http.Request) (*types.ImageCheckResp, error) {
	file, _, err := r.FormFile("file")
	if err != nil {
		return nil, errorx.NewInvalidParamError("file is required")
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, errorx.NewDefaultError("读取文件失败")
	}

	contentType := http.DetectContentType(data)
	if !isImageContentType(contentType) {
		return nil, errorx.NewInvalidParamError("unsupported file type")
	}

	if len(data) > 10*1024*1024 {
		return nil, errorx.NewInvalidParamError("file too large (max 10MB)")
	}

	imgCtx := r.FormValue("context")

	result, err := l.svcCtx.ImageEngine.Check(l.ctx, data, imgCtx)
	if err != nil {
		logx.Errorf("image check error: %v", err)
		return nil, errorx.NewDefaultError("图片审核失败")
	}

	return &types.ImageCheckResp{
		Pass:       result.Pass,
		RiskLevel:  result.RiskLevel,
		Reason:     result.Reason,
		NeedReview: result.NeedReview,
		Details:    convertImageDetails(result.Details),
	}, nil
}

func isImageContentType(ct string) bool {
	switch ct {
	case "image/jpeg", "image/png", "image/gif", "image/webp", "image/bmp":
		return true
	}
	return false
}

func convertImageDetails(details []engine.MatchDetail) []types.MatchDetail {
	if details == nil {
		return []types.MatchDetail{}
	}
	out := make([]types.MatchDetail, 0, len(details))
	for _, d := range details {
		out = append(out, types.MatchDetail{
			Layer:      d.Layer,
			Category:   d.Category,
			Confidence: d.Confidence,
		})
	}
	return out
}
