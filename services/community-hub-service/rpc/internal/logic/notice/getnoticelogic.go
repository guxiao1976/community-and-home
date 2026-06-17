package notice

import (
	"context"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-hub/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetNoticeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetNoticeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetNoticeLogic {
	return &GetNoticeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetNoticeLogic) GetNotice(in *communityv1.GetNoticeRequest) (*communityv1.GetNoticeResponse, error) {
	n, err := l.svcCtx.NoticeModel.FindOne(l.ctx, in.Id)
	if err != nil {
		l.Infof("GetNotice: not found id=%d", in.Id)
		return &communityv1.GetNoticeResponse{
			Base: responsex.NewBaseRespWithError(80001, "通知不存在"),
		}, nil
	}

	// 查询附件
	attachments, _ := l.svcCtx.NoticeAttachmentModel.FindByNoticeId(l.ctx, in.Id)
	pbAttachments := make([]*communityv1.NoticeAttachment, 0, len(attachments))
	for _, a := range attachments {
		pbAttachments = append(pbAttachments, &communityv1.NoticeAttachment{
			Id:       a.Id,
			FileName: a.FileName,
			FileUrl:  a.FileUrl,
			FileSize: a.FileSize,
		})
	}

	pbNotice := toProtoNotice(n)
	pbNotice.Attachments = pbAttachments

	return &communityv1.GetNoticeResponse{
		Base:   responsex.NewBaseResp(),
		Notice: pbNotice,
	}, nil
}
