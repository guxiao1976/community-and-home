package notice

import (
	"context"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-hub/rpc/internal/logic/scope"
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

	// 数据范围读过滤（评审 CRITICAL 补漏：T4.6 只覆盖 List，Get-by-ID 漏网）：
	// reverse-lookup 内容 community_id → FilterAllowed（GLOBAL 放行 / LIMITED IN / EMPTY+无身份拒绝）。
	// 与列表接口同一读路径，避免 LIMITED/EMPTY 用户按 ID（Snowflake 时间有序可枚举/分享链接）越权读全文。
	userID := scope.UserIDFromCtx(l.ctx)
	allowed, err := scope.FilterAllowed(l.ctx, l.svcCtx.PermissionClient, userID, n.CommunityId)
	if err != nil {
		l.Errorf("GetNotice: filter by scope failed: %v", err)
		return nil, err
	}
	if !allowed {
		l.Infof("GetNotice: denied by data scope id=%d community=%d", in.Id, n.CommunityId)
		return &communityv1.GetNoticeResponse{
			Base: scope.DenyBase(),
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
