package notice

import (
	"context"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-hub/api/internal/svc"
	"github.com/guxiao1976/community-hub/api/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateNoticeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateNoticeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateNoticeLogic {
	return &UpdateNoticeLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// UpdateNotice 更新通知公告。
//
// 注入 JWT 身份 metadata 供 rpc 层 AssertPublishScope 数据权限校验（T4.3），
// 并将 rpc 业务错误（如 080006）透出给客户端。
func (l *UpdateNoticeLogic) UpdateNotice(req *types.UpdateNoticeReq) error {
	callCtx, _, err := l.svcCtx.CallCtx(l.ctx)
	if err != nil {
		return err
	}

	resp, err := l.svcCtx.NoticeServiceRpc.UpdateNotice(callCtx, &communityv1.UpdateNoticeRequest{
		Id:       req.Id,
		Title:    req.Title,
		Content:  req.Content,
		IsPinned: req.IsPinned,
	})
	if err != nil {
		return err
	}
	if !responsex.IsSuccess(resp.GetBase()) {
		return responsex.ToError(resp.GetBase())
	}
	return nil
}
